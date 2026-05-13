package types

import (
	"fmt"
	"sync"
)

// SLA metadata key for token budget.
const SLAKeyTokenBudget = "sla.token_budget"

// TokenStage represents a named execution stage with a budget fraction.
type TokenStage string

const (
	StagePrototype  TokenStage = "prototype"  // 10% budget
	StageSmallScale TokenStage = "small"      // 30% budget
	StageLargeScale TokenStage = "large"      // 60% budget
)

// StageAllocation maps stages to their budget fractions.
var StageAllocation = map[TokenStage]float64{
	StagePrototype:  0.10,
	StageSmallScale: 0.30,
	StageLargeScale: 0.60,
}

// TokenBudget tracks token consumption against a fixed budget with stage allocation.
type TokenBudget struct {
	mu       sync.Mutex
	total    int
	consumed int
	stage    TokenStage
}

// NewTokenBudget creates a budget with the given total token limit.
func NewTokenBudget(total int) *TokenBudget {
	return &TokenBudget{total: total, stage: StagePrototype}
}

// SetStage advances to the given stage.
func (b *TokenBudget) SetStage(stage TokenStage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stage = stage
}

// Stage returns the current stage.
func (b *TokenBudget) Stage() TokenStage {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.stage
}

// StageLimit returns the token limit for the current stage.
func (b *TokenBudget) StageLimit() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	frac, ok := StageAllocation[b.stage]
	if !ok {
		frac = 1.0
	}
	return int(float64(b.total) * frac)
}

// Remaining returns tokens remaining in the total budget.
func (b *TokenBudget) Remaining() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	r := b.total - b.consumed
	if r < 0 {
		return 0
	}
	return r
}

// Consumed returns total tokens consumed so far.
func (b *TokenBudget) Consumed() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.consumed
}

// Total returns the total budget.
func (b *TokenBudget) Total() int {
	return b.total
}

// Consume records token usage. Returns ErrTokenBudgetExhausted if budget exceeded.
func (b *TokenBudget) Consume(tokens int) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consumed += tokens
	if b.consumed > b.total {
		return &AgentError{Code: ErrTokenBudgetExhausted, Message: fmt.Sprintf("token budget exhausted: consumed %d / %d", b.consumed, b.total)}
	}
	return nil
}

// ExceedsStage returns true if consumed tokens exceed the current stage's allocation.
func (b *TokenBudget) ExceedsStage() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	frac, ok := StageAllocation[b.stage]
	if !ok {
		frac = 1.0
	}
	limit := int(float64(b.total) * frac)
	return b.consumed > limit
}
