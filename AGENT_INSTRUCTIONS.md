# AGENT HANDOVER — Single Entry Point

This is the **only** entry point for AI agents picking up the Axis project.
All coding constraints live in `CLAUDE.md` (the constitution). This file tells you **what to read and in what order**.

**Project positioning**: Axis is an **OS for Agents** — execution substrate for Agent autogenesis.
9 syscall primitives implemented: `submit_task / query_state / acquire_context / request_capability / compact / spawn / introspect / yield / checkpoint`

## Reading Order

1. **`CLAUDE.md`** — The constitution. All prohibitions, checklists, and engineering constraints.
2. **`docs/guides/SRS-LOOP-AI-REFERENCE.md`** — 作者工作方法论（A0-A8 SRS Loop）。当作者说"按我的工作流"时指的是这个。
3. **`docs/status/current-progress.md`** — Ground truth for milestone status. (Do NOT rely on `HANDOVER.md` for status.)
4. **`HANDOVER.md`** — Known issues, next steps, recent context for the last hand-off.
5. **`docs/architecture/agent-native-first-principles.md`** — Six design principles.
6. **`workflow/entry.md`** — Route your task to the minimal workflow combination.

## Milestone Status

M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Sandboxed Evolution ✅ | Local Control Plane ✅

Active specs (in progress): `docs/specs/actor-comm/` · `docs/specs/skills-system/` · `docs/specs/layered-memory-model/` · `docs/specs/context-compaction/`

## Boundary Files

Read the adjacent `BOUNDARY.md` before editing these directories:

| Directory | Notes |
|-----------|-------|
| `internal/kernel/` | Scheduler must NOT call provider directly |
| `cmd/axis/` | No Web/TUI; scriptable output; no secret leaks |
| `internal/contextpack/` | Preview-only; never push into provider prompts |
| `internal/agent/` | Never bypass contract layer |
| `internal/memory/` | No physical-delete; no external deps; LF-only line terminators |
| `internal/skills/` | Path safety enforced; scheduler isolation required |
| `internal/actor/` | Homogeneous actor model; no identity bias |
| `internal/comm/` | JSONL mailbox; append-only; no delete |

If editing `CLAUDE.md`, `.devin/skills/`, `.claude/commands/`, `scripts/harness-audit.sh`, or anything that defines how agents behave on this repo, read `docs/architecture/harness-composition.md` first.

## Dev Loop

```bash
go build -o axis-dev.exe ./cmd/axis       # Windows: output to axis-dev.exe to avoid locking axis.exe
go test -race ./...                        # must pass before any commit
gofmt -w . && go vet ./...                 # format + vet
staticcheck ./... && gosec ./...           # lint + security
```

> On Windows, use `axis-dev.exe` not `axis.exe` to avoid overwriting the stable binary.

## What NOT to Read

Anything in `docs/deprecated/` or files marked DRAFT / DEPRECATED.

## After Completing a Task

Synchronize docs by scope of impact:

| Changed | Update |
|---------|--------|
| Milestone progress | `docs/status/current-progress.md` |
| Known issues / next steps | `HANDOVER.md` |
| Doc index or directory structure | `docs/README.md` |
| Workflow status or routing | `workflows/README.md` |
| Human-tracked pending work | `WORKFLOW-HUMAN/pending-*.md` |
| Methodology / collaboration rules | `docs/guides/SRS-LOOP-AI-REFERENCE.md` + `CLAUDE.md §0` |
