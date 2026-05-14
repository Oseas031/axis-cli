// Package orchestrator provides task orchestration and coordination.
package orchestrator

import (
	"context"
	"fmt"
	"log"
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
	"github.com/axis-cli/axis/internal/memory/horizon"
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
		o.modelProvider = p
	}
}

// WithToolRegistry sets a custom tool registry, overriding the default.
func WithToolRegistry(r *tool.Registry) OrchestratorOption {
	return func(o *Orchestrator) {
		o.toolRegistry = r
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
	toolRegistry       *tool.Registry
	modelProvider      provider.ModelProvider
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
	humanExec := humanexec.NewHumanExecutor()

	orch := &Orchestrator{
		stateStore:       stateStore,
		lifecycleManager: lifecycleManager,
		scheduler:        sched,
		humanExecutor:    humanExec,
		running:          false,
		taskSubmitted:    make(chan struct{}, 1),
		workerLimit:      defaultWorkerLimit,
		workerSem:        make(chan struct{}, defaultWorkerLimit),
		stopCh:           make(chan struct{}),
		loopDone:         make(chan struct{}),
	}

	for _, opt := range opts {
		opt(orch)
	}

	// Use default tool registry if none was provided via option
	if orch.toolRegistry == nil {
		orch.toolRegistry = defaultToolRegistry()
	}

	// Wire skills loader for Layer 1 prompt injection
	root := project.MustResolveRoot()
	skillsPromptLoader := skills.NewLoader(project.SkillsDir(root))

	// Wire principles loader for Layer 1 prompt injection (derived from dream)
	principlesStore := horizon.NewStore(project.MemoryDir(root))

	// Resolve model provider
	resolvedProvider := provider.ModelProvider(provider.NewMockModelProvider())
	if orch.modelProvider != nil {
		resolvedProvider = orch.modelProvider
	}

	contractExec := contractexec.NewContractExecutorWithConfig(contractexec.ExecutorConfig{
		Provider:     resolvedProvider,
		ToolRegistry: orch.toolRegistry,
		SkillsLoader: skillsPromptLoader,
		PrinciplesLoader: principlesStore,
		CompactionPipeline: &contractexec.ThreeLayerCompaction{
			Micro:  &contractexec.ToolResultCompaction{KeepRecent: 3},
			Auto:   &contractexec.SummarizationCompaction{Provider: nil, KeepRecent: 4},
			Budget: 32000,
		},
	})
	orch.contractExecutor = contractExec
	orch.dispatcher = dispatcher.NewDispatcher(contractExec, humanExec)
	orch.admissionValidator = admission.NewAdmissionValidator(contractExec)

	// Inject agent executor into dispatcher if configured
	if orch.agentExecutor != nil {
		orch.dispatcher.SetAgentExecutor(orch.agentExecutor)
		// Wire tool registry into LLM agent executor if applicable
		if llmExec, ok := orch.agentExecutor.(*agent.LLMAgentExecutor); ok {
			llmExec.SetToolRegistry(orch.toolRegistry)
		}
	}

	return orch
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

	// Mark orphaned tasks from previous run
	o.markOrphanedTasks()

	// Start the task execution loop; waits for workers on exit,
	// then closes loopDone so Shutdown knows all work is finished.
	go func() {
		o.runTaskLoop(ctx)
		o.wg.Wait()
		close(o.loopDone)
	}()

	return nil
}

// markOrphanedTasks marks tasks left in Running state from a previous run as Failed.
func (o *Orchestrator) markOrphanedTasks() {
	states, err := o.stateStore.ListAll()
	if err != nil {
		return
	}
	for id, s := range states {
		if s.Task != nil && s.Task.Status == types.TaskStatusRunning {
			s.Task.Status = types.TaskStatusFailed
			s.UpdatedAt = time.Now()
			_ = o.stateStore.Save(id, s)
			log.Printf("[Orchestrator] Marked orphaned task %s as failed", id)
		}
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

// GetTaskResult returns the result of a completed or failed task
func (o *Orchestrator) GetTaskResult(taskID string) (*types.TaskResult, error) {
	state, err := o.stateStore.Load(taskID)
	if err != nil {
		return nil, err
	}
	return state.Result, nil
}

// saveTaskResult persists a task result into the state store.
func (o *Orchestrator) saveTaskResult(taskID string, result *types.TaskResult) {
	state, err := o.stateStore.Load(taskID)
	if err != nil {
		return
	}
	state.Result = result
	state.UpdatedAt = time.Now()
	_ = o.stateStore.Save(taskID, state)
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
