# Context Compaction Tasks

**Status**: Planned
**Implements**: requirements.md + design.md

---

## T1: Token Estimation + ToolResultCompaction

- `history_compact.go`: EstimateTokens, CompactionStrategy interface, ToolResultCompaction
- `history_compact_test.go`: estimation accuracy, tool result truncation, KeepRecent behavior

## T2: Pipeline + TruncationCompaction

- CompactionPipeline struct with ordered strategy execution
- TruncationCompaction (drop oldest groups)
- Tests: pipeline stops early, truncation respects KeepRecent

## T3: Integration into executeMultiTurn

- Add CompactionPipeline field to ContractExecutorImpl
- Call pipeline after tool results appended each turn
- Add CompactionConfig + SetCompactionConfig setter
- Wire default config in orchestrator

## T4: SummarizationCompaction

- Uses ModelProvider to summarize older messages
- Replace older span with single summary message
- Tests with MockModelProvider

## T5: compact Tool

- Register `compact` tool in ToolRegistry
- Handler triggers pipeline manually, returns compaction summary
- Tests: tool callable, returns result

## T6: Full Test + Commit

- `go test -race ./internal/contract/executor/...`
- `go test -race ./internal/model/tool/...`
- Commit all changes
