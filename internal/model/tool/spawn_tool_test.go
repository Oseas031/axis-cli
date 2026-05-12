package tool

import (
	"context"
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
	if len(schema.Parameters) != 3 {
		t.Errorf("expected 3 params, got %d", len(schema.Parameters))
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
