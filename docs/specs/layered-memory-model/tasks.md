# Layered Memory Model Tasks

**Status**: In Progress (P0 Completed)
**Last Updated**: 2026-05-11

## T1: Spec-RDT Finalization

- [x] Write requirements.md
- [x] Write design.md
- [ ] Write tasks.md (this file)
- [ ] Review against CLAUDE.md constraints
- [ ] Review against BOUNDARY.md files
- [ ] Approve spec before implementation

**Acceptance**: All three spec files exist, reviewed, and accepted.

---

## T2: Immediate Memory Data Model (P0)

Implement `internal/memory/immediate/` with zero external dependencies.

### 2.1 Define data structures

- `ImmediateContext`
- `WorkingSetSnapshot`
- `RetainedBundleSummary`
- `ToolResult`
- `TokenBudget`
- `ContractSnapshot`

### 2.2 ContextBuilder

- `BuildSelfContext` refactored from `internal/agent/context_builder.go`
- Integrates with `internal/contextpack` for bundle retrieval
- Applies exact summary rules:
  - Path: `filepath.ToSlash()` normalized, always relative to project root
  - Summary: first 1024 UTF-8 bytes, truncated at valid character boundary (`utf8.DecodeLastRuneInString`)
  - Hash: `crypto/sha256.Sum256(content)` → first 128 bits → 32 hex chars
  - `file_changed`: compared against `.axis/memory/.seen` per-file last-hash entry
- Applies budget degradation: exhausted → path-only mode (path + hash + file_changed, no summary)

### 2.3 Unit tests

- Nil/empty input validation
- Budget exhaustion → path-only mode
- Summary truncation: exactly 1024 bytes max, never splits multi-byte UTF-8 rune
- Hash generation: SHA-256 → 128-bit truncation produces exactly 32 hex chars
- `file_changed` detection: unchanged file → false, modified file → true, new file → true
- `.seen` file roundtrip: load → compare → update → reload
- Path normalization: `internal\\agent\\context.go` (Windows) → `internal/agent/context.go`

**Acceptance**: `go test ./internal/memory/immediate/...` passes with ≥90% coverage.

---

## T3: Indexed KV Engine (P0)

Implement `internal/memory/kv/` — the storage substrate for Working Memory.

### 3.1 Core types

- `Engine`
- `RecordPos`
- `OpType` (`put`, `del`)
- `Iterator`

### 3.2 Log writer

- Append JSONL to `history.jsonl`
- Atomic fsync after each write
- Handle concurrent writers via `sync.Mutex`

### 3.3 Snapshot manager

- Tiny header read/write (64 bytes)
- Binary record serialization (`key_len:2 + key + val_len:8 + val`)
- Atomic rename replacement

### 3.4 Index manager

- Plain text `index.txt` format
- Load into memory `map[string]RecordPos`
- Save from authoritative memory index

### 3.5 Engine lifecycle

- `Open(rootDir)` — load snapshot + replay log tail
- `Close()` — flush, close handles
- `Get(key)` — O(1) via memory map + file seek
- `Put(key, value)` — append log + update memory map
- `Delete(key)` — tombstone log + remove from memory map
- `ScanPrefix(prefix)` — iterate memory map keys
- `Compact()` — rebuild snapshot + index from memory map

### 3.6 Destructive tests

- Corrupt tiny header → graceful fallback to log-only rebuild
- Malformed JSONL line → skip line, log warning, continue
- Zero-length files → treat as empty
- Concurrent Put/Put, Put/Compact, Get/Compact
- Mid-compact crash → Open recovers from log
- Empty key / oversized key / oversized value

### 3.7 Cross-platform tests

- Windows path separators
- Atomic rename behavior on Windows vs. Unix

**Acceptance**: `go test ./internal/memory/kv/...` passes with ≥90% coverage. All destructive tests pass.

---

## T4: Working Memory Engine (P0)

Implement `internal/memory/working/` using the KV engine.

### 4.1 Data structures

- `Memory` interface
- `Engine` struct (wraps `kv.Engine`)
- `WorkingSetItem`
- `PacketHit`

### 4.2 Operations

- `Retain(bundleID, reason)` → `Put("wm:bundle:{id}", marshaled_bundle)`
- `Release(bundleID)` → `Delete("wm:bundle:{id}")`
- `List()` → `ScanPrefix("wm:bundle:")`
- `Recall(query, limit)` → basic keyword match over retained bundle summaries (P0)
- `Clear()` → iterate and delete all `wm:bundle:*` keys
- `Compact()` → delegate to `kv.Engine.Compact()`

### 4.3 Bundle value format

Self-describing JSON including:
- `bundle_id`, `task_id`, `contract_id`, `goal`
- `packets` array
- `trace`, `budget`
- `retained_at`, `access_count`, `source_digest`

### 4.4 Unit tests

- Retain/Release/List roundtrip
- Recall keyword matching
- Clear empties all
- Compact preserves all retained bundles
- Empty reason rejection
- Invalid bundle_id format rejection

**Acceptance**: `go test ./internal/memory/working/...` passes with ≥90% coverage.

---

## T5: Long-term Memory Event Store (P0)

Implement `internal/memory/longterm/` with append-only event log.

### 5.1 Data structures

- `EventRecord`
- `EventFilter`
- `Store` interface
- `FileStore` implementation (JSONL append-only)

### 5.2 Event types (initial)

- `task.created`, `task.completed`, `task.failed`
- `judgement.submitted`
- `autonomy.transitioned`
- `memory.retained`, `memory.released`, `memory.forgotten`
- `tool.executed` (behavioral summary, not full output)

### 5.3 Operations

- `Append(ctx, event)` — append to `.axis/memory/longterm/events.jsonl`
- `QueryEvents(ctx, filter)` — scan with filter (event types, entity_id, time range, deprecated flag)

### 5.4 Soft-delete (forgetting)

- Forgetting appends a `memory.forgotten` event with `deprecated_at`
- Queries default to excluding deprecated records
- Original event remains in the immutable log

### 5.5 Unit tests

- Append and query roundtrip
- Time range filtering
- Event type filtering
- Deprecated exclusion/inclusion
- Concurrent append safety
- Empty filter returns all

**Acceptance**: `go test ./internal/memory/longterm/...` passes with ≥90% coverage.

---

## T6: Derived Views — CompetenceProfile (P1)

Build read-only derived views from Long-term events.

### 6.1 CompetenceProfile builder

- Incremental update from new `judgement.submitted` and `autonomy.transitioned` events
- Per-project authoritative profile
- `source_digest` binding to raw event subset

### 6.2 Profile data model

- `CompetenceProfile`
- `CompetenceSample`

### 6.3 Store integration

- `GetCompetenceProfile(agentID)` queries derived view
- Rebuild trigger: explicit CLI or event threshold (P1+)

### 6.4 Tests

- Incremental update correctness
- Rebuild parity (incremental == full rebuild)
- Cross-project isolation

**Acceptance**: `go test ./internal/memory/longterm/derived/...` passes. Incremental and full rebuild produce identical results.

---

## T7: CLI Commands (P0)

Integrate memory commands into `cmd/axis/`.

### 7.1 Commands

- `axis memory retain <bundle-id> --reason "..."`
- `axis memory release <bundle-id>`
- `axis memory list [--json]`
- `axis memory inspect <bundle-id> [--json]`
- `axis memory compact`

### 7.2 Output conventions

- Human-readable by default
- `--json` for machine mode with stable snake_case fields
- Success: action + object ID + state summary
- Error: action + object ID + concise cause + next step

### 7.3 Tests

- CLI integration tests using `ExecuteCLI` pattern
- JSON output schema validation
- Error case coverage (unknown bundle, empty reason)

**Acceptance**: `go test ./cmd/axis/...` passes. All memory CLI commands have integration tests.

---

## T8: Integration Wiring (P0)

Connect memory layers to existing Axis modules.

### 8.1 Working Memory initialization

- `axis start` initializes Working Memory engine at `.axis/memory/working/`
- Standalone commands create ephemeral engine (in-memory)

### 8.2 ContextBuilder integration

- `internal/agent/context_builder.go` refactored to use `internal/memory/immediate/`
- Retains backward compatibility with existing `SelfContext` consumers

### 8.3 Event emission

- Dispatcher / Executor emits events to Long-term store
- Memory operations emit events to Long-term store

### 8.4 End-to-end tests

- Full flow: task execution → tool use → memory retain → context recall → judgement

**Acceptance**: E2E tests pass. No existing tests broken.

---

## T9: Documentation Synchronization

- Update `HANDOVER.md` with memory model milestone
- Update `CLAUDE.md` if directory boundaries change
- Update `AGENT_INSTRUCTIONS.md` with memory CLI usage
- Update `docs/architecture/semantic-boundaries.md` if needed

**Acceptance**: All docs synchronized. No stale references.

---

## P1+ Deferred Tasks

| Task | Phase | Description |
|---|---|---|
| T10 | P1 | `axis memory query <query>` — keyword recall with TF-IDF boost |
| T11 | P1 | Adaptive summary length by file type |
| T12 | P1 | Budget auto-degrade to path-only mode |
| T13 | P1 | In-process LRU hotspot cache |
| T14 | P1 | Optional global read-only profile reference |
| T15 | P2 | Domain-level competence profiles |
| T16 | P2 | PatternMiner from Long-term events |
| T17 | P2 | Binary search in sorted snapshot |
| T18 | P2 | Incremental derived view maintenance (vs. full rebuild) |
| T19 | P3 | Cross-domain analogous task retrieval |
| T20 | P3 | Competence prediction from behavior patterns |
| T21 | P3 | Active context query (Agent-initiated) |

---

## Completion Gate

All P0 tasks (T1-T9) must be `Completed` before this spec is marked `Completed`.

A task is `Completed` only when:
- Code or document change is done
- Relevant tests or validation pass (`go test -race ./...`)
- Docs are synchronized
- User-visible behavior is described in CLI help text
