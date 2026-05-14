package provider

import (
	"context"
	"testing"
)

type trackingProvider struct {
	called bool
	name   string
}

func (p *trackingProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	p.called = true
	return &ModelResponse{Output: map[string]any{"provider": p.name}}, nil
}

func TestLayeredProvider_RoutesToUtility(t *testing.T) {
	primary := &trackingProvider{name: "primary"}
	utility := &trackingProvider{name: "utility"}
	lp := NewLayeredProvider(primary, utility)

	req := &ModelRequest{
		Input:    map[string]any{"prompt": "summarize this"},
		Metadata: map[string]any{"request_type": "summarize"},
	}
	resp, err := lp.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !utility.called {
		t.Error("expected utility provider to be called")
	}
	if primary.called {
		t.Error("expected primary provider NOT to be called")
	}
	if resp.Output["provider"] != "utility" {
		t.Errorf("expected utility, got %v", resp.Output["provider"])
	}
}

func TestLayeredProvider_RoutesToPrimary(t *testing.T) {
	primary := &trackingProvider{name: "primary"}
	utility := &trackingProvider{name: "utility"}
	lp := NewLayeredProvider(primary, utility)

	req := &ModelRequest{
		Input: map[string]any{"prompt": "generate code"},
	}
	resp, err := lp.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !primary.called {
		t.Error("expected primary provider to be called")
	}
	if utility.called {
		t.Error("expected utility provider NOT to be called")
	}
	if resp.Output["provider"] != "primary" {
		t.Errorf("expected primary, got %v", resp.Output["provider"])
	}
}

func TestLayeredProvider_ClassifyIsUtility(t *testing.T) {
	primary := &trackingProvider{name: "primary"}
	utility := &trackingProvider{name: "utility"}
	lp := NewLayeredProvider(primary, utility)

	req := &ModelRequest{
		Input:    map[string]any{"prompt": "classify this"},
		Metadata: map[string]any{"request_type": "classify"},
	}
	_, err := lp.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !utility.called {
		t.Error("expected utility provider to be called for classify")
	}
}
