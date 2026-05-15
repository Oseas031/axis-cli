// Package agent provides agent execution capabilities.
package agent

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/model/tool"
)

// SimulationAgentExecutor simulates a user running the full Axis pipeline
// using only the bash tool.
type SimulationAgentExecutor struct {
	bash tool.Tool
}

// NewSimulationAgentExecutor creates a new simulation executor.
func NewSimulationAgentExecutor() *SimulationAgentExecutor {
	return &SimulationAgentExecutor{bash: tool.NewBashTool()}
}

// Execute runs the full pipeline: start → submit → status → stop.
func (e *SimulationAgentExecutor) Execute(ctx context.Context, req *AgentExecutionRequest) (*AgentExecutionResult, error) {
	bashDrivePrefix := e.detectBashDrivePrefix(ctx)
	binary := filepath.ToSlash("./axis-dev.exe")
	if b, ok := req.Task.Input["axis_binary"].(string); ok && b != "" {
		binary = filepath.ToSlash(b)
	}
	prompt := "smoke test"
	if p, ok := req.Task.Input["task_prompt"].(string); ok && p != "" {
		prompt = p
	}
	taskID := "sim-" + time.Now().Format("20060102-150405")
	if t, ok := req.Task.Input["task_id"].(string); ok && t != "" {
		taskID = t
	}
	root := "."
	if r, ok := req.Task.Input["project_root"].(string); ok && r != "" {
		root = toBashPath(filepath.ToSlash(r), bashDrivePrefix)
	}
	binary = toBashPath(binary, bashDrivePrefix)
	rootArg := shellQuote(root)
	binaryArg := shellQuote(binary)
	promptArg := shellQuote(prompt)
	taskIDArg := shellQuote(taskID)

	logs := make([]string, 0, 8)

	run := func(label, cmd string) string {
		res, err := e.bash.Execute(ctx, map[string]any{"command": cmd})
		stdout := ""
		exitCode := 0
		if err != nil {
			stdout = fmt.Sprintf("err:%v", err)
			exitCode = -1
		} else {
			if s, ok := res["stdout"].(string); ok {
				stdout = s
			}
			if c, ok := res["exit_code"].(int); ok {
				exitCode = c
			}
		}
		logs = append(logs, fmt.Sprintf("[%s] exit=%d out=%s", label, exitCode, truncate(stdout, 80)))
		return stdout
	}

	// Step 1: cleanup old runtime.
	run("cleanup", fmt.Sprintf("cd %s && rm -rf .axis/", rootArg))

	// Step 2: start runtime in background, wait briefly for process init.
	run("start", fmt.Sprintf("cd %s && mkdir -p .axis && nohup %s start > .axis/start.log 2>&1 & sleep 2; echo started", rootArg, binaryArg))

	// Step 3: wait for runtime.json (up to 30s).
	var addr string
	var pid string
	for i := 0; i < 30; i++ {
		out := run("wait", fmt.Sprintf("cd %s && test -f .axis/runtime.json && cat .axis/runtime.json || echo missing", rootArg))
		if !strings.Contains(out, "missing") {
			addr = extractField(out, "address")
			pid = extractField(out, "pid")
			if addr != "" {
				break
			}
		}
		time.Sleep(time.Second)
	}
	if addr == "" {
		return result(logs, "runtime_not_ready", false)
	}

	// Step 4: submit task via ask --submit.
	run("submit", fmt.Sprintf("cd %s && %s ask %s --submit --task-id %s", rootArg, binaryArg, promptArg, taskIDArg))

	// Step 5: poll status until terminal (up to 30s).
	finalStatus := ""
	for i := 0; i < 30; i++ {
		out := run("poll", fmt.Sprintf("cd %s && %s status %s", rootArg, binaryArg, taskIDArg))
		finalStatus = parseStatusLine(out)
		if finalStatus != "" && finalStatus != "pending" && finalStatus != "running" {
			break
		}
		time.Sleep(time.Second)
	}

	// Step 6: stop runtime and cleanup.
	if pid != "" {
		stopScript := "Stop-Process -Id " + pid + " -Force"
		run("stop", fmt.Sprintf("if command -v powershell.exe >/dev/null 2>&1; then powershell.exe -NoProfile -Command %s >/dev/null 2>&1 || true; else kill %s 2>/dev/null || true; fi; sleep 1", shellQuote(stopScript), shellQuote(pid)))
	}
	run("cleanup2", fmt.Sprintf("cd %s && rm -rf .axis/", rootArg))

	ok := finalStatus == "completed"
	return result(logs, finalStatus, ok)
}

// GetAutonomyLevel returns the executor autonomy level.
func (e *SimulationAgentExecutor) GetAutonomyLevel() AutonomyLevel {
	return AutonomyLevelExecute
}

func (e *SimulationAgentExecutor) detectBashDrivePrefix(ctx context.Context) string {
	res, err := e.bash.Execute(ctx, map[string]any{"command": "if [ -d /mnt/c ]; then echo /mnt; elif [ -d /c ]; then echo /; else echo /; fi"})
	if err != nil {
		return "/"
	}
	stdout, _ := res["stdout"].(string)
	prefix := strings.TrimSpace(stdout)
	if prefix == "/mnt" {
		return "/mnt"
	}
	return "/"
}

// toBashPath converts a Windows path (C:/foo) to a POSIX path understood by bash.
func toBashPath(p string, drivePrefix string) string {
	if len(p) >= 2 && p[1] == ':' {
		if drivePrefix == "/mnt" {
			return "/mnt/" + strings.ToLower(string(p[0])) + p[2:]
		}
		return "/" + strings.ToLower(string(p[0])) + p[2:]
	}
	return p
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func result(logs []string, status string, ok bool) (*AgentExecutionResult, error) {
	return &AgentExecutionResult{
		Output: map[string]any{
			"logs":   logs,
			"status": status,
			"ok":     ok,
		},
	}, nil
}

func extractField(s, key string) string {
	prefix := fmt.Sprintf("\"%s\":", key)
	i := strings.Index(s, prefix)
	if i < 0 {
		return ""
	}
	rest := s[i+len(prefix):]
	for len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t') {
		rest = rest[1:]
	}
	if len(rest) > 0 && rest[0] == '"' {
		end := strings.Index(rest[1:], "\"")
		if end >= 0 {
			return rest[1 : end+1]
		}
	}
	end := 0
	for end < len(rest) && ((rest[end] >= '0' && rest[end] <= '9') || rest[end] == '.' || rest[end] == '-') {
		end++
	}
	if end > 0 {
		return rest[:end]
	}
	return ""
}

func parseStatusLine(s string) string {
	// Only parse the first line ("Task X status: Y") to avoid matching
	// JSON fields in the output body that also contain "status:".
	firstLine := s
	if nl := strings.IndexByte(s, '\n'); nl >= 0 {
		firstLine = s[:nl]
	}
	i := strings.Index(firstLine, "status:")
	if i < 0 {
		return ""
	}
	return strings.TrimSpace(firstLine[i+7:])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
