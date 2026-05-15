package skills

import (
	"context"
	"errors"
	"testing"
)

func TestCheckDependencies_Satisfied(t *testing.T) {
	loader := NewLoader(testdataDir())
	meta := &SkillMetadata{DependsOn: []string{"code-review"}}
	if err := loader.CheckDependencies(context.Background(), meta); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckDependencies_Missing(t *testing.T) {
	loader := NewLoader(testdataDir())
	meta := &SkillMetadata{DependsOn: []string{"nonexistent-skill"}}
	err := loader.CheckDependencies(context.Background(), meta)
	if !errors.Is(err, ErrDependencyNotFound) {
		t.Fatalf("got %v, want ErrDependencyNotFound", err)
	}
}

func TestCheckDependencies_Conflict(t *testing.T) {
	loader := NewLoader(testdataDir())
	meta := &SkillMetadata{ConflictsWith: []string{"conflicting-skill"}}
	err := loader.CheckDependencies(context.Background(), meta)
	if !errors.Is(err, ErrConflictDetected) {
		t.Fatalf("got %v, want ErrConflictDetected", err)
	}
}

func TestCheckDependencies_Empty(t *testing.T) {
	loader := NewLoader(testdataDir())
	meta := &SkillMetadata{}
	if err := loader.CheckDependencies(context.Background(), meta); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckDependencies_HasDepsFixture(t *testing.T) {
	loader := NewLoader(testdataDir())
	// has-deps depends on valid-skill (missing) and conflicts with conflicting-skill (exists)
	// Test dependency failure first
	meta := &SkillMetadata{DependsOn: []string{"valid-skill"}}
	err := loader.CheckDependencies(context.Background(), meta)
	if !errors.Is(err, ErrDependencyNotFound) {
		t.Fatalf("got %v, want ErrDependencyNotFound for missing valid-skill", err)
	}
}
