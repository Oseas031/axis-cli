// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SyntaxValidationStrategy validates code syntax using go fmt and go vet.
type SyntaxValidationStrategy struct {
	BaseStrategy
}

// NewSyntaxValidationStrategy creates a new syntax validation strategy.
func NewSyntaxValidationStrategy() *SyntaxValidationStrategy {
	return &SyntaxValidationStrategy{}
}

// CanHandle returns true if the criteria type is syntax.
func (s *SyntaxValidationStrategy) CanHandle(criteria JudgementCriteria) bool {
	return criteria.Type == JudgementTypeSyntax
}

// Validate validates the syntax of Go code files.
func (s *SyntaxValidationStrategy) Validate(input any, criteria JudgementCriteria) (*JudgementItem, error) {
	if err := s.BaseStrategy.ValidateInput(input, criteria); err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        err.Error(),
		}, nil
	}

	files, ok := input.([]string)
	if !ok {
		// Try to extract files from a map
		if m, ok := input.(map[string]any); ok {
			if f, ok := m["files"].([]string); ok {
				files = f
			}
		}
	}

	if len(files) == 0 {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       true,
			Score:        1.0,
			Details:      "no files to validate",
		}, nil
	}

	var fmtErrors []string
	var vetErrors []string

	for _, file := range files {
		// Check if file exists and is a Go file
		if !strings.HasSuffix(file, ".go") {
			continue
		}

		// Check formatting
		if err := s.checkFormat(file); err != nil {
			fmtErrors = append(fmtErrors, fmt.Sprintf("%s: %v", file, err))
		}

		// Run go vet
		if err := s.runVet(file); err != nil {
			vetErrors = append(vetErrors, fmt.Sprintf("%s: %v", file, err))
		}
	}

	totalChecks := len(files) * 2 // fmt + vet
	if totalChecks == 0 {
		totalChecks = 1
	}
	passedChecks := totalChecks - len(fmtErrors) - len(vetErrors)
	score := float64(passedChecks) / float64(totalChecks)

	var details strings.Builder
	if len(fmtErrors) > 0 {
		details.WriteString(fmt.Sprintf("format errors in %d files; ", len(fmtErrors)))
	}
	if len(vetErrors) > 0 {
		details.WriteString(fmt.Sprintf("vet errors in %d files", len(vetErrors)))
	}
	if details.Len() == 0 {
		details.WriteString("all syntax checks passed")
	}

	// Determine pass/fail based on thresholds
	minPassRate := criteria.GetThreshold("min_pass_rate", 1.0)
	passed := score >= minPassRate

	return &JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       passed,
		Score:        score,
		Details:      details.String(),
	}, nil
}

func (s *SyntaxValidationStrategy) checkFormat(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	cmd := exec.Command("gofmt", "-l", file)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gofmt failed: %v", err)
	}

	if len(output) > 0 {
		return fmt.Errorf("file is not formatted")
	}
	return nil
}

func (s *SyntaxValidationStrategy) runVet(file string) error {
	dir := filepath.Dir(file)
	cmd := exec.Command("go", "vet", file)
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	// go vet returns non-zero for issues, which is expected
	if err != nil {
		return fmt.Errorf("vet found issues")
	}
	return nil
}
