// Package agent provides self-context management for agent autonomy.
package agent

import (
	"context"

	"github.com/axis-cli/axis/internal/types"
)

// TaskStateReader defines the interface for reading task state.
// This decouples agent from kernel/sharedlayer implementation.
type TaskStateReader interface {
	// LoadState returns the task state for the given task ID.
	LoadState(taskID string) (types.TaskState, error)

	// SaveState persists the task state.
	SaveState(taskID string, state types.TaskState) error
}

// ModelProvider defines the interface for LLM model execution.
// This decouples agent from model/provider implementation.
type ModelProvider interface {
	// Execute runs a model request and returns the response.
	Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error)
}

// ModelRequest represents a request to the LLM model.
type ModelRequest struct {
	ContractID   string
	Input        map[string]any
	Tools        []ToolDefinition
	SystemPrompt string
	History      []types.ModelMessage
}

// ModelResponse represents the LLM model response.
type ModelResponse struct {
	Output    map[string]any
	ToolCalls []ToolCall
	Error     string
}

// ToolDefinition defines a tool available to the agent.
type ToolDefinition struct {
	Name        string
	Description string
	Schema      map[string]any
}

// ToolCall represents a tool invocation from the LLM.
type ToolCall struct {
	ID    string
	Name  string
	Input map[string]any
}

// ToolRegistry defines the interface for tool management.
type ToolRegistry interface {
	// List returns all available tools.
	List() []ToolDefinition

	// Execute runs a tool with the given input.
	Execute(ctx context.Context, name string, input map[string]any) (map[string]any, error)
}

// TaskState is a convenience alias for types.TaskState.
type TaskState = types.TaskState
