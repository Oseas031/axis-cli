package types

// EstimateTokens returns a unified token count estimate for text.
//
// Language-aware heuristic:
//   - CJK characters: ~1 token per rune
//   - ASCII (printable): ~1 token per 4 runes
//   - Other runes: ~1 token per 2 runes
//
// This replaces the three scattered implementations:
//   - executor.EstimateTokens: len/4
//   - immediate.EstimateTokens: CJK-aware with /4
//   - compactor.estimateTokens: CJK/1.5 + other/4
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}

	tokens := 0
	for _, r := range text {
		switch {
		case isCJK(r):
			tokens += 4 // CJK: 1 rune ≈ 1 token; accumulate 4 so /4 yields 1
		case r >= ' ' && r <= '~': // ASCII printable
			tokens += 1
		default:
			tokens += 2 // other Unicode: 1 rune ≈ 0.5 token; accumulate 2 so /4 yields 0.5
		}
	}
	return tokens / 4
}

// EstimateMessages estimates the total token count for a list of ModelMessages,
// including per-message overhead for role/tool metadata.
func EstimateMessages(messages []ModelMessage) int {
	total := 0
	for _, msg := range messages {
		total += EstimateTokens(msg.Content)
		total += 4 // per-message overhead (role, separators)
		for _, tc := range msg.ToolCalls {
			total += EstimateTokens(tc.Name) + 10
		}
	}
	return total
}

// isCJK reports whether r is a CJK unified ideograph or extension.
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x20000 && r <= 0x2A6DF) || // CJK Extension B
		(r >= 0x2A700 && r <= 0x2B73F) || // CJK Extension C
		(r >= 0x2B740 && r <= 0x2B81F) || // CJK Extension D
		(r >= 0x2B820 && r <= 0x2CEAF) || // CJK Extension E
		(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
		(r >= 0x2F800 && r <= 0x2FA1F) // CJK Compatibility Supplement
}

// EstimateString is an alias for EstimateTokens, for call-sites that prefer
// a longer, more explicit name.
func EstimateString(text string) int {
	return EstimateTokens(text)
}
