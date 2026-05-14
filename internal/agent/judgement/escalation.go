package judgement

import (
	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
	"github.com/axis-cli/axis/internal/model/provider"
)

// EscalatingEngine wraps Engine with two-pass judgement.
// Lightweight pass runs cheap strategies (Syntax, Test, Coverage).
// Full pass includes all strategies (adds Semantic, Contract).
// If lightweight score >= threshold, full pass is skipped.
type EscalatingEngine struct {
	lightweight *Engine
	full        *Engine
	threshold   float64
}

// NewEscalatingEngine creates an EscalatingEngine with the given threshold.
func NewEscalatingEngine(modelProvider provider.ModelProvider, contractExec strategies.ContractExecutorProvider, threshold float64) *EscalatingEngine {
	return &EscalatingEngine{
		lightweight: NewEngine(), // Syntax + Test + Coverage only
		full:        NewEngineWithProviders(modelProvider, contractExec),
		threshold:   threshold,
	}
}

// Judge runs lightweight first. If score >= threshold, returns early.
// Otherwise escalates to full engine.
func (ee *EscalatingEngine) Judge(input any, criteria []strategies.JudgementCriteria) (*JudgementResult, error) {
	// Filter criteria to only those the lightweight engine can handle
	lightCriteria := filterByRegistered(ee.lightweight, criteria)
	if len(lightCriteria) == 0 {
		// No lightweight-compatible criteria; go straight to full
		fullResult, err := ee.full.Judge(input, criteria)
		if err != nil {
			return nil, err
		}
		fullResult.SetMetadata("escalated", true)
		return fullResult, nil
	}

	result, err := ee.lightweight.Judge(input, lightCriteria)
	if err != nil {
		return nil, err
	}
	if result.Score >= ee.threshold {
		result.SetMetadata("escalated", false)
		return result, nil
	}
	fullResult, err := ee.full.Judge(input, criteria)
	if err != nil {
		return nil, err
	}
	fullResult.SetMetadata("escalated", true)
	return fullResult, nil
}

// filterByRegistered returns only criteria whose type is registered in the engine.
func filterByRegistered(e *Engine, criteria []strategies.JudgementCriteria) []strategies.JudgementCriteria {
	var filtered []strategies.JudgementCriteria
	for _, c := range criteria {
		if _, ok := e.GetStrategy(c.Type); ok {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
