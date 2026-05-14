package provider

import (
	"context"
	"fmt"
	"strings"
)

// QualityGate determines whether a model response should trigger escalation.
type QualityGate interface {
	ShouldEscalate(response string) bool
}

// ThresholdGate is a QualityGate that checks minimum length and required keys.
type ThresholdGate struct {
	MinLength    int
	RequiredKeys []string
}

func (g *ThresholdGate) ShouldEscalate(response string) bool {
	if len(response) < g.MinLength {
		return true
	}
	for _, key := range g.RequiredKeys {
		if !strings.Contains(response, key) {
			return true
		}
	}
	return false
}

// EscalationProvider routes to a cheap model first, escalating on quality failure.
type EscalationProvider struct {
	primary   ModelProvider
	escalated ModelProvider
	gate      QualityGate
}

// NewEscalationProvider creates an EscalationProvider with the given primary, escalated, and gate.
func NewEscalationProvider(primary, escalated ModelProvider, gate QualityGate) *EscalationProvider {
	return &EscalationProvider{primary: primary, escalated: escalated, gate: gate}
}

func (p *EscalationProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	resp, err := p.primary.Execute(ctx, req)
	if err != nil {
		return p.escalated.Execute(ctx, req)
	}
	if p.gate.ShouldEscalate(responseText(resp)) {
		return p.escalated.Execute(ctx, req)
	}
	return resp, nil
}

func responseText(resp *ModelResponse) string {
	if resp == nil || resp.Output == nil {
		return ""
	}
	if text, ok := resp.Output["text"]; ok {
		return fmt.Sprintf("%v", text)
	}
	// Fallback: concatenate all string values
	var parts []string
	for _, v := range resp.Output {
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return strings.Join(parts, " ")
}
