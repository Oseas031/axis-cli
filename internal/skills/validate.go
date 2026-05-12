package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ValidateFull checks if a skill directory is valid.
// It validates: name format, directory existence, SKILL.md presence,
// frontmatter fields, name consistency, and optional subdirectory types.
func (l *Loader) validateSkill(ctx context.Context, name string) error {
	if err := ValidateSkillName(name); err != nil {
		return err
	}
	dir, err := safeSkillPath(l.skillsDir, name)
	if err != nil {
		return err
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return ErrSkillNotFound
	}
	skillFile := filepath.Join(dir, "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		return ErrMissingSKILLMD
	}
	fm, _, err := parseFrontmatter(string(data))
	if err != nil {
		return fmt.Errorf("invalid frontmatter: %w", err)
	}
	meta := SkillMeta{Name: fm["name"], Description: fm["description"]}
	if err := meta.Validate(); err != nil {
		return err
	}
	if fm["name"] != name {
		return ErrNameMismatch
	}
	// Validate optional subdirectories are actually directories
	for _, sub := range []string{"scripts", "references"} {
		subPath := filepath.Join(dir, sub)
		if si, err := os.Stat(subPath); err == nil && !si.IsDir() {
			return fmt.Errorf("%s must be a directory", sub)
		}
	}
	return nil
}
