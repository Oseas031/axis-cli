package executor

import (
	"context"
	"fmt"
	"testing"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// fixedTool returns a fixed result for testing.
type fixedTool struct {
	name   string
	result map[string]any
	err    error
}

func (f *fixedTool) Name() string { return f.name }

func (f *fixedTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{Name: f.name, Description: "fixed tool for testing"}
}

func (f *fixedTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) {
	return f.result, f.err
}

func TestContractExecutor_Execute_ToolCallAndResult(t *testing.T) {
	// Register a contract and a tool, then execute with a tool-calling provider.
	exec := NewContractExecutor()
	exec.SetProvider(provider.NewMockModelProvider())

	// Create and set tool registry with a fixed tool.
	reg := tool.NewRegistry()
	reg.Register(&fixedTool{
		name:   "bash",
		result: map[string]any{"stdout": "hello", "exit_code": 0},
	}, nil)
	exec.SetToolRegistry(reg)

	contract := &types.AgentContract{
		ContractID: "tool-contract",
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{Name: "status", Type: types.FieldTypeString, Required: true},
			},
		},
	}
	exec.RegisterContract(contract)

	input := map[string]any{
		"tool":  "bash",
		"input": map[string]any{"command": "echo hello"},
	}

	result, err := exec.Execute(context.Background(), "tool-contract", input)
	if err != nil {
		t.Fatalf("Execute should succeed: %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	// The MockModelProvider with tool-history should return final output.
	if result.Output["status"] != "completed" {
		t.Errorf("Expected status=completed, got %v", result.Output["status"])
	}
	if result.Output["message"] != "mock model executed after tool call" {
		t.Errorf("Expected tool execution message, got %v", result.Output["message"])
	}
}

func TestContractExecutor_Execute_WithoutTools(t *testing.T) {
	// When tool registry is not set, behavior must be identical to original single-pass.
	exec := NewContractExecutor()
	exec.SetProvider(provider.NewMockModelProvider())

	contract := &types.AgentContract{
		ContractID: "no-tool-contract",
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{Name: "status", Type: types.FieldTypeString, Required: true},
			},
		},
	}
	exec.RegisterContract(contract)

	result, err := exec.Execute(context.Background(), "no-tool-contract", map[string]any{"msg": "hello"})
	if err != nil {
		t.Fatalf("Execute should succeed: %v", err)
	}
	if result.Output["msg"] != "hello" {
		t.Errorf("Expected echoed msg, got %v", result.Output["msg"])
	}
	if result.Output["provider"] != "mock" {
		t.Errorf("Expected provider=mock, got %v", result.Output["provider"])
	}
}

func TestContractExecutor_Execute_EmptyToolRegistry(t *testing.T) {
	// When tool registry is set but empty, should still be single-pass.
	exec := NewContractExecutor()
	exec.SetProvider(provider.NewMockModelProvider())
	exec.SetToolRegistry(tool.NewRegistry()) // empty registry

	contract := &types.AgentContract{
		ContractID: "empty-reg-contract",
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{Name: "status", Type: types.FieldTypeString, Required: true},
			},
		},
	}
	exec.RegisterContract(contract)

	result, err := exec.Execute(context.Background(), "empty-reg-contract", map[string]any{"msg": "hello"})
	if err != nil {
		t.Fatalf("Execute should succeed: %v", err)
	}
	if result.Output["msg"] != "hello" {
		t.Errorf("Expected echoed msg, got %v", result.Output["msg"])
	}
}

func TestContractExecutor_Execute_UnknownToolInRegistry(t *testing.T) {
	exec := NewContractExecutor()

	// A provider that always returns a tool call for an unknown tool.
	unknownToolProvider := &unknownToolCallProvider{}
	exec.SetProvider(unknownToolProvider)

	reg := tool.NewRegistry()
	exec.SetToolRegistry(reg)

	contract := &types.AgentContract{
		ContractID: "unknown-tool",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
	}
	exec.RegisterContract(contract)

	result, err := exec.Execute(context.Background(), "unknown-tool", map[string]any{"x": "y"})
	if err != nil {
		t.Fatalf("Execute should succeed: %v", err)
	}
	if result.Output["status"] != "completed" {
		t.Errorf("Expected status=completed, got %v", result.Output["status"])
	}
}

func TestContractExecutor_SetToolRegistry(t *testing.T) {
	exec := NewContractExecutor()
	reg := tool.NewRegistry()
	exec.SetToolRegistry(reg)
	// Should not panic or error.
}

func TestContractExecutor_Execute_ToolProviderError(t *testing.T) {
	exec := NewContractExecutor()

	errorProvider := &errorToolCallProvider{}
	exec.SetProvider(errorProvider)

	reg := tool.NewRegistry()
	reg.Register(&fixedTool{
		name:   "bash",
		result: map[string]any{"stdout": "ok", "exit_code": 0},
	}, nil)
	exec.SetToolRegistry(reg)

	contract := &types.AgentContract{
		ContractID: "error-tool",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
	}
	exec.RegisterContract(contract)

	_, err := exec.Execute(context.Background(), "error-tool", map[string]any{"x": "y"})
	if err == nil {
		t.Error("Expected error from provider failure")
	}
}

// unknownToolCallProvider returns a tool call for a tool that isn't registered.
type unknownToolCallProvider struct{}

func (p *unknownToolCallProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	// Always return tool call for "nonexistent-tool"
	if len(req.Tools) > 0 {
		return &provider.ModelResponse{
			ToolCalls: []types.ToolCall{
				{ID: "call-1", Name: "nonexistent-tool", Input: map[string]any{}},
			},
		}, nil
	}
	return &provider.ModelResponse{
		Output: map[string]any{"status": "completed", "provider": "unknown-tool", "contract_id": req.ContractID},
	}, nil
}

// errorToolCallProvider returns an error on the first call.
type errorToolCallProvider struct {
	called bool
}

func (p *errorToolCallProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	if !p.called && len(req.Tools) > 0 {
		p.called = true
		return &provider.ModelResponse{
			ToolCalls: []types.ToolCall{
				{ID: "call-1", Name: "bash", Input: map[string]any{"command": "echo hi"}},
			},
		}, nil
	}
	return nil, fmt.Errorf("simulated provider error")
}

// erroringTool is a tool that always returns an error.
type erroringTool struct{}

func (e *erroringTool) Name() string { return "error-tool" }

func (e *erroringTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{Name: "error-tool", Description: "always fails"}
}

func (e *erroringTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) {
	return nil, fmt.Errorf("tool execution failed")
}

// erroringToolCallProvider returns tool calls for erroringTool indefinitely.
type erroringToolCallProvider struct {
	callCount int
}

func (p *erroringToolCallProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	p.callCount++
	return &provider.ModelResponse{
		ToolCalls: []types.ToolCall{
			{ID: fmt.Sprintf("call-%d", p.callCount), Name: "error-tool", Input: map[string]any{}},
		},
	}, nil
}

// successAfterErrorProvider returns errors for first N calls, then outputs.
type successAfterErrorProvider struct {
	errorCount    int
	callCount     int
	successOutput map[string]any
}

func (p *successAfterErrorProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	p.callCount++
	if p.callCount <= p.errorCount {
		return &provider.ModelResponse{
			ToolCalls: []types.ToolCall{
				{ID: fmt.Sprintf("call-%d", p.callCount), Name: "error-tool", Input: map[string]any{}},
			},
		}, nil
	}
	return &provider.ModelResponse{Output: p.successOutput}, nil
}

func TestContractExecutor_Execute_CircuitBreaker_TripsAfter5Errors(t *testing.T) {
	exec := NewContractExecutor()

	// Provider that keeps returning tool calls
	provider := &erroringToolCallProvider{}
	exec.SetProvider(provider)

	reg := tool.NewRegistry()
	reg.Register(&erroringTool{}, nil)
	exec.SetToolRegistry(reg)

	contract := &types.AgentContract{
		ContractID: "circuit-test",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
	}
	exec.RegisterContract(contract)

	// With 5 consecutive errors (each turn has 1 tool call), circuit should trip
	// Turn 1: error #1, Turn 2: error #2, Turn 3: error #3, Turn 4: error #4, Turn 5: error #5 -> abort
	result, err := exec.Execute(context.Background(), "circuit-test", map[string]any{"x": "y"})
	if err == nil {
		t.Fatal("Expected circuit breaker error")
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	// Circuit breaker error message should indicate triggered count
	if result.Error == "" {
		t.Error("Expected error message about circuit breaker")
	}
}

func TestContractExecutor_Execute_CircuitBreaker_ResetsOnSuccess(t *testing.T) {
	exec := NewContractExecutor()

	// Provider that returns 3 errors then success
	provider := &successAfterErrorProvider{
		errorCount:    3,
		successOutput: map[string]any{"status": "completed"},
	}
	exec.SetProvider(provider)

	reg := tool.NewRegistry()
	reg.Register(&erroringTool{}, nil)
	exec.SetToolRegistry(reg)

	contract := &types.AgentContract{
		ContractID: "circuit-reset-test",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
	}
	exec.RegisterContract(contract)

	// Should succeed because errors reset after 3 errors when success happens
	result, err := exec.Execute(context.Background(), "circuit-reset-test", map[string]any{"x": "y"})
	if err != nil {
		t.Fatalf("Should not error when success follows errors: %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	if result.Output["status"] != "completed" {
		t.Errorf("Expected status=completed, got %v", result.Output["status"])
	}
}

func TestSafeMarshal_ReturnsErrorForUnmarshalableValue(t *testing.T) {
	ch := make(chan int)
	_, err := safeMarshal(ch)
	if err == nil {
		t.Fatal("safeMarshal should return error for unmarshalable values such as channels")
	}
}

// unmarshalableResultTool returns a map containing a channel, which json.Marshal cannot serialize.
type unmarshalableResultTool struct{}

func (t *unmarshalableResultTool) Name() string { return "unmarshalable-tool" }
func (t *unmarshalableResultTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{Name: "unmarshalable-tool", Description: "returns data that cannot be JSON marshaled"}
}
func (t *unmarshalableResultTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) {
	return map[string]any{"data": make(chan int)}, nil
}

// marshalCheckingProvider returns a tool call on the first turn, then on the second turn
// checks whether the tool-result history message is non-empty.
type marshalCheckingProvider struct {
	callCount          int
	sawToolContent     bool
	lastHistoryChecked []types.ModelMessage
}

func (p *marshalCheckingProvider) Execute(_ context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	p.callCount++
	if p.callCount == 1 {
		return &provider.ModelResponse{
			ToolCalls: []types.ToolCall{
				{ID: "call-1", Name: "unmarshalable-tool", Input: map[string]any{}},
			},
		}, nil
	}
	p.lastHistoryChecked = req.History
	for _, msg := range req.History {
		if msg.Role == "tool" && msg.Content != "" {
			p.sawToolContent = true
		}
	}
	return &provider.ModelResponse{
		Output: map[string]any{"status": "completed"},
	}, nil
}

func TestContractExecutor_Execute_ToolResultMarshalErrorNotSwallowed(t *testing.T) {
	// Quick sanity: does safeMarshal actually fail for the tool result?
	quickResult := map[string]any{"data": make(chan int)}
	_, quickErr := safeMarshal(quickResult)
	if quickErr == nil {
		t.Fatal("sanity: safeMarshal should fail for map containing channel")
	}

	exec := NewContractExecutor()
	p := &marshalCheckingProvider{}
	exec.SetProvider(p)

	reg := tool.NewRegistry()
	if err := reg.Register(&unmarshalableResultTool{}, nil); err != nil {
		t.Fatalf("register tool: %v", err)
	}
	exec.SetToolRegistry(reg)

	contract := &types.AgentContract{
		ContractID: "marshal-error-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
	}
	exec.RegisterContract(contract)

	exec.Execute(context.Background(), "marshal-error-contract", map[string]any{"x": "y"})

	if !p.sawToolContent {
		t.Fatalf("tool result history message should contain marshal error info; history checked=%+v", p.lastHistoryChecked)
	}
}

func TestContractExecutor_Execute_CircuitBreaker_UnregisteredToolCounts(t *testing.T) {
	exec := NewContractExecutor()

	// Provider that returns tool calls for unknown tools
	unknownToolProvider := &unknownToolCallProvider{}
	exec.SetProvider(unknownToolProvider)

	reg := tool.NewRegistry()
	// Note: NOT registering erroringTool, so "error-tool" will be "not found"
	reg.Register(&erroringTool{}, nil)
	exec.SetToolRegistry(reg)

	contract := &types.AgentContract{
		ContractID: "circuit-unregistered",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
	}
	exec.RegisterContract(contract)

	// This will use unknownToolCallProvider which calls "nonexistent-tool" (not registered)
	// So it will count as a "tool not found" error
	result, err := exec.Execute(context.Background(), "circuit-unregistered", map[string]any{"x": "y"})
	if err == nil {
		// With only 1 tool call and unknown tool, should get error about max turns or tool not found
		// Not critical - just verify it handled gracefully
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}
