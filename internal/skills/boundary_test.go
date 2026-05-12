package skills

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func projectRoot() string {
	_, f, _, _ := runtime.Caller(0)
	// internal/skills/boundary_test.go -> project root is ../../
	return filepath.Join(filepath.Dir(f), "..", "..")
}

func TestBoundarySchedulerDoesNotImportSkills(t *testing.T) {
	root := projectRoot()
	cmd := exec.Command("go", "list", "-deps", "./internal/kernel/scheduler/...")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("go list failed: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "internal/skills") {
			t.Fatalf("scheduler depends on skills: %s", line)
		}
	}
}

func TestBoundaryLoadSkillIsOptIn(t *testing.T) {
	loader := NewLoader(testdataDir())
	// Discover only returns metadata, not content
	metas, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range metas {
		// SkillMeta has no Content field — this is compile-time proof
		// that discovery never leaks content
		if m.Name == "" {
			t.Error("meta has empty name")
		}
	}
	// BuildSkillsPromptSection also must not contain full content
	section := loader.BuildSkillsPromptSection(context.Background())
	if strings.Contains(section, "# PDF Processing Skill") {
		t.Error("prompt section leaks full skill content")
	}
}

func TestBoundarySkillPathSafety(t *testing.T) {
	loader := NewLoader(testdataDir())
	attacks := []string{"../etc", "..\\windows", "./hidden", "/absolute"}
	for _, name := range attacks {
		_, err := loader.Load(context.Background(), name)
		if err == nil {
			t.Errorf("expected error for path %q", name)
		}
	}
}

func TestBoundarySkillNameFormat(t *testing.T) {
	invalid := []string{"", "A", "has space", "under_score", "-leading", "trailing-", "a"}
	for _, name := range invalid {
		if err := ValidateSkillName(name); err == nil {
			t.Errorf("expected error for %q", name)
		}
	}
	valid := []string{"ab", "pdf", "code-review", "my-tool2"}
	for _, name := range valid {
		if err := ValidateSkillName(name); err != nil {
			t.Errorf("unexpected error for %q: %v", name, err)
		}
	}
}
