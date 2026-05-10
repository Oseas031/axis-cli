// Package judgement provides self-judgement capabilities for agent execution validation.
package judgement

import (
	"testing"

	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
)

func TestNewJudgementResult(t *testing.T) {
	result := NewJudgementResult()

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Judgements == nil {
		t.Error("expected Judgements to be initialized")
	}
	if result.SuggestedFixes == nil {
		t.Error("expected SuggestedFixes to be initialized")
	}
	if result.Metadata == nil {
		t.Error("expected Metadata to be initialized")
	}
}

func TestJudgementResult_AddJudgement(t *testing.T) {
	result := NewJudgementResult()

	// Add a passing judgement
	item := strategies.JudgementItem{
		CriteriaName: "test",
		Passed:       true,
		Score:        0.8,
		Details:      "test passed",
	}
	result.AddJudgement(item)

	if len(result.Judgements) != 1 {
		t.Errorf("expected 1 judgement, got %d", len(result.Judgements))
	}
	if result.Score != 0.8 {
		t.Errorf("expected score 0.8, got %f", result.Score)
	}
}

func TestJudgementResult_AddJudgement_FailingAddsFix(t *testing.T) {
	result := NewJudgementResult()

	// Add a failing judgement
	item := strategies.JudgementItem{
		CriteriaName: "test",
		Passed:       false,
		Score:        0.5,
		Details:      "fix this issue",
	}
	result.AddJudgement(item)

	if len(result.SuggestedFixes) != 1 {
		t.Errorf("expected 1 suggested fix, got %d", len(result.SuggestedFixes))
	}
}

func TestJudgementResult_Recalculate(t *testing.T) {
	tests := []struct {
		name           string
		items          []strategies.JudgementItem
		expectedScore  float64
		expectedPassed bool
	}{
		{
			name:           "empty",
			items:          []strategies.JudgementItem{},
			expectedScore:  0,
			expectedPassed: false,
		},
		{
			name: "all passing",
			items: []strategies.JudgementItem{
				{CriteriaName: "t1", Passed: true, Score: 0.8},
				{CriteriaName: "t2", Passed: true, Score: 0.9},
			},
			expectedScore:  0.85,
			expectedPassed: true,
		},
		{
			name: "some failing",
			items: []strategies.JudgementItem{
				{CriteriaName: "t1", Passed: true, Score: 0.8},
				{CriteriaName: "t2", Passed: false, Score: 0.3},
			},
			expectedScore:  0.8, // Only passing items count
			expectedPassed: true,
		},
		{
			name: "all failing",
			items: []strategies.JudgementItem{
				{CriteriaName: "t1", Passed: false, Score: 0.3},
				{CriteriaName: "t2", Passed: false, Score: 0.4},
			},
			expectedScore:  0, // No passing items
			expectedPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewJudgementResult()
			for _, item := range tt.items {
				result.AddJudgement(item)
			}

			// Use approximate comparison for floating point
			scoreDiff := result.Score - tt.expectedScore
			if scoreDiff < 0 {
				scoreDiff = -scoreDiff
			}
			if scoreDiff > 0.001 {
				t.Errorf("expected score ~%f, got %f", tt.expectedScore, result.Score)
			}
			if result.Passed != tt.expectedPassed {
				t.Errorf("expected passed %v, got %v", tt.expectedPassed, result.Passed)
			}
		})
	}
}

func TestJudgementResult_AddSuggestedFix(t *testing.T) {
	result := NewJudgementResult()
	result.AddSuggestedFix("fix 1")

	if len(result.SuggestedFixes) != 1 {
		t.Errorf("expected 1 fix, got %d", len(result.SuggestedFixes))
	}
}

func TestJudgementResult_SetMetadata(t *testing.T) {
	result := NewJudgementResult()
	result.SetMetadata("key", "value")

	if result.Metadata["key"] != "value" {
		t.Errorf("expected 'value', got %v", result.Metadata["key"])
	}
}

func TestJudgementResult_Clone(t *testing.T) {
	result := NewJudgementResult()
	result.AddJudgement(strategies.JudgementItem{
		CriteriaName: "test",
		Passed:       true,
		Score:        0.8,
	})
	result.AddSuggestedFix("fix")
	result.SetMetadata("key", "value")

	clone := result.Clone()

	if clone.Score != result.Score {
		t.Errorf("expected score %f, got %f", result.Score, clone.Score)
	}
	if len(clone.Judgements) != len(result.Judgements) {
		t.Errorf("expected %d judgements, got %d", len(result.Judgements), len(clone.Judgements))
	}
	if len(clone.SuggestedFixes) != len(result.SuggestedFixes) {
		t.Errorf("expected %d fixes, got %d", len(result.SuggestedFixes), len(clone.SuggestedFixes))
	}
}

func TestJudgementResult_Clone_Nil(t *testing.T) {
	var result *JudgementResult
	clone := result.Clone()

	if clone != nil {
		t.Error("expected nil clone for nil result")
	}
}
