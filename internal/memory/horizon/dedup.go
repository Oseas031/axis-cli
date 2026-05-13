package horizon

import (
	"os"
	"path/filepath"
	"strings"
)

// DedupResult holds the outcome of a deduplication pass.
type DedupResult struct {
	Duplicates int `json:"duplicates"`
	Kept       int `json:"kept"`
}

// Deduplicate scans the patterns directory, groups files by title prefix
// (first 40 chars, lowercased), keeps the newest in each group, and marks
// older duplicates as deprecated.
func Deduplicate(store *Store) (*DedupResult, error) {
	dir := filepath.Join(store.root, DirPatterns)
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return &DedupResult{}, nil
		}
		return nil, err
	}

	// Group entries by title prefix
	type fileEntry struct {
		path  string
		entry Entry
	}
	groups := make(map[string][]fileEntry)
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		entry := parseMarkdown(content, f.Name(), dir)
		prefix := strings.ToLower(truncate(entry.Title, 40))
		groups[prefix] = append(groups[prefix], fileEntry{path: path, entry: entry})
	}

	result := &DedupResult{}
	for _, group := range groups {
		if len(group) <= 1 {
			continue
		}
		// Find newest
		newestIdx := 0
		for i := 1; i < len(group); i++ {
			if group[i].entry.CreatedAt.After(group[newestIdx].entry.CreatedAt) {
				newestIdx = i
			}
		}
		result.Kept++
		// Mark others as deprecated
		for i, fe := range group {
			if i == newestIdx {
				continue
			}
			if err := markDeprecated(fe.path); err != nil {
				continue
			}
			result.Duplicates++
		}
	}
	return result, nil
}

func markDeprecated(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, "deprecated: true") {
		return nil // already marked
	}
	// Insert after first "---\n"
	if strings.HasPrefix(content, "---\n") {
		content = "---\ndeprecated: true\n" + content[4:]
	}
	return os.WriteFile(path, []byte(content), 0600)
}
