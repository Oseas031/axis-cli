package judgement

// v1: threshold-based detection using strategy scores as proxy for internal validity.
// TODO: CoT semantic analysis — compare Judge's reasoning trace against its verdict for true Sovereignty Gap.

// SovereigntyGapThreshold is the divergence threshold above which escalation is triggered.
// Derived from arXiv:2605.10698 where gaps of +0.34 to +0.50 indicated systematic
// Alignment Hallucination in multi-agent settings.
const SovereigntyGapThreshold = 0.3

// SovereigntyGapResult captures the divergence between internal reasoning validity
// and external output accuracy, per arXiv:2605.10698.
type SovereigntyGapResult struct {
	InternalValidity                float64 `json:"internal_validity"`
	ExternalAccuracy                float64 `json:"external_accuracy"`
	Gap                             float64 `json:"gap"`
	AlignmentHallucinationSuspected bool    `json:"alignment_hallucination_suspected"`
	EscalationRequired              bool    `json:"escalation_required"`
}

// DetectSovereigntyGap computes the gap between internal judgement scores and the
// final pass/fail verdict. A large positive gap suggests alignment hallucination
// (model internally correct but externally compliant with swarm). A large negative
// gap suggests unwarranted pass (low internal confidence but passed anyway).
func DetectSovereigntyGap(result *JudgementResult) *SovereigntyGapResult {
	if result == nil {
		return &SovereigntyGapResult{}
	}

	var internalValidity float64
	if len(result.Judgements) > 0 {
		var sum float64
		for _, j := range result.Judgements {
			sum += j.Score
		}
		internalValidity = sum / float64(len(result.Judgements))
	}

	var externalAccuracy float64
	if result.Passed {
		externalAccuracy = 1.0
	}

	gap := internalValidity - externalAccuracy

	sgr := &SovereigntyGapResult{
		InternalValidity: internalValidity,
		ExternalAccuracy: externalAccuracy,
		Gap:              gap,
	}

	if gap > SovereigntyGapThreshold {
		sgr.AlignmentHallucinationSuspected = true
		sgr.EscalationRequired = true
	} else if gap < -SovereigntyGapThreshold {
		sgr.EscalationRequired = true
	}

	return sgr
}
