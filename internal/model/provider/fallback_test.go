package provider

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFallbackProvider_SwitchesOn429(t *testing.T) {
	primary := &rateLimitMock{err: fmt.Errorf("API error (status 429): rate limit exceeded")}
	fallback := &rateLimitMock{resp: &ModelResponse{Output: map[string]any{"text": "from fallback"}}}

	fp := NewFallbackProvider(time.Second, primary, fallback)
	resp, err := fp.Execute(context.Background(), &ModelRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output["text"] != "from fallback" {
		t.Fatalf("expected fallback response, got %v", resp.Output["text"])
	}
}

func TestFallbackProvider_NoSwitchOnOtherErrors(t *testing.T) {
	primary := &rateLimitMock{err: fmt.Errorf("API error (status 400): bad request")}
	fallback := &rateLimitMock{resp: &ModelResponse{Output: map[string]any{"text": "from fallback"}}}

	fp := NewFallbackProvider(time.Second, primary, fallback)
	_, err := fp.Execute(context.Background(), &ModelRequest{})
	if err == nil {
		t.Fatal("expected error to propagate")
	}
}

func TestFallbackProvider_CooldownRestoresPrimary(t *testing.T) {
	callCount := 0
	primary := &rateLimitMock{execFn: func() (*ModelResponse, error) {
		callCount++
		if callCount == 1 {
			return nil, fmt.Errorf("API error (status 429): rate limit")
		}
		return &ModelResponse{Output: map[string]any{"text": "primary recovered"}}, nil
	}}
	fallback := &rateLimitMock{resp: &ModelResponse{Output: map[string]any{"text": "fallback"}}}

	fp := NewFallbackProvider(10*time.Millisecond, primary, fallback)

	// First call: 429 → fallback
	resp, _ := fp.Execute(context.Background(), &ModelRequest{})
	if resp.Output["text"] != "fallback" {
		t.Fatalf("expected fallback, got %v", resp.Output["text"])
	}

	// Wait for cooldown
	time.Sleep(15 * time.Millisecond)

	// Second call: primary recovered
	resp, err := fp.Execute(context.Background(), &ModelRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output["text"] != "primary recovered" {
		t.Fatalf("expected primary recovered, got %v", resp.Output["text"])
	}
}

// rateLimitMock is a test mock for FallbackProvider tests.
type rateLimitMock struct {
	resp   *ModelResponse
	err    error
	execFn func() (*ModelResponse, error)
}

func (m *rateLimitMock) Execute(_ context.Context, _ *ModelRequest) (*ModelResponse, error) {
	if m.execFn != nil {
		return m.execFn()
	}
	return m.resp, m.err
}
