# Skills System Boundary

## Must NOT

- Push skill content into provider prompts automatically (must go through Agent calling `load_skill`)
- Import or be imported by `internal/kernel/scheduler/`
- Modify scheduler behavior, execution paths, or contract semantics
- Spawn goroutines or background watchers
- Access network or remote resources
- Introduce external dependencies (pure stdlib)

## Must

- All file access confined to `.axis/skills/` via `safeSkillPath()`
- Skill content only loaded when Agent explicitly calls `load_skill`
- Layer 1 injection limited to metadata only (name + description, max 20 skills)
- All paths use `path/filepath` for cross-platform safety
- Validate skill names match `^[a-z][a-z0-9-]*[a-z0-9]$`

## Allowed Dependents

- `internal/model/tool/` — registers `load_skill` tool
- `internal/kernel/orchestrator/` — wires Loader to executor and tool registry
- `internal/contract/executor/` — calls `BuildSkillsPromptSection()` for Layer 1
- `cmd/axis/` — CLI surface

## Enforced By

- `boundary_test.go` — verifies scheduler isolation, opt-in loading, path safety, name format
