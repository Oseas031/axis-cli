// Package provider defines the model execution interface for Axis.
package provider

import "context"

// ModelRequest is the input to a model provider call.
type ModelRequest struct {
	ContractID string
	Input      map[string]any
}

// ModelResponse is the output from a model provider call.
type ModelResponse struct {
	Output map[string]any
}

// ModelProvider defines the interface for model execution.
type ModelProvider interface {
	Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error)
}
