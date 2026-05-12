package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// SpawnTool allows an Agent to create an isolated subtask.
// The subtask runs with a clean context (no parent history leakage)
// and returns only a summary to the parent.
type SpawnTool struct{}

func NewSpawnTool() *SpawnTool { return &SpawnTool{} }

func (t *SpawnTool) Name() string { return "spawn" }

func (t *SpawnTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "spawn",
		Description: "Create an isolated subtask. The subtask runs with a clean context and returns a summary. Use for delegating focused work without polluting the current context.",
		Parameters: []types.FieldDef{
			{Name: "task_id", Type: types.FieldTypeString, Required: true, Description: "Unique ID for the subtask"},
			{Name: "prompt", Type: types.FieldTypeString, Required: true, Description: "Instructions for the subtask"},
			{Name: "isolation", Type: types.FieldTypeString, Required: false, Description: "Isolation level: full (default) or shared"},
		},
	}
}

func (t *SpawnTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	taskID, _ := input["task_id"].(string)
	prompt, _ := input["prompt"].(string)
	isolation, _ := input["isolation"].(string)

	if taskID == "" || prompt == "" {
		return map[string]any{"error": "task_id and prompt are required"}, nil
	}
	if isolation == "" {
		isolation = "full"
	}
	if isolation != "full" && isolation != "shared" {
		return map[string]any{"error": fmt.Sprintf("invalid isolation level: %s (use 'full' or 'shared')", isolation)}, nil
	}

	// In P0, spawn records the intent. The orchestrator/scheduler
	// picks up spawned tasks and executes them with isolation guarantees:
	// - full: clean history, no parent context, limited tool scope
	// - shared: inherits parent's context snapshot (read-only), own history
	return map[string]any{
		"status":    "spawned",
		"task_id":   taskID,
		"prompt":    prompt,
		"isolation": isolation,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"message":   fmt.Sprintf("Subtask %s spawned with %s isolation. Result will be returned as a summary.", taskID, isolation),
	}, nil
}
