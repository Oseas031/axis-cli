package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const skillTemplate = "---\nname: %s\ndescription: A skill\ntags: test\nversion: 1.0.0\nauthor: test\n---\n# %s\n"

func createSkill(t *testing.T, dir, name string) {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("---\nname: " + name + "\ndescription: A skill\ntags: test\nversion: 1.0.0\nauthor: test\n---\n# " + name + "\n")
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCacheReturnsCachedWithinTTL(t *testing.T) {
	dir := t.TempDir()
	createSkill(t, dir, "alpha")

	loader := NewLoader(dir, WithCacheTTL(5*time.Second))
	m1, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(m1) != 1 {
		t.Fatalf("got %d, want 1", len(m1))
	}

	// Add a new skill on disk — should NOT appear (cache still valid).
	createSkill(t, dir, "beta")
	m2, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(m2) != 1 {
		t.Fatalf("got %d, want 1 (cached)", len(m2))
	}
}

func TestCacheReScansAfterTTLExpires(t *testing.T) {
	dir := t.TempDir()
	createSkill(t, dir, "alpha")

	loader := NewLoader(dir, WithCacheTTL(1*time.Millisecond))
	m1, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(m1) != 1 {
		t.Fatalf("got %d, want 1", len(m1))
	}

	// Wait for TTL to expire.
	time.Sleep(2 * time.Millisecond)

	// Add a new skill — should appear after re-scan.
	createSkill(t, dir, "beta")
	m2, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(m2) != 2 {
		t.Fatalf("got %d, want 2 (re-scanned)", len(m2))
	}
}

func TestInvalidateForcesReScan(t *testing.T) {
	dir := t.TempDir()
	createSkill(t, dir, "alpha")

	loader := NewLoader(dir, WithCacheTTL(5*time.Second))
	m1, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(m1) != 1 {
		t.Fatalf("got %d, want 1", len(m1))
	}

	// Add a new skill and invalidate.
	createSkill(t, dir, "beta")
	loader.Invalidate()

	m2, err := loader.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(m2) != 2 {
		t.Fatalf("got %d, want 2 (after invalidate)", len(m2))
	}
}
