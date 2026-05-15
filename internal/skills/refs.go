package skills

import (
	"regexp"
	"strings"
)

// skillRefRe matches /skill-name but not URLs (requires non-letter/non-slash before /).
var skillRefRe = regexp.MustCompile(`(?:^|[^a-zA-Z0-9:/])(/([a-z][a-z0-9-]*[a-z0-9]))`)

// SkillRef represents a reference from one skill to another.
type SkillRef struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

// ExtractRefs scans skill content for /skill-name references.
// Only returns refs that match valid skill name format, deduplicated by name.
func ExtractRefs(content string) []SkillRef {
	seen := make(map[string]bool)
	var refs []SkillRef
	for i, line := range strings.Split(content, "\n") {
		for _, match := range skillRefRe.FindAllStringSubmatch(line, -1) {
			name := match[2]
			if seen[name] || ValidateSkillName(name) != nil {
				continue
			}
			seen[name] = true
			refs = append(refs, SkillRef{Name: name, Line: i + 1})
		}
	}
	return refs
}
