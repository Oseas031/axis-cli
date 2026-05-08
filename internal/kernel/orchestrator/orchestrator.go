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
	"github.com/axis-cli/axis/internal/kernel/shared_layer"
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
	stateStore := shared_layer.NewMemoryStateStore()
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
		running:          true,
		taskSubmitted:    make(chan struct{}, 1),
	}
}

// Start starts the orchestrator
func (o *Orchestrator) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.running {
		return fmt.Errorf("orchestrator is already shut down")
	}

	// Mark as started to prevent duplicate starts
	o.running = false

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
	// Update status to running
	if err := o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusRunning); err != nil {
		log.Printf("Error updating task status to running: %v", err)
		return
	}

	// Dispatch the task
	result, err := o.dispatcher.Dispatch(ctx, task)
	if err != nil {
		log.Printf("Error dispatching task %s: %v", task.TaskID, err)
		o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusFailed)
		return
	}

	// Update task status based on result
	if result.Status == types.TaskStatusCompleted {
		o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusCompleted)
	} else {
		o.scheduler.UpdateTaskStatus(task.TaskID, types.TaskStatusFailed)
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
func (o *Orchestrator) GetTaskStatus(taskID string) types.TaskStatus {
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

	return o.lifecycleManager.Shutdown(ctx)
}

// IsRunning returns whether the orchestrator is running
func (o *Orchestrator) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.running
}
