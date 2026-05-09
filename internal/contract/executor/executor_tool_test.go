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

	result, err := exec.Execute("tool-contract", input)
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

	result, err := exec.Execute("no-tool-contract", map[string]any{"msg": "hello"})
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

	result, err := exec.Execute("empty-reg-contract", map[string]any{"msg": "hello"})
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

	result, err := exec.Execute("unknown-tool", map[string]any{"x": "y"})
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

	_, err := exec.Execute("error-tool", map[string]any{"x": "y"})
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
