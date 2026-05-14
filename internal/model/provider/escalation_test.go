package provider

import (
	"context"
	"errors"
	"testing"
)

type mockProvider struct {
	resp *ModelResponse
	err  error
	calls int
}

func (m *mockProvider) Execute(_ context.Context, _ *ModelRequest) (*ModelResponse, error) {
	m.calls++
	return m.resp, m.err
}

func TestEscalation_PrimaryPassesGate(t *testing.T) {
	primary := &mockProvider{resp: &ModelResponse{Output: map[string]any{"text": "hello world result"}}}
	escalated := &mockProvider{resp: &ModelResponse{Output: map[string]any{"text": "escalated"}}}
	gate := &ThresholdGate{MinLength: 5}

	ep := NewEscalationProvider(primary, escalated, gate)
	resp, err := ep.Execute(context.Background(), &ModelRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Output["text"] != "hello world result" {
		t.Fatalf("expected primary response, got %v", resp.Output["text"])
	}
	if escalated.calls != 0 {
		t.Fatal("escalated should not have been called")
	}
}

func TestEscalation_PrimaryFailsGate(t *testing.T) {
	primary := &mockProvider{resp: &ModelResponse{Output: map[string]any{"text": "hi"}}}
	escalated := &mockProvider{resp: &ModelResponse{Output: map[string]any{"text": "full escalated response"}}}
	gate := &ThresholdGate{MinLength: 10}

	ep := NewEscalationProvider(primary, escalated, gate)
	resp, err := ep.Execute(context.Background(), &ModelRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Output["text"] != "full escalated response" {
		t.Fatalf("expected escalated response, got %v", resp.Output["text"])
	}
	if primary.calls != 1 || escalated.calls != 1 {
		t.Fatalf("expected 1 call each, got primary=%d escalated=%d", primary.calls, escalated.calls)
	}
}

func TestEscalation_PrimaryError_FallsToEscalated(t *testing.T) {
	primary := &mockProvider{err: errors.New("timeout")}
	escalated := &mockProvider{resp: &ModelResponse{Output: map[string]any{"text": "fallback"}}}
	gate := &ThresholdGate{MinLength: 1}

	ep := NewEscalationProvider(primary, escalated, gate)
	resp, err := ep.Execute(context.Background(), &ModelRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Output["text"] != "fallback" {
		t.Fatalf("expected fallback response, got %v", resp.Output["text"])
	}
}

func TestEscalation_BothFail_ReturnsEscalatedResponse(t *testing.T) {
	primary := &mockProvider{resp: &ModelResponse{Output: map[string]any{"text": "x"}}}
	escalated := &mockProvider{resp: &ModelResponse{Output: map[string]any{"text": "y"}}}
	gate := &ThresholdGate{MinLength: 100} // both responses too short

	ep := NewEscalationProvider(primary, escalated, gate)
	resp, err := ep.Execute(context.Background(), &ModelRequest{})
	if err != nil {
		t.Fatal(err)
	}
	// Returns escalated response regardless of gate result
	if resp.Output["text"] != "y" {
		t.Fatalf("expected escalated response anyway, got %v", resp.Output["text"])
	}
}

func TestThresholdGate_RequiredKeys(t *testing.T) {
	gate := &ThresholdGate{MinLength: 1, RequiredKeys: []string{"status", "result"}}
	if !gate.ShouldEscalate("only status here") {
		t.Fatal("should escalate when missing 'result' key")
	}
	if gate.ShouldEscalate("status and result present") {
		t.Fatal("should not escalate when both keys present")
	}
}
