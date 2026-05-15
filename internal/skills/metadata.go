package skills

import "strings"

// SkillMetadata holds composable metadata parsed from SKILL.md frontmatter.
type SkillMetadata struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags,omitempty"`
	DependsOn     []string `json:"depends_on,omitempty"`
	ConflictsWith []string `json:"conflicts_with,omitempty"`
	Version       string   `json:"version,omitempty"`
	Source        string   `json:"source,omitempty"`
	SourceVersion string   `json:"source_version,omitempty"`
}

// ParseSkillMetadata splits YAML frontmatter from body and parses composable metadata.
// Returns nil metadata and full content as body if no frontmatter is present.
func ParseSkillMetadata(content string) (*SkillMetadata, string) {
	const delim = "---"
	if !strings.HasPrefix(content, delim) {
		return nil, content
	}
	rest := content[len(delim):]
	rest = strings.TrimPrefix(rest, "\r\n")
	rest = strings.TrimPrefix(rest, "\n")
	idx := strings.Index(rest, "\n"+delim)
	if idx < 0 {
		return nil, content
	}
	yamlBlock := rest[:idx]
	body := rest[idx+1+len(delim):]
	body = strings.TrimPrefix(body, "\r\n")
	body = strings.TrimPrefix(body, "\n")

	meta := &SkillMetadata{}
	for _, line := range strings.Split(yamlBlock, "\n") {
		line = strings.TrimRight(line, "\r")
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		val = stripQuotes(val)
		switch key {
		case "name":
			meta.Name = val
		case "description":
			meta.Description = val
		case "tags":
			meta.Tags = splitList(val)
		case "depends_on":
			meta.DependsOn = splitList(val)
		case "conflicts_with":
			meta.ConflictsWith = splitList(val)
		case "version":
			meta.Version = val
		case "source":
			meta.Source = val
		case "source_version":
			meta.SourceVersion = val
		}
	}
	return meta, body
}

func splitList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// Handle YAML bracket syntax: [a, b, c]
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		raw = raw[1 : len(raw)-1]
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
