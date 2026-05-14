package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── parseSkillFrontmatter tests ──────────────────────────────────────────────

func TestParseSkillFrontmatter_NilCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"empty string", ""},
		{"no frontmatter delimiter", "just some content\nwithout frontmatter"},
		{"only opening delimiter no close", "---\nname: test\n"},
		{"opening delimiter with content but no newline+close", "---\nname: test"},
		{"close delimiter without preceding newline", "---name: test---"},
		{"empty frontmatter block (---\\n---)", "---\n---\nBody"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSkillFrontmatter(tt.content)
			if result != nil {
				t.Errorf("expected nil, got %+v", result)
			}
		})
	}
}

func TestParseSkillFrontmatter_FourDashesStillMatches(t *testing.T) {
	// HasPrefix("---") matches "----" — documents actual behavior
	content := "---- \nname: x\n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("4+ dashes still matches HasPrefix(\"---\") — expected non-nil")
	}
	if meta.Name != "x" {
		t.Errorf("name = %q, want %q", meta.Name, "x")
	}
}

func TestParseSkillFrontmatter_ValidUnix(t *testing.T) {
	content := "---\nname: my-skill\ndescription: A test skill\nversion: 1.0.0\ntags: [workflow, test]\ndepends_on: [base]\nconflicts_with: [other]\n---\nBody content here"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "my-skill" {
		t.Errorf("name = %q, want %q", meta.Name, "my-skill")
	}
	if meta.Description != "A test skill" {
		t.Errorf("description = %q, want %q", meta.Description, "A test skill")
	}
	if meta.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", meta.Version, "1.0.0")
	}
	if len(meta.Tags) != 2 || meta.Tags[0] != "workflow" || meta.Tags[1] != "test" {
		t.Errorf("tags = %v, want [workflow test]", meta.Tags)
	}
	if len(meta.DependsOn) != 1 || meta.DependsOn[0] != "base" {
		t.Errorf("depends_on = %v, want [base]", meta.DependsOn)
	}
	if len(meta.ConflictsWith) != 1 || meta.ConflictsWith[0] != "other" {
		t.Errorf("conflicts_with = %v, want [other]", meta.ConflictsWith)
	}
}

func TestParseSkillFrontmatter_ValidWindows(t *testing.T) {
	content := "---\r\nname: win-skill\r\ndescription: Windows line endings\r\nversion: 2.0\r\n---\r\nBody"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "win-skill" {
		t.Errorf("name = %q, want %q", meta.Name, "win-skill")
	}
	if meta.Description != "Windows line endings" {
		t.Errorf("description = %q, want %q", meta.Description, "Windows line endings")
	}
	if meta.Version != "2.0" {
		t.Errorf("version = %q, want %q", meta.Version, "2.0")
	}
}

func TestParseSkillFrontmatter_PartialFields(t *testing.T) {
	content := "---\nname: minimal\n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "minimal" {
		t.Errorf("name = %q, want %q", meta.Name, "minimal")
	}
	if meta.Description != "" {
		t.Errorf("description should be empty, got %q", meta.Description)
	}
	if meta.Tags != nil {
		t.Errorf("tags should be nil, got %v", meta.Tags)
	}
}

func TestParseSkillFrontmatter_EmptyValues(t *testing.T) {
	content := "---\nname:\ndescription:\ntags:\n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "" {
		t.Errorf("name should be empty, got %q", meta.Name)
	}
	if meta.Description != "" {
		t.Errorf("description should be empty, got %q", meta.Description)
	}
	// tags with empty value → parseFrontmatterList("") → nil
	if meta.Tags != nil {
		t.Errorf("tags should be nil, got %v", meta.Tags)
	}
}

func TestParseSkillFrontmatter_ColonInValue(t *testing.T) {
	// Real-world case: description contains colons (e.g., URLs, time formats)
	content := "---\nname: web-tool\ndescription: Fetches data from http://example.com:8080/api\n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	// strings.Index finds first colon, so value = everything after first ":"
	// "description: Fetches data from http://example.com:8080/api"
	// colonIdx points to first ":" after "description"
	// val = "Fetches data from http://example.com:8080/api"
	expected := "Fetches data from http://example.com:8080/api"
	if meta.Description != expected {
		t.Errorf("description = %q, want %q", meta.Description, expected)
	}
}

func TestParseSkillFrontmatter_NoColonLines(t *testing.T) {
	// Lines without colons should be silently skipped
	content := "---\nname: test\nthis line has no colon\n  \nanother no-colon line\ndescription: works\n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "test" {
		t.Errorf("name = %q, want %q", meta.Name, "test")
	}
	if meta.Description != "works" {
		t.Errorf("description = %q, want %q", meta.Description, "works")
	}
}

func TestParseSkillFrontmatter_UnknownKeys(t *testing.T) {
	content := "---\nname: test\nauthor: someone\ncustom_field: value\n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "test" {
		t.Errorf("name = %q, want %q", meta.Name, "test")
	}
	// Unknown keys are silently ignored — no crash
}

func TestParseSkillFrontmatter_UnicodeContent(t *testing.T) {
	content := "---\nname: 中文技能\ndescription: 这是一个中文描述 🚀\ntags: [标签一, 标签二]\n---\n正文内容"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "中文技能" {
		t.Errorf("name = %q, want %q", meta.Name, "中文技能")
	}
	if meta.Description != "这是一个中文描述 🚀" {
		t.Errorf("description = %q", meta.Description)
	}
	if len(meta.Tags) != 2 || meta.Tags[0] != "标签一" {
		t.Errorf("tags = %v", meta.Tags)
	}
}

func TestParseSkillFrontmatter_AdversarialInjection(t *testing.T) {
	// Attempt to inject extra delimiters in content
	content := "---\nname: legit\n---\nBody\n---\nname: injected\n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	// Should only parse the first frontmatter block
	if meta.Name != "legit" {
		t.Errorf("name = %q, want %q (injection should not work)", meta.Name, "legit")
	}
}

func TestParseSkillFrontmatter_WhitespaceAroundKeys(t *testing.T) {
	content := "---\n  name  :  spaced  \n  description  :  also spaced  \n---\n"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Name != "spaced" {
		t.Errorf("name = %q, want %q", meta.Name, "spaced")
	}
	if meta.Description != "also spaced" {
		t.Errorf("description = %q, want %q", meta.Description, "also spaced")
	}
}

func TestParseSkillFrontmatter_MinimalContentBetweenDelimiters(t *testing.T) {
	// At least one line (even blank) between delimiters → valid empty meta
	content := "---\n \n---\nBody"
	meta := parseSkillFrontmatter(content)
	if meta == nil {
		t.Fatal("expected non-nil meta for frontmatter with whitespace line")
	}
	// The line " " has no colon, so all fields remain empty
	if meta.Name != "" {
		t.Errorf("name should be empty, got %q", meta.Name)
	}
}

// ── parseFrontmatterList tests ───────────────────────────────────────────────

func TestParseFrontmatterList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", nil},
		{"whitespace only", "   ", nil},
		{"single item no brackets", "workflow", []string{"workflow"}},
		{"multiple items no brackets", "a, b, c", []string{"a", "b", "c"}},
		{"bracketed list", "[workflow, test]", []string{"workflow", "test"}},
		{"bracketed single", "[solo]", []string{"solo"}},
		{"bracketed empty", "[]", nil},
		{"consecutive commas", "a,,b,,,c", []string{"a", "b", "c"}},
		{"trailing comma", "a, b,", []string{"a", "b"}},
		{"leading comma", ",a, b", []string{"a", "b"}},
		{"spaces in items", "  foo bar , baz qux  ", []string{"foo bar", "baz qux"}},
		{"only opening bracket", "[a, b", []string{"[a", "b"}},
		{"only closing bracket", "a, b]", []string{"a", "b]"}},
		{"nested brackets", "[[a, b], c]", []string{"[a", "b]", "c"}},
		{"unicode items", "[研究, 自动化]", []string{"研究", "自动化"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFrontmatterList(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("len = %d, want %d; got %v", len(result), len(tt.expected), result)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

// ── serveSkillsList integration tests ────────────────────────────────────────

func TestServeSkillsList_NonexistentDir(t *testing.T) {
	w := httptest.NewRecorder()
	serveSkillsList(w, filepath.Join(t.TempDir(), "nonexistent"))
	assertJSONArray(t, w, 0)
}

func TestServeSkillsList_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	assertJSONArray(t, w, 0)
}

func TestServeSkillsList_DirWithoutSkillMd(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "some-skill"), 0o755)
	// No SKILL.md inside
	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	assertJSONArray(t, w, 0)
}

func TestServeSkillsList_SkillWithFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: custom-name\ndescription: A great skill\ntags: [ai, tool]\nversion: 1.2.3\n---\nContent"), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	skills := assertJSONArray(t, w, 1)

	s := skills[0]
	if s["name"] != "custom-name" {
		t.Errorf("name = %v, want custom-name", s["name"])
	}
	if s["description"] != "A great skill" {
		t.Errorf("description = %v", s["description"])
	}
	if s["version"] != "1.2.3" {
		t.Errorf("version = %v", s["version"])
	}
}

func TestServeSkillsList_SkillWithoutFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "fallback-name")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("Just plain content, no frontmatter"), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	skills := assertJSONArray(t, w, 1)

	// Should fall back to directory name
	if skills[0]["name"] != "fallback-name" {
		t.Errorf("name = %v, want fallback-name", skills[0]["name"])
	}
}

func TestServeSkillsList_EmptySkillMd(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "empty-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(""), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	skills := assertJSONArray(t, w, 1)

	if skills[0]["name"] != "empty-skill" {
		t.Errorf("name = %v, want empty-skill", skills[0]["name"])
	}
}

func TestServeSkillsList_FrontmatterNameOverridesDir(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "dir-name")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: override-name\n---\n"), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	skills := assertJSONArray(t, w, 1)

	if skills[0]["name"] != "override-name" {
		t.Errorf("name = %v, want override-name (frontmatter should override dir name)", skills[0]["name"])
	}
}

func TestServeSkillsList_FrontmatterEmptyNameKeepsDirName(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "keep-dir")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname:\ndescription: has desc\n---\n"), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	skills := assertJSONArray(t, w, 1)

	// Empty name in frontmatter → keep directory name
	if skills[0]["name"] != "keep-dir" {
		t.Errorf("name = %v, want keep-dir (empty frontmatter name should not override)", skills[0]["name"])
	}
}

func TestServeSkillsList_MultipleSkills(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		skillDir := filepath.Join(dir, name)
		os.MkdirAll(skillDir, 0o755)
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: "+name+"\n---\n"), 0o644)
	}

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	assertJSONArray(t, w, 3)
}

func TestServeSkillsList_NestedDirsOnlyScansOneLevel(t *testing.T) {
	dir := t.TempDir()
	// Create nested: dir/parent/child/SKILL.md — child should NOT be discovered
	parent := filepath.Join(dir, "parent")
	child := filepath.Join(parent, "child")
	os.MkdirAll(child, 0o755)
	os.WriteFile(filepath.Join(parent, "SKILL.md"), []byte("---\nname: parent\n---\n"), 0o644)
	os.WriteFile(filepath.Join(child, "SKILL.md"), []byte("---\nname: child\n---\n"), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	skills := assertJSONArray(t, w, 1)

	// Only parent should be found (SkipDir prevents descending)
	if skills[0]["name"] != "parent" {
		t.Errorf("name = %v, want parent", skills[0]["name"])
	}
}

func TestServeSkillsList_ResponseHeaders(t *testing.T) {
	dir := t.TempDir()
	w := httptest.NewRecorder()
	serveSkillsList(w, dir)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestServeSkillsList_OmitsEmptyFields(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "minimal")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: minimal\n---\n"), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)

	// Verify omitempty works: response should NOT contain "description", "tags", etc.
	body := w.Body.String()
	if strings.Contains(body, `"description"`) {
		t.Error("empty description should be omitted from JSON")
	}
	if strings.Contains(body, `"tags"`) {
		t.Error("nil tags should be omitted from JSON")
	}
	if strings.Contains(body, `"version"`) {
		t.Error("empty version should be omitted from JSON")
	}
}

func TestServeSkillsList_RealWorldSkillContent(t *testing.T) {
	// Simulate a real SKILL.md like the ones in .axis/skills/
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "vigil")
	os.MkdirAll(skillDir, 0o755)
	content := `---
name: vigil
description: Cross-session work tracking. Use at session start (resume) and when managing work items.
tags: [workflow, tracking, automation]
version: 1.0.0
---
# Vigil Skill

This skill provides cross-session work tracking capabilities.

## Commands

- ` + "`axis vigil resume`" + ` — Resume work from last session
- ` + "`axis vigil add`" + ` — Add new work item
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)

	w := httptest.NewRecorder()
	serveSkillsList(w, dir)
	skills := assertJSONArray(t, w, 1)

	s := skills[0]
	if s["name"] != "vigil" {
		t.Errorf("name = %v", s["name"])
	}
	if s["description"] != "Cross-session work tracking. Use at session start (resume) and when managing work items." {
		t.Errorf("description = %v", s["description"])
	}
	if s["version"] != "1.0.0" {
		t.Errorf("version = %v", s["version"])
	}
	tags, ok := s["tags"].([]any)
	if !ok || len(tags) != 3 {
		t.Errorf("tags = %v, want 3 items", s["tags"])
	}
}

// ── Test helpers ─────────────────────────────────────────────────────────────

func assertJSONArray(t *testing.T, w *httptest.ResponseRecorder, expectedLen int) []map[string]any {
	t.Helper()
	var arr []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &arr); err != nil {
		t.Fatalf("failed to parse response as JSON array: %v\nbody: %s", err, w.Body.String())
	}
	if len(arr) != expectedLen {
		t.Fatalf("array length = %d, want %d; body: %s", len(arr), expectedLen, w.Body.String())
	}
	return arr
}
