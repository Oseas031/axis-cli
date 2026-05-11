# Closure Ledger Design

**Status**: Planned
**Implements**: `docs/specs/closure-ledger/requirements.md`
**Depends on**: `internal/memory/longterm/` (existing event log)

## Overview

Closure Ledger is a **read-only derived view over `tasks.jsonl`**. It introduces two namespaced metadata fields on each task's terminal event:

- `axis.closure_gain` — did the task close? (C)
- `axis.cost_risk` — aggregated cost & risk (D)

A ledger view rebuilds these into per-task `LedgerEntry` records on demand. No new authoritative storage. The scheduler, dispatcher, and provider layers do not read this view.

```text
existing tasks.jsonl (terminal events carry closure metadata)
        │
        │  scan + aggregate
        ▼
internal/memory/ledger/  (derived view, rebuildable)
        ├── entry.go     (LedgerEntry, query, aggregate)
        ├── builder.go   (scan tasks.jsonl → entries)
        └── builder_test.go
        │
        │  exposed via CLI
        ▼
axis status --closure   /   axis ledger ...
```

## Architecture

```text
internal/memory/ledger/
  entry.go          # LedgerEntry, ClosureGain, CostRisk types
  builder.go        # Build(events []EventRecord) []LedgerEntry
  builder_test.go
  query.go          # Query(filter), Aggregate(group_by)
  query_test.go
  predicate.go      # ClosurePredicate evaluation at task termination
  predicate_test.go

internal/kernel/
  terminate.go      # extended to invoke predicate.Evaluate + emit metadata (existing file edit)

cmd/axis/
  ledger.go         # `axis ledger list / aggregate / rebuild`
  status.go         # extended with --closure flag (existing file edit)
```

Placement choice: **`internal/memory/ledger/`**, not `internal/kernel/ledger/`. Rationale: requirements FR6 forbids kernel from reading the ledger. Putting the package under `internal/memory/` makes accidental kernel imports more visible during review.

## Core Data Model

```go
package ledger

import "time"

// ClosureGain captures whether and how a task closed.
type ClosureGain struct {
    Closed       *bool       `json:"closed"`             // nil = "undeclared", never silently true
    PredicateID  string      `json:"predicate_id,omitempty"`
    EvidenceRefs []string    `json:"evidence_refs,omitempty"` // event IDs or artifact paths
    Residual     string      `json:"residual,omitempty"`      // free-form, for crystal-unit
    JudgedBy     string      `json:"judged_by,omitempty"`     // actor or "self_judgement"
    JudgedAt     time.Time   `json:"judged_at,omitempty"`
}

// CostRisk aggregates measurable cost dimensions from existing events.
type CostRisk struct {
    TokensIn                int            `json:"tokens_in"`
    TokensOut               int            `json:"tokens_out"`
    TokensTotal             int            `json:"tokens_total"`
    WallclockMs             int64          `json:"wallclock_ms"`
    RetryCount              int            `json:"retry_count"`
    ToolPermissionExpansions int           `json:"tool_permission_expansions"`
    HumanReviewTouches      int            `json:"human_review_touches"`
    ProviderCalls           map[string]int `json:"provider_calls"` // provider_id -> count
}

// LedgerEntry is one task's C/D record. Derived; not authoritative.
type LedgerEntry struct {
    TaskID      string      `json:"task_id"`
    IntentKind  string      `json:"intent_kind"`
    ProviderIDs []string    `json:"provider_ids"`
    TerminalAt  time.Time   `json:"terminal_at"`
    C           ClosureGain `json:"c"`
    D           CostRisk    `json:"d"`
}
```

`*bool` for `Closed` is deliberate: the requirements FR3 specifies "never silently default to true" — a missing predicate yields nil, not false.

## Metadata on Terminal Events

The task's terminal event in `tasks.jsonl` gains:

```json
{
  "event_type": "task.terminated",
  "entity_id": "task-abc",
  "timestamp": "...",
  "payload": {
    "outcome": "success",
    "axis.closure_gain": { "closed": true, "predicate_id": "...", "evidence_refs": [...], "judged_by": "self_judgement" },
    "axis.cost_risk":    { "tokens_total": 12345, "wallclock_ms": 4200, "retry_count": 1, "provider_calls": {"claude-opus-4-7": 2} }
  }
}
```

Namespace `axis.closure_gain` / `axis.cost_risk` complies with `metadata-key-conventions.md`. These are payload fields, not new top-level event fields, so existing event consumers are unaffected.

## Closure Predicate

A task's contract MAY declare a closure predicate. Predicate forms (P0):

| Form | Example | Evaluator |
|---|---|---|
| `file_exists` | `{ "kind": "file_exists", "path": "dist/axis" }` | `os.Stat` |
| `event_emitted` | `{ "kind": "event_emitted", "event_type": "build.success", "since": "task_start" }` | scan `tasks.jsonl` window |
| `exit_zero` | `{ "kind": "exit_zero", "command_event_id": "..." }` | look up referenced event |
| `manual` | `{ "kind": "manual", "judged_by_actor": "..." }` | require human flag at termination |

```go
type ClosurePredicate struct {
    Kind   string         `json:"kind"`
    Params map[string]any `json:"params"`
}

func Evaluate(ctx context.Context, p ClosurePredicate, taskCtx EvalContext) (closed bool, evidence []string, err error)
```

Evaluator package is pure stdlib. Unknown predicate kind → return error → `Closed = nil`, predicate recorded, `judged_by = "unevaluated"`. Never default to true.

## Cost Aggregation

`CostRisk` is built by scanning the task's event range `[task.started, task.terminated]` and counting:

| Field | Source events |
|---|---|
| `tokens_in/out/total` | `provider.call` payload (existing) |
| `wallclock_ms` | terminal_at - started_at |
| `retry_count` | count of `task.retry` events |
| `tool_permission_expansions` | count of `contract.tool_allow_widened` events |
| `human_review_touches` | count of `human.confirmation_requested` and `human.takeover` events |
| `provider_calls` | group `provider.call` events by `provider_id` |

If any source event is missing the expected payload field, aggregation **records the missing field name** in a `cost_risk.warnings` slice but never fails. Auditability over strictness.

## Ledger View Build

```go
package ledger

type Builder struct {
    EventSource longterm.Reader
}

// Build scans the event log and emits one LedgerEntry per terminal task.
// O(N) over events; pure function — same input yields same output.
func (b *Builder) Build(ctx context.Context, filter BuildFilter) ([]LedgerEntry, error)

type BuildFilter struct {
    Since *time.Time
    Until *time.Time
    IntentKind string
    ProviderID string
    OnlyUnclosed bool
}
```

Rebuild correctness: requirements acceptance criterion says `axis ledger rebuild` must produce byte-identical output on a fresh checkout. The builder achieves this by:

- Deterministic JSON marshaling (`json.Marshal` with no maps with unsorted keys; we use `[]struct{K,V}` for provider_calls when serializing to canonical form)
- Stable LedgerEntry ordering by `TerminalAt` then `TaskID`

## Query & Aggregate

```go
type QueryFilter struct {
    IntentKind string
    ProviderID string
    OnlyUnclosed bool
    Since, Until *time.Time
    Limit int
}

type AggregateGroup string
const (
    GroupByIntent   AggregateGroup = "intent"
    GroupByProvider AggregateGroup = "provider"
    GroupByDay      AggregateGroup = "day"
)

type AggregateRow struct {
    Group        string  `json:"group"`
    TaskCount    int     `json:"task_count"`
    ClosedCount  int     `json:"closed_count"`
    TokensTotal  int     `json:"tokens_total"`
    WallclockMs  int64   `json:"wallclock_ms"`
    CloseRate    float64 `json:"close_rate"`
}
```

Aggregation is computed in memory after Build. P0 does not maintain incremental aggregates.

## CLI Surface

```
axis status --closure [--limit N]               # default N=20, last terminated tasks
axis ledger list [--intent <kind>] [--provider <id>] [--unclosed] [--since 7d] [--json]
axis ledger aggregate [--by intent|provider|day] [--since 7d] [--json]
axis ledger rebuild                              # explicit; no auto rebuild
```

Output (human, follows cli-output-conventions.md):

```
$ axis ledger aggregate --by intent --since 7d
intent              tasks  closed  close_rate  tokens   wallclock
build.binary        12     11      91.7%       145,210  00:42:18
scan.repo           5      5      100.0%       28,400   00:08:55
fix.test_failure    8      3       37.5%       212,000  01:50:12
```

Color is **never the sole carrier of meaning** per CLI conventions.

## Boundary Enforcement Tests

1. `kernel_does_not_import_ledger` — `go list -deps ./internal/kernel/...` excludes `internal/memory/ledger`.
2. `contract_does_not_import_ledger` — same for `internal/contract/`.
3. `model_does_not_import_ledger` — same for `internal/model/`.
4. `dispatcher_does_not_consult_ledger` — verify by static check (no symbol references from dispatcher to ledger package).
5. `closed_never_silently_true` — terminate a task with no declared predicate; assert `axis.closure_gain.closed == nil`.
6. `rebuild_deterministic` — twice-call `axis ledger rebuild` on fixture event log; assert byte-identical output.

## Concurrency

`Builder` is stateless. CLI commands instantiate it per invocation. No long-lived goroutines.

## Cross-Platform Safety

- File reads via `path/filepath`
- Atomic rebuild output via temp-file + rename
- No timestamp-formatting locale dependencies (RFC3339 throughout)

## Secrets

`CostRisk` aggregation pulls only counts and IDs. The aggregator MUST NOT dereference `provider.call.payload.request_body` or `response_body`. A unit test asserts the aggregator's struct does not contain any field that could carry a body.

## Non-Goals (reinforced)

- No automatic routing on C/D
- No automatic capability-ladder change
- No budget enforcement / blocking
- No real-time streaming
- No new dependency
- No incremental aggregate cache in P0

## Resolved Decisions (with reversal conditions)

### D1: Eager predicate evaluation at task termination

- **Decision**: closure predicates are evaluated when the task emits `task.terminated`. The result is written into the terminal event's `axis.closure_gain` payload immediately.
- **Reason**: single source of truth on the terminal event; reasoning about "when was this closed" requires no second event. Aligns with First Principle 2 (context is queryable, not recomputed).
- **Reverse if**: a predicate kind needs evidence that only becomes available post-termination (e.g., async build artifacts). At that point, introduce a follow-on `task.closure_rejudged` event rather than making evaluation lazy.

### D2: `manual` predicate defaults to `closed = nil` at termination; human emits later event

- **Decision**: tasks with `kind: manual` predicates emit `closed = nil` on termination. A human (or evaluator agent) later emits a `task.closure_judged` event referencing the task. The ledger view merges the later judgement into the LedgerEntry.
- **Reason**: never silently default to true (requirements FR3); never block termination on human availability.
- **Reverse if**: usage shows the deferred-judgement event is too cumbersome. A `axis task close <task-id> --closed=true --evidence "..."` CLI is the natural next step but is **deferred to a follow-on spec** so this spec stays minimal.

### D3: `evidence_refs` stays as `[]string`

- **Decision**: P0 keeps evidence refs as free-form strings (event ID, artifact path, URL).
- **Reason**: Karpathy §2 — no structure until a second consumer (crystal-unit) demonstrates the need.
- **Reverse if**: crystal-unit promotes to Planned and its design requires typed `{kind, ref}` objects to construct stable `answer_program` references.
