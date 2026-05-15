package compactor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestOffloadCompactor_NoCompactionBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	cfg := CompactConfig{DataDir: dir, TokenBudget: 100000, MildRatio: 0.5}
	c, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	history := []types.ModelMessage{
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc1", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc1", Content: "short output"},
	}

	result := c.Compact(context.Background(), history)
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
	if result[1].Content != "short output" {
		t.Fatalf("content should be unchanged, got %q", result[1].Content)
	}
}

func TestOffloadCompactor_MildCompaction(t *testing.T) {
	dir := t.TempDir()
	// Set a very low budget so mild triggers immediately
	cfg := CompactConfig{DataDir: dir, TokenBudget: 100, MildRatio: 0.5}
	c, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create a history with a large tool result
	largeContent := strings.Repeat("x", 2000)
	history := []types.ModelMessage{
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc1", Name: "file_read"}}},
		{Role: "tool", ToolCallID: "tc1", Content: largeContent},
	}

	result := c.Compact(context.Background(), history)

	// Tool content should be replaced with offload marker
	if !strings.HasPrefix(result[1].Content, "[axis:offload:") {
		t.Fatalf("expected offloaded marker, got %q", result[1].Content[:50])
	}

	// Verify ref file was written
	refs, _ := os.ReadDir(filepath.Join(dir, "refs"))
	if len(refs) == 0 {
		t.Fatal("expected ref file to be written")
	}

	// Verify offload.jsonl was written
	data, err := os.ReadFile(filepath.Join(dir, "offload.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "tc1") {
		t.Fatal("offload.jsonl should contain tool_call_id")
	}
}

func TestOffloadCompactor_AggressiveCompaction(t *testing.T) {
	dir := t.TempDir()
	// Budget so low that mild can't save it
	cfg := CompactConfig{DataDir: dir, TokenBudget: 50, MildRatio: 0.3, AggressiveRatio: 0.5}
	c, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Multiple pairs, oldest should be dropped
	history := []types.ModelMessage{
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc1", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc1", Content: strings.Repeat("a", 500)},
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc2", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc2", Content: strings.Repeat("b", 500)},
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc3", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc3", Content: "recent"},
	}

	result := c.Compact(context.Background(), history)

	// Should have fewer messages (oldest pairs dropped)
	if len(result) >= len(history) {
		t.Fatalf("expected fewer messages after aggressive compaction, got %d", len(result))
	}
	// Most recent pair should survive (may be offloaded but still present)
	lastToolFound := false
	for _, msg := range result {
		if msg.ToolCallID == "tc3" {
			lastToolFound = true
		}
	}
	if !lastToolFound {
		t.Fatal("most recent tool result (tc3) should survive aggressive compaction")
	}
}

func TestOffloadCompactor_EmergencyCompaction(t *testing.T) {
	dir := t.TempDir()
	cfg := CompactConfig{DataDir: dir, TokenBudget: 10, MildRatio: 0.1, AggressiveRatio: 0.2, EmergencyRatio: 0.3}
	c, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	history := []types.ModelMessage{
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc1", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc1", Content: strings.Repeat("x", 1000)},
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc2", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc2", Content: strings.Repeat("y", 1000)},
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc3", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc3", Content: "last"},
	}

	result := c.Compact(context.Background(), history)

	// Should be drastically reduced
	if len(result) >= len(history) {
		t.Fatalf("emergency should reduce messages, got %d", len(result))
	}
}

func TestOffloadCompactor_IdempotentOnAlreadyOffloaded(t *testing.T) {
	dir := t.TempDir()
	cfg := CompactConfig{DataDir: dir, TokenBudget: 100, MildRatio: 0.5}
	c, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Pre-offloaded content
	history := []types.ModelMessage{
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc1", Name: "bash"}}},
		{Role: "tool", ToolCallID: "tc1", Content: "[axis:offload: bash: something | ref=refs/x.md]"},
	}

	result := c.Compact(context.Background(), history)
	if result[1].Content != history[1].Content {
		t.Fatal("already offloaded content should not be modified")
	}
}

func TestStore_OffloadAndReadRef(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	content := "full tool output here"
	entry, err := store.Offload("tc-123", "bash", content, "ran bash command", 7)
	if err != nil {
		t.Fatal(err)
	}

	if entry.ToolCallID != "tc-123" {
		t.Fatalf("expected tc-123, got %s", entry.ToolCallID)
	}
	if entry.Score != 7 {
		t.Fatalf("expected score 7, got %d", entry.Score)
	}

	// Read back
	got, err := store.ReadRef(entry.ResultRef)
	if err != nil {
		t.Fatal(err)
	}
	if got != content {
		t.Fatalf("expected %q, got %q", content, got)
	}
}

func TestEstimateTokens(t *testing.T) {
	// Pure ASCII
	got := estimateTokens("hello world") // 11 chars / 4 = 2
	if got < 2 || got > 4 {
		t.Fatalf("unexpected token estimate for ASCII: %d", got)
	}

	// CJK
	got = estimateTokens("你好世界") // 4 CJK chars / 1.5 ≈ 2
	if got < 2 || got > 4 {
		t.Fatalf("unexpected token estimate for CJK: %d", got)
	}
}

func TestHeuristicSummarize(t *testing.T) {
	// Short content → low score
	summary, score := heuristicSummarize("bash", "ok")
	if score != 3 {
		t.Fatalf("expected score 3 for short content, got %d", score)
	}
	if !strings.Contains(summary, "bash") {
		t.Fatal("summary should contain tool name")
	}

	// Long content → high score
	_, score = heuristicSummarize("file_read", strings.Repeat("x", 10000))
	if score < 7 {
		t.Fatalf("expected high score for long content, got %d", score)
	}
}
