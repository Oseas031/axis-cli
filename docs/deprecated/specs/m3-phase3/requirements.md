# M3 Phase 3 Requirements — SLA Strategy Engine & Tool Invocation Layer

## Summary

M3 Phase 3 contains two independent sub-features:

1. **SLA Strategy Engine**: Make `failure_class` actually affect scheduling and retry behavior, add priority ordering and backoff strategies
2. **Tool Invocation Layer**: Give the task execution chain multi-turn tool invocation capability, with Bash as the first tool

Both can be developed, tested, and merged independently.

## Design Philosophy

- **More Context**: Strategy engine handles failures differently based on failure type; tool invocation brings execution results back into model context
- **More Action**: Tool layer lets tasks actually execute Bash commands, not just generate text
- **Zero Control**: Strategies are configurable, tools are registrable, no single hardcoded path
- **Controllable Evolution**: Strategy and tool changes must be observable, testable, and rollbackable
- **Bash is All You Need, simple but robust, composable and extensible**: First tool is Bash, CLI-native, keeping errors clear and extensible

## Users

- Scheduling scenarios requiring differentiated failure handling
- Agents needing to execute Bash commands through tasks
- Future Provider implementations needing multi-turn tool-use

## Functional Requirements

### SLA Strategy Engine

- **FR1**: `failure_class` value determines failure behavior:
  - `"retryable"` — Backoff retry up to max attempts
  - `"fatal"` — Fail immediately, no retry
  - `"degradable"` — Run in degraded mode when dependencies not ready (skip missing dependencies)
  - Unset keeps current default behavior (retry all)
- **FR2**: Backoff strategy configurable: fixed interval, linear growth, exponential backoff, default fixed 100ms
- **FR3**: Priority field `sla.priority` (0-255), higher priority tasks returned first by `GetReadyTasks`
- **FR4**: Scheduler sorts ready tasks by priority, same priority maintains FIFO

### Tool Invocation Layer

- **FR5**: `Tool` interface: `Name()`, `Schema()`, `Execute(ctx, input) → output`
- **FR6**: `ToolRegistry`: register, find, list tools
- **FR7**: `BashTool`: execute shell commands, return stdout/stderr/exit_code, 30-second timeout
- **FR8**: `ModelRequest` extension, supports `Tools []ToolDefinition` field
- **FR9**: `ModelResponse` extension, supports `ToolCalls []ToolCall` field (non-nil during tool_use)
- **FR10**: `ContractExecutor` supports multi-turn execution loop: provider → tool_use? → execute tool → feed result → provider → ... → final output

### Shared

- **FR11**: All new behavior covered by `go test -race ./...`
- **FR12**: Coverage no less than 85%

## Non-Goals

- Real LLM integration (still Mock/Echo only)
- Network tools (http client, etc.)
- File read/write tools (Phase 4, Bash covers this for now)
- SLA compliance tracking / metrics
- Dynamic priority adjustment
- Streaming returns from tool invocations

## Acceptance Criteria

- [x] `failure_class` three behaviors execute correctly
- [x] Priority ordering effective in `GetReadyTasks`
- [x] `BashTool` can execute commands and return results
- [x] Multi-turn tool-use loop works in `ContractExecutor`
- [x] `go test -race ./...` passes
- [x] Coverage ≥ 85%

## Constraints

- Go stdlib only (Bash uses `os/exec`)
- Do not change scheduler core semantics (FIFO remains default)
- Do not change existing API signatures (extension only)
- Do not introduce external DSL or rule engines
