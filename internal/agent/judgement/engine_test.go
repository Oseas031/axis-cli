// Package judgement provides self-judgement capabilities for agent execution validation.
package judgement

import (
	"fmt"
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

func TestEngine_Judge_CanHandleFalse(t *testing.T) {
	e := NewEngine()
	// Syntax strategy is registered but cannot handle contract criteria
	criteria := []strategies.JudgementCriteria{
		{
			Name:    "contract",
			Type:    strategies.JudgementTypeContract,
			Weight:  1.0,
			Enabled: true,
		},
	}

	result, err := e.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Judgements) != 1 {
		t.Fatalf("expected 1 judgement, got %d", len(result.Judgements))
	}
	if result.Judgements[0].Passed {
		t.Error("expected CanHandle=false to produce failed judgement")
	}
	if result.Judgements[0].Error == "" {
		t.Error("expected error message for CanHandle=false")
	}
}

func TestEngine_Judge_ValidateError(t *testing.T) {
	e := NewEngine()
	// Register a custom strategy that always returns an error
	custom := &alwaysErrorStrategy{}
	e.RegisterStrategy(strategies.JudgementTypeCustom, custom)

	criteria := []strategies.JudgementCriteria{
		{
			Name:    "always_error",
			Type:    strategies.JudgementTypeCustom,
			Weight:  1.0,
			Enabled: true,
		},
	}

	result, err := e.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Judgements) != 1 {
		t.Fatalf("expected 1 judgement, got %d", len(result.Judgements))
	}
	if result.Judgements[0].Passed {
		t.Error("expected Validate error to produce failed judgement")
	}
	if result.Judgements[0].Error == "" {
		t.Error("expected error message for Validate error")
	}
}

func TestEngine_Judge_AggregatedScore(t *testing.T) {
	e := NewEngine()
	// Register a custom strategy with a specific score
	custom := &scoredStrategy{score: 0.85}
	e.RegisterStrategy(strategies.JudgementTypeCustom, custom)

	criteria := []strategies.JudgementCriteria{
		{
			Name:    "score_test",
			Type:    strategies.JudgementTypeCustom,
			Weight:  1.0,
			Enabled: true,
		},
	}

	result, err := e.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected score 0.85 to pass default threshold 0.75")
	}
	if result.Score != 0.85 {
		t.Errorf("expected aggregated score 0.85, got %f", result.Score)
	}
	if result.Confidence != 1.0 {
		t.Errorf("expected confidence 1.0 (1 passed / 1 total), got %f", result.Confidence)
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

// alwaysErrorStrategy is a test strategy that always returns an error.
type alwaysErrorStrategy struct{}

func (s *alwaysErrorStrategy) Validate(input any, criteria strategies.JudgementCriteria) (*strategies.JudgementItem, error) {
	return nil, fmt.Errorf("intentional validation error")
}

func (s *alwaysErrorStrategy) CanHandle(criteria strategies.JudgementCriteria) bool {
	return criteria.Type == strategies.JudgementTypeCustom
}

// scoredStrategy is a test strategy that returns a configurable score.
type scoredStrategy struct {
	score float64
}

func (s *scoredStrategy) Validate(input any, criteria strategies.JudgementCriteria) (*strategies.JudgementItem, error) {
	return &strategies.JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       s.score >= DefaultJudgementThresholds.PassingScore,
		Score:        s.score,
		Details:      fmt.Sprintf("scored strategy: %f", s.score),
	}, nil
}

func (s *scoredStrategy) CanHandle(criteria strategies.JudgementCriteria) bool {
	return criteria.Type == strategies.JudgementTypeCustom
}
