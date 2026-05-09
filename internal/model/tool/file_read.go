package tool

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

// FileReadTool reads files from allowed directories.
type FileReadTool struct {
	allowedDirs []string
}

// NewFileReadTool creates a new FileReadTool with the specified allowed directories.
func NewFileReadTool(allowedDirs []string) *FileReadTool {
	return &FileReadTool{allowedDirs: allowedDirs}
}

// Name returns the tool name.
func (t *FileReadTool) Name() string {
	return "file_read"
}

// Schema returns the tool definition for file_read.
func (t *FileReadTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "file_read",
		Description: "Read contents of a file from allowed directories",
		Parameters: []types.FieldDef{
			{Name: "path", Type: types.FieldTypeString, Required: true, Description: "Path to the file to read"},
		},
	}
}

// validatePath checks if the path is within allowed directories.
func (t *FileReadTool) validatePath(requestedPath string) error {
	cleanPath := filepath.Clean(requestedPath)

	for _, dir := range t.allowedDirs {
		cleanDir := filepath.Clean(dir)
		// Check if the clean path starts with the allowed directory
		if strings.HasPrefix(cleanPath, cleanDir) {
			return nil
		}
	}
	return &PathValidationError{Path: requestedPath, Reason: "path is not in allowed directories"}
}

// Execute reads and returns the file contents.
func (t *FileReadTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return map[string]any{"error": "path is required and must be a string"}, nil
	}

	if err := t.validatePath(path); err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	return map[string]any{"content": string(content)}, nil
}

// PathValidationError represents a path validation failure.
type PathValidationError struct {
	Path   string
	Reason string
}

func (e *PathValidationError) Error() string {
	return "path validation failed: " + e.Reason + ": " + e.Path
}
