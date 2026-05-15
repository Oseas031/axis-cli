---
name: vigil
description: Cross-session work tracking. Use at session start (resume) and when managing work items.
tags: [workflow, tracking, automation]
---

# axis vigil — 跨会话工作追踪

## 核心用法

```bash
# 新会话第一步（必做）— 同时自动执行 triage
axis vigil resume

# 首次使用：安装 git hook（之后 commit 自动标记完成）
axis vigil install-hook

# 创建工作项
axis vigil add "title" --priority P1 --tag arch --origin manual

# 开始/完成
axis vigil start <id>
axis vigil done <id> --commit <hash>

# 查询
axis vigil list [--priority P0] [--tag arch] [--status in_progress] [--json]
axis vigil show <id>

# 手动维护（resume 时已自动执行）
axis vigil triage
```

## 自动化行为

- **git hook 自动完成**：commit message 中写 `vigil:<id>` → hook 自动调用 done（需先 `axis vigil install-hook`）
- **resume 自动 triage**：每次 resume 静默执行归档/stale/升级，无需手动调用 triage
- **triage 规则**：pending >7天 → stale；被 ≥3 项依赖 → 升为 P0；completed >48h → 归档
- **无需手动标记完成**：正常路径由 git hook 处理
- **竞态防护（Lock）**：`start` 获取文件锁，`done` 释放。多 AI 会话不会同时操作同一 item

## 竞态防护

`axis vigil start <id>` 在 `.axis/vigil/locks/<id>.lock` 写入锁文件（JSON: holder/PID/started_at）。

行为：
- 另一个活进程已持有锁 → `start` 拒绝，报告持有者 PID
- 持有者进程已死（stale lock）→ 自动回收，新会话可接管
- `done` 完成时自动释放锁
- `resume`/`list` 输出中 🔒 标记表示 item 被活进程锁定

AI 会话规则：看到 🔒 标记的 item，不要尝试 start，选择其他 pending item。

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
- Locks: `.axis/vigil/locks/<id>.lock`
- Archive: `.axis/vigil/archive/YYYY-MM.json`

## 与 CLAUDE.md 的关系

§0 rule #13 指向本文件。§15 External Tools Reference 列出 vigil。
详细设计见 `docs/specs/vigil/`。
