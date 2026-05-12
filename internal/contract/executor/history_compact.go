package executor

import (
	"context"
	"fmt"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/types"
)

// EstimateTokens estimates token count using 4-chars-per-token heuristic.
func EstimateTokens(messages []types.ModelMessage) int {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content)/4 + 4
		for _, tc := range msg.ToolCalls {
			total += len(tc.Name)/4 + 10
		}
	}
	return total
}

// CompactionStrategy compacts message history to reduce token count.
type CompactionStrategy interface {
	Compact(ctx context.Context, history []types.ModelMessage, budget int) ([]types.ModelMessage, bool)
}

// CompactionPipeline runs strategies in order until budget is satisfied.
type CompactionPipeline struct {
	Strategies []CompactionStrategy
	Budget     int
}

// Compact runs all strategies in order, stopping early if budget is met.
func (p *CompactionPipeline) Compact(ctx context.Context, history []types.ModelMessage) []types.ModelMessage {
	for _, s := range p.Strategies {
		if EstimateTokens(history) <= p.Budget {
			break
		}
		history, _ = s.Compact(ctx, history, p.Budget)
	}
	return history
}

// ToolResultCompaction replaces old tool result content with short summaries.
type ToolResultCompaction struct {
	KeepRecent int
}

func (t *ToolResultCompaction) Compact(ctx context.Context, history []types.ModelMessage, budget int) ([]types.ModelMessage, bool) {
	if EstimateTokens(history) <= budget {
		return history, false
	}
	keepRecent := t.KeepRecent
	if keepRecent <= 0 {
		keepRecent = 3
	}
	type toolGroup struct{ start, end int }
	var groups []toolGroup
	for i := 0; i < len(history); i++ {
		if history[i].Role == "assistant" && len(history[i].ToolCalls) > 0 {
			start := i
			j := i + 1
			for j < len(history) && history[j].Role == "tool" {
				j++
			}
			groups = append(groups, toolGroup{start, j - 1})
			i = j - 1
		}
	}
	if len(groups) <= keepRecent {
		return history, false
	}
	modified := false
	for gi := 0; gi < len(groups)-keepRecent; gi++ {
		g := groups[gi]
		for i := g.start + 1; i <= g.end; i++ {
			if history[i].Role == "tool" && len(history[i].Content) > 100 {
				toolName := "unknown"
				for _, tc := range history[g.start].ToolCalls {
					if tc.ID == history[i].ToolCallID {
						toolName = tc.Name
						break
					}
				}
				history[i].Content = fmt.Sprintf("[Tool: %s completed]", toolName)
				modified = true
			}
		}
	}
	return history, modified
}

// TruncationCompaction drops oldest message groups until under budget.
type TruncationCompaction struct {
	KeepRecent int
}

func (t *TruncationCompaction) Compact(ctx context.Context, history []types.ModelMessage, budget int) ([]types.ModelMessage, bool) {
	if EstimateTokens(history) <= budget {
		return history, false
	}
	keepRecent := t.KeepRecent
	if keepRecent <= 0 {
		keepRecent = 2
	}
	type msgGroup struct{ start, end int }
	var groups []msgGroup
	for i := 0; i < len(history); i++ {
		if history[i].Role == "assistant" && len(history[i].ToolCalls) > 0 {
			start := i
			j := i + 1
			for j < len(history) && history[j].Role == "tool" {
				j++
			}
			groups = append(groups, msgGroup{start, j - 1})
			i = j - 1
		} else {
			groups = append(groups, msgGroup{i, i})
		}
	}
	if len(groups) <= keepRecent {
		return history, false
	}
	for len(groups) > keepRecent && EstimateTokens(history) > budget {
		g := groups[0]
		removeCount := g.end - g.start + 1
		history = history[removeCount:]
		groups = groups[1:]
		for i := range groups {
			groups[i].start -= removeCount
			groups[i].end -= removeCount
		}
	}
	return history, true
}

// SummarizationCompaction uses a ModelProvider to summarize older messages.
type SummarizationCompaction struct {
	Provider   provider.ModelProvider
	KeepRecent int
}

func (s *SummarizationCompaction) Compact(ctx context.Context, history []types.ModelMessage, budget int) ([]types.ModelMessage, bool) {
	if EstimateTokens(history) <= budget || s.Provider == nil {
		return history, false
	}
	keepRecent := s.KeepRecent
	if keepRecent <= 0 {
		keepRecent = 4
	}
	type msgGroup struct{ start, end int }
	var groups []msgGroup
	for i := 0; i < len(history); i++ {
		if history[i].Role == "assistant" && len(history[i].ToolCalls) > 0 {
			start := i
			j := i + 1
			for j < len(history) && history[j].Role == "tool" {
				j++
			}
			groups = append(groups, msgGroup{start, j - 1})
			i = j - 1
		} else {
			groups = append(groups, msgGroup{i, i})
		}
	}
	if len(groups) <= keepRecent {
		return history, false
	}
	splitIdx := groups[len(groups)-keepRecent].start
	var summaryInput string
	for _, msg := range history[:splitIdx] {
		if msg.Content != "" {
			summaryInput += fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content)
		}
	}
	resp, err := s.Provider.Execute(ctx, &provider.ModelRequest{
		Input: map[string]any{"message": "Summarize concisely, preserving key decisions and context:\n\n" + summaryInput},
	})
	if err != nil {
		return history, false
	}
	summaryText := "[Session summary]"
	if text, ok := resp.Output["text"].(string); ok && text != "" {
		summaryText = fmt.Sprintf("[Session summary: %s]", text)
	} else if msg, ok := resp.Output["message"].(string); ok && msg != "" {
		summaryText = fmt.Sprintf("[Session summary: %s]", msg)
	}
	result := make([]types.ModelMessage, 0, 1+len(history)-splitIdx)
	result = append(result, types.ModelMessage{Role: "user", Content: summaryText})
	result = append(result, history[splitIdx:]...)
	return result, true
}
