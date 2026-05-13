package horizon

import (
	"os"
	"path/filepath"
	"strings"
)

// BuildPrinciplesPromptSection returns the full body of all principles files
// for injection into the System Prompt. Returns empty string if no principles exist.
// This is Layer 1 injection: zero retrieval cost, always present.
func (s *Store) BuildPrinciplesPromptSection() string {
	dir := filepath.Join(s.root, DirPrinciples)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		body := extractBody(string(data))
		if body == "" {
			continue
		}
		if count == 0 {
			sb.WriteString("\nDerived principles (from experience):\n")
		}
		sb.WriteString(body)
		sb.WriteString("\n")
		count++
	}
	return sb.String()
}

// extractBody returns the markdown body after frontmatter (--- ... ---).
func extractBody(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return strings.TrimSpace(content)
	}
	parts := strings.SplitN(content[4:], "---\n", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(content)
}
