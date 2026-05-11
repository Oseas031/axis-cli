package contextpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentScanner(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("Axis CLI documentation"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignored"), 0644)

	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("git config"), 0644)

	scanner := &DocumentScanner{Root: dir}
	chunks, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	sources := make(map[string]bool)
	for _, c := range chunks {
		sources[c.Source] = true
	}
	if !sources["readme.md"] {
		t.Errorf("expected readme.md")
	}
	if !sources["main.go"] {
		t.Errorf("expected main.go")
	}
}

func TestTFIDFIndex(t *testing.T) {
	chunks := []DocumentChunk{
		{Source: "a.md", Content: "go programming language tutorial for beginners"},
		{Source: "b.md", Content: "python programming language guide for experts"},
		{Source: "c.md", Content: "go go go golang programming"},
	}

	idx := &TFIDFIndex{}
	idx.Build(chunks)

	results := idx.Query("go programming", 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Chunk.Source != "c.md" && results[0].Chunk.Source != "a.md" {
		t.Errorf("expected first result to be a.md or c.md, got %s", results[0].Chunk.Source)
	}

	unrelated := idx.Query("quantum physics", 2)
	if len(unrelated) != 2 {
		t.Fatalf("expected 2 results, got %d", len(unrelated))
	}
	if unrelated[0].Score > 0.5 {
		t.Errorf("expected low score for unrelated query, got %f", unrelated[0].Score)
	}
}

func TestTFIDFIndex_Empty(t *testing.T) {
	idx := &TFIDFIndex{}
	idx.Build(nil)
	results := idx.Query("anything", 5)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty index, got %d", len(results))
	}
}

func TestIndexManager(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doc.md"), []byte("Axis CLI design document"), 0644)

	mgr := NewIndexManager()
	status, err := mgr.Rebuild(dir)
	if err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}
	if !status.Healthy {
		t.Errorf("expected healthy status")
	}
	if status.IndexedFiles != 1 {
		t.Errorf("expected 1 file, got %d", status.IndexedFiles)
	}
	if status.TotalChunks != 1 {
		t.Errorf("expected 1 chunk, got %d", status.TotalChunks)
	}

	path := IndexPath(dir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected index file to exist at %s", path)
	}

	st2 := mgr.Status(dir)
	if !st2.Healthy {
		t.Errorf("expected loaded status to be healthy")
	}

	os.WriteFile(filepath.Join(dir, "code.go"), []byte("package main\n"), 0644)
	status, err = mgr.Update(dir)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if status.IndexedFiles != 2 {
		t.Errorf("expected 2 files after update, got %d", status.IndexedFiles)
	}
}

func TestIndexManager_MissingIndexStatus(t *testing.T) {
	dir := t.TempDir()
	mgr := NewIndexManager()
	status := mgr.Status(dir)
	if status.Healthy {
		t.Errorf("expected unhealthy for missing index")
	}
	if status.Message != "index not found" {
		t.Errorf("expected 'index not found' message, got %q", status.Message)
	}
}
