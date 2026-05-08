package dispatcher

import (
	"context"
	"fmt"
	"time"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	"github.com/axis-cli/axis/internal/types"
)

// Dispatcher interface defines task dispatching to executors
type Dispatcher interface {
	Dispatch(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error)
}

// DispatcherImpl implements task dispatching
type DispatcherImpl struct {
	contractExecutor contractexec.ContractExecutor
	humanExecutor    humanexec.HumanExecutor
	timeout          time.Duration
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(contractExec contractexec.ContractExecutor, humanExec humanexec.HumanExecutor) *DispatcherImpl {
	return &DispatcherImpl{
		contractExecutor: contractExec,
		humanExecutor:    humanExec,
		timeout:          30 * time.Minute, // Default timeout for milestone 1
	}
}

// Dispatch dispatches a task to the appropriate executor
func (d *DispatcherImpl) Dispatch(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	resultChan := make(chan *types.TaskResult, 1)
	errChan := make(chan error, 1)

	// Use a separate context for the goroutine to avoid goroutine leak
	go func() {
		// Check if parent context is cancelled before starting
		select {
		case <-ctx.Done():
			return
		default:
			result, err := d.executeTask(task)
			if err != nil {
				errChan <- err
			} else {
				resultChan <- result
			}
		}
	}()

	select {
	case <-ctx.Done():
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     "task execution timed out",
			Completed: time.Now(),
		}, ctx.Err()
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     err.Error(),
			Completed: time.Now(),
		}, err
	}
}

// executeTask executes a task based on its contract type
func (d *DispatcherImpl) executeTask(task *types.AgentTask) (*types.TaskResult, error) {
	// For milestone 1, we use contract executor for validation
	// In future milestones, this would route to different executors based on contract type

	// Validate input
	if err := d.contractExecutor.ValidateInput(task.ContractID, task.Input); err != nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     fmt.Sprintf("input validation failed: %v", err),
			Completed: time.Now(),
		}, err
	}

	// For milestone 1, we return a simple success result
	// In future milestones, this would execute the actual agent logic
	return &types.TaskResult{
		TaskID:    task.TaskID,
		Output:    map[string]any{"status": "completed", "message": "task executed successfully"},
		Status:    types.TaskStatusCompleted,
		Completed: time.Now(),
	}, nil
}
