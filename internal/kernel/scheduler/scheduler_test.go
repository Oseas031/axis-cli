package scheduler

import (
	"testing"

	"github.com/axis-cli/axis/internal/kernel/lifecycle"
	"github.com/axis-cli/axis/internal/kernel/shared_layer"
	"github.com/axis-cli/axis/internal/types"
)

func TestScheduler_Submit(t *testing.T) {
	stateStore := shared_layer.NewMemoryStateStore()
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

	status := sched.GetStatus("task-1")
	if status != types.TaskStatusPending {
		t.Errorf("Expected status %s, got %s", types.TaskStatusPending, status)
	}
}

func TestScheduler_DuplicateSubmit(t *testing.T) {
	stateStore := shared_layer.NewMemoryStateStore()
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
	stateStore := shared_layer.NewMemoryStateStore()
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

	status := sched.GetStatus("task-1")
	if status != "" {
		t.Error("Cancelled task should not exist")
	}
}

func TestScheduler_DependencyManagement(t *testing.T) {
	stateStore := shared_layer.NewMemoryStateStore()
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
	stateStore := shared_layer.NewMemoryStateStore()
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
	stateStore := shared_layer.NewMemoryStateStore()
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

func TestScheduler_UpdateTaskStatus(t *testing.T) {
	stateStore := shared_layer.NewMemoryStateStore()
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

	status := sched.GetStatus("task-1")
	if status != types.TaskStatusRunning {
		t.Errorf("Expected status %s, got %s", types.TaskStatusRunning, status)
	}

	// Check that StartedAt is set
	state, _ := stateStore.Load("task-1")
	if state.Task.StartedAt == nil {
		t.Error("StartedAt should be set when status is running")
	}
}
