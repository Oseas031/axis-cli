# Crystal Unit Requirements

**Status**: Draft (structural — review carefully before promoting to Planned)
**Inspired by**: Jelly Crystal Commons Architecture — Encoding & Crystallization layer (语义元素 → 答案程序 → 闭合证明 → 数据结晶)
**Related**: `docs/specs/closure-ledger/`, `docs/specs/immunity-memory/`, `docs/specs/sandboxed-evolution/`, `docs/architecture/agent-native-first-principles.md`

> This is the most structurally significant of the three Jelly-Crystal-inspired specs. It introduces a new **typed reuse unit** and therefore touches the boundary between `internal/types/`, `internal/memory/`, and the sandboxed-evolution protocol. Read `docs/architecture/semantic-boundaries.md` and `docs/architecture/refactor-migration-conventions.md` before approving.

## Summary

A **Crystal** is the minimum reusable unit of "a problem that was solved well". It bundles:

1. The **problem signature** (what was being asked)
2. The **answer program** (the tool-call DAG that solved it)
3. The **closure proof** (evidence that the closure predicate held)
4. The **provenance** (which task, which agent, when, at what cost)

A Crystal is created by **explicit promotion** of a completed task that has a positive Closure Ledger entry. It lives in a new Verified Memory layer and can be retrieved by future tasks via contextpack preview as a "this problem class has a known solution shape" hint — never as automatic execution.

This spec is the Axis equivalent of Jelly Crystal's "data crystal", deliberately stripped of: (a) auto-graduation to public commons, (b) cross-project sharing, (c) automatic re-execution. Those belong to future Spec-RDTs under the sandboxed-evolution protocol if and when they are justified.

## Design Philosophy

### A Crystal Is Not a Cached Answer

A Crystal stores the **path** (signature + program + proof), not the answer. Re-running the answer program on new inputs is allowed; copy-pasting the old answer is not what Crystals are for. This matches the Jelly Crystal subtitle: *"not caching answers, caching verifiable closure paths."*

### Promotion Is a Sandboxed-Evolution Event

Creating a Crystal is a structural change to the project's verified-memory layer. It MUST follow the existing sandboxed-evolution protocol: experiment (the task ran) → verify (the closure proof) → explicit promote (`axis crystal promote`). No auto-promotion. The promoter may be human or Agent per `CLAUDE.md §5`; what matters is that the closure proof is machine-checkable, not who triggers the command. This preserves First Principle 4 (Layered Isolation) and §5 Spec-First.

### Crystals Are Per-Project Authoritative

P0–P1: a Crystal belongs to one project. No "public commons" exists in Axis. Cross-project crystal sharing would require a separate Spec-RDT addressing trust, license, and signature-collision concerns.

### Recall Is Preview Only

Like Immunity Memory, Crystal recall surfaces matches through contextpack preview with an opt-in flag. The provider call path is unchanged. An Agent reading a recalled crystal must decide for itself whether the answer program is applicable.

### Composable With the Other Two Specs

- A Crystal references its `closure_proof` via the Closure Ledger's evidence_refs.
- A failed attempt at a crystal-known problem class may itself become an Immunity record, refining future recall.
- These three specs do not implement each other; they reference each other through stable IDs.

## Users

- Agents about to attempt a task whose intent signature matches a known crystal
- Developers building a project-local "known plays" library over time
- Future evaluators measuring reuse rate and reuse quality

## Functional Requirements

### FR1: Crystal schema

A Crystal record MUST contain:

- `crystal_id` — stable identifier (content-addressed hash recommended)
- `problem_signature` — same fields as `immunity-memory` signature: `{intent.kind, intent.normalized_args_schema, contract.tool_allowlist}`. Note: `normalized_args_schema` (the **shape** of args), not the literal args.
- `answer_program` — serialized tool-call DAG: ordered list of `{tool, input_template, output_capture, success_predicate}`
- `closure_proof` — reference to the Closure Ledger entry of the originating task, plus the satisfied closure predicate and its evidence_refs
- `provenance` — `{source_task_id, promoted_by, promoted_at, source_digest}`
- `lifecycle` — one of `{verified, deprecated, retired}` (no `public`, no `pending`)

Metadata namespace: `crystal.*`.

### FR2: Promotion preconditions

`axis crystal promote <task-id>` MUST refuse to create a Crystal unless ALL of the following hold:

1. The source task is in terminal state and was not failed
2. The source task has a Closure Ledger entry with `closed = true` (not null, not false)
3. The source task's contract has a declared closure predicate
4. The promoter (human or agent) is explicitly named via `--by <actor>` flag
5. A one-line `--rationale "..."` is supplied

If any precondition fails, the command MUST exit non-zero with a clear human-readable cause and (in `--json` mode) a structured error code per `error-code-conventions.md`.

### FR3: Storage

Crystals MUST be stored under `internal/memory/crystal/` with:

- Append-only JSONL log: `.axis/memory/crystals.jsonl`
- Signature → crystal-id index (rebuildable, not authoritative)
- Pure Go standard library only

P0 MUST NOT introduce an external dependency.

### FR4: Recall

The memory layer MUST expose:

- `Recall(signature, limit)` — exact-signature match
- `RecallByIntent(intent_kind, limit)` — broader match by intent kind only
- `Show(crystal_id)` — full crystal content

P0 recall is field-equality. Vector or embedding similarity is out of scope.

### FR5: Contextpack opt-in surface

Contextpack MUST gain `--include-crystals` flag (mirroring `--include-immunity`). With the flag, preview attaches up to N (default 2) matching crystals, clearly labeled as "candidate answer programs (advisory)". Without the flag, preview is unchanged. Provider call path is NOT modified.

### FR6: Soft-deprecate, never delete

A crystal may be marked `deprecated` (with `--reason`) when the underlying tool API changes, when better crystals exist, etc. Deprecated crystals are excluded from default Recall but the JSONL events remain immutable. `retired` is a stronger state for "this crystal led to repeated failures" and requires a machine-checkable trigger (e.g., N consecutive failure events referencing this crystal_id) or explicit retire command with `--confirmed-by`.

### FR7: CLI surface

- `axis crystal promote <task-id> --by <actor> --rationale "..."`
- `axis crystal list [--intent <kind>] [--deprecated]`
- `axis crystal show <crystal-id>`
- `axis crystal deprecate <crystal-id> --reason "..."`
- `axis crystal retire <crystal-id> --reason "..." --confirmed-by <actor>` (gated)
- `axis context preview --include-crystals`

Output follows `cli-output-conventions.md`.

### FR8: Boundary enforcement

- `internal/kernel/` MUST NOT read from `internal/memory/crystal/`
- `internal/model/` (provider layer) MUST NOT read crystals directly
- `internal/contract/` MUST NOT change contract semantics based on crystal presence
- Crystal recall happens ONLY when contextpack is explicitly invoked with `--include-crystals`

### FR9: Re-execution of answer programs is out of scope

This spec does NOT add "axis crystal apply <crystal-id> --to <new-task>" semantics. The answer program is data the Agent can read. Whether to mechanically replay it is a separate Spec-RDT under sandboxed-evolution (because it touches execution policy).

### FR10: Cross-platform safety and secrets

- `path/filepath`, atomic rename for index rebuild on Windows
- `answer_program.input_template` MUST NOT capture secrets from the source task. The promote command MUST scrub fields matching known secret patterns (delegating to existing `internal/safego/`); if scrubbing is uncertain, promotion fails closed.

## Non-Goals

- No public commons (no inter-project sharing in P0–P1)
- No automatic crystallization
- No automatic re-execution of `answer_program`
- No vector/embedding similarity in P0–P1
- No quality scoring beyond the existing Closure Ledger
- No GUI, no TUI
- No background indexer or watcher
- No training-export pipeline in this spec
- No Crystal Index over a hypergraph in P0 (a flat per-project index is enough; hypergraph indexing is a future spec if reuse rate justifies it)

## Acceptance Criteria

- `docs/specs/crystal-unit/{requirements,design,tasks}.md` exist; status transitions emit `spec.<status>` events with verification artifacts referenced.
- `axis crystal promote` rejects tasks lacking a positive Closure Ledger entry.
- `axis crystal promote` rejects tasks whose contract has no closure predicate.
- `axis context preview --include-crystals` is byte-identical to baseline preview when the flag is absent.
- `internal/kernel/`, `internal/model/`, `internal/contract/` have zero imports from `internal/memory/crystal/`.
- Promotion attempts that would leak a secret into `answer_program` fail closed and write an audit event.
- `go test -race ./...`, `go vet ./...`, `staticcheck ./...`, `gosec ./...` all pass.
- A new diagram or table in `docs/architecture/semantic-boundaries.md` documents the Crystal boundary.
