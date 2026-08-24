package types

import "testing"

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"ascii_short", "hi", 0},    // 2 chars /4 = 0 (floor)
		{"ascii", "hello world", 2}, // 11 chars → ASCII=11, /4=2
		{"cjk", "你好世界", 4},          // 4 CJK chars, each accumulates 4 → 16/4=4
		{"mixed", "hello你好", 3},     // 5 ASCII(5) + 2 CJK(8) = 13/4=3
		{"single_char_a", "a", 0},   // 1/4=0
		{"four_ascii", "abcd", 1},   // 4/4=1
		{"punctuation", "!@#$", 1},  // 4/4=1
		{"newline", "a\nb\nc", 1},   // 3 printable(3) + 2 newline(4) = 7/4=1
		{"long_ascii", "the quick brown fox jumps over the lazy dog", 10}, // 43 chars → 43/4=10
		{"long_cjk", "测试一下中文分词的估算效果怎么样呢", 17},                             // 17 CJK → 68/4=17
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.input)
			if got != tt.expected {
				t.Errorf("EstimateTokens(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEstimateMessages(t *testing.T) {
	msgs := []ModelMessage{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world", ToolCalls: []ToolCall{{ID: "1", Name: "grep"}}},
	}

	got := EstimateMessages(msgs)
	// msg1: "hello" → 5/4=1, +4 overhead = 5
	// msg2: "world" → 5/4=1, +4 overhead = 5, + tool "grep" → 4/4=1 + 10 = 11
	// total = 5 + 16 = 21
	expected := 21
	if got != expected {
		t.Errorf("EstimateMessages() = %d, want %d", got, expected)
	}
}

func TestEstimateString(t *testing.T) {
	if got := EstimateString("test"); got != EstimateTokens("test") {
		t.Errorf("EstimateString should match EstimateTokens")
	}
}

func TestIsCJK(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'你', true},
		{'中', true},
		{'A', false},
		{'0', false},
		{' ', false},
		{0x3000, false}, // CJK Symbols (not ideograph)
	}
	for _, tt := range tests {
		if got := isCJK(tt.r); got != tt.want {
			t.Errorf("isCJK(%U) = %v, want %v", tt.r, got, tt.want)
		}
	}
}
