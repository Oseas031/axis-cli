// Package orchestrator provides task orchestration and coordination.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/agent"
	"github.com/axis-cli/axis/internal/contract/admission"
	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	dispatcher "github.com/axis-cli/axis/internal/kernel/dispatcher"
	"github.com/axis-cli/axis/internal/kernel/lifecycle"
	"github.com/axis-cli/axis/internal/kernel/scheduler"
	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/project"
	"github.com/axis-cli/axis/internal/skills"
	"github.com/axis-cli/axis/internal/types"
)

const defaultWorkerLimit = 5

// OrchestratorOption is a functional option for Orchestrator construction.
type OrchestratorOption func(*Orchestrator)

// WithModelProvider sets the ModelProvider for contract execution.
func WithModelProvider(p provider.ModelProvider) OrchestratorOption {
	return func(o *Orchestrator) {
		o.contractExecutor.SetProvider(p)
	}
}

// WithAgentExecutor sets the AgentExecutor for agent-based task execution.
func WithAgentExecutor(e agent.AgentExecutor) OrchestratorOption {
	return func(o *Orchestrator) {
		o.agentExecutor = e
	}
}

// Orchestrator coordinates all kernel modules
type Orchestrator struct {
	stateStore         sharedlayer.StateStore
	lifecycleManager   *lifecycle.LifecycleManagerImpl
	scheduler          *scheduler.SchedulerImpl
	dispatcher         *dispatcher.DispatcherImpl
	contractExecutor   *contractexec.ContractExecutorImpl
	admissionValidator *admission.AdmissionValidatorImpl
	humanExecutor      *humanexec.HumanExecutorImpl
	agentExecutor      agent.AgentExecutor
	mu                 sync.Mutex
	running            bool
	started            bool          // true if Start was ever called
	taskSubmitted      chan struct{} // Channel to notify when tasks are submitted
	workerLimit        int           // Max concurrent workers
	workerSem          chan struct{} // Counting semaphore for worker limit
	wg                 sync.WaitGroup
	stopCh             chan struct{} // Closed by Shutdown to signal runTaskLoop to exit
	stopOnce           sync.Once     // Ensures stopCh is closed exactly once
	loopDone           chan struct{} // Closed by Start goroutine after runTaskLoop + wg.Wait finish
}

// NewOrchestrator creates a new orchestrator with the given options.
// Default provider is MockModelProvider.
func NewOrchestrator(opts ...OrchestratorOption) *Orchestrator {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleManager := lifecycle.NewLifecycleManager()
	sched := scheduler.NewScheduler(stateStore, lifecycleManager)
	toolRegistry := defaultToolRegistry()
	contractExec := contractexec.NewContractExecutor()
	contractExec.SetProvider(provider.NewMockModelProvider())
	contractExec.SetToolRegistry(toolRegistry)

	// Wire skills loader for Layer 1 prompt injection
	root := project.MustResolveRoot()
	skillsPromptLoader := skills.NewLoader(project.SkillsDir(root))
	contractExec.SetSkillsLoader(skillsPromptLoader)

	// Wire default compaction pipeline (three-layer model)
	contractExec.SetCompactionPipeline(&contractexec.ThreeLayerCompaction{
		Micro:  &contractexec.ToolResultCompaction{KeepRecent: 3},
		Auto:   &contractexec.SummarizationCompaction{Provider: nil, KeepRecent: 4},
		Budget: 32000,
	})
	humanExec := humanexec.NewHumanExecutor()
	dispatch := dispatcher.NewDispatcher(contractExec, humanExec)
	admissionValidator := admission.NewAdmissionValidator(contractExec)

	orch := &Orchestrator{
		stateStore:         stateStore,
		lifecycleManager:   lifecycleManager,
		scheduler:          sched,
		dispatcher:         dispatch,
		contractExecutor:   contractExec,
		admissionValidator: admissionValidator,
		humanExecutor:      humanExec,
		running:            false,
		taskSubmitted:      make(chan struct{}, 1),
		workerLimit:        defaultWorkerLimit,
		workerSem:          make(chan struct{}, defaultWorkerLimit),
		stopCh:             make(chan struct{}),
		loopDone:           make(chan struct{}),
	}

	for _, opt := range opts {
		opt(orch)
	}

	// Inject agent executor into dispatcher if configured
	if orch.agentExecutor != nil {
		orch.dispatcher.SetAgentExecutor(orch.agentExecutor)
	}

	return orch
}

func defaultToolRegistry() *tool.Registry {
	registry := tool.NewRegistry()
	allowedDirs := []string{}
	allowedDir, err := os.Getwd()
	if err != nil {
		allowedDir = "."
	}
	allowedDirs = append(allowedDirs, allowedDir)
	if homeDir, err := os.UserHomeDir(); err == nil {
		desktopDir := filepath.Join(homeDir, "Desktop")
		if _, err := os.Stat(desktopDir); err == nil {
			allowedDirs = append(allowedDirs, desktopDir)
		}
	}
	_ = registry.Register(tool.NewBashTool(), []string{string(tool.ScopeSubprocess)})
	_ = registry.Register(tool.NewVerifyBashTool(), []string{string(tool.ScopeSubprocess)})
	_ = registry.Register(tool.NewFileReadTool(allowedDirs), []string{string(tool.ScopeFilesystemRead)})
	_ = registry.Register(tool.NewFileWriteTool(allowedDirs), []string{string(tool.ScopeFilesystemWrite)})
	_ = registry.Register(tool.NewHTTPClientTool([]string{"localhost", "127.0.0.1"}), []string{string(tool.ScopeNetwork)})

	// Skills: load_skill tool
	skillsDir := project.SkillsDir(allowedDir)
	skillsLoader := skills.NewLoader(skillsDir)
	_ = registry.Register(tool.NewLoadSkillTool(skillsLoader), []string{string(tool.ScopeFilesystemRead)})
	_ = registry.Register(tool.NewCompactTool(), nil)
	_ = registry.Register(tool.NewYieldTool(), nil)
	_ = registry.Register(tool.NewCheckpointTool(), nil)
	_ = registry.Register(tool.NewSpawnTool(), nil)

	return registry
}

// Start starts the orchestrator
func (o *Orchestrator) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.running {
		return fmt.Errorf("orchestrator is already running")
	}

	// Mark as running to prevent duplicate starts
	o.running = true
	o.started = true

	// Start the task execution loop; waits for workers on exit,
	// then closes loopDone so Shutdown knows all work is finished.
	go func() {
		o.runTaskLoop(ctx)
		o.wg.Wait()
		close(o.loopDone)
	}()

	return nil
}

// runTaskLoop continuously fetches ready tasks and dispatches them in parallel.
// The loop exits when stopCh is closed (graceful shutdown) or ctx is cancelled.
func (o *Orchestrator) runTaskLoop(ctx context.Context) {
	for {
		select {
		case <-o.stopCh:
			return
		case <-ctx.Done():
			return
		case <-o.taskSubmitted:
		default:
		}

		o.mu.Lock()
		if !o.lifecycleManager.IsRunning() {
			o.mu.Unlock()
			return
		}
		o.mu.Unlock()

		available := o.workerLimit - len(o.workerSem)
		if available <= 0 {
			select {
			case <-o.stopCh:
				return
			case <-ctx.Done():
				return
			case <-o.taskSubmitted:
			}
			continue
		}

		tasks, err := o.scheduler.GetReadyTasks(available)
		if err != nil {
			log.Printf("Error getting ready tasks: %v", err)
			// Intentional blocking sleep (not ctx-aware): per Zero Control philosophy,
			// shutdown does not require sub-second graceful exit. The loop already
			// checks ctx.Done()/stopCh at the top of each iteration; this sleep
			// merely throttles error retries without adding select complexity.
			time.Sleep(1 * time.Second)
			continue
		}

		if len(tasks) == 0 {
			select {
			case <-o.taskSubmitted:
			case <-o.stopCh:
				return
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, task := range tasks {
			o.workerSem <- struct{}{}
			o.wg.Add(1)
			go func(t *types.AgentTask) {
				defer o.wg.Done()
				defer func() {
					<-o.workerSem
					// Notify scheduler that a worker slot is free.
					select {
					case o.taskSubmitted <- struct{}{}:
					default:
					}
				}()
				o.executeTask(ctx, t)
			}(task)
		}
	}
}

// executeTask executes a single task with SLA timeout, failure class routing, and retry behavior.
func (o *Orchestrator) executeTask(ctx context.Context, task *types.AgentTask) {
	timeoutMs, maxRetries, failureClass, backoff := parseSLA(task.Metadata)

	// Check current status to ensure idempotency
	currentStatus, err := o.scheduler.GetStatus(task.TaskID)
	if err != nil {
		log.Printf("Error getting task status: %v", err)
		return
	}
	if currentStatus != types.TaskStatusPending && currentStatus != types.TaskStatusRunning {
		log.Printf("Task %s is already in status %s, skipping execution", task.TaskID, currentStatus)
		return
	}

	// Update status to running
	if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusRunning); err != nil {
		log.Printf("Error updating task status to running: %v", err)
		return
	}

	// Fatal failure class: admission rejects fatal + retries > 0, so maxRetries is already 0 here.
	// This assertion is a safety net, not a silent override.
	if failureClass == types.FailureClassFatal && maxRetries > 0 {
		maxRetries = 0
		log.Printf("WARN: Task %s has fatal class with retries=%d (should have been rejected at admission); forcing retries=0", task.TaskID, maxRetries)
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, dispatchErr := func() (*types.TaskResult, error) {
			execCtx := ctx
			if timeoutMs > 0 {
				var cancel context.CancelFunc
				execCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
				defer cancel()
			}
			return o.dispatcher.Dispatch(execCtx, task)
		}()

		// Check for success
		success := dispatchErr == nil && result != nil && result.Status == types.TaskStatusCompleted

		if success {
			if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusCompleted); err != nil {
				time.Sleep(100 * time.Millisecond)
				if err2 := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusCompleted); err2 != nil {
					log.Printf("ERROR: Task %s completed but status persistence failed after retry: %v", task.TaskID, err2)
				}
			}
			log.Printf("Task %s completed with status %s", task.TaskID, result.Status)
			return
		}

		// For degradable tasks with dependency errors, retry once (skip dependency check)
		if failureClass == types.FailureClassDegradable && dispatchErr != nil {
			var ae *types.AgentError
			if errors.As(dispatchErr, &ae) && ae.Code == types.ErrDependencyNotReady {
				log.Printf("Task %s attempt %d/%d dependency not ready (degradable), retrying: %v", task.TaskID, attempt+1, maxRetries+1, dispatchErr)
				time.Sleep(backoffDelay(backoff, attempt))
				continue
			}
		}

		if attempt < maxRetries {
			log.Printf("Task %s attempt %d/%d failed, retrying: %v", task.TaskID, attempt+1, maxRetries+1, dispatchErr)
			time.Sleep(backoffDelay(backoff, attempt))
			continue
		}

		// All retries exhausted
		retryErr := dispatchErr
		if retryErr == nil {
			retryErr = types.NewAgentError(types.ErrDispatchFailed, "dispatch returned non-success result")
		} else if _, ok := retryErr.(*types.AgentError); !ok {
			retryErr = types.NewAgentErrorWithCause(types.ErrDispatchFailed, "dispatch failed", retryErr)
		}
		if maxRetries > 0 {
			retryErr = types.NewAgentErrorWithCause(types.ErrTaskRetryExhausted,
				fmt.Sprintf("retry exhausted (%d attempts)", maxRetries+1), retryErr)
		}
		if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusFailed); err != nil {
			time.Sleep(100 * time.Millisecond)
			if err2 := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusFailed); err2 != nil {
				log.Printf("ERROR: Task %s failed but status persistence failed after retry: %v", task.TaskID, err2)
			}
		}
		log.Printf("Task %s failed: %s", task.TaskID, retryErr.Error())
		return
	}
}

// parseSLA extracts SLA metadata. Missing keys return zero values (use defaults).
func parseSLA(metadata map[string]string) (timeoutMs int, maxRetries int, failureClass string, backoff string) {
	if metadata == nil {
		return
	}
	if v, ok := metadata[types.SLAKeyTimeoutMs]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			timeoutMs = n
		}
	}
	if v, ok := metadata[types.SLAKeyMaxRetries]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			maxRetries = n
		}
	}
	if v, ok := metadata[types.SLAKeyFailureClass]; ok {
		failureClass = v
	}
	if v, ok := metadata[types.SLAKeyBackoff]; ok {
		backoff = v
	}
	return
}

// backoffDelay calculates the delay before a retry attempt based on strategy.
func backoffDelay(strategy string, attempt int) time.Duration {
	base := 100 * time.Millisecond
	switch strategy {
	case types.BackoffExponential:
		d := base * (1 << attempt)
		if d > 30*time.Second {
			d = 30 * time.Second
		}
		return d
	case types.BackoffLinear:
		return base * time.Duration(attempt+1)
	default: // fixed
		return base
	}
}

// SubmitTask submits a task to the orchestrator after admission validation.
func (o *Orchestrator) SubmitTask(task *types.AgentTask) error {
	if err := o.admissionValidator.Validate(task); err != nil {
		return fmt.Errorf("submit %s: admission rejected: %w", task.TaskID, err)
	}

	if err := o.scheduler.Submit(task); err != nil {
		return fmt.Errorf("submit %s: scheduling failed: %w", task.TaskID, err)
	}

	// Notify task loop that a task was submitted
	select {
	case o.taskSubmitted <- struct{}{}:
	default:
		// Channel already has notification, don't block
	}

	return nil
}

// GetTaskStatus returns the status of a task
func (o *Orchestrator) GetTaskStatus(taskID string) (types.TaskStatus, error) {
	return o.scheduler.GetStatus(taskID)
}

// RegisterContract registers a contract
func (o *Orchestrator) RegisterContract(contract *types.AgentContract) error {
	return o.contractExecutor.RegisterContract(contract)
}

// Shutdown gracefully shuts down the orchestrator.
// It signals the task loop to exit, waits for the Start goroutine to finish
// (which includes draining all in-flight workers), then shuts down the
// lifecycle manager. The caller-provided ctx bounds total wait time.
func (o *Orchestrator) Shutdown(ctx context.Context) error {
	o.mu.Lock()
	wasStarted := o.started
	o.running = false
	o.mu.Unlock()

	// Signal task loop to exit. stopCh is closed once and only once.
	o.stopOnce.Do(func() { close(o.stopCh) })

	// Wait for the Start goroutine to finish runTaskLoop + wg.Wait.
	// This guarantees no concurrent wg.Add / wg.Wait, avoiding sync races.
	if wasStarted {
		select {
		case <-o.loopDone:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Wait for lifecycle manager to complete shutdown.
	return o.lifecycleManager.Shutdown(ctx)
}

// IsRunning returns whether the orchestrator is running
func (o *Orchestrator) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.running
}

// GetAllTasks returns all tasks known to the scheduler.
func (o *Orchestrator) GetAllTasks() []*types.AgentTask {
	return o.scheduler.GetAllTasks()
}

// GetDependencyGraph returns the task-to-dependencies mapping.
func (o *Orchestrator) GetDependencyGraph() map[string][]string {
	return o.scheduler.GetDependencyGraph()
}

// ResolveCall resolves a pending human call with the given output.
func (o *Orchestrator) ResolveCall(callID string, output map[string]any) error {
	return o.humanExecutor.ResolveCall(callID, output, nil)
}
