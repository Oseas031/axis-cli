package judgement

import (
	"fmt"
	"testing"
)

func TestIsolateContext_StripsIntermediateToolCalls(t *testing.T) {
	fullInput := map[string]any{
		"task_spec":       "build feature X",
		"contract_schema": "schema-v1",
		"tool_calls": []any{
			map[string]any{"tool": "bash", "output": "intermediate step 1"},
			map[string]any{"tool": "bash", "output": "intermediate step 2"},
			map[string]any{"tool": "file_read", "output": "reading something"},
		},
		"final_artifacts": []any{
			map[string]any{"type": "file", "path": "main.go", "content": "package main"},
		},
	}

	result := IsolateContext(fullInput)

	if len(result.FinalArtifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(result.FinalArtifacts))
	}
	if result.TaskSpec != "build feature X" {
		t.Errorf("expected task_spec preserved, got %q", result.TaskSpec)
	}
	if result.ContractSchema != "schema-v1" {
		t.Errorf("expected contract_schema preserved, got %q", result.ContractSchema)
	}
}

func TestIsolateContext_PreservesFinalArtifacts(t *testing.T) {
	fullInput := map[string]any{
		"task_spec": "deploy service",
		"final_artifacts": []any{
			map[string]any{"type": "file", "path": "cmd/main.go", "content": "package main", "diff": "+package main"},
			map[string]any{"type": "config", "path": "config.yaml", "content": "key: val"},
		},
	}

	result := IsolateContext(fullInput)

	if len(result.FinalArtifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(result.FinalArtifacts))
	}

	a := result.FinalArtifacts[0]
	if a.Type != "file" || a.Path != "cmd/main.go" || a.Content != "package main" || a.Diff != "+package main" {
		t.Errorf("first artifact not preserved correctly: %+v", a)
	}

	b := result.FinalArtifacts[1]
	if b.Type != "config" || b.Path != "config.yaml" || b.Content != "key: val" {
		t.Errorf("second artifact not preserved correctly: %+v", b)
	}
}

func TestIsolateContext_LargeInputReducedToFinalState(t *testing.T) {
	// Simulate 100+ intermediate entries
	toolCalls := make([]any, 150)
	for i := range toolCalls {
		toolCalls[i] = map[string]any{
			"tool":   "bash",
			"output": fmt.Sprintf("step %d output with lots of data", i),
		}
	}

	fullInput := map[string]any{
		"task_spec":  "large task",
		"tool_calls": toolCalls,
		"steps":      toolCalls, // another bloat field
		"final_artifacts": []any{
			map[string]any{"type": "file", "path": "result.go", "content": "done"},
		},
	}

	result := IsolateContext(fullInput)

	// Only final artifact survives — no intermediate data
	if len(result.FinalArtifacts) != 1 {
		t.Fatalf("expected 1 final artifact, got %d", len(result.FinalArtifacts))
	}
	if result.FinalArtifacts[0].Path != "result.go" {
		t.Errorf("expected result.go, got %s", result.FinalArtifacts[0].Path)
	}
	if result.TaskSpec != "large task" {
		t.Errorf("expected task_spec preserved, got %q", result.TaskSpec)
	}
}

func TestIsolateContext_NilInput(t *testing.T) {
	result := IsolateContext(nil)
	if result == nil {
		t.Fatal("expected non-nil result for nil input")
	}
	if len(result.FinalArtifacts) != 0 {
		t.Errorf("expected empty artifacts for nil input")
	}
}

func TestIsolateContext_AlreadyIsolated(t *testing.T) {
	input := &IsolatedJudgeInput{
		TaskSpec:       "already done",
		FinalArtifacts: []Artifact{{Type: "file", Path: "x.go"}},
	}

	result := IsolateContext(input)
	if result != input {
		t.Error("expected same pointer returned for already-isolated input")
	}
}
