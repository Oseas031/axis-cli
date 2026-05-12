package skills

import (
	"context"
	"fmt"
	"strings"
)

// MaxLayer1Skills is the maximum number of skills to include in the system prompt.
const MaxLayer1Skills = 20

// BuildSkillsPromptSection returns the Layer 1 skills section for system prompt injection.
// Returns empty string if no skills are available.
func (l *Loader) BuildSkillsPromptSection(ctx context.Context) string {
	metas, err := l.Discover(ctx)
	if err != nil || len(metas) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\nSkills available (use load_skill tool to load detailed instructions):\n")
	limit := len(metas)
	if limit > MaxLayer1Skills {
		limit = MaxLayer1Skills
	}
	for i := 0; i < limit; i++ {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", metas[i].Name, metas[i].Description))
	}
	return sb.String()
}
