package tool

import (
	"context"

	"github.com/axis-cli/axis/internal/types"
)

// YieldTool allows an Agent to voluntarily yield execution.
// The scheduler can reassign the task or resume later.
type YieldTool struct{}

func NewYieldTool() *YieldTool { return &YieldTool{} }

func (t *YieldTool) Name() string { return "yield" }

func (t *YieldTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "yield",
		Description: "Voluntarily yield execution. Use when waiting for external input or when the task should be paused and resumed later.",
		Parameters: []types.FieldDef{
			{Name: "reason", Type: types.FieldTypeString, Required: false, Description: "Why yielding (optional)"},
		},
	}
}

func (t *YieldTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	return map[string]any{
		"status":  "yielded",
		"message": "Agent yielded. Execution paused.",
	}, nil
}

// CheckpointTool allows an Agent to persist intermediate state.
// Enables crash recovery and long-task resumption.
type CheckpointTool struct{}

func NewCheckpointTool() *CheckpointTool { return &CheckpointTool{} }

func (t *CheckpointTool) Name() string { return "checkpoint" }

func (t *CheckpointTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "checkpoint",
		Description: "Save intermediate execution state. Use during long tasks to enable resumption after interruption.",
		Parameters: []types.FieldDef{
			{Name: "summary", Type: types.FieldTypeString, Required: true, Description: "Summary of current progress"},
			{Name: "next_step", Type: types.FieldTypeString, Required: false, Description: "What to do next when resumed"},
		},
	}
}

func (t *CheckpointTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	summary, _ := input["summary"].(string)
	return map[string]any{
		"status":  "checkpoint_recorded",
		"summary": summary,
	}, nil
}
