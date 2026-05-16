---
name: mind-magnifier
description: Research information preprocessing tool (MindMagnifier/amp). Use when the user asks about latest research, papers, AI news, or when you need to gather research context for a task.
triggers:
  - user
  - model
---

# MindMagnifier (amp)

> AI-native research information preprocessor. Fetches, deduplicates, categorizes, summarizes, and ranks multi-source research content.

## Location

```
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe
```

## When to Use

- User asks about latest research / papers / AI news
- User says "查一下最新的..." / "有什么新论文" / "research update"
- You need research context to inform a design decision
- User explicitly mentions MindMagnifier or amp

## Commands

```bash
# Sync latest from all sources (arXiv, DeepMind, HN, etc.)
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe sync

# List recent entries (table format)
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe list --limit 10

# Filter by category
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe list --category llm --json

# Export structured JSON for consumption
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe export --format json --limit 20

# Show entries from last 24 hours
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe list --since 24h --json
```

## Categories

`llm` | `vision` | `rl` | `nlp` | `audio` | `robotics` | `general`

## Output Schema (JSON)

```json
[{
  "source_org": "arXiv CS.AI",
  "title": "Paper Title",
  "published_at": "2026-05-13T10:00:00Z",
  "category": "llm",
  "summary": "...",
  "rank_score": 0.85,
  "url": "https://..."
}]
```

## Notes

- Run `amp sync` first if data might be stale (last sync > 1 hour ago)
- Use `--json` flag for programmatic consumption
- The tool does NOT make value judgments — it only aggregates and ranks
- Data stored in `.amp/data.db` (SQLite) at the MindMagnifier project root
