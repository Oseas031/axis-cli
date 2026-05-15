package skills

import (
	"regexp"
	"strings"
	"time"
)

var skillNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)

// SkillMeta is lightweight metadata for discovery (Layer 1).
type SkillMeta struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags,omitempty"`
	Version       string   `json:"version,omitempty"`
	Author        string   `json:"author,omitempty"`
	Source        string   `json:"source,omitempty"`
	SourceVersion string   `json:"source_version,omitempty"`
}

// Validate checks required fields and name format.
func (m *SkillMeta) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return ErrSkillNameRequired
	}
	if !skillNameRe.MatchString(m.Name) {
		return ErrInvalidSkillName
	}
	if strings.TrimSpace(m.Description) == "" {
		return ErrDescriptionRequired
	}
	return nil
}

// SubFile represents an additional file in a skill directory.
type SubFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// Skill is the full skill content returned by Load.
type Skill struct {
	Meta     SkillMeta `json:"meta"`
	Content  string    `json:"content"`
	Path     string    `json:"path"`
	SubFiles []SubFile  `json:"sub_files,omitempty"`
	LoadedAt time.Time  `json:"loaded_at"`
	Refs     []SkillRef `json:"refs,omitempty"`
}

// LoadSkillInput is the tool input schema.
type LoadSkillInput struct {
	Name string `json:"name"`
}

// LoadSkillOutput is the tool output.
type LoadSkillOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// ValidateSkillName checks if a skill name is valid.
func ValidateSkillName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrSkillNameRequired
	}
	if !skillNameRe.MatchString(name) {
		return ErrInvalidSkillName
	}
	return nil
}
