# Execution-time Context Consumption Requirements

## Summary

Execution-time Context Consumption defines how Axis execution paths may safely read and use previously assembled context readiness records.

This is the next phase after Adaptive Context Assembly acceptance. The goal is to let execution become context-aware without turning context into permission, scheduler policy, hidden prompt injection, or a control plane.

## Users

- Agent executors preparing to execute `AgentTask`s
- Contract executors calling model providers
- Developers auditing what context was visible at execution time
- Future evaluation loops comparing readiness quality with execution outcomes

## Functional Requirements

### FR1: Readiness-aware execution boundary

Execution paths may inspect task readiness metadata:

```text
context.bundle_id
context.assembly_mode
context.packet_count
context.truncated
context.source_digest
```

This inspection must be read-only.

### FR2: Preflight before consumption

Before any execution path consumes context, it must be able to run the same readiness check as:

```bash
axis context preflight <task-id>
```

A task with missing or untraceable readiness must remain executable by default unless a strict gate is explicitly enabled.

### FR3: Audit record

Execution must be able to record whether context readiness was seen, ignored, or consumed.

At minimum, future execution records should be able to include:

```text
execution.context.bundle_id
execution.context.status
execution.context.consumption_mode
execution.context.packet_count
execution.context.truncated
```

### FR4: Summary-only consumption first

The first implementation must not inject full context packets into provider prompts.

The first safe consumption mode should expose only a compact summary, such as:

```text
selected sources
packet count
truncated flag
source digest
readiness status
```

### FR5: Explicit prompt augmentation gate

If context is ever inserted into provider prompts, it must require an explicit opt-in flag or config.

Prompt augmentation must be:

- traceable
- budgeted
- reversible
- covered by tests
- clearly marked in execution audit metadata

### FR6: No permission expansion

Context consumption must not grant tool, file, network, or autonomy permissions.

Tool access remains governed by existing contracts, tool registries, and future permission systems.

### FR7: No scheduler or contract mutation

Context consumption must not change:

- scheduler readiness
- dependency resolution
- contract input schema
- contract output schema
- provider selection
- provider credentials

## Acceptance Criteria

- Requirements, design, and tasks exist under `docs/specs/execution-context-consumption/`.
- The first implementation path is summary-only and read-only.
- Prompt augmentation is explicitly out of scope for P0.
- Strict gating is explicit, not default.
- Safety boundaries preserve accepted Adaptive Context Assembly semantics.

## Constraints

- Must reuse existing `contextpack` readiness artifacts and preflight semantics.
- Must not require persistence, vector DB, or LLM ranking.
- Must not break existing `axis run`, `axis ask`, shell, provider, or scheduler behavior.
- Must remain CLI-first and shell-friendly.

## Non-Goals

- No automatic prompt injection in P0
- No persistent readiness store in this spec's P0
- No vector retrieval
- No LLM reranking
- No permission escalation
- No automatic execution blocking by default
- No Web UI or TUI
- No new provider SDKs

## Open Questions

- Should execution context audit live in `types.ExecutionResult`, `types.TaskResult`, or a dedicated execution trace object?
- Should strict readiness gating be a CLI-only preflight or an optional orchestrator/executor policy?
- Should summary-only context be passed through provider input metadata or kept outside provider requests in P0?
