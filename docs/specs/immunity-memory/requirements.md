# Immunity Memory Requirements

**Status**: Draft
**Inspired by**: Jelly Crystal Commons Architecture — Immunity Pool (失败免疫库)
**Related**: `docs/specs/layered-memory-model/`, `docs/architecture/agent-native-first-principles.md`

## Summary

Immunity Memory turns task failures from "events that scroll off the log" into a first-class, queryable memory layer. When an Agent fails at a task, the failure signature (cause, inputs, attempted plan, error class) can be explicitly promoted into an Immunity record. Future tasks with similar signatures may surface these records via contextpack preview, so an Agent can recognize "I've seen this fail before" before re-trying the same path.

Immunity Memory is a **preview-only, opt-in** addition to the existing Long-term Memory event log. It does not change scheduler policy, does not auto-block any execution path, and does not push context into provider prompts.

## Design Philosophy

### Failure Is an Asset, Not Noise

The existing `.axis/events/tasks.jsonl` already records failures. What it does not do is make them **retrievable by failure signature** at the moment a similar task is being assembled. Immunity Memory closes that gap with a thin index — not a new store.

### Explicit Promotion, Never Implicit

A failed task does NOT automatically become an Immunity record. A human or Agent must explicitly promote it via `axis memory immunity promote <task-id> --cause "..."`. This preserves Axis First Principle 4 (Layered Isolation) and rejects the Jelly Crystal "Verified Pool auto-graduation" pattern.

### Preview, Never Push

Immunity records may be surfaced through contextpack preview (`axis context preview --include-immunity`). They are never inlined into a provider prompt by the system. The Agent decides whether the immunity record is relevant.

### No Negative Authority

An Immunity record never blocks a task, never lowers an Agent's permissions, never modifies scheduler decisions. It is informational. Capability-ladder descent remains driven by the existing competence-profile path, not by raw immunity count.

## Users

- Agents about to retry a class of task that has historically failed (e.g., "parse vendor X invoice")
- Developers post-mortem-ing a series of related failures
- Future evaluation loops measuring whether immunity surfacing reduces repeat-failure rate

## Functional Requirements

### FR1: Immunity record schema

An Immunity record MUST contain:

- `immunity_id` — stable identifier
- `source_task_id` — the failed task this was promoted from
- `signature` — a deterministic hash of `{intent.kind, intent.normalized_args, contract.tool_allowlist, error.class}` for similarity matching
- `cause` — human-written one-line cause (required at promotion time)
- `failure_class` — namespaced enum (e.g., `failure.provider.timeout`, `failure.tool.permission_denied`, `failure.contract.unsatisfied`)
- `promoted_by` — actor (human or agent ID)
- `promoted_at` — RFC3339 timestamp
- `source_digest` — digest of the underlying event subset for audit reproduction

Metadata keys MUST use the `memory.immunity.*` namespace per `metadata-key-conventions.md`.

### FR2: Explicit promotion only

The system MUST NOT auto-promote any failed task into Immunity. Promotion MUST be triggered by:

- `axis memory immunity promote <task-id> --cause "..." [--class <failure.class>]`

A failed task that is never promoted MUST remain only as a normal event in `tasks.jsonl`.

### FR3: Immunity layer storage

Immunity records MUST be stored as events in the **existing** `tasks.jsonl` event log, under namespaced event types:

- `memory.immunity.promoted` — emitted on successful promotion
- `memory.immunity.forgotten` — emitted on soft-forget

A signature → record-id index MUST be maintained as a rebuildable in-memory map (not authoritative). An optional on-disk snapshot file (`.axis/memory/immunity.snapshot`) MAY accelerate cold start; corruption MUST trigger full event-log rescan with no data loss.

Pure Go standard library only. No SQLite, no embedding store, no separate authoritative JSONL.

> Rationale for design-phase update: an early Draft of FR3 specified a separate `.axis/memory/immunity.jsonl`. During design, `internal/memory/BOUNDARY.md` review confirmed this would multiply sources of truth for failure history. Reusing `tasks.jsonl` keeps a single immutable audit chain.

### FR4: Signature-based recall

The memory layer MUST expose:

- `Recall(signature, limit)` — exact signature match
- `RecallSimilar(partial_signature, limit)` — match on subset of signature fields (e.g., same intent.kind + same tool_allowlist regardless of args)

P0 MUST NOT use vector similarity. Field-equality similarity is sufficient.

### FR5: Contextpack opt-in surface

Contextpack MUST gain an opt-in flag (`--include-immunity`) that, when set during preview, attaches up to N (configurable, default 3) matching Immunity records to the preview output under a clearly labeled section. Without the flag, behavior is unchanged.

The preview MUST mark immunity records as **advisory only** in its output. The provider call path is NOT modified.

### FR6: Soft-forget, never delete

An Immunity record may be marked `deprecated` (e.g., "this failure mode was fixed by tool upgrade"). The original JSONL event is preserved per `layered-memory-model` FR8. Deprecated records are excluded from default Recall.

### FR7: CLI surface

P0 commands:

- `axis memory immunity promote <task-id> --cause "..." [--class <failure.class>]`
- `axis memory immunity list [--class <failure.class>] [--since <duration>]`
- `axis memory immunity show <immunity-id>`
- `axis memory immunity forget <immunity-id> --reason "..."`
- `axis context preview --include-immunity` (extends existing preview command)

Output MUST follow `cli-output-conventions.md` (human default; `--json` for stable snake_case).

### FR8: Non-invasive boundaries

Immunity Memory MUST NOT:

- Inject records into provider prompts automatically
- Block, delay, or re-route scheduler decisions
- Change capability-ladder state (no auto-descent on N immunity hits)
- Mutate `tasks.jsonl` — promotion writes a new event referencing the source
- Have any background goroutine, watcher, or auto-compaction

### FR9: Cross-platform safety

All file ops use `path/filepath`. Append/read use the same patterns as existing memory layer. No POSIX-only syscalls.

## Non-Goals

- No vector/embedding similarity in P0–P1
- No automatic promotion based on failure-count thresholds
- No automatic provider-call modification (no "refuse to call if immunity match")
- No immunity-driven capability-ladder changes
- No cross-project shared immunity in P0–P1 (per-project authoritative, mirror of competence-profile rule)
- No GUI, no TUI, no Web UI
- No background indexer or watcher
- No training-export pipeline in this spec (a future spec may export immunity as negative samples, but not here)

## Acceptance Criteria

- `docs/specs/immunity-memory/{requirements,design,tasks}.md` exist and are linked from `docs/architecture/semantic-boundaries.md`.
- A failed task that is never promoted leaves no Immunity record.
- Promotion is a single CLI command and writes one append-only event.
- `axis context preview --include-immunity` attaches matching records; without the flag, preview is byte-identical to current behavior.
- No code path in `internal/kernel/` reads from Immunity Memory.
- No code path in `internal/agent/` or `internal/model/` reads from Immunity Memory unless the Agent's contextpack request explicitly opts in.
- All Immunity operations leave traces in `tasks.jsonl` (promotion, forget) for audit.
- `go test -race ./internal/memory/...` passes including failure-injection tests.
