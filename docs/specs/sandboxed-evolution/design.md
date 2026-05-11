# Sandboxed Evolution Protocol Design

## Overview

Sandboxed Evolution Protocol is the safety envelope for Axis self-modification.

It defines how Axis can propose, isolate, verify, inspect, promote, or discard changes to its own contracts, workflows, specs, context rules, and future architecture surfaces.

P0 is intentionally conservative:

```text
proposal -> isolated workspace -> atomic steps -> verification -> explicit decision
```

It does not implement autonomous architecture rewriting. It creates the engineering substrate required before that becomes safe.

## Current Problem

Axis already has forward-looking design philosophy and early self-bootstrapping primitives, but system changes are still ordinary repository edits.

That means a faulty change can directly affect:

- specs
- contracts
- execution behavior
- context readiness logic
- provider configuration
- future agent control surfaces

Without isolation and step trace, failures are hard to attribute, rollback, or use as competence evidence.

## Design Principles

### Isolation before influence

A proposed system change must not affect the main system until promoted.

### Atomicity before autonomy

Large changes must be decomposed into inspectable steps before they can earn higher autonomy.

### Trace before intelligence

Axis should record what happened before trying to infer high-level architecture state.

### Verification before promotion

Promotion is a decision over evidence, not a side effect of generating a patch.

### Filesystem as contract

P0 should use project-local files so shell tools, humans, and agents can inspect the same state.

## Architecture

### Components

```text
EvolutionIntent
EvolutionRun
EvolutionStep
VerificationRecord
EvolutionDecision
TraceLedger
EvolutionWorkspace
```

### P0 Data Flow

```text
user/agent intent
  -> create evolution run
  -> create isolated workspace
  -> apply atomic step patches
  -> run verification command
  -> write trace ledger
  -> promote or discard
```

### Project-local Layout

P0 should prefer a layout similar to:

```text
.axis/evolution/
  <run-id>/
    intent.json
    run.json
    steps.jsonl
    verification.json
    decision.json
    workspace/
    patches/
```

This keeps the protocol scriptable and inspectable without introducing a database.

## Data Model Sketch

### EvolutionIntent

```text
id
created_at
actor
summary
target_domain
risk_level
status
```

### EvolutionStep

```text
step_id
run_id
sequence
target_path
action
patch_ref
status
started_at
completed_at
error
```

### VerificationRecord

```text
run_id
command
started_at
completed_at
exit_code
stdout_ref
stderr_ref
status
```

### EvolutionDecision

```text
run_id
decision
actor
reason
created_at
```

Decision values:

```text
promoted
discarded
paused
```

## CLI Shape

P0 CLI can be introduced incrementally:

```bash
axis evolve start <summary>
axis evolve inspect <run-id>
axis evolve step <run-id> --target <path>
axis evolve verify <run-id> -- go test ./...
axis evolve promote <run-id>
axis evolve discard <run-id>
```

The exact command names may be refined during implementation, but the semantics must remain stable.

## Verification Gate

Verification is evidence capture, not hidden control.

P0 rules:

- verification command must be explicit
- command, exit code, and output references must be recorded
- failed verification prevents promotion by default
- promotion remains a separate explicit decision

## Promotion Semantics

Promotion means moving verified changes from the isolated workspace into the main working tree.

P0 must avoid implicit promotion.

Promotion should fail if:

- verification is missing
- latest verification failed
- workspace is missing
- target files changed incompatibly since workspace creation

## Discard Semantics

Discard means closing an evolution run without promoting its changes.

Discard must preserve the trace ledger unless the user explicitly deletes historical records.

## Relationship to Existing Systems

### Adaptive Context Assembly

Context readiness remains a source of audit evidence, not a control plane.

Sandboxed evolution may later use context summaries to explain why a change was proposed, but P0 must not inject context into prompts or execution.

### Execution-time Context Consumption

Execution context summary remains summary-only and read-only.

Sandboxed evolution should not depend on prompt augmentation.

### Contracts and Workflows

Contracts and workflow files are likely early targets for evolution, but their modification must happen inside isolated evolution runs.

## Trade-offs

| Decision | Chosen | Rejected | Rationale |
|---|---|---|---|
| Storage | project-local files | database | preserves shell-native inspection and low complexity |
| Isolation | workspace/patch envelope | direct edit | prevents main-system contamination |
| Verification | explicit command gate | automatic hidden gate | keeps control visible and auditable |
| Promotion | separate decision | auto-merge after tests | avoids uncontrolled self-modification |
| State estimation | trace first | abstract observer first | facts are required before reliable inference |
| Exploration | single candidate P0 | parallel candidates P0 | parallelism needs mature isolation and comparison |

## Risks

| Risk | Mitigation |
|---|---|
| Workspace copying is expensive | start with patch-only or narrow workspace if needed |
| Verification commands can be unsafe | require explicit user command and future allowlist |
| Trace files drift from reality | write trace at protocol boundaries and add tests |
| Promotion conflicts with changed main tree | detect conflicts and require manual resolution |
| Protocol becomes too heavy | keep P0 CLI small and file-native |

## Deferred Capabilities

The following are valid future directions but not P0:

- parallel candidate evolution
- architecture self-reference interface
- competence-based automatic promotion
- adaptive feedback parameters
- stochastic exploration
- state estimator over historical traces
- prompt-level self-modification

## P0 Success Definition

P0 succeeds when Axis can represent a system change as an isolated, inspectable, verified, and explicitly promoted or discarded evolution run without changing existing execution semantics.
