package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestOrchestrator_NewOrchestrator(t *testing.T) {
	orch := NewOrchestrator()
	if orch == nil {
		t.Fatal("NewOrchestrator should return a non-nil orchestrator")
	}
	if orch.IsRunning() {
		t.Error("NewOrchestrator should create a stopped orchestrator")
	}
}

func TestOrchestrator_Start(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	err := orch.Start(ctx)
	if err != nil {
		t.Fatalf("Start should succeed: %v", err)
	}

	err = orch.Start(ctx)
	if err == nil {
		t.Error("Start should fail when orchestrator is already running")
	}

	err = orch.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown should succeed: %v", err)
	}
}

func TestOrchestrator_Shutdown(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()

	// Shutdown should succeed even when orchestrator has not been started
	err := orch.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown should succeed: %v", err)
	}

	// Should not be running after shutdown
	if orch.IsRunning() {
		t.Error("Orchestrator should not be running after shutdown")
	}
}

func TestOrchestrator_SubmitTask(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()
	defer orch.Shutdown(ctx)

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
	}

	err := orch.SubmitTask(task)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Check task status
	status, err := orch.GetTaskStatus("task-1")
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}
	if status != types.TaskStatusPending && status != types.TaskStatusRunning {
		t.Errorf("Task should be pending or running, got %s", status)
	}
}

func TestOrchestrator_SubmitTask_AdmissionRejectsUnknownContract(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()
	defer orch.Shutdown(ctx)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "nonexistent",
		Input:      map[string]any{"message": "test"},
	}

	err := orch.SubmitTask(task)
	if err == nil {
		t.Error("SubmitTask should reject task with unknown contract")
	}
}

func TestOrchestrator_SubmitTask_AdmissionRejectsInvalidInput(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()
	defer orch.Shutdown(ctx)

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{}, // missing required "message" field
	}

	err := orch.SubmitTask(task)
	if err == nil {
		t.Error("SubmitTask should reject task with invalid input")
	}
}

func TestOrchestrator_GetTaskStatus(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()
	defer orch.Shutdown(ctx)

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
	}

	// Get status before submission should fail
	_, err := orch.GetTaskStatus("task-1")
	if err == nil {
		t.Error("GetTaskStatus should fail for non-existent task")
	}

	// Submit task
	err = orch.SubmitTask(task)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Get status after submission should succeed
	status, err := orch.GetTaskStatus("task-1")
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}
	if status != types.TaskStatusPending && status != types.TaskStatusRunning {
		t.Errorf("Task should be pending or running, got %s", status)
	}
}

func TestOrchestrator_RegisterContract(t *testing.T) {
	orch := NewOrchestrator()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "name",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "result",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}

	err := orch.RegisterContract(contract)
	if err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	// Duplicate registration should fail
	err = orch.RegisterContract(contract)
	if err == nil {
		t.Error("Duplicate contract registration should fail")
	}
}

func TestOrchestrator_IsRunning(t *testing.T) {
	orch := NewOrchestrator()

	if orch.IsRunning() {
		t.Error("Orchestrator should not be running after creation")
	}

	ctx := context.Background()

	err := orch.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}
	if !orch.IsRunning() {
		t.Error("Orchestrator should be running after start")
	}

	err = orch.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Failed to shutdown orchestrator: %v", err)
	}
	if orch.IsRunning() {
		t.Error("Orchestrator should not be running after shutdown")
	}
}

func TestOrchestrator_TaskExecution(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	task := &types.AgentTask{
		TaskID:     "task-execution",
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusCompleted)
	if status != types.TaskStatusCompleted {
		t.Fatalf("Expected task to complete, got %s", status)
	}
}

func TestOrchestrator_MultipleTasks(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	for _, taskID := range []string{"task-one", "task-two"} {
		task := &types.AgentTask{
			TaskID:     taskID,
			ContractID: "default",
			Input:      map[string]any{"message": "test"},
		}
		if err := orch.SubmitTask(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", taskID, err)
		}
	}

	for _, taskID := range []string{"task-one", "task-two"} {
		status := waitForTaskStatus(t, orch, taskID, types.TaskStatusCompleted)
		if status != types.TaskStatusCompleted {
			t.Fatalf("Expected task %s to complete, got %s", taskID, status)
		}
	}
}

func TestOrchestrator_ParseSLA_Valid(t *testing.T) {
	timeoutMs, maxRetries := parseSLA(map[string]string{
		types.SLAKeyTimeoutMs:  "5000",
		types.SLAKeyMaxRetries: "2",
	})
	if timeoutMs != 5000 {
		t.Errorf("Expected timeoutMs=5000, got %d", timeoutMs)
	}
	if maxRetries != 2 {
		t.Errorf("Expected maxRetries=2, got %d", maxRetries)
	}
}

func TestOrchestrator_ParseSLA_Missing(t *testing.T) {
	timeoutMs, maxRetries := parseSLA(nil)
	if timeoutMs != 0 {
		t.Errorf("Expected default timeoutMs=0, got %d", timeoutMs)
	}
	if maxRetries != 0 {
		t.Errorf("Expected default maxRetries=0, got %d", maxRetries)
	}
}

func TestOrchestrator_ParseSLA_Invalid(t *testing.T) {
	timeoutMs, maxRetries := parseSLA(map[string]string{
		types.SLAKeyTimeoutMs:  "not-a-number",
		types.SLAKeyMaxRetries: "xyz",
	})
	if timeoutMs != 0 {
		t.Errorf("Invalid timeoutMs should default to 0, got %d", timeoutMs)
	}
	if maxRetries != 0 {
		t.Errorf("Invalid maxRetries should default to 0, got %d", maxRetries)
	}
}

func TestOrchestrator_TaskWithSLA(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	task := &types.AgentTask{
		TaskID:     "task-sla",
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
		Metadata:   map[string]string{types.SLAKeyTimeoutMs: "60000", types.SLAKeyMaxRetries: "1"},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusCompleted)
	if status != types.TaskStatusCompleted {
		t.Fatalf("Expected task to complete with SLA, got %s", status)
	}
}

func TestOrchestrator_TaskWithSLA_ZeroRetries(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	// max_retries=0 means no retry on failure (single attempt)
	task := &types.AgentTask{
		TaskID:     "task-zero-retries",
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
		Metadata:   map[string]string{types.SLAKeyMaxRetries: "0"},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusCompleted)
	if status != types.TaskStatusCompleted {
		t.Errorf("Task with max_retries=0 should complete normally, got %s", status)
	}
}

func waitForTaskStatus(t *testing.T, orch *Orchestrator, taskID string, expected types.TaskStatus) types.TaskStatus {
	t.Helper()

	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			status, err := orch.GetTaskStatus(taskID)
			if err != nil {
				t.Fatalf("Failed to get final task status: %v", err)
			}
			return status
		case <-ticker.C:
			status, err := orch.GetTaskStatus(taskID)
			if err != nil {
				t.Fatalf("Failed to get task status: %v", err)
			}
			if status == expected || status == types.TaskStatusFailed {
				return status
			}
		}
	}
}

func testDefaultContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: "default",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "message",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}
}
