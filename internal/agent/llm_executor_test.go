package agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// mockProvider implements provider.ModelProvider for testing.
type mockLLMProvider struct {
	responses []*provider.ModelResponse
	callCount int
}

func (m *mockLLMProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	if m.callCount >= len(m.responses) {
		return &provider.ModelResponse{Output: map[string]any{"done": true}}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

// mockTool implements tool.Tool for testing.
type mockToolImpl struct {
	name   string
	result map[string]any
	err    error
	execFn func(map[string]any) (map[string]any, error)
}

func (t *mockToolImpl) Name() string { return t.name }
func (t *mockToolImpl) Schema() types.ToolDefinition {
	return types.ToolDefinition{Name: t.name, Description: "mock tool"}
}
func (t *mockToolImpl) Execute(_ context.Context, input map[string]any) (map[string]any, error) {
	if t.execFn != nil {
		return t.execFn(input)
	}
	if t.err != nil {
		return nil, t.err
	}
	return t.result, nil
}

func newTestRegistry(tools ...tool.Tool) *tool.Registry {
	r := tool.NewRegistry()
	for _, t := range tools {
		_ = r.Register(t, nil)
	}
	return r
}

func TestLLMAgentExecutor_SingleTurnNoTools(t *testing.T) {
	p := &mockLLMProvider{
		responses: []*provider.ModelResponse{
			{Output: map[string]any{"result": "hello"}},
		},
	}
	reg := newTestRegistry()
	exec := NewLLMAgentExecutor(p, reg)

	req := &AgentExecutionRequest{
		Task: &types.AgentTask{TaskID: "t1", ContractID: "default", Input: map[string]any{}},
	}
	result, err := exec.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.Output["result"] != "hello" {
		t.Errorf("expected result=hello, got %v", result.Output["result"])
	}
	if result.AgentID != "default" {
		t.Errorf("expected AgentID=default, got %s", result.AgentID)
	}
}

func TestLLMAgentExecutor_MultiTurnToolLoop(t *testing.T) {
	p := &mockLLMProvider{
		responses: []*provider.ModelResponse{
			{ToolCalls: []types.ToolCall{{ID: "c1", Name: "bash", Input: map[string]any{"cmd": "echo hi"}}}},
			{Output: map[string]any{"result": "done"}},
		},
	}
	bashTool := &mockToolImpl{name: "bash", result: map[string]any{"stdout": "hi"}}
	reg := newTestRegistry(bashTool)
	exec := NewLLMAgentExecutor(p, reg, WithAgentID("coding-agent"))

	req := &AgentExecutionRequest{
		Task: &types.AgentTask{TaskID: "t2", ContractID: "default", Input: map[string]any{}},
	}
	result, err := exec.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.AgentID != "coding-agent" {
		t.Errorf("expected AgentID=coding-agent, got %s", result.AgentID)
	}
	// Should have tool traces
	traces, ok := result.Output["_tool_traces"].([]ToolTrace)
	if !ok || len(traces) == 0 {
		t.Error("expected tool traces in output")
	}
}

func TestLLMAgentExecutor_CircuitBreaker(t *testing.T) {
	// Provider always requests a non-existent tool
	p := &mockLLMProvider{
		responses: make([]*provider.ModelResponse, 10),
	}
	for i := range p.responses {
		p.responses[i] = &provider.ModelResponse{
			ToolCalls: []types.ToolCall{{ID: fmt.Sprintf("c%d", i), Name: "nonexistent", Input: map[string]any{}}},
		}
	}
	reg := newTestRegistry()
	exec := NewLLMAgentExecutor(p, reg, WithMaxErrors(3))

	req := &AgentExecutionRequest{
		Task: &types.AgentTask{TaskID: "t3", ContractID: "default", Input: map[string]any{}},
	}
	result, err := exec.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected circuit breaker error")
	}
	if p.callCount > 3 {
		t.Errorf("circuit breaker should have stopped after 3 errors, got %d calls", p.callCount)
	}
}

func TestLLMAgentExecutor_IterationBudget(t *testing.T) {
	// Provider always requests tools, never completes
	p := &mockLLMProvider{
		responses: make([]*provider.ModelResponse, 50),
	}
	callNum := 0
	bashTool := &mockToolImpl{name: "bash", execFn: func(_ map[string]any) (map[string]any, error) {
		callNum++
		return map[string]any{"stdout": fmt.Sprintf("ok-%d", callNum)}, nil
	}}
	for i := range p.responses {
		p.responses[i] = &provider.ModelResponse{
			ToolCalls: []types.ToolCall{{ID: fmt.Sprintf("c%d", i), Name: "bash", Input: map[string]any{}}},
		}
	}
	reg := newTestRegistry(bashTool)
	exec := NewLLMAgentExecutor(p, reg, WithMaxIterations(5))

	req := &AgentExecutionRequest{
		Task: &types.AgentTask{TaskID: "t4", ContractID: "default", Input: map[string]any{}},
	}
	result, _ := exec.Execute(context.Background(), req)
	if result.Error == "" {
		t.Fatal("expected iteration budget error")
	}
	if p.callCount != 5 {
		t.Errorf("expected exactly 5 iterations, got %d", p.callCount)
	}
}

func TestLLMAgentExecutor_CustomTermination(t *testing.T) {
	// LLM returns no tool calls but termination says Continue first time, Complete second
	callCount := 0
	p := &mockLLMProvider{
		responses: []*provider.ModelResponse{
			{Output: map[string]any{"status": "thinking"}},
			{Output: map[string]any{"status": "done"}},
		},
	}
	reg := newTestRegistry()
	exec := NewLLMAgentExecutor(p, reg, WithTerminationFn(func(history []types.ModelMessage, last *provider.ModelResponse) TerminationDecision {
		callCount++
		if callCount == 1 {
			return Continue
		}
		return Complete
	}))

	req := &AgentExecutionRequest{
		Task: &types.AgentTask{TaskID: "t5", ContractID: "default", Input: map[string]any{}},
	}
	result, err := exec.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if p.callCount != 2 {
		t.Errorf("expected 2 LLM calls (continue then complete), got %d", p.callCount)
	}
}

func TestLLMAgentExecutor_ToolError(t *testing.T) {
	p := &mockLLMProvider{
		responses: []*provider.ModelResponse{
			{ToolCalls: []types.ToolCall{{ID: "c1", Name: "bash", Input: map[string]any{}}}},
			{Output: map[string]any{"result": "recovered"}},
		},
	}
	bashTool := &mockToolImpl{name: "bash", err: fmt.Errorf("command failed")}
	reg := newTestRegistry(bashTool)
	exec := NewLLMAgentExecutor(p, reg)

	req := &AgentExecutionRequest{
		Task: &types.AgentTask{TaskID: "t6", ContractID: "default", Input: map[string]any{}},
	}
	result, err := exec.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Agent should recover after tool error (LLM sees error, responds with output)
	if result.Output["result"] != "recovered" {
		t.Errorf("expected recovery, got %v", result.Output)
	}
}

func TestLLMAgentExecutor_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	p := &mockLLMProvider{
		responses: []*provider.ModelResponse{{Output: map[string]any{}}},
	}
	reg := newTestRegistry()
	exec := NewLLMAgentExecutor(p, reg)

	req := &AgentExecutionRequest{
		Task: &types.AgentTask{TaskID: "t7", ContractID: "default", Input: map[string]any{}},
	}
	result, err := exec.Execute(ctx, req)
	if err == nil {
		t.Fatal("expected context error")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
}
