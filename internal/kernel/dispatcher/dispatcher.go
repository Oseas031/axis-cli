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
	"github.com/axis-cli/axis/internal/evolution"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	"github.com/axis-cli/axis/internal/kernel/capability"
	"github.com/axis-cli/axis/internal/kernel/featuregate"
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
	capRegistry      *capability.CapabilityRegistry
	gate             *featuregate.Gate
	evolutionStore   *evolution.Store
	projectRoot      string
	timeout          time.Duration
	autonomyResolver func(task *types.AgentTask) agent.AutonomyLevel
	auditLog         []AuditEntry
	auditMu          sync.RWMutex
	auditFn              func(taskID, event, detail string)
	followUpFn           func(tasks []*types.AgentTask)
	candidatePoolEnabled bool
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
// Checks "agent.autonomy_level" first, then falls back to "autonomy_level".
func DefaultAutonomyResolver(task *types.AgentTask) agent.AutonomyLevel {
	if task.Metadata != nil {
		val, ok := task.Metadata["agent.autonomy_level"]
		if !ok {
			val, ok = task.Metadata["autonomy_level"]
		}
		if ok {
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

// SetCapabilityRegistry sets the capability registry for capability checks.
func (d *DispatcherImpl) SetCapabilityRegistry(reg *capability.CapabilityRegistry) {
	d.capRegistry = reg
}

// SetAuditFunc sets an optional external audit callback.
func (d *DispatcherImpl) SetAuditFunc(fn func(taskID, event, detail string)) {
	d.auditFn = fn
}

// SetFollowUpHandler sets a callback for follow-up task submission.
func (d *DispatcherImpl) SetFollowUpHandler(fn func(tasks []*types.AgentTask)) {
	d.followUpFn = fn
}

// SetCandidatePoolEnabled enables candidate pool evaluation for high-risk tasks.
func (d *DispatcherImpl) SetCandidatePoolEnabled(enabled bool) {
	d.candidatePoolEnabled = enabled
}

// SetFeatureGate sets the feature gate for pre-execution checks.
func (d *DispatcherImpl) SetFeatureGate(g *featuregate.Gate) {
	d.gate = g
}

// SetEvolutionStore sets the evolution store for evolution protocol routing.
func (d *DispatcherImpl) SetEvolutionStore(store *evolution.Store, projectRoot string) {
	d.evolutionStore = store
	d.projectRoot = projectRoot
}

// checkFeatureGate verifies all required features are unlocked.
func (d *DispatcherImpl) checkFeatureGate(task *types.AgentTask) error {
	if d.gate == nil || task.Metadata == nil {
		return nil
	}
	raw, ok := task.Metadata["axis.required_features"]
	if !ok || raw == "" {
		return nil
	}
	for _, f := range strings.Split(raw, ",") {
		feat := featuregate.Feature(strings.TrimSpace(f))
		if !d.gate.IsUnlocked(feat) {
			return fmt.Errorf("feature %q is locked", feat)
		}
	}
	return nil
}

func (d *DispatcherImpl) audit(taskID, event, detail string) {
	if d.auditFn != nil {
		d.auditFn(taskID, event, detail)
	}
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
	executorType := "contract"
	if task.Metadata != nil {
		if v, ok := task.Metadata[types.TaskMetadataKeyExecutor]; ok {
			executorType = v
		}
	}
	d.audit(task.TaskID, "dispatch_start", executorType)

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
		d.audit(task.TaskID, "dispatch_end", string(result.Status))
		return result, timeoutErr
	case result := <-resultChan:
		d.recordAudit(start, task, result)
		d.audit(task.TaskID, "dispatch_end", string(result.Status))
		return result, nil
	case err := <-errChan:
		result := &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     err.Error(),
			Completed: time.Now(),
		}
		d.recordAudit(start, task, result)
		d.audit(task.TaskID, "dispatch_end", string(result.Status))
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
	if err := d.checkFeatureGate(task); err != nil {
		return &types.TaskResult{
			TaskID:    task.TaskID,
			Status:    types.TaskStatusFailed,
			Error:     err.Error(),
			Completed: time.Now(),
		}, err
	}

	if d.isEvolutionRequired(task) {
		return d.executeViaEvolution(ctx, task)
	}

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
	if d.capRegistry != nil {
		if reqCaps, ok := task.Metadata["axis.required_capabilities"]; ok && reqCaps != "" {
			for _, name := range strings.Split(reqCaps, ",") {
				name = strings.TrimSpace(name)
				if _, found := d.capRegistry.Get(name); !found {
					errMsg := fmt.Sprintf("required capability not registered: %s", name)
					return &types.TaskResult{
						TaskID:    task.TaskID,
						Status:    types.TaskStatusFailed,
						Error:     errMsg,
						Completed: time.Now(),
					}, fmt.Errorf("required capability not registered: %s", name)
				}
			}
		}
	} else {
		log.Printf("[dispatcher] scope_check task=%s: permission scopes not enforced (v1). TODO: check against autonomy level", task.TaskID)
	}

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

	// Submit follow-up tasks if any were generated
	if d.followUpFn != nil && len(agentResult.FollowUpTasks) > 0 {
		d.followUpFn(agentResult.FollowUpTasks)
		d.audit(task.TaskID, "followup_generated", fmt.Sprintf("%d tasks", len(agentResult.FollowUpTasks)))
	}

	// CandidatePool evaluation for high-risk tasks
	// v1: single-candidate mode. TODO: execute with multiple providers and use CandidatePool.Partition() to select dominant result
	if d.candidatePoolEnabled && task.Metadata != nil && task.Metadata["axis.candidate_pool"] == "true" {
		log.Printf("[dispatcher] candidate_pool task=%s: single-candidate mode (v1). TODO: multi-provider differential testing", task.TaskID)
		d.audit(task.TaskID, "candidate_pool_triggered", "single-candidate v1")
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

// isEvolutionRequired checks whether a task should be routed through the evolution protocol.
func (d *DispatcherImpl) isEvolutionRequired(task *types.AgentTask) bool {
	if d.evolutionStore == nil || d.gate == nil {
		return false
	}
	if !d.gate.IsUnlocked(featuregate.FeatureEvolution) {
		return false
	}
	if task.Metadata == nil {
		return false
	}
	return task.Metadata["axis.evolution_required"] == "true"
}

// executeViaEvolution wraps task execution with the staged evolution protocol.
func (d *DispatcherImpl) executeViaEvolution(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
	runID := evolution.GenerateRunID()

	intent := evolution.EvolutionIntent{
		ID:           runID,
		CreatedAt:    time.Now(),
		Actor:        "dispatcher",
		Summary:      fmt.Sprintf("evolution for task %s", task.TaskID),
		TargetDomain: task.Metadata["axis.evolution_domain"],
		RiskLevel:    evolution.RiskLevel(task.Metadata["axis.evolution_risk"]),
		Status:       evolution.StatusRunning,
	}
	if intent.RiskLevel == "" {
		intent.RiskLevel = evolution.RiskLow
	}

	run := evolution.EvolutionRun{
		RunID:         runID,
		IntentID:      runID,
		Status:        evolution.StatusRunning,
		CreatedAt:     time.Now(),
		WorkspacePath: d.evolutionStore.RunDir(runID),
	}

	if err := d.evolutionStore.CreateRun(intent, run); err != nil {
		log.Printf("[dispatcher] evolution store CreateRun failed: %v; falling back to normal execution", err)
		return d.executeTaskNormal(ctx, task)
	}

	// Inject evolution metadata into task
	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}
	task.Metadata["evolution.run_id"] = runID
	task.Metadata["evolution.workspace"] = run.WorkspacePath

	d.audit(task.TaskID, "evolution_start", runID)

	result, err := d.executeTaskNormal(ctx, task)

	if result != nil {
		if result.Output == nil {
			result.Output = make(map[string]any)
		}
		result.Output["evolution.run_id"] = runID
	}

	d.audit(task.TaskID, "evolution_end", runID)
	return result, err
}

// executeTaskNormal runs the normal executor routing (agent/contract/human).
func (d *DispatcherImpl) executeTaskNormal(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
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
