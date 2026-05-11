# Spec Lifecycle Conventions

## Purpose

Axis is spec-first. Specs are implementation contracts, not decorative notes.

## Status Vocabulary

| Status | Meaning |
|---|---|
| `Draft` | being shaped; not ready for implementation |
| `Planned` | accepted direction; implementation not started |
| `In Progress` | implementation underway |
| `Completed` | implemented and verified |
| `Paused` | intentionally deferred, not cancelled |
| `Deprecated` | replaced or no longer authoritative |
| `Cancelled` | intentionally abandoned |

## Required Spec Shape

Feature specs should use:

```text
docs/specs/<feature>/requirements.md
docs/specs/<feature>/design.md
docs/specs/<feature>/tasks.md
```

Large features may add:

```text
workflow-binding.md
```

## Source of Truth

- Requirements define what must be true.
- Design defines how Axis intends to satisfy requirements.
- Tasks define execution order and status.
- Code must not contradict active specs.
- If code and spec diverge, update one intentionally; do not leave silent drift.

## Implementation Gate

Before implementation:

- requirements exist
- design exists for non-trivial features
- tasks exist with clear acceptance criteria
- non-goals are explicit
- affected modules are named

## Evolution Gate

Changes that modify Axis's system structure should be treated as evolution work, not ordinary edits.

Use the [Sandboxed Evolution Protocol](../specs/sandboxed-evolution/) for high-impact changes to:

- specs that redefine execution or permission semantics
- contracts that affect task admission or validation
- workflow files that change implementation process
- context rules that affect readiness or execution context summaries
- future control surfaces that influence autonomy, promotion, or self-modification

Evolution work must separate:

- draft changes
- verified changes
- promoted changes
- discarded changes

## Promotion Semantics (CLAUDE.md v1.1)

A status transition is a **promotion**. Promotion is gated by verification quality, not by promoter identity.

### Who may promote

The promoter MAY be human or Agent. The constitution (`CLAUDE.md §1.4`, `§5`, `§11`) does not require a human gate. What it requires is that the verification criteria themselves are machine-checkable.

### What every promotion must do

1. **Pass all declared verification criteria** — every criterion must be reproducible from the recorded workspace digest. Subjective criteria (e.g. "reviewed by a trusted agent", "looks reasonable") are NOT valid criteria and must be reframed before they can gate promotion.
2. **Emit a `spec.<new-status>` event** to the long-term event log (e.g. `spec.planned`, `spec.in_progress`, `spec.completed`, `spec.deprecated`). The event payload MUST include:
   - `spec_id` (e.g. `immunity-memory`)
   - `from_status`, `to_status`
   - `promoted_by` — actor identifier (human user ID or agent ID); never anonymous
   - `verification_artifacts` — list of refs (test names, event IDs, digest hashes) that constitute the evidence
   - `source_digest` — workspace digest at promotion time, for audit reproduction
3. **Update the spec status atomically** — the status header in `requirements.md`/`tasks.md` and the event log entry MUST be consistent. A reader observing one but not the other indicates a botched promotion that MUST be re-emitted.

### Reversibility

Audit, not approval, is the trust mechanism. Any promoted change can be:

- **reverted** by a counter-event (`spec.demoted`, with `from_status`, `to_status`, reason)
- **quarantined** by setting `Deprecated` with an explicit replacement reference

Both are themselves promotions in the same sense above and follow the same rules.

### What verification criteria look like (good vs bad)

| Good (machine-checkable) | Bad (subjective) |
|---|---|
| `go test -race ./internal/memory/immunity/... passes` | "Code looks clean" |
| `go list -deps ./internal/kernel/... excludes internal/memory/ledger` | "Doesn't seem to leak abstraction" |
| `axis context preview --no-flag output byte-identical to golden file` | "Backward compatible" |
| Event log contains exactly one `memory.immunity.promoted` per CLI invocation | "Works as expected" |
| `gosec ./... reports zero issues` | "Probably secure enough" |

If a spec's `Acceptance Criteria` section contains "Bad" column language, that spec is not promotion-ready until those criteria are rewritten.

## Completion Gate

A task may be marked `Completed` only when:

- code or document change is done
- relevant tests or validation pass
- docs are synchronized
- user-visible behavior is described when applicable

## Pausing M-Series or Large Tracks

Use `Paused` for tracks that remain valid but are not current priority.

Do not use `Deprecated` unless the direction is superseded.

## Deprecation Rule

Deprecated specs should say:

- what replaced them
- why they are deprecated
- whether any code still follows them
