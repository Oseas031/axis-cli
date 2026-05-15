# COMPACTOR BOUNDARY

> 展开自 CLAUDE.md §6 (语义边界)

`internal/model/compactor/` implements context-aware history compaction for the multi-turn loop.

## What This Package Does

- Offloads tool result full text to `{dataDir}/refs/*.md`
- Maintains an append-only index at `{dataDir}/offload.jsonl`
- Implements `multiturn.Compactor` interface with tiered compression:
  - Mild: replace high-score tool results with summaries
  - Aggressive: drop oldest assistant+tool pairs
  - Emergency: truncate to 60% budget

## What This Package Must NEVER Do

1. **Never run background goroutines** — compaction is synchronous, triggered by the multi-turn loop
2. **Never inject content into provider prompts** — it only reduces/replaces existing history
3. **Never physically delete offload.jsonl entries** — append-only
4. **Never introduce external dependencies** — pure Go standard library
5. **Never modify scheduler/dispatcher/contract semantics**
6. **Never access network** — all I/O is local filesystem
7. **Never break assistant→tool message pairing** — pairs must remain intact or be dropped together

## Storage Layout

```
{dataDir}/
├── refs/{timestamp}-{hash}.md   # Full tool result text (immutable once written)
└── offload.jsonl                # Append-only index of OffloadEntry records
```

## Key Invariants

- `refs/` files are write-once, never modified after creation
- `offload.jsonl` is append-only, LF-terminated
- `estimateTokens` is a heuristic (CJK/1.5 + other/4), not exact
- `Summarizer` interface allows plugging in LLM-based summarization later
- Already-offloaded messages (prefix `[offloaded:`) are never re-processed
