# Model Provider Tasks

## Related Documents

  [requirements.md](requirements.md)
  [design.md](design.md)
  [workflow binding.md](workflow binding.md)

## Upstream Workflow Binding

This feature is governed by the existing project workflow system:

  `wf doc 004` Meta Workflow Management: documentation first, explicit dependencies, HANDOVER synchronization
  `wf occams` Occam's Razor Architecture Simplification: MockModelProvider only, no real provider expansion
  `wf pr check` PR Quality Check Workflow: quality gates and non blocking documentation context
  `wf ci` Continuous Integration Workflow: build, format, tests
  `wf doc 006` Document Audit: beginner guide and handover consistency

## Progress Tracking

| Task | Status |
|---|---|
| T1: Add model provider package | Completed |
| T2: Implement MockModelProvider | Completed |
| T3: Integrate provider with ContractExecutor | Completed |
| T4: Update Dispatcher to use Execute | Completed |
| T5: Add and update tests | Completed |
| T6: Verify shell execution path | Completed |
| T7: Update docs and HANDOVER.md | Completed |

   

## T1: Add model provider package

**Goal**: Create the provider abstraction.

**Files**:

  `internal/model/provider/provider.go`

**Acceptance Criteria**:

  `ModelProvider` interface exists
  `ModelRequest` exists
  `ModelResponse` exists
  Package compiles with no external dependencies

**Depends on**: None

   

## T2: Implement MockModelProvider

**Goal**: Add deterministic local provider implementation.

**Files**:

  `internal/model/provider/mock.go`

**Acceptance Criteria**:

  Mock provider implements `ModelProvider`
  Output includes `status`, `message`, and `provider`
  No API key or network required

**Depends on**: T1

   

## T3: Integrate provider with ContractExecutor

**Goal**: Make contract execution call the model provider after input validation.

**Files**:

  `internal/contract/executor/executor.go`

**Acceptance Criteria**:

  `ContractExecutorImpl` owns a provider
  `Execute` validates input, calls provider, validates output, returns result
  Existing contract registration behavior remains unchanged

**Depends on**: T2

   

## T4: Update Dispatcher to use Execute

**Goal**: Replace placeholder dispatcher execution with provider backed contract execution.

**Files**:

  `internal/kernel/dispatcher/dispatcher.go`

**Acceptance Criteria**:

  Dispatcher calls `contractExecutor.Execute`
  Dispatcher returns provider output in `TaskResult.Output`
  Dispatcher still returns failed result on validation/execution errors

**Depends on**: T3

   

## T5: Add and update tests

**Goal**: Cover provider and updated execution behavior.

**Files**:

  `internal/model/provider/mock_test.go`
  `internal/contract/executor/executor_test.go`
  `internal/kernel/dispatcher/dispatcher_test.go`

**Acceptance Criteria**:

  Mock provider test passes
  Contract executor test asserts provider backed output
  Dispatcher test asserts provider output flows to task result
  `go test ./...` passes

**Depends on**: T4

   

## T6: Verify shell execution path

**Goal**: Confirm beginner path works from shell.

**Commands**:

```bash
go build  o axis.exe cmd/axis/main.go
```

Manual or piped check:

```powershell
@('run demo task','status demo task','exit') | .\axis.exe shell
```

**Acceptance Criteria**:

  No `contract default not found`
  No API key required
  Task can be submitted and status queried

**Depends on**: T5

   

## T7: Update docs and HANDOVER.md

**Goal**: Record that Axis now has a mock model provider layer, not real model integration.

**Files**:

  `docs/guides/BEGINNER_GUIDE.md`
  `HANDOVER.md`
  `docs/specs/model provider/tasks.md`

**Acceptance Criteria**:

  Beginner guide explains mock provider
  HANDOVER records MockModelProvider completion
  Tasks are marked completed

**Depends on**: T6


