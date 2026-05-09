package agent

import (
	"testing"

	"github.com/axis-cli/axis/internal/kernel/scheduler"
	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/types"
)

func TestNewContextBuilder(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/test/root")

	if cb == nil {
		t.Fatal("NewContextBuilder returned nil")
	}
	if cb.rootDir != "/test/root" {
		t.Errorf("expected rootDir '/test/root', got '%s'", cb.rootDir)
	}
	if cb.scheduler == nil {
		t.Error("scheduler should not be nil")
	}
	if cb.stateStore == nil {
		t.Error("stateStore should not be nil")
	}
}

func TestContextBuilder_collectStateSnapshot(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/test/root")

	// Submit some tasks (scheduler sets all to pending initially)
	tasks := []*types.AgentTask{
		{TaskID: "task-1", Status: types.TaskStatusPending},
		{TaskID: "task-2", Status: types.TaskStatusPending},
		{TaskID: "task-3", Status: types.TaskStatusPending},
		{TaskID: "task-4", Status: types.TaskStatusPending},
	}

	for _, task := range tasks {
		if err := sched.Submit(task); err != nil {
			t.Fatalf("failed to submit task: %v", err)
		}
	}

	// Update task statuses to reflect desired states
	if err := sched.UpdateTaskStatus("task-2", types.TaskStatusRunning); err != nil {
		t.Fatalf("failed to update task-2 status: %v", err)
	}
	if err := sched.UpdateTaskStatus("task-3", types.TaskStatusCompleted); err != nil {
		t.Fatalf("failed to update task-3 status: %v", err)
	}
	if err := sched.UpdateTaskStatus("task-4", types.TaskStatusFailed); err != nil {
		t.Fatalf("failed to update task-4 status: %v", err)
	}

	snapshot, err := cb.collectStateSnapshot()
	if err != nil {
		t.Fatalf("collectStateSnapshot failed: %v", err)
	}

	if snapshot.PendingTasks != 1 {
		t.Errorf("expected PendingTasks=1, got %d", snapshot.PendingTasks)
	}
	if snapshot.RunningTasks != 1 {
		t.Errorf("expected RunningTasks=1, got %d", snapshot.RunningTasks)
	}
	if snapshot.CompletedTasks != 1 {
		t.Errorf("expected CompletedTasks=1, got %d", snapshot.CompletedTasks)
	}
	if snapshot.FailedTasks != 1 {
		t.Errorf("expected FailedTasks=1, got %d", snapshot.FailedTasks)
	}
}

func TestContextBuilder_collectTaskLineage(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/test/root")

	// Create a dependency chain: task-3 -> task-2 -> task-1
	tasks := []*types.AgentTask{
		{TaskID: "task-1", Dependencies: []string{}},
		{TaskID: "task-2", Dependencies: []string{"task-1"}},
		{TaskID: "task-3", Dependencies: []string{"task-2"}},
	}

	for _, task := range tasks {
		if err := sched.Submit(task); err != nil {
			t.Fatalf("failed to submit task: %v", err)
		}
	}

	ctx := NewSelfContext("task-3")
	if err := cb.collectTaskLineage(ctx, "task-3"); err != nil {
		t.Fatalf("collectTaskLineage failed: %v", err)
	}

	// Lineage should contain task-1 and task-2 (ancestors)
	if len(ctx.TaskLineage) != 2 {
		t.Errorf("expected lineage length 2, got %d", len(ctx.TaskLineage))
	}
}

func TestContextBuilder_collectTaskLineageWithCycle(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/test/root")

	// Create tasks but use state store directly to simulate cycle scenario
	state := types.TaskState{
		Task: &types.AgentTask{
			TaskID:       "task-1",
			Dependencies: []string{"task-2"},
		},
	}
	ss.Save("task-1", state)

	state2 := types.TaskState{
		Task: &types.AgentTask{
			TaskID:       "task-2",
			Dependencies: []string{"task-1"}, // cycle
		},
	}
	ss.Save("task-2", state2)

	ctx := NewSelfContext("task-1")
	err := cb.collectTaskLineage(ctx, "task-1")
	// Should not error but should handle cycle gracefully
	if err != nil {
		t.Fatalf("collectTaskLineage should handle cycles gracefully: %v", err)
	}
}

func TestContextBuilder_collectDocSnapshot(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, ".") // Use current dir

	snapshot, err := cb.collectDocSnapshot()
	if err != nil {
		t.Fatalf("collectDocSnapshot failed: %v", err)
	}

	// Should find doc files in docs directory
	if snapshot == nil {
		t.Fatal("snapshot should not be nil")
	}
	// SpecFiles, DocFiles, Reports should be initialized (may be empty if no files found)
	if snapshot.SpecFiles == nil {
		t.Error("SpecFiles should not be nil")
	}
	if snapshot.DocFiles == nil {
		t.Error("DocFiles should not be nil")
	}
	if snapshot.Reports == nil {
		t.Error("Reports should not be nil")
	}
}

func TestContextBuilder_collectCodeSnapshot(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, ".") // Use current dir

	snapshot, err := cb.collectCodeSnapshot()
	if err != nil {
		t.Fatalf("collectCodeSnapshot failed: %v", err)
	}

	if snapshot == nil {
		t.Fatal("snapshot should not be nil")
	}
	// Should find some Go files
	if len(snapshot.ModifiedFiles) == 0 {
		t.Log("Note: No Go files found in internal/ directory")
	}
	// TaskCount should match scheduler
	if snapshot.TaskCount != 0 {
		t.Errorf("expected TaskCount 0, got %d", snapshot.TaskCount)
	}
	// SpecVersion should be populated
	if snapshot.SpecVersion == "" {
		t.Log("Note: SpecVersion is empty (expected if docs/current-progress.md not found)")
	}
}

func TestBuildSelfContext(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})

	// Submit some tasks
	task := &types.AgentTask{
		TaskID:       "task-1",
		Dependencies: []string{},
	}
	if err := sched.Submit(task); err != nil {
		t.Fatalf("failed to submit task: %v", err)
	}

	cb := NewContextBuilder(sched, ss, ".")
	ctx, err := cb.BuildSelfContext("task-1")

	if err != nil {
		t.Fatalf("BuildSelfContext failed: %v", err)
	}
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	if ctx.TaskID != "task-1" {
		t.Errorf("expected TaskID 'task-1', got '%s'", ctx.TaskID)
	}
	if ctx.CodeSnapshot == nil {
		t.Error("CodeSnapshot should not be nil")
	}
	if ctx.DocSnapshot == nil {
		t.Error("DocSnapshot should not be nil")
	}
	if ctx.StateSnapshot == nil {
		t.Error("StateSnapshot should not be nil")
	}
}

func TestBuildSelfContext_WithLineage(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})

	// Create dependency chain: task-3 -> task-2 -> task-1
	task1 := &types.AgentTask{TaskID: "task-1", Dependencies: []string{}}
	task2 := &types.AgentTask{TaskID: "task-2", Dependencies: []string{"task-1"}}
	task3 := &types.AgentTask{TaskID: "task-3", Dependencies: []string{"task-2"}}

	for _, task := range []*types.AgentTask{task1, task2, task3} {
		if err := sched.Submit(task); err != nil {
			t.Fatalf("failed to submit task: %v", err)
		}
	}

	cb := NewContextBuilder(sched, ss, ".")
	ctx, err := cb.BuildSelfContext("task-3")

	if err != nil {
		t.Fatalf("BuildSelfContext failed: %v", err)
	}
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	// Lineage should contain ancestors
	if len(ctx.TaskLineage) != 2 {
		t.Errorf("expected lineage length 2, got %d", len(ctx.TaskLineage))
	}
}

func TestBuildSelfContext_WithNonExistentTask(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, ".")

	ctx, err := cb.BuildSelfContext("non-existent-task")

	// Should still build context but with empty lineage
	if err != nil {
		t.Fatalf("BuildSelfContext should not fail for non-existent task: %v", err)
	}
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	if len(ctx.TaskLineage) != 0 {
		t.Errorf("expected empty lineage for non-existent task, got %d", len(ctx.TaskLineage))
	}
}

func TestTraverseDependencies_WithStateStore(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/test/root")

	// Add task directly to state store
	state := types.TaskState{
		Task: &types.AgentTask{
			TaskID:       "direct-task",
			Dependencies: []string{"parent-1", "parent-2"},
		},
	}
	ss.Save("direct-task", state)

	// Save parent tasks too
	for i := 1; i <= 2; i++ {
		parentState := types.TaskState{
			Task: &types.AgentTask{
				TaskID:       "parent-" + string(rune('0'+i)),
				Dependencies: []string{},
			},
		}
		ss.Save("parent-"+string(rune('0'+i)), parentState)
	}

	ctx := NewSelfContext("direct-task")
	err := cb.collectTaskLineage(ctx, "direct-task")

	if err != nil {
		t.Fatalf("collectTaskLineage failed: %v", err)
	}
	if len(ctx.TaskLineage) != 2 {
		t.Errorf("expected lineage length 2, got %d", len(ctx.TaskLineage))
	}
}

func TestReadSpecVersion(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, ".") // Use current dir with docs/current-progress.md

	version := cb.readSpecVersion()

	// Version should be determined from docs/current-progress.md
	if version == "" {
		t.Error("spec version should not be empty when docs/current-progress.md exists")
	}
}

func TestCollectDocSnapshot_WithFiles(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, ".") // Use current dir

	snapshot, err := cb.collectDocSnapshot()
	if err != nil {
		t.Fatalf("collectDocSnapshot failed: %v", err)
	}

	// With real docs directory, should have some files
	if snapshot.SpecFiles == nil {
		t.Error("SpecFiles should not be nil")
	}
	if snapshot.DocFiles == nil {
		t.Error("DocFiles should not be nil")
	}
	if snapshot.Reports == nil {
		t.Error("Reports should not be nil")
	}

	// Log what we found for debugging
	if len(snapshot.SpecFiles) > 0 {
		t.Logf("Found %d spec files", len(snapshot.SpecFiles))
	}
	if len(snapshot.DocFiles) > 0 {
		t.Logf("Found %d doc files", len(snapshot.DocFiles))
	}
	if len(snapshot.Reports) > 0 {
		t.Logf("Found %d report files", len(snapshot.Reports))
	}
}

func TestCollectDocSnapshot_NonExistentRoot(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/non/existent/path")

	snapshot, err := cb.collectDocSnapshot()

	// Should not fail even if directory doesn't exist
	if err != nil {
		t.Fatalf("collectDocSnapshot should not fail for non-existent root: %v", err)
	}
	if snapshot == nil {
		t.Fatal("snapshot should not be nil")
	}
}

func TestCollectCodeSnapshot_NonExistentRoot(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/non/existent/path")

	snapshot, err := cb.collectCodeSnapshot()

	// Should not fail
	if err != nil {
		t.Fatalf("collectCodeSnapshot should not fail for non-existent root: %v", err)
	}
	if snapshot == nil {
		t.Fatal("snapshot should not be nil")
	}
}

func TestTraverseDependencies_WithNestedDeps(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/test/root")

	// Add a task with nested dependencies
	state := types.TaskState{
		Task: &types.AgentTask{
			TaskID:       "root-task",
			Dependencies: []string{"level-1a", "level-1b"},
		},
	}
	ss.Save("root-task", state)

	state1a := types.TaskState{
		Task: &types.AgentTask{
			TaskID:       "level-1a",
			Dependencies: []string{"level-2"},
		},
	}
	ss.Save("level-1a", state1a)

	state1b := types.TaskState{
		Task: &types.AgentTask{
			TaskID:       "level-1b",
			Dependencies: []string{},
		},
	}
	ss.Save("level-1b", state1b)

	state2 := types.TaskState{
		Task: &types.AgentTask{
			TaskID:       "level-2",
			Dependencies: []string{},
		},
	}
	ss.Save("level-2", state2)

	ctx := NewSelfContext("root-task")
	err := cb.collectTaskLineage(ctx, "root-task")

	if err != nil {
		t.Fatalf("collectTaskLineage failed: %v", err)
	}
	// Should have all ancestors: level-1a, level-1b, level-2
	if len(ctx.TaskLineage) != 3 {
		t.Errorf("expected lineage length 3, got %d", len(ctx.TaskLineage))
	}
}

func TestReadSpecVersion_NonExistentFile(t *testing.T) {
	ss := sharedlayer.NewMemoryStateStore()
	sched := scheduler.NewScheduler(ss, &mockLifecycle{running: true})
	cb := NewContextBuilder(sched, ss, "/non/existent")

	version := cb.readSpecVersion()

	// Should return "unknown" for non-existent path
	if version != "unknown" {
		t.Errorf("expected 'unknown' for non-existent path, got '%s'", version)
	}
}

// mockLifecycle implements LifecycleChecker for testing
type mockLifecycle struct {
	running bool
}

func (m *mockLifecycle) IsRunning() bool {
	return m.running
}
