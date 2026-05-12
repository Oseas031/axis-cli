package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseFrontmatter extracts YAML frontmatter from markdown content.
// Returns (meta map, markdown body, error).
func parseFrontmatter(content string) (map[string]string, string, error) {
	const delim = "---"
	// Must start with ---
	if !strings.HasPrefix(content, delim) {
		return nil, content, fmt.Errorf("missing opening frontmatter delimiter")
	}
	// Find closing ---
	rest := content[len(delim):]
	rest = strings.TrimPrefix(rest, "\r\n")
	rest = strings.TrimPrefix(rest, "\n")
	idx := strings.Index(rest, "\n"+delim)
	if idx < 0 {
		return nil, content, fmt.Errorf("missing closing frontmatter delimiter")
	}
	yamlBlock := rest[:idx]
	body := rest[idx+1+len(delim):]
	body = strings.TrimPrefix(body, "\r\n")
	body = strings.TrimPrefix(body, "\n")

	meta := make(map[string]string)
	for _, line := range strings.Split(yamlBlock, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		meta[key] = val
	}
	return meta, body, nil
}

// parseTags splits a comma-separated tag string into a slice.
func parseTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// Discover scans .axis/skills/ and returns all valid skill metadata.
func (l *Loader) Discover(ctx context.Context) ([]SkillMeta, error) {
	l.mu.RLock()
	if len(l.index) > 0 {
		metas := make([]SkillMeta, 0, len(l.index))
		for _, m := range l.index {
			metas = append(metas, m)
		}
		l.mu.RUnlock()
		return metas, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	entries, err := os.ReadDir(l.skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillFile := filepath.Join(l.skillsDir, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillFile)
		if err != nil {
			continue // skip directories without SKILL.md
		}
		fm, _, err := parseFrontmatter(string(data))
		if err != nil {
			continue
		}
		meta := SkillMeta{
			Name:        fm["name"],
			Description: fm["description"],
			Tags:        parseTags(fm["tags"]),
			Version:     fm["version"],
			Author:      fm["author"],
		}
		if meta.Validate() != nil {
			continue
		}
		l.index[meta.Name] = meta
	}

	metas := make([]SkillMeta, 0, len(l.index))
	for _, m := range l.index {
		metas = append(metas, m)
	}
	return metas, nil
}
