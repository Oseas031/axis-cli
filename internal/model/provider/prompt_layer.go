package provider

import "strings"

// PromptLayer represents one layer in the prompt assembly chain.
type PromptLayer struct {
	Name     string // "default", "project", "task", "skills", "append"
	Content  string
	Priority int // higher = applied later (overrides earlier)
}

// PromptAssembler builds an effective system prompt from layers.
type PromptAssembler struct {
	layers []PromptLayer
}

// NewPromptAssembler creates an assembler with default layer ordering.
func NewPromptAssembler() *PromptAssembler {
	return &PromptAssembler{}
}

// AddLayer adds a prompt layer. Layers are applied in priority order.
func (pa *PromptAssembler) AddLayer(name, content string, priority int) {
	if content == "" {
		return
	}
	pa.layers = append(pa.layers, PromptLayer{Name: name, Content: content, Priority: priority})
}

// Build assembles the effective system prompt.
// Layers are sorted by priority (ascending). Each layer's content is appended.
// Project layer constraints override default layer if both set the same directive.
// Append layers are always added at the end, never replace.
func (pa *PromptAssembler) Build() string {
	if len(pa.layers) == 0 {
		return ""
	}
	// Sort by priority
	sorted := make([]PromptLayer, len(pa.layers))
	copy(sorted, pa.layers)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].Priority < sorted[j-1].Priority; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	var parts []string
	for _, l := range sorted {
		parts = append(parts, l.Content)
	}
	return strings.Join(parts, "\n\n")
}

// Standard priority constants.
const (
	PriorityDefault = 0
	PriorityProject = 10
	PriorityTask    = 20
	PrioritySkills  = 30
	PriorityAppend  = 100
)
