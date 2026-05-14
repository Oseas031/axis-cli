package provider

import (
	"context"
	"log"
	"sync"

	"github.com/axis-cli/axis/internal/types"
)

// MockModelProvider returns an echo of the input for testing.
// When tools are configured, it simulates tool calling behavior.
type MockModelProvider struct {
	warnOnce       sync.Once
	warningHandler func(string)
}

func NewMockModelProvider() *MockModelProvider {
	return &MockModelProvider{}
}

// SetWarningHandler allows overriding the default log warning for testability.
func (m *MockModelProvider) SetWarningHandler(fn func(string)) {
	m.warningHandler = fn
}

func (m *MockModelProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	m.warnOnce.Do(func() {
		msg := "[WARN] MockModelProvider: no real provider configured. Response will be empty."
		if m.warningHandler != nil {
			m.warningHandler(msg)
		} else {
			log.Println(msg)
		}
	})

	// If history has a tool result as the last entry, incorporate it.
	// This check must come first so the multi-turn loop can terminate.
	if len(req.History) > 0 {
		last := req.History[len(req.History)-1]
		if last.Role == "tool" {
			output := map[string]any{
				"status":      "completed",
				"message":     "mock model executed after tool call",
				"provider":    "mock",
				"tool_result": last.Content,
			}
			output["contract_id"] = req.ContractID
			return &ModelResponse{Output: output, InputTokens: 100, OutputTokens: 50}, nil
		}
	}

	// Tool-aware mode: if tools are present and input has "tool" key, return a ToolCall.
	if len(req.Tools) > 0 {
		if toolName, ok := req.Input["tool"].(string); ok && toolName != "" {
			toolInput, _ := req.Input["input"].(map[string]any)
			if toolInput == nil {
				toolInput = make(map[string]any)
			}
			return &ModelResponse{
				ToolCalls: []types.ToolCall{
					{ID: "call-1", Name: toolName, Input: toolInput},
				},
				InputTokens:  100,
				OutputTokens: 50,
			}, nil
		}
	}

	// Default echo behavior.
	output := make(map[string]any)
	for k, v := range req.Input {
		output[k] = v
	}
	output["contract_id"] = req.ContractID
	output["status"] = "completed"
	output["provider"] = "mock"
	return &ModelResponse{Output: output, InputTokens: 100, OutputTokens: 50}, nil
}
