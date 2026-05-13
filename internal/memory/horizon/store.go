// Package horizon implements the long-horizon memory layer for Axis.
// It manages patterns, principles, and narrative memories stored as
// markdown files in .axis/memory/{patterns,principles,narrative}/.
package horizon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Subdirectories within .axis/memory/ for long-horizon storage.
const (
	DirPatterns   = "patterns"
	DirPrinciples = "principles"
	DirNarrative  = "narrative"
	DirMilestones = "narrative/milestones"
	DirDreams     = "dreams"
)

// Category represents a memory storage category.
type Category string

const (
	CategoryPatterns   Category = "patterns"
	CategoryPrinciples Category = "principles"
	CategoryNarrative  Category = "narrative"
)

// Entry is a single memory entry with frontmatter metadata.
type Entry struct {
	ID        string    `json:"id"`
	Category  Category  `json:"category"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Body      string    `json:"body"`
}

// Store manages long-horizon memory files on disk.
type Store struct {
	root string // .axis/memory/ directory
}

// NewStore creates a Store rooted at the given memory directory.
func NewStore(memoryDir string) *Store {
	return &Store{root: memoryDir}
}

// Init creates the directory structure.
func (s *Store) Init() error {
	dirs := []string{DirPatterns, DirPrinciples, DirNarrative, DirMilestones, DirDreams}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(s.root, d), 0755); err != nil {
			return fmt.Errorf("horizon: init %s: %w", d, err)
		}
	}
	return nil
}

// Store writes a memory entry as a markdown file.
func (s *Store) Store(entry Entry) error {
	dir := filepath.Join(s.root, string(entry.Category))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filename := sanitizeFilename(entry.ID) + ".md"
	path := filepath.Join(dir, filename)
	content := renderMarkdown(entry)
	return os.WriteFile(path, []byte(content), 0600)
}

// Recall searches memory files by keyword across all categories (or a specific one).
func (s *Store) Recall(keyword string, category Category) ([]Entry, error) {
	var dirs []string
	if category != "" {
		dirs = []string{filepath.Join(s.root, string(category))}
	} else {
		dirs = []string{
			filepath.Join(s.root, DirPatterns),
			filepath.Join(s.root, DirPrinciples),
			filepath.Join(s.root, DirNarrative),
		}
	}

	var results []Entry
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			content := string(data)
			if keyword == "" || strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
				entry := parseMarkdown(content, e.Name(), dir)
				results = append(results, entry)
			}
		}
	}
	return results, nil
}

func renderMarkdown(e Entry) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", e.ID))
	sb.WriteString(fmt.Sprintf("title: %s\n", e.Title))
	sb.WriteString(fmt.Sprintf("category: %s\n", e.Category))
	if len(e.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: %s\n", strings.Join(e.Tags, ", ")))
	}
	sb.WriteString(fmt.Sprintf("created: %s\n", e.CreatedAt.Format(time.RFC3339)))
	sb.WriteString("---\n\n")
	sb.WriteString(e.Body)
	sb.WriteString("\n")
	return sb.String()
}

func parseMarkdown(content, filename, dir string) Entry {
	entry := Entry{}
	entry.ID = strings.TrimSuffix(filename, ".md")
	// Derive category from directory name
	entry.Category = Category(filepath.Base(dir))

	// Parse frontmatter
	if strings.HasPrefix(content, "---\n") {
		parts := strings.SplitN(content[4:], "---\n", 2)
		if len(parts) == 2 {
			for _, line := range strings.Split(parts[0], "\n") {
				kv := strings.SplitN(line, ": ", 2)
				if len(kv) != 2 {
					continue
				}
				switch kv[0] {
				case "title":
					entry.Title = kv[1]
				case "tags":
					entry.Tags = strings.Split(kv[1], ", ")
				case "created":
					t, _ := time.Parse(time.RFC3339, kv[1])
					entry.CreatedAt = t
				}
			}
			entry.Body = strings.TrimSpace(parts[1])
		}
	}
	return entry
}

func sanitizeFilename(id string) string {
	r := strings.NewReplacer("/", "-", "\\", "-", " ", "-", ":", "-")
	return r.Replace(id)
}
