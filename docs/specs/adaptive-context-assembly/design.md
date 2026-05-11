# Adaptive Context Assembly Design

## Overview

Adaptive Context Assembly sits between task creation and Agent execution.

```text
AgentTask
  -> Context Assembly
  -> ContextBundle preview or injection
  -> Agent execution
```

P0 should only create previewable bundles. It must not change execution semantics.

## Architecture

Recommended package for later implementation:

```text
internal/contextpack/
  packet.go
  assembler.go
  rules.go
  assembler_test.go
```

`contextpack` avoids naming conflicts with Go's standard `context` package.

## Core Data Model

### ContextPacket

```go
type ContextPacket struct {
    ID          string
    Type        string
    Source      string
    Summary     string
    Content     string
    Relevance   float64
    Reason      string
    IsPartial   bool
    TruncatedAt int
}
```

`IsPartial` and `TruncatedAt` are set when a packet’s content is truncated to fit the remaining byte budget, rather than being dropped entirely. This preserves high-relevance context at reduced granularity.

Recommended packet types:

```text
spec
code
memory
recent_change
contract
tool
error
workflow
principle
```

### ContextBundle

```go
type ContextBundle struct {
    TaskID      string
    ContractID  string
    Goal        string
    Packets     []ContextPacket
    Trace       AssemblyTrace
    Budget      ContextBudget
}
```

### AssemblyTrace

```go
type AssemblyTrace struct {
    Selected []TraceItem
    Excluded []TraceItem
    Notes    []string
}
```

### ContextBudget

```go
type ContextBudget struct {
    MaxPackets int
    MaxBytes   int
    UsedBytes  int
    Truncated  bool
}
```

### ReadinessArtifact

`ReadinessArtifact` is a stable audit reference derived from a `ContextBundle`.

It stores only lightweight reproducibility metadata:

```text
context.bundle_id
context.assembly_mode
context.packet_count
context.truncated
context.source_digest
```

It must not store the full context bundle inside `AgentTask.Metadata`.

### ReadinessRegistry

`ReadinessRegistry` maps `context.bundle_id` to an inspectable readiness record.

P2 starts with an in-process registry:

```text
context.bundle_id -> ReadinessArtifact + ContextBundle summary
```

It makes readiness records inspectable during the current Axis process without adding execution injection.

T10 adds an optional persistent backing store (`FileStore`) under `.axis/context/readiness.json`:

```text
ReadinessRegistry (in-memory) + ReadinessStore (file-backed JSON, atomic write)
```

- `NewReadinessRegistry()` creates a memory-only registry for tests and backward compatibility.
- `NewReadinessRegistryWithStore(store)` loads existing records from disk on creation.
- `Register` writes to disk after updating memory; disk failure does not block registration.
- `Inspect` syncs from disk before lookup, enabling cross-process reads.
- `Reset` clears memory and deletes the store file.
- `InitDefaultRegistry(root)` replaces the global `DefaultRegistry` with a file-backed instance.

The local control plane (`axis start`) initializes a persistent registry so that dispatcher `ExecutionContextSummary` resolution survives server restart. Standalone tests continue to use the in-memory default.

## P0 Assembly Flow

```text
1. Read AgentTask input.message / input.goal
2. Identify task keywords and contract ID
3. Collect deterministic candidates from known docs/spec paths
4. Rank by rule score, authority, and freshness
5. Enforce packet and byte budget
   - Packet count limit is absolute; exceeding packets are excluded.
   - If a candidate exceeds the byte budget, attempt semantic truncation of
     Content (then Summary) at paragraph, sentence, line, or whitespace boundaries.
   - If truncation cannot reduce the packet below the remaining budget
     (fixed metadata overhead already exceeds capacity), exclude the packet.
   - Truncated packets are marked `IsPartial` with `TruncatedAt` set.
6. Emit ContextBundle and AssemblyTrace
```

## Candidate Rules

Initial rule examples:

| Signal | Candidate context |
|---|---|
| `ask`, `natural language`, `prompt` | `docs/specs/natural-language-scheduling/` |
| `provider`, `model`, `deepseek`, `minimax` | `docs/specs/model-provider/`, provider README sections |
| `shell`, `interactive` | `docs/specs/interactive-shell/` |
| `scheduler`, `dag`, `dependency` | scheduler specs and architecture docs |
| `axis-up`, `onboarding`, `start` | `tools/axis-up/README.md`, `tools/axis-up/DESIGN.md` |
| `context`, `assembly` | `docs/specs/adaptive-context-assembly/` |

## Authority and Freshness

Ranking should prefer:

```text
current task > current code > current specs > current test output > accepted principles > recent memory > old reports
```

Deprecated documents should only be selected when the task explicitly asks for historical context.

## CLI Preview Shape

P0 exposes:

```bash
axis context preview "fix provider config"
```

Natural language scheduling can also preview assembled context:

```bash
axis ask "fix provider config" --with-context
```

Both commands preview context only. They do not submit tasks or inject context into execution.

Submitted readiness records can be inspected in the same process:

```bash
axis context inspect ctx-...
```

Tasks can be checked for traceable readiness before execution:

```bash
axis context preflight ask-...
```

Scripts may opt into a failing gate:

```bash
axis context preflight ask-... --strict
```

## Safety Boundaries

- Context does not grant permissions.
- Context does not execute tools.
- Context does not override contracts.
- Context does not change scheduler readiness.
- Context does not bypass provider configuration.
- Context does not silently expand file/network access.

## Evolution Path

### P0: Rule-based preview

Produce deterministic context bundles from task goal and known project sources.

### P1: Readiness artifact attachment

Attach a deterministic readiness artifact reference to submitted tasks when users explicitly request context assembly.

```bash
axis ask "fix provider config" --with-context --submit
```

This attaches lightweight `context.*` metadata to the ordinary `AgentTask`. It does not inject context into provider prompts or change scheduler, contract, dispatcher, or provider semantics.

### P2: Inspectable readiness registry

Register readiness artifacts in an in-process registry and expose `axis context inspect <bundle-id>`.

This makes `context.bundle_id` traceable to the selected packets, reasons, exclusions, budget, and source digest without persistent storage.

### P3: Context readiness preflight

Check whether a submitted task has traceable context readiness before execution.

```bash
axis context preflight <task-id>
```

The check is read-only. It does not execute, block, mutate, inject context, or change scheduler semantics.

`--strict` returns an error when status is not `ready`; default preflight remains non-blocking.

### P4: Competence-tuned assembly

Adjust context detail level using explicit reliability signals.

### P4.5: Agent-declared needs

Agents declare required context sources via `context.requested_sources`; the system resolves them against the readiness registry and reports satisfied/missing. This shifts from system-push to Agent-query, implemented in the execution-context-consumption layer.

### P5: Retrieval-backed assembly

Add richer retrieval only after deterministic bundle semantics are stable.

#### P5 Design Principles

- **Zero-config first run**: TF-IDF requires no API key, no provider setup, no external service.
- **Explicit index lifecycle**: Users choose when to build/update; no hidden background process.
- **Graceful degradation**: Missing or stale index falls back to rule-based assembly without error.
- **Hybrid traceability**: Trace shows both rule origin and retrieval score for every packet.
- **Axis boundary only**: Index covers Axis own docs and the user's current project code. External knowledge bases are out of scope.

#### P5 Data Flow

```text
1. Rule-based recall (deterministic candidates from keywords)
2. If index exists and healthy:
   a. Scan project files (docs, code) -> Document chunks
   b. TF-IDF index over chunks
   c. Query = task goal -> TF-IDF vector
   d. Cosine similarity rank against index
   e. Merge rule candidates + retrieval candidates
   f. Deduplicate by source, boost rule-matched items
3. If index missing or empty:
   a. Fall back to rule-based candidates only
4. Apply budget, emit bundle + hybrid trace
```

#### P5 Index Components

```text
internal/contextpack/
  index_scanner.go    - Scan project files into DocumentChunk
  index_tfidf.go      - TFIDFIndex: build, query, cosine rank
  index_store.go      - Persist/load index as JSON under .axis/context/index.json
  index.go            - IndexManager: rebuild, update, status, path
```

**DocumentChunk**:

```go
type DocumentChunk struct {
    Source    string // file path
    Content   string // extracted text
    ModTime   int64  // file mtime for incremental update
    DocType   string // doc | code | spec
}
```

**TFIDFIndex**:

- `Build(chunks []DocumentChunk)` - computes term frequency, global IDF, document vectors
- `Query(text string, topK int)` - vectorizes query, returns scored chunks by cosine similarity
- Pure Go, no external dependencies

**IndexManager**:

- `Rebuild(root string)` - full scan, compute, persist
- `Update(root string)` - compare mtimes, remove stale, add new, recompute affected, persist
- `Status(root string)` - report indexed files, total chunks, last build time
- `Path(root string)` - returns `.axis/context/index.json`

#### P5 Hybrid Assembly

`Assembler` gains an optional `*IndexManager`:

- `WithIndex(manager)` option
- `candidates()` first collects rule packets (as before)
- If index is present and non-empty, also runs `index.Query(goal, topK)`
- Retrieval results are converted to `ContextPacket` with:
  - `Type: PacketTypeCode` or `PacketTypeDoc`
  - `Source: chunk.Source`
  - `Content: chunk.Content` (truncated to budget)
  - `Reason: "retrieval: tf-idf cosine similarity"`
  - `Relevance: similarityScore` (normalized 0-1)
- Deduplication: if a retrieval result matches an existing rule packet by source prefix, keep the rule packet (determinism wins) and boost its relevance.
- Trace notes include `"hybrid mode: rule + retrieval"` or `"rule-only fallback: index not found"`.

#### P5 CLI

```bash
# Full rebuild from scratch
axis context index --rebuild

# Incremental update based on mtime changes
axis context index --update

# Show index status
axis context index --status
```

All three commands render machine-friendly JSON under `Context index:` header.

#### P5 Storage Layout

```text
.axis/
  context/
    readiness.json   (existing T10)
    index.json       (new: TF-IDF index + metadata)
```

`index.json` contains:
- `version` (for migration)
- `last_build_at`
- `indexed_files` (count)
- `total_chunks`
- `idf` map
- `document_vectors` (sparse representation: term -> weight)
- `chunks` (source, modTime, docType for incremental update)

### P6: LLM-assisted ranking

Use model providers for ranking or summarization only behind schema validation, budget limits, and trace output.
