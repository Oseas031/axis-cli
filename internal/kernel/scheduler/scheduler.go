package scheduler

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/kernel/shared_layer"
	"github.com/axis-cli/axis/internal/types"
)

// Scheduler interface defines agent task scheduling
type Scheduler interface {
	Submit(task *types.AgentTask) error
	Cancel(taskID string) error
	GetStatus(taskID string) types.TaskStatus
	GetNextTask() (*types.AgentTask, error)
	UpdateTaskStatus(taskID string, status types.TaskStatus) error
}

// SchedulerImpl implements FIFO task scheduling with dependency management
type SchedulerImpl struct {
	mu         sync.RWMutex
	queue      []*types.AgentTask
	taskMap    map[string]*types.AgentTask
	stateStore shared_layer.StateStore
	lifecycle  LifecycleChecker
	running    bool
}

// LifecycleChecker defines the interface to check if the system is running
type LifecycleChecker interface {
	IsRunning() bool
}

// NewScheduler creates a new scheduler
func NewScheduler(stateStore shared_layer.StateStore, lifecycle LifecycleChecker) *SchedulerImpl {
	return &SchedulerImpl{
		queue:      make([]*types.AgentTask, 0),
		taskMap:    make(map[string]*types.AgentTask),
		stateStore: stateStore,
		lifecycle:  lifecycle,
		running:    true,
	}
}

// Submit submits a task to the scheduler
func (s *SchedulerImpl) Submit(task *types.AgentTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lifecycle.IsRunning() {
		return errors.New("scheduler is not running")
	}

	// Check for circular dependencies
	if err := s.detectCircularDependencies(task.TaskID, task.Dependencies, make(map[string]bool)); err != nil {
		return fmt.Errorf("circular dependency detected: %w", err)
	}

	// Check if task already exists
	if _, exists := s.taskMap[task.TaskID]; exists {
		return fmt.Errorf("task %s already exists", task.TaskID)
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
		visited[dep] = true
		if task, exists := s.taskMap[dep]; exists {
			if err := s.detectCircularDependencies(dep, task.Dependencies, visited); err != nil {
				return err
			}
		}
		delete(visited, dep)
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
		return fmt.Errorf("task %s not found", taskID)
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
func (s *SchedulerImpl) GetStatus(taskID string) types.TaskStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.taskMap[taskID]
	if !exists {
		return ""
	}
	return task.Status
}

// GetNextTask returns the next task to execute (FIFO with dependency check)
func (s *SchedulerImpl) GetNextTask() (*types.AgentTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lifecycle.IsRunning() {
		return nil, errors.New("scheduler is not running")
	}

	for _, task := range s.queue {
		if task.Status == types.TaskStatusPending {
			// Check if all dependencies are completed
			if s.areDependenciesCompleted(task.Dependencies) {
				// Mark as dispatched to prevent duplicate returns
				task.Status = types.TaskStatusRunning
				now := time.Now()
				task.StartedAt = &now

				// Update state store
				state := types.TaskState{
					Task:      task,
					UpdatedAt: time.Now(),
				}
				if err := s.stateStore.Save(task.TaskID, state); err != nil {
					// Revert status on error
					task.Status = types.TaskStatusPending
					task.StartedAt = nil
					return nil, fmt.Errorf("failed to update task state: %w", err)
				}

				return task, nil
			}
		}
	}

	return nil, nil // No ready tasks
}

// areDependenciesCompleted checks if all dependencies are completed
func (s *SchedulerImpl) areDependenciesCompleted(dependencies []string) bool {
	for _, depID := range dependencies {
		task, exists := s.taskMap[depID]
		if !exists {
			return false
		}
		if task.Status != types.TaskStatusCompleted {
			return false
		}
	}
	return true
}

// UpdateTaskStatus updates the status of a task
func (s *SchedulerImpl) UpdateTaskStatus(taskID string, status types.TaskStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.taskMap[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
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
