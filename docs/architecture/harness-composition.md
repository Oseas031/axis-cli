# Harness Composition — How CLAUDE.md Specializes Agent Defaults

> 展开自 CLAUDE.md §13（治理框架）— harness 层级组合


**Status**: Adopted (2026-05-11)
**Audience**: Anyone editing CLAUDE.md, AGENT_INSTRUCTIONS.md, `.devin/skills/`, or `.claude/commands/`.
**Sibling docs**: `engineering-guardrails.md` (compiler-level constraints), `external-tool-boundaries.md` (tool surface), `semantic-boundaries.md` (module surface).

---

## 1. The Problem

Axis is operated by external agent runtimes (Devin CLI, Claude Code, lmh-harness, etc.). Each runtime ships its own "harness": a system prompt + skill protocol + tool policy + safety rules.

`CLAUDE.md` is Axis's **project-level harness**. It sits *underneath* the runtime harness and is loaded as an always-on rule.

When the two harnesses define overlapping rules, three failure modes appear:

1. **Silent override** — runtime default wins because CLAUDE.md never gets re-read. Example: agent runs `go test ./...` instead of CLAUDE.md §9's `go test -race ./...` because the runtime's generic "run tests" instruction fires first.
2. **Double work** — agent re-reads both rule sets every turn. The anti-indicator "agent references CLAUDE.md > 10× per session = rules not internalised, forced re-reading" trips.
3. **Bootstrap bloat** — agent reads 6+ files before the first useful action, because each layer adds its own reading chain.

This document fixes the failure modes by stating one principle and applying it to the existing layers.

---

## 2. The Principle

> **The specialization layer states deltas, not duplicates.**

CLAUDE.md and every `.devin/skills/*/SKILL.md` MUST NOT repeat the runtime's default behavior. They MUST state, in one place each:

1. **What runtime default we keep** (sentence reference is enough)
2. **What we override and why** (the diff)
3. **Where the full rule lives** (link, not copy)

If a rule lives in two files, the second copy is the bug. Delete it; replace with a one-line reference.

Corollary: if the runtime default is fine, say nothing. Silence = compliance.

---

## 3. Concrete Override Surface (Axis ↔ Devin defaults)

The following are the *only* places Axis intentionally overrides the Devin CLI default behavior. Anything not listed below SHOULD follow the runtime default.

| # | Devin default | Axis override | Authority |
|---|---|---|---|
| O1 | "Explore codebase with search tools" before coding | Fixed entry chain: `CLAUDE.md` → `docs/status/current-progress.md` → `HANDOVER.md` → `workflow/entry.md` → relevant `BOUNDARY.md` | `AGENT_INSTRUCTIONS.md` |
| O2 | Plan Mode produces an in-chat plan | Non-trivial structural change produces a **Spec-RDT** under `docs/specs/<feature>/` | `CLAUDE.md` §5 |
| O3 | "Run lint / typecheck / build / tests" (generic) | Literal 6-step Axis dev loop: `go build → go test -race → gofmt → go vet → staticcheck → gosec` | `CLAUDE.md` §9 |
| O4 | Conventional-commit subject is sufficient | `feat` / `fix` subjects MUST also carry a milestone tag (`M*`, `Phase N.N`, `T<NN>`) or Spec-RDT reference. `chore` / `docs` / `refactor` / `test` / `ci` / `build` / `perf` are exempt. | `CLAUDE.md` §9 + `scripts/harness-audit.sh` |
| O5 | `todo_write` (in-session) tracks tasks | Cross-session state lives in `docs/status/current-progress.md` (milestone-grained) and `HANDOVER.md` (issues/next-steps). No `session-state.md`, no chat memory. | `AGENT_INSTRUCTIONS.md` |
| O6 | New config files go under `.devin/` | Project-specific config under `.devin/` (skills) or repo root; do **not** write to `.claude/`, `.windsurf/`, `.agents/` — those are read-only compatibility shims. | Devin runtime default (kept) |
| O7 | Pre-commit destructive-op confirmation | Same, **plus** the §1 7-item prohibition self-check before every commit | `CLAUDE.md` §1, axis-bootstrap |

Anything beyond this table — file refs format, tone, parallel tool calls, safety on secrets, cross-platform paths — Axis defers to the runtime. No restating.

---

## 4. Layout Rule

```
Runtime harness (Devin / Claude Code / lmh-harness)
    │  defines: tools, modes, skills protocol, generic safety, style
    │
    ├──[always-on rule]──> CLAUDE.md
    │       defines: §1 prohibitions, §6 boundaries, §9 commit hygiene,
    │                §5 Spec-RDT, §10 engineering practices
    │       MUST NOT redefine: tool semantics, file-ref format, mode names
    │
    └──[on-demand skill]──> .devin/skills/axis-bootstrap/SKILL.md
            content: the 7 overrides from §3 above, in checklist form
            MUST NOT: re-paste §1 prohibitions, re-paste §9 commands,
                      re-paste the reading chain prose from CLAUDE.md
            MUST: link to the canonical section in CLAUDE.md by anchor
```

If a future skill (e.g. `axis-evolution`, `axis-memory`) needs project-specific rules, it follows the same pattern: state its delta, link the canonical authority. Skills are diffs, not copies.

---

## 5. Canonical Worked Example — The 2026-05-11 Fix

This section is the **reference precedent** for "two harnesses are competing, fix it without adding weight."

### Symptoms
- `docs/status/session-state.md` referenced from 4 harness files but **never tracked in git** → phantom state.
- `axis-bootstrap/SKILL.md` Override 1 listed **6 files** to read at start → tripped the "bootstrap > 6 files" anti-indicator.
- `axis-bootstrap/SKILL.md` repeated CLAUDE.md §1 (7-item check) and §9 (commit gate) verbatim → double-source drift risk.
- `scripts/harness-audit.sh` flagged `chore(infra):` and `docs:` commits as missing a milestone, even though `CLAUDE.md` §9 calls scope tag *or* milestone tag sufficient → false positives.

### Diagnosis
All four symptoms were instances of **specialization layer duplicating runtime/canonical content** instead of stating a delta. Once the layer drifted from its source, the agent had to re-read both to find the truth.

### Fix
- Deleted `docs/status/session-state.md`; replaced references in `AGENT_INSTRUCTIONS.md`, `HANDOVER.md`, `axis-bootstrap/SKILL.md` with the already-tracked `current-progress.md` + `HANDOVER.md` pair. (O5 made explicit.)
- Rewrote `axis-bootstrap/SKILL.md` to state only the 7 overrides above, with links to `CLAUDE.md` anchors instead of copies.
- Tightened `scripts/harness-audit.sh`'s milestone metric to exempt `chore` / `docs` / `refactor` / `test` / `ci` / `build` / `perf` scopes. (O4 made explicit.)
- Re-anchored harness-handoff staleness metric to `HANDOVER.md` (tracked) instead of `session-state.md` (phantom).

### Result
- Bootstrap file-read count: 6 → 3 mandatory + N conditional.
- Audit Tier-1 noise: `no-milestone` flag dropped from 9/30 to 4/30 (the remaining 4 are real misses).
- Duplication: ~120 lines of repeated rules removed from the skill; canonical rules now have exactly one source each.

### Reference

When a future contributor sees a similar pattern — a skill, a rule file, or a doc that repeats content from `CLAUDE.md` or from a runtime default — they SHOULD:

1. Identify which side is canonical (CLAUDE.md, BOUNDARY.md, or runtime).
2. Replace the duplicate with a one-line reference + the *delta*, if any.
3. If there is no delta, delete the duplicate entirely.
4. If the duplicate exists because the canonical source is unclear, fix the canonical source first; only then remove the duplicate.

Cite this section (`harness-composition.md` §5) in the PR description so the precedent compounds rather than being re-discovered.

---

## 6. Anti-Indicators to Watch

These signals mean the layering has drifted again and §2 needs re-application:

- Bootstrap file-read count > 5
- Same rule appears verbatim in CLAUDE.md and in any skill
- Skill paragraph > 5 lines without a link to a canonical anchor
- `harness-audit.sh` Tier-1 flags rising over 3 consecutive weekly runs
- Agent references CLAUDE.md > 10× in a single session
- A new runtime (next agent CLI, next IDE plugin) is integrated by *copying* CLAUDE.md content into its own config file

Any of the above → run §5's 4-step playbook on the offending duplicate.
