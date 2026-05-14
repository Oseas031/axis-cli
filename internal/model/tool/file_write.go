package tool

import (
	"context"
	"os"
	"path/filepath"

	"github.com/axis-cli/axis/internal/types"
)

// FileWriteTool writes files to allowed directories.
type FileWriteTool struct {
	allowedDirs []string
}

// NewFileWriteTool creates a new FileWriteTool with the specified allowed directories.
func NewFileWriteTool(allowedDirs []string) *FileWriteTool {
	return &FileWriteTool{allowedDirs: allowedDirs}
}

// Name returns the tool name.
func (t *FileWriteTool) Name() string {
	return "file_write"
}

// Schema returns the tool definition for file_write.
func (t *FileWriteTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "file_write",
		Description: "Write content to a file in allowed directories",
		Parameters: []types.FieldDef{
			{Name: "path", Type: types.FieldTypeString, Required: true, Description: "Path to the file to write"},
			{Name: "content", Type: types.FieldTypeString, Required: true, Description: "Content to write to the file"},
		},
	}
}

// validatePath checks if the path is within allowed directories.
func (t *FileWriteTool) validatePath(requestedPath string) error {
	return validateAllowedPath(requestedPath, t.allowedDirs)
}

// Execute writes content to the file.
func (t *FileWriteTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return map[string]any{"error": "path is required and must be a string"}, nil
	}

	content, ok := input["content"].(string)
	if !ok {
		return map[string]any{"error": "content is required and must be a string"}, nil
	}

	// Clean and resolve to absolute path BEFORE validation and use.
	// This ensures validation and execution operate on the same canonical path.
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return map[string]any{"error": "invalid path: " + err.Error()}, nil
	}

	if err := t.validatePath(cleanPath); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(parentDir, 0750); err != nil {
		return map[string]any{"error": "failed to create parent directory: " + err.Error()}, nil
	}

	if err := os.WriteFile(cleanPath, []byte(content), 0600); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	return map[string]any{"success": true, "path": cleanPath}, nil
}
