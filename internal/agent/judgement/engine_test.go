// Package judgement provides self-judgement capabilities for agent execution validation.
package judgement

import (
	"testing"

	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()

	if e == nil {
		t.Fatal("expected non-nil engine")
	}

	// Check default strategies are registered
	types := e.ListStrategies()
	if len(types) != 3 {
		t.Errorf("expected 3 default strategies, got %d", len(types))
	}
}

func TestEngine_Judge_EmptyCriteria(t *testing.T) {
	e := NewEngine()

	result, err := e.Judge(nil, []strategies.JudgementCriteria{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Passed {
		t.Error("expected result to pass with empty criteria")
	}
	if result.Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", result.Score)
	}
}

func TestEngine_Judge_DisabledCriteria(t *testing.T) {
	e := NewEngine()

	criteria := []strategies.JudgementCriteria{
		{
			Name:    "test",
			Type:    strategies.JudgementTypeSyntax,
			Enabled: false,
		},
	}

	result, err := e.Judge([]string{"test.go"}, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Disabled criteria should be skipped
	if len(result.Judgements) != 0 {
		t.Errorf("expected 0 judgements, got %d", len(result.Judgements))
	}
}

func TestEngine_Judge_NoStrategy(t *testing.T) {
	e := NewEngine()

	criteria := []strategies.JudgementCriteria{
		{
			Name:    "test",
			Type:    strategies.JudgementTypeCustom, // No strategy registered
			Enabled: true,
		},
	}

	result, err := e.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have an error judgement item
	if len(result.Judgements) != 1 {
		t.Errorf("expected 1 judgement, got %d", len(result.Judgements))
	}
	if result.Judgements[0].Passed {
		t.Error("expected judgement to fail")
	}
}

func TestEngine_Judge_SyntaxValidation(t *testing.T) {
	e := NewEngine()

	criteria := []strategies.JudgementCriteria{
		{
			Name:    "syntax",
			Type:    strategies.JudgementTypeSyntax,
			Weight:  1.0,
			Enabled: true,
			Thresholds: map[string]float64{
				"min_pass_rate": 1.0,
			},
		},
	}

	// Empty file list should pass
	result, err := e.Judge([]string{}, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Judgements) != 1 {
		t.Errorf("expected 1 judgement, got %d", len(result.Judgements))
	}
	if !result.Judgements[0].Passed {
		t.Error("expected empty file list to pass syntax check")
	}
}

func TestEngine_RegisterStrategy(t *testing.T) {
	e := NewEngine()

	// Create a custom strategy
	customStrategy := &customTestStrategy{}
	e.RegisterStrategy(strategies.JudgementTypeCustom, customStrategy)

	strategy, ok := e.GetStrategy(strategies.JudgementTypeCustom)
	if !ok {
		t.Error("expected to get custom strategy")
	}

	if strategy == nil {
		t.Error("expected non-nil strategy")
	}
}

func TestEngine_ListStrategies(t *testing.T) {
	e := NewEngine()

	types := e.ListStrategies()

	if len(types) != 3 {
		t.Errorf("expected 3 strategies, got %d", len(types))
	}
}

// customTestStrategy is a test strategy that always returns a passing result.
type customTestStrategy struct{}

func (s *customTestStrategy) Validate(input any, criteria strategies.JudgementCriteria) (*strategies.JudgementItem, error) {
	return &strategies.JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       true,
		Score:        1.0,
		Details:      "custom strategy",
	}, nil
}

func (s *customTestStrategy) CanHandle(criteria strategies.JudgementCriteria) bool {
	return criteria.Type == strategies.JudgementTypeCustom
}
