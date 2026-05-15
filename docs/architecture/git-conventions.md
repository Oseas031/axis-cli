---
type: architecture
status: active
created: 2026-05-15
last_verified: 2026-05-15
related:
  - docs/architecture/axis-system-conventions.md
---

> 展开自 CLAUDE.md §9

# Git Conventions

Authoritative git workflow reference for the Axis project.

## Branch Strategy

**`main`**: always deployable, bisect-safe. Every commit compiles (`go build ./...`) and passes `go vet ./...`.

**Direct main commit allowed when**:
- RDM operations (docs-only, ≤5 files, ≤100 lines)

**Feature branch required when** (any one):
- >5 files changed
- >200 lines delta
- Crosses ≥2 `internal/` subdirectories

**Branch naming**: `<type>/<scope>-<description>`

Types: `feature` | `fix` | `docs` | `research` | `refactor` | `chore`

Examples: `feature/kernel-crash-recovery`, `fix/vigil-done-grep`, `docs/git-conventions`

## Commit Message Convention

Format (Conventional Commits):

```
<type>(<scope>): <description>
                                        ← blank line
<body>
                                        ← blank line
<footer>
```

**Header**: ≤70 characters, imperative mood, no trailing period.

**Types**: `feat` | `fix` | `docs` | `refactor` | `test` | `chore` | `perf` | `research` | `rdm`

**Scopes**: `agent` | `kernel` | `contextpack` | `model` | `vigil` | `gui` | `methodology`

**Body**: what + why (not how). Wrap at 72 characters.

**Footer**:
- `vigil: <id>` — links to vigil work item
- `Refs: <spec-id>` — links to Spec-RDT (e.g., `Refs: M6 T13`)

At least one of `vigil:` or `Refs:` is required for non-RDM commits.

## Commit Granularity

- One logical concern = one commit
- Target: ≤5 files, ≤200 lines per commit
- Research pipeline: report + vigil update = one commit; code implementation = separate commit(s)

**Never mix in one commit**:
- Code changes + methodology updates
- Code changes + documentation
- Multiple unrelated fixes

## Push Policy

- Push at end of every work session (minimum daily)
- Never accumulate >10 unpushed commits
- New branches: `git push -u origin <branch>`

## Never Commit (Private / Sensitive)

The following must never be tracked in git — enforced by `.gitignore`:

| Pattern | Reason |
|---------|--------|
| `.env`, `.env.*` | Environment secrets (API keys, tokens) |
| `.claude/`, `.devin/`, `.swarm/` | Internal AI tooling state |
| `WORKFLOW-HUMAN/` | Personal workflow notes |
| `workflow/`, `workflows/` | Internal process docs |
| `configs/` | Local configuration (may contain credentials) |
| `*.pem`, `*.key`, `*.p12` | Private keys / certificates |
| `.axis/providers.json` | Contains API keys |
| Any file with hardcoded secrets | Use environment variables instead |

**Pre-commit check**: before every commit, verify `git diff --cached --name-only` contains no sensitive paths.

## Prohibited

| Action | Reason |
|--------|--------|
| `git add .` | Risk of staging secrets, binaries, temp files |
| WIP/temp commits on `main` | Breaks bisect-safety |
| `git push --force` to `main` | Destructive, non-recoverable |
| Committing `.exe`, `.out`, `coverage.out`, `dist/`, `.cache/`, editor temp files | Build artifacts pollute history |
