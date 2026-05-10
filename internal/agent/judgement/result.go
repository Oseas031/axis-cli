// Package judgement provides self-judgement capabilities for agent execution validation.
package judgement

import (
	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
)

// JudgementResult contains the aggregated result of all judgement evaluations.
type JudgementResult struct {
	Passed         bool                       `json:"passed"`
	Score          float64                    `json:"score"`
	Judgements     []strategies.JudgementItem `json:"judgements"`
	Confidence     float64                    `json:"confidence"`
	SuggestedFixes []string                   `json:"suggested_fixes"`
	Metadata       map[string]any             `json:"metadata"`
}

// DefaultJudgementThresholds contains the default thresholds for judging.
var DefaultJudgementThresholds = struct {
	MinCoverage     float64
	MinTestPassRate float64
	MinConfidence   float64
	PassingScore    float64
}{
	MinCoverage:     0.85,
	MinTestPassRate: 0.90,
	MinConfidence:   0.70,
	PassingScore:    0.75,
}

// NewJudgementResult creates a new JudgementResult with initialized fields.
func NewJudgementResult() *JudgementResult {
	return &JudgementResult{
		Judgements:     make([]strategies.JudgementItem, 0),
		SuggestedFixes: make([]string, 0),
		Metadata:       make(map[string]any),
	}
}

// AddJudgement adds a judgement item to the result and recalculates aggregates.
func (r *JudgementResult) AddJudgement(item strategies.JudgementItem) {
	r.Judgements = append(r.Judgements, item)
	if !item.Passed && item.Details != "" {
		r.SuggestedFixes = append(r.SuggestedFixes, item.Details)
	}
	r.recalculate()
}

// recalculate recalculates the score and confidence based on all judgements.
func (r *JudgementResult) recalculate() {
	if len(r.Judgements) == 0 {
		r.Score = 0
		r.Confidence = 0
		r.Passed = false
		return
	}

	// Calculate score as average of passed items
	passedCount := 0
	totalScore := 0.0
	for _, j := range r.Judgements {
		if j.Passed {
			passedCount++
			totalScore += j.Score
		}
	}

	if passedCount > 0 {
		r.Score = totalScore / float64(passedCount)
	} else {
		r.Score = 0
	}

	// Calculate confidence based on agreement between strategies
	r.Confidence = float64(passedCount) / float64(len(r.Judgements))

	// Passed if score meets threshold
	r.Passed = r.Score >= DefaultJudgementThresholds.PassingScore
}

// AddSuggestedFix adds a suggested fix to the result.
func (r *JudgementResult) AddSuggestedFix(fix string) {
	r.SuggestedFixes = append(r.SuggestedFixes, fix)
}

// SetMetadata sets a metadata value.
func (r *JudgementResult) SetMetadata(key string, value any) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]any)
	}
	r.Metadata[key] = value
}

// Clone creates a deep copy of the result.
func (r *JudgementResult) Clone() *JudgementResult {
	if r == nil {
		return nil
	}
	clone := &JudgementResult{
		Passed:     r.Passed,
		Score:      r.Score,
		Confidence: r.Confidence,
	}
	if r.Judgements != nil {
		clone.Judgements = make([]strategies.JudgementItem, len(r.Judgements))
		copy(clone.Judgements, r.Judgements)
	}
	if r.SuggestedFixes != nil {
		clone.SuggestedFixes = make([]string, len(r.SuggestedFixes))
		copy(clone.SuggestedFixes, r.SuggestedFixes)
	}
	if r.Metadata != nil {
		clone.Metadata = make(map[string]any, len(r.Metadata))
		for k, v := range r.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}

// CriteriaFromStrategies converts strategies.JudgementCriteria to the internal format.
// This is a compatibility helper.
func CriteriaFromStrategies(c []strategies.JudgementCriteria) []strategies.JudgementCriteria {
	return c
}
