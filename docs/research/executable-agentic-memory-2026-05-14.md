# Executable Agentic Memory：对 Axis Memory 系统的启示

> 基于 arXiv:2605.12294 — "Executable Agentic Memory for GUI Agent"
> 作者: Qin et al., 2026-05-12

## 1. 核心机制

| 组件 | 作用 |
|------|------|
| **Knowledge Graph** | GUI 操作序列编码为有向图：节点=UI 状态，边=动作组（多步压缩为单边） |
| **State-aware DFS + Action-group Mining** | 自动发现重复多步操作，压缩为原子 action-group |
| **Q-function Model** | 轻量价值模型，为 KG 边估算 Q 值（bias-consistency 已证明） |
| **MCTS over KG** | 在 KG 上做 Monte Carlo Tree Search，Q-model 引导搜索方向 |

**范式转换**：规划从「每步重新生成」→「在结构化记忆上检索+执行」。

## 2. 关键数据

- **+19.6% accuracy** vs UI-TARS-7B (AndroidWorld)
- **6× token cost reduction** vs GPT-4o
- **2.8s average latency**

## 3. 对 Axis 的启示

### 当前状态

Axis memory 是**被动的**：
- `patterns/` — 被检索的失败模式
- `principles/` — 被注入 prompt 的规则
- `narrative/` — 被遗忘的叙事

Agent 每次从零规划，不复用已验证的执行路径。

### EAM 的 memory 是**主动的**

KG 本身就是可执行的规划图。Agent 沿图行走而非自由生成。

### 可借鉴

1. **操作序列图化**：成功的 task execution traces 编码为 DAG，相似任务复用路径
2. **Action-group 压缩**：ToolTrace 序列挖掘重复 pattern，压缩为复合动作
3. **Value-guided 路径选择**：用历史成功率为边赋权，替代纯 LLM 规划

### 不能直接借鉴

1. **GUI 状态可枚举**：Axis 面对开放代码/系统状态空间，不可穷举
2. **完整 MCTS**：代码任务分支因子远大于 GUI
3. **离线 DFS 构建**：Axis 任务空间无法预先探索

## 4. 可行动建议

| 优先级 | 行动 | 模块 |
|--------|------|------|
| P1 | 将成功的 ToolTrace 序列持久化为 execution trace graph | `internal/memory/` |
| P1 | context assembly 时查询 trace graph，相似任务注入已验证路径 | `internal/contextpack/` |
| P2 | judgement 结果反馈更新 graph 边权重（success-rate weighting） | `internal/agent/judgement/` |

**一句话**：Axis 应从「每次从零规划」演进为「优先复用已验证的执行路径」。EAM 的 KG+Q-value 降维适配为 trace graph + success-rate weighting。

---

*生成日期: 2026-05-14 | 论文: arXiv:2605.12294*
