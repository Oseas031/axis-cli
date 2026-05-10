// Package judgement provides self-judgement capabilities for agent execution validation.
package judgement

import (
	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
)

// Re-export types from strategies package for convenience.
type JudgementType = strategies.JudgementType

const (
	JudgementTypeSyntax   = strategies.JudgementTypeSyntax
	JudgementTypeSemantic = strategies.JudgementTypeSemantic
	JudgementTypeContract = strategies.JudgementTypeContract
	JudgementTypeTest     = strategies.JudgementTypeTest
	JudgementTypeCoverage = strategies.JudgementTypeCoverage
	JudgementTypeCustom   = strategies.JudgementTypeCustom
)

// JudgementCriteria is re-exported from strategies.
type JudgementCriteria = strategies.JudgementCriteria

// DefaultJudgementCriteria returns a set of default criteria for agent execution.
func DefaultJudgementCriteria() []JudgementCriteria {
	return []JudgementCriteria{
		{
			Name:   "syntax_validation",
			Type:   JudgementTypeSyntax,
			Weight: 0.15,
			Thresholds: map[string]float64{
				"min_pass_rate": 1.0,
			},
			Enabled: true,
		},
		{
			Name:   "test_validation",
			Type:   JudgementTypeTest,
			Weight: 0.30,
			Thresholds: map[string]float64{
				"min_pass_rate": 0.90,
			},
			Enabled: true,
		},
		{
			Name:   "coverage_validation",
			Type:   JudgementTypeCoverage,
			Weight: 0.25,
			Thresholds: map[string]float64{
				"min_coverage": 0.85,
			},
			Enabled: true,
		},
		{
			Name:   "contract_validation",
			Type:   JudgementTypeContract,
			Weight: 0.30,
			Thresholds: map[string]float64{
				"min_pass_rate": 1.0,
			},
			Enabled: true,
		},
	}
}

// NewJudgementCriteria creates a new JudgementCriteria with defaults.
func NewJudgementCriteria(name string, judgementType JudgementType, weight float64) JudgementCriteria {
	return JudgementCriteria{
		Name:       name,
		Type:       judgementType,
		Weight:     weight,
		Thresholds: make(map[string]float64),
		Enabled:    true,
	}
}
