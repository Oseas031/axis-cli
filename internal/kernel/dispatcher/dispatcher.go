// Package dispatcher provides task dispatching to executors.
package dispatcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/axis-cli/axis/internal/agent"
	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	"github.com/axis-cli/axis/internal/types"
)

// Dispatcher interface defines task dispatching to executors
type Dispatcher interface {
	Dispatch(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error)
	SetAgentExecutor(e agent.AgentExecutor)
}

// DispatcherImpl implements task dispatching
type DispatcherImpl struct {
	contractExecutor contractexec.ContractExecutor
	humanExecutor    humanexec.HumanExecutor
	agentExecutor    agent.AgentExecutor
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

// SetAgentExecutor sets the agent executor for agent-based task execution.
func (d *DispatcherImpl) SetAgentExecutor(e agent.AgentExecutor) {
	d.agentExecutor = e
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
		// Capture any late-arriving goroutine error
		select {
		case err := <-errChan:
			log.Printf("Task %s timed out; underlying error: %v", task.TaskID, err)
		default:
		}
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

// executeTask executes a task by routing to the appropriate executor.
func (d *DispatcherImpl) executeTask(task *types.AgentTask) (*types.TaskResult, error) {
	executorType := task.Metadata[types.TaskMetadataKeyExecutor]

	if executorType == types.ExecutorTypeHuman {
		return d.executeHumanTask(task)
	}

	if executorType == types.ExecutorTypeAgent {
		return d.executeAgentTask(task)
	}

	execResult, err := d.contractExecutor.Execute(task.ContractID, task.Input)
	if err != nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     execResult.Error,
			Completed: time.Now(),
		}, fmt.Errorf("dispatch: contract %s execute: %w", task.ContractID, err)
	}

	return &types.TaskResult{
		TaskID:    task.TaskID,
		Output:    execResult.Output,
		Status:    types.TaskStatusCompleted,
		Completed: time.Now(),
	}, nil
}

// executeAgentTask routes a task to the agent executor.
func (d *DispatcherImpl) executeAgentTask(task *types.AgentTask) (*types.TaskResult, error) {
	if d.agentExecutor == nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     "agent executor not configured",
			Completed: time.Now(),
		}, fmt.Errorf("agent executor not configured")
	}

	selfContext := agent.NewSelfContext(task.TaskID)

	agentReq := &agent.AgentExecutionRequest{
		Task:        task,
		SelfContext: selfContext,
		Autonomy:    agent.AutonomyLevelLow,
	}

	agentResult, err := d.agentExecutor.Execute(context.Background(), agentReq)
	if err != nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     fmt.Sprintf("agent execution failed: %v", err),
			Completed: time.Now(),
		}, fmt.Errorf("agent execution failed: %w", err)
	}

	if agentResult.Error != "" {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Output:    agentResult.Output,
			Status:    types.TaskStatusFailed,
			Error:     agentResult.Error,
			Completed: time.Now(),
		}, fmt.Errorf("agent execution error: %s", agentResult.Error)
	}

	return &types.TaskResult{
		TaskID:    task.TaskID,
		Output:    agentResult.Output,
		Status:    types.TaskStatusCompleted,
		Completed: time.Now(),
	}, nil
}

// executeHumanTask routes a task to the human executor and polls until resolved or timed out.
func (d *DispatcherImpl) executeHumanTask(task *types.AgentTask) (*types.TaskResult, error) {
	callReq := &types.HumanCallRequest{
		CallID:   task.TaskID,
		TaskID:   task.TaskID,
		Input:    task.Input,
		Metadata: task.Metadata,
	}

	if _, err := d.humanExecutor.ExecuteCall(callReq); err != nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     fmt.Sprintf("human call failed: %v", err),
			Completed: time.Now(),
		}, err
	}

	pollInterval := 100 * time.Millisecond
	deadline := time.Now().Add(d.timeout)

	for {
		status, err := d.humanExecutor.GetCallStatus(task.TaskID)
		if err != nil {
			return &types.TaskResult{
				TaskID:    task.TaskID,
				Status:    types.TaskStatusFailed,
				Error:     fmt.Sprintf("call status check failed: %v", err),
				Completed: time.Now(),
			}, fmt.Errorf("call status check failed: %v", err)
		}

		switch status {
		case types.CallStatusCompleted:
			return &types.TaskResult{
				TaskID:    task.TaskID,
				Status:    types.TaskStatusCompleted,
				Completed: time.Now(),
			}, nil
		case types.CallStatusFailed:
			return &types.TaskResult{
				TaskID:    task.TaskID,
				Status:    types.TaskStatusFailed,
				Error:     "human call failed",
				Completed: time.Now(),
			}, fmt.Errorf("human call %s failed", task.TaskID)
		}

		if time.Now().After(deadline) {
			timeoutErr := types.NewAgentError(types.ErrTaskTimeout, fmt.Sprintf("human call %s timed out waiting for resolution", task.TaskID))
			return &types.TaskResult{
				TaskID:    task.TaskID,
				Status:    types.TaskStatusFailed,
				Error:     timeoutErr.Error(),
				Completed: time.Now(),
			}, timeoutErr
		}

		time.Sleep(pollInterval)
	}
}
