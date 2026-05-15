package judgement

import (
	"testing"

	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
)

func TestDetectSovereigntyGap_NormalCase(t *testing.T) {
	result := &JudgementResult{
		Passed: true,
		Judgements: []strategies.JudgementItem{
			{Score: 0.9, Passed: true},
			{Score: 0.8, Passed: true},
		},
	}
	sgr := DetectSovereigntyGap(result)
	if sgr.AlignmentHallucinationSuspected {
		t.Error("expected no alignment hallucination suspected")
	}
	if sgr.EscalationRequired {
		t.Error("expected no escalation required")
	}
	// Gap should be negative (0.85 - 1.0 = -0.15), within threshold
	if sgr.Gap > SovereigntyGapThreshold || sgr.Gap < -SovereigntyGapThreshold {
		t.Errorf("expected gap within threshold, got %v", sgr.Gap)
	}
}

func TestDetectSovereigntyGap_PositiveGap(t *testing.T) {
	// High internal scores but failed overall → alignment hallucination
	result := &JudgementResult{
		Passed: false,
		Judgements: []strategies.JudgementItem{
			{Score: 0.8, Passed: true},
			{Score: 0.7, Passed: true},
		},
	}
	sgr := DetectSovereigntyGap(result)
	if sgr.Gap <= SovereigntyGapThreshold {
		t.Errorf("expected gap > %v, got %v", SovereigntyGapThreshold, sgr.Gap)
	}
	if !sgr.AlignmentHallucinationSuspected {
		t.Error("expected alignment hallucination suspected")
	}
	if !sgr.EscalationRequired {
		t.Error("expected escalation required")
	}
}

func TestDetectSovereigntyGap_NegativeGap(t *testing.T) {
	// Low internal scores but passed overall → suspicious
	result := &JudgementResult{
		Passed: true,
		Judgements: []strategies.JudgementItem{
			{Score: 0.2, Passed: false},
			{Score: 0.3, Passed: false},
		},
	}
	sgr := DetectSovereigntyGap(result)
	if sgr.Gap >= -SovereigntyGapThreshold {
		t.Errorf("expected gap < %v, got %v", -SovereigntyGapThreshold, sgr.Gap)
	}
	if sgr.AlignmentHallucinationSuspected {
		t.Error("expected no alignment hallucination suspected for negative gap")
	}
	if !sgr.EscalationRequired {
		t.Error("expected escalation required")
	}
}

func TestDetectSovereigntyGap_NilResult(t *testing.T) {
	sgr := DetectSovereigntyGap(nil)
	if sgr == nil {
		t.Fatal("expected non-nil result for nil input")
	}
	if sgr.Gap != 0 {
		t.Errorf("expected zero gap, got %v", sgr.Gap)
	}
	if sgr.EscalationRequired {
		t.Error("expected no escalation for nil input")
	}
}

func TestDetectSovereigntyGap_EmptyJudgements(t *testing.T) {
	result := &JudgementResult{
		Passed:     false,
		Judgements: nil,
	}
	sgr := DetectSovereigntyGap(result)
	if sgr.InternalValidity != 0 {
		t.Errorf("expected zero internal validity for empty judgements, got %v", sgr.InternalValidity)
	}
	if sgr.EscalationRequired {
		t.Error("expected no escalation for empty judgements with failed result")
	}
}
