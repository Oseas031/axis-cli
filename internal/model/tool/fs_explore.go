package tool

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

func resolveAllowedPath(rawPath string, allowedDirs []string) (string, error) {
	p := strings.TrimSpace(rawPath)
	if p == "" {
		p = "."
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if len(allowedDirs) == 0 {
		return abs, nil
	}
	for _, dir := range allowedDirs {
		dirAbs, derr := filepath.Abs(dir)
		if derr != nil {
			continue
		}
		if abs == dirAbs || strings.HasPrefix(strings.ToLower(abs), strings.ToLower(dirAbs+string(filepath.Separator))) {
			return abs, nil
		}
	}
	return "", fmt.Errorf("path validation failed: %s is not in allowed directories", abs)
}

func shouldSkipExploreDir(name string) bool {
	switch name {
	case ".git", "node_modules", "__pycache__", ".venv", "venv", "dist":
		return true
	}
	return false
}

// ListDirTool lists files and subdirectories inside an allowed directory.
type ListDirTool struct {
	allowedDirs []string
}

func NewListDirTool(allowedDirs []string) *ListDirTool {
	return &ListDirTool{allowedDirs: allowedDirs}
}

func (t *ListDirTool) Name() string { return "list_dir" }

func (t *ListDirTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "list_dir",
		Description: "List files and subdirectories in a directory within allowed paths.",
		Parameters: []types.FieldDef{
			{Name: "path", Type: types.FieldTypeString, Required: false, Description: "Directory path (default: current directory)"},
			{Name: "max_depth", Type: types.FieldTypeInt, Required: false, Description: "Traversal depth 1-4 (default: 2)"},
		},
	}
}

func (t *ListDirTool) Execute(_ context.Context, input map[string]any) (map[string]any, error) {
	rawPath, _ := input["path"].(string)
	target, err := resolveAllowedPath(rawPath, t.allowedDirs)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	maxDepth := 2
	if d, ok := input["max_depth"].(float64); ok && d >= 1 && d <= 4 {
		maxDepth = int(d)
	}

	var entries []string
	const maxEntries = 200
	baseDepth := strings.Count(target, string(filepath.Separator))
	_ = filepath.WalkDir(target, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if p == target {
			return nil
		}
		if d.IsDir() && shouldSkipExploreDir(d.Name()) {
			return filepath.SkipDir
		}
		depth := strings.Count(p, string(filepath.Separator)) - baseDepth
		if depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(target, p)
		kind := "FILE"
		if d.IsDir() {
			kind = "DIR"
		}
		entries = append(entries, fmt.Sprintf("[%s] %s", kind, filepath.ToSlash(rel)))
		if len(entries) >= maxEntries {
			return fs.SkipAll
		}
		return nil
	})

	return map[string]any{
		"path":    target,
		"count":   len(entries),
		"entries": entries,
	}, nil
}

// GrepSearchTool searches text files for a substring/pattern within allowed directories.
type GrepSearchTool struct {
	allowedDirs []string
}

func NewGrepSearchTool(allowedDirs []string) *GrepSearchTool {
	return &GrepSearchTool{allowedDirs: allowedDirs}
}

func (t *GrepSearchTool) Name() string { return "grep_search" }

func (t *GrepSearchTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "grep_search",
		Description: "Search for a text pattern inside files under an allowed directory.",
		Parameters: []types.FieldDef{
			{Name: "pattern", Type: types.FieldTypeString, Required: true, Description: "Substring to search for (case-insensitive)"},
			{Name: "path", Type: types.FieldTypeString, Required: false, Description: "Directory or file path (default: current directory)"},
		},
	}
}

func (t *GrepSearchTool) Execute(_ context.Context, input map[string]any) (map[string]any, error) {
	pattern, _ := input["pattern"].(string)
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return map[string]any{"error": "pattern is required"}, nil
	}
	rawPath, _ := input["path"].(string)
	target, err := resolveAllowedPath(rawPath, t.allowedDirs)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	lowerPat := strings.ToLower(pattern)
	var matches []string
	const maxMatches = 60

	_ = filepath.WalkDir(target, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipExploreDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil || info.Size() > 512*1024 {
			return nil
		}
		f, ferr := os.Open(p)
		if ferr != nil {
			return nil
		}
		defer f.Close()

		rel, _ := filepath.Rel(target, p)
		scanner := bufio.NewScanner(f)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), lowerPat) {
				trimmed := strings.TrimSpace(line)
				if len(trimmed) > 160 {
					trimmed = trimmed[:160] + "..."
				}
				matches = append(matches, fmt.Sprintf("%s:%d: %s", filepath.ToSlash(rel), lineNo, trimmed))
				if len(matches) >= maxMatches {
					return fs.SkipAll
				}
			}
		}
		return nil
	})

	return map[string]any{
		"pattern": pattern,
		"count":   len(matches),
		"matches": matches,
	}, nil
}
