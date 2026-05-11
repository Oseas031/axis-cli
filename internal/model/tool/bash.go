package tool

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

const bashOutputLimit = 64 * 1024

// BashTool executes bash shell commands.
type BashTool struct{}

// NewBashTool creates a new BashTool.
func NewBashTool() *BashTool { return &BashTool{} }

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

// Execute runs a bash command and returns its output.
func (t *BashTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	cmdStr, ok := input["command"].(string)
	if !ok || cmdStr == "" {
		return map[string]any{"error": "command is required and must be a string"}, nil
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
