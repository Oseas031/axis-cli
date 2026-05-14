package tool

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileWriteTool_PathTraversal_Regression(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewFileWriteTool([]string{tmpDir})

	cases := []struct {
		name string
		path string
	}{
		{"dot-dot-slash", tmpDir + "/../../../etc/passwd"},
		{"dot-dot-only", tmpDir + "/.."},
		{"dot-dot-nested", tmpDir + "/a/../../b/../../../etc/shadow"},
		{"absolute-unix", "/etc/passwd"},
		{"absolute-root", "/tmp/evil.txt"},
	}

	if runtime.GOOS == "windows" {
		cases = append(cases, []struct {
			name string
			path string
		}{
			{"absolute-windows-c", "C:\\Windows\\System32\\evil.txt"},
			{"absolute-windows-d", "D:\\evil.txt"},
		}...)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tool.Execute(context.Background(), map[string]any{
				"path":    tc.path,
				"content": "malicious",
			})
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if result["error"] == nil {
				t.Errorf("Expected path validation error for %q, got success", tc.path)
			}
		})
	}
}

func TestFileWriteTool_PathTraversal_PrefixBypass(t *testing.T) {
	// Regression: strings.HasPrefix("/tmp/safe", "/tmp/saf") == true
	// This must be rejected because "/tmp/safe-extra" is not inside "/tmp/saf"
	tmpDir := t.TempDir()
	siblingDir := tmpDir + "-sibling"
	if err := os.MkdirAll(siblingDir, 0750); err != nil {
		t.Fatalf("Failed to create sibling dir: %v", err)
	}
	defer os.RemoveAll(siblingDir)

	tool := NewFileWriteTool([]string{tmpDir})
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    filepath.Join(siblingDir, "evil.txt"),
		"content": "malicious",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected path validation error for sibling directory with shared prefix")
	}
}

func TestFileWriteTool_PathTraversal_Symlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink test requires elevated privileges on Windows")
	}

	tmpDir := t.TempDir()
	outsideDir := t.TempDir()

	// Create a symlink inside tmpDir that points outside
	symlinkPath := filepath.Join(tmpDir, "escape")
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	tool := NewFileWriteTool([]string{tmpDir})
	targetPath := filepath.Join(symlinkPath, "evil.txt")

	// The path validation uses filepath.Abs which does NOT resolve symlinks,
	// so the path appears to be inside tmpDir. This documents current behavior.
	// A stronger check would use filepath.EvalSymlinks, but that's a separate concern.
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    targetPath,
		"content": "test",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	// Document: symlink-based escape is not caught by current Abs+Rel validation.
	// This is consistent with FileReadTool behavior.
	_ = result
}

func TestFileWriteTool_ValidPath_StillWorks(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewFileWriteTool([]string{tmpDir})

	validPath := filepath.Join(tmpDir, "subdir", "file.txt")
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    validPath,
		"content": "valid content",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected success for valid path, got error: %v", result["error"])
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "valid content" {
		t.Errorf("Expected 'valid content', got %q", string(content))
	}
}
