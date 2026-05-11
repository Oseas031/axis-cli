# Layered Memory Model Design

## Overview

The Layered Memory Model provides structured memory services for Axis Agent execution. It consists of three layers — Immediate, Working, and Long-term — each with distinct storage contracts, lifecycles, and query semantics.

```text
Long-term Memory (immutable events + derived views)
         |
         |  cache miss / pattern query
         v
Working Memory (retained bundles, queryable working set)
         |
         |  explicit retain / recall
         v
Immediate Memory (single execution cycle context)
         |
         |  budgeted injection
         v
Agent Execution
```

## Architecture

```text
internal/memory/
  immediate/
    immediate.go      # ImmediateContext data model
    builder.go        # Assembles ImmediateContext from Working + runtime state
    builder_test.go
  working/
    engine.go         # WorkingMemory interface + Engine implementation
    engine_test.go
    bundle.go         # WorkingBundle / WorkingSetItem data model
    bundle_test.go
  kv/
    engine.go         # IndexedKV: log + snapshot + index
    engine_test.go    # Destructive tests (corruption, crash, concurrent)
    log.go            # JSONL append-only log writer
    log_test.go
    snapshot.go       # Snapshot file with tiny header
    snapshot_test.go
    index.go          # Offset index loader/saver
    index_test.go
    types.go          # RecordPos, OpType, etc.
  longterm/
    store.go          # Event log append + query interface
    store_test.go
    event.go          # EventRecord data model
    event_test.go
    derived/
      profile.go      # CompetenceProfile builder
      profile_test.go
      miner.go        # PatternMiner (P2+)
  cli/
    # CLI commands integrated into cmd/axis/
```

## Core Data Model

### Immediate Memory

```go
package immediate

import (
    "github.com/axis-cli/axis/internal/types"
)

// ImmediateContext represents the complete situational context of a single
// Agent execution cycle. It is built fresh for every turn.
type ImmediateContext struct {
    TaskID      string
    Intent      string
    Contract    *ContractSnapshot
    WorkingSet  *WorkingSetSnapshot   // pulled from Working Memory
    ToolResults []ToolResult
    TurnCount   int
    Budget      TokenBudget
}

// WorkingSetSnapshot is the subset of Working Memory injected into
// ImmediateContext, after budget and relevance filtering.
type WorkingSetSnapshot struct {
    Bundles []RetainedBundleSummary
}

type RetainedBundleSummary struct {
    BundleID     string
    Type         string
    Source       string        // filepath.ToSlash() normalized path
    Summary      string        // UTF-8 safe head truncation, max 1024 bytes (P0)
    ContentHash  string        // SHA-256 truncated to 128 bit (32 hex chars)
    FileChanged  bool          // true if hash differs from last seen in .seen file
    PacketCount  int
}

// ToolResult captures a tool execution from the current cycle.
type ToolResult struct {
    ToolName   string
    Input      map[string]any
    Output     map[string]any
    Success    bool
    DurationMs int64
}

// TokenBudget tracks how much context budget has been consumed.
// P0 uses rune-count estimation with language-weighted heuristics;
// no external tokenization library is required.
type TokenBudget struct {
    MaxTokens  int
    UsedTokens int
    Remaining  int
}

// EstimateTokens returns a language-aware token approximation.
//   ASCII letters/digits/punctuation: 1 rune ≈ 0.25 token
//   CJK unified ideographs:           1 rune ≈ 1.0 token
//   Other runes:                      1 rune ≈ 0.5 token
// This is deliberately conservative for safety margins.
func EstimateTokens(s string) int {
    tokens := 0
    for _, r := range s {
        switch {
        case r >= ' ' && r <= '~': // ASCII printable
            tokens += 1 // will be divided by 4 below
        case r >= '\u4e00' && r <= '\u9fff': // CJK
            tokens += 4 // 1.0 token per rune after division
        default:
            tokens += 2 // 0.5 token per rune after division
        }
    }
    return tokens / 4
}

// ContractSnapshot captures the contract constraints applicable to this task.
type ContractSnapshot struct {
    ContractID   string
    RequiredTools []string
    Constraints  map[string]string
}
```

**Content Exposure Rule**: Immediate Memory exposes four fields per file: normalized path, UTF-8-safe summary, truncated hash, and a `file_changed` boolean.

| Field | Specification | Rationale |
|---|---|---|
| **Path** | `filepath.ToSlash()` normalized, always relative to project root | Cross-platform deterministic; no `\` vs `/` ambiguity |
| **Summary** | First 1024 UTF-8 bytes of file content, truncated at valid character boundary using `utf8.DecodeLastRuneInString()` | Enough for ~500 CJK chars or ~1024 ASCII chars; captures package/import/head comment; zero dependency |
| **Hash** | `crypto/sha256.Sum256(content)` truncated to first 128 bits (32 hex chars) | Zero external dependency; same length as BLAKE3-128; sufficient for change detection |
| **`file_changed`** | `true` if current hash ≠ last seen hash in `.axis/memory/.seen`; `true` if never seen; updated after Agent observes the file | Allows Agent to skip unchanged files without parsing content |

**Seen File**: `.axis/memory/.seen` is a single-line-per-entry text file:
```
path/to/file.go  a1b2c3d4e5f6...  1746960000
```
Fields: normalized_path, 32-hex-hash, last_seen_unix_timestamp. Loaded at startup, appended on Agent file observation.

**Budget Degradation**: If token budget is exhausted, summaries are stripped entirely, leaving only `path + hash + file_changed` per file. Path-only mode is the absolute minimum.

### Working Memory

```go
package working

import (
    "context"
    "time"

    "github.com/axis-cli/axis/internal/memory/kv"
)

// Memory defines the interface for Working Memory operations.
type Memory interface {
    // Retain adds a bundle to the working set with a reason.
    Retain(ctx context.Context, bundleID string, reason string) error

    // Release removes a bundle from the working set.
    Release(ctx context.Context, bundleID string) error

    // Recall retrieves relevant packets from retained bundles.
    Recall(ctx context.Context, query string, limit int) ([]PacketHit, error)

    // List returns all retained bundles in the working set.
    List(ctx context.Context) ([]WorkingSetItem, error)

    // Clear empties the entire working set.
    Clear(ctx context.Context) error

    // Compact triggers explicit snapshot rebuild.
    Compact() error
}

// Engine is the filesystem-backed implementation of Working Memory.
type Engine struct {
    kv     *kv.Engine
    rootDir string
}

// WorkingSetItem represents a retained bundle in the working set.
type WorkingSetItem struct {
    BundleID    string    `json:"bundle_id"`
    RetainedAt  time.Time `json:"retained_at"`
    Reason      string    `json:"reason"`
    AccessCount int       `json:"access_count"`
}

// PacketHit represents a single packet retrieved by Recall.
type PacketHit struct {
    BundleID    string
    PacketID    string
    Type        string
    Source      string
    Summary     string
    Relevance   float64
}
```

**Key Format**: `wm:bundle:{bundle_id}`

Example: `wm:bundle:ctx-abc123def456`

**Value Format**: JSON self-describing bundle.

```json
{
  "bundle_id": "ctx-abc123def456",
  "task_id": "task-001",
  "contract_id": "default",
  "goal": "fix provider config",
  "packets": [
    {
      "id": "pkt-001",
      "type": "spec",
      "source": "docs/specs/model-provider/design.md",
      "summary": "Provider config validation rules...",
      "relevance": 0.95,
      "reason": "contract_id match"
    }
  ],
  "trace": {
    "selected": [...],
    "excluded": [...]
  },
  "budget": {
    "max_packets": 10,
    "max_bytes": 32768,
    "used_bytes": 24576,
    "truncated": false
  },
  "retained_at": "2026-05-11T12:00:00Z",
  "access_count": 3,
  "source_digest": "sha256:abc..."
}
```

### Long-term Memory

```go
package longterm

import (
    "context"
    "time"
)

// Store defines the interface for Long-term Memory event log operations.
type Store interface {
    // Append writes an immutable event record.
    Append(ctx context.Context, event EventRecord) error

    // QueryEvents retrieves events matching a filter.
    QueryEvents(ctx context.Context, filter EventFilter) ([]EventRecord, error)

    // GetCompetenceProfile returns the derived competence profile for an agent.
    GetCompetenceProfile(ctx context.Context, agentID string) (*CompetenceProfile, error)

    // FindAnalogousTasks searches for historically similar tasks.
    FindAnalogousTasks(ctx context.Context, intent string, limit int) ([]TaskAnalogy, error)

    // GetContractLineage returns the evolution history of a contract.
    GetContractLineage(ctx context.Context, contractID string) ([]ContractVersion, error)
}

// EventRecord is the immutable unit of Long-term Memory.
type EventRecord struct {
    EventType    string         `json:"event_type"`
    EntityID     string         `json:"entity_id"`
    Timestamp    time.Time      `json:"timestamp"`
    Payload      map[string]any `json:"payload"`
    SourceDigest string         `json:"source_digest"`
    DeprecatedAt *time.Time     `json:"deprecated_at,omitempty"` // soft-delete mark
}

// EventFilter defines query constraints for events.
type EventFilter struct {
    EventTypes   []string
    EntityID     string
    After        time.Time
    Before       time.Time
    IncludeDeprecated bool
}

// CompetenceProfile is a derived view from judgement and transition events.
type CompetenceProfile struct {
    AgentID       string
    ProjectID     string
    DomainID      string     // P2+
    LatestScore   float64
    AutonomyLevel string
    History       []CompetenceSample
    SourceDigest  string     // verifiable rebuild anchor
}

type CompetenceSample struct {
    Timestamp time.Time
    Score     float64
    Criteria  map[string]float64
}
```

## Indexed KV Storage (P0)

The `internal/memory/kv` package provides a pure Go, standard-library-only, append-only key-value engine used by Working Memory.

### Three-File Model

```
.axis/memory/working/
  history.jsonl     # append-only log of all put/release operations
  snapshot.bin      # compacted snapshot with tiny header
  index.txt         # key -> (offset, length) mapping
```

#### history.jsonl

Append-only JSONL. Every `Put` or `Delete` appends one line.

```json
{"op":"put","k":"wm:bundle:ctx-001","v":{...},"ts":"2026-05-11T12:00:00Z"}
{"op":"del","k":"wm:bundle:ctx-001","ts":"2026-05-11T12:01:00Z"}
```

Properties:
- Only this file is ever appended to during normal operation.
- Deletion is a tombstone (`op:"del"`), not a physical removal.
- Each line is self-describing and human-readable.

**Encoding rules**:
- Compact JSON: no unnecessary whitespace between tokens.
- `omitempty` on all optional struct fields; empty maps/slices omitted entirely.
- Smart escaping: only escape `"`, `\`, and control characters; no over-escaping of `/` or Unicode.
- `v` is raw JSON (not base64): the value is marshaled as an embedded JSON object, so the outer JSON parser sees it as a nested object.
- Line terminator: **always LF (`\n`)**, never platform-native `CRLF`. `bufio.Scanner` on the read side transparently strips both `LF` and `CRLF`.
- Line-level atomicity: each line is written in a single `Write` + `Sync` cycle; partial lines are detectable and skippable on replay.

#### snapshot.bin

JSON compacted snapshot with a tiny header. Format is the same serialization logic as `history.jsonl`, minus the `op` and `ts` fields.

```
offset  size   content
─────────────────────────────────
0       4      magic: "AXSN"
4       4      version: uint32 = 1
8       8      record_count: uint64
16      8      data_offset: uint64   (header ends here)
24      8      created_at_unix_nano: uint64
32      8      compacted_history_offset: uint64  (history.jsonl byte offset at time of compact)
40      24     reserved (future use)
64      N      records (sorted by key, compact JSONL, one line per record)
```

Each record line:
```json
{"k":"wm:bundle:ctx-001","v":{...}}
```

Properties:
- Records sorted by key for deterministic binary search fallback (P1+).
- Serialization logic 100% shared with `history.jsonl`; only the top-level field set differs (no `op`/`ts`).
- `compacted_history_offset` records the history.jsonl byte offset at compact time; subsequent replays start from here.
- Created from the authoritative in-memory index during `Compact()`.

#### index.txt

Plain text offset index for the snapshot. Generated during compact by pre-computing the exact byte offset and length of each JSON record line in `snapshot.bin`.

```
# Format: key  offset_in_snapshot  length_in_bytes
wm:bundle:ctx-001  64    1024
wm:bundle:ctx-002  1088  512
```

Loaded into an in-memory `map[string]RecordPos` at startup. Enables O(1) random access into the snapshot without scanning.

### Engine Lifecycle

```go
package kv

import (
    "context"
    "os"
    "sync"
)

// Engine is the IndexedKV implementation.
// P0 uses a single sync.Mutex for all operations. Data volumes are small;
// simplicity and correctness outweigh read concurrency.
type Engine struct {
    mu       sync.Mutex
    rootDir  string
    index    map[string]RecordPos  // in-memory authoritative index
    logFile  *os.File              // history.jsonl
    // snapshot and index file handles opened on-demand
}

// Engine limits (defensive boundaries)
const (
    maxKeyLen   = 256   // bytes; "wm:bundle:{id}" should never exceed this
    maxValueLen = 256 * 1024  // 256 KiB per bundle; enough for ~20 packets with summaries
    maxRecordLen = maxKeyLen + maxValueLen + 1024  // generous envelope for JSON overhead
)

// RecordPos identifies where a record lives.
type RecordPos struct {
    File   string // "log" or "snapshot"
    Offset int64
    Length int64
}

func Open(rootDir string) (*Engine, error)
func (e *Engine) Close() error

func (e *Engine) Get(ctx context.Context, key string) ([]byte, error)
func (e *Engine) Put(ctx context.Context, key string, value []byte) error
func (e *Engine) Delete(ctx context.Context, key string) error   // tombstone
func (e *Engine) ScanPrefix(ctx context.Context, prefix string) (Iterator, error)
func (e *Engine) Compact() error  // explicit rebuild of snapshot + index
```

#### Open

1. Lock engine mutex.
2. Create `rootDir` if missing.
3. Open `history.jsonl`; record its current size as `historySize`.
4. If `snapshot.bin` exists:
   - Attempt to read tiny header (64 bytes) and validate magic/version.
   - **If header valid and `index.txt` exists**:
     - Load `index.txt`; verify every offset points inside `snapshot.bin`.
     - If index validation passes:
       - Extract `compacted_history_offset` from header.
       - Load entries into memory `map[string]RecordPos` (all point to `snapshot.bin`).
       - Set replay start = `compacted_history_offset`.
     - If index validation fails (corrupt, offset out of range):
       - Discard index; attempt to rebuild by scanning `snapshot.bin` JSONL lines sequentially.
       - If rebuild succeeds: use rebuilt index; set replay start = `compacted_history_offset`.
       - If rebuild fails: discard snapshot entirely; set replay start = 0.
   - **If header valid but `index.txt` missing**:
     - Rebuild index by scanning `snapshot.bin` JSONL lines sequentially.
     - Extract `compacted_history_offset` from header.
     - Set replay start = `compacted_history_offset`.
   - **If header invalid**:
     - Discard snapshot entirely.
     - Log warning: `snapshot corrupted; falling back to full history replay`.
     - Set replay start = 0.
5. If no snapshot:
   - Memory index is empty.
   - Set replay start = 0.
6. Scan `history.jsonl` from `replay start` to `historySize`:
   - Parse each JSON line; skip malformed lines with warning (do not abort).
   - `op:"put"` → `index[key] = RecordPos{File:"log", Offset:off, Length:len}`.
   - `op:"del"` → delete `key` from index.
7. Return engine.

**Invariant**: The in-memory index is always authoritative. Snapshot is a fast cold-start accelerator; `history.jsonl` is the immutable source of truth. Corruption in snapshot/index never loses data — it only costs a full replay.

#### Get

1. Lock.
2. Lookup `key` in memory index → `RecordPos`.
3. Open the file indicated by `pos.File` (`snapshot.bin` or `history.jsonl`).
4. `Seek(pos.Offset)` → `ReadN(pos.Length)`.
5. Return raw bytes (caller unmarshals).
6. Unlock.

#### Put

1. Validate: key non-empty and ≤ maxKeyLen; value non-nil and ≤ maxValueLen.
2. Lock.
3. Marshal record as compact JSON: `{"op":"put","k":key,"v":valueJSON,"ts":now}`.
   - `valueJSON` is the raw JSON bytes of the value (not base64), embedded directly as a JSON object field.
4. Append to `history.jsonl` with newline.
5. `fsync` log file.
6. Update memory index: `index[key] = RecordPos{File:"log", Offset:newOff, Length:len}`.
7. Unlock.

#### Delete

1. Lock.
2. Append `{"op":"del","k":key,"ts":now}` to `history.jsonl`.
3. `fsync`.
4. Delete `key` from memory index.
5. Unlock.

#### Compact (explicit only)

1. Lock.
2. Record current `history.jsonl` size → `historyOffset`.
3. Create temporary files: `.snapshot.bin.tmp`, `.index.txt.tmp`.
4. For each entry in memory index (sorted by key):
   - Read value from current location (snapshot or log).
   - Write compact JSON line to `.snapshot.bin.tmp`: `{"k":key,"v":value}`.
5. While writing, pre-compute exact byte offset and length of each line; write `.index.txt.tmp`.
6. Write tiny header to `.snapshot.bin.tmp` with `compacted_history_offset = historyOffset`.
7. **Atomic replace** (cross-platform safe):
   - **Unix**: `os.Rename(.tmp, target)` — single atomic operation.
   - **Windows**: rename existing target → `.old`; rename `.tmp` → target; remove `.old`.
     Two-phase, non-atomic window is acceptable for P0; crash during window leaves `.old` as recoverable backup.
8. **Do not truncate, rename, or modify `history.jsonl` in any way.**
9. Unlock.

**P0 constraint**: Compact is triggered only by explicit `axis memory compact` CLI. No background goroutine, no threshold trigger, no exit-time auto-compact.

## Layer Interaction Protocol

### Immediate ← Working

```text
1. Agent declares context.requested_sources (optional in P0)
2. WorkingMemory.List() → all retained bundles
3. ContextBuilder selects relevant bundles by goal matching
4. ContextCompressor applies budget, generates summaries + hashes
5. ImmediateContext.WorkingSet receives the compressed result
```

### Working ← Long-term

```text
1. WorkingMemory.Recall() cache miss
2. LongTermStore.QueryEvents(filter) → historical bundles
3. System suggests historical context; Agent decides to Retain or ignore
4. If Retained, bundle enters Working Memory with reason="recalled_from_ltm"
```

### Long-term ← Working/Immediate

```text
1. Task completion / judgement / transition events
2. LongTermStore.Append(event) → immutable append
3. Derived view builders (incremental) consume new events
4. CompetenceProfile updated; binds to new source_digest
```

## Forgetting Protocol

1. User or Agent requests `axis memory forget <record-id>`.
2. System appends a `forget` event to Long-term log (or `del` to Working log).
3. For Long-term: `EventRecord.DeprecatedAt` is set; original record preserved.
4. For Working: tombstone appended to `history.jsonl`.
5. Queries default to `IncludeDeprecated=false`.
6. Un-forgetting = append a "revoke forget" event or rebuild from raw log.

Physical deletion is architecturally prohibited.

## Competence Profile Attribution

| Phase | Profile Granularity | Authority |
|---|---|---|
| P0-P1 | per-project | Project-level profile is authoritative for all decisions |
| P2 | per-project + per-domain (optional) | Project > Domain > User |
| P3+ | per-project + per-domain + per-user (reference only) | Project remains authoritative |

Global profiles (if any) are read-only cold-start references. They do not influence:
- Autonomy transitions
- Permission decisions
- Task admission

## CLI Commands

### `axis memory retain <bundle-id> --reason "..."`

Retain a context bundle in the working set.

Output: `Retained wm:bundle:<id> at <timestamp>. Working set: <n> bundles.`

### `axis memory release <bundle-id>`

Remove a bundle from the working set.

Output: `Released wm:bundle:<id>. Working set: <n> bundles.`

### `axis memory list`

List all retained bundles.

Output (default human-readable):
```
BUNDLE ID            RETAINED AT          REASON
wm:bundle:ctx-001    2026-05-11 12:00     fix provider config
wm:bundle:ctx-002    2026-05-11 12:05     review judgement
```

`--json` flag emits stable snake_case JSON.

### `axis memory inspect <bundle-id>`

Display full bundle content (packets, trace, budget).

### `axis memory compact`

Explicitly rebuild snapshot and index from history.

Output: `Compacted <n> records. Snapshot: <size> bytes. Index: <entries>.`

### `axis memory query <query>` (P1+)

Recall relevant packets from working set by keyword.

## Safety Boundaries

- Memory layers do not push context into provider prompts.
- Memory layers do not execute tools, schedule tasks, or mutate contracts.
- Memory layers do not auto-compact or auto-evict.
- Memory layers do not grant or escalate permissions.
- Memory layers do not store secrets or credentials.
- Memory layers do not make network requests.
- Working Memory does not cross project boundaries.

## Evolution Path

### P0: Explicit layers + minimal KV

- Immediate/Working/Long-term interfaces defined
- `internal/memory/kv` Engine with log + snapshot + index
- CLI: retain, release, list, inspect, compact
- Working Memory key format: `wm:bundle:{bundle_id}`
- Per-project competence profiles

### P1: Query + adaptive summary

- `axis memory query`
- Adaptive summary length by file type
- Budget auto-degrade to path-only mode
- Memory hotspot caching (in-process LRU)
- Optional global read-only profile reference

### P2: Domain-level + pattern mining

- Domain-level profile layer
- PatternMiner from Long-term events
- Binary search in snapshot (sorted keys)
- Incremental derived view maintenance

### P3: Organizational intelligence

- Cross-domain analogous task retrieval
- Competence prediction from behavior patterns
- Active context query (Agent-initiated)

## Testing Strategy

- Destructive tests: corrupt snapshot header, malformed JSONL, zero-length files
- Concurrent tests: parallel Get/Put/Compact
- Crash recovery: simulate mid-compact, verify Open rebuilds correctly
- Boundary tests: key size limits, value size limits, empty key
- Cross-platform tests: Windows path handling, atomic rename behavior
