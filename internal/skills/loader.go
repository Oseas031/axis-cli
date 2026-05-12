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

// Loader manages skill discovery and loading.
type Loader struct {
	skillsDir string
	index     map[string]SkillMeta
	mu        sync.RWMutex
}

// NewLoader creates a new Loader rooted at skillsDir.
func NewLoader(skillsDir string) *Loader {
	return &Loader{
		skillsDir: skillsDir,
		index:     make(map[string]SkillMeta),
	}
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
		Name:        fm["name"],
		Description: fm["description"],
		Tags:        parseTags(fm["tags"]),
		Version:     fm["version"],
		Author:      fm["author"],
	}
	return &Skill{
		Meta:     meta,
		Content:  body,
		Path:     skillFile,
		LoadedAt: time.Now(),
	}, nil
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
