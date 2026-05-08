package dispatcher

import (
	"context"
	"testing"
	"time"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
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
