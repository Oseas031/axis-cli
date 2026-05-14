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
		t.Errorf("status = %v, want yielded", result["status"])
	}
	if result["message"] != "Agent yielded. Execution paused." {
		t.Errorf("message = %v", result["message"])
	}
}

func TestCheckpointTool(t *testing.T) {
	ct := NewCheckpointTool()
	if ct.Name() != "checkpoint" {
		t.Errorf("Name() = %q", ct.Name())
	}
	result, err := ct.Execute(context.Background(), map[string]any{
		"summary": "completed 3 of 5 files",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "checkpoint_recorded" {
		t.Errorf("status = %v, want checkpoint_recorded", result["status"])
	}
	if result["summary"] != "completed 3 of 5 files" {
		t.Errorf("summary = %v", result["summary"])
	}
}
