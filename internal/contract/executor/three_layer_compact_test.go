package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestTranscriptStore_Save(t *testing.T) {
	dir := t.TempDir()
	store := &TranscriptStore{Dir: dir}

	history := []types.ModelMessage{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
	}

	path, err := store.Save("test-task", history)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("transcript file not created: %v", err)
	}
}

func TestTranscriptStore_EmptyDir(t *testing.T) {
	store := &TranscriptStore{Dir: ""}
	path, err := store.Save("task", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path for empty dir, got %q", path)
	}
}

func TestThreeLayerCompaction_MicroOnly(t *testing.T) {
	c := &ThreeLayerCompaction{
		Micro:  &ToolResultCompaction{KeepRecent: 1},
		Budget: 100000, // high budget so auto doesn't trigger
	}

	// Build history with 3 tool groups
	history := buildToolHistory(3)
	result := c.Compact(context.Background(), history)

	// Micro should have compacted older tool results
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestThreeLayerCompaction_AutoTriggered(t *testing.T) {
	dir := t.TempDir()
	c := &ThreeLayerCompaction{
		Micro:      &ToolResultCompaction{KeepRecent: 1},
		Auto:       &SummarizationCompaction{Provider: nil, KeepRecent: 2}, // nil provider = no-op summarize
		Budget:     10, // very low budget to trigger auto
		Transcript: &TranscriptStore{Dir: dir},
		TaskID:     "auto-test",
	}

	history := buildToolHistory(5)
	c.Compact(context.Background(), history)

	// Verify transcript was saved (auto layer triggered because tokens > budget)
	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Error("expected transcript file to be saved")
	}
}

func TestThreeLayerCompaction_TranscriptNaming(t *testing.T) {
	dir := t.TempDir()
	store := &TranscriptStore{Dir: dir}
	store.Save("my-task", []types.ModelMessage{{Role: "user", Content: "x"}})

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	name := entries[0].Name()
	if len(name) < len("my-task-") {
		t.Fatalf("unexpected filename: %s", name)
	}
	if filepath.Ext(name) != ".json" {
		t.Errorf("expected .json extension, got %s", name)
	}
}

// buildToolHistory creates a history with N tool call groups.
func buildToolHistory(n int) []types.ModelMessage {
	var history []types.ModelMessage
	for i := 0; i < n; i++ {
		history = append(history, types.ModelMessage{
			Role:      "assistant",
			ToolCalls: []types.ToolCall{{ID: "tc" + string(rune('0'+i)), Name: "bash", Input: map[string]any{"command": "echo"}}},
		})
		history = append(history, types.ModelMessage{
			Role:       "tool",
			ToolCallID: "tc" + string(rune('0'+i)),
			Content:    "output from tool call " + string(rune('0'+i)) + " with some content to take up space",
		})
	}
	return history
}
