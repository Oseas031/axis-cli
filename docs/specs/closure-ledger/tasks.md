# Closure Ledger Tasks

**Status**: Draft
**Last Updated**: 2026-05-12
**Implements**: `requirements.md` + `design.md` in this directory.

> All tasks are P0 unless marked otherwise. The ledger is a **derived view** — no authoritative on-disk state of its own.

---

## T1: Spec-RDT Finalization

- [x] requirements.md
- [x] design.md
- [x] tasks.md (this file)
- [ ] Spec triplet passes consistency check (no requirements/design contradiction; resolved decisions documented)
- [ ] Cross-link from `docs/architecture/semantic-boundaries.md`
- [ ] Status promoted Draft → Planned (promoter may be human or Agent per `CLAUDE.md §5`; promoter records `spec.promoted` event)

**Acceptance**: spec triplet exists, internally consistent, `spec.promoted` event emitted.

---

## T2: Package Skeleton (`internal/memory/ledger/`)

### 2.1 Files

- `types.go` — `ClosureGain`, `CostRisk`, `LedgerEntry`, `QueryFilter`, `AggregateGroup`, `AggregateRow`, `BuildFilter`
- `errors.go` — typed errors (`ErrUnknownPredicateKind`, `ErrPredicateEvalFailed`, `ErrMissingEventField`)
- `package_test.go` — type round-trip and zero-value behavior

### 2.2 Critical schema invariant

- `ClosureGain.Closed` is `*bool` (NOT `bool`). nil means "undeclared". Tests must lock this invariant.

**Acceptance**: `go test ./internal/memory/ledger/...` passes; no imports outside stdlib.

---

## T3: Closure Predicate Evaluator

### 3.1 `predicate.go`

- `type ClosurePredicate struct { Kind string; Params map[string]any }`
- `type EvalContext struct { TaskID string; TaskStart, TaskEnd time.Time; EventReader longterm.Reader; ProjectRoot string }`
- `func Evaluate(ctx context.Context, p ClosurePredicate, ec EvalContext) (closed bool, evidence []string, err error)`

### 3.2 Predicate kinds (P0)

| Kind | Params | Evaluator |
|---|---|---|
| `file_exists` | `{ "path": "<rel>" }` | `os.Stat(filepath.Join(ProjectRoot, path))` |
| `event_emitted` | `{ "event_type": "...", "since": "task_start" }` | scan events in `[TaskStart, TaskEnd]` |
| `exit_zero` | `{ "command_event_id": "..." }` | resolve referenced event; check `payload.exit_code == 0` |
| `manual` | `{}` | return `(false, nil, nil)` paired with `Closed = nil` upstream |

### 3.3 Tests

- Each of 4 kinds: happy path + 2 failure paths each
- Unknown predicate kind → `Closed = nil`, recorded `judged_by = "unevaluated"`, no panic
- `file_exists`: path traversal attempt (`../../etc/passwd`) → reject via `filepath.Clean` + project-root containment check

**Acceptance**: `go test -race ./internal/memory/ledger/...` passes; path-traversal test green.

---

## T4: Cost Aggregator

### 4.1 `cost.go`

- `func AggregateCost(ctx context.Context, taskID string, events []longterm.EventRecord) (CostRisk, []string)`
- Returns the `CostRisk` struct + a `warnings` slice for any missing-field cases
- Never fails: a malformed event subtracts nothing, adds a warning string

### 4.2 Field-source map (locked schema)

| Field | Source event type | Payload field |
|---|---|---|
| `tokens_in/out/total` | `provider.call` | `tokens_in`, `tokens_out` |
| `wallclock_ms` | `task.started` + `task.terminated` | timestamps |
| `retry_count` | `task.retry` | count of events |
| `tool_permission_expansions` | `contract.tool_allow_widened` | count |
| `human_review_touches` | `human.confirmation_requested`, `human.takeover` | count |
| `provider_calls` | `provider.call` | group by `provider_id` |

### 4.3 Tests

- Each field aggregates correctly from fixture events
- Missing payload field → warning string, no panic, no negative value
- Aggregator MUST NOT dereference `request_body` / `response_body` — enforced by a test that fails compilation if `CostRisk` gains any such field (use `reflect`-based assertion)

**Acceptance**: aggregator is pure; same input slice yields same output every call.

---

## T5: Terminal-Event Metadata Emission

### 5.1 Extend `internal/kernel/terminate.go` (existing file edit)

At the point where `task.terminated` event is constructed:

1. Look up the task's declared `ClosurePredicate` from contract (if any)
2. If declared: call `predicate.Evaluate(...)` to compute `closed` + `evidence_refs`
3. If undeclared: set `closed = nil`, `judged_by = "undeclared"`
4. Call `cost.AggregateCost(...)` over the task's event range
5. Attach both as payload fields `axis.closure_gain` and `axis.cost_risk` on the terminal event

### 5.2 Critical safety

- If predicate evaluation errors, write `closed = nil` + `judged_by = "unevaluated"` + the error string under `judgement_error`. **NEVER default to true.**
- Pinned by a test: `TestUnevaluatedNeverClosed`.

### 5.3 Tests

- Predicate declared + satisfied → `closed = true`, evidence populated
- Predicate declared + unsatisfied → `closed = false`
- Predicate declared + evaluator error → `closed = nil`, `judgement_error` set
- No predicate declared → `closed = nil`, `judged_by = "undeclared"`
- Cost metadata always present and well-formed

**Acceptance**: 5 integration tests covering above matrix all pass.

---

## T6: Ledger View Builder

### 6.1 `builder.go`

- `type Builder struct { events longterm.Reader }`
- `func (b *Builder) Build(ctx context.Context, filter BuildFilter) ([]LedgerEntry, error)`
- Reads `task.terminated` events from event log; extracts `axis.closure_gain` + `axis.cost_risk` payload fields; emits `LedgerEntry`

### 6.2 Determinism requirements

- Stable ordering: `TerminalAt` ASC, then `TaskID` ASC
- Canonical JSON marshaling (no random map ordering): when emitting `provider_calls`, convert to sorted `[]struct{ProviderID string; Count int}` representation before marshaling

### 6.3 Tests

- `TestBuildDeterministic`: build twice on same event log; outputs are `bytes.Equal`
- `TestBuildFilter`: each filter dimension (intent, provider, since, until, only_unclosed) returns expected subset
- `TestBuildHandlesMissingMetadata`: a terminal event without `axis.closure_gain` payload produces `LedgerEntry` with `C.Closed = nil`

**Acceptance**: byte-equality test green across reruns.

---

## T7: Query & Aggregate

### 7.1 `query.go`

- `func Query(entries []LedgerEntry, filter QueryFilter) []LedgerEntry`
- `func Aggregate(entries []LedgerEntry, group AggregateGroup) []AggregateRow`
- Pure functions over an entries slice — Build is the only IO

### 7.2 Tests

- Each `QueryFilter` field
- Aggregate by intent / provider / day: counts, close_rate calculation (treat `Closed = nil` as NOT closed for rate, but expose `undeclared_count` separately)

**Acceptance**: aggregation tests cover undeclared correctly (no false-true inflation of close_rate).

---

## T8: CLI Surface

### 8.1 `cmd/axis/ledger.go` — new file

- `axis ledger list [--intent <kind>] [--provider <id>] [--unclosed] [--since 7d] [--limit N] [--json]`
- `axis ledger aggregate [--by intent|provider|day] [--since 7d] [--json]`
- `axis ledger rebuild` — explicit; recompute from raw events; print summary

### 8.2 `cmd/axis/status.go` — existing file edit

- Add `--closure` flag: prints per-task C/D summary for last N terminated tasks (default 20)

### 8.3 Output rules

- Tabular human output for `list` and `aggregate`
- `--json` mode: stable snake_case fields
- Color never sole carrier of meaning
- Exit codes per `error-code-conventions.md`

### 8.4 Tests

- Golden-file tests for human output of each subcommand (success + empty + filter cases)
- JSON schema stability test
- `axis ledger rebuild` against fixture event log: output byte-identical across two runs

**Acceptance**: golden + JSON-stability + rebuild-determinism tests all pass.

---

## T9: Boundary Enforcement Tests

In `boundary_test.go`:

- T9.1 `TestKernelDoesNotImportLedger` — `go list -deps ./internal/kernel/...` excludes `internal/memory/ledger`
- T9.2 `TestContractDoesNotImportLedger` — same for `internal/contract/...`
- T9.3 `TestModelDoesNotImportLedger` — same for `internal/model/...`
- T9.4 `TestDispatcherDoesNotConsultLedger` — static check that dispatcher source references zero ledger symbols
- T9.5 `TestClosedNeverSilentlyTrue` — terminate a task with no predicate; assert `axis.closure_gain.closed == null` in event JSON
- T9.6 `TestCostRiskHasNoBodyFields` — `reflect`-walk `CostRisk` struct; fail if any field name contains `body` / `payload` / `content`

**Acceptance**: all six tests run under `go test ./...` and pass on a clean checkout.

---

## T10: Documentation

- [ ] Add row to `docs/architecture/semantic-boundaries.md` for Closure Ledger boundary (kernel/model/contract MUST NOT read)
- [ ] Add lines to `docs/architecture/metadata-key-conventions.md` for `axis.closure_gain` and `axis.cost_risk`
- [ ] Add a short paragraph to `docs/architecture/agent-native-first-principles.md` clarifying that C/D ledger is observation, not routing input
- [ ] No new top-level user docs (CLI `--help` is the user surface)

**Acceptance**: `grep -r "axis.closure_gain\|axis.cost_risk" docs/` finds the registered entries.

---

## T11: P1 Follow-ups (Out of Scope for First Cut)

- `axis task close <task-id> --closed=true --evidence "..."` CLI for retroactive manual judgement
- Incremental aggregate cache (avoid full event scan for large logs)
- Cross-project C/D comparison
- C/D feeding into competence-profile updates (requires separate spec)
- C/D as scheduler routing input (REJECTED at architecture level until evidence justifies a new sandboxed-evolution proposal)

**Acceptance**: each P1 item is listed here so the next planner can pick it up without rediscovery.

---

## Definition of Done (Whole Spec)

- All P0 tasks checked off
- `go test -race ./...` green
- `go vet`, `staticcheck`, `gosec` clean
- No new entry in `go.mod`
- Boundary tests (T9) all pass
- Determinism tests (T6, T8.4) all pass
- No code path in `internal/kernel/`, `internal/contract/`, or `internal/model/` imports `internal/memory/ledger`
- Status transitions (Draft → Planned → In Progress → Completed) per `spec-lifecycle-conventions.md`; each transition emits a `spec.<status>` event referencing verification artifacts. Promoter may be human or Agent per `CLAUDE.md §5`.
