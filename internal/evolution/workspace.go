// Package evolution provides data models and storage for the Sandboxed Evolution Protocol.
package evolution

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Workspace manages an isolated filesystem area for an evolution run.
type Workspace struct {
	RunID    string
	BasePath string
}

// NewWorkspace creates or returns an existing workspace for a run.
// It refuses to overwrite an existing workspace.
func NewWorkspace(runDir string, runID string) (*Workspace, error) {
	wsPath := filepath.Join(runDir, "workspace")
	if info, err := os.Stat(wsPath); err == nil && info.IsDir() {
		return &Workspace{RunID: runID, BasePath: wsPath}, nil
	}
	if err := os.MkdirAll(wsPath, 0755); err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	return &Workspace{RunID: runID, BasePath: wsPath}, nil
}

// Path returns the full path for a relative path inside the workspace.
func (w *Workspace) Path(rel string) string {
	return filepath.Join(w.BasePath, rel)
}

// CopyFrom copies a file from the main project tree into the workspace.
// It preserves the relative directory structure.
//
// DESIGN NOTE: This is an internal API; path validation is the caller's
// responsibility. The Sandboxed Evolution Protocol assumes the caller
// (e.g. BootstrapOrchestrator) only copies from the legitimate project tree.
func (w *Workspace) CopyFrom(srcPath string, relPath string) error {
	dstPath := w.Path(relPath)
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create dst dir: %w", err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}

// WriteFile writes data to a file inside the workspace.
func (w *Workspace) WriteFile(relPath string, data []byte) error {
	path := w.Path(relPath)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ReadFile reads a file from inside the workspace.
func (w *Workspace) ReadFile(relPath string) ([]byte, error) {
	return os.ReadFile(w.Path(relPath))
}

// Exists checks whether a relative path exists inside the workspace.
func (w *Workspace) Exists(relPath string) bool {
	_, err := os.Stat(w.Path(relPath))
	return err == nil
}

// ListFiles returns all relative file paths inside the workspace recursively.
func (w *Workspace) ListFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(w.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(w.BasePath, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk workspace: %w", err)
	}
	return files, nil
}

// PromoteTo copies all workspace files back to the given target root (typically the project root).
// It does not delete files from the workspace.
func (w *Workspace) PromoteTo(targetRoot string) error {
	files, err := w.ListFiles()
	if err != nil {
		return err
	}
	for _, rel := range files {
		srcPath := w.Path(rel)
		dstPath := filepath.Join(targetRoot, rel)
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("create target dir: %w", err)
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("read workspace file: %w", err)
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return fmt.Errorf("write target file: %w", err)
		}
	}
	return nil
}
