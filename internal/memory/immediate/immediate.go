// Package immediate provides the Immediate Memory layer for Axis.
// It represents the situational context of a single Agent execution cycle.
package immediate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"unicode/utf8"
)

// ImmediateContext represents the complete situational context of a single
// Agent execution cycle. It is built fresh for every turn.
type ImmediateContext struct {
	TaskID      string
	Intent      string
	Contract    *ContractSnapshot
	WorkingSet  *WorkingSetSnapshot
	ToolResults []ToolResult
	TurnCount   int
	Budget      TokenBudget
}

// WorkingSetSnapshot is the subset of Working Memory injected into
// ImmediateContext, after budget and relevance filtering.
type WorkingSetSnapshot struct {
	Bundles []RetainedBundleSummary
}

// RetainedBundleSummary represents a file or context packet in the working set.
type RetainedBundleSummary struct {
	BundleID    string // optional; empty for standalone file entries
	Type        string // e.g. "spec", "code", "tool", "memory"
	Source      string // filepath.ToSlash() normalized path
	Summary     string // UTF-8 safe head truncation, max 1024 bytes (P0)
	ContentHash string // SHA-256 truncated to 128 bit (32 hex chars)
	FileChanged bool   // true if hash differs from last seen in .seen file
	PacketCount int
}

// ToolResult captures a tool execution from the current cycle.
type ToolResult struct {
	ToolName   string
	Input      map[string]any
	Output     map[string]any
	Success    bool
	DurationMs int64
}

// TokenBudget tracks how much context budget has been consumed.
// P0 uses rune-count estimation with language-weighted heuristics;
// no external tokenization library is required.
type TokenBudget struct {
	MaxTokens  int
	UsedTokens int
	Remaining  int
}

// ContractSnapshot captures the contract constraints applicable to this task.
type ContractSnapshot struct {
	ContractID    string
	RequiredTools []string
	Constraints   map[string]string
}

// summaryMaxBytes is the P0 fixed-length summary threshold.
const summaryMaxBytes = 1024

// EstimateTokens returns a language-aware token approximation.
//
//	ASCII letters/digits/punctuation: 1 rune ≈ 0.25 token
//	CJK unified ideographs:           1 rune ≈ 1.0 token
//	Other runes:                      1 rune ≈ 0.5 token
//
// This is deliberately conservative for safety margins.
func EstimateTokens(s string) int {
	tokens := 0
	for _, r := range s {
		switch {
		case r >= ' ' && r <= '~': // ASCII printable
			tokens += 1 // will be divided by 4 below
		case r >= '\u4e00' && r <= '\u9fff': // CJK
			tokens += 4 // 1.0 token per rune after division
		default:
			tokens += 2 // 0.5 token per rune after division
		}
	}
	return tokens / 4
}

// TruncateSummary returns the first up-to-1024 UTF-8 bytes of content,
// truncated at a valid character boundary.
func TruncateSummary(content string) string {
	b := []byte(content)
	if len(b) <= summaryMaxBytes {
		return content
	}
	truncated := b[:summaryMaxBytes]
	for len(truncated) > 0 {
		r, size := utf8.DecodeLastRune(truncated)
		if r == utf8.RuneError && size == 1 {
			// Truncated in the middle of a multi-byte rune; remove the stray byte.
			truncated = truncated[:len(truncated)-1]
			continue
		}
		break
	}
	return string(truncated)
}

// ContentHash computes the SHA-256 of content and returns the first 128 bits
// as a 32-character hex string.
func ContentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:16]) // 128 bits = 16 bytes = 32 hex chars
}

// NewTokenBudget creates a TokenBudget with the given max.
func NewTokenBudget(maxTokens int) TokenBudget {
	return TokenBudget{
		MaxTokens:  maxTokens,
		Remaining:  maxTokens,
		UsedTokens: 0,
	}
}

// Consume deducts tokens from the budget. Returns an error if over-budget.
func (tb *TokenBudget) Consume(tokens int) error {
	if tokens < 0 {
		return fmt.Errorf("immediate: cannot consume negative tokens")
	}
	tb.UsedTokens += tokens
	tb.Remaining = tb.MaxTokens - tb.UsedTokens
	if tb.Remaining < 0 {
		return fmt.Errorf("immediate: budget exceeded (max %d, used %d)", tb.MaxTokens, tb.UsedTokens)
	}
	return nil
}
