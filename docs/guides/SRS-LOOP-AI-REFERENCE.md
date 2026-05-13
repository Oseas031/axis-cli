# SRS Loop — AI Collaboration Reference

> 这是作者工作方法论的 AI 可读摘要。当作者说"按我的工作流"或"用 SRS Loop"时，指的就是这套流程。
> 完整版：`docs/guides/YOUR_IMPLICIT_METHODOLOGY.md`
> 所有路径均以**项目根目录**为基准。

---

## 什么是 SRS Loop

**Spec-Reflect-Sublate 循环**：作者从哲学训练中继承的工程方法论。
每个工作单元都经历 A0→A8 的完整循环，而不是"写代码→提交"。

---

## 完整 8 阶段

```
A0. Posture（姿态）
    确认本次工作的性质：修 bug？长出新能力？扬弃旧结构？
    → 不同姿态决定后续步骤的侧重点

A1. Externalize（外化）
    把当前想法写成文档（.md），无论多粗糙
    → 写作是思想生成器，不是记录工具

A2. Inventory（盘点）
    列出已有什么：已完成的模块、已通过的测试、已稳定的接口
    → 参考 docs/status/current-progress.md

A3. Diagnose（诊断）
    列出缺什么、错什么、违背了什么原则
    → 用 bug/缺口 反向检验 first principles

A4. Realign（归位）
    把诊断映射到四命题 + 六 first principles
    → 参考 docs/architecture/agent-native-first-principles.md
    → 每个问题必须能对应到至少一条原则

A5. Minimize（最小化）
    奥卡姆三问：① 现在必须做吗？② 最小可行是什么？③ 会破坏 Scaffold-to-Self 吗？
    → 参考 workflow/occams-razor-architecture-simplification.md
    → 选出今天能完成的最小一步

A6. Execute & Verify（执行验证）
    实施 A5 选出的最小步骤，跑测试，记录验证结果
    → go test -race ./...
    → 这是唯一真正"写代码"的阶段

A7. Distill（萃取）
    写复盘，分四栏：
    - 保留：哪些做法有效，继续用
    - 修正：哪些做法需要调整
    - 剔除：哪些做法是错的，停止
    - 沉淀：哪些经验要变成规则

A8. Sublate（扬弃）
    把 A7"沉淀"栏的内容写回：
    - CLAUDE.md（约束规则）
    - workflow/entry.md（工作流入口）
    - 对应 spec 的 requirements/design/tasks
    → 工作不以"代码合并"为终点，以"规则更新"为终点
```

---

## AI 协作指南

### 当作者说"按 SRS Loop 走"时

按 A0→A8 顺序推进，不要跳步。尤其不要跳过 A4（归位）直接到 A6（执行）。

### 当作者说"帮我做 A5"时

给出奥卡姆三问的答案，然后列出 1-3 个候选的"最小可行步骤"，让作者选择。

### 当作者说"帮我做 A7"时

按四栏格式输出复盘：保留 / 修正 / 剔除 / 沉淀。

### 当作者说"帮我做 A8"时

把 A7 的"沉淀"栏转化为具体的文件修改建议，指出要写入哪个文件的哪个位置。

### 关于 A1（外化）的特殊说明

作者写的 strategy/ 报告和长篇分析文档，**大部分不需要被完整执行**。
它们的作用是帮作者看清下一步。AI 不应该把这些文档当作任务清单全部实施。
应该问："这份文档的 A5 结论是什么？"

### 关于"脚手架"的态度

workflow / contract / permission / spec 这些结构是**过渡性的**。
作者明确说过它们最终会被 Agent 内化、改写、废弃。
AI 不应该把维护这些结构本身当作目标，而应该把"让系统能自我改写"当作目标。

### 关于原则优先级

遇到实现选择时，优先顺序：
1. 不违背 CLAUDE.md §1 绝对禁止项
2. 对应到 agent-native-first-principles.md 的某条原则
3. 奥卡姆最小化（A5）
4. 技术最优解

---

## 四命题（作者的核心信念）

```
More Context    → 系统提供查询基础设施；Agent 主动组装上下文
More Action     → 执行、组合、验证、修正、生成后续任务
Zero Control    → 系统提供契约和可观测性，不规定单一行动路径
Controllable Evolution → 自举、自生成、自修改必须在可审计、可回滚边界内
```

---

## 快速判断：这个任务处于哪个阶段？

| 作者在做什么 | 对应阶段 | AI 应该做什么 |
|---|---|---|
| 写分析文档、战略报告 | A1 外化 | 帮助结构化思路，不要催促执行 |
| 列已有能力、查进度 | A2 盘点 | 读 current-progress.md，给出准确现状 |
| 描述 bug 或缺口 | A3 诊断 | 帮助定位根因，映射到原则 |
| 问"这违背了哪条原则" | A4 归位 | 引用具体原则条目，给出判断 |
| 问"现在该做什么" | A5 最小化 | 给出奥卡姆三问答案 + 候选最小步骤 |
| 要求实现某功能 | A6 执行 | 实施，跑测试，报告验证结果 |
| 说"帮我复盘" | A7 萃取 | 四栏格式输出 |
| 说"写回规则" | A8 扬弃 | 给出具体文件修改建议 |
