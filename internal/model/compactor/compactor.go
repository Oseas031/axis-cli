package compactor

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

// OffloadCompactor implements multiturn.Compactor with tiered compression.
// It offloads tool results to disk and replaces them with summaries based on
// replaceability score when token budget thresholds are exceeded.
//
// Usage:
//
//	cfg := compactor.DefaultConfig(".axis/memory/offload")
//	c, err := compactor.New(cfg)
//	executor := agent.NewLLMAgentExecutor(provider, tools,
//	    agent.WithHistoryCompactor(c),
//	)
//
// Known limitations:
//   - isAlreadyOffloaded uses "[offloaded:" prefix detection (magic string)
//   - estimateHistoryTokens is called per-candidate in mildCompact (O(candidates × history))
//   - offloaded map is session-scoped; cross-session dedup relies on offload.jsonl
type OffloadCompactor struct {
	cfg   CompactConfig
	store *Store
	// offloaded tracks which tool_call_ids have already been offloaded this session.
	offloaded map[string]*OffloadEntry
}

// New creates an OffloadCompactor. Returns error if store init fails.
func New(cfg CompactConfig) (*OffloadCompactor, error) {
	if cfg.MildRatio <= 0 {
		cfg.MildRatio = 0.5
	}
	if cfg.AggressiveRatio <= 0 {
		cfg.AggressiveRatio = 0.8
	}
	if cfg.EmergencyRatio <= 0 {
		cfg.EmergencyRatio = 0.95
	}
	if cfg.TokenBudget <= 0 {
		cfg.TokenBudget = 100000
	}

	store, err := NewStore(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	return &OffloadCompactor{
		cfg:       cfg,
		store:     store,
		offloaded: make(map[string]*OffloadEntry),
	}, nil
}

// Compact implements multiturn.Compactor.
// Strategy:
//   - Estimate total tokens in history
//   - If below mild threshold: no-op
//   - Mild (>50%): offload + replace highest-score tool results with summaries
//   - Aggressive (>80%): drop oldest assistant+tool pairs (keep last N)
//   - Emergency (>95%): truncate to 60% budget
func (c *OffloadCompactor) Compact(_ context.Context, history []types.ModelMessage) []types.ModelMessage {
	totalTokens := c.estimateHistoryTokens(history)
	mildThreshold := int(float64(c.cfg.TokenBudget) * c.cfg.MildRatio)
	aggressiveThreshold := int(float64(c.cfg.TokenBudget) * c.cfg.AggressiveRatio)
	emergencyThreshold := int(float64(c.cfg.TokenBudget) * c.cfg.EmergencyRatio)

	if totalTokens < mildThreshold {
		return history
	}

	// Mild: replace tool results with summaries (highest score first)
	history = c.mildCompact(history, mildThreshold)
	totalTokens = c.estimateHistoryTokens(history)
	if totalTokens < aggressiveThreshold {
		return history
	}

	// Aggressive: drop oldest pairs
	history = c.aggressiveCompact(history, aggressiveThreshold)
	totalTokens = c.estimateHistoryTokens(history)
	if totalTokens < emergencyThreshold {
		return history
	}

	// Emergency: hard truncate keeping only recent messages
	return c.emergencyCompact(history)
}

// offloadPrefix is the marker prefix for offloaded tool messages.
const offloadPrefix = "[axis:offload:"

// mildCompact replaces tool result content with summaries for entries with
// the highest replaceability scores, until tokens drop below threshold.
func (c *OffloadCompactor) mildCompact(history []types.ModelMessage, target int) []types.ModelMessage {
	// Collect replaceable tool messages (not already replaced)
	type candidate struct {
		idx       int
		score     int
		entry     *OffloadEntry
		origToken int // token count of original content
	}
	var candidates []candidate

	for i, msg := range history {
		if msg.Role != "tool" || msg.ToolCallID == "" {
			continue
		}
		if isAlreadyOffloaded(msg.Content) {
			continue
		}
		// Offload to disk if not already done
		entry, ok := c.offloaded[msg.ToolCallID]
		if !ok {
			toolName := c.findToolName(history, msg.ToolCallID)
			summary, score := c.summarize(toolName, msg.Content)
			var err error
			entry, err = c.store.Offload(msg.ToolCallID, toolName, msg.Content, summary, score)
			if err != nil {
				continue // skip on write error
			}
			c.offloaded[msg.ToolCallID] = entry
		}
		candidates = append(candidates, candidate{
			idx: i, score: entry.Score, entry: entry,
			origToken: estimateTokens(msg.Content),
		})
	}

	// Sort by score descending (highest score = most replaceable)
	sort.Slice(candidates, func(a, b int) bool {
		return candidates[a].score > candidates[b].score
	})

	// Replace until under target, maintaining running token total
	runningTotal := c.estimateHistoryTokens(history)
	for _, cand := range candidates {
		if runningTotal < target {
			break
		}
		replacement := fmt.Sprintf("%s %s | ref=%s]", offloadPrefix, cand.entry.Summary, cand.entry.ResultRef)
		newTokens := estimateTokens(replacement)
		runningTotal -= cand.origToken
		runningTotal += newTokens
		history[cand.idx].Content = replacement
	}

	return history
}

// aggressiveCompact drops the oldest assistant+tool pairs, keeping the most recent.
func (c *OffloadCompactor) aggressiveCompact(history []types.ModelMessage, target int) []types.ModelMessage {
	// Find pair boundaries: each assistant message with ToolCalls + subsequent tool messages
	type pair struct {
		start, end int
		tokens     int
	}
	var pairs []pair
	for i := 0; i < len(history); i++ {
		if history[i].Role == "assistant" && len(history[i].ToolCalls) > 0 {
			start := i
			end := i + 1
			for end < len(history) && history[end].Role == "tool" {
				end++
			}
			tokens := 0
			for j := start; j < end; j++ {
				tokens += estimateTokens(history[j].Content)
			}
			pairs = append(pairs, pair{start: start, end: end, tokens: tokens})
			i = end - 1
		}
	}

	// Drop oldest pairs until under target
	totalTokens := c.estimateHistoryTokens(history)
	dropUntil := 0
	for _, p := range pairs {
		if totalTokens < target {
			break
		}
		totalTokens -= p.tokens
		dropUntil = p.end
	}

	if dropUntil > 0 && dropUntil < len(history) {
		return history[dropUntil:]
	}
	return history
}

// emergencyCompact keeps only the last 60% of token budget worth of messages.
// v1: hard truncation. TODO: inject summary of dropped content.
func (c *OffloadCompactor) emergencyCompact(history []types.ModelMessage) []types.ModelMessage {
	targetTokens := int(float64(c.cfg.TokenBudget) * 0.6)

	// Safeguard: always preserve at least the last 4 messages (2 turns)
	const minKeep = 4
	if len(history) <= minKeep {
		return history
	}

	// Walk from end, accumulate tokens
	accum := 0
	cutoff := len(history)
	for i := len(history) - 1; i >= 0; i-- {
		accum += estimateTokens(history[i].Content)
		if accum > targetTokens {
			cutoff = i + 1
			break
		}
	}

	// Never cut below minKeep
	maxCutoff := len(history) - minKeep
	if cutoff > maxCutoff {
		cutoff = maxCutoff
	}

	// Ensure we don't split a tool pair: if cutoff lands on a "tool" message,
	// advance to include the preceding assistant message's full pair.
	for cutoff < len(history) && history[cutoff].Role == "tool" {
		cutoff++
	}
	if cutoff >= len(history) {
		return history[len(history)-minKeep:]
	}
	return history[cutoff:]
}

func (c *OffloadCompactor) estimateHistoryTokens(history []types.ModelMessage) int {
	total := 0
	for _, msg := range history {
		total += estimateTokens(msg.Content)
		// Account for tool call definitions in assistant messages
		for _, tc := range msg.ToolCalls {
			total += estimateTokens(tc.Name) + estimateTokens(fmt.Sprintf("%v", tc.Input))
		}
	}
	return total
}

func (c *OffloadCompactor) summarize(toolName, content string) (string, int) {
	if c.cfg.Summarizer != nil {
		return c.cfg.Summarizer.Summarize(toolName, content)
	}
	return heuristicSummarize(toolName, content)
}

// heuristicSummarize is the fallback summarizer: keeps first line + last lines + error lines.
// v1: heuristic only. TODO: LLM-based summarizer with semantic scoring.
func heuristicSummarize(toolName, content string) (string, int) {
	lines := splitLines(content)
	tokens := estimateTokens(content)

	// Collect meaningful lines: first line, error/fail lines, last 3 lines
	var kept []string
	if len(lines) > 0 {
		kept = append(kept, lines[0])
	}
	// Error/fail lines (up to 5)
	errorCount := 0
	for i := 1; i < len(lines) && errorCount < 5; i++ {
		lower := strings.ToLower(lines[i])
		if strings.Contains(lower, "error") || strings.Contains(lower, "fail") ||
			strings.Contains(lower, "panic") || strings.Contains(lower, "fatal") {
			kept = append(kept, lines[i])
			errorCount++
		}
	}
	// Last 3 lines (if not already included)
	tailStart := len(lines) - 3
	if tailStart < 1 {
		tailStart = 1
	}
	for i := tailStart; i < len(lines); i++ {
		kept = append(kept, lines[i])
	}

	summary := toolName + ": " + strings.Join(kept, " | ")
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}

	// v1: score by length + error presence. TODO: reference-aware scoring.
	score := 5
	if tokens > 2000 {
		score = 8
	} else if tokens > 500 {
		score = 6
	} else if tokens < 100 {
		score = 3
	}
	if errorCount > 0 {
		score -= 2 // errors are less replaceable
		if score < 1 {
			score = 1
		}
	}
	return summary, score
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := strings.TrimSpace(s[start:i])
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		line := strings.TrimSpace(s[start:])
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// findToolName looks up the tool name from the assistant's ToolCalls by ID.
func (c *OffloadCompactor) findToolName(history []types.ModelMessage, toolCallID string) string {
	for i := len(history) - 1; i >= 0; i-- {
		for _, tc := range history[i].ToolCalls {
			if tc.ID == toolCallID {
				return tc.Name
			}
		}
	}
	return "unknown"
}

// isAlreadyOffloaded checks if a tool message content has been replaced.
func isAlreadyOffloaded(content string) bool {
	return len(content) >= len(offloadPrefix) && content[:len(offloadPrefix)] == offloadPrefix
}
