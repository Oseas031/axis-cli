# Immunity Memory Tasks

**Status**: Draft
**Last Updated**: 2026-05-12
**Implements**: `requirements.md` + `design.md` in this directory.

> All tasks are P0 unless marked otherwise. P0 = minimum viable, no external dependencies, no background work. P1 marks reasonable follow-ups left out of the first cut.

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

## T2: Package Skeleton (`internal/memory/immunity/`)

Create the package with no real logic — just types and stub functions — so downstream tasks can be developed in isolation.

### 2.1 Files

- `types.go` — `ImmunityRecord`, `Signature`, `FailureClass`, `PartialSignature`, `PromoteInput`, `ListFilter`
- `classes.go` — known `FailureClass` prefix list + validator
- `errors.go` — typed errors (`ErrTaskNotTerminal`, `ErrTaskNotFailed`, `ErrCauseRequired`, `ErrUnknownFailureClass`)

### 2.2 Validation

- `FailureClass` validator accepts: `failure.provider.*`, `failure.tool.*`, `failure.contract.*`, `failure.intent.*`, `failure.safego.*`, `failure.runtime.*`. Reject everything else with `ErrUnknownFailureClass`.
- `PromoteInput.Validate()` enforces non-empty trimmed `SourceTaskID`, `Cause`, `PromotedBy`.

### 2.3 Tests

- `types_test.go`: zero-value behavior, JSON round-trip
- `classes_test.go`: every accepted prefix + rejected examples
- `errors_test.go`: error wrapping with `errors.Is`

**Acceptance**: `go test ./internal/memory/immunity/...` passes. No imports outside Go stdlib.

---

## T3: Signature Construction & Hashing

### 3.1 `signature.go`

- `func BuildSignature(intentKind string, args map[string]string, toolAllow []string, errClass FailureClass) Signature`
- `func (s Signature) Hash() string` — canonical JSON (sorted map keys, sorted slices, dropped empty values) → SHA-256 → first 32 hex chars
- `func NormalizeArgs(raw map[string]any) map[string]string` — drops keys matching `internal/safego/` secret-key list; stringifies remaining values deterministically

### 3.2 Tests

- Canonical hash stability: same inputs in different key order yield identical hash
- Sensitive-key dropping: `{"api_key": "..."}` → hash equals hash with that key absent
- Empty-value dropping
- 32-hex-char output length

**Acceptance**: hash stability test passes across `GOOS=linux,windows,darwin` in CI matrix.

---

## T4: Event Schema & Store

### 4.1 Extend Long-term Event Reader

In `internal/memory/longterm/`, add (without modifying existing types):

- Constants: `EventTypeImmunityPromoted = "memory.immunity.promoted"`, `EventTypeImmunityForgotten = "memory.immunity.forgotten"`
- No changes to `EventRecord` schema; payloads ride in existing `Payload map[string]any`

### 4.2 `store.go`

- `type Store struct { events longterm.Reader; appender longterm.Appender }`
- `Promote(ctx, in) (ImmunityRecord, error)` — load source task terminal event; refuse if not terminal or not failed; build Signature; append `memory.immunity.promoted` event
- `Forget(ctx, immunityID, reason, actor)` — append `memory.immunity.forgotten` event; never mutates earlier events
- `Show(ctx, immunityID)` — scan events for `promoted` matching ID
- `List(ctx, filter)` — scan + filter (class, since, include_deprecated, limit)

### 4.3 Tests

- `store_test.go` covers (with fixture event-log files):
  - Promote: success path with explicit `--by` actor
  - Promote: rejects non-terminal task → returns `ErrTaskNotTerminal`
  - Promote: rejects successful task → returns `ErrTaskNotFailed`
  - Promote: rejects empty Cause → returns `ErrCauseRequired`
  - Forget: marks deprecated; original `promoted` event untouched
  - List: filter by class returns only matching
  - List: `IncludeDeprecated=false` excludes forgotten records
  - Show: returns record from arbitrary event-log offset

**Acceptance**: `go test -race ./internal/memory/immunity/...` passes including failure-injection tests.

---

## T5: Recall

### 5.1 `recall.go`

- `type Recaller struct { store *Store; index *Index }`
- `Recall(ctx, sig, limit) ([]ImmunityRecord, error)` — exact `SignatureHash` match
- `RecallSimilar(ctx, partial, limit) ([]ImmunityRecord, error)` — match where:
  - if `partial.IntentKind != ""`, signature.IntentKind must equal it
  - if `len(partial.ContractToolAllow) > 0`, signature.ContractToolAllow must be a superset

### 5.2 Index

- `index.go`: `map[signatureHash][]immunityID` rebuilt on startup
- Optional `.axis/memory/immunity.snapshot` cold-start accelerator (write via temp + atomic rename)
- Corruption → discard snapshot, full rescan, no error to caller

### 5.3 Tests

- Exact recall: 10 promoted, 1 query, returns the right one
- Similar recall: superset matching of tool allowlist
- Snapshot corruption: deliberately truncate snapshot file → rebuild succeeds → results identical to no-snapshot run
- Deprecated records excluded from both Recall and RecallSimilar by default

**Acceptance**: snapshot-corruption test passes; all recall tests deterministic.

---

## T6: CLI Surface

In `cmd/axis/`:

### 6.1 `memory_immunity.go` — new file

Subcommands under `axis memory immunity`:

- `promote <task-id> --cause "..." [--class <failure.class>] [--by <actor>]`
- `list [--class ...] [--since <duration>] [--deprecated] [--limit N] [--json]`
- `show <immunity-id> [--json]`
- `forget <immunity-id> --reason "..." [--by <actor>]`

Output rules per `docs/architecture/cli-output-conventions.md`:

- Human default: action verb + primary ID + summary line + "next: ..." hint
- `--json` mode: stable snake_case fields
- Exit codes per `error-code-conventions.md`

### 6.2 Tests

- Golden-file tests for human output of each subcommand (success + failure)
- JSON schema test: `axis memory immunity list --json` produces parseable, schema-stable output
- Help text test: `axis memory immunity --help` lists all four subcommands

**Acceptance**: `axis memory immunity promote` works end-to-end against a real `.axis/events/tasks.jsonl`.

---

## T7: Contextpack Preview Integration

### 7.1 Extend `internal/contextpack/preview.go`

- Add `IncludeImmunity bool` and `ImmunityLimit int` (default 3) to `PreviewOptions`
- When set, call `Recaller.Recall` (then `RecallSimilar` fallback if fewer than limit)
- Attach result under labeled section `immunity_advisory` (json) / `Immunity advisory (advisory only — not enforced):` (human)

### 7.2 Extend `cmd/axis/context_preview.go`

- Add `--include-immunity` flag (boolean) and `--immunity-limit` flag (int)
- Default off → output byte-identical to current behavior

### 7.3 Tests

- **Golden regression**: `axis context preview` with no `--include-immunity` flag → byte-identical to baseline (pinned golden file)
- With flag + matching records → advisory section present and well-formed
- With flag + no matches → advisory section absent (not "empty advisory")
- JSON mode: `immunity_advisory` field absent when flag off, present when flag on

**Acceptance**: golden regression test catches any accidental default-behavior change.

---

## T8: Boundary Enforcement Tests

In `boundary_test.go` at repo root (or in `internal/memory/immunity/boundary_test.go`):

- T8.1 `TestKernelDoesNotImportImmunity` — uses `go list -deps -json ./internal/kernel/...` and asserts `internal/memory/immunity` not in the dep set
- T8.2 `TestModelDoesNotImportImmunity` — same for `internal/model/...`
- T8.3 `TestContractDoesNotImportImmunity` — same for `internal/contract/...`
- T8.4 `TestPromoteRequiresExplicitInvocation` — run a failing task through normal kernel; assert zero `memory.immunity.promoted` events in resulting event log
- T8.5 `TestSecretsScrubbedFromSignature` — promote with args containing `api_key`; assert hash equals hash with key dropped

**Acceptance**: all five tests run under `go test ./...` and pass on a clean checkout.

---

## T9: Documentation

- [ ] Add row to `docs/architecture/semantic-boundaries.md` for Immunity Memory boundary
- [ ] Add line to `docs/architecture/metadata-key-conventions.md` registering `memory.immunity.*` namespace
- [ ] Update `internal/memory/BOUNDARY.md` with one line referencing immunity as a derived-view consumer of the event log (not a new authoritative store)
- [ ] No new top-level user docs in this spec (the CLI `--help` is the user surface)

**Acceptance**: `grep -r "memory.immunity" docs/` returns the three additions above.

---

## T10: P1 Follow-ups (Out of Scope for First Cut)

- Cross-project shared immunity (per-project authoritative in P0)
- Vector / embedding similarity (field-equality only in P0)
- Auto-suggested promotion from N repeated failures of same signature (must remain human-initiated in P0)
- Training-export pipeline for negative samples (separate spec)
- Hypergraph indexing for residual-signature traversal (separate spec, depends on `crystal-unit` outcome)

**Acceptance**: each P1 item is listed here so the next planner can pick it up without rediscovery.

---

## Definition of Done (Whole Spec)

- All P0 tasks above checked off
- `go test -race ./...` green
- `go vet`, `staticcheck`, `gosec` clean
- No new entry in `go.mod`
- Boundary tests (T8) all pass
- Golden regression on `axis context preview` (T7) holds
- Status transitions (Draft → Planned → In Progress → Completed) per `spec-lifecycle-conventions.md`; each transition emits a `spec.<status>` event referencing verification artifacts. Promoter may be human or Agent per `CLAUDE.md §5`.
