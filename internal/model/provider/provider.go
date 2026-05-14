// Package provider defines the model execution interface for Axis.
package provider

import (
	"context"

	"github.com/axis-cli/axis/internal/types"
)

// ModelRequest is the input to a model provider call.
type ModelRequest struct {
	ContractID   string
	Input        map[string]any
	Tools        []types.ToolDefinition // Available tools for this execution
	History      []types.ModelMessage   // Prior turns in multi-turn execution
	SystemPrompt string                 // Optional system prompt (includes skills metadata)
	Metadata     map[string]any         // Optional metadata for routing (e.g. request_type)
}

// ModelResponse is the output from a model provider call.
type ModelResponse struct {
	Output          map[string]any   // Final output (nil when tool calls are requested)
	ToolCalls       []types.ToolCall // Tool calls requested by the provider
	InputTokens     int              // Number of input tokens consumed
	OutputTokens    int              // Number of output tokens generated
	CostEstimateUSD float64          // Optional cost estimate based on model pricing
}

// ModelProvider defines the interface for model execution.
type ModelProvider interface {
	Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error)
}
