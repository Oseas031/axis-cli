package tool

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListDirAndGrepSearch(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "app")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "main.py"), []byte("def hello_axis():\n    return 'ok'\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ld := NewListDirTool([]string{dir})
	res, err := ld.Execute(context.Background(), map[string]any{"path": dir})
	if err != nil {
		t.Fatalf("list_dir error: %v", err)
	}
	entries, _ := res["entries"].([]string)
	if len(entries) < 2 {
		t.Fatalf("expected >= 2 entries, got %v", entries)
	}

	gs := NewGrepSearchTool([]string{dir})
	gres, err := gs.Execute(context.Background(), map[string]any{"path": dir, "pattern": "hello_axis"})
	if err != nil {
		t.Fatalf("grep_search error: %v", err)
	}
	matches, _ := gres["matches"].([]string)
	if len(matches) != 1 || !strings.Contains(matches[0], "main.py:1:") {
		t.Fatalf("unexpected grep matches: %v", matches)
	}

	// Out-of-bounds path is rejected
	bad, _ := ld.Execute(context.Background(), map[string]any{"path": filepath.Dir(dir)})
	if errMsg, _ := bad["error"].(string); !strings.Contains(errMsg, "path validation failed") {
		t.Fatalf("expected path validation error, got %v", bad)
	}
}
