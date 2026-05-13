# Skills System Design

**Status**: Planned
**Implements**: `docs/specs/skills-system/requirements.md`
**Depends on**: `internal/tools/` (existing ToolRegistry), `internal/contextpack/` (optional integration)

## Overview

Skills System is a lightweight on-demand knowledge loading mechanism. It scans SKILL.md files under the `.axis/skills/` directory, injects metadata into the system prompt at startup, and returns full content when the Agent calls the `load_skill` tool.

```
.axis/skills/
  ├── pdf/
  │   └── SKILL.md
  └── code-review/
      └── SKILL.md
        │
        │  startup: scan and build SkillMeta index
        ▼
internal/skills/
  ├── loader.go      # Loader interface implementation
  ├── loader_test.go
  ├── discover.go    # Discover() - scan .axis/skills/
  ├── discover_test.go
  ├── types.go       # SkillMeta, Skill types
  └── validate.go    # Validate skill directory structure
        │
        │  Layer 1: inject meta to system prompt
        ▼
Agent System Prompt:
  "Skills available: pdf: ..., code-review: ..."
        │
        │  Layer 2: Agent calls load_skill("pdf")
        ▼
Tool Result:
  <skill name="pdf">
    Full PDF processing instructions...
  </skill>
```

No background goroutines. No remote fetching. Pure local file operations.

## Architecture

```text
internal/skills/
  loader.go          # Loader struct with Discover/Load/Validate methods
  loader_test.go
  discover.go        # scan .axis/skills/ directory
  discover_test.go
  types.go           # SkillMeta, Skill, LoadSkillInput, LoadSkillOutput
  validate.go        # validate SKILL.md format and directory structure
  validate_test.go
  errors.go          # typed errors

cmd/axis/
  skills.go          # subcommands under `axis skills ...`

internal/tools/
  skills_tool.go     # register load_skill tool (new file)
```

## Core Data Model

```go
package skills

import "time"

// SkillMeta is the lightweight metadata for discovery.
// This is what gets injected into system prompt.
type SkillMeta struct {
    Name        string   `json:"name"`                  // kebab-case, required
    Description string   `json:"description"`           // one-line, required
    Tags        []string `json:"tags,omitempty"`        // optional
    Version     string   `json:"version,omitempty"`     // semver, optional
    Author      string   `json:"author,omitempty"`      // optional
}

// Skill is the full skill content returned by Load.
type Skill struct {
    Meta      SkillMeta `json:"meta"`
    Content   string    `json:"content"`    // raw markdown body (after frontmatter)
    Path      string    `json:"path"`       // absolute path to SKILL.md
    LoadedAt  time.Time `json:"loaded_at"`  // when the skill was loaded
}

// LoadSkillInput is the tool input schema.
type LoadSkillInput struct {
    Name string `json:"name"` // skill name in kebab-case
}

// LoadSkillOutput is the tool output.
type LoadSkillOutput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Content     string `json:"content"`  // full markdown body
}
```

## SKILL.md Format

```yaml
---
name: skill-name
description: One-line description for discovery
tags: tag1, tag2
version: 1.0.0
author: author-name
---

# Skill Title

Markdown content...

## Section 1
...

## Section 2
...
```

Frontmatter parsing rules:
- Uses `---` delimiters
- Supports YAML format
- `name` and `description` are required
- `tags` is parsed as a comma-separated string array
- All other fields are optional

## Interfaces

```go
package skills

import "context"

// Loader manages skill discovery and loading.
type Loader struct {
    skillsDir string       // absolute path to .axis/skills/
    index     map[string]SkillMeta  // name -> meta cache
    mu        sync.RWMutex
}

// NewLoader creates a new Loader.
func NewLoader(skillsDir string) *Loader {
    return &Loader{
        skillsDir: skillsDir,
        index:     make(map[string]SkillMeta),
    }
}

// Discover returns all available skill metadata.
// Scans .axis/skills/ directory on first call, then returns cached index.
func (l *Loader) Discover(ctx context.Context) ([]SkillMeta, error)

// Load returns full skill content by name.
func (l *Loader) Load(ctx context.Context, name string) (*Skill, error)

// Validate checks if a skill directory is valid.
// Validates SKILL.md format and required frontmatter fields.
func (l *Loader) Validate(ctx context.Context, name string) error

// Reload rescans the skills directory and rebuilds the index.
func (l *Loader) Reload(ctx context.Context) error
```

## Storage & Indexing

- **Skills Directory**: `.axis/skills/`
- **Runtime Cache**: `map[string]SkillMeta` in memory, built on first `Discover()` call
- **No persistence**: Skills are loaded from filesystem on demand

File structure validation:
```
.axis/skills/<skill-name>/
  ├── SKILL.md           # required
  ├── scripts/           # optional
  └── references/        # optional
```

All file I/O uses `path/filepath`. Path validation ensures no escape from `.axis/skills/`.

## Frontmatter Parsing Strategy

P0 uses manual YAML frontmatter parsing without external dependencies:

```go
// parseFrontmatter extracts YAML frontmatter from markdown content.
// Returns (meta map, markdown body, error).
func parseFrontmatter(content string) (map[string]any, string, error) {
    // 1. Check content starts with "---\n"
    // 2. Find closing "---\n"
    // 3. Parse YAML between delimiters (simple key: value parsing for P0)
    // 4. Return rest of content as markdown body
}
```

**P0 supports only flat key-value pairs** (no nested structures):
```yaml
---
name: skill-name
description: One-line description
tags: tag1, tag2
version: 1.0.0
---
```

For P1, consider adding `gopkg.in/yaml.v3` if nested frontmatter is needed.

## Discovery Flow

```
Startup:
  │
  ├─► NewLoader(skillsDir)
  │
  ├─► Discover() on first access
  │      │
  │      ├─► scan .axis/skills/ subdirectories
  │      │
  │      ├─► for each subdirectory:
  │      │      ├─► check SKILL.md exists
  │      │      ├─► parse frontmatter
  │      │      ├─► validate required fields (name, description)
  │      │      └─► add to index map[name]SkillMeta
  │      │
  │      └─► return []SkillMeta
  │
  └─► Inject SkillMeta list into system prompt
         "Skills available:\n  - pdf: Process PDF files...\n  - code-review: Review code..."
```

## Loading Flow

```
Agent calls load_skill("pdf"):
  │
  ├─► Lookup "pdf" in index
  │      └─► error if not found: ErrSkillNotFound
  │
  ├─► Read .axis/skills/pdf/SKILL.md
  │
  ├─► Parse frontmatter and extract markdown body
  │
  └─► Return LoadSkillOutput{
         Name:        "pdf",
         Description: "Process PDF files...",
         Content:     "Full PDF processing instructions...",
       }
```

## Tool Registration

```go
// In internal/tools/skills_tool.go

func RegisterSkillTools(registry *ToolRegistry, loader *skills.Loader) {
    registry.Register(&Tool{
        Name:        "load_skill",
        Description: "Load a skill by name to get detailed instructions. Use this when you need domain-specific knowledge.",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "name": map[string]any{
                    "type":        "string",
                    "description": "Skill name in kebab-case (e.g., 'pdf', 'code-review')",
                },
            },
            "required": []string{"name"},
        },
        Handler: func(ctx context.Context, input map[string]any) (any, error) {
            name, ok := input["name"].(string)
            if !ok || name == "" {
                return nil, skills.ErrSkillNameRequired
            }
            skill, err := loader.Load(ctx, name)
            if err != nil {
                return nil, err
            }
            return &skills.LoadSkillOutput{
                Name:        skill.Meta.Name,
                Description: skill.Meta.Description,
                Content:     skill.Content,
            }, nil
        },
    })
}
```

## CLI Surface

```go
// In cmd/axis/skills.go

// axis skills list
func runSkillsList(cmd *cobra.Command, args []string) error {
    loader := skills.NewLoader(skillsDir)
    metas, err := loader.Discover(context.Background())
    if err != nil {
        return err
    }
    // format output per cli-output-conventions.md
    // default: human-readable table
    // --json: JSON array
}

// axis skills show <skill-name>
func runSkillsShow(cmd *cobra.Command, args []string) error {
    name := args[0]
    loader := skills.NewLoader(skillsDir)
    skill, err := loader.Load(context.Background(), name)
    if err != nil {
        return err
    }
    // format output
}

// axis skills validate [<skill-name>]
func runSkillsValidate(cmd *cobra.Command, args []string) error {
    loader := skills.NewLoader(skillsDir)
    if len(args) > 0 {
        // validate specific skill
        return loader.Validate(context.Background(), args[0])
    }
    // validate all skills
    metas, _ := loader.Discover(context.Background())
    for _, meta := range metas {
        if err := loader.Validate(context.Background(), meta.Name); err != nil {
            return err
        }
    }
    return nil
}

// axis skills create <skill-name>
func runSkillsCreate(cmd *cobra.Command, args []string) error {
    name := args[0]
    // validate name format (kebab-case)
    // create directory .axis/skills/<name>/
    // create SKILL.md with template content
}
```

## Layer 1 Injection

```go
// In internal/agent/prompt.go (or equivalent)

func buildSystemPrompt(skillsLoader *skills.Loader) string {
    var sb strings.Builder
    
    sb.WriteString("You are an Axis agent...\n\n")
    
    // Layer 1: inject available skills
    metas, err := skillsLoader.Discover(context.Background())
    if err == nil && len(metas) > 0 {
        sb.WriteString("Skills available:\n")
        for _, meta := range metas {
            sb.WriteString(fmt.Sprintf("  - %s: %s\n", meta.Name, meta.Description))
        }
        sb.WriteString("\nUse load_skill(name) to load detailed instructions.\n\n")
    }
    
    // ... rest of system prompt
    
    return sb.String()
}
```

Estimated token overhead: ~20 tokens base + ~80 tokens per skill = ~100 tokens/skill.

## Boundary Enforcement Tests

The following tests are mandatory:

1. `TestKernelDoesNotImportSkills` — `go list -deps ./internal/kernel/...` MUST NOT include `internal/skills`
2. `TestSchedulerDoesNotImportSkills` — same for `internal/scheduler/`
3. `TestLoadSkillIsOptIn` — without calling `load_skill`, skill content MUST NOT appear in agent context
4. `TestSkillPathSafety` — skill name `../escape` MUST be rejected
5. `TestSkillNameFormat` — skill names must match `^[a-z][a-z0-9-]*[a-z0-9]$`

## Concurrency

- Single `map[string]SkillMeta` guarded by `sync.RWMutex`
- No goroutines
- Discover, Load, Validate are all synchronous

## Cross-Platform Safety

- All paths via `path/filepath`
- Slash normalization: `filepath.ToSlash()`
- LF-only line terminators
- UTF-8 encoding required

## Non-Goals (reinforced from requirements)

- No nested skills (skill depending on another skill)
- No version conflict resolution
- No remote skill repositories
- No runtime isolation for scripts
- No automatic injection into provider prompts
- No scheduler/contract modifications

## Resolved Decisions

### D1: Skills live in `.axis/skills/`

- **Decision**: Skills are stored under `.axis/skills/<skill-name>/SKILL.md`
- **Reason**: Follows Axis convention of storing configuration in `.axis/` directory
- **Reverse if**: need to support per-user global skills, then add `~/.axis/skills/` as secondary location

### D2: No skill versioning in P0

- **Decision**: Each skill has one version (latest). No version conflict resolution.
- **Reason**: Karpathy §2 — minimum code. Versioning adds complexity without clear use case yet.
- **Reverse if**: real usage shows need for multiple versions of same skill

### D3: No nested skills in P0

- **Decision**: A skill cannot depend on another skill.
- **Reason**: Keeps loading logic simple. No dependency resolution.
- **Reverse if**: real usage shows skill composition is valuable

### D4: Scripts in `scripts/` subdirectory

- **Decision**: Skill may have a `scripts/` subdirectory for helper scripts.
- **Reason**: Keeps skill self-contained while separating knowledge from executable code.
- **Reverse if**: security review shows need for script isolation

These decisions are recorded for future Spec-RDT review.
