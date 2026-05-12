package tool

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/axis-cli/axis/internal/skills"
)

func skillsTestdataDir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "..", "..", "skills", "testdata", "skills")
}

func TestLoadSkillTool_Name(t *testing.T) {
	tool := NewLoadSkillTool(skills.NewLoader(""))
	if tool.Name() != "load_skill" {
		t.Errorf("Name() = %q, want load_skill", tool.Name())
	}
}

func TestLoadSkillTool_Schema(t *testing.T) {
	tool := NewLoadSkillTool(skills.NewLoader(""))
	schema := tool.Schema()
	if schema.Name != "load_skill" {
		t.Errorf("schema name = %q", schema.Name)
	}
	if len(schema.Parameters) != 1 || schema.Parameters[0].Name != "name" {
		t.Error("expected single 'name' parameter")
	}
}

func TestLoadSkillTool_Execute_Valid(t *testing.T) {
	loader := skills.NewLoader(skillsTestdataDir())
	tool := NewLoadSkillTool(loader)
	result, err := tool.Execute(context.Background(), map[string]any{"name": "pdf"})
	if err != nil {
		t.Fatal(err)
	}
	if result["error"] != nil {
		t.Fatalf("got error: %v", result["error"])
	}
	if result["name"] != "pdf" {
		t.Errorf("name = %v", result["name"])
	}
	if result["content"] == nil || result["content"] == "" {
		t.Error("content is empty")
	}
}

func TestLoadSkillTool_Execute_MissingName(t *testing.T) {
	tool := NewLoadSkillTool(skills.NewLoader(""))
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result["error"] == nil {
		t.Error("expected error for missing name")
	}
}

func TestLoadSkillTool_Execute_NotFound(t *testing.T) {
	loader := skills.NewLoader(skillsTestdataDir())
	tool := NewLoadSkillTool(loader)
	result, err := tool.Execute(context.Background(), map[string]any{"name": "nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	if result["error"] == nil {
		t.Error("expected error for nonexistent skill")
	}
}
