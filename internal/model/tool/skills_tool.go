package tool

import (
	"context"

	"github.com/axis-cli/axis/internal/skills"
	"github.com/axis-cli/axis/internal/types"
)

// LoadSkillTool loads skill content by name.
type LoadSkillTool struct {
	loader *skills.Loader
}

// NewLoadSkillTool creates a new LoadSkillTool.
func NewLoadSkillTool(loader *skills.Loader) *LoadSkillTool {
	return &LoadSkillTool{loader: loader}
}

// Name returns the tool name.
func (t *LoadSkillTool) Name() string {
	return "load_skill"
}

// Schema returns the tool definition.
func (t *LoadSkillTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "load_skill",
		Description: "Load a skill by name to get detailed instructions. Use this when you need domain-specific knowledge.",
		Parameters: []types.FieldDef{
			{Name: "name", Type: types.FieldTypeString, Required: true, Description: "Skill name in kebab-case (e.g., 'pdf', 'code-review')"},
		},
	}
}

// Execute loads and returns the skill content.
func (t *LoadSkillTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	name, ok := input["name"].(string)
	if !ok || name == "" {
		return map[string]any{"error": "name is required"}, nil
	}
	skill, err := t.loader.Load(ctx, name)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}
	return map[string]any{
		"name":        skill.Meta.Name,
		"description": skill.Meta.Description,
		"content":     skill.Content,
	}, nil
}
