// Package strategies provides validation strategy implementations for self-judgement.
package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axis-cli/axis/internal/model/provider"
)

// SemanticValidationStrategy validates code semantics using an LLM model provider.
type SemanticValidationStrategy struct {
	BaseStrategy
	provider provider.ModelProvider
}

// NewSemanticValidationStrategy creates a new semantic validation strategy.
func NewSemanticValidationStrategy(p provider.ModelProvider) *SemanticValidationStrategy {
	return &SemanticValidationStrategy{
		provider: p,
	}
}

// CanHandle returns true if the criteria type is semantic.
func (s *SemanticValidationStrategy) CanHandle(criteria JudgementCriteria) bool {
	return criteria.Type == JudgementTypeSemantic
}

// Validate validates the semantics of code using an LLM.
func (s *SemanticValidationStrategy) Validate(input any, criteria JudgementCriteria) (*JudgementItem, error) {
	if err := s.BaseStrategy.ValidateInput(input, criteria); err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        err.Error(),
		}, nil
	}

	// Extract code from input
	code, err := s.extractCode(input)
	if err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        fmt.Sprintf("failed to extract code: %v", err),
		}, nil
	}

	if code == "" {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       true,
			Score:        1.0,
			Details:      "no code to validate",
		}, nil
	}

	// Build validation prompt
	prompt := s.buildValidationPrompt(code)

	// Call model provider
	resp, err := s.provider.Execute(context.Background(), &provider.ModelRequest{
		ContractID: "self/semantic-validation",
		Input: map[string]any{
			"code":      code,
			"prompt":    prompt,
			"task_type": "semantic_validation",
		},
	})
	if err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      "",
			Error:        fmt.Sprintf("model provider error: %v", err),
		}, nil
	}

	// Parse response
	var content string
	if resp.Output != nil {
		if outputStr, ok := resp.Output["output"].(string); ok {
			content = outputStr
		} else if contentMap, ok := resp.Output["content"].(string); ok {
			content = contentMap
		} else {
			// Fallback: serialize the whole output
			data, _ := json.Marshal(resp.Output)
			content = string(data)
		}
	}
	return s.parseResponse(content, criteria)
}

func (s *SemanticValidationStrategy) extractCode(input any) (string, error) {
	switch v := input.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case map[string]any:
		if code, ok := v["code"].(string); ok {
			return code, nil
		}
		if files, ok := v["files"].([]string); ok {
			return strings.Join(files, "\n"), nil
		}
		return "", fmt.Errorf("unsupported map format")
	default:
		return "", fmt.Errorf("unsupported input type %T", input)
	}
}

func (s *SemanticValidationStrategy) buildValidationPrompt(code string) string {
	return fmt.Sprintf(`Analyze the following Go code for:
1. Logic correctness
2. Potential bugs or edge cases
3. Code quality issues
4. Security vulnerabilities

Code:
%s

Respond with a JSON object:
{
  "passed": true/false,
  "score": 0.0-1.0,
  "issues": ["list of issues found"],
  "confidence": 0.0-1.0
}`, code)
}

func (s *SemanticValidationStrategy) parseResponse(content string, criteria JudgementCriteria) (*JudgementItem, error) {
	// Try to extract JSON from the response
	content = strings.TrimSpace(content)

	// Find JSON boundaries
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 {
		// Fallback: treat entire response as content analysis
		passed := !strings.Contains(strings.ToLower(content), "error")
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       passed,
			Score:        0.5,
			Details:      content,
		}, nil
	}

	jsonStr := content[start : end+1]
	var result struct {
		Passed     bool     `json:"passed"`
		Score      float64  `json:"score"`
		Issues     []string `json:"issues"`
		Confidence float64  `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &JudgementItem{
			CriteriaName: criteria.Name,
			Passed:       false,
			Score:        0,
			Details:      fmt.Sprintf("failed to parse response: %v", err),
			Error:        err.Error(),
		}, nil
	}

	details := ""
	if len(result.Issues) > 0 {
		details = strings.Join(result.Issues, "; ")
	} else {
		details = "no issues found"
	}

	return &JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       result.Passed,
		Score:        result.Score,
		Details:      details,
	}, nil
}
