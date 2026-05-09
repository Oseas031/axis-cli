package tool

import (
	"context"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

// mockTool is a simple tool implementation for testing the registry.
type mockTool struct {
	name string
}

func (m *mockTool) Name() string { return m.name }

func (m *mockTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        m.name,
		Description: "mock tool for testing",
	}
}

func (m *mockTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) {
	return map[string]any{"result": "ok"}, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "test-tool"}

	err := r.Register(tool)
	if err != nil {
		t.Fatalf("Register should succeed: %v", err)
	}

	got, ok := r.Get("test-tool")
	if !ok {
		t.Fatal("Get should return ok=true for registered tool")
	}
	if got.Name() != "test-tool" {
		t.Errorf("Expected name test-tool, got %s", got.Name())
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "dup"}

	err := r.Register(tool)
	if err != nil {
		t.Fatalf("First register should succeed: %v", err)
	}

	err = r.Register(&mockTool{name: "dup"})
	if err == nil {
		t.Error("Duplicate register should return error")
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get for unknown tool should return ok=false")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "a"})
	r.Register(&mockTool{name: "b"})

	defs := r.List()
	if len(defs) != 2 {
		t.Fatalf("Expected 2 tool definitions, got %d", len(defs))
	}

	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}
	if !names["a"] || !names["b"] {
		t.Error("List should contain both registered tool names")
	}
}

func TestRegistry_ListEmpty(t *testing.T) {
	r := NewRegistry()
	defs := r.List()
	if len(defs) != 0 {
		t.Errorf("Expected empty list, got %d items", len(defs))
	}
}

func TestRegistry_ImplementsInterface(t *testing.T) {
	var _ Tool = &mockTool{}
}
