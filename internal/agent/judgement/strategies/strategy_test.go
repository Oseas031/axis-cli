// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import (
	"testing"
)

func TestNewStrategyRegistry(t *testing.T) {
	r := NewStrategyRegistry()

	if r == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestStrategyRegistry_Register(t *testing.T) {
	r := NewStrategyRegistry()

	strategy := &customTestStrategy{}
	r.Register(JudgementTypeCustom, strategy)

	got, ok := r.Get(JudgementTypeCustom)
	if !ok {
		t.Fatal("expected to get registered strategy")
	}
	if got == nil {
		t.Error("expected non-nil strategy")
	}
}

func TestStrategyRegistry_Get_NotFound(t *testing.T) {
	r := NewStrategyRegistry()

	_, ok := r.Get(JudgementTypeSyntax)
	if ok {
		t.Error("expected not found")
	}
}

func TestStrategyRegistry_List(t *testing.T) {
	r := NewStrategyRegistry()

	r.Register(JudgementTypeSyntax, NewSyntaxValidationStrategy())
	r.Register(JudgementTypeTest, NewTestValidationStrategy())

	types := r.List()

	if len(types) != 2 {
		t.Errorf("expected 2 types, got %d", len(types))
	}
}

func TestBaseStrategy_ValidateInput(t *testing.T) {
	s := BaseStrategy{}

	tests := []struct {
		name      string
		criteria  JudgementCriteria
		wantError bool
	}{
		{
			name: "valid",
			criteria: JudgementCriteria{
				Name:   "test",
				Weight: 0.5,
			},
			wantError: false,
		},
		{
			name: "empty name",
			criteria: JudgementCriteria{
				Name:   "",
				Weight: 0.5,
			},
			wantError: true,
		},
		{
			name: "negative weight",
			criteria: JudgementCriteria{
				Name:   "test",
				Weight: -0.1,
			},
			wantError: true,
		},
		{
			name: "weight too high",
			criteria: JudgementCriteria{
				Name:   "test",
				Weight: 1.5,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(nil, tt.criteria)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDefaultJudgementItem(t *testing.T) {
	item := DefaultJudgementItem("test", true, 0.8, "details")

	if item.CriteriaName != "test" {
		t.Errorf("expected 'test', got %s", item.CriteriaName)
	}
	if !item.Passed {
		t.Error("expected Passed to be true")
	}
	if item.Score != 0.8 {
		t.Errorf("expected 0.8, got %f", item.Score)
	}
	if item.Details != "details" {
		t.Errorf("expected 'details', got %s", item.Details)
	}
}

func TestErrorJudgementItem(t *testing.T) {
	item := ErrorJudgementItem("test", nil)

	if item.CriteriaName != "test" {
		t.Errorf("expected 'test', got %s", item.CriteriaName)
	}
	if item.Passed {
		t.Error("expected Passed to be false")
	}
	if item.Score != 0 {
		t.Errorf("expected 0, got %f", item.Score)
	}
}

func TestJudgementCriteria_GetThreshold(t *testing.T) {
	c := JudgementCriteria{
		Name: "test",
		Thresholds: map[string]float64{
			"min_coverage": 0.85,
		},
	}

	if c.GetThreshold("min_coverage", 0.75) != 0.85 {
		t.Error("expected 0.85")
	}
	if c.GetThreshold("missing", 0.75) != 0.75 {
		t.Error("expected default 0.75")
	}
	if c.GetThreshold("missing", 0.0) != 0.0 {
		t.Error("expected default 0.0")
	}
}

// customTestStrategy is a simple test strategy.
type customTestStrategy struct{}

func (s *customTestStrategy) Validate(input any, criteria JudgementCriteria) (*JudgementItem, error) {
	return &JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       true,
		Score:        1.0,
	}, nil
}

func (s *customTestStrategy) CanHandle(criteria JudgementCriteria) bool {
	return criteria.Type == JudgementTypeCustom
}
