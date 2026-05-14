# Coding Agent — Tasks

> 实现 `docs/specs/coding-agent/design.md`

**Status**: In Progress

## Completed

- [x] **T1**: Define `LLMAgentExecutor` struct with functional options — `internal/agent/llm_executor.go`
- [x] **T2**: Implement multi-turn loop (LLM → ToolCalls → Execute → History → repeat) — same file
- [x] **T3**: Implement circuit breaker (consecutive error counter) — same file
- [x] **T4**: Implement `TerminationFn` pluggable termination — same file
- [x] **T5**: Implement `HistoryCompactor` interface with no-op default — same file
- [x] **T6**: Implement `ToolTrace` recording — same file
- [x] **T7**: Wire into orchestrator via `WithAgentExecutor` + `SetToolRegistry` — `internal/kernel/orchestrator/orchestrator.go`
- [x] **T8**: Register in `cmd/axis/main.go` `initOrchestrator()` — done
- [x] **T9**: Unit tests (7 cases: single-turn, multi-turn, circuit breaker, budget, custom termination, tool error recovery, context cancellation) — `internal/agent/llm_executor_test.go`
- [x] **T10**: Permission Ladder L0 whitelist for BashTool — `internal/model/tool/bash.go`
- [x] **T11**: FileWriteTool path validation hardening — `internal/model/tool/file_write.go`

## Remaining

- [ ] **T12**: Integration test — submit `executor_type: "agent"` task via `axis ask --submit`, verify end-to-end with real provider
- [ ] **T13**: Coding-specific `TerminationFn` — "go build passes" as completion criterion
- [ ] **T14**: Wire `PermissionL1` into orchestrator's BashTool registration (currently Unrestricted)
- [ ] **T15**: History compaction implementation (reuse `ThreeLayerCompaction` from contract executor)
- [ ] **T16**: System prompt templating — inject task description, project context, tool list into prompt

## Dependencies

- T13 depends on: T12 (need real execution to test termination)
- T14 depends on: autonomy level propagation from dispatcher to tool registry
- T15 depends on: provider configured (compaction needs summarization model)
