# M3 Phase 3 Design — SLA Strategy Engine & Tool Invocation Layer

## Architecture Overview

```
Orchestrator.SubmitTask
  → AdmissionValidator (SLA validation: priority range check)
  → Scheduler.Submit (stores task with metadata)
  → Scheduler.GetReadyTasks (priority-sorted, same-priority FIFO)
  → Orchestrator.executeTask
    → parseSLA (timeout, retries, failureClass, backoff)
    → retry loop with backoff + failure_class routing
    → Dispatcher.Dispatch
      → ContractExecutor.Execute
        → ValidateInput
        → execution loop:
          → ModelProvider.Execute(ctx, request with tools)
          → if tool_calls in response:
              → ToolRegistry.Execute(tool_call)
              → append result to conversation
              → loop back to provider
          → final response
        → ValidateOutput
      → TaskResult
```

## 1. SLA Strategy Engine

### 1.1 New Metadata Keys

```go
const (
    SLAKeyTimeoutMs    = "sla.timeout_ms"
    SLAKeyMaxRetries   = "sla.max_retries"
    SLAKeyFailureClass = "sla.failure_class"
    SLAKeyPriority     = "sla.priority"       // NEW: 0-255
    SLAKeyBackoff      = "sla.backoff"        // NEW: "fixed" | "linear" | "exponential"
)
```

### 1.2 Failure Class Behaviors

```go
const (
    FailureClassRetryable  = "retryable"   // default: retry with backoff
    FailureClassFatal      = "fatal"       // no retry, immediate fail
    FailureClassDegradable = "degradable"  // skip missing deps, proceed
)
```

### 1.3 Backoff Strategy

```go
type BackoffStrategy interface {
    Delay(attempt int) time.Duration
}
// fixed: 100ms constant
// linear: 100ms * (attempt+1)
// exponential: 100ms * 2^attempt, max 30s
```

### 1.4 Priority Sorting

In `SchedulerImpl.GetReadyTasks`:
1. Parse each ready task's `sla.priority` (default 128)
2. Sort by priority descending
3. Same priority maintains FIFO order
4. Only affects return order of ready tasks, does not affect Submit/Status semantics

### 1.5 Admission Changes

`validateSLA` additions:
- `sla.priority`: Must be an integer 0-255
- `sla.backoff`: Must be one of "fixed" | "linear" | "exponential"
- `sla.failure_class`: Must be one of "retryable" | "fatal" | "degradable"

## 2. Tool Invocation Layer

### 2.1 New Types

```go
// ToolDefinition describes a tool for the model provider.
type ToolDefinition struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Parameters  []FieldDef  `json:"parameters"`
}

// ToolCall represents a request from the provider to use a tool.
type ToolCall struct {
    ID     string         `json:"id"`
    Name   string         `json:"name"`
    Input  map[string]any `json:"input"`
}

// ToolResult is the result of a tool execution.
type ToolResult struct {
    CallID string         `json:"call_id"`
    Output map[string]any `json:"output"`
    Error  string         `json:"error,omitempty"`
}
```

### 2.2 Extended ModelRequest / ModelResponse

```go
type ModelRequest struct {
    ContractID string
    Input      map[string]any
    Tools      []ToolDefinition   // NEW
    History    []ModelMessage     // NEW: prior turns
}

type ModelResponse struct {
    Output    map[string]any
    ToolCalls []ToolCall          // NEW: non-nil when model wants tools
}

type ModelMessage struct {
    Role    string         // "user" | "assistant" | "tool"
    Content string
    ToolCallID string
    ToolCalls  []ToolCall
}
```

### 2.3 Tool Interface and Registry

```go
type Tool interface {
    Name() string
    Schema() ToolDefinition
    Execute(ctx context.Context, input map[string]any) (map[string]any, error)
}

type ToolRegistry struct {
    tools map[string]Tool
}
```

Package: `internal/model/tool/` (new)

### 2.4 BashTool

```go
type BashTool struct{}

func (t *BashTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    cmd := exec.CommandContext(ctx, "bash", "-c", input["command"].(string))
    output, err := cmd.CombinedOutput()
    return map[string]any{
        "stdout":   string(output),
        "exit_code": cmd.ProcessState.ExitCode(),
    }, err
}
```

### 2.5 Multi-Turn Execution Loop

`ContractExecutor.Execute` changes:

```go
func (e *ContractExecutorImpl) Execute(contractID string, input map[string]any) (*ExecutionResult, error) {
    // 1. ValidateInput (unchanged)
    // 2. Build initial request with tools
    // 3. Loop (max N turns):
    //    a. provider.Execute(ctx, request)
    //    b. if response.ToolCalls != nil:
    //       - execute each tool via registry
    //       - append tool results to request.History
    //       - continue loop
    //    c. if response.Output != nil:
    //       - break (final response)
    // 4. ValidateOutput (unchanged)
}
```

Max turns: 10 (prevents infinite loops)

### 2.6 MockModelProvider Adaptation

Mock provider needs to support tool-use simulation:
- When request contains tools and input contains `"tool": "<name>"`, return tool_call
- When the last history entry is a tool result, return final output
- Default behavior without tools remains unchanged

### 2.7 Orchestrator Construction Changes

```go
func NewOrchestrator(opts ...OrchestratorOption) *Orchestrator {
    // ...
    toolRegistry := tool.NewRegistry()
    toolRegistry.Register(tool.NewBashTool())
    contractExec := contractexec.NewContractExecutor()
    contractExec.SetToolRegistry(toolRegistry)
    // ...
}
```

## File Structure

```
internal/types/types.go              # +ToolCall, ToolResult, ToolDefinition,
                                     #  SLA constants, failure class constants,
                                     #  ModelMessage, ModelRequest/Response extension
internal/model/tool/tool.go          # Tool interface, ToolRegistry
internal/model/tool/bash.go          # BashTool implementation
internal/model/tool/tool_test.go     # Tool tests
internal/model/provider/provider.go  # Extended ModelRequest/ModelResponse
internal/model/provider/mock.go      # Tool-aware mock provider
internal/contract/executor/executor.go # Multi-turn execution loop
internal/kernel/scheduler/scheduler.go # Priority sorting in GetReadyTasks
internal/kernel/orchestrator/orchestrator.go # parseSLA + failure_class routing + backoff
internal/contract/admission/admission.go    # Extended SLA validation
```

## Trade-offs

| Option | Decision | Rationale |
|---|---|---|
| Tool in ContractExecutor vs Dispatcher | ContractExecutor | Tool use is part of contract execution, not routing |
| Priority in Scheduler vs Orchestrator | Scheduler | GetReadyTasks is the natural ordering point |
| Real shell vs sandbox | Real shell (os/exec) | M3 scope; sandbox in M4+ |
| Max turns hardcoded vs configurable | Hardcoded 10 | Simplicity; make configurable when needed |
| Tool history in provider vs executor | Executor manages loop, provider sees history | Clear separation: executor drives loop, provider responds |

## Risks

| Risk | Mitigation |
|---|---|
| Bash tool is dangerous (rm -rf, etc.) | Phase 3 scope is limited; sandbox/safety in M4 |
| Priority breaks FIFO assumption in tests | Priority defaults to 128, same-priority → FIFO preserved |
| Multi-turn loop inflates latency | Max 10 turns, each with 30s tool timeout |
