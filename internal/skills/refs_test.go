package skills

import (
	"testing"
)

func TestExtractRefs_BasicExtraction(t *testing.T) {
	content := "hand off to /diagnose for debugging\nthen use /improve-codebase-architecture"
	refs := ExtractRefs(content)
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
	if refs[0].Name != "diagnose" || refs[0].Line != 1 {
		t.Errorf("refs[0] = %+v, want {diagnose, 1}", refs[0])
	}
	if refs[1].Name != "improve-codebase-architecture" || refs[1].Line != 2 {
		t.Errorf("refs[1] = %+v, want {improve-codebase-architecture, 2}", refs[1])
	}
}

func TestExtractRefs_InvalidFiltered(t *testing.T) {
	content := "/A uppercase\n/123 starts with number\n/a single char"
	refs := ExtractRefs(content)
	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d: %+v", len(refs), refs)
	}
}

func TestExtractRefs_Deduplicated(t *testing.T) {
	content := "use /diagnose here\nand /diagnose again"
	refs := ExtractRefs(content)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if refs[0].Line != 1 {
		t.Errorf("expected first occurrence line 1, got %d", refs[0].Line)
	}
}

func TestExtractRefs_NoRefs(t *testing.T) {
	content := "no skill references here\njust plain text"
	refs := ExtractRefs(content)
	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}

func TestExtractRefs_InsideCodeBlocks(t *testing.T) {
	content := "```\n/diagnose inside code\n```"
	refs := ExtractRefs(content)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if refs[0].Name != "diagnose" {
		t.Errorf("expected diagnose, got %s", refs[0].Name)
	}
}

func TestExtractRefs_URLsNotMatched(t *testing.T) {
	content := "visit https://example.com/path and http://foo.bar/something"
	refs := ExtractRefs(content)
	if len(refs) != 0 {
		t.Errorf("expected 0 refs from URLs, got %d: %+v", len(refs), refs)
	}
}
