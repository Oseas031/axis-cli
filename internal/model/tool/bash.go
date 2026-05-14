package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

const bashOutputLimit = 64 * 1024

// PermissionLevel controls which commands the BashTool is allowed to execute.
type PermissionLevel int

const (
	// PermissionL0 allows only safe read-only commands (whitelist).
	PermissionL0 PermissionLevel = iota
	// PermissionL1 allows build/test commands in addition to L0.
	PermissionL1
	// PermissionUnrestricted allows any command (current default for backward compat).
	PermissionUnrestricted
)

// l0Whitelist is the Level 0 command prefix whitelist — safe, read-only operations.
var l0Whitelist = []string{
	"ls", "dir", "cat", "head", "tail", "less", "more",
	"grep", "rg", "find", "wc", "sort", "uniq", "diff",
	"echo", "printf", "pwd", "env", "which", "where",
	"git status", "git log", "git diff", "git show", "git branch",
	"git rev-parse", "git ls-files", "git ls-tree",
	"go vet", "go doc", "gofmt -l",
}

// l1Whitelist extends L0 with build/test commands.
var l1Whitelist = []string{
	"go build", "go test", "go run",
	"go mod tidy", "go mod download",
	"staticcheck", "gosec",
	"npm run build", "npm test", "npm run",
}

// BashTool executes bash shell commands.
type BashTool struct {
	permLevel PermissionLevel
}

// NewBashTool creates a new BashTool with unrestricted permissions (backward compat).
func NewBashTool() *BashTool { return &BashTool{permLevel: PermissionUnrestricted} }

// NewBashToolWithLevel creates a BashTool with the specified permission level.
func NewBashToolWithLevel(level PermissionLevel) *BashTool {
	return &BashTool{permLevel: level}
}

// Name returns the tool name.
func (t *BashTool) Name() string { return "bash" }

// Schema returns the tool definition for bash.
func (t *BashTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "bash",
		Description: "Execute a bash shell command",
		Parameters: []types.FieldDef{
			{Name: "command", Type: types.FieldTypeString, Required: true, Description: "The command to execute"},
		},
	}
}

// checkPermission validates the command against the current permission level.
func (t *BashTool) checkPermission(cmd string) error {
	if t.permLevel == PermissionUnrestricted {
		return nil
	}
	trimmed := strings.TrimSpace(cmd)
	allowed := l0Whitelist
	if t.permLevel >= PermissionL1 {
		allowed = append(allowed, l1Whitelist...)
	}
	for _, prefix := range allowed {
		if strings.HasPrefix(trimmed, prefix) {
			return nil
		}
	}
	return fmt.Errorf("command not allowed at permission level %d: %s", t.permLevel, firstWord(trimmed))
}

func firstWord(s string) string {
	if i := strings.IndexByte(s, ' '); i > 0 {
		return s[:i]
	}
	return s
}

// Execute runs a bash command and returns its output.
func (t *BashTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	cmdStr, ok := input["command"].(string)
	if !ok || cmdStr == "" {
		return map[string]any{"error": "command is required and must be a string"}, nil
	}
	if err := t.checkPermission(cmdStr); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	start := time.Now()
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// #nosec G204
	cmd := exec.CommandContext(timeoutCtx, "bash", "-c", cmdStr)
	cmd.Dir = cwd
	// Inject Windows tool paths so Agent can find go, git, etc. without guessing
	cmd.Env = enrichedEnv()
	output, execErr := cmd.CombinedOutput()
	exitCode := 0
	timedOut := false
	if execErr != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			timedOut = true
			exitCode = -1
		} else if exitErr, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, execErr
		}
	}
	outputTruncated := false
	if len(output) > bashOutputLimit {
		output = output[:bashOutputLimit]
		outputTruncated = true
	}
	return map[string]any{
		"command":          cmdStr,
		"cwd":              cwd,
		"stdout":           string(output),
		"exit_code":        exitCode,
		"duration_ms":      time.Since(start).Milliseconds(),
		"timed_out":        timedOut,
		"output_truncated": outputTruncated,
	}, nil
}


// enrichedEnv returns the current environment with additional tool paths appended.
// On Windows/WSL, this ensures go, git, cmd.exe etc. are discoverable.
func enrichedEnv() []string {
	env := os.Environ()
	// Additional paths for common Windows tools accessible from WSL/bash
	extraPaths := []string{
		"/usr/local/go/bin",
		"/usr/local/bin",
		"/mnt/c/Program Files/Go/bin",
		"/mnt/c/Windows/System32",
		"/mnt/c/Windows",
	}
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = e + ":" + strings.Join(extraPaths, ":")
			return env
		}
	}
	// No PATH found, create one
	env = append(env, "PATH=/usr/local/bin:/usr/bin:/bin:"+strings.Join(extraPaths, ":"))
	return env
}
