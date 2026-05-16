package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// ScopedBashTool wraps BashTool with a fixed workspace as cwd.
type ScopedBashTool struct {
	inner     *BashTool
	workspace string
}

// NewScopedBashTool creates a bash tool scoped to the given workspace directory.
func NewScopedBashTool(workspace string) *ScopedBashTool {
	return &ScopedBashTool{inner: NewBashTool(), workspace: workspace}
}

func (s *ScopedBashTool) Name() string                 { return "bash" }
func (s *ScopedBashTool) Schema() types.ToolDefinition { return s.inner.Schema() }

func (s *ScopedBashTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	cmdStr, ok := input["command"].(string)
	if !ok || cmdStr == "" {
		return map[string]any{"error": "command is required and must be a string"}, nil
	}
	if err := s.inner.checkPermission(cmdStr); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	start := time.Now()
	if err := os.MkdirAll(s.workspace, 0o755); err != nil {
		return map[string]any{"error": "workspace not accessible: " + err.Error()}, nil
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(timeoutCtx, "bash", "-c", cmdStr)
	cmd.Dir = s.workspace
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
		"cwd":              toWSLPath(s.workspace),
		"stdout":           string(output),
		"exit_code":        exitCode,
		"duration_ms":      time.Since(start).Milliseconds(),
		"timed_out":        timedOut,
		"output_truncated": outputTruncated,
	}, nil
}

// ScopedFileWriteTool restricts file_write to a workspace directory.
type ScopedFileWriteTool struct {
	inner *FileWriteTool
}

// NewScopedFileWriteTool creates a file_write tool restricted to the workspace.
func NewScopedFileWriteTool(workspace string) *ScopedFileWriteTool {
	return &ScopedFileWriteTool{inner: NewFileWriteTool([]string{workspace})}
}

func (s *ScopedFileWriteTool) Name() string                 { return "file_write" }
func (s *ScopedFileWriteTool) Schema() types.ToolDefinition { return s.inner.Schema() }
func (s *ScopedFileWriteTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	return s.inner.Execute(ctx, input)
}

// NewScopedRegistry creates a registry with bash and file_write scoped to workspace.
// All other tools are copied unchanged.
func NewScopedRegistry(original *Registry, workspace string) *Registry {
	scoped := NewRegistry()
	original.mu.RLock()
	defer original.mu.RUnlock()
	for name, t := range original.tools {
		switch name {
		case "bash":
			_ = scoped.Register(NewScopedBashTool(workspace), nil)
		case "file_write":
			_ = scoped.Register(NewScopedFileWriteTool(workspace), nil)
		default:
			_ = scoped.Register(t, original.scopes[name])
		}
	}
	return scoped
}

// v1: ScopedBashTool duplicates execution logic from BashTool.Execute for workspace override.
// TODO: refactor BashTool to accept cwd parameter to reduce duplication.
var _ Tool = (*ScopedBashTool)(nil)
var _ Tool = (*ScopedFileWriteTool)(nil)

func init() {
	// Ensure format import is used (for godoc, not runtime)
	_ = fmt.Sprintf
}
