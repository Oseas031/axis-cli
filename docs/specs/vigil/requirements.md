# Vigil Requirements

> 实现 agent-native-first-principles.md P1（CLI First）+ CLAUDE.md §11（演化原则：审计）

**Status**: In Progress
**Date**: 2026-05-14
**Name**: vigil（守夜）— 跨会话工作追踪，系统替你守着未完成之事

## Problem Statement

当前工作追踪是手动 markdown 文件（`WORKFLOW-HUMAN/pending-*.md`）。每次新会话需要：
1. 读文件了解状态（慢，容易遗漏）
2. 手动编辑标记完成（容易忘）
3. 手动创建新待办（judge/test 发现的问题不会自动变成待办）

AI Agent 每次会话都是新的——没有跨会话记忆。工作连续性完全依赖文件状态。

## Design Decision

**自动化基础设施，不是 CRUD 工具**。核心原则：
- 系统自动维护状态，Agent 查询结果
- 自动化做秘书工作（创建、标记完成、归档），不做上司工作（不阻止、不强制顺序）
- Agent 保留完全掌控权（可手动覆盖任何自动操作）

## Functional Requirements

### FR1: resume — 新会话第一条命令

`axis vigil resume` 输出：
- 上次会话的 in_progress 项（未完成的工作）
- 最近完成的项（上下文恢复）
- 最高优先级的 pending 项（下一步建议）

这是 Agent 新会话的**唯一入口**——不需要读 markdown、不需要盘点。

### FR2: add — 显式创建

`axis vigil add "title" --priority P1 --tag arch --origin manual`

必填：title
可选：priority（默认 P1）、tag、origin、depends_on、notes

### FR3: start/done — 状态流转

- `axis vigil start <id>` — pending → in_progress（记录开始时间）
- `axis vigil done <id> [--commit <hash>]` — → completed（记录完成时间）
- 正常路径：git hook 自动 done，手动为兜底

### FR4: list — 筛选查询

`axis vigil list [--priority P0] [--tag arch] [--status open] [--since 7d]`

默认只显示 active 项（pending + in_progress）。

### FR5: show — 详情 + 溯源

`axis vigil show <id>` 输出完整信息 + origin 溯源 + 关联 commits。

### FR6: triage — 自动优先级维护

`axis vigil triage` 执行：
- 被 ≥3 项依赖的 → 升为 P0
- 超过 7 天未变更 → 标记 stale
- 已完成超过 48h → 自动归档

可由 hook 定时触发，也可手动调用。

### FR7: git post-commit hook — 自动完成

解析 commit message 中 `vigil:<id>` 标记 → 自动调用 done。

### FR8: ingest — 系统产出自动灌入（P2，预留接口）

接收来自 judge/test/followup 的自动创建请求。当前只定义接口，不实现自动触发。

## Non-Functional Requirements

### NFR1: 纯本地，零外部依赖

数据存储：`.axis/vigil/items.json` + `.axis/vigil/archive/`

### NFR2: JSON 输出

`--json` flag 输出机器可解析的 JSON。默认人类可读。

### NFR3: 幂等

重复 done 同一个 id 不报错。重复 start 不报错。

### NFR4: 不阻塞

任何 vigil 命令失败不应阻塞 git commit 或其他 axis 命令。

## Acceptance Criteria

- `axis vigil resume` 在无数据时输出 "No active work. Use: axis vigil add"
- `axis vigil add "test" --priority P0` 创建项并返回 id
- `axis vigil start <id>` + `axis vigil done <id>` 完整流转
- `axis vigil list --status in_progress` 只显示进行中
- `axis vigil triage` 标记 stale 项
- git commit with `vigil:xxx` 自动标记完成
- `go test -race ./internal/vigil/...` 通过
- `go test -race ./cmd/axis/...` 通过
