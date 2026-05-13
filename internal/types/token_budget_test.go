package types

import "testing"

func TestTokenBudget_Basic(t *testing.T) {
	b := NewTokenBudget(1000)
	if b.Total() != 1000 {
		t.Errorf("expected total 1000, got %d", b.Total())
	}
	if b.Remaining() != 1000 {
		t.Errorf("expected remaining 1000, got %d", b.Remaining())
	}
	if err := b.Consume(100); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if b.Consumed() != 100 {
		t.Errorf("expected consumed 100, got %d", b.Consumed())
	}
	if b.Remaining() != 900 {
		t.Errorf("expected remaining 900, got %d", b.Remaining())
	}
}

func TestTokenBudget_Exhausted(t *testing.T) {
	b := NewTokenBudget(100)
	if err := b.Consume(101); err == nil {
		t.Error("expected error when budget exhausted")
	}
}

func TestTokenBudget_StageAllocation(t *testing.T) {
	b := NewTokenBudget(1000)

	b.SetStage(StagePrototype)
	if b.StageLimit() != 100 {
		t.Errorf("prototype stage limit: expected 100, got %d", b.StageLimit())
	}

	b.SetStage(StageSmallScale)
	if b.StageLimit() != 300 {
		t.Errorf("small stage limit: expected 300, got %d", b.StageLimit())
	}

	b.SetStage(StageLargeScale)
	if b.StageLimit() != 600 {
		t.Errorf("large stage limit: expected 600, got %d", b.StageLimit())
	}
}

func TestTokenBudget_ExceedsStage(t *testing.T) {
	b := NewTokenBudget(1000)
	b.SetStage(StagePrototype) // 10% = 100 tokens

	b.Consume(50)
	if b.ExceedsStage() {
		t.Error("should not exceed stage at 50/100")
	}

	b.Consume(60) // total 110 > 100
	if !b.ExceedsStage() {
		t.Error("should exceed stage at 110/100")
	}
}

func TestTokenBudget_RemainingNeverNegative(t *testing.T) {
	b := NewTokenBudget(10)
	b.Consume(20) // over budget
	if b.Remaining() != 0 {
		t.Errorf("expected remaining 0, got %d", b.Remaining())
	}
}
