package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LoaderOption configures a Loader.
type LoaderOption func(*Loader)

// WithCacheTTL sets the cache time-to-live for discovery results.
func WithCacheTTL(d time.Duration) LoaderOption {
	return func(l *Loader) { l.cacheTTL = d }
}

const defaultCacheTTL = 30 * time.Second

// Loader manages skill discovery and loading.
type Loader struct {
	skillsDir      string
	index          map[string]SkillMeta
	mu             sync.RWMutex
	lastDiscoverAt time.Time
	cacheTTL       time.Duration
}

// NewLoader creates a new Loader rooted at skillsDir.
func NewLoader(skillsDir string, opts ...LoaderOption) *Loader {
	l := &Loader{
		skillsDir: skillsDir,
		index:     make(map[string]SkillMeta),
		cacheTTL:  defaultCacheTTL,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Invalidate clears the cached discovery results, forcing a re-scan on next Discover call.
func (l *Loader) Invalidate() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.index = make(map[string]SkillMeta)
	l.lastDiscoverAt = time.Time{}
}

// Discover is implemented in discover.go.

// Load returns full skill content by name.
func (l *Loader) Load(ctx context.Context, name string) (*Skill, error) {
	if err := ValidateSkillName(name); err != nil {
		return nil, err
	}
	dir, err := safeSkillPath(l.skillsDir, name)
	if err != nil {
		return nil, err
	}
	skillFile := filepath.Join(dir, "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSkillNotFound
		}
		return nil, fmt.Errorf("reading skill file: %w", err)
	}
	fm, body, err := parseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing skill %s: %w", name, err)
	}
	meta := SkillMeta{
		Name:          fm["name"],
		Description:   fm["description"],
		Tags:          parseTags(fm["tags"]),
		Version:       fm["version"],
		Author:        fm["author"],
		Source:        fm["source"],
		SourceVersion: fm["source_version"],
	}

	var subFiles []SubFile
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			n := e.Name()
			if !strings.HasSuffix(strings.ToLower(n), ".md") || strings.EqualFold(n, "SKILL.md") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			subFiles = append(subFiles, SubFile{
				Name: n,
				Path: filepath.Join(dir, n),
				Size: info.Size(),
			})
		}
	}

	return &Skill{
		Meta:     meta,
		Content:  body,
		Path:     skillFile,
		SubFiles: subFiles,
		LoadedAt: time.Now(),
		Refs:     ExtractRefs(body),
	}, nil
}

// LoadSubFile loads a specific sub-file's content by skill name and file name.
func (l *Loader) LoadSubFile(ctx context.Context, skillName, fileName string) (string, error) {
	if err := ValidateSkillName(skillName); err != nil {
		return "", err
	}
	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") || filepath.IsAbs(fileName) {
		return "", ErrInvalidPath
	}
	dir, err := safeSkillPath(l.skillsDir, skillName)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrSkillNotFound
		}
		return "", fmt.Errorf("reading sub-file: %w", err)
	}
	return string(data), nil
}

// safeSkillPath validates and returns the absolute path for a skill directory.
// It rejects any path escape attempts.
func safeSkillPath(baseDir, name string) (string, error) {
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") || filepath.IsAbs(name) {
		return "", ErrInvalidPath
	}
	abs, err := filepath.Abs(filepath.Join(baseDir, name))
	if err != nil {
		return "", ErrInvalidPath
	}
	base, err := filepath.Abs(baseDir)
	if err != nil {
		return "", ErrInvalidPath
	}
	if !strings.HasPrefix(abs, base+string(filepath.Separator)) {
		return "", ErrInvalidPath
	}
	return abs, nil
}

// Validate checks if a skill directory is valid.
func (l *Loader) Validate(ctx context.Context, name string) error {
	return l.validateSkill(ctx, name)
}
