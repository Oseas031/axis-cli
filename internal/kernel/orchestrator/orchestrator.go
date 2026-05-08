// Package orchestrator provides task orchestration and coordination.
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	dispatcher "github.com/axis-cli/axis/internal/kernel/dispatcher"
	"github.com/axis-cli/axis/internal/kernel/lifecycle"
	"github.com/axis-cli/axis/internal/kernel/scheduler"
	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/types"
)

// Orchestrator coordinates all kernel modules
type Orchestrator struct {
	stateStore       sharedlayer.StateStore
	lifecycleManager *lifecycle.LifecycleManagerImpl
	scheduler        *scheduler.SchedulerImpl
	dispatcher       *dispatcher.DispatcherImpl
	contractExecutor *contractexec.ContractExecutorImpl
	humanExecutor    *humanexec.HumanExecutorImpl
	mu               sync.Mutex
	running          bool
	taskSubmitted    chan struct{} // Channel to notify when tasks are submitted
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator() *Orchestrator {
	stateStore := sharedlayer.NewMemoryStateStore()
	lifecycleManager := lifecycle.NewLifecycleManager()
	sched := scheduler.NewScheduler(stateStore, lifecycleManager)
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := dispatcher.NewDispatcher(contractExec, humanExec)

	return &Orchestrator{
		stateStore:       stateStore,
		lifecycleManager: lifecycleManager,
		scheduler:        sched,
		dispatcher:       dispatch,
		contractExecutor: contractExec,
		humanExecutor:    humanExec,
		running:          false,
		taskSubmitted:    make(chan struct{}, 1),
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

	// Start the task execution loop
	go o.runTaskLoop(ctx)

	return nil
}

// runTaskLoop continuously processes tasks from the scheduler
func (o *Orchestrator) runTaskLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.taskSubmitted:
			// Task submitted, process immediately
		default:
			// No notification, check periodically
			o.mu.Lock()
			if !o.lifecycleManager.IsRunning() {
				o.mu.Unlock()
				return
			}
			o.mu.Unlock()

			task, err := o.scheduler.GetNextTask()
			if err != nil {
				log.Printf("Error getting next task: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if task == nil {
				// No tasks ready, wait for notification
				select {
				case <-o.taskSubmitted:
					continue
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			// Execute the task
			o.executeTask(ctx, task)
		}
	}
}

// executeTask executes a single task
func (o *Orchestrator) executeTask(ctx context.Context, task *types.AgentTask) {
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

	// Dispatch the task
	result, err := o.dispatcher.Dispatch(ctx, task)
	if err != nil {
		log.Printf("Error dispatching task %s: %v", task.TaskID, err)
		if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusFailed); err != nil {
			log.Printf("Error updating task status to failed: %v", err)
		}
		return
	}

	// Update task status based on result
	if result.Status == types.TaskStatusCompleted {
		if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusCompleted); err != nil {
			log.Printf("Error updating task status to completed: %v", err)
		}
	} else {
		if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusFailed); err != nil {
			log.Printf("Error updating task status to failed: %v", err)
		}
	}

	log.Printf("Task %s completed with status %s", task.TaskID, result.Status)
}

// SubmitTask submits a task to the orchestrator
func (o *Orchestrator) SubmitTask(task *types.AgentTask) error {
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
