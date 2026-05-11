# Layered Memory Model Requirements

## Summary

Axis Layered Memory Model provides a structured, queryable, and auditable memory system for Agent execution. It is not a chatbot conversation history, not a vector database, and not a control plane. It is the substrate that enables Agents to retain, retrieve, and reason over context across execution cycles.

The model consists of three layers:

- **Immediate Memory**: the situational context of a single execution cycle
- **Working Memory**: the active working set of retained bundles across a task chain or session
- **Long-term Memory**: immutable event history and derived competence profiles

Each layer has explicit boundaries, interfaces, and storage contracts. No layer silently injects context into execution, escalates authority, or mutates scheduler semantics.

## Design Philosophy

### Context is Queried, Not Pushed

Agents declare what they need. The system resolves against available layers and reports satisfied, missing, or stale context. This aligns with Axis First Principle 2: "Query is Context."

### Explicit Over Implicit

Retention, release, and compaction are explicit actions with observable traces. No automatic background maintenance, no hidden GC, no silent eviction.

### Immutable Baseline, Derived Views

Long-term Memory is append-only. Derived views (competence profiles, pattern indices) may be rebuilt or incrementally maintained, but never mutate the raw event log.

### Local-First, Zero External Dependencies

P0 uses only the Go standard library and local filesystem. No SQLite, no Redis, no vector database, no network services.

## Users

- Agents executing tasks that need cross-turn context retention
- Developers inspecting why an Agent retained or forgot a bundle
- CLI users managing working sets via `axis memory` commands
- Future evaluation loops comparing memory quality with task outcomes

## Functional Requirements

### FR1: Three explicit layers

Axis must define three memory layers with distinct lifecycles and interfaces:

1. **Immediate Memory** — per-execution-cycle context
2. **Working Memory** — per-session/task-chain retained bundles
3. **Long-term Memory** — cross-session immutable event history

### FR2: Immediate Memory composition

Immediate Memory must include:

- Task identity and intent
- Contract constraints
- Working Set snapshot (content pulled from Working Memory)
- Tool results from the current execution cycle
- Turn counter to prevent unbounded self-iteration
- Token budget

Immediate Memory must expose file paths, fixed-length summaries, and content hashes. It must never inline full file contents by default.

### FR3: Working Memory management

Working Memory must support explicit operations:

- `Retain(bundle_id, reason)` — add a bundle to the working set
- `Release(bundle_id)` — remove a bundle from the working set
- `Recall(query, limit)` — retrieve relevant packets from retained bundles
- `List()` — enumerate all retained bundles
- `Clear()` — empty the working set

Working Memory must persist across process restarts using local filesystem storage.

### FR4: Working Memory storage format (P0)

Working Memory must use a local filesystem-based storage engine with:

- JSONL append-only log for raw writes
- Snapshot file with tiny header for fast cold-start
- Index file mapping keys to file offsets for O(1) reads
- Pure Go standard library implementation (bufio, os, encoding/json, sync)

### FR5: Long-term Memory append-only event log

Long-term Memory must store immutable event records:

- Task lifecycle events
- Self-judgement results
- Autonomy transitions
- Tool executions (behavioral data, not full tool outputs)
- Memory retention/release actions

Events must include: event_type, entity_id, timestamp, payload, source_digest.

### FR6: Long-term derived views

Long-term Memory must support read-only derived views:

- CompetenceProfile per agent per project
- Contract evolution lineage
- Analogous task history

Derived views must bind to a `source_digest` of the raw event log subset they were computed from.

### FR7: Incremental update with periodic rebuild

Derived views must be updated incrementally from new events. Periodic full rebuilds must be triggerable to correct drift. Incremental logic must never mutate the raw event log.

### FR8: Forgetting as soft marking only

Agents or users may request forgetting a memory record. The system must:

- Mark the record as deprecated via metadata
- Continue to preserve the original event in the append-only log
- Exclude deprecated records from normal query results
- Support un-forgetting by removing the deprecated mark

Physical deletion of raw events is prohibited.

### FR9: Per-project authority for competence profiles

Competence profiles must be per-project authoritative in P0-P1. Global profiles may exist as read-only cold-start references but must not influence project-level autonomy decisions.

P2 may introduce domain-level profiles as an intermediate layer.

### FR10: Non-invasive execution path

Memory layers must not:

- Inject context into provider prompts without explicit opt-in
- Change scheduler, dispatcher, or contract semantics
- Grant or escalate permissions
- Execute tools or mutate task state

### FR11: CLI observability

P0 must expose:

- `axis memory retain <bundle-id> --reason "..."`
- `axis memory release <bundle-id>`
- `axis memory list`
- `axis memory inspect <bundle-id>`
- `axis memory compact` — explicit snapshot rebuild

### FR12: Cross-platform safety

All file operations must use `path/filepath`. Locking must use `sync.Mutex` (not POSIX-specific). Snapshot rebuild must be safe on Windows (atomic rename semantics).

## Non-Goals

- No vector database or embedding search in P0-P1
- No SQLite or external KV store
- No automatic background compaction, indexing, or watcher
- No automatic context injection into model prompts
- No policy engine hidden inside memory layers
- No cross-project autonomy decisions in P0-P1
- No physical deletion of immutable event logs
- No Web UI or TUI for memory browsing
- No multi-agent shared memory market
- No LLM-based memory summarization in P0

## Acceptance Criteria

- Requirements, design, and tasks exist under `docs/specs/layered-memory-model/`.
- The three-layer model is defined with explicit interfaces and boundaries.
- Working Memory uses pure Go standard library + local filesystem with JSONL + snapshot + index.
- Key format is `wm:bundle:{bundle_id}` with JSON value self-describing bundles.
- Compact is explicit CLI only; no background/auto logic in P0.
- Long-term events are append-only; forgetting is soft-mark only.
- Competence profiles are per-project authoritative in P0-P1.
- Scheduler, dispatcher, contract, and provider semantics remain unchanged.
