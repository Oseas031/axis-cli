// Package judgement provides self-judgement capabilities for agent execution validation.
package judgement

import (
	"testing"

	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
)

func TestDefaultJudgementCriteria(t *testing.T) {
	criteria := DefaultJudgementCriteria()

	if len(criteria) != 4 {
		t.Errorf("expected 4 default criteria, got %d", len(criteria))
	}

	// Verify each criteria type exists
	typeMap := make(map[strategies.JudgementType]bool)
	for _, c := range criteria {
		if !c.Enabled {
			t.Errorf("expected criteria %s to be enabled", c.Name)
		}
		if c.Weight <= 0 || c.Weight > 1 {
			t.Errorf("expected weight between 0 and 1 for %s, got %f", c.Name, c.Weight)
		}
		typeMap[c.Type] = true
	}

	expectedTypes := []strategies.JudgementType{
		strategies.JudgementTypeSyntax,
		strategies.JudgementTypeTest,
		strategies.JudgementTypeCoverage,
		strategies.JudgementTypeContract,
	}

	for _, et := range expectedTypes {
		if !typeMap[et] {
			t.Errorf("expected criteria type %s not found", et)
		}
	}
}

func TestNewJudgementCriteria(t *testing.T) {
	c := NewJudgementCriteria("test", strategies.JudgementTypeSyntax, 0.5)

	if c.Name != "test" {
		t.Errorf("expected name 'test', got %s", c.Name)
	}
	if c.Type != strategies.JudgementTypeSyntax {
		t.Errorf("expected type syntax, got %s", c.Type)
	}
	if c.Weight != 0.5 {
		t.Errorf("expected weight 0.5, got %f", c.Weight)
	}
	if !c.Enabled {
		t.Error("expected enabled to be true")
	}
	if c.Thresholds == nil {
		t.Error("expected thresholds to be initialized")
	}
}
