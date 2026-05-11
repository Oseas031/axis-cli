# Execution-time Context Consumption Design

## Overview

Execution-time Context Consumption sits after Adaptive Context Assembly.

```text
AgentTask
  -> ContextBundle
  -> ReadinessArtifact
  -> Preflight
  -> Execution-time Context Consumption
  -> Contract / Agent execution
```

P0 must make execution aware of context readiness without changing execution behavior.

## Current Relevant Flow

### Task submission

`axis ask --with-context --submit` currently:

1. Parses natural language into `AgentTask`.
2. Assembles `ContextBundle`.
3. Registers `ReadinessRecord` in `contextpack.DefaultRegistry`.
4. Attaches lightweight `context.*` metadata.
5. Submits the ordinary `AgentTask`.

### Contract execution

`internal/contract/executor.Execute` currently:

1. Validates contract input.
2. Creates `provider.ModelRequest{ContractID, Input}`.
3. Optionally attaches tools.
4. Calls provider.
5. Validates output.

It does not currently receive full `AgentTask` metadata.

### Agent execution

`internal/agent.AgentExecutionRequest` currently contains:

```go
Task        *types.AgentTask
SelfContext *SelfContext
Contract    *types.AgentContract
Autonomy    AutonomyLevel
```

This path is structurally more ready for context-aware execution because it already carries the full task.

## Design Principle

```text
Execution may observe readiness before it consumes context.
Consumption may summarize before it augments prompts.
Prompt augmentation must be explicit, auditable, and reversible.
```

## Query Is Context

The system must not push context to Agents. Agents declare what they need; the system queries and reports what is satisfied or missing.

P0 introduces `context.requested_sources` as a metadata key where an Agent (or task preparer) declares required context sources. The system resolves these declarations against the readiness registry and reports:

- `requested_sources`: what the Agent asked for
- `satisfied_sources`: what the registry can provide
- `missing_sources`: what remains unavailable

This changes the flow from:

```text
System assembles context -> pushes to Agent -> Agent consumes
```

to:

```text
Agent declares needs -> System queries registry -> reports satisfied/missing
```

## Proposed Components

### ExecutionContextSummary

A compact summary derived from `contextpack.PreflightResult` and `contextpack.ReadinessRecord`.

Proposed package:

```text
internal/contextpack
```

Proposed shape:

```go
type ExecutionContextSummary struct {
    BundleID         string
    Status           PreflightStatus
    ConsumptionMode  string
    PacketCount      int
    Truncated        bool
    SourceDigest     string
    Sources          []string
    RequestedSources []string
    SatisfiedSources []string
    MissingSources   []string
}
```

P0 summary must not include full packet contents.

### ConsumptionMode

Recommended values:

```text
none
observed
summary
prompt_augmented
```

P0 should support only:

```text
none
observed
summary
```

`prompt_augmented` is reserved for a future explicit opt-in phase.

### ExecutionContextConsumer

A small service that converts task readiness into an execution-safe summary.

```go
type ExecutionContextConsumer struct {
    Registry *contextpack.ReadinessRegistry
}

func (c *ExecutionContextConsumer) Summarize(task *types.AgentTask) ExecutionContextSummary
```

The method should:

1. Run `contextpack.Preflight(task, registry)`.
2. If not ready, return `ConsumptionMode=none` or `observed` with reason.
3. If ready, inspect the registry record.
4. Return compact summary fields.
5. Never mutate task, registry, scheduler, provider request, or contracts.

## Data Flow P0

```text
AgentTask.Metadata
  -> contextpack.Preflight
  -> contextpack.ReadinessRegistry.Inspect
  -> ExecutionContextSummary
  -> execution audit metadata / logs
```

P0 does not pass packet content to provider prompts.

## Integration Options

### Option A: Agent executor first

Because `AgentExecutionRequest` already carries `Task`, the first implementation can add an optional summary field:

```go
type AgentExecutionRequest struct {
    Task             *types.AgentTask
    SelfContext      *SelfContext
    Contract         *types.AgentContract
    Autonomy         AutonomyLevel
    ContextSummary   *contextpack.ExecutionContextSummary
    RequestedSources []string
}
```

Pros:

- Does not require contract executor API changes.
- Keeps model provider path untouched.
- Aligns with agent-native execution.

Cons:

- Does not affect contract-only execution yet.

### Option B: Contract executor metadata extension

Extend contract execution to accept metadata:

```go
ExecuteWithTask(task *types.AgentTask)
```

Pros:

- Contract execution can audit context readiness.

Cons:

- Higher risk: API expansion and more call-site changes.
- Must avoid provider prompt changes.

### Chosen P0 Direction

Use Option A for first implementation planning.

P0 should keep contract executor and provider request unchanged unless a separate task explicitly adds audit-only metadata.

## Prompt Augmentation Boundary

Prompt augmentation is not P0.

If later enabled, it must satisfy:

- explicit flag or config
- bounded summary length
- no raw full-repo dumps
- no hidden tool permissions
- audit field `execution.context.consumption_mode=prompt_augmented`
- tests proving disabled-by-default behavior

## Audit Metadata

Future execution audit should be able to emit:

```text
execution.context.status
execution.context.bundle_id
execution.context.consumption_mode
execution.context.packet_count
execution.context.truncated
execution.context.source_digest
```

These keys should not replace existing `context.*` readiness keys.

## Safety Boundaries

Execution-time consumption must not:

- mutate task metadata silently
- alter scheduler readiness
- override contract schemas
- expand tool access
- change provider profile or credentials
- inject prompt content by default
- fail execution by default when readiness is missing

## Evolution Path

### P0: Summary-only execution awareness

Create summary model and optional agent execution request field. No prompt changes.

### P1: Execution audit emission

Record whether execution observed or consumed context readiness.

### P2: Contract executor audit-only awareness

Allow contract executor to receive task metadata for audit without changing provider prompt.

### P3: Explicit prompt augmentation

Only after P0-P2 are stable, allow explicit opt-in summary injection.

### P4: Outcome feedback loop

Compare task outcomes with readiness quality to improve future assembly.
