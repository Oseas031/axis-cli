// Package orchestrator provides task orchestration and coordination.
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/contract/admission"
	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	dispatcher "github.com/axis-cli/axis/internal/kernel/dispatcher"
	"github.com/axis-cli/axis/internal/kernel/lifecycle"
	"github.com/axis-cli/axis/internal/kernel/scheduler"
	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/types"
)

const defaultWorkerLimit = 5

// Orchestrator coordinates all kernel modules
type Orchestrator struct {
	stateStore         sharedlayer.StateStore
	lifecycleManager   *lifecycle.LifecycleManagerImpl
	scheduler          *scheduler.SchedulerImpl
	dispatcher         *dispatcher.DispatcherImpl
	contractExecutor   *contractexec.ContractExecutorImpl
	admissionValidator *admission.AdmissionValidatorImpl
	humanExecutor      *humanexec.HumanExecutorImpl
	mu                 sync.Mutex
	running            bool
	taskSubmitted      chan struct{} // Channel to notify when tasks are submitted
	workerLimit        int           // Max concurrent workers
	workerSem          chan struct{} // Counting semaphore for worker limit
	wg                 sync.WaitGroup
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator() *Orchestrator {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleManager := lifecycle.NewLifecycleManager()
	sched := scheduler.NewScheduler(stateStore, lifecycleManager)
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := dispatcher.NewDispatcher(contractExec, humanExec)
	admissionValidator := admission.NewAdmissionValidator(contractExec)

	return &Orchestrator{
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
	}
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

	// Start the task execution loop; waits for workers on exit
	go func() {
		o.runTaskLoop(ctx)
		o.wg.Wait()
	}()

	return nil
}

// runTaskLoop continuously fetches ready tasks and dispatches them in parallel.
func (o *Orchestrator) runTaskLoop(ctx context.Context) {
	for {
		select {
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
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}

		tasks, err := o.scheduler.GetReadyTasks(available)
		if err != nil {
			log.Printf("Error getting ready tasks: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if len(tasks) == 0 {
			select {
			case <-o.taskSubmitted:
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}

		for _, task := range tasks {
			o.workerSem <- struct{}{}
			o.wg.Add(1)
			go func(t *types.AgentTask) {
				defer o.wg.Done()
				defer func() { <-o.workerSem }()
				o.executeTask(ctx, t)
			}(task)
		}
	}
}

// executeTask executes a single task with SLA timeout and retry behavior.
func (o *Orchestrator) executeTask(ctx context.Context, task *types.AgentTask) {
	timeoutMs, maxRetries := parseSLA(task.Metadata)

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

	for attempt := 0; attempt <= maxRetries; attempt++ {
		execCtx := ctx
		if timeoutMs > 0 {
			var cancel context.CancelFunc
			execCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
			defer cancel()
		}

		result, dispatchErr := o.dispatcher.Dispatch(execCtx, task)
		if dispatchErr == nil && result != nil && result.Status == types.TaskStatusCompleted {
			if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusCompleted); err != nil {
				log.Printf("Error updating task status to completed: %v", err)
			}
			log.Printf("Task %s completed with status %s", task.TaskID, result.Status)
			return
		}

		if attempt < maxRetries {
			log.Printf("Task %s attempt %d/%d failed, retrying: %v", task.TaskID, attempt+1, maxRetries+1, dispatchErr)
			continue
		}

		// All retries exhausted
		errMsg := "dispatch failed"
		if dispatchErr != nil {
			errMsg = dispatchErr.Error()
		}
		if maxRetries > 0 {
			errMsg = fmt.Sprintf("retry exhausted (%d attempts): %s", maxRetries+1, errMsg)
		}
		if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusFailed); err != nil {
			log.Printf("Error updating task status to failed: %v", err)
		}
		log.Printf("Task %s failed: %s", task.TaskID, errMsg)
		return
	}
}

// parseSLA extracts SLA metadata. Missing keys return zero values (use defaults).
func parseSLA(metadata map[string]string) (timeoutMs int, maxRetries int) {
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
	return
}

// SubmitTask submits a task to the orchestrator after admission validation.
func (o *Orchestrator) SubmitTask(task *types.AgentTask) error {
	if err := o.admissionValidator.Validate(task); err != nil {
		return err
	}

	if err := o.scheduler.Submit(task); err != nil {
		return err
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

// Shutdown gracefully shuts down the orchestrator
func (o *Orchestrator) Shutdown(ctx context.Context) error {
	o.mu.Lock()
	o.running = false
	o.mu.Unlock()

	// Notify task loop to stop
	select {
	case o.taskSubmitted <- struct{}{}:
	default:
	}

	// Wait for lifecycle manager to complete shutdown
	return o.lifecycleManager.Shutdown(ctx)
}

// IsRunning returns whether the orchestrator is running
func (o *Orchestrator) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.running
}
