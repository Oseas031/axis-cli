package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// CheckDependencies validates that all depends_on skills exist and no conflicts_with skills are present.
// Returns nil if all checks pass.
func (l *Loader) CheckDependencies(ctx context.Context, meta *SkillMetadata) error {
	if len(meta.DependsOn) == 0 && len(meta.ConflictsWith) == 0 {
		return nil
	}
	for _, dep := range meta.DependsOn {
		if err := ctx.Err(); err != nil {
			return err
		}
		p := filepath.Join(l.skillsDir, dep, "SKILL.md")
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("%w: %s", ErrDependencyNotFound, dep)
		}
	}
	for _, conflict := range meta.ConflictsWith {
		if err := ctx.Err(); err != nil {
			return err
		}
		p := filepath.Join(l.skillsDir, conflict, "SKILL.md")
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("%w: %s", ErrConflictDetected, conflict)
		}
	}
	return nil
}
