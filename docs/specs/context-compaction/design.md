# Context Compaction Design

**Status**: Planned
**Implements**: `docs/specs/context-compaction/requirements.md`

## Architecture

```
internal/contract/executor/
  history_compact.go       # CompactionStrategy interface + Pipeline + token estimation
  history_compact_test.go
  tool_result_compact.go   # ToolResultCompaction strategy
  summarize_compact.go     # SummarizationCompaction strategy
  truncation_compact.go    # TruncationCompaction strategy (sliding window)

internal/model/tool/
  compact_tool.go          # compact tool registration
  compact_tool_test.go
```

## Core Interface

```go
// CompactionStrategy compacts a message history in place.
type CompactionStrategy interface {
    // Compact modifies history to reduce token count.
    // Returns true if any modification was made.
    Compact(ctx context.Context, history []types.ModelMessage, budget int) ([]types.ModelMessage, bool)
}

// CompactionPipeline runs strategies in order until budget is satisfied.
type CompactionPipeline struct {
    strategies []CompactionStrategy
    budget     int // max tokens for history
}
```

## Token Estimation

```go
// EstimateTokens estimates token count using 4-chars-per-token heuristic.
func EstimateTokens(messages []types.ModelMessage) int {
    total := 0
    for _, msg := range messages {
        total += len(msg.Content)/4 + 4 // 4 overhead per message
        for _, tc := range msg.ToolCalls {
            total += len(tc.Name)/4 + 10
        }
    }
    return total
}
```

## Message Groups

Messages are grouped into atomic units:
- **ToolCallGroup**: assistant message (with ToolCalls) + subsequent tool messages
- **UserGroup**: single user message
- **AssistantTextGroup**: assistant message without tool calls

ToolCallGroups are never split — either all messages in the group are kept or all are compacted together.

## ToolResultCompaction

```go
type ToolResultCompaction struct {
    KeepRecent int // number of recent tool-call groups to preserve (default 3)
}
```

Algorithm:
1. Identify all tool-call groups (assistant with ToolCalls + tool results)
2. Count from the end, preserve last `KeepRecent` groups
3. For older groups: replace tool message Content with `[Tool: {name} completed]`
4. Never touch assistant messages or user messages

## SummarizationCompaction

```go
type SummarizationCompaction struct {
    Provider      provider.ModelProvider
    KeepRecent    int // groups to preserve at end (default 4)
    SummaryPrompt string
}
```

Algorithm:
1. Split history into [older | recent] at KeepRecent boundary
2. Send older messages to provider with summarization prompt
3. Replace older messages with single user message: `[Session summary: ...]`

## TruncationCompaction

```go
type TruncationCompaction struct {
    KeepRecent int // minimum groups to preserve (default 2)
}
```

Algorithm:
1. Drop oldest message groups until under budget
2. Always preserve at least KeepRecent groups from the end

## Pipeline Execution

```go
func (p *CompactionPipeline) Compact(ctx context.Context, history []types.ModelMessage) []types.ModelMessage {
    for _, strategy := range p.strategies {
        if EstimateTokens(history) <= p.budget {
            break // budget satisfied, stop early
        }
        history, _ = strategy.Compact(ctx, history, p.budget)
    }
    return history
}
```

## Integration Point

In `executeMultiTurn`, after appending all tool results for a turn:

```go
// After tool results appended
if e.compactionPipeline != nil {
    history = e.compactionPipeline.Compact(ctx, history)
}
```

## CompactionConfig

```go
type CompactionConfig struct {
    Enabled    bool
    Budget     int // max tokens for history (default 32000)
    KeepRecent int // tool-call groups to keep intact (default 3)
}
```

## Resolved Decisions

### D1: Compaction in executor, not provider
- Compaction operates on history before it's sent to provider
- Provider sees already-compacted history

### D2: 4-char heuristic for token estimation
- Good enough for P0, avoids tiktoken dependency
- Can be replaced with model-specific tokenizer later

### D3: No background threading
- Go's model differs from Python; synchronous compaction is simpler
- Compaction is fast (L0 is string replacement, L2 is slice truncation)
- Only L1 (summarization) is slow, and it's P1
