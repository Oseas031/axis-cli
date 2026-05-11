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

Verification is evidence for promotion. It is not promotion by itself.

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
