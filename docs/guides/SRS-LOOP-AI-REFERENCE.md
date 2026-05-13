# SRS Loop — AI Collaboration Reference

> 展开自 CLAUDE.md §0（作者工作方法论）
> AI 可读摘要。当作者说"按我的工作流"或"用 SRS Loop"时，指的就是这套流程。
> 本体论层：`docs/architecture/dialectical-development-methodology.md`
> 完整操作层：`docs/guides/dialectical-development-methodology.md (docs/architecture/)`

---

> 核心三元组定义见 CLAUDE.md §0。本文档只展开操作层细节。

------|------|---------|
| **Construct**（对象化） | 意图→客观存在 | 执行生成 |
| **Constraint**（规定性） | 划定质的界限 | 在边界内工作 |
| **Judge**（扬弃） | 保留内核，否定偏差 | 辅助判断 |

---

## 三阶段结构

### Phase I — Objectification（生成）

| 步骤 | 操作 |
|------|------|
| A0 Posture | 确认性质：修 bug？长出新能力？扬弃旧结构？ |
| A1 Externalize | 把想法写成文档，无论多粗糙 |

**退出条件**：想法已成为可被规定的客观文本。

### Phase II — Determinateness（规定）

| 步骤 | 操作 |
|------|------|
| A2 Inventory | 列已有什么（参考 current-progress.md） |
| A3 Diagnose | 列缺什么、错什么、违背什么（基于代码事实，不靠记忆） |
| A4 Realign | 映射到四命题 + 六 first principles |
| A5 Minimize | 奥卡姆三问 → 选最小可行步骤 |

**退出条件**：下一步行动的边界清晰到不可误解。

### Phase III — Sublation（扬弃）

| 步骤 | 操作 |
|------|------|
| A6 Execute | 实施最小步骤，`go test -race ./...` |
| A7 Distill | 四栏复盘：保留 / 修正 / 剔除 / 沉淀 |
| A8 Sublate | 把"沉淀"写回 CLAUDE.md / spec / workflow |

**退出条件**：新规则已写回系统。

### 失败回退

- A6 失败 → 回 Phase II（A5 重选最小单元）
- A7 发现方向错 → 回 Phase I（A1 重新外化）
- A8 发现原则需修正 → 更高阶 Sublation，更新 `docs/architecture/`

---

## AI 行为约束

1. **不跳过 Phase II 直接执行**。收到任务后先确认边界。
2. **Phase 转换时主动声明**。说明当前 Phase 和退出条件是否满足。
3. **工作终点是规则更新（A8）**，不是代码合并。
4. **主要矛盾优先**。A5 时问"这解决主要矛盾还是次要矛盾？"
5. **实践检验**。设计必须在一个循环内可 `go test` 验证。
6. **阶段论**。不用终态理想批判当前阶段的妥协。

---

## AI 协作指南

| 作者在做什么 | Phase | AI 应该做什么 |
|---|---|---|
| 写分析文档、战略报告 | I | 帮助结构化，不催促执行 |
| 列已有能力、查进度 | II | 读代码/文档，给准确现状 |
| 描述 bug 或缺口 | II | 定位根因，映射到原则 |
| 问"现在该做什么" | II | 奥卡姆三问 + 候选最小步骤 |
| 要求实现某功能 | III | 实施 + 测试 + 报告验证 |
| 说"帮我复盘" | III | 四栏格式输出 |
| 说"写回规则" | III | 具体文件修改建议 |

### 特殊说明

- 作者的 strategy/ 报告是 A1 外化工具，**不是任务清单**。问"A5 结论是什么？"
- workflow/contract/spec 是**过渡性脚手架**，最终会被 Agent 内化和废弃。
- 遇到原则冲突时，用辩证方法论裁决（主要矛盾、阶段论、对立统一）。

---

## 四命题

```
More Context         → Agent 主动查询组装上下文
More Action          → 执行、验证、修正、生成后续任务
Zero Control         → 系统提供契约和可观测性，不规定行动路径
Controllable Evolution → 自修改必须在可审计、可回滚边界内
```
