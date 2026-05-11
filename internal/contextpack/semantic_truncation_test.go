package contextpack

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateAtSemanticBoundary_noTruncationNeeded(t *testing.T) {
	content := "short text"
	truncated, pos := truncateAtSemanticBoundary(content, 100)
	if truncated != content || pos != len(content) {
		t.Fatalf("expected no truncation for short content, got %q at %d", truncated, pos)
	}
}

func TestTruncateAtSemanticBoundary_paragraphBoundary(t *testing.T) {
	content := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
	truncated, pos := truncateAtSemanticBoundary(content, 28)
	// 28 bytes lands inside "Second paragraph."; should cut after "First paragraph.\n\n"
	if !strings.Contains(truncated, "First paragraph") {
		t.Fatalf("expected first paragraph, got %q", truncated)
	}
	if strings.Contains(truncated, "Second paragraph") {
		t.Fatalf("should not contain second paragraph, got %q", truncated)
	}
	if pos <= 0 || pos > 28 {
		t.Fatalf("unexpected truncation position %d", pos)
	}
}

func TestTruncateAtSemanticBoundary_sentenceBoundary(t *testing.T) {
	content := "Hello world. This is a test. Another sentence here."
	truncated, pos := truncateAtSemanticBoundary(content, 28)
	// 28 lands inside " Another"; should cut after "Hello world. " or "Hello world. This is a test. "
	if !strings.Contains(truncated, "Hello world") {
		t.Fatalf("expected at least first sentence, got %q", truncated)
	}
	if pos <= 0 || pos > 28 {
		t.Fatalf("unexpected truncation position %d", pos)
	}
}

func TestTruncateAtSemanticBoundary_lineBoundary(t *testing.T) {
	content := "line one\nline two\nline three\nline four"
	truncated, pos := truncateAtSemanticBoundary(content, 22)
	// 22 lands inside "line three"; should cut after "line two\n"
	if !strings.Contains(truncated, "line two") {
		t.Fatalf("expected at least line two, got %q", truncated)
	}
	if strings.Contains(truncated, "line three") {
		t.Fatalf("should not contain line three, got %q", truncated)
	}
	if pos <= 0 || pos > 22 {
		t.Fatalf("unexpected truncation position %d", pos)
	}
}

func TestTruncateAtSemanticBoundary_zeroBudget(t *testing.T) {
	truncated, pos := truncateAtSemanticBoundary("anything", 0)
	if truncated != "" || pos != 0 {
		t.Fatalf("expected empty for zero budget, got %q at %d", truncated, pos)
	}
}

func TestTruncateAtSemanticBoundary_UTF8Safe(t *testing.T) {
	content := "你好世界这是一段中文测试文本"
	// Each Chinese rune is 3 bytes. maxBytes=10 cuts inside a rune.
	truncated, pos := truncateAtSemanticBoundary(content, 10)
	if pos > 10 {
		t.Fatalf("position %d exceeds maxBytes 10", pos)
	}
	// Verify every rune in truncated is valid.
	for _, r := range truncated {
		_ = r
	}
	if !utf8.ValidString(truncated) {
		t.Fatalf("truncated string is not valid UTF-8: %q", truncated)
	}
}

func TestTruncateAtSemanticBoundary_preservesLeadingContent(t *testing.T) {
	content := "A. B. C. D. E. F. G."
	truncated, pos := truncateAtSemanticBoundary(content, 10)
	// Should cut after "A. " or similar, not produce empty.
	if len(truncated) == 0 {
		t.Fatalf("expected non-empty truncated content, got empty")
	}
	if pos == 0 {
		t.Fatalf("expected positive truncation position")
	}
}
