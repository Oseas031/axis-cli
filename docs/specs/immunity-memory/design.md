# Immunity Memory Design

**Status**: Planned
**Implements**: `docs/specs/immunity-memory/requirements.md`
**Depends on**: `internal/memory/longterm/` (existing event log), `internal/contextpack/` (existing preview)

## Overview

Immunity Memory is a **thin index over the existing long-term event log**. It does not introduce a new authoritative store. Promotion writes one new event into `tasks.jsonl`; a rebuildable signature index makes Immunity records queryable by failure shape.

```text
existing tasks.jsonl (append-only)
        │
        │  promote: write immunity.promoted event referencing source_task_id
        ▼
internal/memory/immunity/
        ├── index.go     (signature → immunity_id, rebuildable in-memory)
        ├── store.go     (Promote / List / Show / Forget over events)
        └── recall.go    (Recall / RecallSimilar)
        │
        │  surfaced via opt-in flag only
        ▼
contextpack preview (--include-immunity)
```

No background goroutines. No new on-disk file other than the existing event log and one optional snapshot file for cold-start acceleration.

## Architecture

```text
internal/memory/immunity/
  store.go         # Promote, List, Show, Forget, rebuilds via longterm events
  store_test.go
  recall.go        # Recall(signature), RecallSimilar(partial)
  recall_test.go
  signature.go     # Deterministic Signature() construction + hashing
  signature_test.go
  types.go         # ImmunityRecord, Signature, FailureClass

cmd/axis/
  memory_immunity.go    # subcommands under `axis memory immunity ...`
  context_preview.go    # extended with --include-immunity (existing file edit)

internal/contextpack/
  preview.go        # gain AttachImmunity option (existing file edit)
```

## Core Data Model

```go
package immunity

import "time"

// FailureClass is a namespaced failure category.
// Format: "failure.<subsystem>.<reason>" — validated against a known prefix list.
type FailureClass string

// Signature is the deterministic shape used for similarity matching.
// All fields are required; nil maps are represented as empty maps for stable hashing.
type Signature struct {
    IntentKind         string            `json:"intent_kind"`
    NormalizedArgs     map[string]string `json:"normalized_args"`     // sorted-key canonical form
    ContractToolAllow  []string          `json:"contract_tool_allow"` // sorted, deduped
    ErrorClass         FailureClass      `json:"error_class"`
}

// Hash returns the canonical SHA-256 hex of the JSON-encoded Signature
// with sorted map keys and sorted slices. 32 hex chars (first 128 bits).
func (s Signature) Hash() string { /* sha256 over canonical encoding */ }

// ImmunityRecord is the in-memory shape of a promoted failure.
type ImmunityRecord struct {
    ImmunityID    string       `json:"immunity_id"`     // sig-hash + ts suffix
    SourceTaskID  string       `json:"source_task_id"`
    Signature     Signature    `json:"signature"`
    SignatureHash string       `json:"signature_hash"`
    Cause         string       `json:"cause"`           // required, one-line
    FailureClass  FailureClass `json:"failure_class"`
    PromotedBy    string       `json:"promoted_by"`     // actor identifier
    PromotedAt    time.Time    `json:"promoted_at"`
    SourceDigest  string       `json:"source_digest"`
    Deprecated    bool         `json:"deprecated,omitempty"`
    DeprecatedAt  *time.Time   `json:"deprecated_at,omitempty"`
    DeprecateReason string     `json:"deprecate_reason,omitempty"`
}
```

`ImmunityID` format: `imm-<first-12-of-signature-hash>-<unix-ms>`. Stable per promotion event.

## Event Schema (in existing `tasks.jsonl`)

Two new event types under existing event log:

```json
{"event_type":"memory.immunity.promoted","entity_id":"task-abc","timestamp":"2026-05-12T03:00:00Z","payload":{"immunity_id":"imm-...","cause":"...","failure_class":"failure.provider.timeout","signature_hash":"...","signature":{...},"promoted_by":"user:alex"},"source_digest":"sha256:..."}
{"event_type":"memory.immunity.forgotten","entity_id":"imm-...","timestamp":"...","payload":{"reason":"...","forgotten_by":"agent:axis-bootstrap"}}
```

No separate `immunity.jsonl` file. The existing event log is the source of truth. The rationale: per `memory/BOUNDARY.md`, we do not multiply on-disk authoritative stores.

> Requirements FR3 was updated in-place to reflect this design (reuse existing `tasks.jsonl`). No deviation remains between requirements and design.

## Interfaces

```go
package immunity

import "context"

type Store interface {
    Promote(ctx context.Context, in PromoteInput) (ImmunityRecord, error)
    Forget(ctx context.Context, immunityID, reason, actor string) error
    Show(ctx context.Context, immunityID string) (ImmunityRecord, error)
    List(ctx context.Context, filter ListFilter) ([]ImmunityRecord, error)
}

type Recaller interface {
    Recall(ctx context.Context, sig Signature, limit int) ([]ImmunityRecord, error)
    RecallSimilar(ctx context.Context, partial PartialSignature, limit int) ([]ImmunityRecord, error)
}

type PromoteInput struct {
    SourceTaskID string
    Cause        string         // required, non-empty after TrimSpace
    FailureClass FailureClass   // optional; auto-derived if empty
    PromotedBy   string         // required actor ID
}

type ListFilter struct {
    Class           FailureClass
    Since           *time.Time
    IncludeDeprecated bool
    Limit           int
}

// PartialSignature: any field left zero/nil means "any value matches".
type PartialSignature struct {
    IntentKind        string
    ContractToolAllow []string
}
```

## Storage & Indexing

- **Authoritative**: existing `tasks.jsonl`.
- **Runtime state**: `map[signatureHash][]ImmunityRecord` built by scanning `memory.immunity.*` events on startup.
- **Optional snapshot**: `.axis/memory/immunity.snapshot` — a tiny JSON file with `{last_event_offset, records_by_sig}` to skip full event scan on cold start. Corruption triggers full rescan (no data loss; mirrors existing `kv/` snapshot pattern).

All file I/O uses `path/filepath`. Snapshot writes use temp-file + atomic rename.

## Promotion Flow

```
axis memory immunity promote <task-id> --cause "..." [--class <failure.class>] [--by <actor>]
  │
  ├─► load source task terminal event from tasks.jsonl
  │      └─► error if task not terminal OR not failed
  ├─► derive Signature from {intent.kind, intent.normalized_args, contract.tool_allow, error.class}
  │      └─► normalize: sort args by key, sort tool list, drop empty values
  ├─► compute SignatureHash
  ├─► validate Cause (non-empty trimmed), FailureClass (matches known prefix or empty)
  ├─► assemble ImmunityRecord, append memory.immunity.promoted event
  └─► return ImmunityID + suggested next command (per cli-output-conventions.md)
```

Failure path (any precondition violation) → non-zero exit, structured error per `error-code-conventions.md`, no event written.

## Contextpack Integration

`internal/contextpack/preview.go` gains:

```go
type PreviewOptions struct {
    // ... existing fields unchanged
    IncludeImmunity bool
    ImmunityLimit   int // default 3
}
```

When `IncludeImmunity` is true and the current task's intent yields a Signature:

1. Call `Recaller.Recall(ctx, sig, ImmunityLimit)`
2. If fewer than limit, fall back to `RecallSimilar` with `IntentKind` only
3. Attach to preview output under a labeled section: `immunity_advisory` (json) / `Immunity advisory:` (human)

Each advisory line clearly states: `advisory only — not enforced`.

When the flag is absent (default), preview output is byte-identical to current behavior. A regression test pins this.

## CLI Surface

```
axis memory immunity promote <task-id> --cause "..." [--class <failure.class>] [--by <actor>]
axis memory immunity list [--class <failure.class>] [--since 24h] [--deprecated] [--limit N] [--json]
axis memory immunity show <immunity-id> [--json]
axis memory immunity forget <immunity-id> --reason "..." [--by <actor>]
axis context preview --include-immunity [other existing flags]
```

Output examples follow `docs/architecture/cli-output-conventions.md`:

```
Promoted task-abc123 to immunity record imm-9f2c1d-1746960123.
  cause:  provider returned 504 on every retry
  class:  failure.provider.timeout
  next:   axis memory immunity show imm-9f2c1d-1746960123
```

## Boundary Enforcement Tests

The following tests are mandatory (they encode FR8 / requirements §C rejection list):

1. `kernel_does_not_import_immunity` — `go list -deps ./internal/kernel/...` MUST NOT include `internal/memory/immunity`.
2. `model_does_not_import_immunity` — same for `internal/model/`.
3. `contract_does_not_import_immunity` — same for `internal/contract/`.
4. `preview_default_byte_identical` — golden file comparison of `axis context preview` output with and without `--include-immunity` (default off).
5. `no_auto_promotion` — running a failing task does not produce a `memory.immunity.promoted` event.
6. `promote_requires_failed_task` — promoting a successful or non-terminal task exits non-zero.

These run under `go test -race ./internal/memory/immunity/...` and a top-level `boundary_test.go`.

## Concurrency

Single in-memory `map[string]...` guarded by `sync.RWMutex`. No goroutines. Promote, Forget, List, Show, Recall are all synchronous and short. Index rebuild on startup runs in the foreground.

## Cross-Platform Safety

- All paths via `path/filepath`. Slash normalization for any logged path uses `filepath.ToSlash()`.
- Snapshot writes: write to `immunity.snapshot.tmp`, then `os.Rename` (atomic on Windows for same-volume rename).
- LF-only line terminators when writing event payloads (per `memory/BOUNDARY.md`).

## Secrets

- `Cause` is human-written; we do not parse it for secrets and we do NOT echo it into provider prompts. The human authored the string and accepts authorship.
- Signature fields exclude raw args content (only normalized arg **shapes** are hashed, not values, when args may contain secrets). `signature.NormalizeArgs` drops any arg whose key matches a sensitive-key pattern list maintained locally in `internal/memory/immunity/sensitive_keys.go` (P0 hardcoded set: `api_key`, `token`, `bearer`, `password`, `secret`, `credential`, `auth` — case-insensitive substring match). A centralized redaction utility does not currently exist in the repo (`internal/safego/` is panic-recovery, not redaction); per `CLAUDE.md §12` metadata-promotion rule, we keep the list local until a second package needs it.

## Non-Goals (reinforced from requirements)

- No vector / embedding similarity
- No auto-promotion
- No execution path modification
- No capability-ladder change
- No background goroutines
- No new external dependencies

## Resolved Decisions (with reversal conditions)

These were Open Questions during the first Draft pass. Each is now resolved with a documented default and a condition under which a future Spec-RDT may revisit.

### D1: Storage location — reuse `tasks.jsonl`

- **Decision**: events live in existing `tasks.jsonl`; no separate `immunity.jsonl`.
- **Reason**: `internal/memory/BOUNDARY.md` rule "never multiply sources of truth"; one immutable audit chain.
- **Reverse if**: event-log throughput becomes a bottleneck AND profiling shows immunity scan is the hot path. A separate index file is the first remedy; a separate log is a last resort.

### D2: `FailureClass` lives in `internal/memory/immunity/`

- **Decision**: the known-prefix list and `FailureClass` type stay in `internal/memory/immunity/classes.go`.
- **Reason**: per `CLAUDE.md §12` metadata-promotion rule, types move to `internal/types/` only when multiple core modules need them, validation depends on them, or tests need stable cross-package access. None of those hold yet.
- **Reverse if**: a second package (e.g., `internal/kernel/` self-judgement or `internal/contract/`) needs to read or emit `FailureClass` values.

### D3: `PartialSignature` stays minimal (IntentKind + ContractToolAllow)

- **Decision**: P0 partial match supports only two fields; "match any subset" generalization is deferred.
- **Reason**: Karpathy §2 — minimum code. No real-world usage data exists yet to justify the more flexible API surface.
- **Reverse if**: real recall traces show users want similarity on `NormalizedArgs` keys or `ErrorClass` only.

These decisions are recorded so the next Draft → Planned review pass can evaluate them as a batch, not rediscover them.
