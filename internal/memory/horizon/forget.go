package horizon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ForgetResult holds the outcome of a forget operation.
type ForgetResult struct {
	Archived int      `json:"archived"`
	Deleted  int      `json:"deleted"`
	Skipped  int      `json:"skipped"`
	Details  []string `json:"details,omitempty"`
}

// Forget scans narrative/ entries and archives or deletes old ones.
// Files older than 7 days have their body replaced with [archived].
// Files older than 30 days are deleted entirely.
// patterns/ and principles/ are never touched.
func Forget(store *Store, dryRun bool) (*ForgetResult, error) {
	narrativeDir := filepath.Join(store.root, DirNarrative)
	result := &ForgetResult{}
	now := time.Now()

	entries, err := os.ReadDir(narrativeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("forget: read narrative dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(narrativeDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		entry := parseMarkdown(content, e.Name(), narrativeDir)
		if entry.CreatedAt.IsZero() {
			result.Skipped++
			result.Details = append(result.Details, fmt.Sprintf("skip (no created): %s", e.Name()))
			continue
		}

		age := now.Sub(entry.CreatedAt)
		switch {
		case age > 30*24*time.Hour:
			result.Deleted++
			result.Details = append(result.Details, fmt.Sprintf("delete (%.0fd old): %s", age.Hours()/24, e.Name()))
			if !dryRun {
				os.Remove(path)
			}
		case age > 7*24*time.Hour:
			result.Archived++
			result.Details = append(result.Details, fmt.Sprintf("archive (%.0fd old): %s", age.Hours()/24, e.Name()))
			if !dryRun {
				archived := replacebody(content, "[archived]")
				_ = os.WriteFile(path, []byte(archived), 0600)
			}
		default:
			result.Skipped++
			result.Details = append(result.Details, fmt.Sprintf("skip (%.0fd old): %s", age.Hours()/24, e.Name()))
		}
	}
	return result, nil
}

// replacebody keeps frontmatter intact and replaces the body.
func replacebody(content, newBody string) string {
	if !strings.HasPrefix(content, "---\n") {
		return newBody + "\n"
	}
	parts := strings.SplitN(content[4:], "---\n", 2)
	if len(parts) < 2 {
		return content
	}
	return "---\n" + parts[0] + "---\n\n" + newBody + "\n"
}
