package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/model/provider"
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

func TestOrchestrator_WithModelProvider(t *testing.T) {
	echo := provider.NewEchoModelProvider()
	orch := NewOrchestrator(WithModelProvider(echo))
	if orch == nil {
		t.Fatal("NewOrchestrator should return non-nil")
	}
	ctx := context.Background()
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Start should succeed: %v", err)
	}
	if err := orch.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown should succeed: %v", err)
	}
}

func TestOrchestrator_DefaultToolRegistryExposesCoreTools(t *testing.T) {
	p := &capturingToolsProvider{called: make(chan struct{}, 1)}
	orch := NewOrchestrator(WithModelProvider(p))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())
	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("RegisterContract should succeed: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Start should succeed: %v", err)
	}
	if err := orch.SubmitTask(&types.AgentTask{TaskID: "tools-task", ContractID: "default", Input: map[string]any{"message": "use tools"}}); err != nil {
		t.Fatalf("SubmitTask should succeed: %v", err)
	}
	select {
	case <-p.called:
	case <-time.After(2 * time.Second):
		t.Fatal("provider was not called")
	}
	got := append([]string(nil), p.toolNames...)
	sort.Strings(got)
	want := []string{"bash", "checkpoint", "compact", "file_read", "file_write", "http_request", "load_skill", "spawn", "yield"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("expected default tools %v, got %v", want, got)
	}
}

func TestOrchestrator_DefaultToolRegistryExecutesFileWriteToolCall(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer os.Chdir(cwd)

	outputPath := filepath.Join(tmp, "axis-intro.md")
	p := &fileWriteToolCallProvider{path: outputPath, content: "Axis intro"}
	orch := NewOrchestrator(WithModelProvider(p))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())
	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("RegisterContract should succeed: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Start should succeed: %v", err)
	}
	if err := orch.SubmitTask(&types.AgentTask{TaskID: "file-write-task", ContractID: "default", Input: map[string]any{"message": "write file"}}); err != nil {
		t.Fatalf("SubmitTask should succeed: %v", err)
	}
	status := waitForTaskStatus(t, orch, "file-write-task", types.TaskStatusCompleted)
	if status != types.TaskStatusCompleted {
		t.Fatalf("expected file-write task to complete, got %s", status)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected file_write tool to create file: %v", err)
	}
	if string(data) != "Axis intro" {
		t.Fatalf("unexpected file content: %q", string(data))
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

func TestOrchestrator_SubmitTask_Duplicate(t *testing.T) {
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

	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("First submit should succeed: %v", err)
	}
	if err := orch.SubmitTask(task); err == nil {
		t.Error("Duplicate submit should fail")
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
	timeoutMs, maxRetries, failureClass, backoff := parseSLA(map[string]string{
		types.SLAKeyTimeoutMs:    "5000",
		types.SLAKeyMaxRetries:   "2",
		types.SLAKeyFailureClass: "fatal",
		types.SLAKeyBackoff:      "linear",
	})
	if timeoutMs != 5000 {
		t.Errorf("Expected timeoutMs=5000, got %d", timeoutMs)
	}
	if maxRetries != 2 {
		t.Errorf("Expected maxRetries=2, got %d", maxRetries)
	}
	if failureClass != "fatal" {
		t.Errorf("Expected failureClass=fatal, got %s", failureClass)
	}
	if backoff != "linear" {
		t.Errorf("Expected backoff=linear, got %s", backoff)
	}
}

func TestOrchestrator_ParseSLA_Missing(t *testing.T) {
	timeoutMs, maxRetries, failureClass, backoff := parseSLA(nil)
	if timeoutMs != 0 {
		t.Errorf("Expected default timeoutMs=0, got %d", timeoutMs)
	}
	if maxRetries != 0 {
		t.Errorf("Expected default maxRetries=0, got %d", maxRetries)
	}
	if failureClass != "" {
		t.Errorf("Expected default failureClass=empty, got %s", failureClass)
	}
	if backoff != "" {
		t.Errorf("Expected default backoff=empty, got %s", backoff)
	}
}

func TestOrchestrator_ParseSLA_Invalid(t *testing.T) {
	timeoutMs, maxRetries, failureClass, backoff := parseSLA(map[string]string{
		types.SLAKeyTimeoutMs:    "not-a-number",
		types.SLAKeyMaxRetries:   "xyz",
		types.SLAKeyFailureClass: "bad-value",
		types.SLAKeyBackoff:      "unknown",
	})
	if timeoutMs != 0 {
		t.Errorf("Invalid timeoutMs should default to 0, got %d", timeoutMs)
	}
	if maxRetries != 0 {
		t.Errorf("Invalid maxRetries should default to 0, got %d", maxRetries)
	}
	if failureClass != "bad-value" {
		t.Errorf("failureClass should pass through as-is, got %s", failureClass)
	}
	if backoff != "unknown" {
		t.Errorf("backoff should pass through as-is, got %s", backoff)
	}
}

func TestOrchestrator_ParseSLA_MaxRetryCap(t *testing.T) {
	_, maxRetries, _, _ := parseSLA(map[string]string{
		types.SLAKeyMaxRetries: "10",
	})
	if maxRetries != types.MaxRetryLimit {
		t.Errorf("Expected maxRetries capped at %d, got %d", types.MaxRetryLimit, maxRetries)
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

func TestOrchestrator_ParallelExecution(t *testing.T) {
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

	// Submit 10 independent tasks — exercises parallel dispatch
	for i := 0; i < 10; i++ {
		task := &types.AgentTask{
			TaskID:     fmt.Sprintf("task-%d", i),
			ContractID: "default",
			Input:      map[string]any{"message": "test"},
		}
		if err := orch.SubmitTask(task); err != nil {
			t.Fatalf("Failed to submit task %d: %v", i, err)
		}
	}

	for i := 0; i < 10; i++ {
		status := waitForTaskStatus(t, orch, fmt.Sprintf("task-%d", i), types.TaskStatusCompleted)
		if status != types.TaskStatusCompleted {
			t.Errorf("Task %d should complete, got %s", i, status)
		}
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

func TestOrchestrator_TaskWithRetryExhaustion(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	// Register a contract whose output validation will fail because
	// MockModelProvider never provides the required "missing_field"
	failingContract := &types.AgentContract{
		ContractID: "failing-output",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "msg", Type: types.FieldTypeString, Required: false}},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{{Name: "missing_field", Type: types.FieldTypeString, Required: true}},
		},
	}
	if err := orch.RegisterContract(failingContract); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	task := &types.AgentTask{
		TaskID:     "retry-fail",
		ContractID: "failing-output",
		Input:      map[string]any{"msg": "test"},
		Metadata:   map[string]string{types.SLAKeyMaxRetries: "3"},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusFailed)
	if status != types.TaskStatusFailed {
		t.Errorf("Expected task to fail after retries, got %s", status)
	}
}

func TestOrchestrator_TaskFailsNoRetry(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	failingContract := &types.AgentContract{
		ContractID: "failing-once",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "msg", Type: types.FieldTypeString, Required: false}},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{{Name: "required_output", Type: types.FieldTypeString, Required: true}},
		},
	}
	if err := orch.RegisterContract(failingContract); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	// max_retries=0 means no retry on first failure, wraps with ErrTaskTimeout
	task := &types.AgentTask{
		TaskID:     "fail-no-retry",
		ContractID: "failing-once",
		Input:      map[string]any{"msg": "test"},
		Metadata:   map[string]string{types.SLAKeyMaxRetries: "0"},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusFailed)
	if status != types.TaskStatusFailed {
		t.Errorf("Expected task to fail without retry, got %s", status)
	}
}

func TestOrchestrator_TaskWithTimeoutRetry(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	failingContract := &types.AgentContract{
		ContractID: "timeout-fail",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "msg", Type: types.FieldTypeString, Required: false}},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{{Name: "missing", Type: types.FieldTypeString, Required: true}},
		},
	}
	if err := orch.RegisterContract(failingContract); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	// timeout + retries — timeout triggers retry, retry exhaustion wraps both
	task := &types.AgentTask{
		TaskID:     "timeout-retry",
		ContractID: "timeout-fail",
		Input:      map[string]any{"msg": "test"},
		Metadata:   map[string]string{types.SLAKeyTimeoutMs: "60000", types.SLAKeyMaxRetries: "1"},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusFailed)
	if status != types.TaskStatusFailed {
		t.Errorf("Expected task to fail with timeout+retry, got %s", status)
	}
}

func TestOrchestrator_GetAllTasks(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()
	defer orch.Shutdown(ctx)

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	// Submit a few tasks
	for _, taskID := range []string{"a", "b", "c"} {
		task := &types.AgentTask{
			TaskID:     taskID,
			ContractID: "default",
			Input:      map[string]any{"message": "test"},
		}
		if err := orch.SubmitTask(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", taskID, err)
		}
	}

	tasks := orch.GetAllTasks()
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	seen := make(map[string]bool)
	for _, task := range tasks {
		seen[task.TaskID] = true
	}
	for _, taskID := range []string{"a", "b", "c"} {
		if !seen[taskID] {
			t.Errorf("Expected task %s in GetAllTasks result", taskID)
		}
	}
}

func TestOrchestrator_GetAllTasks_Empty(t *testing.T) {
	orch := NewOrchestrator()
	tasks := orch.GetAllTasks()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks from fresh orchestrator, got %d", len(tasks))
	}
}

func TestOrchestrator_GetDependencyGraph(t *testing.T) {
	orch := NewOrchestrator()
	ctx := context.Background()
	defer orch.Shutdown(ctx)

	if err := orch.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	// Submit tasks with dependencies: a depends on b, b depends on c
	taskA := &types.AgentTask{
		TaskID:       "a",
		ContractID:   "default",
		Input:        map[string]any{"message": "test"},
		Dependencies: []string{"b"},
	}
	taskB := &types.AgentTask{
		TaskID:       "b",
		ContractID:   "default",
		Input:        map[string]any{"message": "test"},
		Dependencies: []string{"c"},
	}
	taskC := &types.AgentTask{
		TaskID:     "c",
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
	}

	for _, task := range []*types.AgentTask{taskA, taskB, taskC} {
		if err := orch.SubmitTask(task); err != nil {
			t.Fatalf("Failed to submit task %s: %v", task.TaskID, err)
		}
	}

	graph := orch.GetDependencyGraph()
	if len(graph) == 0 {
		t.Fatal("Expected non-empty dependency graph")
	}

	// Verify dependency relationships
	if deps, ok := graph["a"]; !ok {
		t.Error("Expected task 'a' in dependency graph")
	} else if len(deps) != 1 || deps[0] != "b" {
		t.Errorf("Expected task 'a' to depend on ['b'], got %v", deps)
	}

	if deps, ok := graph["b"]; !ok {
		t.Error("Expected task 'b' in dependency graph")
	} else if len(deps) != 1 || deps[0] != "c" {
		t.Errorf("Expected task 'b' to depend on ['c'], got %v", deps)
	}
}

func TestOrchestrator_GetDependencyGraph_Empty(t *testing.T) {
	orch := NewOrchestrator()
	graph := orch.GetDependencyGraph()
	if len(graph) != 0 {
		t.Errorf("Expected empty dependency graph from fresh orchestrator, got %v", graph)
	}
}

func TestOrchestrator_ResolveCall(t *testing.T) {
	orch := NewOrchestrator()
	err := orch.ResolveCall("nonexistent-call", map[string]any{"output": "test"})
	if err == nil {
		t.Error("Expected error resolving non-existent call, got nil")
	}
}

func TestOrchestrator_BackoffDelay_Default(t *testing.T) {
	d := backoffDelay("", 0)
	if d != 100*time.Millisecond {
		t.Errorf("Expected default backoff 100ms, got %v", d)
	}
	d = backoffDelay("fixed", 5)
	if d != 100*time.Millisecond {
		t.Errorf("Expected fixed backoff 100ms, got %v", d)
	}
}

func TestOrchestrator_BackoffDelay_Linear(t *testing.T) {
	d := backoffDelay("linear", 0)
	if d != 100*time.Millisecond {
		t.Errorf("Expected linear backoff[0]=100ms, got %v", d)
	}
	d = backoffDelay("linear", 2)
	if d != 300*time.Millisecond {
		t.Errorf("Expected linear backoff[2]=300ms, got %v", d)
	}
}

func TestOrchestrator_BackoffDelay_Exponential(t *testing.T) {
	d := backoffDelay("exponential", 0)
	if d != 100*time.Millisecond {
		t.Errorf("Expected exp backoff[0]=100ms, got %v", d)
	}
	d = backoffDelay("exponential", 2)
	if d != 400*time.Millisecond {
		t.Errorf("Expected exp backoff[2]=400ms, got %v", d)
	}
	d = backoffDelay("exponential", 10)
	if d > 30*time.Second {
		t.Errorf("Expected exp backoff[10] capped at 30s, got %v", d)
	}
}

func TestOrchestrator_FatalFailureClass(t *testing.T) {
	orch := NewOrchestrator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer orch.Shutdown(context.Background())

	failingContract := &types.AgentContract{
		ContractID: "fatal-fail",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "msg", Type: types.FieldTypeString, Required: false}},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{{Name: "required", Type: types.FieldTypeString, Required: true}},
		},
	}
	if err := orch.RegisterContract(failingContract); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("Failed to start orchestrator: %v", err)
	}

	// fatal failure class with max_retries=3: should not retry, immediate fail
	task := &types.AgentTask{
		TaskID:     "fatal-task",
		ContractID: "fatal-fail",
		Input:      map[string]any{"msg": "test"},
		Metadata: map[string]string{
			types.SLAKeyMaxRetries:   "3",
			types.SLAKeyFailureClass: types.FailureClassFatal,
		},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusFailed)
	if status != types.TaskStatusFailed {
		t.Errorf("Expected fatal task to fail, got %s", status)
	}
}

func TestOrchestrator_DegradableFailureClass_CompletesNormally(t *testing.T) {
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

	// degradable task that completes normally
	task := &types.AgentTask{
		TaskID:     "degradable-ok",
		ContractID: "default",
		Input:      map[string]any{"message": "test"},
		Metadata: map[string]string{
			types.SLAKeyFailureClass: types.FailureClassDegradable,
		},
	}
	if err := orch.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	status := waitForTaskStatus(t, orch, task.TaskID, types.TaskStatusCompleted)
	if status != types.TaskStatusCompleted {
		t.Errorf("Expected degradable task to complete, got %s", status)
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

type capturingToolsProvider struct {
	toolNames []string
	called    chan struct{}
}

func (p *capturingToolsProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	for _, def := range req.Tools {
		p.toolNames = append(p.toolNames, def.Name)
	}
	select {
	case p.called <- struct{}{}:
	default:
	}
	return &provider.ModelResponse{Output: map[string]any{"status": "completed"}}, nil
}

type fileWriteToolCallProvider struct {
	path    string
	content string
	called  bool
}

func (p *fileWriteToolCallProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	if !p.called {
		p.called = true
		return &provider.ModelResponse{
			ToolCalls: []types.ToolCall{
				{ID: "call-file-write", Name: "file_write", Input: map[string]any{"path": p.path, "content": p.content}},
			},
		}, nil
	}
	return &provider.ModelResponse{Output: map[string]any{"status": "completed"}}, nil
}
