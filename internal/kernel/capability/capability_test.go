package capability

import "testing"

type mockCap struct {
	name    string
	capType CapabilityType
	desc    string
}

func (m *mockCap) CapName() string        { return m.name }
func (m *mockCap) CapType() CapabilityType { return m.capType }
func (m *mockCap) CapDescription() string  { return m.desc }

func TestRegisterAndGet(t *testing.T) {
	reg := NewCapabilityRegistry()
	c := &mockCap{name: "bash", capType: CapTool, desc: "execute bash"}

	if err := reg.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := reg.Get("bash")
	if !ok {
		t.Fatal("expected capability to be found")
	}
	if got.CapName() != "bash" {
		t.Fatalf("expected name 'bash', got %q", got.CapName())
	}
}

func TestRegisterDuplicate(t *testing.T) {
	reg := NewCapabilityRegistry()
	c := &mockCap{name: "bash", capType: CapTool, desc: "execute bash"}

	_ = reg.Register(c)
	if err := reg.Register(c); err == nil {
		t.Fatal("expected error on duplicate registration")
	}
}

func TestGetNotFound(t *testing.T) {
	reg := NewCapabilityRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Fatal("expected capability to not be found")
	}
}

func TestListByType(t *testing.T) {
	reg := NewCapabilityRegistry()
	_ = reg.Register(&mockCap{name: "bash", capType: CapTool, desc: "shell"})
	_ = reg.Register(&mockCap{name: "search", capType: CapSkill, desc: "search skill"})
	_ = reg.Register(&mockCap{name: "file-write", capType: CapTool, desc: "write files"})

	tools := reg.ListByType(CapTool)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	skills := reg.ListByType(CapSkill)
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	memories := reg.ListByType(CapMemory)
	if len(memories) != 0 {
		t.Fatalf("expected 0 memories, got %d", len(memories))
	}
}
