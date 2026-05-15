package skills

import (
	"context"
	"testing"
)

func TestLoad_PopulatesSubFiles(t *testing.T) {
	loader := NewLoader("testdata/skills")
	skill, err := loader.Load(context.Background(), "with-subfiles")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(skill.SubFiles) != 1 {
		t.Fatalf("expected 1 sub-file, got %d", len(skill.SubFiles))
	}
	sf := skill.SubFiles[0]
	if sf.Name != "REFERENCE.md" {
		t.Errorf("expected sub-file name REFERENCE.md, got %s", sf.Name)
	}
	if sf.Size == 0 {
		t.Error("expected non-zero sub-file size")
	}
}

func TestLoadSubFile_ReturnsContent(t *testing.T) {
	loader := NewLoader("testdata/skills")
	content, err := loader.LoadSubFile(context.Background(), "with-subfiles", "REFERENCE.md")
	if err != nil {
		t.Fatalf("LoadSubFile() error: %v", err)
	}
	if content == "" {
		t.Fatal("expected non-empty content")
	}
	if content != "# Reference\nDetailed reference content for testing.\n" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestLoadSubFile_RejectsPathTraversal(t *testing.T) {
	loader := NewLoader("testdata/skills")
	cases := []string{"../foo.md", "..\\foo.md", "sub/file.md", "sub\\file.md"}
	for _, name := range cases {
		_, err := loader.LoadSubFile(context.Background(), "with-subfiles", name)
		if err != ErrInvalidPath {
			t.Errorf("LoadSubFile(%q) expected ErrInvalidPath, got %v", name, err)
		}
	}
}

func TestLoadSubFile_NonExistentFile(t *testing.T) {
	loader := NewLoader("testdata/skills")
	_, err := loader.LoadSubFile(context.Background(), "with-subfiles", "NONEXISTENT.md")
	if err != ErrSkillNotFound {
		t.Errorf("expected ErrSkillNotFound, got %v", err)
	}
}
