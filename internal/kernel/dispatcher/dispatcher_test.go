package dispatcher

import (
	"context"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/agent"
	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/types"
)

func TestDispatcher_NewDispatcher(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	if dispatch == nil {
		t.Fatal("NewDispatcher should return non-nil")
	}
}

func TestDispatcher_Dispatch(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	// Register a contract
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
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "test-contract",
		Input:      map[string]any{"name": "test"},
	}

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)

	if err != nil {
		t.Errorf("Dispatch should succeed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.TaskID != task.TaskID {
		t.Errorf("Expected task ID %s, got %s", task.TaskID, result.TaskID)
	}
}

func TestDispatcher_DispatchInvalidInput(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	// Register a contract with required field
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
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "test-contract",
		Input:      map[string]any{}, // Missing required field
	}

	ctx := context.Background()
	result, _ := dispatch.Dispatch(ctx, task)

	// Should return result with error status
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected status %s, got %s", types.TaskStatusFailed, result.Status)
	}
}

func TestDispatcher_DispatchParentContextCancelled(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "name", Type: types.FieldTypeString, Required: true}},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "ctx-cancelled",
		ContractID: "test-contract",
		Input:      map[string]any{"name": "test"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := dispatch.Dispatch(ctx, task)
	if err == nil {
		t.Fatal("Dispatch should return error when parent context is cancelled")
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected failed status, got %s", result.Status)
	}
}

func TestDispatcher_DispatchErrChan(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	// Set a provider so Execute goes through the full path
	contractExec.SetProvider(provider.NewMockModelProvider())

	// Contract whose output schema requires a field the mock won't provide
	contract := &types.AgentContract{
		ContractID: "err-chan",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{{Name: "missing_field", Type: types.FieldTypeString, Required: true}},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "err-chan-task",
		ContractID: "err-chan",
		Input:      map[string]any{"x": "y"},
	}

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)
	if err == nil {
		t.Fatal("Dispatch should return error when executeTask fails output validation")
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected failed status, got %s", result.Status)
	}
}

func TestDispatcher_DispatchTimeout(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()

	// Create dispatcher with very short timeout
	dispatch := NewDispatcher(contractExec, humanExec)
	dispatch.timeout = 10 * time.Millisecond

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
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "test-contract",
		Input:      map[string]any{"name": "test"},
	}

	ctx := context.Background()
	result, _ := dispatch.Dispatch(ctx, task)

	// Should still succeed quickly (task execution is fast)

	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestDispatcher_HumanExecutorRoute(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)
	dispatch.timeout = 500 * time.Millisecond

	task := &types.AgentTask{
		TaskID:     "human-task",
		ContractID: "any",
		Input:      map[string]any{"prompt": "hello"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeHuman},
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		humanExec.ResolveCall("human-task", map[string]any{"answer": "hi"}, nil)
	}()

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Human dispatch should succeed after resolution: %v", err)
	}
	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected completed, got %s", result.Status)
	}
}

func TestDispatcher_HumanExecutorTimeout(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)
	dispatch.timeout = 100 * time.Millisecond

	task := &types.AgentTask{
		TaskID:     "human-timeout",
		ContractID: "any",
		Input:      map[string]any{"prompt": "hello"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeHuman},
	}

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)
	if err == nil {
		t.Fatal("Human dispatch should timeout")
	}
	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected failed status, got %s", result.Status)
	}
}

func TestDispatcher_SetAgentExecutor(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	if dispatcher.agentExecutor == nil {
		t.Error("agentExecutor should be set after SetAgentExecutor")
	}
}

func TestDispatcher_Dispatch_AgentExecutor(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-task-1",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test input",
			"task_type": "default",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestDispatcher_Dispatch_AgentExecutorNotConfigured(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	// Don't set agent executor

	task := &types.AgentTask{
		TaskID:     "agent-task-unconfigured",
		ContractID: "test-contract",
		Input: map[string]any{
			"input": "test input",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err == nil {
		t.Error("Expected error when agent executor not configured")
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected status failed, got %s", result.Status)
	}

	if result.Error == "" {
		t.Error("Error should be set when agent executor not configured")
	}
}

func TestDispatcher_Dispatch_AgentExecutorCodeGen(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-code-gen",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "generate code",
			"task_type": "code_generation",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestDispatcher_Dispatch_AgentExecutorDebugging(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-debug",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "debug issue",
			"task_type": "debugging",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}
