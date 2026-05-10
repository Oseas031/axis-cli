// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// TestValidationStrategy validates test execution results.
type TestValidationStrategy struct {
	BaseStrategy
}

// NewTestValidationStrategy creates a new test validation strategy.
func NewTestValidationStrategy() *TestValidationStrategy {
	return &TestValidationStrategy{}
}

// CanHandle returns true if the criteria type is test.
func (s *TestValidationStrategy) CanHandle(criteria JudgementCriteria) bool {
	return criteria.Type == JudgementTypeTest
}

// Validate validates test execution results.
func (s *TestValidationStrategy) Validate(input any, criteria JudgementCriteria) (*JudgementItem, error) {
	if err := s.BaseStrategy.ValidateInput(input, criteria); err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        err.Error(),
		}, nil
	}

	// Extract test result from input
	testResult, err := s.extractTestResult(input)
	if err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        fmt.Sprintf("failed to extract test result: %v", err),
		}, nil
	}

	if testResult == nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       true,
			Score:        1.0,
			Details:      "no test result to validate",
		}, nil
	}

	// Calculate pass rate
	totalTests := testResult.Passed + testResult.Failed
	var passRate float64
	if totalTests > 0 {
		passRate = float64(testResult.Passed) / float64(totalTests)
	}

	// Get threshold
	minPassRate := criteria.GetThreshold("min_pass_rate", 0.90)
	passed := passRate >= minPassRate

	details := fmt.Sprintf("tests passed: %d, failed: %d, pass rate: %.1f%%",
		testResult.Passed, testResult.Failed, passRate*100)

	return &JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       passed,
		Score:        passRate,
		Details:      details,
	}, nil
}

type testResult struct {
	Passed int
	Failed int
}

func (s *TestValidationStrategy) extractTestResult(input any) (*testResult, error) {
	switch v := input.(type) {
	case *testResult:
		return v, nil
	case map[string]any:
		result := &testResult{}
		if passed, ok := v["tests_passed"].(float64); ok {
			result.Passed = int(passed)
		}
		if failed, ok := v["tests_failed"].(float64); ok {
			result.Failed = int(failed)
		}
		// Also accept string keys
		if passed, ok := v["tests_passed"].(string); ok {
			if p, err := strconv.Atoi(passed); err == nil {
				result.Passed = p
			}
		}
		if failed, ok := v["tests_failed"].(string); ok {
			if f, err := strconv.Atoi(failed); err == nil {
				result.Failed = f
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported input type %T", input)
	}
}

// RunTests runs tests for the given package and returns the result.
func RunTests(pkg string) (*testResult, error) {
	cmd := exec.Command("go", "test", "-v", "-count=1", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Test command might fail due to test failures, which is okay
	}

	result := &testResult{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "--- PASS:") {
			result.Passed++
		} else if strings.Contains(line, "--- FAIL:") {
			result.Failed++
		}
	}

	return result, nil
}
