# Execution-time Context Consumption Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)
- [../adaptive-context-assembly/](../adaptive-context-assembly/)

## Progress Tracking

| Task | Status |
|---|---|
| T1: Document execution consumption boundary | Completed |
| T2: Add execution context summary model | Completed |
| T3: Add summary consumer over readiness registry | Completed |
| T4: Add agent execution request summary field | Completed |
| T5: Add audit-only tests | Completed |
| T6: Decide contract executor audit path | Planned |
| T7: Gate prompt augmentation behind future spec | Planned |
| T8: Add Agent-declared context.requested_sources | Completed |

---

## T1: Document execution consumption boundary

**Goal**: Define how execution may safely observe context readiness without changing behavior.

**Files**:

- `docs/specs/execution-context-consumption/requirements.md`
- `docs/specs/execution-context-consumption/design.md`
- `docs/specs/execution-context-consumption/tasks.md`
- `docs/README.md`

**Acceptance Criteria**:

- Requirements, design, and tasks exist.
- P0 is summary-only and read-only.
- Prompt augmentation is explicitly non-P0.
- Safety boundaries preserve Adaptive Context Assembly acceptance.

**Depends on**: Adaptive Context Assembly acceptance

---

## T2: Add execution context summary model

**Goal**: Define a compact execution-safe summary of context readiness.

**Files**:

- `internal/contextpack/execution_summary.go`
- `internal/contextpack/execution_summary_test.go`

**Acceptance Criteria**:

- `ExecutionContextSummary` exists.
- Consumption modes include `none`, `observed`, `summary`, and reserved `prompt_augmented`.
- Summary excludes full packet content.
- Tests prove source list and packet count are summary-only.

**Depends on**: T1

---

## T3: Add summary consumer over readiness registry

**Goal**: Convert an `AgentTask` plus readiness registry into `ExecutionContextSummary`.

**Files**:

- `internal/contextpack/execution_consumer.go`
- `internal/contextpack/execution_consumer_test.go`

**Acceptance Criteria**:

- Ready tasks produce `ConsumptionMode=summary`.
- Missing readiness produces `ConsumptionMode=none` or `observed` with clear status.
- Untraceable readiness does not panic or mutate state.
- Consumer uses existing `Preflight` and `ReadinessRegistry.Inspect`.

**Depends on**: T2

---

## T4: Add agent execution request summary field

**Goal**: Let agent execution receive a summary without changing provider prompts.

**Files**:

- `internal/agent/executor.go`
- related agent tests if present

**Acceptance Criteria**:

- `AgentExecutionRequest` can carry optional `ContextSummary`.
- Existing agent executor implementations remain compatible.
- No provider request or contract executor behavior changes.

**Depends on**: T3

---

## T5: Add audit-only tests

**Goal**: Verify execution context awareness does not mutate execution behavior.

**Files**:

- `internal/contextpack/*_test.go`
- `internal/agent/*_test.go` if needed
- `cmd/axis/*_test.go` only if CLI output changes

**Acceptance Criteria**:

- Summary creation is read-only.
- No prompt content is generated in P0.
- Missing context does not fail execution by default.
- Strict preflight remains CLI-controlled unless a later spec changes it.

**Depends on**: T4

---

## T6: Decide contract executor audit path

**Goal**: Decide whether contract executor should receive task metadata for audit-only context awareness.

**Files**:

- `docs/specs/execution-context-consumption/design.md`
- possibly a follow-up spec if API changes are material

**Acceptance Criteria**:

- Decision is documented before implementation.
- Provider prompt path remains unchanged unless a new accepted task explicitly changes it.
- Migration impact on existing tests is understood.

**Depends on**: T5

---

## T7: Gate prompt augmentation behind future spec

**Goal**: Keep prompt injection out of this implementation phase.

**Files**:

- `docs/specs/execution-context-consumption/design.md`
- future spec if prompt augmentation is accepted

**Acceptance Criteria**:

- Prompt augmentation remains disabled by default.
- Any future prompt augmentation must have explicit opt-in, trace, budget, and tests.
- This task remains Planned until a future spec is requested.

**Depends on**: T6

---

## T8: Add Agent-declared context.requested_sources

**Goal**: Shift from system-push to Agent-query by letting Agents declare context needs.

**Files**:

- `internal/contextpack/artifact.go`
- `internal/contextpack/execution_summary.go`
- `internal/contextpack/execution_consumer.go`
- `internal/agent/executor.go`
- `internal/kernel/dispatcher/dispatcher.go`
- `docs/specs/execution-context-consumption/design.md`
- `docs/architecture/metadata-key-conventions.md`

**Acceptance Criteria**:

- `context.requested_sources` metadata key exists.
- `AgentExecutionRequest.RequestedSources` carries the Agent's declared needs.
- `ExecutionContextSummary` reports `RequestedSources`, `SatisfiedSources`, and `MissingSources`.
- `ExecutionContextConsumer.Summarize` resolves requests against the readiness registry.
- Dispatcher populates `RequestedSources` from task metadata.
- Tests verify satisfied/missing resolution and dispatcher flow-through.
- Safety: no prompt injection, no scheduler/contract/provider semantic changes.

**Review Fixes Applied**:

- Dispatcher eliminated duplicate `parseRequestedSources` call by reusing `summary.RequestedSources`.
- Added zero-value assertions in existing tests to protect backward compatibility.
- Added `TestExecutionContextConsumerSummarizeUntraceableWithRequests` to cover untraceable + requested sources path.
- Added design-intent comments to `Summarize`, `parseRequestedSources`, `resolveSources`, `AgentExecutionRequest.RequestedSources`, and dispatcher population logic to document the Query Is Context philosophy and modular boundaries.

**Depends on**: T5
