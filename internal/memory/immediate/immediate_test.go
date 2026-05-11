package immediate

import (
	"testing"
	"unicode/utf8"
)

func TestTruncateSummary_BelowThreshold(t *testing.T) {
	input := "short content"
	got := TruncateSummary(input)
	if got != input {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

func TestTruncateSummary_ExactASCII(t *testing.T) {
	input := make([]byte, summaryMaxBytes)
	for i := range input {
		input[i] = 'x'
	}
	got := TruncateSummary(string(input))
	if len(got) != summaryMaxBytes {
		t.Fatalf("expected %d bytes, got %d", summaryMaxBytes, len(got))
	}
}

func TestTruncateSummary_UTF8Boundary(t *testing.T) {
	// Build a string that exceeds 1024 bytes with multi-byte UTF-8 chars.
	// CJK chars are 3 bytes each. 400 chars = 1200 bytes.
	input := ""
	for i := 0; i < 400; i++ {
		input += "中"
	}
	got := TruncateSummary(input)
	if len(got) > summaryMaxBytes {
		t.Fatalf("truncated summary exceeds %d bytes: %d", summaryMaxBytes, len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatal("truncated summary is not valid UTF-8")
	}
}

func TestTruncateSummary_SplitMultiByte(t *testing.T) {
	// Create content where 1024th byte falls in the middle of a 3-byte CJK char.
	// We need exactly 1022 ASCII bytes + 1 CJK char (3 bytes) = 1025 bytes.
	input := string(make([]byte, 1022)) + "中"
	got := TruncateSummary(input)
	if len(got) > summaryMaxBytes {
		t.Fatalf("exceeds threshold: %d > %d", len(got), summaryMaxBytes)
	}
	if !utf8.ValidString(got) {
		t.Fatal("not valid UTF-8 after truncation")
	}
	// The CJK char should be dropped because it doesn't fit fully.
	if len(got) != 1022 {
		t.Fatalf("expected 1022 bytes (CJK dropped), got %d", len(got))
	}
}

func TestContentHash_Length(t *testing.T) {
	h := ContentHash([]byte("hello world"))
	if len(h) != 32 {
		t.Fatalf("expected 32 hex chars, got %d (%q)", len(h), h)
	}
}

func TestContentHash_Deterministic(t *testing.T) {
	b := []byte("test data")
	h1 := ContentHash(b)
	h2 := ContentHash(b)
	if h1 != h2 {
		t.Fatalf("hash not deterministic: %q vs %q", h1, h2)
	}
}

func TestEstimateTokens(t *testing.T) {
	// Pure ASCII: 40 printable chars = 40 runes -> 40/4 = 10 tokens.
	ascii := "abcdefghijklmnopqrstuvwxyz0123456789!@#$"
	if got := EstimateTokens(ascii); got != 10 {
		t.Fatalf("ASCII: expected 10 tokens, got %d", got)
	}

	// Pure CJK: 10 CJK chars = 10 runes -> 10*4/4 = 10 tokens.
	cjk := "中文字符测试十个字符"
	if got := EstimateTokens(cjk); got != 10 {
		t.Fatalf("CJK: expected 10 tokens, got %d", got)
	}
}

func TestTokenBudget_Consume(t *testing.T) {
	b := NewTokenBudget(100)
	if err := b.Consume(30); err != nil {
		t.Fatalf("Consume 30: %v", err)
	}
	if b.UsedTokens != 30 {
		t.Fatalf("expected UsedTokens 30, got %d", b.UsedTokens)
	}
	if b.Remaining != 70 {
		t.Fatalf("expected Remaining 70, got %d", b.Remaining)
	}
}

func TestTokenBudget_OverBudget(t *testing.T) {
	b := NewTokenBudget(10)
	if err := b.Consume(20); err == nil {
		t.Fatal("expected over-budget error")
	}
}
