---
name: vigil
description: Cross-session work tracking. Use at session start (resume) and when managing work items.
tags: [workflow, tracking, automation]
---

# axis vigil — 跨会话工作追踪

## 核心用法

```bash
# 新会话第一步（必做）
axis vigil resume

# 创建工作项
axis vigil add "title" --priority P1 --tag arch --origin manual

# 开始/完成
axis vigil start <id>
axis vigil done <id> --commit <hash>

# 查询
axis vigil list [--priority P0] [--tag arch] [--status in_progress] [--json]
axis vigil show <id>

# 自动维护（标记 stale、升级优先级、归档）
axis vigil triage
```

## 自动化行为

- **git hook 自动完成**：commit message 中写 `vigil:<id>` → hook 自动调用 done
- **triage 规则**：pending >7天 → stale；被 ≥3 项依赖 → 升为 P0；completed >48h → 归档
- **无需手动标记完成**：正常路径由 git hook 处理

## 状态流转

```
pending → in_progress → completed → (48h后自动归档)
    ↓
  stale (>7天未动)
```

## commit message 格式

```
fix(cli): unify root resolution vigil:vigil-a3f

feat(vigil): core data layer vigil:vigil-b7c vigil:vigil-d2e
```

多个 ID 用空格分隔，每个都会被自动标记完成。

## 优先级

- **P0**：阻塞其他工作，必须立即处理
- **P1**：中期目标，本周/下周完成
- **P2**：长期演进，条件成熟时再做

## 数据位置

- Active items: `.axis/vigil/items.json`
- Archive: `.axis/vigil/archive/YYYY-MM.json`

## 与 CLAUDE.md 的关系

§0 rule #13 指向本文件。§15 External Tools Reference 列出 vigil。
详细设计见 `docs/specs/vigil/`。
