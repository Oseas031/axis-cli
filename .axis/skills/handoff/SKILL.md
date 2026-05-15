---
name: handoff
description: Compact the current conversation into a handoff document for another agent/session to continue. Use when user says "handoff"/"交接"/"下次继续"/"收工", or session is ending with unfinished work.
tags: [productivity, cross-session]
source: mattpocock/skills
source_version: 2026-05-15
---

# Handoff

> 来源：mattpocock/skills (MIT)，适配 Axis 语境。

将当前对话压缩为交接文档，让新 Agent/会话能无缝继续。

## 流程

1. 写交接文档到 `.axis/handoffs/handoff-<timestamp>.md`
2. 建议下次会话使用的 skills
3. 不重复已捕获到其他 artifact 中的内容（specs、commits、vigil items）— 引用路径即可

## 交接文档结构

```markdown
# Handoff: <简短标题>

## 上下文
<项目当前状态，1-3 句>

## 已完成
- <本次完成的事项，引用 commit/file>

## 未完成
- <剩余工作，按优先级排序>

## 关键决策
- <本次做出的重要决策及理由>

## 下次建议
- 使用 skills: <推荐的 skills>
- 先做: <最高优先级的下一步>
- 注意: <陷阱/已知问题>

## 相关文件
- <路径列表>
```

## 与 vigil 的关系

- handoff 是**一次性快照**（会话结束时的状态）
- vigil 是**持久化追踪**（跨多个会话的工作项）
- 两者互补：handoff 捕获会话上下文，vigil 追踪工作项
- 如果未完成项还没在 vigil 中 → 同时 `axis vigil add`

## Axis 特定

- 如果用户传了参数，视为下次会话的焦点方向
- 交接文档不进 git（`.axis/handoffs/` 在 .gitignore 中）
- 新会话第一步仍是 `axis vigil resume`（handoff 是补充，不替代 vigil）
