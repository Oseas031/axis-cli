// Package scheduler provides FIFO task scheduling with dependency management.
package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/kernel/sharedlayer"
	"github.com/axis-cli/axis/internal/types"
)

// Scheduler interface defines agent task scheduling
type Scheduler interface {
	Submit(task *types.AgentTask) error
	Cancel(taskID string) error
	GetStatus(taskID string) (types.TaskStatus, error)
	GetNextTask() (*types.AgentTask, error)
	GetReadyTasks(limit int) ([]*types.AgentTask, error)
	UpdateTaskStatus(taskID string, status types.TaskStatus) error
}

// SchedulerImpl implements FIFO task scheduling with dependency management
type SchedulerImpl struct {
	mu         sync.RWMutex
	queue      []*types.AgentTask
	taskMap    map[string]*types.AgentTask
	stateStore sharedlayer.StateStore
	lifecycle  LifecycleChecker
}

// LifecycleChecker defines the interface to check if the system is running
type LifecycleChecker interface {
	IsRunning() bool
}

// NewScheduler creates a new scheduler
func NewScheduler(stateStore sharedlayer.StateStore, lifecycle LifecycleChecker) *SchedulerImpl {
	return &SchedulerImpl{
		queue:      make([]*types.AgentTask, 0),
		taskMap:    make(map[string]*types.AgentTask),
		stateStore: stateStore,
		lifecycle:  lifecycle,
	}
}

// Submit submits a task to the scheduler
func (s *SchedulerImpl) Submit(task *types.AgentTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lifecycle.IsRunning() {
		return types.NewAgentError(types.ErrSchedulerNotRunning, "scheduler is not running")
	}

	// Check for circular dependencies
	if err := s.detectCircularDependencies(task.TaskID, task.Dependencies, make(map[string]bool)); err != nil {
		return types.NewAgentErrorWithCause(types.ErrDependencyCycle, "circular dependency detected", err)
	}

	// Check if task already exists
	if _, exists := s.taskMap[task.TaskID]; exists {
		return types.NewAgentError(types.ErrTaskAlreadyExists, fmt.Sprintf("task %s already exists", task.TaskID))
	}

	// Set initial status
	task.Status = types.TaskStatusPending
	task.CreatedAt = time.Now()

	// Add to queue and map
	s.queue = append(s.queue, task)
	s.taskMap[task.TaskID] = task

	// Save to state store
	state := types.TaskState{
		Task:      task,
		UpdatedAt: time.Now(),
	}
	if err := s.stateStore.Save(task.TaskID, state); err != nil {
		return fmt.Errorf("failed to save task state: %w", err)
	}

	return nil
}

// detectCircularDependencies checks for circular dependencies using DFS
func (s *SchedulerImpl) detectCircularDependencies(taskID string, dependencies []string, visited map[string]bool) error {
	// Mark current task as visited to detect if it appears in its own dependency chain
	visited[taskID] = true

	for _, dep := range dependencies {
		if visited[dep] {
			return fmt.Errorf("circular dependency involving task %s", dep)
		}
		if task, exists := s.taskMap[dep]; exists {
			if err := s.detectCircularDependencies(dep, task.Dependencies, visited); err != nil {
				return err
			}
		}
	}

	// Clean up current task from visited
	delete(visited, taskID)
	return nil
}

// Cancel cancels a task
func (s *SchedulerImpl) Cancel(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.taskMap[taskID]
	if !exists {
		return types.NewAgentError(types.ErrTaskNotFound, fmt.Sprintf("task %s not found", taskID))
	}

	// Only allow cancellation of pending tasks
	if task.Status != types.TaskStatusPending {
		return fmt.Errorf("cannot cancel task in status %s", task.Status)
	}

	// Remove from queue
	for i, t := range s.queue {
		if t.TaskID == taskID {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			break
		}
	}

	// Remove from map
	delete(s.taskMap, taskID)

	// Delete from state store
	if err := s.stateStore.Delete(taskID); err != nil {
		return fmt.Errorf("failed to delete task state: %w", err)
	}

	return nil
}

// GetStatus returns the status of a task
func (s *SchedulerImpl) GetStatus(taskID string) (types.TaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.taskMap[taskID]
	if !exists {
		return "", types.NewAgentError(types.ErrTaskNotFound, fmt.Sprintf("task %s not found", taskID))
	}
	return task.Status, nil
}

// GetNextTask returns the next task to execute (FIFO with dependency check)
func (s *SchedulerImpl) GetNextTask() (*types.AgentTask, error) {
	tasks, err := s.GetReadyTasks(1)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks[0], nil
}

func (s *SchedulerImpl) GetReadyTasks(limit int) ([]*types.AgentTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lifecycle.IsRunning() {
		return nil, types.NewAgentError(types.ErrSchedulerNotRunning, "scheduler is not running")
	}

	readyTasks := make([]*types.AgentTask, 0)
	for _, task := range s.queue {
		if task.Status == types.TaskStatusPending {
			if s.areDependenciesCompleted(task.Dependencies) {
				task.Status = types.TaskStatusRunning
				now := time.Now()
				task.StartedAt = &now

				state := types.TaskState{
					Task:      task,
					UpdatedAt: time.Now(),
				}
				if err := s.stateStore.Save(task.TaskID, state); err != nil {
					task.Status = types.TaskStatusPending
					task.StartedAt = nil
					return nil, fmt.Errorf("failed to update task state: %w", err)
				}

				readyTasks = append(readyTasks, task)
				if limit > 0 && len(readyTasks) >= limit {
					break
				}
			}
		}
	}

	return readyTasks, nil
}

// areDependenciesCompleted checks if all dependencies are done (completed or failed).
func (s *SchedulerImpl) areDependenciesCompleted(dependencies []string) bool {
	for _, depID := range dependencies {
		task, exists := s.taskMap[depID]
		if !exists {
			return false
		}
		if task.Status == types.TaskStatusFailed {
			continue
		}
		if task.Status != types.TaskStatusCompleted {
			return false
		}
	}
	return true
}

// GetAllTasks returns copies of all tasks currently known to the scheduler.
func (s *SchedulerImpl) GetAllTasks() []*types.AgentTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*types.AgentTask, 0, len(s.taskMap))
	for _, task := range s.taskMap {
		taskCopy := *task
		if len(task.Dependencies) > 0 {
			taskCopy.Dependencies = make([]string, len(task.Dependencies))
			copy(taskCopy.Dependencies, task.Dependencies)
		}
		if task.Metadata != nil {
			taskCopy.Metadata = make(map[string]string, len(task.Metadata))
			for k, v := range task.Metadata {
				taskCopy.Metadata[k] = v
			}
		}
		if task.Input != nil {
			taskCopy.Input = make(map[string]any, len(task.Input))
			for k, v := range task.Input {
				taskCopy.Input[k] = v
			}
		}
		result = append(result, &taskCopy)
	}
	return result
}

// GetDependencyGraph returns the task-to-dependencies mapping for all known tasks.
func (s *SchedulerImpl) GetDependencyGraph() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]string, len(s.taskMap))
	for id, task := range s.taskMap {
		if len(task.Dependencies) == 0 {
			result[id] = nil
		} else {
			deps := make([]string, len(task.Dependencies))
			copy(deps, task.Dependencies)
			result[id] = deps
		}
	}
	return result
}

// UpdateTaskStatus updates the status of a task
func (s *SchedulerImpl) UpdateTaskStatus(taskID string, status types.TaskStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.taskMap[taskID]
	if !exists {
		return types.NewAgentError(types.ErrTaskNotFound, fmt.Sprintf("task %s not found", taskID))
	}

	task.Status = status
	now := time.Now()

	switch status {
	case types.TaskStatusRunning:
		task.StartedAt = &now
	case types.TaskStatusCompleted, types.TaskStatusFailed:
		task.CompletedAt = &now
	}

	// Update state store
	state := types.TaskState{
		Task:      task,
		UpdatedAt: time.Now(),
	}
	if err := s.stateStore.Save(taskID, state); err != nil {
		return fmt.Errorf("failed to update task state: %w", err)
	}

	return nil
}
