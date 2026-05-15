package skills

import "testing"

func TestParseSkillMetadata(t *testing.T) {
	input := "---\nname: code-review\ndescription: Automated code review\ntags: dev, quality\ndepends_on: linter, formatter\nconflicts_with: legacy-review\nversion: 1.2.0\n---\n# Code Review\n\nBody content here."

	meta, body := ParseSkillMetadata(input)
	if meta == nil {
		t.Fatal("expected metadata, got nil")
	}
	if meta.Name != "code-review" {
		t.Errorf("Name = %q, want %q", meta.Name, "code-review")
	}
	if meta.Description != "Automated code review" {
		t.Errorf("Description = %q, want %q", meta.Description, "Automated code review")
	}
	if len(meta.Tags) != 2 || meta.Tags[0] != "dev" || meta.Tags[1] != "quality" {
		t.Errorf("Tags = %v, want [dev quality]", meta.Tags)
	}
	if len(meta.DependsOn) != 2 || meta.DependsOn[0] != "linter" || meta.DependsOn[1] != "formatter" {
		t.Errorf("DependsOn = %v, want [linter formatter]", meta.DependsOn)
	}
	if len(meta.ConflictsWith) != 1 || meta.ConflictsWith[0] != "legacy-review" {
		t.Errorf("ConflictsWith = %v, want [legacy-review]", meta.ConflictsWith)
	}
	if meta.Version != "1.2.0" {
		t.Errorf("Version = %q, want %q", meta.Version, "1.2.0")
	}
	if body != "# Code Review\n\nBody content here." {
		t.Errorf("Body = %q", body)
	}
}

func TestParseSkillMetadata_NoFrontmatter(t *testing.T) {
	input := "# Just markdown\n\nNo frontmatter."
	meta, body := ParseSkillMetadata(input)
	if meta != nil {
		t.Errorf("expected nil metadata, got %+v", meta)
	}
	if body != input {
		t.Errorf("body should equal input")
	}
}

func TestParseSkillMetadata_EmptyFields(t *testing.T) {
	input := "---\nname: minimal\ndescription: A minimal skill\n---\nBody only."

	meta, body := ParseSkillMetadata(input)
	if meta == nil {
		t.Fatal("expected metadata")
	}
	if meta.Name != "minimal" {
		t.Errorf("Name = %q", meta.Name)
	}
	if meta.DependsOn != nil {
		t.Errorf("DependsOn should be nil, got %v", meta.DependsOn)
	}
	if meta.ConflictsWith != nil {
		t.Errorf("ConflictsWith should be nil, got %v", meta.ConflictsWith)
	}
	if body != "Body only." {
		t.Errorf("Body = %q", body)
	}
}


func TestParseSkillMetadata_WithSource(t *testing.T) {
	input := "---\nname: tdd\ndescription: Test-driven development\nsource: mattpocock/skills\nsource_version: 2026-05-15\n---\n# TDD\n\nBody."

	meta, body := ParseSkillMetadata(input)
	if meta == nil {
		t.Fatal("expected metadata, got nil")
	}
	if meta.Source != "mattpocock/skills" {
		t.Errorf("Source = %q, want %q", meta.Source, "mattpocock/skills")
	}
	if meta.SourceVersion != "2026-05-15" {
		t.Errorf("SourceVersion = %q, want %q", meta.SourceVersion, "2026-05-15")
	}
	if body != "# TDD\n\nBody." {
		t.Errorf("Body = %q", body)
	}
}

func TestParseSkillMetadata_WithoutSource(t *testing.T) {
	input := "---\nname: custom\ndescription: A custom skill\nversion: 1.0.0\n---\nBody."

	meta, _ := ParseSkillMetadata(input)
	if meta == nil {
		t.Fatal("expected metadata, got nil")
	}
	if meta.Source != "" {
		t.Errorf("Source should be empty, got %q", meta.Source)
	}
	if meta.SourceVersion != "" {
		t.Errorf("SourceVersion should be empty, got %q", meta.SourceVersion)
	}
}
