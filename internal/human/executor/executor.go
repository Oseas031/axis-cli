package executor

import (
	"fmt"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// HumanExecutor interface defines human task execution
type HumanExecutor interface {
	ExecuteCall(req *types.HumanCallRequest) (*types.HumanCallResult, error)
	GetCallStatus(callID string) types.CallStatus
	ResolveCall(callID string, output map[string]any, err error) error
}

// HumanExecutorImpl implements human task execution
type HumanExecutorImpl struct {
	mu    sync.RWMutex
	calls map[string]*types.HumanCallResult
}

// NewHumanExecutor creates a new human executor
func NewHumanExecutor() *HumanExecutorImpl {
	return &HumanExecutorImpl{
		calls: make(map[string]*types.HumanCallResult),
	}
}

// ExecuteCall submits a human call request
func (e *HumanExecutorImpl) ExecuteCall(req *types.HumanCallRequest) (*types.HumanCallResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if call already exists
	if _, exists := e.calls[req.CallID]; exists {
		return nil, fmt.Errorf("call %s already exists", req.CallID)
	}

	// Create pending call result
	result := &types.HumanCallResult{
		CallID: req.CallID,
		Status: types.CallStatusPending,
	}

	e.calls[req.CallID] = result
	return result, nil
}

// GetCallStatus returns the status of a human call
func (e *HumanExecutorImpl) GetCallStatus(callID string) types.CallStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	call, exists := e.calls[callID]
	if !exists {
		return ""
	}
	return call.Status
}

// ResolveCall resolves a human call with output or error
func (e *HumanExecutorImpl) ResolveCall(callID string, output map[string]any, err error) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	call, exists := e.calls[callID]
	if !exists {
		return fmt.Errorf("call %s not found", callID)
	}

	if call.Status != types.CallStatusPending {
		return fmt.Errorf("call %s is not in pending status", callID)
	}

	if err != nil {
		call.Status = types.CallStatusFailed
		call.Error = err.Error()
	} else {
		call.Status = types.CallStatusCompleted
		call.Output = output
	}

	call.Resolved = time.Now()
	return nil
}
