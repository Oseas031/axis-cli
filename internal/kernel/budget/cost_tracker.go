package budget

import (
	"fmt"
	"sync"
)

// CostTracker tracks accumulated cost per task and enforces budget limits.
type CostTracker struct {
	mu    sync.Mutex
	costs map[string]float64 // taskID -> accumulated USD cost
}

// NewCostTracker creates a new cost tracker.
func NewCostTracker() *CostTracker {
	return &CostTracker{costs: make(map[string]float64)}
}

// Add records cost for a task. Returns the new total.
func (ct *CostTracker) Add(taskID string, cost float64) float64 {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.costs[taskID] += cost
	return ct.costs[taskID]
}

// Get returns the accumulated cost for a task.
func (ct *CostTracker) Get(taskID string) float64 {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.costs[taskID]
}

// CheckBudget returns an error if the task has exceeded its budget.
// Returns nil if budget is 0 (unlimited) or not exceeded.
func (ct *CostTracker) CheckBudget(taskID string, budget float64) error {
	if budget <= 0 {
		return nil // unlimited
	}
	ct.mu.Lock()
	accumulated := ct.costs[taskID]
	ct.mu.Unlock()
	if accumulated >= budget {
		return fmt.Errorf("cost budget exceeded for task %s: accumulated $%.4f >= budget $%.4f", taskID, accumulated, budget)
	}
	return nil
}

// ShouldDowngrade returns true if accumulated cost exceeds 80% of budget.
func (ct *CostTracker) ShouldDowngrade(taskID string, budget float64) bool {
	if budget <= 0 {
		return false
	}
	ct.mu.Lock()
	accumulated := ct.costs[taskID]
	ct.mu.Unlock()
	return accumulated >= budget*0.8
}

// Reset clears cost tracking for a task.
func (ct *CostTracker) Reset(taskID string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	delete(ct.costs, taskID)
}
