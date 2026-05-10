// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import (
	"testing"
)

func TestNewSyntaxValidationStrategy(t *testing.T) {
	s := NewSyntaxValidationStrategy()
	if s == nil {
		t.Fatal("expected non-nil strategy")
	}
}

func TestSyntaxValidationStrategy_CanHandle(t *testing.T) {
	s := NewSyntaxValidationStrategy()

	if !s.CanHandle(JudgementCriteria{Type: JudgementTypeSyntax}) {
		t.Error("expected to handle syntax type")
	}
	if s.CanHandle(JudgementCriteria{Type: JudgementTypeTest}) {
		t.Error("expected not to handle test type")
	}
}

func TestSyntaxValidationStrategy_Validate_EmptyInput(t *testing.T) {
	s := NewSyntaxValidationStrategy()

	result, err := s.Validate([]string{}, JudgementCriteria{
		Name:   "syntax",
		Type:   JudgementTypeSyntax,
		Weight: 1.0,
		Thresholds: map[string]float64{
			"min_pass_rate": 1.0,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected empty input to pass")
	}
	if result.Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", result.Score)
	}
}

func TestSyntaxValidationStrategy_Validate_InvalidCriteria(t *testing.T) {
	s := NewSyntaxValidationStrategy()

	// Empty name should cause validation to return failed item
	result, err := s.Validate([]string{}, JudgementCriteria{
		Name:   "",
		Type:   JudgementTypeSyntax,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected invalid criteria to fail")
	}
}

func TestTestValidationStrategy_CanHandle(t *testing.T) {
	s := NewTestValidationStrategy()

	if !s.CanHandle(JudgementCriteria{Type: JudgementTypeTest}) {
		t.Error("expected to handle test type")
	}
	if s.CanHandle(JudgementCriteria{Type: JudgementTypeSyntax}) {
		t.Error("expected not to handle syntax type")
	}
}

func TestTestValidationStrategy_Validate(t *testing.T) {
	s := NewTestValidationStrategy()

	result, err := s.Validate(map[string]any{
		"tests_passed": float64(10),
		"tests_failed": float64(0),
	}, JudgementCriteria{
		Name:   "test",
		Type:   JudgementTypeTest,
		Weight: 1.0,
		Thresholds: map[string]float64{
			"min_pass_rate": 0.9,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected test validation to pass")
	}
}

func TestTestValidationStrategy_Validate_NoTests(t *testing.T) {
	s := NewTestValidationStrategy()

	// With 0 tests, pass rate is 0%, which fails the 90% threshold
	result, err := s.Validate(map[string]any{}, JudgementCriteria{
		Name:   "test",
		Type:   JudgementTypeTest,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 0 tests = 0% pass rate < 90% threshold = fail
	if result.Passed {
		t.Error("expected 0 tests to fail")
	}
}

func TestCoverageValidationStrategy_CanHandle(t *testing.T) {
	s := NewCoverageValidationStrategy()

	if !s.CanHandle(JudgementCriteria{Type: JudgementTypeCoverage}) {
		t.Error("expected to handle coverage type")
	}
	if s.CanHandle(JudgementCriteria{Type: JudgementTypeSyntax}) {
		t.Error("expected not to handle syntax type")
	}
}

func TestCoverageValidationStrategy_Validate(t *testing.T) {
	s := NewCoverageValidationStrategy()

	result, err := s.Validate(0.90, JudgementCriteria{
		Name:   "coverage",
		Type:   JudgementTypeCoverage,
		Weight: 1.0,
		Thresholds: map[string]float64{
			"min_coverage": 0.85,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected 90% coverage to pass")
	}
}

func TestCoverageValidationStrategy_Validate_StringInput(t *testing.T) {
	s := NewCoverageValidationStrategy()

	result, err := s.Validate("85.5%", JudgementCriteria{
		Name:   "coverage",
		Type:   JudgementTypeCoverage,
		Weight: 1.0,
		Thresholds: map[string]float64{
			"min_coverage": 0.85,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected 85.5% coverage to pass")
	}
}
