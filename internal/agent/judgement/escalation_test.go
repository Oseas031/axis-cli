package judgement

import (
	"testing"

	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
)

// trackingStrategy records whether Validate was called.
type trackingStrategy struct {
	called bool
	score  float64
}

func (s *trackingStrategy) Validate(input any, criteria strategies.JudgementCriteria) (*strategies.JudgementItem, error) {
	s.called = true
	return &strategies.JudgementItem{
		CriteriaName: criteria.Name,
		Passed:       s.score >= DefaultJudgementThresholds.PassingScore,
		Score:        s.score,
		Details:      "tracking",
	}, nil
}

func (s *trackingStrategy) CanHandle(criteria strategies.JudgementCriteria) bool {
	return true
}

func TestEscalatingEngine_LightweightPasses_NoEscalation(t *testing.T) {
	ee := NewEscalatingEngine(nil, nil, 0.8)

	// Replace lightweight with a high-scoring strategy
	lightTracker := &trackingStrategy{score: 0.95}
	ee.lightweight.RegisterStrategy(strategies.JudgementTypeSyntax, lightTracker)

	// Replace full with a tracker to verify it's NOT called
	fullTracker := &trackingStrategy{score: 1.0}
	ee.full.RegisterStrategy(strategies.JudgementTypeSyntax, fullTracker)

	criteria := []strategies.JudgementCriteria{
		{Name: "syntax", Type: strategies.JudgementTypeSyntax, Weight: 1.0, Enabled: true},
	}

	result, err := ee.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lightTracker.called {
		t.Error("expected lightweight to be called")
	}
	if fullTracker.called {
		t.Error("expected full engine NOT to be called when lightweight passes")
	}
	if result.Metadata["escalated"] != false {
		t.Error("expected escalated=false in metadata")
	}
}

func TestEscalatingEngine_LightweightFails_Escalates(t *testing.T) {
	ee := NewEscalatingEngine(nil, nil, 0.8)

	// Lightweight returns low score
	lightTracker := &trackingStrategy{score: 0.5}
	ee.lightweight.RegisterStrategy(strategies.JudgementTypeSyntax, lightTracker)

	// Full engine should be called
	fullTracker := &trackingStrategy{score: 0.9}
	ee.full.RegisterStrategy(strategies.JudgementTypeSyntax, fullTracker)

	criteria := []strategies.JudgementCriteria{
		{Name: "syntax", Type: strategies.JudgementTypeSyntax, Weight: 1.0, Enabled: true},
	}

	result, err := ee.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lightTracker.called {
		t.Error("expected lightweight to be called")
	}
	if !fullTracker.called {
		t.Error("expected full engine to be called on escalation")
	}
	if result.Metadata["escalated"] != true {
		t.Error("expected escalated=true in metadata")
	}
}

func TestEscalatingEngine_ThresholdBoundary(t *testing.T) {
	threshold := 0.8

	// Exactly at threshold → no escalation
	ee := NewEscalatingEngine(nil, nil, threshold)
	lightTracker := &trackingStrategy{score: 0.8}
	ee.lightweight.RegisterStrategy(strategies.JudgementTypeSyntax, lightTracker)
	fullTracker := &trackingStrategy{score: 1.0}
	ee.full.RegisterStrategy(strategies.JudgementTypeSyntax, fullTracker)

	criteria := []strategies.JudgementCriteria{
		{Name: "syntax", Type: strategies.JudgementTypeSyntax, Weight: 1.0, Enabled: true},
	}

	result, err := ee.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fullTracker.called {
		t.Error("expected no escalation at exact threshold")
	}
	if result.Metadata["escalated"] != false {
		t.Error("expected escalated=false at exact threshold")
	}

	// Just below threshold → escalation
	ee2 := NewEscalatingEngine(nil, nil, threshold)
	lightTracker2 := &trackingStrategy{score: 0.79}
	ee2.lightweight.RegisterStrategy(strategies.JudgementTypeSyntax, lightTracker2)
	fullTracker2 := &trackingStrategy{score: 1.0}
	ee2.full.RegisterStrategy(strategies.JudgementTypeSyntax, fullTracker2)

	result2, err := ee2.Judge(nil, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fullTracker2.called {
		t.Error("expected escalation below threshold")
	}
	if result2.Metadata["escalated"] != true {
		t.Error("expected escalated=true below threshold")
	}
}
