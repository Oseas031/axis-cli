package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/types"
)

type mockSummarizer struct{}

func (m *mockSummarizer) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	return &provider.ModelResponse{Output: map[string]any{"text": "Summary of previous conversation."}}, nil
}

func makeToolGroup(id, name, content string) []types.ModelMessage {
	return []types.ModelMessage{
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: id, Name: name}}},
		{Role: "tool", ToolCallID: id, Content: content},
	}
}

func TestEstimateTokens(t *testing.T) {
	msgs := []types.ModelMessage{
		{Role: "user", Content: "hello world"},
		{Role: "assistant", Content: "hi there"},
	}
	tokens := EstimateTokens(msgs)
	if tokens <= 0 {
		t.Error("expected positive token count")
	}
}

func TestToolResultCompaction_KeepsRecent(t *testing.T) {
	var history []types.ModelMessage
	for i := 0; i < 5; i++ {
		history = append(history, makeToolGroup(
			string(rune('a'+i)), "file_read", strings.Repeat("data", 100))...)
	}
	strategy := &ToolResultCompaction{KeepRecent: 3}
	result, modified := strategy.Compact(context.Background(), history, 100)
	if !modified {
		t.Error("expected modification")
	}
	compacted := 0
	for _, msg := range result {
		if msg.Role == "tool" && strings.HasPrefix(msg.Content, "[Tool:") {
			compacted++
		}
	}
	if compacted != 2 {
		t.Errorf("expected 2 compacted, got %d", compacted)
	}
}

func TestToolResultCompaction_NoOpWhenFewGroups(t *testing.T) {
	history := makeToolGroup("a", "bash", strings.Repeat("x", 200))
	strategy := &ToolResultCompaction{KeepRecent: 3}
	_, modified := strategy.Compact(context.Background(), history, 10)
	if modified {
		t.Error("should not modify when fewer groups than KeepRecent")
	}
}

func TestTruncationCompaction_DropsOldest(t *testing.T) {
	var history []types.ModelMessage
	for i := 0; i < 10; i++ {
		history = append(history, makeToolGroup(
			string(rune('a'+i)), "bash", strings.Repeat("output", 50))...)
	}
	strategy := &TruncationCompaction{KeepRecent: 2}
	result, modified := strategy.Compact(context.Background(), history, 200)
	if !modified {
		t.Error("expected modification")
	}
	if len(result) >= len(history) {
		t.Error("expected fewer messages")
	}
}

func TestCompactionPipeline_StopsEarly(t *testing.T) {
	history := []types.ModelMessage{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
	}
	pipeline := &CompactionPipeline{
		Strategies: []CompactionStrategy{&ToolResultCompaction{KeepRecent: 3}},
		Budget:     10000,
	}
	result := pipeline.Compact(context.Background(), history)
	if len(result) != len(history) {
		t.Error("should not modify when under budget")
	}
}

func TestCompactionPipeline_RunsStrategies(t *testing.T) {
	var history []types.ModelMessage
	for i := 0; i < 20; i++ {
		history = append(history, makeToolGroup(
			string(rune('a'+i%26)), "bash", strings.Repeat("x", 200))...)
	}
	pipeline := &CompactionPipeline{
		Strategies: []CompactionStrategy{
			&ToolResultCompaction{KeepRecent: 3},
			&TruncationCompaction{KeepRecent: 4},
		},
		Budget: 500,
	}
	result := pipeline.Compact(context.Background(), history)
	if EstimateTokens(result) > 500 {
		t.Errorf("should reduce to budget, got %d", EstimateTokens(result))
	}
}

func TestSummarizationCompaction_WithMockProvider(t *testing.T) {
	var history []types.ModelMessage
	for i := 0; i < 8; i++ {
		history = append(history, makeToolGroup(
			string(rune('a'+i)), "bash", strings.Repeat("output ", 50))...)
	}
	strategy := &SummarizationCompaction{Provider: &mockSummarizer{}, KeepRecent: 4}
	result, modified := strategy.Compact(context.Background(), history, 100)
	if !modified {
		t.Error("expected modification")
	}
	if !strings.Contains(result[0].Content, "[Session summary") {
		t.Errorf("first message should be summary, got %q", result[0].Content)
	}
}

func TestSummarizationCompaction_NilProvider(t *testing.T) {
	var history []types.ModelMessage
	for i := 0; i < 8; i++ {
		history = append(history, makeToolGroup(string(rune('a'+i)), "bash", strings.Repeat("x", 200))...)
	}
	strategy := &SummarizationCompaction{Provider: nil, KeepRecent: 4}
	_, modified := strategy.Compact(context.Background(), history, 100)
	if modified {
		t.Error("should not modify with nil provider")
	}
}
