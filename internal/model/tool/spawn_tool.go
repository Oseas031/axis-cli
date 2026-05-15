package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// IsolationPolicy declares the context boundary for a spawned subtask.
type IsolationPolicy struct {
	InheritMemory   bool     `json:"inherit_memory"`   // default: false
	InheritContext  bool     `json:"inherit_context"`  // default: false
	SharedArtifacts []string `json:"shared_artifacts"` // explicit list of artifacts to pass
	// v1: binary diversity flag. TODO: explicit provider exclusion list
	RequireProviderDiversity bool `json:"require_provider_diversity"` // spawned agent must use different provider
	CoTIsolation             bool `json:"cot_isolation"`              // hide parent's chain-of-thought from spawned agent
}

// DefaultIsolationPolicy returns a policy with full isolation (no inheritance).
func DefaultIsolationPolicy() IsolationPolicy {
	return IsolationPolicy{
		InheritMemory:   false,
		InheritContext:  false,
		SharedArtifacts: nil,
	}
}

// DevilsAdvocatePolicy returns an IsolationPolicy configured for adversarial
// review: no shared memory/context, provider diversity enforced, and CoT hidden
// from the reviewer to prevent Alignment Hallucination (arXiv:2605.10698).
func DevilsAdvocatePolicy() IsolationPolicy {
	return IsolationPolicy{
		InheritMemory:            false,
		InheritContext:           false,
		RequireProviderDiversity: true,
		CoTIsolation:             true,
	}
}

// SpawnTool allows an Agent to create an isolated subtask.
// The subtask runs with a clean context (no parent history leakage)
// and returns only a summary to the parent.
type SpawnTool struct {
	execFn func(ctx context.Context, taskID, prompt, isolation string) (map[string]any, error)
}

func NewSpawnTool() *SpawnTool { return &SpawnTool{} }

// SetExecFn injects an active executor. When set, Execute delegates to it
// instead of returning a passive stub response.
func (t *SpawnTool) SetExecFn(fn func(ctx context.Context, taskID, prompt, isolation string) (map[string]any, error)) {
	t.execFn = fn
}

func (t *SpawnTool) Name() string { return "spawn" }

func (t *SpawnTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "spawn",
		Description: "Create an isolated subtask. The subtask runs with a clean context and returns a summary. Use for delegating focused work without polluting the current context.",
		Parameters: []types.FieldDef{
			{Name: "task_id", Type: types.FieldTypeString, Required: true, Description: "Unique ID for the subtask"},
			{Name: "prompt", Type: types.FieldTypeString, Required: true, Description: "Instructions for the subtask"},
			{Name: "isolation", Type: types.FieldTypeString, Required: false, Description: "Isolation level: full (default) or shared"},
			{Name: "shared_artifacts", Type: types.FieldTypeArray, Required: false, Description: "Explicit list of artifact IDs to pass to subtask"},
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

	// Delegate to active executor if wired
	if t.execFn != nil {
		return t.execFn(ctx, taskID, prompt, isolation)
	}

	// Build isolation policy based on isolation level
	policy := DefaultIsolationPolicy()
	if isolation == "shared" {
		policy.InheritContext = true
	}

	// Parse shared_artifacts if provided
	if artifacts, ok := input["shared_artifacts"]; ok {
		if arr, ok := artifacts.([]any); ok {
			for _, a := range arr {
				if s, ok := a.(string); ok {
					policy.SharedArtifacts = append(policy.SharedArtifacts, s)
				}
			}
		}
	}

	policyJSON, _ := json.Marshal(policy)

	return map[string]any{
		"status":           "spawned",
		"task_id":          taskID,
		"prompt":           prompt,
		"isolation":        isolation,
		"isolation_policy": string(policyJSON),
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"message":          fmt.Sprintf("Subtask %s spawned with %s isolation. Result will be returned as a summary.", taskID, isolation),
	}, nil
}
