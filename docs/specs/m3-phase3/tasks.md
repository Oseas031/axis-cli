# M3 Phase 3 Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)
- [workflow-binding.md](workflow-binding.md)

## Progress Tracking

| Task | Status | Depends On |
|---|---|---|
| T1: SLA types & constants | Completed | — |
| T2: Failure class routing + backoff | Completed | T1 |
| T3: Priority sorting in scheduler | Completed | T1 |
| T4: SLA admission extension | Completed | T1 |
| T5: Tool types & registry | Completed | — |
| T6: BashTool | Completed | T5 |
| T7: Extended ModelProvider (multi-turn) | Completed | T5 |
| T8: MockModelProvider tool-aware | Completed | T6, T7 |
| T9: Multi-turn execution loop | Completed | T5, T7 |
| T10: Orchestrator wiring | Completed | T2, T3, T4, T8, T9 |
| T11: Tests & coverage | Completed | T10 |
| T12: CLI/docs update | Completed | T11 |

---

## T1: SLA types & constants

**Goal**: Add failure class, priority, backoff constants to types package.

**Files**: `internal/types/types.go`

**Acceptance Criteria**:
- `FailureClassRetryable`, `FailureClassFatal`, `FailureClassDegradable` constants exist
- `SLAKeyPriority`, `SLAKeyBackoff` metadata keys exist
- `BackoffFixed`, `BackoffLinear`, `BackoffExponential` constants exist
- Compilation passes

---

## T2: Failure class routing + backoff

**Goal**: `parseSLA` returns failure class and backoff strategy; `executeTask` branches by failure class.

**Files**: `internal/kernel/orchestrator/orchestrator.go`

**Acceptance Criteria**:
- `parseSLA` returns `(timeout, retries, failureClass, backoff string)`
- `fatal` type does not retry, directly marks failed
- `degradable` type does not block on unready dependencies (skips dependency check)
- `retryable` + unset keeps current retry behavior
- Backoff strategy takes effect between retries

---

## T3: Priority sorting in scheduler

**Goal**: `GetReadyTasks` returns ready tasks sorted by `sla.priority` descending.

**Files**: `internal/kernel/scheduler/scheduler.go`

**Acceptance Criteria**:
- Parses each task's priority metadata (default 128)
- `GetReadyTasks` returns tasks sorted by priority descending
- Same priority maintains FIFO
- Does not affect Submit/GetStatus/Cancel behavior

---

## T4: SLA admission extension

**Goal**: Admission validates priority and backoff fields.

**Files**: `internal/contract/admission/admission.go`

**Acceptance Criteria**:
- `sla.priority`: Must be integer 0-255
- `sla.backoff`: Must be "fixed" | "linear" | "exponential"
- `sla.failure_class`: Must be "retryable" | "fatal" | "degradable"
- Invalid values rejected by admission

---

## T5: Tool types & registry

**Goal**: Define Tool interface, ToolRegistry, ToolDefinition/ToolCall/ToolResult types.

**Files**:
- `internal/types/types.go` — ToolCall, ToolResult, ToolDefinition, ModelMessage
- `internal/model/tool/tool.go` — Tool interface + ToolRegistry

**Acceptance Criteria**:
- `Tool` interface: Name(), Schema(), Execute(ctx, input) (output, error)
- `ToolRegistry`: Register, Get, List methods
- Compiler passes

---

## T6: BashTool

**Goal**: Implement Bash tool, executing commands via `os/exec`.

**Files**: `internal/model/tool/bash.go`

**Acceptance Criteria**:
- Reads input["command"] as bash command
- 30-second timeout (context.WithTimeout)
- Returns stdout, stderr, exit_code
- Command failure does not return error (non-zero exit_code is normal result), only system errors return error

---

## T7: Extended ModelProvider (multi-turn)

**Goal**: Extend ModelRequest/ModelResponse/ModelProvider to support multi-turn and tool calls.

**Files**: `internal/model/provider/provider.go`

**Acceptance Criteria**:
- `ModelRequest` adds `Tools []ToolDefinition` and `History []ModelMessage` fields
- `ModelResponse` adds `ToolCalls []ToolCall` field
- `ModelProvider` interface unchanged (Execute signature unchanged)
- Backward compatible (existing callers need no modification)

---

## T8: MockModelProvider tool-aware

**Goal**: Mock provider can simulate tool-use multi-turn interaction.

**Files**: `internal/model/provider/mock.go`

**Acceptance Criteria**:
- When input contains `"tool"` key: returns tool_call (not Output)
- When last History entry is tool result: returns final output
- No-tool scenario behavior unchanged

---

## T9: Multi-turn execution loop

**Goal**: ContractExecutor supports provider → tool → provider loop.

**Files**: `internal/contract/executor/executor.go`

**Acceptance Criteria**:
- If request contains tools, enters multi-turn mode
- Provider returns tool_calls → execute tools → feed back → repeat
- Maximum 10 turns
- If request has no tools, behavior unchanged (single turn)
- Tool execution errors recorded in ToolResult.Error

---

## T10: Orchestrator wiring

**Goal**: Assemble ToolRegistry + BashTool + SLA strategies in NewOrchestrator.

**Files**: `internal/kernel/orchestrator/orchestrator.go`

**Acceptance Criteria**:
- ToolRegistry created and BashTool registered
- ToolRegistry injected into ContractExecutor
- SLA strategies take effect through orchestrator
- `go build -o axis-dev.exe cmd/axis/main.go` succeeds

---

## T11: Tests & coverage

**Goal**: All new code has test coverage, coverage ≥ 85%.

**Files**:
- `internal/kernel/orchestrator/orchestrator_test.go` — SLA routing tests
- `internal/kernel/scheduler/scheduler_test.go` — priority ordering tests
- `internal/contract/admission/admission_test.go` — SLA validation tests
- `internal/model/tool/tool_test.go` — registry tests
- `internal/model/tool/bash_test.go` — bash execution tests
- `internal/model/provider/mock_test.go` — tool-aware mock tests
- `internal/contract/executor/executor_test.go` — multi-turn tests

**Acceptance Criteria**:
- `go test -race ./...` passes
- Coverage ≥ 85%

---

## T12: CLI/docs update

**Goal**: Update documentation to reflect M3 Phase 3 changes.

**Files**:
- `docs/status/current-progress.md`
- `HANDOVER.md`
- `docs/guides/QUICKSTART.md`

**Acceptance Criteria**:
- Progress docs mark Phase 3 completed
- HANDOVER records new capabilities
- Coverage data updated
