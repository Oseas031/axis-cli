# Milestone 2 Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)
- [workflow-binding.md](workflow-binding.md)

## Progress Tracking

| Task | Status |
|---|---|
| T0: Confirm workflow binding | Completed |
| T1: Baseline verification | Completed |
| T2: Add scheduler ready-set API | Completed |
| T2.5: Align ordinary CLI with Bash-first semantics | Completed |
| T3: Add contract admission layer | Completed |
| T4: Add SLA parsing and execution timeout | Completed |
| T5: Add orchestrator parallel execution loop | Completed |
| T6: Add error code foundation | Completed |
| T7: Update CLI/docs and run acceptance checks | Completed |

---

## T0: Confirm workflow binding

**Goal**: Ensure Milestone 2 development is routed through the project workflow mechanism before implementation starts.

**Files**:

- `workflow/entry.md`
- `workflow/meta-workflow-management.md`
- `workflow/occams-razor-architecture-simplification.md`
- `workflows/README.md`
- `.github/config/registry.yml`
- `docs/specs/milestone2/workflow-binding.md`

**Design References**:

- `workflow-binding.md` 鈥?Feature Workflow Execution Order
- `requirements.md` 鈥?Acceptance Criteria

**Acceptance Criteria**:

- Milestone 2 is classified as a new feature using `workflow/entry.md`
- upstream workflows are explicitly declared as `wf-doc-004 + wf-pr-check + wf-ci + wf-doc-006`
- no new workflow is introduced for Milestone 2
- user confirms specs and workflow binding before implementation

**Depends on**: None

---

## T1: Baseline verification

**Goal**: Confirm the current Milestone 1 codebase is green before changing behavior.

**Files**:

- None, verification only

**Commands**:

```bash
go test ./...
go build -o axis.exe cmd/axis/main.go
```

**Acceptance Criteria**:

- All existing tests pass
- CLI builds successfully
- Any pre-existing failure is documented before Milestone 2 implementation begins

**Current Result**:

- `go test ./...` passes
- `go build -o axis-baseline-check.exe cmd/axis/main.go` passes without overwriting existing `axis.exe`
- GitHub CI equivalent race/coverage command passes and reports total coverage `62.8%`, above the CI threshold of `60%`
- `gofmt -s -l .` reports no unformatted files
- `go vet ./...` passes
- Local machine does not currently have `staticcheck`, `gosec`, `govulncheck`, or `markdownlint` on PATH, so those GitHub Actions checks still require CI execution or explicit local tool installation

**Depends on**: T0

---

## T2: Add scheduler ready-set API

**Goal**: Extend the scheduler so the orchestrator can fetch multiple dependency-ready tasks in one call.

**Files**:

- `internal/kernel/scheduler/scheduler.go`
- `internal/kernel/scheduler/scheduler_test.go`

**Design References**:

- `design.md` 鈥?DAG scheduling
- `requirements.md` 鈥?FR1, FR2, FR3

**Acceptance Criteria**:

- `GetReadyTasks(limit int)` exists
- `GetNextTask` remains backward compatible
- independent pending tasks can be returned together
- returned tasks are marked `running` before execution
- tasks blocked by incomplete dependencies are not returned
- cycle detection tests still pass

**Current Result**:

- `GetReadyTasks(limit int)` added to the scheduler interface and implementation
- `GetNextTask` delegates to `GetReadyTasks(1)`
- scheduler tests cover FIFO ready-set selection, no-limit selection, dependency blocking, and dependency unblocking
- `go test ./internal/kernel/scheduler` passes
- GitHub CI equivalent `go test -v -race -coverprofile=coverage.out -covermode=atomic ./...` passes with total coverage `63.6%`
- `gofmt` and `go vet ./...` pass

**Depends on**: T1

---

## T2.5: Align ordinary CLI with Bash-first semantics

**Goal**: Ensure ordinary CLI commands do not mislead users about cross-process orchestrator state and remain aligned with `bash is all you need, simple but robust, composable and extensible`.

**Files**:

- `cmd/axis/main.go`
- `cmd/axis/main_test.go`

**Design References**:

- `workflow-binding.md` 鈥?Keep CLI/shell-native validation as the default interface
- `requirements.md` 鈥?FR3 backward compatibility
- `docs/architecture/bash-is-all-you-need.md` 鈥?CLI First, Shell Native, Composable

**Acceptance Criteria**:

- `axis run <task-id>` does not tell users to run `axis start` first
- `axis run <task-id>` initializes the local orchestrator path and submits a task in the current process
- `axis status <task-id>` returns a contextual not-found message instead of implying cross-process state exists
- `axis shell` remains unchanged
- CI-equivalent tests pass

**Current Result**:

- `axis run <task-id>` now initializes the local orchestrator path and submits without telling users to run `axis start`
- `axis status <task-id>` now initializes the local orchestrator path and returns a local-process not-found message when the task is absent
- `cmd/axis/main_test.go` covers ordinary CLI run/status behavior
- `go test ./cmd/axis` passes
- GitHub CI equivalent `go test -v -race -coverprofile=coverage.out -covermode=atomic ./...` passes with total coverage `67.3%`
- `gofmt`, `go vet ./...`, `staticcheck ./...`, `gosec ./...`, and `govulncheck ./...` pass

**Depends on**: T2

---

## T3: Add contract admission layer

**Goal**: Reject invalid tasks before they enter the scheduler queue.

**Files**:

- `internal/contract/admission/admission.go`
- `internal/contract/admission/admission_test.go`
- `internal/kernel/orchestrator/orchestrator.go`

**Design References**:

- `design.md` 鈥?Contract admission
- `requirements.md` 鈥?FR4

**Acceptance Criteria**:

- unknown contracts are rejected before scheduling
- invalid input schema is rejected before scheduling
- valid tasks continue to submit successfully
- errors include stable contextual information suitable for CLI output

**Depends on**: T2

---

## T4: Add SLA parsing and execution timeout

**Goal**: Interpret minimal SLA metadata and apply timeout/retry behavior around dispatch.

**Files**:

- `internal/types/types.go`
- `internal/kernel/orchestrator/orchestrator.go`
- `internal/kernel/orchestrator/orchestrator_test.go`

**Design References**:

- `design.md` 鈥?SLA metadata
- `requirements.md` 鈥?FR5, FR7

**Acceptance Criteria**:

- `sla.timeout_ms` is parsed strictly
- `sla.max_retries` is parsed strictly
- missing SLA metadata keeps current default behavior
- timeout marks task failed with contextual error information
- retry exhaustion is visible in task result or error context

**Depends on**: T3

---

## T5: Add orchestrator parallel execution loop

**Goal**: Execute multiple independent ready tasks concurrently while preserving safe shutdown.

**Files**:

- `internal/kernel/orchestrator/orchestrator.go`
- `internal/kernel/orchestrator/orchestrator_test.go`

**Design References**:

- `design.md` 鈥?Orchestrator parallel execution
- `requirements.md` 鈥?FR2, FR7

**Acceptance Criteria**:

- independent tasks execute concurrently
- dependent tasks wait until dependencies complete
- worker limit prevents unbounded goroutine creation
- shutdown cancels outstanding work without goroutine leaks
- existing single-task behavior still works

**Depends on**: T4

---

## T6: Add error code foundation

**Goal**: Introduce a small stable error vocabulary without building a large framework.

**Files**:

- `internal/types/types.go`
- scheduler/admission/orchestrator files touched by prior tasks
- relevant tests

**Design References**:

- `design.md` 鈥?Error code foundation
- `requirements.md` 鈥?FR6

**Acceptance Criteria**:

- error codes exist for scheduler lifecycle, task lookup, duplicate task, dependency cycle, contract missing, input invalid, timeout, retry exhaustion
- tests assert codes where behavior depends on them
- CLI-visible errors remain readable

**Depends on**: T5

---

## T7: Update CLI/docs and run acceptance checks

**Goal**: Complete handoff-quality documentation and validate Milestone 2 readiness.

**Files**:

- `docs/status/current-progress.md`
- `HANDOVER.md`
- `AGENT_INSTRUCTIONS.md`
- `docs/product/ROADMAP.md` if status wording needs synchronization

**Commands**:

```bash
go test ./...
go build -o axis.exe cmd/axis/main.go
```

**Acceptance Criteria**:

- progress docs reflect Milestone 2 implementation status
- handover docs identify remaining tasks clearly
- all tests pass
- build succeeds
- no Milestone 3 scope is introduced

**Depends on**: T6


