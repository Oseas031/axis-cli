---
type: validation
status: active
created: 2026-05-16
title: "E2E Combat 001: Cost Budget via axis ask --submit"
---

# E2E Combat Test 001: Cost Budget via axis ask --submit

> 目标：用 `axis ask --submit` 一句话触发 cost_budget 功能实现，全程无人工干预

## Session Metrics

| 指标 | 数值 |
|------|------|
| 时长 | ~3min 40s (10:33:04 → 10:36:44) |
| Provider | MiniMax-M2.7 (mm-official) |
| Prompt | "为AgentTask添加cost_budget字段，当token消耗超过预算80%时自动降级执行" |
| 任务ID | ask-20260516-103304-5480 |
| 结果 | failed - iteration budget exhausted (20 turns) |
| 人工介入 | 0 次 |
| Agent 修改文件 | 1 (internal/types/types.go) |

## 执行流程

```
axis ask --submit "为AgentTask添加cost_budget字段..."
  → intent parser (deterministic)
  → control client → orchestrator (port 9091)
  → dispatcher → agent executor
  → LLM (MiniMax-M2.7) → tool use → file edit
  → ⏱ 20 turns exhausted → FAILED
```

## 完成的工作

Agent 在 20 轮内完成了部分类型定义层面的工作：

1. **AgentTask.CostBudget 字段** — 完整添加到 struct
   ```go
   // CostBudget defines the maximum USD cost allowed for this task.
   // When accumulated cost exceeds 80% of this budget, the executor
   // will automatically switch to a lower-cost model or approach.
   // A value of 0 means unlimited budget.
   CostBudget float64 `json:"cost_budget"`
   ```

2. **SLAKeyTokenBudget 常量**
   ```go
   SLAKeyTokenBudget = "sla.token_budget"
   ```

3. **错误码**
   ```go
   ErrTokenBudgetExhausted ErrorCode = "TOKEN_BUDGET_EXHAUSTED"
   ErrCostBudgetExceeded   ErrorCode = "COST_BUDGET_EXCEEDED"
   ```

## 未完成的工作

- [ ] CostTracker / budget package 的实现与集成
- [ ] Dispatcher 中的 CostBudget 检查（80% 阈值自动降级）
- [ ] 自动降级逻辑（切换低消耗模型）
- [ ] 测试
- [ ] 构建验证

## 失败分析

| 因素 | 影响 |
|------|------|
| 20 轮迭代限制 | 完成类型定义后已无剩余轮次完成后续逻辑 |
| MiniMax-M2.7 速度 | 平均每轮 ~11s，20 轮 ≈ 220s，非主要瓶颈 |
| 单文件修改 | Agent 只修改了 types.go，未修改 dispatcher.go 等 |
| 任务粒度 | 单一 prompt 对于多文件 feature 过大，subtask 分解未触发 |

## 与上一轮（人工多轮 Session）对比

| 维度 | 本轮（axis ask --submit） | 上一轮（人工多轮） |
|------|--------------------------|-------------------|
| 时长 | 3min 40s | 2h 36min |
| Commits | 0 (未触发提交) | 12 |
| 代码产出 | 1 file, ~10 行 | ~1,400 行 + 测试 |
| 人工介入 | 0 | 0 (但 Prompt 多次) |
| 失败模式 | Iteration budget | N/A (完成) |

## 结论

1. **Pipeline 端到端可用**：`ask --submit` → orchestrator → agent executor → tool use 链路通
2. **Agent 能理解任务并生成代码**：类型定义部分正确且语义完整
3. **20 轮迭代限制是主要瓶颈**：对于多文件 feature 实现，需要更大的 budget 或任务分解
4. **Prompt 设计需改进**：单一长句包含了字段+降级逻辑，Agent 倾向于线性执行而非并行分解

## Next Steps

- [ ] 增加 iteration budget（20 → 50+）或改为动态预算
- [ ] 添加 subtask 分解能力（Phase III A6 Execute 的 subagent 派发机制）
- [ ] 考虑 `axis run --background` 异步执行模式
- [ ] 换用更有能力的模型（DeepSeek V4 Flash）重试
