package provider

import (
	"context"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestMockModelProvider_ToolCallRequest(t *testing.T) {
	m := NewMockModelProvider()
	req := &ModelRequest{
		ContractID: "tool-test",
		Input: map[string]any{
			"tool":  "bash",
			"input": map[string]any{"command": "echo hello"},
		},
		Tools: []types.ToolDefinition{
			{Name: "bash", Description: "Execute bash commands"},
		},
	}

	resp, err := m.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ID != "call-1" {
		t.Errorf("Expected call ID 'call-1', got %q", resp.ToolCalls[0].ID)
	}
	if resp.ToolCalls[0].Name != "bash" {
		t.Errorf("Expected tool name 'bash', got %q", resp.ToolCalls[0].Name)
	}
	if resp.Output != nil {
		t.Error("Expected Output to be nil when tool calls are returned")
	}
}

func TestMockModelProvider_ToolCallWithEmptyInput(t *testing.T) {
	m := NewMockModelProvider()
	req := &ModelRequest{
		ContractID: "tool-test",
		Input: map[string]any{
			"tool": "bash",
		},
		Tools: []types.ToolDefinition{
			{Name: "bash", Description: "Execute bash commands"},
		},
	}

	resp, err := m.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "bash" {
		t.Errorf("Expected tool name 'bash', got %q", resp.ToolCalls[0].Name)
	}
	// Should have empty input map instead of nil
	if resp.ToolCalls[0].Input == nil {
		t.Error("Expected non-nil input map")
	}
}

func TestMockModelProvider_ToolCallWithMissingToolKey(t *testing.T) {
	m := NewMockModelProvider()
	req := &ModelRequest{
		ContractID: "tool-test",
		Input:      map[string]any{"msg": "hello"},
		Tools: []types.ToolDefinition{
			{Name: "bash", Description: "Execute bash commands"},
		},
	}

	resp, err := m.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}
	// Should fall back to echo mode since input has no "tool" key.
	if resp.Output == nil {
		t.Fatal("Expected non-nil Output")
	}
	if resp.Output["msg"] != "hello" {
		t.Errorf("Expected echoed msg, got %v", resp.Output["msg"])
	}
	if resp.Output["provider"] != "mock" {
		t.Errorf("Expected provider=mock, got %v", resp.Output["provider"])
	}
}

func TestMockModelProvider_HistoryWithToolResult(t *testing.T) {
	m := NewMockModelProvider()
	req := &ModelRequest{
		ContractID: "tool-result-test",
		Input:      map[string]any{"msg": "after tool"},
		History: []types.ModelMessage{
			{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "call-1", Name: "bash", Input: map[string]any{"command": "echo hi"}}}},
			{Role: "tool", ToolCallID: "call-1", Content: `{"stdout": "hi\n", "exit_code": 0}`},
		},
	}

	resp, err := m.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}
	if resp.Output == nil {
		t.Fatal("Expected non-nil Output")
	}
	if resp.Output["status"] != "completed" {
		t.Errorf("Expected status=completed, got %v", resp.Output["status"])
	}
	if resp.Output["message"] != "mock model executed after tool call" {
		t.Errorf("Expected tool result message, got %v", resp.Output["message"])
	}
	if resp.Output["tool_result"] != `{"stdout": "hi\n", "exit_code": 0}` {
		t.Errorf("Expected tool_result to be preserved, got %v", resp.Output["tool_result"])
	}
}

func TestMockModelProvider_HistoryLastEntryNotTool(t *testing.T) {
	m := NewMockModelProvider()
	req := &ModelRequest{
		ContractID: "some-test",
		Input:      map[string]any{"msg": "hello"},
		History: []types.ModelMessage{
			{Role: "assistant", Content: "previous response"},
		},
	}

	resp, err := m.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}
	if resp.Output == nil {
		t.Fatal("Expected non-nil Output")
	}
	// Should fall through to default echo behavior.
	if resp.Output["msg"] != "hello" {
		t.Errorf("Expected echoed msg, got %v", resp.Output["msg"])
	}
}
