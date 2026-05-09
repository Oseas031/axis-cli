// Package tool provides a pluggable tool system for model providers.
package tool

import (
	"context"
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/types"
)

// ToolPermissionScope represents a permission category for a tool.
type ToolPermissionScope string

const (
	ScopeFilesystemRead  ToolPermissionScope = "filesystem:read"
	ScopeFilesystemWrite ToolPermissionScope = "filesystem:write"
	ScopeNetwork         ToolPermissionScope = "network"
	ScopeSubprocess      ToolPermissionScope = "subprocess"
)

// Tool defines a tool that can be invoked by a model provider.
type Tool interface {
	Name() string
	Schema() types.ToolDefinition
	Execute(ctx context.Context, input map[string]any) (map[string]any, error)
}

// Registry manages a set of registered tools.
type Registry struct {
	mu     sync.RWMutex
	tools  map[string]Tool
	scopes map[string][]string // tool name -> scopes
}

// NewRegistry creates a new empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool), scopes: make(map[string][]string)}
}

// Register adds a tool to the registry with the given permission scopes.
func (r *Registry) Register(t Tool, scopes []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[t.Name()]; exists {
		return fmt.Errorf("tool %s already registered", t.Name())
	}
	r.tools[t.Name()] = t
	r.scopes[t.Name()] = scopes
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

// GetScopes returns the permission scopes for a registered tool.
func (r *Registry) GetScopes(toolName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.scopes[toolName]
}
