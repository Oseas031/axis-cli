package strategies

import (
	"context"
	"testing"
)

// mockContractProvider is a test double for ContractExecutorProvider.
type mockContractProvider struct{}

func (m *mockContractProvider) ValidateOutput(ctx any, contractID string, output map[string]any) error {
	return nil
}

func TestContractValidationStrategy_CanHandle(t *testing.T) {
	s := NewContractValidationStrategy(&mockContractProvider{})
	if !s.CanHandle(JudgementCriteria{Type: JudgementTypeContract}) {
		t.Error("expected to handle contract type")
	}
	if s.CanHandle(JudgementCriteria{Type: JudgementTypeSyntax}) {
		t.Error("expected not to handle syntax type")
	}
}

func TestContractValidationStrategy_Validate_NilResult(t *testing.T) {
	s := NewContractValidationStrategy(&mockContractProvider{})

	result, err := s.Validate(nil, JudgementCriteria{
		Name:   "contract",
		Type:   JudgementTypeContract,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Nil input is treated as unsupported type -> fails extraction
	if result.Passed {
		t.Error("expected nil result to fail (unsupported type)")
	}
	if result.Error == "" {
		t.Error("expected error message for unsupported input")
	}
}

func TestContractValidationStrategy_Validate_PassingTests(t *testing.T) {
	s := NewContractValidationStrategy(&mockContractProvider{})

	result, err := s.Validate(map[string]any{
		"is_acceptable": true,
		"tests_passed":  float64(10),
		"tests_failed":  float64(0),
		"coverage":      0.90,
	}, JudgementCriteria{
		Name:   "contract",
		Type:   JudgementTypeContract,
		Weight: 1.0,
		Thresholds: map[string]float64{
			"min_pass_rate": 0.9,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected passing tests to pass")
	}
	if result.Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", result.Score)
	}
}

func TestContractValidationStrategy_Validate_FailingTests(t *testing.T) {
	s := NewContractValidationStrategy(&mockContractProvider{})

	result, err := s.Validate(map[string]any{
		"is_acceptable": true,
		"tests_passed":  float64(5),
		"tests_failed":  float64(5),
		"coverage":      0.90,
	}, JudgementCriteria{
		Name:   "contract",
		Type:   JudgementTypeContract,
		Weight: 1.0,
		Thresholds: map[string]float64{
			"min_pass_rate": 0.9,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected 50% pass rate to fail 90% threshold")
	}
	if result.Score != 0.5 {
		t.Errorf("expected score 0.5, got %f", result.Score)
	}
}

func TestContractValidationStrategy_Validate_CoverageOnly(t *testing.T) {
	s := NewContractValidationStrategy(&mockContractProvider{})

	result, err := s.Validate(map[string]any{
		"is_acceptable": true,
		"coverage":      0.80,
	}, JudgementCriteria{
		Name:   "contract",
		Type:   JudgementTypeContract,
		Weight: 1.0,
		Thresholds: map[string]float64{
			"min_pass_rate": 0.9,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected 80% coverage to fail 90% threshold")
	}
	if result.Score != 0.80 {
		t.Errorf("expected score 0.80, got %f", result.Score)
	}
}

func TestContractValidationStrategy_Validate_InvalidCriteria(t *testing.T) {
	s := NewContractValidationStrategy(&mockContractProvider{})

	result, err := s.Validate(map[string]any{}, JudgementCriteria{
		Name:   "",
		Type:   JudgementTypeContract,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected invalid criteria to fail")
	}
}

func TestContractValidationStrategy_Validate_UnsupportedInput(t *testing.T) {
	s := NewContractValidationStrategy(&mockContractProvider{})

	result, err := s.Validate(12345, JudgementCriteria{
		Name:   "contract",
		Type:   JudgementTypeContract,
		Weight: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected unsupported input to fail")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
}

func TestContractValidationStrategy_ValidateContractOutput_NilProvider(t *testing.T) {
	s := NewContractValidationStrategy(nil)

	err := s.ValidateContractOutput(context.Background(), "test-contract", nil)
	if err != nil {
		t.Fatalf("expected nil provider to return nil error, got %v", err)
	}
}
