// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import (
	"context"
	"fmt"
)

// ContractValidationStrategy validates contract execution results.
type ContractValidationStrategy struct {
	BaseStrategy
	contractProvider ContractExecutorProvider
}

// NewContractValidationStrategy creates a new contract validation strategy.
func NewContractValidationStrategy(provider ContractExecutorProvider) *ContractValidationStrategy {
	return &ContractValidationStrategy{
		contractProvider: provider,
	}
}

// CanHandle returns true if the criteria type is contract.
func (s *ContractValidationStrategy) CanHandle(criteria JudgementCriteria) bool {
	return criteria.Type == JudgementTypeContract
}

// Validate validates contract execution output.
func (s *ContractValidationStrategy) Validate(input any, criteria JudgementCriteria) (*JudgementItem, error) {
	if err := s.BaseStrategy.ValidateInput(input, criteria); err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        err.Error(),
		}, nil
	}

	// Extract validation result from input
	validationResult, err := s.extractValidationResult(input)
	if err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        fmt.Sprintf("failed to extract validation result: %v", err),
		}, nil
	}

	if validationResult == nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       true,
			Score:        1.0,
			Details:      "no validation result to check",
		}, nil
	}

	// Calculate score based on validation
	passed := validationResult.IsAcceptable
	score := 0.0

	if validationResult.TestsPassed > 0 || validationResult.TestsFailed > 0 {
		totalTests := validationResult.TestsPassed + validationResult.TestsFailed
		passRate := float64(validationResult.TestsPassed) / float64(totalTests)
		score = passRate
	} else {
		// No tests, use coverage as score
		score = validationResult.Coverage
	}

	minPassRate := criteria.GetThreshold("min_pass_rate", 1.0)
	if score < minPassRate {
		passed = false
	}

	details := fmt.Sprintf("tests passed: %d, failed: %d, coverage: %.1f%%",
		validationResult.TestsPassed, validationResult.TestsFailed, validationResult.Coverage)

	return &JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       passed,
		Score:        score,
		Details:      details,
	}, nil
}

type validationResult struct {
	IsAcceptable bool
	TestsPassed  int
	TestsFailed  int
	Coverage     float64
}

func (s *ContractValidationStrategy) extractValidationResult(input any) (*validationResult, error) {
	switch v := input.(type) {
	case *validationResult:
		return v, nil
	case map[string]any:
		result := &validationResult{}
		if acceptable, ok := v["is_acceptable"].(bool); ok {
			result.IsAcceptable = acceptable
		}
		if passed, ok := v["tests_passed"].(float64); ok {
			result.TestsPassed = int(passed)
		}
		if failed, ok := v["tests_failed"].(float64); ok {
			result.TestsFailed = int(failed)
		}
		if coverage, ok := v["coverage"].(float64); ok {
			result.Coverage = coverage
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported input type %T", input)
	}
}

// ValidateContractOutput validates contract output using the provider.
func (s *ContractValidationStrategy) ValidateContractOutput(ctx context.Context, contractID string, output map[string]any) error {
	if s.contractProvider == nil {
		return nil
	}
	return s.contractProvider.ValidateOutput(ctx, contractID, output)
}
