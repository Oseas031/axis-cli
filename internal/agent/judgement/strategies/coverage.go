// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CoverageValidationStrategy validates code coverage results.
type CoverageValidationStrategy struct {
	BaseStrategy
}

// NewCoverageValidationStrategy creates a new coverage validation strategy.
func NewCoverageValidationStrategy() *CoverageValidationStrategy {
	return &CoverageValidationStrategy{}
}

// CanHandle returns true if the criteria type is coverage.
func (s *CoverageValidationStrategy) CanHandle(criteria JudgementCriteria) bool {
	return criteria.Type == JudgementTypeCoverage
}

// Validate validates code coverage results.
func (s *CoverageValidationStrategy) Validate(input any, criteria JudgementCriteria) (*JudgementItem, error) {
	if err := s.BaseStrategy.ValidateInput(input, criteria); err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        err.Error(),
		}, nil
	}

	// Extract coverage from input
	coverage, err := s.extractCoverage(input)
	if err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        fmt.Sprintf("failed to extract coverage: %v", err),
		}, nil
	}

	if coverage < 0 {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "no coverage data available",
		}, nil
	}

	// Get threshold
	minCoverage := criteria.GetThreshold("min_coverage", 0.85)
	passed := coverage >= minCoverage

	details := fmt.Sprintf("coverage: %.1f%%, threshold: %.1f%%", coverage*100, minCoverage*100)

	return &JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       passed,
		Score:        coverage,
		Details:      details,
	}, nil
}

func (s *CoverageValidationStrategy) extractCoverage(input any) (float64, error) {
	switch v := input.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v) / 100.0, nil
	case string:
		// Parse percentage string like "85.5%"
		v = strings.TrimSuffix(v, "%")
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return -1, err
		}
		return f / 100.0, nil
	case map[string]any:
		// Try to extract coverage from various keys
		for _, key := range []string{"coverage", "coverage_percent", "coverage_percentage"} {
			if cov, ok := v[key].(float64); ok {
				return cov, nil
			}
			if cov, ok := v[key].(string); ok {
				cov = strings.TrimSuffix(cov, "%")
				if f, err := strconv.ParseFloat(cov, 64); err == nil {
					return f / 100.0, nil
				}
			}
		}
		return -1, fmt.Errorf("coverage not found in map")
	default:
		return -1, fmt.Errorf("unsupported input type %T", input)
	}
}

// RunCoverage runs go test with coverage for the given package and returns coverage percentage.
func RunCoverage(pkg string) (float64, error) {
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "-covermode=atomic", pkg)
	_, err := cmd.CombinedOutput()
	if err != nil {
		// Test might fail but coverage report is still generated
	}

	// Read coverage percentage from output
	cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get coverage: %v", err)
	}

	// Parse total coverage from output like "total:	coverage: 85.5%"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "total:") {
			parts := strings.Split(line, "\t")
			for _, part := range parts {
				if strings.HasPrefix(part, "coverage:") {
					cov := strings.TrimPrefix(strings.TrimSpace(part), "coverage:")
					cov = strings.TrimSuffix(cov, "%")
					if f, err := strconv.ParseFloat(cov, 64); err == nil {
						return f / 100.0, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("coverage not found in output")
}
