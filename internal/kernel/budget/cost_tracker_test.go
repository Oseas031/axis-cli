package budget

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestCostTracker_AddAndGet(t *testing.T) {
	ct := NewCostTracker()
	ct.Add("task-1", 0.10)
	ct.Add("task-1", 0.05)
	if got := ct.Get("task-1"); !almostEqual(got, 0.15) {
		t.Fatalf("expected 0.15, got %f", got)
	}
}

func TestCostTracker_CheckBudget_Unlimited(t *testing.T) {
	ct := NewCostTracker()
	ct.Add("task-1", 999.0)
	if err := ct.CheckBudget("task-1", 0); err != nil {
		t.Fatal("unlimited budget should never error")
	}
}

func TestCostTracker_CheckBudget_Exceeded(t *testing.T) {
	ct := NewCostTracker()
	ct.Add("task-1", 0.60)
	if err := ct.CheckBudget("task-1", 0.50); err == nil {
		t.Fatal("expected budget exceeded error")
	}
}

func TestCostTracker_CheckBudget_NotExceeded(t *testing.T) {
	ct := NewCostTracker()
	ct.Add("task-1", 0.30)
	if err := ct.CheckBudget("task-1", 0.50); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCostTracker_ShouldDowngrade(t *testing.T) {
	ct := NewCostTracker()
	ct.Add("task-1", 0.40)
	if !ct.ShouldDowngrade("task-1", 0.50) {
		t.Fatal("80% threshold reached, should downgrade")
	}
	if ct.ShouldDowngrade("task-1", 0) {
		t.Fatal("unlimited should never downgrade")
	}
}

func TestCostTracker_Reset(t *testing.T) {
	ct := NewCostTracker()
	ct.Add("task-1", 0.50)
	ct.Reset("task-1")
	if got := ct.Get("task-1"); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}
