// Package tool provides a pluggable tool system for model providers.
package tool

import (
	"context"
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/types"
)

// Tool defines a tool that can be invoked by a model provider.
type Tool interface {
	Name() string
	Schema() types.ToolDefinition
	Execute(ctx context.Context, input map[string]any) (map[string]any, error)
}

// Registry manages a set of registered tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[t.Name()]; exists {
		return fmt.Errorf("tool %s already registered", t.Name())
	}
	r.tools[t.Name()] = t
	return nil
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List returns the schema definitions for all registered tools.
func (r *Registry) List() []types.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]types.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.Schema())
	}
	return defs
}
