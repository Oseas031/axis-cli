package contextpack

import (
	"path/filepath"
	"testing"
)

func TestIndexStore_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	chunks := []DocumentChunk{
		{Source: "a.md", Content: "hello world", ModTime: 123, DocType: "doc"},
	}
	idx := &TFIDFIndex{}
	idx.Build(chunks)

	store := &IndexStore{}
	if err := store.Save(path, idx); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(loaded.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(loaded.Chunks))
	}
	if loaded.Chunks[0].Source != "a.md" {
		t.Errorf("expected a.md, got %s", loaded.Chunks[0].Source)
	}
	if len(loaded.IDF) == 0 {
		t.Errorf("expected non-empty IDF")
	}
	if len(loaded.Vectors) != 1 {
		t.Errorf("expected 1 vector, got %d", len(loaded.Vectors))
	}
}

func TestIndexStore_LoadMissing(t *testing.T) {
	store := &IndexStore{}
	_, err := store.Load("/nonexistent/path/index.json")
	if err == nil {
		t.Errorf("expected error for missing file")
	}
}
