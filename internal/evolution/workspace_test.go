package evolution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewWorkspace_CreatesDirectory(t *testing.T) {
	runDir := t.TempDir()
	ws, err := NewWorkspace(runDir, "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.RunID != "run-1" {
		t.Errorf("expected run ID run-1, got %s", ws.RunID)
	}
	if _, err := os.Stat(ws.BasePath); os.IsNotExist(err) {
		t.Fatal("expected workspace directory to exist")
	}
}

func TestNewWorkspace_ExistingNotOverwritten(t *testing.T) {
	runDir := t.TempDir()
	// Create workspace with a marker file
	wsPath := filepath.Join(runDir, "workspace")
	if err := os.MkdirAll(wsPath, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	marker := filepath.Join(wsPath, "marker.txt")
	if err := os.WriteFile(marker, []byte("existing"), 0644); err != nil {
		t.Fatalf("write marker failed: %v", err)
	}

	// NewWorkspace should return existing workspace without overwriting
	ws, err := NewWorkspace(runDir, "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ws.Exists("marker.txt") {
		t.Fatal("expected existing marker file to survive")
	}
}

func TestWorkspace_WriteAndReadFile(t *testing.T) {
	runDir := t.TempDir()
	ws, err := NewWorkspace(runDir, "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []byte("hello workspace")
	if err := ws.WriteFile("test.txt", data); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	read, err := ws.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}
	if string(read) != string(data) {
		t.Errorf("expected %s, got %s", string(data), string(read))
	}
}

func TestWorkspace_CopyFrom(t *testing.T) {
	// Main project tree
	mainDir := t.TempDir()
	srcPath := filepath.Join(mainDir, "src", "main.go")
	if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(srcPath, []byte("package main"), 0644); err != nil {
		t.Fatalf("write src failed: %v", err)
	}

	// Workspace
	runDir := t.TempDir()
	ws, err := NewWorkspace(runDir, "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := ws.CopyFrom(srcPath, "src/main.go"); err != nil {
		t.Fatalf("copy from failed: %v", err)
	}

	copied, err := ws.ReadFile("src/main.go")
	if err != nil {
		t.Fatalf("read copied file failed: %v", err)
	}
	if string(copied) != "package main" {
		t.Errorf("expected 'package main', got %s", string(copied))
	}
}

func TestWorkspace_CopyFrom_DoesNotModifyMain(t *testing.T) {
	mainDir := t.TempDir()
	srcPath := filepath.Join(mainDir, "main.go")
	if err := os.WriteFile(srcPath, []byte("original"), 0644); err != nil {
		t.Fatalf("write src failed: %v", err)
	}

	runDir := t.TempDir()
	ws, err := NewWorkspace(runDir, "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ws.CopyFrom(srcPath, "main.go"); err != nil {
		t.Fatalf("copy from failed: %v", err)
	}

	// Modify workspace copy
	if err := ws.WriteFile("main.go", []byte("modified")); err != nil {
		t.Fatalf("write workspace file failed: %v", err)
	}

	// Main tree should be unchanged
	mainContent, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read main file failed: %v", err)
	}
	if string(mainContent) != "original" {
		t.Errorf("main tree was modified: %s", string(mainContent))
	}
}

func TestWorkspace_ListFiles(t *testing.T) {
	runDir := t.TempDir()
	ws, err := NewWorkspace(runDir, "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws.WriteFile("a.txt", []byte("a"))
	ws.WriteFile("sub/b.txt", []byte("b"))

	files, err := ws.ListFiles()
	if err != nil {
		t.Fatalf("list files failed: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestWorkspace_PromoteTo(t *testing.T) {
	runDir := t.TempDir()
	ws, err := NewWorkspace(runDir, "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws.WriteFile("new.txt", []byte("new content"))
	ws.WriteFile("sub/updated.txt", []byte("updated content"))

	targetDir := t.TempDir()
	if err := ws.PromoteTo(targetDir); err != nil {
		t.Fatalf("promote failed: %v", err)
	}

	// Verify promoted files
	content, err := os.ReadFile(filepath.Join(targetDir, "new.txt"))
	if err != nil {
		t.Fatalf("read promoted file failed: %v", err)
	}
	if string(content) != "new content" {
		t.Errorf("expected 'new content', got %s", string(content))
	}

	content2, err := os.ReadFile(filepath.Join(targetDir, "sub/updated.txt"))
	if err != nil {
		t.Fatalf("read promoted nested file failed: %v", err)
	}
	if string(content2) != "updated content" {
		t.Errorf("expected 'updated content', got %s", string(content2))
	}
}
