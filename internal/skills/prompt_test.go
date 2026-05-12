package skills

import (
	"context"
	"strings"
	"testing"
)

func TestBuildSkillsPromptSection_WithFixtures(t *testing.T) {
	loader := NewLoader(testdataDir())
	section := loader.BuildSkillsPromptSection(context.Background())
	if section == "" {
		t.Fatal("expected non-empty section")
	}
	if !strings.Contains(section, "pdf") {
		t.Error("expected pdf in section")
	}
	if !strings.Contains(section, "code-review") {
		t.Error("expected code-review in section")
	}
	if !strings.Contains(section, "load_skill") {
		t.Error("expected load_skill instruction")
	}
}

func TestBuildSkillsPromptSection_Empty(t *testing.T) {
	loader := NewLoader(t.TempDir())
	section := loader.BuildSkillsPromptSection(context.Background())
	if section != "" {
		t.Errorf("expected empty, got %q", section)
	}
}

func TestBuildSkillsPromptSection_NoContent(t *testing.T) {
	loader := NewLoader(testdataDir())
	section := loader.BuildSkillsPromptSection(context.Background())
	// Section should NOT contain full skill content
	if strings.Contains(section, "# PDF Processing Skill") {
		t.Error("section should not contain full skill content")
	}
}
