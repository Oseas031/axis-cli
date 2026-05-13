# Context Compaction Requirements

> 实现 agent-native-first-principles.md P2（Query is Context）


**Status**: Planned
**Inspired by**: Anthropic Context Engineering (2026), Microsoft Agent Framework Compaction, learn-claude-code s06
**Related**: `internal/contract/executor/executor.go` (executeMultiTurn)

## Summary

Context Compaction is a pipeline-based conversation history compression system that prevents token overflow in multi-turn agent loops. It operates on `[]types.ModelMessage` history within the executor's tool-calling loop.

Three layers, from gentle to aggressive:
- **L0 ToolResultCompaction**: Truncate old tool result content to short summaries
- **L1 SummarizationCompaction**: Use LLM to summarize older conversation spans
- **L2 TruncationCompaction**: Drop oldest message groups as emergency backstop

## Design Philosophy

- **Pipeline**: Strategies execute in order, gentlest first. Stop early if budget satisfied.
- **Atomic Groups**: assistant+tool messages are atomic units (never split).
- **Threshold-triggered**: Only compact when token estimate exceeds threshold.
- **Non-destructive**: Original messages preserved in structure; only Content field modified.

## Functional Requirements

### FR1: Token Estimation

Estimate token count of `[]ModelMessage` using 4-chars-per-token heuristic.

### FR2: ToolResultCompaction

- Trigger: estimated tokens > threshold (default 80% of budget)
- Keep last N tool-call groups intact (default N=3)
- Replace older tool message Content with `[Tool: {name} completed]`
- Never modify assistant messages or user messages

### FR3: SummarizationCompaction

- Trigger: estimated tokens > higher threshold (default 90% of budget)
- Summarize older messages using provider, replace with single Summary message
- Preserve last M message groups intact (default M=4)
- Requires ModelProvider for summarization call

### FR4: TruncationCompaction (Sliding Window)

- Trigger: estimated tokens still > budget after L0+L1
- Drop oldest message groups until under budget
- Always preserve at least the last K groups (default K=2)

### FR5: Pipeline

- Execute strategies in order: ToolResult → Summarization → Truncation
- Each strategy checks its own trigger before acting
- Pipeline stops early if token estimate drops below target

### FR6: compact Tool

- Register `compact` tool in ToolRegistry
- Agent can call it to manually trigger compaction
- Returns summary of what was compacted

### FR7: Integration

- Compaction runs inside `executeMultiTurn` after each tool result batch
- Configurable via CompactionConfig on ContractExecutorImpl
- Default config: enabled with sensible thresholds

## Non-Goals

- No background threading (Go's concurrency model differs from Python)
- No disk persistence of transcripts in P0
- No prompt caching optimization (provider-level concern)
- No modification of provider interface

## Acceptance Criteria

- `go test -race ./internal/contract/executor/...` passes
- History stays bounded even with 50+ tool calls
- ToolResultCompaction alone handles 80% of cases without LLM call
- compact tool callable by Agent
