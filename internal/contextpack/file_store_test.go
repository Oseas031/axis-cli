package contextpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore_LoadAll_MissingReturnsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(tmpDir)
	records, err := store.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll on missing file should succeed: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected empty records, got %d", len(records))
	}
}

func TestFileStore_SaveAllAndLoadAll(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(tmpDir)

	records := map[string]ReadinessRecord{
		"ctx-abc": {
			Artifact: ReadinessArtifact{BundleID: "ctx-abc", TaskID: "t1", PacketCount: 2},
			Bundle:   ContextBundle{TaskID: "t1", Goal: "test goal"},
		},
	}
	if err := store.SaveAll(records); err != nil {
		t.Fatalf("SaveAll should succeed: %v", err)
	}

	loaded, err := store.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll should succeed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 record, got %d", len(loaded))
	}
	rec := loaded["ctx-abc"]
	if rec.Artifact.BundleID != "ctx-abc" {
		t.Fatalf("expected bundle id ctx-abc, got %s", rec.Artifact.BundleID)
	}
	if rec.Bundle.Goal != "test goal" {
		t.Fatalf("expected goal %q, got %q", "test goal", rec.Bundle.Goal)
	}
}

func TestFileStore_DeleteAll(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(tmpDir)
	if err := store.SaveAll(map[string]ReadinessRecord{"ctx-x": {}}); err != nil {
		t.Fatalf("SaveAll should succeed: %v", err)
	}
	if err := store.DeleteAll(); err != nil {
		t.Fatalf("DeleteAll should succeed: %v", err)
	}
	if _, err := os.Stat(store.Path()); !os.IsNotExist(err) {
		t.Fatal("expected store file to be deleted")
	}
}

func TestFileStore_DeleteAll_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(tmpDir)
	if err := store.DeleteAll(); err != nil {
		t.Fatalf("DeleteAll on missing file should succeed: %v", err)
	}
}

func TestFileStore_LoadAll_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(tmpDir)
	path := store.Path()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := store.LoadAll(); err == nil {
		t.Fatal("expected error loading invalid json")
	}
}

func TestFileStore_Path(t *testing.T) {
	store := NewFileStore("/project")
	want := filepath.Join("/project", ".axis", "context", "readiness.json")
	if got := store.Path(); got != want {
		t.Fatalf("expected path %q, got %q", want, got)
	}
}

func TestFileStore_DefaultRoot(t *testing.T) {
	store := NewFileStore("")
	if got := store.Path(); got != filepath.Join(".", ".axis", "context", "readiness.json") {
		t.Fatalf("expected default root path, got %q", got)
	}
}
