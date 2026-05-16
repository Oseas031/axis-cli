---
name: prototype
description: Build a throwaway prototype to answer a design question — runnable terminal app for logic/state questions, or multiple UI variations. Use when user says "prototype"/"试试看"/"验证一下"/"探索设计".
tags: [engineering, exploration, throwaway]
source: mattpocock/skills
source_version: 2026-05-15
---

# Prototype

> 来源：mattpocock/skills (MIT)，适配 Axis 语境。

Prototype = **回答问题的一次性代码**。问题决定形状。

## 选择分支

- **"这个逻辑/状态模型对吗？"** → 建一个小型交互式终端 app，推动状态机走过难以纸上推理的 case
- **"这应该长什么样？"** → 生成几个截然不同的变体，可切换对比

不确定时，看周围代码：backend module → logic 分支；page/component → UI 分支。

## 通用规则

1. **从第一天就是一次性的，且明确标记。** 放在它 prototype 的模块旁边（上下文明显），但命名让人一眼看出不是生产代码。
2. **一条命令运行。** `go run ./prototype/xxx` 或项目已有的 task runner。
3. **默认无持久化。** 状态在内存中。
4. **跳过打磨。** 无测试、无错误处理（除了让它能跑的）、无抽象。
5. **暴露状态。** 每次操作后打印完整相关状态。
6. **完成后删除或吸收。** 不让它在 repo 中腐烂。

## 完成时

唯一值得保留的是**答案**。捕获到持久位置：

- commit message
- ADR（如果是架构决策）
- vigil item（如果产生了后续行动）

连同它回答的问题一起记录。然后删除 prototype 代码。

## Axis 特定

- Prototype 代码放 `.scratch/prototype-<name>/`（不进 git）
- 如果验证了设计决策 → 走 CLAUDE.md §5 Spec-First 协议
- 如果产生了后续行动 → `axis vigil add`
