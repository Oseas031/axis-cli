# Adaptive Context Assembly Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)

## Progress Tracking

| Task | Status |
|---|---|
| T1: Document design philosophy and P0 boundary | Completed |
| T2: Define context packet and bundle types | Completed |
| T3: Implement rule-based assembler | Completed |
| T4: Add context preview command | Completed |
| T5: Add audit trace tests | Completed |
| T6: Attach readiness artifact metadata | Completed |
| T7: Add inspectable readiness registry | Completed |
| T8: Add context readiness preflight | Completed |
| T9: Add strict preflight gate | Completed |
| T10: Persistent readiness registry | Completed |
| T11: TF-IDF index core (scanner + indexer + ranker) | Completed |
| T12: Index persistence and incremental update | Completed |
| T13: Hybrid assembler (rule recall + TF-IDF rerank) | Completed |
| T14: CLI axis context index --rebuild/--update/--status | Completed |
| T15: Hybrid assembler tests and fallback verification | Completed |
| T16: Update spec docs and acceptance | Completed |
| T17: Semantic truncation for budget-aware context retention | Completed |

---

## T1: Document design philosophy and P0 boundary

**Goal**: Establish Adaptive Context Assembly as situational readiness, not memory or control.

**Files**:

- `docs/specs/adaptive-context-assembly/requirements.md`
- `docs/specs/adaptive-context-assembly/design.md`
- `docs/specs/adaptive-context-assembly/tasks.md`
- `docs/README.md`

**Acceptance Criteria**:

- Five guiding philosophies are documented.
- P0 is rule-based and preview-first.
- Non-goals exclude vector DB, permission escalation, and hidden execution injection.

**Depends on**: None

---

## T2: Define context packet and bundle types

**Goal**: Add minimal internal data structures for context assembly.

**Files**:

- `internal/contextpack/packet.go`
- `internal/contextpack/packet_test.go`

**Acceptance Criteria**:

- `ContextPacket`, `ContextBundle`, `AssemblyTrace`, and `ContextBudget` exist.
- Types require source and reason.
- Tests cover validation of required provenance fields.

**Depends on**: T1

---

## T3: Implement rule-based assembler

**Goal**: Build a deterministic assembler from `AgentTask` to `ContextBundle`.

**Files**:

- `internal/contextpack/assembler.go`
- `internal/contextpack/rules.go`
- `internal/contextpack/assembler_test.go`

**Acceptance Criteria**:

- Natural-language scheduling tasks select natural-language scheduling specs.
- Provider tasks select model-provider docs.
- Shell tasks select interactive-shell docs.
- Budget limits are enforced.
- Trace explains selected and excluded packets.

**Depends on**: T2

---

## T4: Add context preview command

**Goal**: Make context assembly visible from CLI without changing execution.

**Files**:

- `cmd/axis/context_cmd.go`
- `cmd/axis/context_cmd_test.go`

**Acceptance Criteria**:

- CLI can preview context for a prompt or task-shaped input.
- Output includes selected packet source, reason, and budget usage.
- No task execution occurs.

**Depends on**: T3

---

## T5: Add audit trace tests

**Goal**: Ensure context assembly remains explainable.

**Files**:

- `internal/contextpack/*_test.go`
- `cmd/axis/*_test.go`

**Acceptance Criteria**:

- Tests assert every selected packet has source and reason.
- Tests assert budget exclusions are visible.
- Tests assert deprecated or irrelevant sources are not selected by default.

**Depends on**: T4

---

## T6: Attach readiness artifact metadata

**Goal**: Make assembled context auditable and traceable before execution without injecting it into execution.

**Files**:

- `internal/contextpack/artifact.go`
- `internal/contextpack/artifact_test.go`
- `cmd/axis/ask_cmd.go`
- `cmd/axis/ask_cmd_test.go`
- `docs/specs/adaptive-context-assembly/design.md`

**Acceptance Criteria**:

- `ContextBundle` can produce a deterministic readiness artifact.
- `axis ask "..." --with-context --submit` attaches lightweight `context.*` metadata to the submitted `AgentTask`.
- Full context bundles are not embedded in task metadata.
- Scheduler, contract, dispatcher, provider, and execution prompt semantics remain unchanged.

**Depends on**: T5

---

## T7: Add inspectable readiness registry

**Goal**: Make `context.bundle_id` resolve to an inspectable readiness record.

**Files**:

- `internal/contextpack/registry.go`
- `internal/contextpack/registry_test.go`
- `cmd/axis/context_cmd.go`
- `cmd/axis/context_cmd_test.go`
- `docs/specs/adaptive-context-assembly/design.md`

**Acceptance Criteria**:

- `axis ask "..." --with-context --submit` registers a readiness record.
- `axis context inspect <bundle-id>` prints the registered artifact and bundle summary.
- Missing bundle IDs return a clear error.
- Registry is in-process only; no persistence or execution injection is added.

**Depends on**: T6

---

## T8: Add context readiness preflight

**Goal**: Check whether a submitted task has traceable context readiness before execution.

**Files**:

- `internal/contextpack/preflight.go`
- `internal/contextpack/preflight_test.go`
- `cmd/axis/context_cmd.go`
- `cmd/axis/context_cmd_test.go`
- `docs/specs/adaptive-context-assembly/design.md`

**Acceptance Criteria**:

- `axis context preflight <task-id>` reports `ready`, `missing`, or `untraceable`.
- Ready status requires `context.bundle_id`, inspectable registry record, selected packets, and matching source digest.
- Missing context does not fail command execution; it reports a clear readiness result.
- Preflight is read-only and does not change task state, scheduler behavior, or execution prompts.

**Depends on**: T7

---

## T9: Add strict preflight gate

**Goal**: Let scripts explicitly fail when context readiness is not ready.

**Files**:

- `cmd/axis/context_cmd.go`
- `cmd/axis/context_cmd_test.go`
- `docs/specs/adaptive-context-assembly/tasks.md`

**Acceptance Criteria**:

- `axis context preflight <task-id> --strict` returns success only when status is `ready`.
- Strict mode still renders the preflight result before returning an error.
- Default preflight remains non-blocking and read-only.
- No execution semantics change.

**Depends on**: T8

---

## T10: Persistent readiness registry

**Goal**: Make readiness records survive process restart so that cross-process inspect and preflight work.

**Files**:

- `internal/contextpack/store.go`
- `internal/contextpack/file_store.go`
- `internal/contextpack/file_store_test.go`
- `internal/contextpack/registry.go`
- `internal/contextpack/registry_test.go`
- `internal/contextpack/init.go`
- `cmd/axis/main.go`
- `cmd/axis/control_runtime.go`
- `cmd/axis/main_test.go`

**Acceptance Criteria**:

- `ReadinessStore` interface exists with `LoadAll`, `SaveAll`, `DeleteAll`.
- `FileStore` persists records as JSON under `.axis/context/readiness.json` using atomic temp+rename writes.
- `ReadinessRegistry` supports an optional backing store; `NewReadinessRegistryWithStore` loads existing records on creation.
- `Register` persists to disk without blocking in-memory success.
- `Inspect` syncs from disk before lookup, enabling cross-process reads.
- `Reset` deletes the persisted file.
- `contextpack.InitDefaultRegistry(root)` replaces `DefaultRegistry` with a file-backed instance.
- `runLocalRuntime` initializes persistent registry so the local control plane dispatcher can resolve readiness after restart.
- Tests verify cross-process read, Reset file deletion, atomic write, and backward compatibility with in-memory-only usage.
- No scheduler/contract/dispatcher/provider semantic changes.

**Depends on**: T9

---

## T17: Semantic truncation for budget-aware context retention

**Goal**: Improve budget utilization by truncating packet content at semantic boundaries instead of dropping entire high-relevance packets.

**Files**:

- `internal/contextpack/packet.go`
- `internal/contextpack/assembler.go`
- `internal/contextpack/semantic_truncation.go`
- `internal/contextpack/semantic_truncation_test.go`
- `internal/contextpack/assembler_test.go`
- `cmd/axis/context_cmd_test.go`
- `docs/specs/adaptive-context-assembly/design.md`
- `docs/specs/adaptive-context-assembly/tasks.md`

**Acceptance Criteria**:

- `ContextPacket` carries `IsPartial` and `TruncatedAt` fields.
- When a candidate exceeds the remaining byte budget, the assembler attempts semantic truncation of `Content` (then `Summary`) at paragraph, sentence, line, or whitespace boundaries.
- If fixed metadata overhead exceeds the remaining budget, the packet is excluded with a clear trace reason.
- Truncated packets retain `IsPartial=true` and `TruncatedAt>0`; trace items note the truncation byte position.
- `MaxPackets` and `MaxBytes` dual constraints remain unchanged.
- UTF-8 safety: truncation never splits a multi-byte rune.
- Tests cover: no-truncation-needed, paragraph boundary, sentence boundary, line boundary, zero budget, UTF-8 safe fallback, truncation succeeds, truncation fails (fixed overhead too large).
- CLI test for budget exclusions is updated to accept the new exclusion reason format.

**Depends on**: T13
