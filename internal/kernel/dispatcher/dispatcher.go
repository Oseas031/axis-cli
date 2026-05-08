// Package dispatcher provides task dispatching to executors.
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
	timeoutCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	resultChan := make(chan *types.TaskResult, 1)
	errChan := make(chan error, 1)

	// Use a separate context for the goroutine to avoid goroutine leak
	go func() {
		// Check if parent context is cancelled before starting
		select {
		case <-ctx.Done():
			return
		case <-timeoutCtx.Done():
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
	case <-timeoutCtx.Done():
		timeoutErr := types.NewAgentError(types.ErrTaskTimeout, fmt.Sprintf("task %s timed out", task.TaskID))
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     timeoutErr.Error(),
			Completed: time.Now(),
		}, timeoutErr
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

// executeTask executes a task by routing through the contract executor.
func (d *DispatcherImpl) executeTask(task *types.AgentTask) (*types.TaskResult, error) {
	execResult, err := d.contractExecutor.Execute(task.ContractID, task.Input)
	if err != nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     execResult.Error,
			Completed: time.Now(),
		}, err
	}

	return &types.TaskResult{
		TaskID:    task.TaskID,
		Output:    execResult.Output,
		Status:    types.TaskStatusCompleted,
		Completed: time.Now(),
	}, nil
}
