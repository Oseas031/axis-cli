package tool

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSpawnTool_Name(t *testing.T) {
	st := NewSpawnTool()
	if st.Name() != "spawn" {
		t.Errorf("Name() = %q", st.Name())
	}
}

func TestSpawnTool_Schema(t *testing.T) {
	st := NewSpawnTool()
	schema := st.Schema()
	if schema.Name != "spawn" {
		t.Error("schema name mismatch")
	}
	if len(schema.Parameters) != 4 {
		t.Errorf("expected 4 params, got %d", len(schema.Parameters))
	}
}

func TestSpawnTool_Execute_FullIsolation(t *testing.T) {
	st := NewSpawnTool()
	result, err := st.Execute(context.Background(), map[string]any{
		"task_id": "sub-1",
		"prompt":  "analyze this file",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "spawned" {
		t.Errorf("status = %v", result["status"])
	}
	if result["isolation"] != "full" {
		t.Errorf("isolation = %v (expected full default)", result["isolation"])
	}
}

func TestSpawnTool_Execute_SharedIsolation(t *testing.T) {
	st := NewSpawnTool()
	result, err := st.Execute(context.Background(), map[string]any{
		"task_id":   "sub-2",
		"prompt":    "review code",
		"isolation": "shared",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["isolation"] != "shared" {
		t.Errorf("isolation = %v", result["isolation"])
	}
}

func TestSpawnTool_Execute_MissingRequired(t *testing.T) {
	st := NewSpawnTool()
	result, _ := st.Execute(context.Background(), map[string]any{})
	if result["error"] == nil {
		t.Error("expected error for missing fields")
	}
}

func TestSpawnTool_Execute_InvalidIsolation(t *testing.T) {
	st := NewSpawnTool()
	result, _ := st.Execute(context.Background(), map[string]any{
		"task_id":   "sub-3",
		"prompt":    "test",
		"isolation": "invalid",
	})
	if result["error"] == nil {
		t.Error("expected error for invalid isolation level")
	}
}

func TestSpawnTool_DefaultIsolationPolicy(t *testing.T) {
	st := NewSpawnTool()
	result, err := st.Execute(context.Background(), map[string]any{
		"task_id": "sub-isolated",
		"prompt":  "do isolated work",
	})
	if err != nil {
		t.Fatal(err)
	}

	policyStr, ok := result["isolation_policy"].(string)
	if !ok {
		t.Fatal("isolation_policy not present in result")
	}

	var policy IsolationPolicy
	if err := json.Unmarshal([]byte(policyStr), &policy); err != nil {
		t.Fatalf("failed to parse isolation_policy: %v", err)
	}

	if policy.InheritMemory {
		t.Error("default InheritMemory should be false")
	}
	if policy.InheritContext {
		t.Error("default InheritContext should be false")
	}
	if len(policy.SharedArtifacts) != 0 {
		t.Errorf("default SharedArtifacts should be empty, got %v", policy.SharedArtifacts)
	}
}

func TestSpawnTool_SharedIsolationPolicy(t *testing.T) {
	st := NewSpawnTool()
	result, err := st.Execute(context.Background(), map[string]any{
		"task_id":   "sub-shared",
		"prompt":    "shared work",
		"isolation": "shared",
	})
	if err != nil {
		t.Fatal(err)
	}

	var policy IsolationPolicy
	json.Unmarshal([]byte(result["isolation_policy"].(string)), &policy)

	if policy.InheritMemory {
		t.Error("shared isolation should not inherit memory")
	}
	if !policy.InheritContext {
		t.Error("shared isolation should inherit context")
	}
}

func TestSpawnTool_SharedArtifacts(t *testing.T) {
	st := NewSpawnTool()
	result, err := st.Execute(context.Background(), map[string]any{
		"task_id":          "sub-artifacts",
		"prompt":           "use these artifacts",
		"shared_artifacts": []any{"artifact-1", "artifact-2"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var policy IsolationPolicy
	json.Unmarshal([]byte(result["isolation_policy"].(string)), &policy)

	if len(policy.SharedArtifacts) != 2 {
		t.Fatalf("expected 2 shared artifacts, got %d", len(policy.SharedArtifacts))
	}
	if policy.SharedArtifacts[0] != "artifact-1" || policy.SharedArtifacts[1] != "artifact-2" {
		t.Errorf("unexpected artifacts: %v", policy.SharedArtifacts)
	}
}


func TestDevilsAdvocatePolicy(t *testing.T) {
	p := DevilsAdvocatePolicy()

	if p.InheritMemory {
		t.Error("expected InheritMemory=false")
	}
	if p.InheritContext {
		t.Error("expected InheritContext=false")
	}
	if !p.RequireProviderDiversity {
		t.Error("expected RequireProviderDiversity=true")
	}
	if !p.CoTIsolation {
		t.Error("expected CoTIsolation=true")
	}
	if len(p.SharedArtifacts) != 0 {
		t.Error("expected empty SharedArtifacts")
	}
}

func TestIsolationPolicyDefaults(t *testing.T) {
	var p IsolationPolicy

	if p.RequireProviderDiversity {
		t.Error("expected default RequireProviderDiversity=false")
	}
	if p.CoTIsolation {
		t.Error("expected default CoTIsolation=false")
	}
}
