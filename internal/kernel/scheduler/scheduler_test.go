package scheduler

import (
	"testing"

	"github.com/axis-cli/axis/internal/kernel/lifecycle"
	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/types"
)

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
