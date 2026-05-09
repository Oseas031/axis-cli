package tool

import (
	"context"
	"os/exec"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

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
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(timeoutCtx, "bash", "-c", cmdStr)
	output, execErr := cmd.CombinedOutput()
	exitCode := 0
	if execErr != nil {
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, execErr
		}
	}
	return map[string]any{
		"stdout":    string(output),
		"exit_code": exitCode,
	}, nil
}
