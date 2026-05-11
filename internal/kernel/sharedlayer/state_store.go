// Package sharedlayer provides shared state storage for tasks.
package sharedlayer

import (
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/types"
)

// StateStore interface defines task state persistence
type StateStore interface {
	Save(taskID string, state types.TaskState) error
	Load(taskID string) (types.TaskState, error)
	Delete(taskID string) error
	ListAll() (map[string]types.TaskState, error)
}

// MemoryStateStore implements in-memory state storage
type MemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]types.TaskState
}

// NewMemoryStateStore creates a new in-memory state store
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		states: make(map[string]types.TaskState),
	}
}

// Save stores a task state
func (s *MemoryStateStore) Save(taskID string, state types.TaskState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[taskID] = state
	return nil
}

// Load retrieves a task state
func (s *MemoryStateStore) Load(taskID string) (types.TaskState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, exists := s.states[taskID]
	if !exists {
		return types.TaskState{}, fmt.Errorf("task %s not found", taskID)
	}
	return state, nil
}

// Delete removes a task state
func (s *MemoryStateStore) Delete(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, taskID)
	return nil
}

// ListAll returns a copy of all stored states.
func (s *MemoryStateStore) ListAll() (map[string]types.TaskState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]types.TaskState, len(s.states))
	for k, v := range s.states {
		result[k] = v
	}
	return result, nil
}
