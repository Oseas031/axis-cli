# Coding Agent — Design

> 实现 `docs/architecture/agent-design-first-principles.md` §3-§4

**Status**: In Progress

## Architecture

```
AgentExecutionRequest
  → LLMAgentExecutor.Execute()
    → for iter < maxIter:
        → provider.Execute(ctx, modelReq)  // LLM thinks
        → if no tool calls: checkTermination() → Complete/Failed/NeedHuman/Continue
        → for each tool call:
            → registry.Get(name) → tool.Execute(ctx, input)
            → record ToolTrace
            → append result to history
            → circuit breaker check
        → compactor.Compact(history)
    → return AgentExecutionResult (with traces + AgentID)
```

## Key Types

```go
type LLMAgentExecutor struct {
    provider     provider.ModelProvider  // injected
    tools        *tool.Registry          // injected by orchestrator
    systemPrompt string                  // defines agent role
    maxIter      int                     // external budget
    terminate    TerminationFn           // pluggable stop condition
    compactor    HistoryCompactor        // v1: no-op
    maxErrors    int                     // circuit breaker threshold
    agentID      string                  // identity for tracing
}

type TerminationFn func(history []ModelMessage, last *ModelResponse) TerminationDecision
type TerminationDecision int // Continue | Complete | Failed | NeedHuman

type HistoryCompactor interface {
    Compact(ctx context.Context, history []ModelMessage) []ModelMessage
}

type ToolTrace struct {
    Name, Output, Error string
    Input               map[string]any
    Duration            time.Duration
}
```

## Integration Points

| Component | How it connects |
|---|---|
| Orchestrator | `WithAgentExecutor(exec)` option → wires into dispatcher |
| Dispatcher | Routes `executor_type: "agent"` tasks to LLMAgentExecutor |
| Tool Registry | Orchestrator calls `exec.SetToolRegistry(registry)` after construction |
| Provider | Same provider instance used by ContractExecutor |

## Design Decisions

1. **Reuses same loop shape as `ContractExecutorImpl.executeMultiTurn`** — not a third implementation, same pattern with Agent-specific termination/tracing.
2. **Functional options pattern** (`WithSystemPrompt`, `WithMaxIterations`, etc.) — extensible without breaking existing callers.
3. **Default termination = "no tool calls means complete"** — simplest correct default, overridable for coding tasks that need "tests pass" verification.
4. **ToolTrace stored in `Output["_tool_traces"]`** — no new fields needed on AgentExecutionResult, traces are part of output.

## Permission Ladder Integration

BashTool now supports `PermissionLevel` (L0/L1/Unrestricted):
- L0: read-only commands (ls, cat, grep, git status, etc.)
- L1: + build/test (go build, go test, npm run build)
- Unrestricted: anything (backward compat default)

Future: orchestrator selects permission level based on Agent's autonomy level.
