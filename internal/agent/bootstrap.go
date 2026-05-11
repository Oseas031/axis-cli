// Package agent provides bootstrap loop integration for autonomous agent execution.
package agent

import (
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/agent/judgement"
	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
	"github.com/axis-cli/axis/internal/types"
)

// SchedulerAccess is the interface for scheduler operations needed by BootstrapLoop.
type SchedulerAccess interface {
	SubmitTask(task *types.AgentTask) error
	GetAllTasks() []*types.AgentTask
	GetDependencyGraph() map[string][]string
}

// BootstrapLoop is the interface for bootstrap loop management.
// It manages the self-iteration loop that allows agents to spawn follow-up tasks.
type BootstrapLoop interface {
	// SubmitSelfIterationTask submits a task with self-context injection and loop tracking.
	SubmitSelfIterationTask(task *types.AgentTask) error
	// TrackIteration increments and returns the iteration count for a task.
	TrackIteration(taskID string) int
	// GetIterationCount returns the current iteration count for a task.
	GetIterationCount(taskID string) int
	// IsIterationAllowed checks if a task can proceed with another iteration.
	IsIterationAllowed(taskID string) bool
	// ResetIteration resets the iteration count for a task.
	ResetIteration(taskID string)
	// GetLoopStatus returns the loop tracking status for all tasks.
	GetLoopStatus() map[string]int
	// GetAllTasks returns all tasks from the scheduler.
	GetAllTasks() []*types.AgentTask
	// GetDependencyGraph returns the dependency graph from the scheduler.
	GetDependencyGraph() map[string][]string
}

// bootstrapOrchestrator is the internal implementation of BootstrapLoop.
type bootstrapOrchestrator struct {
	scheduler       SchedulerAccess
	loopTracking    map[string]int
	maxIterations   int
	judgementEngine *judgement.Engine
	mu              sync.RWMutex
}

// BootstrapOption configures a bootstrapOrchestrator.
type BootstrapOption func(*bootstrapOrchestrator)

// WithJudgementEngine sets the judgement engine for the bootstrap orchestrator.
func WithJudgementEngine(engine *judgement.Engine) BootstrapOption {
	return func(bo *bootstrapOrchestrator) {
		bo.judgementEngine = engine
	}
}

// NewBootstrapOrchestrator creates a new BootstrapLoop implementation with the given scheduler and max iterations.
func NewBootstrapOrchestrator(scheduler SchedulerAccess, maxIterations int, opts ...BootstrapOption) BootstrapLoop {
	bo := &bootstrapOrchestrator{
		scheduler:     scheduler,
		loopTracking:  make(map[string]int),
		maxIterations: maxIterations,
	}
	for _, opt := range opts {
		opt(bo)
	}
	return bo
}

// SubmitSelfIterationTask submits a task with self-context injection and loop tracking.
func (bo *bootstrapOrchestrator) SubmitSelfIterationTask(task *types.AgentTask) error {
	bo.mu.Lock()
	defer bo.mu.Unlock()

	// 1. Check if loop count exceeds maxIterations
	loopCount := bo.loopTracking[task.TaskID]
	if loopCount >= bo.maxIterations {
		return fmt.Errorf("task %s exceeded max iterations (%d)", task.TaskID, bo.maxIterations)
	}

	// 2. Inject SelfContext to task.Metadata
	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}

	// Create self context metadata
	selfCtx := bo.buildSelfContextMetadata(task)
	for k, v := range selfCtx {
		task.Metadata[k] = v
	}

	// 3. Submit to Scheduler
	if err := bo.scheduler.SubmitTask(task); err != nil {
		return fmt.Errorf("submit self-iteration task %s: %w", task.TaskID, err)
	}

	return nil
}

// buildSelfContextMetadata builds metadata from self context for task injection.
func (bo *bootstrapOrchestrator) buildSelfContextMetadata(task *types.AgentTask) map[string]string {
	metadata := make(map[string]string)

	// Inject loop iteration count
	metadata["self.iteration"] = fmt.Sprintf("%d", bo.loopTracking[task.TaskID]+1)
	metadata["self.max_iterations"] = fmt.Sprintf("%d", bo.maxIterations)

	// Inject parent task ID if this is a follow-up
	if task.Metadata != nil {
		if parentID, ok := task.Metadata["parent_task_id"]; ok {
			metadata["self.parent_task_id"] = parentID
		}
	}

	// Build lineage chain
	lineage := task.TaskID
	if deps := task.Dependencies; len(deps) > 0 {
		// Get the immediate parent (first dependency)
		lineage = deps[0] + " -> " + lineage
	}
	metadata["self.lineage"] = lineage

	return metadata
}

// TrackIteration increments and returns the iteration count for a task.
func (bo *bootstrapOrchestrator) TrackIteration(taskID string) int {
	bo.mu.Lock()
	defer bo.mu.Unlock()
	bo.loopTracking[taskID]++
	return bo.loopTracking[taskID]
}

// GetIterationCount returns the current iteration count for a task.
func (bo *bootstrapOrchestrator) GetIterationCount(taskID string) int {
	bo.mu.RLock()
	defer bo.mu.RUnlock()
	return bo.loopTracking[taskID]
}

// IsIterationAllowed checks if a task can proceed with another iteration.
func (bo *bootstrapOrchestrator) IsIterationAllowed(taskID string) bool {
	bo.mu.RLock()
	defer bo.mu.RUnlock()
	return bo.loopTracking[taskID] < bo.maxIterations
}

// ResetIteration resets the iteration count for a task.
func (bo *bootstrapOrchestrator) ResetIteration(taskID string) {
	bo.mu.Lock()
	defer bo.mu.Unlock()
	delete(bo.loopTracking, taskID)
}

// GetLoopStatus returns the loop tracking status for all tasks.
func (bo *bootstrapOrchestrator) GetLoopStatus() map[string]int {
	bo.mu.RLock()
	defer bo.mu.RUnlock()
	status := make(map[string]int, len(bo.loopTracking))
	for k, v := range bo.loopTracking {
		status[k] = v
	}
	return status
}

// GetAllTasks returns all tasks from the scheduler.
func (bo *bootstrapOrchestrator) GetAllTasks() []*types.AgentTask {
	return bo.scheduler.GetAllTasks()
}

// GetDependencyGraph returns the dependency graph from the scheduler.
func (bo *bootstrapOrchestrator) GetDependencyGraph() map[string][]string {
	return bo.scheduler.GetDependencyGraph()
}

// JudgeExecutionResult performs self-judgement on an execution result using the configured judgement engine.
// Returns nil if no judgement engine is configured.
func (bo *bootstrapOrchestrator) JudgeExecutionResult(result *AgentExecutionResult) (*judgement.JudgementResult, error) {
	bo.mu.RLock()
	engine := bo.judgementEngine
	bo.mu.RUnlock()

	if engine == nil {
		return nil, nil
	}

	criteria := bo.defaultJudgementCriteria()
	return engine.Judge(result, criteria)
}

// CalculateAutonomyDelta computes an autonomy adjustment based on the judgement result.
// Positive delta means earned autonomy; negative means lost.
func (bo *bootstrapOrchestrator) CalculateAutonomyDelta(jr *judgement.JudgementResult) AutonomyDelta {
	if jr == nil {
		return AutonomyDelta{Delta: 0, Reason: "no judgement performed"}
	}

	if jr.Passed {
		if jr.Score >= 0.95 && jr.Confidence >= 0.90 {
			return AutonomyDelta{Delta: 2, Reason: fmt.Sprintf("excellent judgement: score %.2f, confidence %.2f", jr.Score, jr.Confidence)}
		}
		return AutonomyDelta{Delta: 1, Reason: fmt.Sprintf("passed judgement: score %.2f, confidence %.2f", jr.Score, jr.Confidence)}
	}

	if jr.Score >= 0.50 {
		return AutonomyDelta{Delta: 0, Reason: fmt.Sprintf("marginal judgement: score %.2f, needs improvement", jr.Score)}
	}
	return AutonomyDelta{Delta: -1, Reason: fmt.Sprintf("failed judgement: score %.2f, confidence %.2f", jr.Score, jr.Confidence)}
}

// EvaluateAndDecide runs judgement on the execution result and computes the corresponding autonomy delta.
// It mutates result.JudgementResult and result.AutonomyDelta in place.
func (bo *bootstrapOrchestrator) EvaluateAndDecide(result *AgentExecutionResult) error {
	if result == nil {
		return fmt.Errorf("execution result is nil")
	}

	jr, err := bo.JudgeExecutionResult(result)
	if err != nil {
		return fmt.Errorf("judgement failed: %w", err)
	}

	result.JudgementResult = jr
	result.AutonomyDelta = bo.CalculateAutonomyDelta(jr)
	return nil
}

// defaultJudgementCriteria returns the default criteria set for bootstrap self-judgement.
func (bo *bootstrapOrchestrator) defaultJudgementCriteria() []strategies.JudgementCriteria {
	return []strategies.JudgementCriteria{
		{
			Name:    "syntax_check",
			Type:    strategies.JudgementTypeSyntax,
			Weight:  0.20,
			Enabled: true,
			Thresholds: map[string]float64{
				"min_pass_rate": 1.0,
			},
		},
		{
			Name:    "test_pass_rate",
			Type:    strategies.JudgementTypeTest,
			Weight:  0.40,
			Enabled: true,
			Thresholds: map[string]float64{
				"min_pass_rate": judgement.DefaultJudgementThresholds.MinTestPassRate,
			},
		},
		{
			Name:    "coverage_threshold",
			Type:    strategies.JudgementTypeCoverage,
			Weight:  0.40,
			Enabled: true,
			Thresholds: map[string]float64{
				"min_coverage": judgement.DefaultJudgementThresholds.MinCoverage,
			},
		},
	}
}
