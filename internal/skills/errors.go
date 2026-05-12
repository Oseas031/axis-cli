package skills

import "errors"

var (
	ErrSkillNotFound       = errors.New("skill not found")
	ErrSkillNameRequired   = errors.New("skill name is required")
	ErrInvalidSkillName    = errors.New("invalid skill name: must be kebab-case matching ^[a-z][a-z0-9-]*[a-z0-9]$")
	ErrInvalidPath         = errors.New("invalid skill path: path escape detected")
	ErrDescriptionRequired = errors.New("skill description is required")
	ErrMissingSKILLMD      = errors.New("SKILL.md not found in skill directory")
)
