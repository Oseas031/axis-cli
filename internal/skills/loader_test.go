package skills

import (
	"context"
	"strings"
	"testing"
)

func TestLoadValidSkill(t *testing.T) {
	loader := NewLoader(testdataDir())
	skill, err := loader.Load(context.Background(), "pdf")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if skill.Meta.Name != "pdf" {
		t.Errorf("name = %q, want pdf", skill.Meta.Name)
	}
	if skill.Meta.Description == "" {
		t.Error("description is empty")
	}
	if !strings.Contains(skill.Content, "# PDF Processing Skill") {
		t.Error("content missing expected heading")
	}
	if skill.LoadedAt.IsZero() {
		t.Error("LoadedAt not set")
	}
}

func TestLoadContentExcludesFrontmatter(t *testing.T) {
	loader := NewLoader(testdataDir())
	skill, err := loader.Load(context.Background(), "pdf")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(skill.Content, "---") {
		t.Error("content should not contain frontmatter delimiters")
	}
}

func TestLoadNotFound(t *testing.T) {
	loader := NewLoader(testdataDir())
	_, err := loader.Load(context.Background(), "nonexistent")
	if err != ErrSkillNotFound {
		t.Errorf("got %v, want ErrSkillNotFound", err)
	}
}

func TestLoadPathEscape(t *testing.T) {
	loader := NewLoader(testdataDir())
	attacks := []string{
		"../escape",
		"..\\escape",
		"../../etc",
		"./hidden",
		"valid/../escape",
	}
	for _, name := range attacks {
		_, err := loader.Load(context.Background(), name)
		if err == nil {
			t.Errorf("expected error for %q", name)
		}
	}
}

func TestSafeSkillPath(t *testing.T) {
	base := testdataDir()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "pdf", false},
		{"dotdot", "../x", true},
		{"backslash escape", "..\\x", true},
		{"absolute", "/etc/passwd", true},
		{"slash in name", "a/b", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := safeSkillPath(base, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("safeSkillPath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
