# Aspirational Directions

> 展开自 CLAUDE.md §13.4 模式 A（理论超前实践）
> 本文档保留方向性命题和激活条件。不含具体实现设计。

---

## 方向性命题

### D1: 结构化错误码

**方向**: Agent 和脚本应该能通过结构化错误码做自动化决策，而非解析人类可读的错误消息。

**激活条件**: 当 ≥3 个 CLI 命令需要被脚本/Agent 解析错误类型时。

**来源**: docs/deprecated/error-code-conventions.md

---

### D2: 状态可重放

**方向**: 系统状态变更应该可以从日志重放，支持任意时间点的状态重建。

**激活条件**: 当需要调试跨多个任务的状态不一致问题时，或当需要实现 checkpoint/restore 时。

**来源**: docs/architecture/kernel-abstraction-model.md Infrastructure Layer

---

### D3: 自动化 Spec Promotion

**方向**: Spec 状态转换应该产生结构化事件（spec.promoted / spec.demoted），支持自动化审计和回溯。

**激活条件**: 当 spec 数量 >20 且手动状态管理成为瓶颈时，或当需要自动化 spec 状态报告时。

**来源**: docs/architecture/spec-lifecycle-conventions.md Promotion Semantics

---

### D4: 语义索引

**方向**: Context assembly 应该支持基于语义相似度的检索，而非仅靠规则匹配。

**激活条件**: 当规则匹配的 context recall 准确率不足以支撑 Agent 决策质量时。

**来源**: docs/architecture/kernel-abstraction-model.md Index + Retrieval

---

> 本文档属于 CLAUDE.md §13.1 渐进条款，可被实践反馈修正或扬弃。
