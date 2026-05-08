package shared_layer

import (
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestMemoryStateStore_SaveAndLoad(t *testing.T) {
	store := NewMemoryStateStore()

	task := &types.AgentTask{
		TaskID:     "test-1",
		ContractID: "contract-1",
		Input:      map[string]any{"key": "value"},
		Status:     types.TaskStatusPending,
	}

	state := types.TaskState{
		Task:      task,
		UpdatedAt: time.Now(),
	}

	err := store.Save("test-1", state)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	loaded, err := store.Load("test-1")
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if loaded.Task.TaskID != task.TaskID {
		t.Errorf("Expected task ID %s, got %s", task.TaskID, loaded.Task.TaskID)
	}
}

func TestMemoryStateStore_LoadNonExistent(t *testing.T) {
	store := NewMemoryStateStore()

	_, err := store.Load("non-existent")
	if err != nil {
		t.Fatalf("Load should not error for non-existent key: %v", err)
	}
}

func TestMemoryStateStore_Delete(t *testing.T) {
	store := NewMemoryStateStore()

	task := &types.AgentTask{
		TaskID: "test-1",
	}

	state := types.TaskState{
		Task:      task,
		UpdatedAt: time.Now(),
	}

	store.Save("test-1", state)
	err := store.Delete("test-1")
	if err != nil {
		t.Fatalf("Failed to delete state: %v", err)
	}

	_, err = store.Load("test-1")
	if err != nil {
		t.Fatalf("Load after delete should not error: %v", err)
	}
}
