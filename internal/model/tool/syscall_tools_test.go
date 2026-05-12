package tool

import (
	"context"
	"testing"
)

func TestYieldTool(t *testing.T) {
	yt := NewYieldTool()
	if yt.Name() != "yield" {
		t.Errorf("Name() = %q", yt.Name())
	}
	schema := yt.Schema()
	if schema.Name != "yield" {
		t.Error("schema name mismatch")
	}
	result, err := yt.Execute(context.Background(), map[string]any{"reason": "waiting for human"})
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "yielded" {
		t.Errorf("status = %v", result["status"])
	}
	if result["reason"] != "waiting for human" {
		t.Errorf("reason = %v", result["reason"])
	}
}

func TestCheckpointTool(t *testing.T) {
	ct := NewCheckpointTool()
	if ct.Name() != "checkpoint" {
		t.Errorf("Name() = %q", ct.Name())
	}
	schema := ct.Schema()
	if len(schema.Parameters) != 2 {
		t.Errorf("expected 2 params, got %d", len(schema.Parameters))
	}
	result, err := ct.Execute(context.Background(), map[string]any{
		"summary":   "completed 3 of 5 files",
		"next_step": "process remaining 2 files",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "checkpointed" {
		t.Errorf("status = %v", result["status"])
	}
	if result["summary"] != "completed 3 of 5 files" {
		t.Errorf("summary = %v", result["summary"])
	}
}
