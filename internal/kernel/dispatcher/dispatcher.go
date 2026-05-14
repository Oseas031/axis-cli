// Package dispatcher provides task dispatching to executors.
package dispatcher

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/agent"
	"github.com/axis-cli/axis/internal/contextpack"
	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	"github.com/axis-cli/axis/internal/types"
)

// AuditEntry records a single dispatch event.
type AuditEntry struct {
	Timestamp    time.Time
	TaskID       string
	ExecutorType string // "contract", "agent", "human"
	Duration     time.Duration
	Status       string // "completed", "failed", "timeout"
	Error        string
}

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
	autonomyResolver func(task *types.AgentTask) agent.AutonomyLevel
	auditLog         []AuditEntry
	auditMu          sync.RWMutex
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(contractExec contractexec.ContractExecutor, humanExec humanexec.HumanExecutor) *DispatcherImpl {
	return &DispatcherImpl{
		contractExecutor: contractExec,
		humanExecutor:    humanExec,
		timeout:          30 * time.Minute, // Default timeout for milestone 1
		autonomyResolver: DefaultAutonomyResolver,
	}
}

// DefaultAutonomyResolver returns AutonomyLevelLow unless task metadata specifies otherwise.
func DefaultAutonomyResolver(task *types.AgentTask) agent.AutonomyLevel {
	if task.Metadata != nil {
		if val, ok := task.Metadata["autonomy_level"]; ok {
			switch val {
			case "none":
				return agent.AutonomyLevelNone
			case "low":
				return agent.AutonomyLevelLow
			case "medium":
				return agent.AutonomyLevelMedium
			case "high":
				return agent.AutonomyLevelHigh
			case "execute":
				return agent.AutonomyLevelExecute
			case "decide":
				return agent.AutonomyLevelDecide
			case "plan":
				return agent.AutonomyLevelPlan
			case "learn":
				return agent.AutonomyLevelLearn
			case "full":
				return agent.AutonomyLevelFull
			}
		}
	}
	return agent.AutonomyLevelLow
}

// WithAutonomyResolver sets a custom autonomy resolver function.
func (d *DispatcherImpl) WithAutonomyResolver(fn func(*types.AgentTask) agent.AutonomyLevel) {
	d.autonomyResolver = fn
}

// SetAgentExecutor sets the agent executor for agent-based task execution.
func (d *DispatcherImpl) SetAgentExecutor(e agent.AgentExecutor) {
	d.agentExecutor = e
}

// AuditLog returns a copy of the current audit log entries.
func (d *DispatcherImpl) AuditLog() []AuditEntry {
	d.auditMu.RLock()
	defer d.auditMu.RUnlock()
	out := make([]AuditEntry, len(d.auditLog))
	copy(out, d.auditLog)
	return out
}

// Dispatch dispatches a task to the appropriate executor
func (d *DispatcherImpl) Dispatch(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
	start := time.Now()
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
			result, err := d.executeTask(timeoutCtx, task)
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
		result := &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     timeoutErr.Error(),
			Completed: time.Now(),
		}
		d.recordAudit(start, task, result)
		return result, timeoutErr
	case result := <-resultChan:
		d.recordAudit(start, task, result)
		return result, nil
	case err := <-errChan:
		result := &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     err.Error(),
			Completed: time.Now(),
		}
		d.recordAudit(start, task, result)
		return result, err
	}
}

// recordAudit appends an audit entry for a completed dispatch.
func (d *DispatcherImpl) recordAudit(start time.Time, task *types.AgentTask, result *types.TaskResult) {
	executorType := "contract"
	if task.Metadata != nil {
		switch task.Metadata[types.TaskMetadataKeyExecutor] {
		case types.ExecutorTypeAgent:
			executorType = "agent"
		case types.ExecutorTypeHuman:
			executorType = "human"
		}
	}
	status := "completed"
	if result.Status == types.TaskStatusFailed {
		status = "failed"
		if result.Error != "" && strings.Contains(result.Error, string(types.ErrTaskTimeout)) {
			status = "timeout"
		}
	}
	entry := AuditEntry{
		Timestamp:    start,
		TaskID:       task.TaskID,
		ExecutorType: executorType,
		Duration:     time.Since(start),
		Status:       status,
		Error:        result.Error,
	}
	d.auditMu.Lock()
	d.auditLog = append(d.auditLog, entry)
	d.auditMu.Unlock()
}

// executeTask executes a task by routing to the appropriate executor.
func (d *DispatcherImpl) executeTask(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
	executorType := task.Metadata[types.TaskMetadataKeyExecutor]

	if executorType == types.ExecutorTypeHuman {
		return d.executeHumanTask(ctx, task)
	}

	if executorType == types.ExecutorTypeAgent {
		return d.executeAgentTask(ctx, task)
	}

	execResult, err := d.contractExecutor.Execute(ctx, task.ContractID, task.Input)
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
func (d *DispatcherImpl) executeAgentTask(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
	if d.agentExecutor == nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     "agent executor not configured",
			Completed: time.Now(),
		}, fmt.Errorf("agent executor not configured")
	}

	selfContext := agent.NewSelfContext(task.TaskID)
	summary := executionContextSummary(task)

	// RequestedSources is populated from the summary rather than parsed again
	// to avoid duplicate work. The summary already parsed the metadata once;
	// dispatcher only carries it forward to the executor.
	agentReq := &agent.AgentExecutionRequest{
		Task:             task,
		SelfContext:      selfContext,
		Autonomy:         d.autonomyResolver(task),
		ContextSummary:   summary,
		RequestedSources: summary.RequestedSources,
	}

	agentResult, err := d.agentExecutor.Execute(ctx, agentReq)
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

func executionContextSummary(task *types.AgentTask) *contextpack.ExecutionContextSummary {
	summary := contextpack.NewExecutionContextConsumer(contextpack.DefaultRegistry).Summarize(task)
	return &summary
}

// executeHumanTask routes a task to the human executor and polls until resolved or timed out.
func (d *DispatcherImpl) executeHumanTask(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
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
		select {
		case <-ctx.Done():
			return &types.TaskResult{
				TaskID:    task.TaskID,
				Status:    types.TaskStatusFailed,
				Error:     ctx.Err().Error(),
				Completed: time.Now(),
			}, ctx.Err()
		default:
		}
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
