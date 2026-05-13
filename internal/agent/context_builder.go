// Package agent provides self-context management for agent autonomy.
package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/types"
)

// ContextBuilder builds SelfContext by collecting task state, code structure, and documentation.
type ContextBuilder struct {
	scheduler  SchedulerProvider
	stateStore sharedlayer.StateStore
	rootDir    string
}

// NewContextBuilder creates a new ContextBuilder.
func NewContextBuilder(scheduler SchedulerProvider, stateStore sharedlayer.StateStore, rootDir string) *ContextBuilder {
	return &ContextBuilder{
		scheduler:  scheduler,
		stateStore: stateStore,
		rootDir:    rootDir,
	}
}

// BuildSelfContext builds a complete SelfContext for the given task ID.
func (cb *ContextBuilder) BuildSelfContext(taskID string) (*SelfContext, error) {
	ctx := NewSelfContext(taskID)

	// Collect task lineage by traversing dependencies
	if err := cb.collectTaskLineage(ctx, taskID); err != nil {
		return nil, fmt.Errorf("failed to collect task lineage: %w", err)
	}

	// Collect code snapshot
	codeSnapshot, err := cb.collectCodeSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to collect code snapshot: %w", err)
	}
	ctx.CodeSnapshot = codeSnapshot

	// Collect documentation snapshot
	docSnapshot, err := cb.collectDocSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to collect doc snapshot: %w", err)
	}
	ctx.DocSnapshot = docSnapshot

	// Collect state snapshot
	stateSnapshot, err := cb.collectStateSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to collect state snapshot: %w", err)
	}
	ctx.StateSnapshot = stateSnapshot

	return ctx, nil
}

// collectTaskLineage traverses the dependency graph to build task lineage.
func (cb *ContextBuilder) collectTaskLineage(ctx *SelfContext, taskID string) error {
	visited := make(map[string]bool)
	return cb.traverseDependencies(ctx, taskID, visited)
}

// traverseDependencies recursively traverses dependencies to build lineage.
func (cb *ContextBuilder) traverseDependencies(ctx *SelfContext, taskID string, visited map[string]bool) error {
	if visited[taskID] {
		return nil
	}
	visited[taskID] = true

	// Load the task from state store to get its dependencies
	state, err := cb.stateStore.Load(taskID)
	if err != nil {
		// Task not found in state store, try scheduler
		tasks := cb.scheduler.GetAllTasks()
		for _, t := range tasks {
			if t.TaskID == taskID {
				for _, depID := range t.Dependencies {
					if err := cb.traverseDependencies(ctx, depID, visited); err != nil {
						return err
					}
					ctx.AddLineage(depID)
				}
				return nil
			}
		}
		// Task not found anywhere, skip
		return nil
	}

	if state.Task != nil {
		for _, depID := range state.Task.Dependencies {
			if err := cb.traverseDependencies(ctx, depID, visited); err != nil {
				return err
			}
			ctx.AddLineage(depID)
		}
	}
	return nil
}

// collectCodeSnapshot collects information about the codebase.
func (cb *ContextBuilder) collectCodeSnapshot() (*CodeSnapshot, error) {
	snapshot := NewCodeSnapshot()

	// Find modified Go files in internal/
	if err := filepath.Walk(filepath.Join(cb.rootDir, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // Skip inaccessible files
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			snapshot.ModifiedFiles = append(snapshot.ModifiedFiles, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Count tasks from scheduler
	tasks := cb.scheduler.GetAllTasks()
	snapshot.TaskCount = len(tasks)

	// Read spec version from docs/current-progress.md if exists
	specVersion := cb.readSpecVersion()
	snapshot.SpecVersion = specVersion

	// Estimate tool count based on code files
	snapshot.ToolCount = len(snapshot.ModifiedFiles)

	return snapshot, nil
}

// readSpecVersion reads the current spec version.
func (cb *ContextBuilder) readSpecVersion() string {
	specPath := filepath.Join(cb.rootDir, "docs", "current-progress.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		return "unknown"
	}
	content := string(data)
	// Try to find version in the file
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "version") || strings.Contains(strings.ToLower(line), "milestone") {
			return strings.TrimSpace(line)
		}
	}
	return "unknown"
}

// collectDocSnapshot collects documentation files.
func (cb *ContextBuilder) collectDocSnapshot() (*DocSnapshot, error) {
	snapshot := NewDocSnapshot()

	// Walk docs directory
	docsDir := filepath.Join(cb.rootDir, "docs")
	if err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(cb.rootDir, path)
		if err != nil {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".md" {
			if strings.HasPrefix(relPath, "docs/specs") {
				snapshot.SpecFiles = append(snapshot.SpecFiles, relPath)
			} else if strings.HasPrefix(relPath, "docs") {
				snapshot.DocFiles = append(snapshot.DocFiles, relPath)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Collect reports
	reportsDir := filepath.Join(cb.rootDir, "reports")
	if err := filepath.Walk(reportsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, err := filepath.Rel(cb.rootDir, path)
			if err != nil {
				return nil
			}
			snapshot.Reports = append(snapshot.Reports, relPath)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return snapshot, nil
}

// collectStateSnapshot collects current task execution state.
func (cb *ContextBuilder) collectStateSnapshot() (*StateSnapshot, error) {
	snapshot := NewStateSnapshot()

	tasks := cb.scheduler.GetAllTasks()
	for _, task := range tasks {
		switch task.Status {
		case types.TaskStatusRunning:
			snapshot.RunningTasks++
		case types.TaskStatusPending:
			snapshot.PendingTasks++
		case types.TaskStatusCompleted:
			snapshot.CompletedTasks++
		case types.TaskStatusFailed:
			snapshot.FailedTasks++
		}
	}

	return snapshot, nil
}
