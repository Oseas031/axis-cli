// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import "fmt"

// JudgementType represents the type of validation performed.
type JudgementType string

const (
	JudgementTypeSyntax   JudgementType = "syntax"
	JudgementTypeSemantic JudgementType = "semantic"
	JudgementTypeContract JudgementType = "contract"
	JudgementTypeTest     JudgementType = "test"
	JudgementTypeCoverage JudgementType = "coverage"
	JudgementTypeCustom   JudgementType = "custom"
)

// JudgementCriteria defines the criteria for judging an execution result.
type JudgementCriteria struct {
	Name       string             `json:"name"`
	Type       JudgementType      `json:"type"`
	Weight     float64            `json:"weight"`
	Thresholds map[string]float64 `json:"thresholds"`
	Enabled    bool               `json:"enabled"`
}

// GetThreshold gets a threshold value with a default if not found.
func (jc JudgementCriteria) GetThreshold(name string, defaultValue float64) float64 {
	if jc.Thresholds == nil {
		return defaultValue
	}
	if v, ok := jc.Thresholds[name]; ok {
		return v
	}
	return defaultValue
}

// JudgementItem represents the result of a single judgement criteria evaluation.
type JudgementItem struct {
	CriteriaName string  `json:"criteria_name"`
	Passed       bool    `json:"passed"`
	Score        float64 `json:"score"`
	Details      string  `json:"details"`
	Error        string  `json:"error,omitempty"`
}

// ValidationStrategy is the interface for implementing judgement strategies.
type ValidationStrategy interface {
	// Validate performs the validation and returns a judgement item.
	Validate(input any, criteria JudgementCriteria) (*JudgementItem, error)
	// CanHandle returns true if this strategy can handle the given criteria.
	CanHandle(criteria JudgementCriteria) bool
}

// StrategyRegistry manages all registered validation strategies.
type StrategyRegistry struct {
	strategies map[JudgementType]ValidationStrategy
}

// NewStrategyRegistry creates a new strategy registry.
func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{
		strategies: make(map[JudgementType]ValidationStrategy),
	}
}

// Register registers a validation strategy for a judgement type.
func (r *StrategyRegistry) Register(judgementType JudgementType, strategy ValidationStrategy) {
	r.strategies[judgementType] = strategy
}

// Get returns the strategy for the given judgement type.
func (r *StrategyRegistry) Get(judgementType JudgementType) (ValidationStrategy, bool) {
	strategy, ok := r.strategies[judgementType]
	return strategy, ok
}

// List returns all registered judgement types.
func (r *StrategyRegistry) List() []JudgementType {
	types := make([]JudgementType, 0, len(r.strategies))
	for t := range r.strategies {
		types = append(types, t)
	}
	return types
}

// BaseStrategy provides common functionality for strategies.
type BaseStrategy struct{}

// ValidateInput performs common input validation.
func (BaseStrategy) ValidateInput(input any, criteria JudgementCriteria) error {
	if criteria.Name == "" {
		return fmt.Errorf("criteria name is empty")
	}
	if criteria.Weight < 0 || criteria.Weight > 1 {
		return fmt.Errorf("criteria weight must be between 0 and 1, got %f", criteria.Weight)
	}
	return nil
}

// DefaultJudgementItem creates a default judgement item.
func DefaultJudgementItem(criteriaName string, passed bool, score float64, details string) JudgementItem {
	return JudgementItem{
		CriteriaName: criteriaName,
		Passed:       passed,
		Score:        score,
		Details:      details,
	}
}

// ErrorJudgementItem creates a judgement item for an error case.
func ErrorJudgementItem(criteriaName string, err error) JudgementItem {
	item := JudgementItem{
		CriteriaName: criteriaName,
		Passed:       false,
		Score:        0,
		Details:      "",
	}
	if err != nil {
		item.Error = err.Error()
	}
	return item
}

// ContractExecutorProvider interface for contract-based validation.
type ContractExecutorProvider interface {
	ValidateOutput(ctx any, contractID string, output map[string]any) error
}
