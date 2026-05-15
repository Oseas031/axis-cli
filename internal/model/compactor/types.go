// Package compactor implements context-aware history compaction for the multi-turn loop.
// It offloads tool results to external files and replaces them with compact summaries
// based on a replaceability score.
package compactor

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// OffloadEntry represents a single offloaded tool result stored in offload.jsonl.
type OffloadEntry struct {
	Version    int       `json:"v"`              // Schema version. Current: 1.
	Timestamp  time.Time `json:"timestamp"`
	ToolCallID string    `json:"tool_call_id"`
	ToolName   string    `json:"tool_name"`
	Summary    string    `json:"summary"`
	ResultRef  string    `json:"result_ref"`
	Score      int       `json:"score"` // 0-10: how well summary replaces original
	TokensOrig int       `json:"tokens_orig"`
}

// CompactConfig configures the tiered compaction strategy.
type CompactConfig struct {
	// DataDir is the root directory for offload storage (e.g. .axis/memory/offload/).
	DataDir string

	// TokenBudget is the max token estimate for history before compaction triggers.
	TokenBudget int

	// MildRatio triggers mild compaction when usage exceeds this fraction of budget.
	// Default: 0.5
	MildRatio float64

	// AggressiveRatio triggers aggressive compaction.
	// Default: 0.8
	AggressiveRatio float64

	// EmergencyRatio triggers emergency truncation.
	// Default: 0.95
	EmergencyRatio float64

	// Summarizer generates a summary and replaceability score for a tool result.
	// If nil, a heuristic summarizer is used (truncate to first 200 chars, score=5).
	Summarizer Summarizer
}

// Summarizer produces a summary and replaceability score for a tool result.
type Summarizer interface {
	Summarize(toolName string, content string) (summary string, score int)
}

// DefaultConfig returns a CompactConfig with sensible defaults.
func DefaultConfig(dataDir string) CompactConfig {
	return CompactConfig{
		DataDir:         dataDir,
		TokenBudget:     100000,
		MildRatio:       0.5,
		AggressiveRatio: 0.8,
		EmergencyRatio:  0.95,
	}
}

// refFilename generates a deterministic filename for a tool result ref.
func refFilename(toolCallID string, ts time.Time) string {
	h := sha256.Sum256([]byte(toolCallID + ts.Format(time.RFC3339Nano)))
	return fmt.Sprintf("%s-%s.md", ts.Format("20060102T150405"), hex.EncodeToString(h[:4]))
}
