package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidate_ValidSkill(t *testing.T) {
	loader := NewLoader(testdataDir())
	if err := loader.Validate(context.Background(), "pdf"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidate_MissingSKILLMD(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "empty-skill"), 0o755)
	loader := NewLoader(dir)
	err := loader.Validate(context.Background(), "empty-skill")
	if err != ErrMissingSKILLMD {
		t.Errorf("got %v, want ErrMissingSKILLMD", err)
	}
}

func TestValidate_InvalidFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "bad-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("no frontmatter here"), 0o644)
	loader := NewLoader(dir)
	err := loader.Validate(context.Background(), "bad-skill")
	if err == nil {
		t.Error("expected error for invalid frontmatter")
	}
}

func TestValidate_NameMismatch(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(skillDir, 0o755)
	content := "---\nname: other-name\ndescription: A skill\n---\n# Content"
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
	loader := NewLoader(dir)
	err := loader.Validate(context.Background(), "my-skill")
	if err != ErrNameMismatch {
		t.Errorf("got %v, want ErrNameMismatch", err)
	}
}

func TestValidate_ScriptsAsFile(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(skillDir, 0o755)
	content := "---\nname: my-skill\ndescription: A skill\n---\n# Content"
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
	// Create scripts as a file instead of directory
	os.WriteFile(filepath.Join(skillDir, "scripts"), []byte("not a dir"), 0o644)
	loader := NewLoader(dir)
	err := loader.Validate(context.Background(), "my-skill")
	if err == nil {
		t.Error("expected error when scripts is a file")
	}
}
