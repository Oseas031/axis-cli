# Skills System Tasks

**Status**: Planned
**Last Updated**: 2026-05-12
**Implements**: `requirements.md` + `design.md` in this directory.

> All tasks are P0 unless marked otherwise. P0 = minimum viable, no external dependencies, no background work.

---

## T1: Spec-RDT Finalization

- [x] requirements.md
- [x] design.md
- [x] tasks.md (this file)
- [ ] Spec triplet passes consistency check (no requirements/design contradiction; resolved decisions documented)
- [ ] Cross-link from `docs/architecture/semantic-boundaries.md`
- [x] Status promoted Draft → Planned (per `spec-lifecycle-conventions.md`)

**Acceptance**: spec triplet exists, internally consistent, status = Planned.

---

## T2: Package Skeleton (`internal/skills/`)

Create the package with types and interfaces — no implementation logic yet.

### 2.1 Files

- `types.go` — `SkillMeta`, `Skill`, `LoadSkillInput`, `LoadSkillOutput`
- `errors.go` — typed errors (`ErrSkillNotFound`, `ErrSkillNameRequired`, `ErrInvalidSkillName`, `ErrInvalidPath`)
- `loader.go` — `Loader` struct with method stubs

### 2.2 Validation

- `SkillMeta.Validate()` enforces:
  - `Name` is required, matches `^[a-z][a-z0-9-]*[a-z0-9]$`
  - `Description` is required, non-empty after TrimSpace
- `ValidateSkillName(name string) error` — validates skill name format

### 2.3 Tests

- `types_test.go`: SkillMeta validation, zero-value behavior, JSON round-trip
- `errors_test.go`: error wrapping with `errors.Is`

**Acceptance**: `go test ./internal/skills/...` passes. No imports outside Go stdlib.

---

## T3: Skill Discovery (`discover.go`)

### 3.1 Implementation

- `func (l *Loader) Discover(ctx context.Context) ([]SkillMeta, error)`
  - Scan `.axis/skills/` subdirectories on first call
  - For each subdirectory, check `SKILL.md` exists
  - Parse frontmatter (YAML between `---` delimiters)
  - Validate required fields
  - Cache in `l.index` map
  - Return cached on subsequent calls

### 3.2 Helper Functions

- `parseFrontmatter(content string) (map[string]any, string, error)` — parse YAML frontmatter, return meta map and markdown body
- `parseTags(raw string) []string` — parse comma-separated tags

### 3.3 Tests

- Discover finds all skills in fixture directory
- Discover returns cached results on second call
- Discover handles empty skills directory
- Frontmatter parsing: valid YAML, missing delimiters, invalid YAML
- SkillMeta validation: missing name, missing description, invalid name format

**Acceptance**: `go test ./internal/skills/...` passes including edge cases.

---

## T4: Skill Loading (`loader.go`)

### 4.1 Implementation

- `func (l *Loader) Load(ctx context.Context, name string) (*Skill, error)`
  - Validate skill name format
  - Lookup in index (call Discover if index empty)
  - Read `SKILL.md` file
  - Parse frontmatter and extract markdown body
  - Return `Skill` struct

### 4.2 Path Safety

- `func safeSkillPath(baseDir, name string) (string, error)`
  - Join `baseDir` and `name`
  - Resolve to absolute path
  - Verify result is under `baseDir` (no path escape)
  - Reject `..`, `./`, absolute paths in name

### 4.3 Tests

- Load returns skill content for valid skill
- Load returns `ErrSkillNotFound` for non-existent skill
- Load rejects path escape attempts (`../escape`, `..\\escape`)
- Load returns markdown body without frontmatter
- LoadedAt timestamp is set correctly

**Acceptance**: Path safety tests pass, no path escape possible.

---

## T5: Skill Validation (`validate.go`)

### 5.1 Implementation

- `func (l *Loader) Validate(ctx context.Context, name string) error`
  - Check skill directory exists
  - Check `SKILL.md` exists
  - Parse and validate frontmatter
  - Check `scripts/` and `references/` are directories if they exist
  - Return nil if valid, specific error otherwise

### 5.2 Validation Rules

- Directory name must match skill name
- `SKILL.md` must be valid UTF-8
- Frontmatter must contain `name` and `description`
- `name` in frontmatter must match directory name

### 5.3 Tests

- Validate returns nil for valid skill
- Validate returns error for missing SKILL.md
- Validate returns error for invalid frontmatter
- Validate handles optional `scripts/` and `references/` directories

**Acceptance**: All validation rules covered by tests.

---

## T6: CLI Surface (`cmd/axis/skills.go`)

### 6.1 Subcommands

- `axis skills list [--json]`
  - List all available skills
  - Human output: table format
  - `--json`: JSON array

- `axis skills show <skill-name> [--json]`
  - Show full skill content
  - Human output: formatted markdown
  - `--json`: JSON object

- `axis skills validate [<skill-name>]`
  - Validate skill format
  - No name: validate all skills
  - Exit 0 if valid, non-zero otherwise

- `axis skills create <skill-name>`
  - Create skill directory and `SKILL.md` template
  - Validate name format first

### 6.2 Output Format

Per `cli-output-conventions.md`:

```
$ axis skills list
NAME         DESCRIPTION
pdf          Process PDF files - extract text, create PDFs
code-review  Review code for quality and security

$ axis skills show pdf
Name: pdf
Description: Process PDF files - extract text, create PDFs
---
# PDF Processing Skill

You now have expertise in PDF manipulation...

$ axis skills create new-skill
Created skill directory: .axis/skills/new-skill/
Created: .axis/skills/new-skill/SKILL.md
Edit the file to add your instructions.
```

### 6.3 Tests

- Golden-file tests for human output
- JSON schema test for `--json` output
- Help text test

**Acceptance**: `axis skills list` works against real `.axis/skills/` directory.

---

## T7: Tool Registration (`internal/tools/skills_tool.go`)

### 7.1 Implementation

- Create `load_skill` tool
- Register to ToolRegistry
- Handler calls `loader.Load()`

### 7.2 Tool Schema

```json
{
  "name": "load_skill",
  "description": "Load a skill by name to get detailed instructions. Use this when you need domain-specific knowledge.",
  "input_schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "description": "Skill name in kebab-case (e.g., 'pdf', 'code-review')"
      }
    },
    "required": ["name"]
  }
}
```

### 7.3 Tests

- Tool is registered in ToolRegistry
- Tool handler returns skill content for valid name
- Tool handler returns error for invalid name
- Tool output is valid JSON

**Acceptance**: `load_skill` tool callable from agent loop.

---

## T8: Layer 1 System Prompt Injection

### 8.1 Implementation

- Extend system prompt builder to include skills metadata
- Call `loader.Discover()` at startup
- Format as: `Skills available:\n  - <name>: <description>\n`

### 8.2 Token Budget

- Estimate: ~20 tokens base + ~80 tokens per skill
- If total exceeds threshold, truncate list (future: prioritize by context)

### 8.3 Tests

- System prompt contains skills metadata
- System prompt does NOT contain skill content (only metadata)
- Empty skills directory: no skills section in prompt

**Acceptance**: Skills listed in system prompt, content NOT included.

---

## T9: Boundary Enforcement Tests

In `boundary_test.go` or package-level test:

- T9.1 `TestKernelDoesNotImportSkills` — `go list -deps ./internal/kernel/...` excludes `internal/skills`
- T9.2 `TestSchedulerDoesNotImportSkills` — same for `internal/scheduler/`
- T9.3 `TestLoadSkillIsOptIn` — verify skill content not in agent context unless `load_skill` called
- T9.4 `TestSkillPathSafety` — path escape attempts rejected
- T9.5 `TestSkillNameFormat` — invalid names rejected

**Acceptance**: All boundary tests pass under `go test ./...`.

---

## T10: Documentation

- [ ] Add row to `docs/architecture/semantic-boundaries.md` for Skills System boundary
- [ ] Add entry to `docs/architecture/module-and-naming-conventions.md` for `.axis/skills/` directory
- [ ] Create `internal/skills/BOUNDARY.md` with constraint rules
- [ ] Update `WORKFLOW-HUMAN/today-5-12-learnclaude.md` with completion status

**Acceptance**: `grep -r "skills" docs/architecture/` returns the additions above.

---

## T11: Example Skills

Create example skills for testing and documentation:

- `axis skills create pdf`
- `axis skills create code-review`
- `axis skills create database`

Each with realistic SKILL.md content.

**Acceptance**: At least one example skill exists in `.axis/skills/`.

---

## T12: P1 Follow-ups (Out of Scope for First Cut)

- Remote skill repositories (Git clone)
- Skill versioning and conflict resolution
- Nested skills (skill dependencies)
- Script runtime isolation
- Skill marketplace / discovery service
- Auto-prioritization of skills by context relevance

**Note on `scripts/` and `references/` directories**:
P0 only validates these directories exist (if present). Actual usage is deferred to P1:
- `scripts/` — helper scripts that Agent may invoke (needs security review)
- `references/` — supplementary files (templates, schemas) that may be referenced in skill content

**Acceptance**: Each P1 item documented for future planning.

---

## Definition of Done (Whole Spec)

- All P0 tasks above checked off
- `go test -race ./...` green
- `go vet`, `staticcheck`, `gosec` clean
- No new entry in `go.mod` (pure stdlib)
- Boundary tests (T9) all pass
- `axis skills list` and `axis skills show` work end-to-end
- `load_skill` tool callable by Agent
- Status = Planned per `spec-lifecycle-conventions.md`
