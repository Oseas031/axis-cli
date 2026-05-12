package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRoot_FindsAxisDir(t *testing.T) {
	tmp := t.TempDir()
	// Create .axis/ in tmp
	os.MkdirAll(filepath.Join(tmp, ".axis"), 0o755)
	// Start from a subdirectory
	sub := filepath.Join(tmp, "a", "b", "c")
	os.MkdirAll(sub, 0o755)

	got := ResolveRoot(sub)
	if got != tmp {
		t.Errorf("ResolveRoot(%q) = %q, want %q", sub, got, tmp)
	}
}

func TestResolveRoot_FallsBackToStartDir(t *testing.T) {
	tmp := t.TempDir()
	// No .axis/ anywhere
	sub := filepath.Join(tmp, "x")
	os.MkdirAll(sub, 0o755)

	got := ResolveRoot(sub)
	if got != sub {
		t.Errorf("ResolveRoot(%q) = %q, want fallback to startDir", sub, got)
	}
}

func TestResolveRoot_DirectMatch(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".axis"), 0o755)

	got := ResolveRoot(tmp)
	if got != tmp {
		t.Errorf("got %q, want %q", got, tmp)
	}
}

func TestAxisDir(t *testing.T) {
	got := AxisDir("/project")
	want := filepath.Join("/project", ".axis")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSkillsDir(t *testing.T) {
	got := SkillsDir("/project")
	want := filepath.Join("/project", ".axis", "skills")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMemoryDir(t *testing.T) {
	got := MemoryDir("/project")
	want := filepath.Join("/project", ".axis", "memory")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
