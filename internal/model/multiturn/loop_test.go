package multiturn

import (
	"context"
	"fmt"
	"testing"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

type mockProvider struct {
	responses []*provider.ModelResponse
	callCount int
}

func (m *mockProvider) Execute(_ context.Context, _ *provider.ModelRequest) (*provider.ModelResponse, error) {
	if m.callCount >= len(m.responses) {
		return &provider.ModelResponse{Output: map[string]any{"done": true}}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

type mockTool struct {
	name   string
	result map[string]any
	err    error
}

func (t *mockTool) Name() string                                                    { return t.name }
func (t *mockTool) Schema() types.ToolDefinition                                    { return types.ToolDefinition{Name: t.name} }
func (t *mockTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) { return t.result, t.err }

func reg(tools ...tool.Tool) *tool.Registry {
	r := tool.NewRegistry()
	for _, t := range tools {
		_ = r.Register(t, nil)
	}
	return r
}

// dynamicMockTool returns a different result each call to avoid runaway detection.
type dynamicMockTool struct {
	name  string
	calls int
}

func (t *dynamicMockTool) Name() string                 { return t.name }
func (t *dynamicMockTool) Schema() types.ToolDefinition { return types.ToolDefinition{Name: t.name} }
func (t *dynamicMockTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) {
	t.calls++
	return map[string]any{"call": t.calls}, nil
}

func TestRun_SingleTurn(t *testing.T) {
	p := &mockProvider{responses: []*provider.ModelResponse{{Output: map[string]any{"ok": true}}}}
	res, err := Run(context.Background(), LoopConfig{Provider: p, Tools: reg(), MaxIterations: 5, MaxErrors: 3}, &provider.ModelRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Output["ok"] != true {
		t.Errorf("got %v", res.Output)
	}
}

func TestRun_ToolLoop(t *testing.T) {
	p := &mockProvider{responses: []*provider.ModelResponse{
		{ToolCalls: []types.ToolCall{{ID: "1", Name: "bash", Input: map[string]any{}}}},
		{Output: map[string]any{"done": true}},
	}}
	res, err := Run(context.Background(), LoopConfig{
		Provider: p, Tools: reg(&mockTool{name: "bash", result: map[string]any{"out": "hi"}}),
		MaxIterations: 10, MaxErrors: 3,
	}, &provider.ModelRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Output["done"] != true {
		t.Errorf("got %v", res.Output)
	}
}

func TestRun_CircuitBreaker(t *testing.T) {
	resps := make([]*provider.ModelResponse, 10)
	for i := range resps {
		resps[i] = &provider.ModelResponse{ToolCalls: []types.ToolCall{{ID: fmt.Sprintf("c%d", i), Name: "missing"}}}
	}
	p := &mockProvider{responses: resps}
	res, _ := Run(context.Background(), LoopConfig{Provider: p, Tools: reg(), MaxIterations: 10, MaxErrors: 3}, &provider.ModelRequest{})
	if res.Error == "" {
		t.Fatal("expected circuit breaker error")
	}
}

func TestRun_IterationBudget(t *testing.T) {
	resps := make([]*provider.ModelResponse, 20)
	for i := range resps {
		resps[i] = &provider.ModelResponse{ToolCalls: []types.ToolCall{{ID: fmt.Sprintf("c%d", i), Name: "bash"}}}
	}
	p := &mockProvider{responses: resps}
	callNum := 0
	countingTool := &mockTool{name: "bash"}
	countingTool.result = nil // will be set dynamically
	// Use a tool that returns different output each time to avoid runaway detection
	dynamicTool := &dynamicMockTool{name: "bash"}
	res, _ := Run(context.Background(), LoopConfig{
		Provider: p, Tools: reg(dynamicTool),
		MaxIterations: 5, MaxErrors: 10,
	}, &provider.ModelRequest{})
	_ = callNum
	if res.Error == "" {
		t.Fatal("expected budget error")
	}
	if p.callCount != 5 {
		t.Errorf("expected 5 calls, got %d", p.callCount)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &mockProvider{responses: []*provider.ModelResponse{{Output: map[string]any{}}}}
	res, err := Run(ctx, LoopConfig{Provider: p, Tools: reg(), MaxIterations: 5, MaxErrors: 3}, &provider.ModelRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestRun_RunawayDetection(t *testing.T) {
	resps := make([]*provider.ModelResponse, 10)
	for i := range resps {
		resps[i] = &provider.ModelResponse{ToolCalls: []types.ToolCall{{ID: fmt.Sprintf("c%d", i), Name: "bash"}}}
	}
	p := &mockProvider{responses: resps}
	// Tool always returns identical output → should trigger runaway after 3
	res, _ := Run(context.Background(), LoopConfig{
		Provider: p, Tools: reg(&mockTool{name: "bash", result: map[string]any{"same": "output"}}),
		MaxIterations: 10, MaxErrors: 10,
	}, &provider.ModelRequest{})
	if res.Error == "" || res.Error != "runaway detected: last 3 tool outputs identical" {
		t.Fatalf("expected runaway error, got: %q", res.Error)
	}
	if p.callCount > 4 {
		t.Errorf("expected early termination, got %d calls", p.callCount)
	}
}

func TestRun_OnToolExecutedHook(t *testing.T) {
	p := &mockProvider{responses: []*provider.ModelResponse{
		{ToolCalls: []types.ToolCall{{ID: "1", Name: "bash", Input: map[string]any{}}}},
		{Output: map[string]any{}},
	}}
	var called bool
	_, _ = Run(context.Background(), LoopConfig{
		Provider: p, Tools: reg(&mockTool{name: "bash", result: map[string]any{"x": 1}}),
		MaxIterations:  5,
		MaxErrors:      3,
		OnToolExecuted: func(name string, result map[string]any, err error) { called = true },
	}, &provider.ModelRequest{})
	if !called {
		t.Error("hook not called")
	}
}

func TestClosePendingToolCalls(t *testing.T) {
	// No pending calls
	h := []types.ModelMessage{{Role: "user", Content: "hi"}}
	got := closePendingToolCalls(h)
	if len(got) != 1 {
		t.Fatal("should not modify history without pending tool calls")
	}

	// Pending tool call
	h = []types.ModelMessage{
		{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "tc-1", Name: "bash"}}},
	}
	got = closePendingToolCalls(h)
	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got))
	}
	if got[1].Role != "tool" || got[1].ToolCallID != "tc-1" {
		t.Fatal("synthetic result not correct")
	}
}


func TestRun_CostGuardAborts(t *testing.T) {
	// Provider returns tool calls with cost, CostGuard aborts after 2nd call
	resps := []*provider.ModelResponse{
		{ToolCalls: []types.ToolCall{{ID: "c1", Name: "bash"}}, CostEstimateUSD: 0.05},
		{ToolCalls: []types.ToolCall{{ID: "c2", Name: "bash"}}, CostEstimateUSD: 0.06},
		{Output: map[string]any{"done": true}},
	}
	p := &mockProvider{responses: resps}

	totalCost := 0.0
	res, _ := Run(context.Background(), LoopConfig{
		Provider:      p,
		Tools:         reg(&dynamicMockTool{name: "bash"}),
		MaxIterations: 10,
		MaxErrors:     3,
		CostGuard: func(costUSD float64) error {
			totalCost += costUSD
			if totalCost >= 0.10 {
				return fmt.Errorf("[COST_BUDGET_EXCEEDED] budget exceeded")
			}
			return nil
		},
	}, &provider.ModelRequest{})

	if res.Error == "" {
		t.Fatal("expected cost guard error")
	}
	if p.callCount != 2 {
		t.Errorf("expected 2 provider calls, got %d", p.callCount)
	}
}
