---
name: devils-advocate
description: "Mechanical Devil's Advocate: spawn an isolated subagent to critique the current work product. Prevents Sovereignty Gap (self-sycophancy). Use when user says \"唱反调\"/\"反调\"/\"devil\"/\"critique\", or auto-triggered at Phase III end when mechanical conditions are met."
triggers:
  - user
  - model
tags: [methodology, quality, multi-agent]
---

# Devil's Advocate

> 展开自 CLAUDE.md §0 rule #10。主上下文禁止自行扮演反调角色。

## 触发词（主动路径）

用户消息包含以下任一词时立即执行：

```
唱反调 | 反调 | devil | critique | grill
```

## 机械条件（自动路径）

Phase III 结束时（A7 Distill 之前），检查本次工作：

```yaml
auto_trigger_predicate:  # 任一为 true 即触发
  - new_files: ">=2"
  - changed_modules: ">=3"       # 不同 internal/ 子目录
  - new_exported_types: ">=3"
```

满足时主 Agent 必须 spawn 反调 subagent，不得跳过。

## 执行协议

### 1. Spawn 反调 subagent

使用 kiro subagent 工具，模板如下：

```
subagent(
  name: "devils-advocate",
  role: "axis",
  prompt: """
你是一个严格的批判者。你的任务是找出以下工作产出的问题。

## 工作摘要
{work_summary}

## 变更文件
{changed_files}

## 批判框架（必须逐条回答）
1. 它声称解决了什么？实际解决了吗？有没有自欺欺人？
2. 它遵守了哪些规则？遵守规则的同时是否产出了垃圾？
3. 遗漏了什么主要矛盾？什么问题被回避了？
4. 最弱的环节是什么？如果只能砍掉一半代码，砍哪些？
5. 6 个月后回看，什么会让人后悔？

## 约束
- 只输出批判，不输出赞美
- 每条批判必须具体到文件/函数/行为，不允许泛泛而谈
- 如果找不到问题，说明为什么找不到（这本身是一个信号）
"""
)
```

### 2. 主 Agent 回应

subagent 返回后，主 Agent 必须：

- 逐条显式回应（接受/反驳/记录为 TODO）
- 不得忽略任何一条批判
- 接受的批判立即修复或创建 vigil 待办

### 3. 记录

在 Phase III 反馈闭环中声明：
```
Devil's Advocate 已执行。接受 N 条，反驳 M 条，TODO K 条。
```

## 隔离要求

反调 subagent 的 IsolationPolicy：

| 字段 | 值 | 原因 |
|------|-----|------|
| InheritMemory | false | 防止被主 Agent 的推理路径污染 |
| InheritContext | false | 只看最终产出，不看过程 |
| RequireProviderDiversity | true | 防止 Kinship 同族谄媚 (arXiv:2605.10698) |
| CoTIsolation | true | 不暴露主 Agent 的 CoT 给反调者 |

## 不做的事

- 不替代 rule #12 自审（自审是快速轻量的第一道防线）
- 不阻塞用户（subagent 结果返回后才继续，但不需要用户确认）
- 不自动修改代码（只有建议权）
- 主上下文不得在没有 spawn subagent 的情况下自行输出"反调结果"
