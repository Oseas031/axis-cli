package capability

import "fmt"

// CapabilityType classifies what kind of capability this is.
type CapabilityType string

const (
	CapTool   CapabilityType = "tool"
	CapSkill  CapabilityType = "skill"
	CapMemory CapabilityType = "memory"
)

// Capability is the unified interface for all registrable capabilities.
type Capability interface {
	CapName() string
	CapType() CapabilityType
	CapDescription() string
}

// CapabilityRegistry holds all registered capabilities.
type CapabilityRegistry struct {
	capabilities map[string]Capability
}

func NewCapabilityRegistry() *CapabilityRegistry {
	return &CapabilityRegistry{capabilities: make(map[string]Capability)}
}

func (r *CapabilityRegistry) Register(c Capability) error {
	name := c.CapName()
	if _, exists := r.capabilities[name]; exists {
		return fmt.Errorf("capability already registered: %s", name)
	}
	r.capabilities[name] = c
	return nil
}

func (r *CapabilityRegistry) Get(name string) (Capability, bool) {
	c, ok := r.capabilities[name]
	return c, ok
}

func (r *CapabilityRegistry) ListByType(t CapabilityType) []Capability {
	var result []Capability
	for _, c := range r.capabilities {
		if c.CapType() == t {
			result = append(result, c)
		}
	}
	return result
}
