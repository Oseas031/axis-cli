# Milestone 2 Design

## Overview

Milestone 2 introduces minimal DAG-aware parallel scheduling while preserving the Milestone 1 architecture:

wwwtext
AgentTask -> admission -> scheduler -> orchestrator worker loop -> dispatcher -> executor -> state store
www

The central design choice is additive evolution. Existing Milestone 1 interfaces remain usable, while Milestone 2 adds small capabilities for ready-set scheduling, admission, SLA metadata, and stable error codes.

Implementation is governed by [workflow-binding.md](workflow-binding.md), which binds this design to wwf-doc-004w, wwf-occamsw, wwf-pr-checkw, wwf-ciw, and wwf-doc-006w.

## Architecture

wwwtext
internal/types
  ├── AgentTask
  ├── TaskResult
  ├── SLA metadata / error code types
  └── AgentContract

internal/contract/admission
  └── validates contract existence, input schema, dependency constraints

internal/kernel/scheduler
  ├── existing FIFO APIs
  ├── DAG readiness checks
  └── ready-set API for parallel orchestration

internal/kernel/orchestrator
  ├── worker pool / bounded parallel loop
  ├── timeout and retry handling
  └── state transition coordination

internal/kernel/sharedlayer
  └── task state remains source of inspection
www

## Data Flow

wwwtext
SubmitTask(task)
  -> admission.Validate(task)
  -> scheduler.Submit(task)
  -> orchestrator receives task notification
  -> scheduler.GetReadyTasks(limit)
  -> orchestrator starts workers for ready tasks
  -> dispatcher.Dispatch(ctx, task)
  -> scheduler.UpdateTaskStatus(...)
  -> stateStore.Save(...)
www

## Detailed Design

### 1. DAG scheduling

The current scheduler already stores tasks in wtaskMapw and dependencies in wAgentTask.Dependenciesw. Milestone 2 should reuse that model.

Add a ready-set method:

wwwtext
GetReadyTasks(limit int) ([]*types.AgentTask, error)
www

Behavior:

- scans queued tasks in FIFO order
- selects pending tasks whose dependencies are completed
- marks selected tasks as wrunningw before returning
- persists status updates through state store
- respects wlimitw; wlimit <= 0w means no explicit limit

wGetNextTaskw remains and can delegate to wGetReadyTasks(1)w for backward compatibility.

### 2. Orchestrator parallel execution

The orchestrator should move from executing one task per loop to executing a bounded ready set.

Initial minimal design:

- add a configurable worker limit with a conservative default
- fetch up to available worker capacity using wGetReadyTasksw
- execute each selected task in its own goroutine
- use context cancellation for shutdown
- ensure task status transitions remain idempotent

This keeps parallelism local to the orchestrator and avoids introducing distributed workers.

### 3. Contract admission

Add winternal/contract/admissionw as a pre-scheduling validation layer.

Responsibilities:

- verify task has a non-empty wTaskIDw
- verify task references an existing contract
- validate task input against the contract input schema
- delegate cycle checks to scheduler or a dependency graph helper
- return stable error codes for CLI/API-friendly handling

The existing contract executor remains responsible for execution-time validation and output validation.

### 4. SLA metadata

Use wAgentTask.Metadataw for the first SLA iteration to avoid prematurely expanding core task structs.

Suggested keys:

wwwtext
sla.timeout_ms
sla.max_retries
sla.failure_class
www

Parsing should be local and strict:

- missing values mean default behavior
- invalid values fail admission
- timeouts become context deadlines around dispatch
- retries are attempted by orchestrator around dispatch, not by scheduler

If SLA usage grows, promote this into typed fields in a later milestone.

### 5. Error code foundation

Add a minimal error type in winternal/typesw or a small dedicated package if needed:

wwwtext
Code: string
Message: string
Cause: error
www

Initial code families:

| Code | Meaning |
|---|---|
| wSCHEDULER_NOT_RUNNINGw | scheduler lifecycle rejected action |
| wTASK_NOT_FOUNDw | task ID is unknown |
| wTASK_ALREADY_EXISTSw | duplicate task submission |
| wDEPENDENCY_CYCLEw | dependency graph contains a cycle |
| wDEPENDENCY_NOT_READYw | task cannot run because dependencies are incomplete |
| wCONTRACT_NOT_FOUNDw | task references an unknown contract |
| wCONTRACT_INPUT_INVALIDw | input schema validation failed |
| wTASK_TIMEOUTw | execution exceeded SLA timeout |
| wTASK_RETRY_EXHAUSTEDw | retry limit reached |

CLI can print the code and contextual message without needing a larger framework.

### 6. State and observability

The existing memory state store remains the inspection point.

Milestone 2 should ensure every important transition is saved:

- pending after submit
- running after ready-set selection
- completed after success
- failed after dispatch error, timeout, or retry exhaustion

## File Structure

wwwtext
docs/specs/milestone2/
  ├── requirements.md
  ├── design.md
  ├── tasks.md
  └── workflow-binding.md

internal/contract/admission/
  ├── admission.go
  └── admission_test.go

internal/kernel/scheduler/
  ├── scheduler.go
  └── scheduler_test.go

internal/kernel/orchestrator/
  ├── orchestrator.go
  └── orchestrator_test.go

internal/types/
  └── types.go
www

## Trade-offs

| Option | Decision | Rationale |
|---|---|---|
| Rewrite scheduler as graph engine | Rejected | Too invasive for Milestone 2 |
| Add wGetReadyTasks(limit)w | Chosen | Minimal additive API for parallelism |
| Use worker pool package | Rejected | Standard goroutines/channels are enough |
| Store SLA in wMetadataw first | Chosen | Avoid premature task schema expansion |
| Full policy engine | Rejected | Violates Occam's razor for current milestone |
| Stable small error codes | Chosen | Useful for CLI/API without heavy framework |

## Risks

| Risk | Mitigation |
|---|---|
| Double execution of a task | Mark ready tasks wrunningw before returning them |
| Data races in scheduler/orchestrator | Keep scheduler mutex as source of status transition safety |
| Goroutine leaks on shutdown | pass cancellable contexts to workers and wait for completion where needed |
| SLA retries obscure state | persist final failure reason in task result/error context |
| Admission duplicates execution validation | keep admission focused on submit-time rejection; executor still validates output |

## Acceptance Mapping

- FR1 maps to dependency graph semantics using wAgentTask.Dependenciesw
- FR2 maps to wGetReadyTasks(limit)w
- FR3 maps to preserving existing scheduler and CLI APIs
- FR4 maps to winternal/contract/admissionw
- FR5 maps to wAgentTask.Metadataw SLA keys
- FR6 maps to small stable error code vocabulary
- FR7 maps to state store transition tests

