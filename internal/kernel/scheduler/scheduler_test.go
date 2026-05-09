package scheduler

import (
	"context"
	"testing"

	"github.com/axis-cli/axis/internal/kernel/lifecycle"
	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/types"
)

type mockLifecycleChecker struct{ running bool }

func (m *mockLifecycleChecker) IsRunning() bool { return m.running }

func TestScheduler_Submit(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
		Input:      map[string]any{"key": "value"},
		Status:     types.TaskStatusPending,
	}

	err := sched.Submit(task)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status, err := sched.GetStatus("task-1")
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}
	if status != types.TaskStatusPending {
		t.Errorf("Expected status %s, got %s", types.TaskStatusPending, status)
	}
}

func TestScheduler_DuplicateSubmit(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
	}

	sched.Submit(task)

	err := sched.Submit(task)
	if err == nil {
		t.Error("Duplicate submit should fail")
	}
}

func TestScheduler_Cancel(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
		Status:     types.TaskStatusPending,
	}

	sched.Submit(task)

	err := sched.Cancel("task-1")
	if err != nil {
		t.Fatalf("Failed to cancel task: %v", err)
	}

	_, err = sched.GetStatus("task-1")
	if err == nil {
		t.Error("Cancelled task should not exist")
	}
}

func TestScheduler_DependencyManagement(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task1 := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
		Status:     types.TaskStatusPending,
	}

	task2 := &types.AgentTask{
		TaskID:       "task-2",
		ContractID:   "contract-1",
		Dependencies: []string{"task-1"},
		Status:       types.TaskStatusPending,
	}

	sched.Submit(task1)
	sched.Submit(task2)

	// task-2 should not be ready since task-1 is not completed
	task, _ := sched.GetNextTask()
	if task != nil && task.TaskID == "task-2" {
		t.Error("task-2 should not be ready before task-1 completes")
	}

	// Complete task-1
	sched.UpdateTaskStatus("task-1", types.TaskStatusCompleted)

	// Now task-2 should be ready
	task, _ = sched.GetNextTask()
	if task == nil || task.TaskID != "task-2" {
		t.Error("task-2 should be ready after task-1 completes")
	}
}

func TestScheduler_CircularDependency(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task1 := &types.AgentTask{
		TaskID:       "task-1",
		ContractID:   "contract-1",
		Dependencies: []string{"task-2"},
		Status:       types.TaskStatusPending,
	}

	task2 := &types.AgentTask{
		TaskID:       "task-2",
		ContractID:   "contract-1",
		Dependencies: []string{"task-1"},
		Status:       types.TaskStatusPending,
	}

	sched.Submit(task1)

	err := sched.Submit(task2)
	if err == nil {
		t.Error("Circular dependency should be detected")
	}
}

func TestScheduler_FIFOOrdering(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	for i := 0; i < 5; i++ {
		task := &types.AgentTask{
			TaskID:     string(rune('a' + i)),
			ContractID: "contract-1",
			Status:     types.TaskStatusPending,
		}
		sched.Submit(task)
	}

	// Tasks should be returned in FIFO order
	for i := 0; i < 5; i++ {
		task, _ := sched.GetNextTask()
		expectedID := string(rune('a' + i))
		if task.TaskID != expectedID {
			t.Errorf("Expected task %s, got %s", expectedID, task.TaskID)
		}
		sched.UpdateTaskStatus(task.TaskID, types.TaskStatusCompleted)
	}
}

func TestScheduler_GetReadyTasks(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	for _, taskID := range []string{"task-1", "task-2", "task-3"} {
		task := &types.AgentTask{
			TaskID:     taskID,
			ContractID: "contract-1",
			Status:     types.TaskStatusPending,
		}
		if err := sched.Submit(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", taskID, err)
		}
	}

	tasks, err := sched.GetReadyTasks(2)
	if err != nil {
		t.Fatalf("GetReadyTasks should succeed: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 ready tasks, got %d", len(tasks))
	}
	if tasks[0].TaskID != "task-1" || tasks[1].TaskID != "task-2" {
		t.Fatalf("Expected FIFO ready tasks task-1/task-2, got %s/%s", tasks[0].TaskID, tasks[1].TaskID)
	}

	for _, task := range tasks {
		status, err := sched.GetStatus(task.TaskID)
		if err != nil {
			t.Fatalf("Failed to get status for %s: %v", task.TaskID, err)
		}
		if status != types.TaskStatusRunning {
			t.Fatalf("Expected selected task %s to be running, got %s", task.TaskID, status)
		}
	}

	nextTasks, err := sched.GetReadyTasks(0)
	if err != nil {
		t.Fatalf("GetReadyTasks should succeed without limit: %v", err)
	}
	if len(nextTasks) != 1 || nextTasks[0].TaskID != "task-3" {
		t.Fatalf("Expected remaining ready task task-3, got %#v", nextTasks)
	}
}

func TestScheduler_GetReadyTasksWithDependencies(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	root := &types.AgentTask{
		TaskID:     "root",
		ContractID: "contract-1",
		Status:     types.TaskStatusPending,
	}
	child := &types.AgentTask{
		TaskID:       "child",
		ContractID:   "contract-1",
		Dependencies: []string{"root"},
		Status:       types.TaskStatusPending,
	}
	independent := &types.AgentTask{
		TaskID:     "independent",
		ContractID: "contract-1",
		Status:     types.TaskStatusPending,
	}

	for _, task := range []*types.AgentTask{root, child, independent} {
		if err := sched.Submit(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", task.TaskID, err)
		}
	}

	tasks, err := sched.GetReadyTasks(0)
	if err != nil {
		t.Fatalf("GetReadyTasks should succeed: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("Expected root and independent to be ready, got %d tasks", len(tasks))
	}
	if tasks[0].TaskID != "root" || tasks[1].TaskID != "independent" {
		t.Fatalf("Expected ready tasks root/independent, got %s/%s", tasks[0].TaskID, tasks[1].TaskID)
	}

	if err := sched.UpdateTaskStatus("root", types.TaskStatusCompleted); err != nil {
		t.Fatalf("Failed to complete root: %v", err)
	}

	tasks, err = sched.GetReadyTasks(0)
	if err != nil {
		t.Fatalf("GetReadyTasks should succeed: %v", err)
	}
	if len(tasks) != 1 || tasks[0].TaskID != "child" {
		t.Fatalf("Expected child to become ready, got %#v", tasks)
	}
}

func TestScheduler_Cancel_NonExistent(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	err := sched.Cancel("nonexistent")
	if err == nil {
		t.Error("Cancel of non-existent task should fail")
	}
}

func TestScheduler_Cancel_NonPending(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
	}
	sched.Submit(task)
	sched.UpdateTaskStatus("task-1", types.TaskStatusRunning)

	err := sched.Cancel("task-1")
	if err == nil {
		t.Error("Cancel of non-pending task should fail")
	}
}

func TestScheduler_Submit_NotRunning(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)
	lifecycleMgr.Shutdown(context.Background())

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
	}
	err := sched.Submit(task)
	if err == nil {
		t.Error("Submit should fail when scheduler is not running")
	}
}

func TestScheduler_UpdateTaskStatus_Completed(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
	}
	sched.Submit(task)
	sched.UpdateTaskStatus("task-1", types.TaskStatusRunning)

	err := sched.UpdateTaskStatus("task-1", types.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	status, _ := sched.GetStatus("task-1")
	if status != types.TaskStatusCompleted {
		t.Errorf("Expected completed, got %s", status)
	}

	state, _ := stateStore.Load("task-1")
	if state.Task.CompletedAt == nil {
		t.Error("CompletedAt should be set when status is completed")
	}
}

func TestScheduler_UpdateTaskStatus_Failed(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
	}
	sched.Submit(task)
	sched.UpdateTaskStatus("task-1", types.TaskStatusRunning)

	err := sched.UpdateTaskStatus("task-1", types.TaskStatusFailed)
	if err != nil {
		t.Fatalf("Failed to mark task failed: %v", err)
	}

	status, _ := sched.GetStatus("task-1")
	if status != types.TaskStatusFailed {
		t.Errorf("Expected failed, got %s", status)
	}

	state, _ := stateStore.Load("task-1")
	if state.Task.CompletedAt == nil {
		t.Error("CompletedAt should be set when status is failed")
	}
}

func TestScheduler_UpdateTaskStatus_NonExistent(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	err := sched.UpdateTaskStatus("nonexistent", types.TaskStatusRunning)
	if err == nil {
		t.Error("UpdateTaskStatus for non-existent task should fail")
	}
}

func TestScheduler_AreDependenciesCompleted_FailedDep(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	dep := &types.AgentTask{TaskID: "dep", ContractID: "c"}
	child := &types.AgentTask{TaskID: "child", ContractID: "c", Dependencies: []string{"dep"}}
	sched.Submit(dep)
	sched.Submit(child)
	sched.UpdateTaskStatus("dep", types.TaskStatusFailed)

	tasks, err := sched.GetReadyTasks(0)
	if err != nil {
		t.Fatalf("GetReadyTasks failed: %v", err)
	}
	found := false
	for _, task := range tasks {
		if task.TaskID == "child" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Task with failed dependency should be ready (failed = done)")
	}
}

func TestScheduler_AreDependenciesCompleted_MissingDep(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:       "task-1",
		ContractID:   "contract-1",
		Dependencies: []string{"nonexistent"},
	}
	sched.Submit(task)

	tasks, err := sched.GetReadyTasks(0)
	if err != nil {
		t.Fatalf("GetReadyTasks should succeed: %v", err)
	}
	if len(tasks) != 0 {
		t.Error("Task with missing dependency should not be ready")
	}
}

func TestScheduler_GetNextTask_NotRunning(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)
	lifecycleMgr.Shutdown(context.Background())

	_, err := sched.GetNextTask()
	if err == nil {
		t.Error("GetNextTask should fail when scheduler is not running")
	}
}

func TestScheduler_GetNextTask_Empty(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task, err := sched.GetNextTask()
	if err != nil {
		t.Fatalf("GetNextTask should not error on empty queue: %v", err)
	}
	if task != nil {
		t.Error("GetNextTask should return nil for empty queue")
	}
}

func TestScheduler_UpdateTaskStatus(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleMgr := lifecycle.NewLifecycleManager()
	sched := NewScheduler(stateStore, lifecycleMgr)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "contract-1",
		Status:     types.TaskStatusPending,
	}

	sched.Submit(task)

	err := sched.UpdateTaskStatus("task-1", types.TaskStatusRunning)
	if err != nil {
		t.Fatalf("Failed to update task status: %v", err)
	}

	status, err := sched.GetStatus("task-1")
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}
	if status != types.TaskStatusRunning {
		t.Errorf("Expected status %s, got %s", types.TaskStatusRunning, status)
	}

	// Check that StartedAt is set
	state, _ := stateStore.Load("task-1")
	if state.Task.StartedAt == nil {
		t.Error("StartedAt should be set when status is running")
	}
}

func TestScheduler_GetAllTasks(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycle := &mockLifecycleChecker{running: true}
	sched := NewScheduler(stateStore, lifecycle)
	sched.Submit(&types.AgentTask{TaskID: "a"})
	sched.Submit(&types.AgentTask{TaskID: "b"})
	tasks := sched.GetAllTasks()
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}
}

func TestScheduler_GetAllTasks_Empty(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycle := &mockLifecycleChecker{running: true}
	sched := NewScheduler(stateStore, lifecycle)
	tasks := sched.GetAllTasks()
	if len(tasks) != 0 {
		t.Fatalf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestScheduler_GetDependencyGraph(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycle := &mockLifecycleChecker{running: true}
	sched := NewScheduler(stateStore, lifecycle)
	sched.Submit(&types.AgentTask{TaskID: "a"})
	sched.Submit(&types.AgentTask{TaskID: "b", Dependencies: []string{"a"}})
	sched.Submit(&types.AgentTask{TaskID: "c", Dependencies: []string{"a", "b"}})
	graph := sched.GetDependencyGraph()
	if len(graph) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(graph))
	}
	if len(graph["a"]) != 0 {
		t.Errorf("Task a should have no deps, got %v", graph["a"])
	}
	if len(graph["b"]) != 1 || graph["b"][0] != "a" {
		t.Errorf("Task b should depend on a, got %v", graph["b"])
	}
	if len(graph["c"]) != 2 {
		t.Errorf("Task c should have 2 deps, got %v", graph["c"])
	}
}

func TestScheduler_GetReadyTasks_PrioritySorting(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycle := &mockLifecycleChecker{running: true}
	sched := NewScheduler(stateStore, lifecycle)

	tasks := []*types.AgentTask{
		{TaskID: "low", ContractID: "c", Metadata: map[string]string{types.SLAKeyPriority: "50"}},
		{TaskID: "high", ContractID: "c", Metadata: map[string]string{types.SLAKeyPriority: "200"}},
		{TaskID: "medium", ContractID: "c", Metadata: map[string]string{types.SLAKeyPriority: "128"}},
		{TaskID: "default", ContractID: "c"}, // no priority → 128
	}
	for _, task := range tasks {
		if err := sched.Submit(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", task.TaskID, err)
		}
	}

	ready, err := sched.GetReadyTasks(0)
	if err != nil {
		t.Fatalf("GetReadyTasks failed: %v", err)
	}
	if len(ready) != 4 {
		t.Fatalf("Expected 4 ready tasks, got %d", len(ready))
	}

	// Expected order: high (200), medium (128), default (128), low (50)
	expected := []string{"high", "medium", "default", "low"}
	for i, task := range ready {
		if task.TaskID != expected[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expected[i], task.TaskID)
		}
	}
}

func TestScheduler_GetReadyTasks_PriorityPreservesFIFO(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycle := &mockLifecycleChecker{running: true}
	sched := NewScheduler(stateStore, lifecycle)

	// Submit two tasks with same priority; FIFO order should be preserved
	tasks := []*types.AgentTask{
		{TaskID: "first", ContractID: "c", Metadata: map[string]string{types.SLAKeyPriority: "100"}},
		{TaskID: "second", ContractID: "c", Metadata: map[string]string{types.SLAKeyPriority: "100"}},
	}
	for _, task := range tasks {
		if err := sched.Submit(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", task.TaskID, err)
		}
	}

	ready, err := sched.GetReadyTasks(0)
	if err != nil {
		t.Fatalf("GetReadyTasks failed: %v", err)
	}
	if len(ready) != 2 {
		t.Fatalf("Expected 2 ready tasks, got %d", len(ready))
	}
	if ready[0].TaskID != "first" || ready[1].TaskID != "second" {
		t.Errorf("Expected FIFO order first,second for same priority, got %s,%s", ready[0].TaskID, ready[1].TaskID)
	}
}

func TestScheduler_GetReadyTasks_PriorityLimit(t *testing.T) {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycle := &mockLifecycleChecker{running: true}
	sched := NewScheduler(stateStore, lifecycle)

	tasks := []*types.AgentTask{
		{TaskID: "low", ContractID: "c", Metadata: map[string]string{types.SLAKeyPriority: "10"}},
		{TaskID: "high", ContractID: "c", Metadata: map[string]string{types.SLAKeyPriority: "200"}},
	}
	for _, task := range tasks {
		if err := sched.Submit(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", task.TaskID, err)
		}
	}

	ready, err := sched.GetReadyTasks(1)
	if err != nil {
		t.Fatalf("GetReadyTasks failed: %v", err)
	}
	if len(ready) != 1 {
		t.Fatalf("Expected 1 ready task, got %d", len(ready))
	}
	if ready[0].TaskID != "high" {
		t.Errorf("Expected high-priority task first with limit, got %s", ready[0].TaskID)
	}
}
