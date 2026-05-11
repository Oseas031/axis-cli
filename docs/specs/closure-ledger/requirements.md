# Closure Ledger Requirements

**Status**: Draft
**Inspired by**: Jelly Crystal Commons Architecture — C+D=1 Ledger (闭合证 / 闭合增益 + 成本风险)
**Related**: `docs/architecture/agent-native-first-principles.md`, `docs/specs/layered-memory-model/`

## Summary

Closure Ledger introduces a **measurement-only** accounting layer for two quantities per task:

- **C — Closure Gain**: how much the task closed (problem resolved, residual reduced, artifact produced)
- **D — Cost & Risk**: token spend, wallclock, retries, human-review touches, tool-permission expansions

The original Jelly Crystal proposal treats `C + D = 1` as a routing optimization target. **Axis explicitly rejects making this a default scheduler input** (see semantic-boundaries: scheduler must not own provider-cost policy). Instead, Axis only **records** these quantities as namespaced metadata on completed tasks. A future spec may, after evidence accumulates, consider promoting any of them to typed fields or routing inputs.

This spec is intentionally small. It is the smallest viable step to make "was this task worth what it cost?" a queryable property of the event log.

## Design Philosophy

### Record First, Route Later

Axis's first principle "Capability is Decision Right" is grounded in **observable behavior**. C and D give us numbers to observe. Routing on them is a separate, later decision that requires its own Spec-RDT and verification.

### Closure Is Task-Defined, Not System-Imposed

The system does NOT decide what "closed" means for a given task. The task's contract (its requirements/design) declares closure predicates. The Closure Ledger only records the boolean outcome and any quantitative evidence the task produced.

### Cost Is Already Mostly Measured

Token usage, wallclock, retry count, and human-review touches are already recorded somewhere in `tasks.jsonl`. This spec primarily aggregates them into a single per-task `cost_risk` view; it does not invent new measurements.

### Pure Sidecar

The ledger is a derived view over `tasks.jsonl`. Rebuilding it from raw events MUST produce identical results. No state lives only in the ledger.

## Users

- Developers asking "which provider × tool combo gives best C/D ratio for intent class X?"
- Future evaluation loops feeding C/D into competence-profile updates
- Humans reviewing weekly audit reports

## Functional Requirements

### FR1: Closure Gain (C) recording

A completed task MUST be able to record one ClosureGain entry containing:

- `closed` — boolean (did the contract's closure predicate hold?)
- `evidence_refs` — list of event IDs or artifact paths supporting the claim
- `residual` — optional string describing what was NOT closed (free-form, for crystal-unit spec to consume later)
- `judged_by` — actor that decided closure (the task's self-judgement, a human review, or an evaluator agent)

Stored under metadata key `axis.closure_gain` on the task's terminal event.

### FR2: Cost & Risk (D) recording

A completed task MUST aggregate from existing events:

- `tokens_in`, `tokens_out`, `tokens_total`
- `wallclock_ms`
- `retry_count`
- `tool_permission_expansions` (count of `tool.allowed_paths` widenings during task)
- `human_review_touches` (count of human-confirmation events)
- `provider_calls` (count, grouped by provider ID)

Stored under metadata key `axis.cost_risk` on the task's terminal event. Aggregation is from existing event records; this spec adds NO new mid-task events.

### FR3: Closure predicate declaration

A task's contract MAY declare a closure predicate. If declared, the system MUST evaluate it at task termination and write the result into `axis.closure_gain.closed`. If not declared, `closed = null` and the field is recorded as "undeclared" — never silently defaulted to true.

### FR4: Ledger view

`internal/kernel/ledger/` (or `internal/memory/ledger/` — design decides) MUST expose a read-only view:

- `LedgerEntry{ task_id, intent_class, provider_id, c, d, judged_by, ts }`
- `Query(filter)` — filter by intent class, provider, time range, closed/unclosed
- `Aggregate(group_by)` — group totals for reporting

The view MUST be rebuildable from `tasks.jsonl` alone. No authoritative state.

### FR5: CLI surface

- `axis status --closure` — per-task C/D summary for recent tasks (default last 20)
- `axis ledger list [--intent <kind>] [--provider <id>] [--unclosed]` — query entries
- `axis ledger aggregate [--by intent|provider|day]` — grouped totals
- `axis ledger rebuild` — explicit rebuild from raw events (no auto-rebuild)

`--json` flag MUST emit stable snake_case fields.

### FR6: No scheduler/dispatcher influence

`internal/kernel/` and `internal/contract/` MUST NOT read from the Closure Ledger. The dispatcher MUST NOT consult C/D when selecting providers. Adding any such read-path requires a new Spec-RDT under `sandboxed-evolution` protocol.

### FR7: No auto-promotion of metrics to typed fields

The aggregated metrics live as namespaced metadata only. Promotion of any metric to a typed Go field follows the rule in `CLAUDE.md §12` (multiple core modules need it, validation depends on it, tests need stable access). This spec MUST NOT promote anything.

### FR8: Cross-platform safety

Standard file I/O via `path/filepath`. Rebuild MUST be atomic on Windows (write-temp-then-rename).

### FR9: Secrets must not leak into ledger

Cost aggregation MUST NOT include provider request/response bodies. Only the token counts and IDs already present in events. The ledger MUST pass `gosec` cleanly.

## Non-Goals

- No automatic provider routing based on C/D
- No automatic capability-ladder changes based on C/D
- No "budget enforcement" — the ledger only observes, it does not block
- No real-time C/D streaming (the ledger is computed at task termination)
- No cross-project C/D comparison in P0 (per-project authoritative)
- No machine-learning-based C-prediction
- No new mid-task events beyond what existing tasks already emit
- No external database, no SQLite, no Prometheus exporter in this spec

## Acceptance Criteria

- `docs/specs/closure-ledger/{requirements,design,tasks}.md` exist.
- A completed task with a declared closure predicate has both `axis.closure_gain` and `axis.cost_risk` metadata.
- A completed task with no declared predicate has `axis.closure_gain.closed = null` and clear "undeclared" status — never silently `true`.
- `axis ledger rebuild` produces byte-identical output to building from scratch on a fresh checkout of the same `tasks.jsonl`.
- `go vet`, `staticcheck`, and `gosec` clean.
- No new dependency in `go.mod`.
- `internal/kernel/` has zero imports from the ledger package.
