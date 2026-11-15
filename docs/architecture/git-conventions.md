---
type: architecture
status: active
created: 2026-05-15
last_verified: 2026-11-15
related:
  - docs/architecture/axis-system-conventions.md
---

> 展开自 CLAUDE.md §9（2026-11-15 轻量化收敛，适配单操作者节奏）

# Git Conventions

Authoritative git workflow reference for the Axis project.

## Branch Strategy

**`main`**: always bisect-safe. Every commit compiles (`go build ./...`), passes `go vet ./...`, and passes `go test -race -short ./...`.

**Direct `main` commit**: default for single-operator iteration (RDM docs, bug fixes, and scoped changes).

**Feature branch (recommended)**: large experimental refactors or multi-day work where isolation helps.

**Branch naming**: `<type>/<scope>-<description>`

Types: `feature` | `fix` | `docs` | `research` | `refactor` | `chore`

## Commit Message Convention

Format (Conventional Commits):

```
<type>(<scope>): <description>

[optional body: what + why]
[optional footer: vigil: <id> | Refs: <spec-id>]
```

**Header**: ≤70 characters, imperative mood; avoid bare `wip` or `fix typo` on `main`.

**Types**: `feat` | `fix` | `docs` | `refactor` | `test` | `chore` | `perf` | `research` | `rdm`

**Scopes**: `agent` | `kernel` | `contextpack` | `model` | `vigil` | `gui` | `human` | `comm` | `docs`

## Commit Granularity

- One logical concern = one commit (target small, bisect-safe commits; keep code and doc sync together when they describe the same feature).

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

## Push discipline

- Push to main 后必须监控 CI（gh run watch），失败立即修复或回滚，不留红 main。
- 每次推送的 commit 集合应 bisect-safe（独立可编译）。
