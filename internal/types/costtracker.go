// Package types provides core data types for the agent system.
package types

import "fmt"

// CostTracker tracks accumulated cost against a budget during task execution.
// When cost exceeds 80% of budget, the tracker signals degradation.
type CostTracker struct {
	TotalCostUSD float64 `json:"total_cost_usd"`
	BudgetUSD    float64 `json:"budget_usd"`
	Degraded     bool    `json:"degraded"`
}

// NewCostTracker creates a CostTracker with the given budget in USD.
func NewCostTracker(budgetUSD float64) *CostTracker {
	return &CostTracker{
		BudgetUSD: budgetUSD,
	}
}

// NewCostTrackerFromTask creates a CostTracker from an AgentTask's CostBudget
// and optional metadata override (sla.cost_budget).
func NewCostTrackerFromTask(task *AgentTask) *CostTracker {
	budget := task.CostBudget
	if task.Metadata != nil {
		if raw, ok := task.Metadata[SLAKeyCostBudget]; ok {
			var parsed float64
			if _, err := fmt.Sscanf(raw, "%f", &parsed); err == nil && parsed > 0 {
				budget = parsed
			}
		}
	}
	return NewCostTracker(budget)
}

// AddCost adds cost in USD and marks as degraded if over 80% of budget.
func (ct *CostTracker) AddCost(cost float64) {
	ct.TotalCostUSD += cost
	if ct.BudgetUSD > 0 && ct.TotalCostUSD >= ct.BudgetUSD*0.8 {
		ct.Degraded = true
	}
}

// IsDegraded returns true if cost has exceeded 80% of budget.
func (ct *CostTracker) IsDegraded() bool {
	return ct.Degraded
}

// IsExhausted returns true if cost has reached or exceeded 100% of budget.
// Returns false if budget is zero or negative (unlimited).
func (ct *CostTracker) IsExhausted() bool {
	if ct.BudgetUSD <= 0 {
		return false
	}
	return ct.TotalCostUSD >= ct.BudgetUSD
}

// RemainingBudget returns the remaining budget in USD.
// Returns 0 if exhausted. Returns -1 if budget is unlimited (zero/negative).
func (ct *CostTracker) RemainingBudget() float64 {
	if ct.BudgetUSD <= 0 {
		return -1
	}
	remaining := ct.BudgetUSD - ct.TotalCostUSD
	if remaining < 0 {
		return 0
	}
	return remaining
}

// BudgetExceededError creates an AgentError for budget exceeded scenarios.
func (ct *CostTracker) BudgetExceededError() *AgentError {
	return NewAgentError(
		ErrCostBudgetExceeded,
		fmt.Sprintf("cost budget $%.4f exceeded (total: $%.4f)", ct.BudgetUSD, ct.TotalCostUSD),
	)
}
