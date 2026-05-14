package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

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
			o.saveTaskResult(task.TaskID, result)
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
		o.saveTaskResult(task.TaskID, &types.TaskResult{TaskID: task.TaskID, Status: types.TaskStatusFailed, Error: retryErr.Error(), Completed: time.Now()})
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
