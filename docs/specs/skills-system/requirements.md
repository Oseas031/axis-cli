# Skills System Requirements

**Status**: Planned
**Inspired by**: learn-claude-code s05: Skills — on-demand knowledge loading
**Related**: `docs/architecture/agent-native-first-principles.md`, `reports/analysis/learn-claude-code-agent-execution-layer-analysis-2026-05-12.md`

**[Chinese version](../../zh/specs/skills-system/requirements.md)**

## Summary

Skills System is an on-demand knowledge loading mechanism that allows Agents to dynamically acquire domain knowledge when needed, rather than preloading all knowledge at system startup. This implements the "Query is Context" principle from First Principles — Agents actively query domain knowledge rather than passively accepting system-pushed context.

Skills System uses a two-layer loading architecture:
- **Layer 1 (cheap)**: List available skill names and short descriptions in the system prompt (~100 tokens/skill)
- **Layer 2 (on demand)**: Return full skill content when Agent calls the `load_skill` tool

## Design Philosophy

### Knowledge on Demand

Agents should not carry all possible knowledge at startup. The Skills system lets Agents "load knowledge only when needed." This reduces initial context usage while maintaining flexibility in knowledge acquisition.

### Two-Layer Injection

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: System Prompt (cheap, ~100 tokens/skill)          │
│   Skills available:                                          │
│     - pdf: Process PDF files...                             │
│     - code-review: Review code...                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Agent calls load_skill("pdf")
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Tool Result (on demand, full content)             │
│   <skill name="pdf">                                        │
│     Full PDF processing instructions...                     │
│   </skill>                                                  │
└─────────────────────────────────────────────────────────────┘
```

### Contract is Structure

SKILL.md files are structured contracts containing frontmatter metadata and markdown body content. All Skills follow a unified format that can be programmatically parsed and validated.

### Zero Control

The Skills system does not control the Agent's behavior path. It only provides knowledge; the Agent autonomously decides whether and how to use it. Skills do not change scheduler semantics or modify execution paths.

## Users

- Agents needing domain knowledge to complete tasks (e.g., PDF processing, code review, database operations)
- Developers creating and maintaining Skills
- CI/CD pipelines validating Skills format and consistency

## Functional Requirements

### FR1: SKILL.md Format

Each Skill MUST be a directory containing a `SKILL.md` file:

```yaml
---
name: skill-name           # required, kebab-case format
description: One-line description for discovery  # required, single line
tags: tag1, tag2          # optional, comma-separated
version: 1.0.0            # optional, semantic version
author: author-name       # optional
---

# Skill Title

Detailed instructions in markdown format...

## Section 1
...

## Section 2
...
```

Directory structure example:
```
.axis/
├── skills/
│   ├── pdf/
│   │   └── SKILL.md
│   ├── code-review/
│   │   └── SKILL.md
│   └── agent-builder/
│       ├── SKILL.md
│       ├── scripts/           # optional helper scripts
│       │   └── validate.py
│       └── references/        # optional reference files
│           └── template.yaml
```

### FR2: Two-Layer Loading

**Layer 1: Discovery**

At system startup, scan the `.axis/skills/` directory and inject all Skill metadata (name, description) into the system prompt:

```
Skills available:
  - pdf: Process PDF files - extract text, create PDFs, merge documents.
  - code-review: Review code for quality, security, and best practices.
  - database: Query and manage databases with SQL and ORM support.
```

Layer 1 overhead: approximately 100 tokens per Skill (name and description only).

**Layer 2: Loading**

When Agent calls the `load_skill` tool:
1. Validate skill name exists
2. Read SKILL.md file
3. Return full content as tool_result

```go
// Tool definition
type LoadSkillInput struct {
    Name string `json:"name"` // skill name in kebab-case
}

// Tool result
type LoadSkillOutput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Content     string `json:"content"`  // full markdown body
}
```

### FR3: Skill Directory Structure

```
.axis/skills/
├── <skill-name>/
│   ├── SKILL.md           # required: knowledge body
│   ├── scripts/           # optional: helper scripts
│   │   └── *.py
│   └── references/        # optional: reference files
│       └── *.yaml
```

Constraints:
- Skill name MUST be kebab-case (`^[a-z][a-z0-9-]*[a-z0-9]$`)
- `SKILL.md` is required
- `scripts/` and `references/` directories are optional
- Nested Skill directories are NOT allowed

### FR4: Skill Loader Interface

```go
package skills

// Loader manages skill discovery and loading
type Loader interface {
    // Discover returns all available skill metadata
    Discover(ctx context.Context) ([]SkillMeta, error)
    
    // Load returns full skill content by name
    Load(ctx context.Context, name string) (*Skill, error)
    
    // Validate checks if a skill directory is valid
    Validate(ctx context.Context, name string) error
}

type SkillMeta struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Tags        []string `json:"tags,omitempty"`
    Version     string   `json:"version,omitempty"`
    Author      string   `json:"author,omitempty"`
}

type Skill struct {
    Meta    SkillMeta `json:"meta"`
    Content string    `json:"content"`  // raw markdown body
    Path    string    `json:"path"`     // absolute path to SKILL.md
}
```

### FR5: CLI Surface

P0 commands:

```
axis skills list [--json]
  List all available Skills

axis skills show <skill-name> [--json]
  Show full Skill content

axis skills validate [<skill-name>]
  Validate Skill format (validates all if no name specified)

axis skills create <skill-name>
  Create new Skill directory and SKILL.md template
```

Output rules follow `docs/architecture/cli-output-conventions.md`.

### FR6: Tool Registration

The `load_skill` tool MUST be registered in ToolRegistry:

```go
// In internal/tools/skills.go
func NewLoadSkillTool(loader skills.Loader) *Tool {
    return &Tool{
        Name:        "load_skill",
        Description: "Load a skill by name to get detailed instructions",
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
            name, _ := input["name"].(string)
            return loader.Load(ctx, name)
        },
    }
}
```

### FR7: Non-invasive Boundaries

Skills System MUST NOT:
- Automatically inject Skill content into provider prompts (must go through Agent calling `load_skill`)
- Modify scheduler behavior or execution paths
- Block or delay any task execution
- Change permission model or Contract semantics
- Have any background goroutine or watcher

### FR8: Cross-platform Safety

- All file operations use `path/filepath`
- Path validation: Skill directory MUST be under `.axis/skills/`, path escape is forbidden
- Line endings use LF only
- File encoding MUST be UTF-8

### FR9: Integration with Contextpack

Contextpack MAY optionally include references to loaded Skills:

```go
type ContextPack struct {
    // ... existing fields
    LoadedSkills []string `json:"loaded_skills,omitempty"`  // names of loaded skills
}
```

This is recording only — content is NOT automatically injected.

## Non-Goals

- No nested Skills (one Skill depending on another)
- No Skill version conflict resolution (P0 loads latest version only)
- No remote Skill repositories (local `.axis/skills/` only)
- No Skill runtime isolation (scripts execute in main Agent context)
- No automatic pushing of Skill content to provider prompts
- No scheduler or contract semantic modifications
- No GUI, TUI, or Web UI

## Acceptance Criteria

- `docs/specs/skills-system/{requirements,design,tasks}.md` exist
- `axis skills list` lists all Skills under `.axis/skills/`
- `axis skills show <name>` returns full Skill content
- `axis skills validate` validates all Skills have correct format
- `load_skill` tool is registered in ToolRegistry and callable by Agent
- No code path in `internal/kernel/` or `internal/scheduler/` reads from Skills
- Layer 1 injection token overhead ≤ 150 tokens/skill
- `go test -race ./internal/skills/...` passes
