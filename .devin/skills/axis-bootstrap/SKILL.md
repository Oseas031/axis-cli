---
name: axis-bootstrap
description: Axis-specific deltas against the runtime agent harness (Devin CLI / Claude Code / lmh-harness). Invoke at the start of any Axis task, or when the user says things like "开始 axis 的新任务" / "let's pick this up" / "what's next on axis". Lists only what Axis overrides; defers everything else to the runtime default. Canonical rationale lives in `docs/architecture/harness-composition.md`.
triggers:
  - user
  - model
---

# Axis Bootstrap — Deltas Against Runtime Defaults

> This skill is a **diff**, not a duplicate. It enumerates the seven points where Axis overrides the runtime agent harness. Everything else (tone, tool-call batching, file references, parallel calls, secret safety, cross-platform paths, destructive-op confirmation) follows the runtime default — do not restate it here.
>
> Full rationale, layout, and worked precedent: `docs/architecture/harness-composition.md`.

---

## The Seven Overrides

For each override, the **authority** column points to the single canonical source. If this skill ever conflicts with the authority, the authority wins and this skill is the bug.

| # | Runtime default | Axis override | Authority |
|---|---|---|---|
| O1 | Search-and-explore the codebase | Fixed entry chain (see below) | `AGENT_INSTRUCTIONS.md` |
| O2 | Plan Mode → in-chat plan | Non-trivial structural change → Spec-RDT under `docs/specs/<feature>/` | `CLAUDE.md` §5 |
| O3 | Generic "lint / typecheck / test" | Literal Axis dev loop (see below) | `CLAUDE.md` §9 |
| O4 | Conventional commit subject sufficient | `feat`/`fix` MUST carry milestone or Spec-RDT tag; other scopes exempt | `CLAUDE.md` §9 |
| O5 | `todo_write` for task state | Cross-session state in `current-progress.md` + `HANDOVER.md` only | `AGENT_INSTRUCTIONS.md` |
| O6 | Pre-commit destructive-op gate | Same, **plus** §1 prohibition self-check (see below) | `CLAUDE.md` §1 |
| O7 | One `go build` is enough | Every commit MUST be bisect-safe; no build artifacts staged | `CLAUDE.md` §9 |

---

## O1 — Entry chain (replaces "explore codebase")

Three mandatory reads, in this order. Stop here unless the conditional table below triggers.

```
1. CLAUDE.md                        constitution (§1, §5, §6, §9, §10)
2. docs/status/current-progress.md  milestone ground truth, "where did we pause"
3. HANDOVER.md                      known issues, next steps, recent context
```

Conditional reads — only when scope demands:

| If you will edit | Also read |
|---|---|
| `internal/kernel/`     | `internal/kernel/BOUNDARY.md` |
| `cmd/axis/`            | `cmd/axis/BOUNDARY.md` |
| `internal/contextpack/`| `internal/contextpack/BOUNDARY.md` |
| `internal/agent/`      | `internal/agent/BOUNDARY.md` |
| `internal/memory/`     | `internal/memory/BOUNDARY.md` |
| A milestone task (M1–M6, evolution, control-plane) | matching `docs/specs/<feature>/{requirements,design,tasks}.md` |
| Composition of this skill with another runtime | `docs/architecture/harness-composition.md` |

**Silent probe**: read these yourself before asking. Escalate only on genuine ambiguity.

---

## O2 — Spec-RDT gate (replaces in-chat plan)

Before any non-trivial change, answer three questions. Any "yes" → write or update a Spec-RDT first; do not code inline.

```
1. Does the change touch a §6 semantic-boundary module?
   (AgentTask, AgentContract, Scheduler, Orchestrator, Dispatcher,
    Provider, Tool, Intent Parser, ContextBundle, EvolutionRun)

2. Does it change permission semantics, contract shape, workflow shape,
   context-assembly rules, autonomy surfaces, or scheduler policy?

3. Does it introduce a new metadata key, CLI subcommand, event type,
   or file format?
```

Full criteria, statuses, and lifecycle: `CLAUDE.md` §5 and `docs/architecture/spec-lifecycle-conventions.md`.

---

## O3 — Axis dev loop (replaces generic verification)

Run **literally these four commands** for any code change. Do not summarise to "tests look fine".

```bash
go build -o axis-dev.exe ./cmd/axis
go test -race ./...
gofmt -w . && go vet ./...
staticcheck ./... && gosec ./...
```

If a step is skipped (e.g. `gosec` not installed), say **"not verified"** — not "passed". Authority: `CLAUDE.md` §9, `workflow/entry.md` rule #4.

---

## O4 — Commit message tag

| Scope | Requirement |
|---|---|
| `feat:` / `fix:` | MUST reference milestone (`M*`, `Phase N.N`, `T<NN>`) or Spec-RDT path |
| `chore:` / `docs:` / `refactor:` / `test:` / `ci:` / `build:` / `perf:` / `merge:` | Conventional scope tag alone is sufficient |
| Intentionally test-red | Subject MUST start `wip(red):`; next commit MUST turn it green |

Audit: `scripts/harness-audit.sh` reports weekly. Full rule: `CLAUDE.md` §9.

---

## O5 — State lives in git, not in chat

| Layer | Lives in | Lifetime |
|---|---|---|
| Current task todos | runtime `todo_write` | one session |
| Milestone status, last atomic step, next concrete action | `docs/status/current-progress.md` (tracked) | project |
| Known issues, next steps, recent context | `HANDOVER.md` (tracked) | project |
| Task acceptance state | `docs/specs/<feature>/tasks.md` (tracked) | spec lifetime |

There is no `session-state.md` and no chat-memory anchor. Anything project-specific belongs in a tracked file.

---

## O6 — §1 prohibition self-check (before every commit)

The full 7-item checklist lives in `CLAUDE.md` §1. Reproducing it here would split the source of truth. Read §1 directly and confirm each item against the staged diff.

Quick mnemonic (not a substitute for reading §1): *no frameworks, no daemons, no semantic drift, no autonomy escape, no context push, namespaced metadata, no secrets*. Any "no" → stop, do not commit, fix the violation first.

---

## O7 — Bisect-safe commits

Before `git commit`:

```
□ Message follows O4
□ No build artifacts staged (axis-dev.exe / *.exe / *.test / coverage.out / dist/)
□ `go build ./...` and `go vet ./...` pass on this exact commit
□ If a hook modifies files, stage the modifications and retry the commit
```

Full rule: `CLAUDE.md` §9 *Commit Hygiene*.

---

## When to ask, not guess (Axis-specific)

The runtime already covers generic "don't guess". These additional triggers are Axis-specific and override the runtime default of "interpret and proceed":

| Trigger | Why |
|---|---|
| Task scope might cross §6 boundaries | Spec-RDT gate is the user's call |
| About to physically delete from `internal/memory/` | §4 boundary forbids physical-delete |
| About to add a Go dependency | §7 — no external deps without justification |
| About to change metadata key naming | §12 promotion rule |
| About to bypass `axis start` (any auto-spawn / background process) | §1 absolute prohibition |

---

## Maintenance

If you find yourself adding more than a paragraph to this skill, or restating something already in `CLAUDE.md`, stop. Read `docs/architecture/harness-composition.md` §5 first — the 2026-05-11 fix is the precedent for trimming this kind of drift, not adding to it. The skill is a *diff*; if the diff is growing, either the runtime changed (state which one) or the authority changed (point at the new section).
