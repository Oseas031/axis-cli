---
name: improve-codebase-architecture
description: Find deepening opportunities — turn shallow modules into deep ones. Improve testability and AI-navigability. Use when user wants architecture review, refactoring opportunities, or says "架构审查"/"深化模块"/"improve architecture".
tags: [engineering, architecture, deep-modules]
source: mattpocock/skills
source_version: 2026-05-15
---

# Improve Codebase Architecture

> 来源：mattpocock/skills (MIT)，适配 Axis 语境。

发现架构摩擦，提出 **深化机会** — 将浅模块变为深模块的重构。

## 术语（严格使用）

- **Module** — 任何有接口和实现的东西（函数、struct、package、slice）
- **Interface** — 调用者必须知道的一切：类型、不变量、错误模式、顺序、配置
- **Implementation** — 内部代码
- **Depth** — 接口处的杠杆：小接口后面大量行为 = 深 = 高杠杆
- **Seam** — 接口所在的位置；可以不编辑原地就改变行为的地方
- **Adapter** — 在 seam 处满足接口的具体实现
- **Deletion Test** — 删除模块后复杂度消失 → pass-through；复杂度散布到 N 个调用者 → 有价值

## 原则

- **接口即测试面**
- **一个 adapter = 假设性 seam。两个 adapter = 真实 seam。**
- 用 Axis `docs/architecture/semantic-boundaries.md` 的边界定义指导 seam 识别

## 流程

### 1. 探索

先读 CLAUDE.md §6 语义边界和相关 BOUNDARY.md。然后有机地探索代码库，注意摩擦点：

- 理解一个概念需要在多个小模块间跳转？
- 模块是否 **浅** — 接口几乎和实现一样复杂？
- 纯函数是否只为可测试性提取，但真正的 bug 藏在调用方式中（无 locality）？
- 紧耦合模块是否跨 seam 泄漏？
- 哪些部分未测试或难以通过当前接口测试？

对可疑的浅模块应用 **Deletion Test**。

### 2. 呈现候选

编号列表，每个候选：

- **Files** — 涉及哪些文件/模块
- **Problem** — 当前架构为何产生摩擦
- **Solution** — 纯文字描述变更
- **Benefits** — 用 locality 和 leverage 解释，以及测试如何改善

**不要提议接口。** 问用户："你想探索哪个？"

### 3. Grilling Loop

用户选择后，进入拷问式对话：约束、依赖、深化模块的形状、seam 后面是什么、哪些测试能存活。

决策结晶时的副作用：
- 命名了新概念 → 记录到相关文档
- 用户拒绝候选且理由有分量 → 提议 ADR
- 需要探索替代接口 → 展示选项

## Axis 特定

- 探索前必读 `internal/*/BOUNDARY.md`
- 遵循 CLAUDE.md §6 语义边界（每个模块的"不得做"清单）
- 深化建议不得违反 §1 绝对禁令
- 结构性修改走 Staged Evolution Protocol（CLAUDE.md §5）
