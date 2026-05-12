// Package project provides project root resolution for .axis/ directory location.
package project

import (
	"os"
	"path/filepath"
)

const axisDir = ".axis"

// ResolveRoot returns the project root directory containing .axis/.
// It walks up from startDir looking for a .axis/ directory.
// If not found, falls back to startDir itself.
func ResolveRoot(startDir string) string {
	dir := startDir
	for {
		if info, err := os.Stat(filepath.Join(dir, axisDir)); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, fall back to startDir
			return startDir
		}
		dir = parent
	}
}

// MustResolveRoot resolves from cwd. Panics if cwd cannot be determined.
func MustResolveRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic("project: cannot determine working directory: " + err.Error())
	}
	return ResolveRoot(cwd)
}

// AxisDir returns the .axis/ directory path for the given root.
func AxisDir(root string) string {
	return filepath.Join(root, axisDir)
}

// SkillsDir returns the .axis/skills/ directory path.
func SkillsDir(root string) string {
	return filepath.Join(root, axisDir, "skills")
}

// MemoryDir returns the .axis/memory/ directory path.
func MemoryDir(root string) string {
	return filepath.Join(root, axisDir, "memory")
}

// EventsDir returns the .axis/events/ directory path.
func EventsDir(root string) string {
	return filepath.Join(root, axisDir, "events")
}
