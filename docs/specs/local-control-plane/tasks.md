# Local Control Plane Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)
- [../../architecture/spec-lifecycle-conventions.md](../../architecture/spec-lifecycle-conventions.md)
- [../../architecture/semantic-boundaries.md](../../architecture/semantic-boundaries.md)
- [../natural-language-scheduling/](../natural-language-scheduling/)
- [../interactive-shell/](../interactive-shell/)

## Progress Tracking

| Task | Status |
|---|---|
| T1: Create Local Control Plane spec | Completed |
| T2: Define control protocol models | Completed |
| T3: Implement runtime locator | Completed |
| T4: Implement local control server | Completed |
| T5: Implement local control client | Completed |
| T6: Wire `axis start` to runtime server | Completed |
| T7: Route `axis ask --submit` through client | Completed |
| T8: Route `axis status` through client | Completed |
| T9: Align shell behavior with local runtime | Completed |
| T10: Add event log and safety tests | Completed |
| T11: Document operational workflow | Completed |

---

## T1: Create Local Control Plane spec

**Goal**: Define the Axis-aligned boundary for cross-process local runtime behavior.

**Files**:

- `docs/specs/local-control-plane/requirements.md`
- `docs/specs/local-control-plane/design.md`
- `docs/specs/local-control-plane/tasks.md`
- `docs/README.md`

**Design References**:

- `requirements.md#summary`
- `design.md#overview`

**Acceptance Criteria**:

- Requirements, design, and tasks exist.
- The spec explains the current cross-process task state failure.
- The design keeps local runtime separate from hidden control policy.
- The spec is linked from `docs/README.md`.

**Depends on**: None

---

## T2: Define control protocol models

**Goal**: Add typed request/response models for local task submission, status query, health, and errors.

**Files**:

- `internal/control/types.go`
- `internal/control/types_test.go`

**Design References**:

- `design.md#local-control-boundary`
- `design.md#localcontrolserver`

**Acceptance Criteria**:

- Submit request carries a `types.AgentTask`.
- Submit response returns task ID and accepted status.
- Status response returns task ID and task status.
- Error response is stable and machine-readable.
- JSON round-trip tests pass.

**Depends on**: T1

---

## T3: Implement runtime locator

**Goal**: Let short-lived CLI clients discover a running project-local runtime.

**Files**:

- `internal/control/locator.go`
- `internal/control/locator_test.go`

**Design References**:

- `design.md#runtime-locator`

**Acceptance Criteria**:

- Locator writes `.axis/runtime.json` atomically.
- Locator records protocol, address, pid, project root, and started time.
- Locator load handles missing and malformed files.
- Tests use temporary project roots.
- No secrets are written to the locator.

**Depends on**: T2

---

## T4: Implement local control server

**Goal**: Expose local-only submit, status, and health operations over the selected IPC mechanism.

**Files**:

- `internal/control/server.go`
- `internal/control/server_test.go`

**Design References**:

- `design.md#localcontrolserver`
- `design.md#local-control-boundary`

**Acceptance Criteria**:

- Server binds only to local access.
- Submit calls `Orchestrator.SubmitTask`.
- Status calls `Orchestrator.GetTaskStatus`.
- Health returns runtime metadata without secrets.
- Tests cover submit, status, not found, and malformed request.

**Depends on**: T2

---

## T5: Implement local control client

**Goal**: Provide a CLI-safe client for submit, status, and health operations.

**Files**:

- `internal/control/client.go`
- `internal/control/client_test.go`

**Design References**:

- `design.md#localcontrolclient`

**Acceptance Criteria**:

- Client loads the runtime locator.
- Client returns actionable no-runtime errors.
- Client can submit a task to a test server.
- Client can query status from a test server.
- Client does not read provider credentials.

**Depends on**: T3, T4

---

## T6: Wire `axis start` to runtime server

**Goal**: Make `axis start` own the local runtime server and orchestrator execution loop.

**Files**:

- `cmd/axis/main.go`
- `cmd/axis/control_runtime.go`
- `cmd/axis/main_test.go`

**Design References**:

- `design.md#localruntime`
- `design.md#cli-integration`

**Acceptance Criteria**:

- `axis start` starts the orchestrator and local control server.
- Runtime locator is written on start.
- Shutdown removes or invalidates the locator when possible.
- Tests verify startup wiring without requiring fixed ports.

**Depends on**: T3, T4

---

## T7: Route `axis ask --submit` through client

**Goal**: Submit natural-language tasks to the running local runtime from a separate CLI process.

**Files**:

- `cmd/axis/ask_cmd.go`
- `cmd/axis/ask_cmd_test.go`

**Design References**:

- `design.md#cli-integration`
- `requirements.md#fr2-cross-process-task-submission`

**Acceptance Criteria**:

- Dry-run `axis ask` remains local and does not require runtime.
- `axis ask --submit` uses the local control client.
- Missing runtime returns guidance to run `axis start`.
- Submitted task preserves intent metadata and context readiness metadata when requested.

**Depends on**: T5, T6

---

## T8: Route `axis status` through client

**Goal**: Query task status from the running local runtime instead of a new in-process orchestrator.

**Files**:

- `cmd/axis/main.go`
- `cmd/axis/main_test.go`

**Design References**:

- `requirements.md#fr3-cross-process-status-query`
- `design.md#cli-integration`

**Acceptance Criteria**:

- `axis status <task-id>` uses the local control client.
- Missing runtime returns actionable guidance.
- Unknown task returns a stable not-found error.
- Cross-process submit/status behavior is covered by tests or a deterministic integration seam.

**Depends on**: T5, T6

---

## T9: Align shell behavior with local runtime

**Goal**: Keep shell and CLI task semantics aligned after introducing the local runtime.

**Files**:

- `cmd/axis/shell_cmd.go`
- `cmd/axis/main_test.go`
- `docs/specs/interactive-shell/design.md`

**Design References**:

- `design.md#shell-integration`

**Acceptance Criteria**:

- Shell either uses the local control client or clearly owns an explicit in-process session.
- Shell `ask --submit` and CLI `ask --submit` do not diverge silently.
- Tests document the chosen behavior.

**Depends on**: T7, T8

---

## T10: Add event log and safety tests

**Goal**: Record observable local task lifecycle events and protect safety boundaries.

**Files**:

- `internal/control/events.go`
- `internal/control/events_test.go`
- `cmd/axis/main_test.go`

**Design References**:

- `requirements.md#fr4-observable-local-events`
- `design.md#task-event-log`

**Acceptance Criteria**:

- Submission and terminal status events are recorded locally.
- Event records do not contain provider API keys or raw secrets.
- Tests cover local-only communication boundary.
- Tests verify contract/admission path is not bypassed.

**Depends on**: T4, T6, T7, T8

---

## T11: Document operational workflow

**Goal**: Explain how users and agents should operate the local runtime.

**Files**:

- `README.md`
- `docs/README.md`
- `docs/guides/QUICKSTART.md`
- `docs/specs/local-control-plane/design.md`

**Design References**:

- `requirements.md#acceptance-criteria`
- `design.md#p0-success-definition`

**Acceptance Criteria**:

- Docs explain `axis start`, `axis ask --submit`, and `axis status` workflow.
- Docs explain no-runtime errors.
- Docs state that the control plane is local-only and not a hidden permission system.
- Docs include the user's reported workflow as a supported scenario.

**Depends on**: T8, T9, T10
