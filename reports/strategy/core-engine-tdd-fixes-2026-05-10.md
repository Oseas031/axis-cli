# Core Engine TDD Bug Fixes Report

**Date:** 2026-05-10
**Scope:** Scheduler, Dispatcher, ContractExecutor, AnthropicProvider
**Approach:** Test-Driven Development (Red-Green-Refactor)

---

## Summary

5 critical bugs in the core engine domain were identified, reproduced with regression tests, and fixed. All fixes pass both new regression tests and the full existing test suite (`go test ./...`).

---

## Fix #1: Scheduler Partial Claim Rollback Failure

**Bug:** `GetReadyTasks` claims multiple candidates by marking them as `Running` and persisting to `stateStore`. If the second (or later) `Save` fails, previously claimed tasks in that batch remain `Running` instead of rolling back to `Pending`.

**Root Cause:**

```go
internal/kernel/scheduler/scheduler.go:203
if err := s.stateStore.Save(task.TaskID, state); err != nil {
    return nil, fmt.Errorf("failed to update task state: %w", err)
}
```

The error return path only aborted; it did not revert the in-memory status of already-processed candidates.

**Fix:** Roll back all previously claimed candidates `[0..i]` on failure.

```go
if err := s.stateStore.Save(task.TaskID, state); err != nil {
    for j := 0; j <= i; j++ {
        candidates[j].Status = types.TaskStatusPending
        candidates[j].StartedAt = nil
    }
    return nil, fmt.Errorf("failed to update task state: %w", err)
}
```

**Regression Test:** `TestScheduler_GetReadyTasks_RollbackOnSaveFailure`

---

## Fix #3 & #4: Dispatcher Context Propagation to Agent/Human Executors

**Bug:**
1. `Dispatcher.Dispatch` started a goroutine calling `d.executeTask(task)`, but `executeAgentTask` then called `agentExecutor.Execute(context.Background(), ...)` — losing the cancellable `timeoutCtx`.
2. `executeHumanTask` used a polling loop (`time.Sleep`) with no `ctx.Done()` check, so it would continue even after the dispatch context was cancelled.

**Fix:**
1. Pass `timeoutCtx` into `executeTask`, then into `executeAgentTask` and `executeHumanTask`.
2. In `executeHumanTask`, add a `select { case <-ctx.Done(): return }` before each `GetCallStatus` poll.

**Regression Tests:**
- `TestDispatcher_Dispatch_AgentExecutorPropagatesCancellation` — verifies the agent executor receives a cancellable context, not `Background`.
- `TestDispatcher_Dispatch_HumanExecutorRespectsCancellation` — verifies the human polling goroutine does not leak after context cancellation.

---

## Fix #5: ContractExecutor → Provider Context Propagation

**Bug:** `ContractExecutor.Execute` took no `context.Context` parameter. Both single-pass (`p.Execute(context.Background(), req)`) and multi-turn (`p.Execute(context.Background(), req)`) paths discarded the dispatcher's cancellable context.

**Fix:**
1. Changed `ContractExecutor` interface signature to `Execute(ctx context.Context, contractID string, input map[string]any)`.
2. Propagated `ctx` through `executeMultiTurn` to all provider and tool calls.
3. Updated all call sites (dispatcher, mock agent executor) to pass `ctx`.

**Regression Test:** `TestContractExecutor_Execute_PropagatesContext`

---

## Fix #6: safeMarshal Error Handling in Multi-Turn

**Bug:** After a tool call, the multi-turn loop JSON-marshaled the result via `safeMarshal(result)`. The error was silently discarded with `_`:

```go
content, _ := safeMarshal(result)
history = append(history, types.ModelMessage{..., Content: string(content)})
```

If `json.Marshal` failed (e.g., result contained a `chan`), `content` was `nil` → the model received an empty string message, breaking the tool-calling conversation.

**Fix:** Check the error. On failure, send a structured error message to the model instead of an empty string.

```go
content, marshalErr := safeMarshal(result)
if marshalErr != nil {
    history = append(history, types.ModelMessage{
        Role: "tool", ToolCallID: tc.ID,
        Content: fmt.Sprintf("error: failed to marshal tool result: %v", marshalErr),
    })
} else {
    history = append(history, types.ModelMessage{
        Role: "tool", ToolCallID: tc.ID,
        Content: string(content),
    })
}
```

**Regression Tests:**
- `TestSafeMarshal_ReturnsErrorForUnmarshalableValue`
- `TestContractExecutor_Execute_ToolResultMarshalErrorNotSwallowed`

---

## Fix #7: AnthropicProvider Tool Schema & Response Parsing

**Bug:**
1. Tool schema was sent as `{"type": "object"}` with no `properties` or `required` fields. Anthropic would reject or ignore such tools.
2. Assistant tool-use history messages were flattened into text strings (`"[ToolCall: ...]"`) instead of structured content blocks.
3. Response parsing ignored `tool_use` blocks; `resp.ToolCalls` was always empty.

**Fix:**
1. Refactored `anthropicTool.InputSchema` to a full JSON Schema type with `Properties` and `Required`.
2. Changed `anthropicMessage.Content` from `string` to `any`, supporting either a string or a slice of `anthropicContentBlock` structs.
3. History conversion now emits proper `tool_use` / `tool_result` content blocks.
4. Response parsing now iterates over all content blocks, extracting `tool_use` blocks into `ModelResponse.ToolCalls`.

**Regression Tests:**
- `TestAnthropicProvider_Execute_ToolSchemaIncludesProperties`
- `TestAnthropicProvider_Execute_ToolUseResponseParsed`

---

## Files Modified

| File | Lines Modified | Summary |
|---|---|---|
| `internal/kernel/scheduler/scheduler.go` | ~18 | Rollback partial claims on Save failure |
| `internal/kernel/scheduler/scheduler_test.go` | ~55 | `failingStateStore`, rollback test |
| `internal/kernel/dispatcher/dispatcher.go` | ~25 | Context propagation + human polling ctx check |
| `internal/kernel/dispatcher/dispatcher_test.go` | ~90 | Agent/human cancellation regression tests |
| `internal/contract/executor/executor.go` | ~20 | safeMarshal error handling, ctx propagation |
| `internal/contract/executor/executor_test.go` | ~45 | `TestContractExecutor_Execute_PropagatesContext` |
| `internal/contract/executor/executor_tool_test.go` | ~85 | safeMarshal + marshal error tests |
| `internal/model/provider/anthropic.go` | ~100 | Tool schema, content blocks, response parsing |
| `internal/model/provider/anthropic_test.go` | ~120 | Tool schema & tool_use parsing tests |
| `internal/agent/mock_executor.go` | ~1 | Pass ctx to contract executor |

## Verification Commands

```bash
go test ./internal/kernel/scheduler ./internal/kernel/dispatcher ./internal/contract/executor ./internal/model/provider -count=1 -timeout=60s
go test ./... -count=1 -timeout=120s
```

All commands pass with `EXIT:0`.
