---
type: session-report
date: 2026-05-16
tags: [e2e, budget, subagent, refactor]
---

# Session Report: 2026-05-16 — E2E Rework + Budget & Subagent Improvements

## Summary

本会话完成了三个主要目标：(1) 重跑 E2E Combat 001，验证 `axis ask --submit` 端到端链路；(2) 根据 E2E 暴露的瓶颈，实施迭代预算改进；(3) 接入同步 Subagent 派发能力。

---

## Phase 1: E2E Combat 001 — 重跑

### 背景

上一轮 E2E 实战（vigil-2ff991）目标是「用 `axis ask --submit` 一句话触发 cost_budget 功能实现，全程无人工干预」，但实际执行是 2.5 小时的人工多轮 Prompt 会话，完全偏离测试定义。

### 执行过程

| 时间 | 事件 |
|------|------|
| 10:27 | `axis ask --dry-run` 验证 intent parser |
| 10:28 | 首次提交 → MiniMax mm 额度不足（$0） |
| 10:32 | 切到 mm-official（MiniMax M2.7） |
| 10:33 | `axis ask --submit` 提交成功，task running |
| 10:33→10:36 | Agent 执行 20 轮 |
| 10:36:43 | Agent 修改 `internal/types/types.go`（CostBudget 字段+常量+错误码） |
| 10:36:44 | `iteration budget exhausted (20 turns)` — 失败 |

### 关键发现

1. **Pipeline 端到端可用**：`ask --submit` → orchestrator → agent executor → tool use → file write 链路完整通
2. **Agent 能正确理解任务**：生成的 CostBudget 字段注释语义完整（80% threshold 被保留）
3. **20 轮不够**：完成类型定义后已无剩余轮次处理 dispatcher 集成
4. **Spawn 工具存在但未使用**：agent 不知道可以用 spawn 分解任务

### 战斗报告

更新写入 `docs/validation/e2e-combat-001.md`

---

## Phase 2: 代码回退

将上次会话中人工写入的 cost_budget 代码从以下文件移除：

| 文件 | 操作 |
|------|------|
| `internal/types/types.go` | 移除 `CostBudget` 字段 |
| `internal/types/types_test.go` | 移除 `TestAgentTask_CostBudgetJSON` |
| `internal/kernel/dispatcher/dispatcher.go` | 移除 budget import、`costTracker` 字段、`SetCostTracker` 方法、预算检查块 |
| `internal/kernel/budget/*` | 标记为死代码（无引用，sandbox 阻止删除） |

构建 ✅ | types 测试 ✅ | dispatcher 测试 ✅

---

## Phase 3: 迭代预算改进 + Subagent 派发

### 3a. Iteration Budget — 默认 50 + Metadata 可覆盖

**改动：**

- `internal/model/multiturn/loop.go` — `MaxIterations` 默认 20 → 50
- `internal/agent/llm_executor.go` — `maxIter` 默认 20 → 50
- `internal/agent/llm_executor.go` — 新增 metadata 覆盖逻辑：`task.Metadata["axis.max_iterations"]` 可覆盖默认值
- `cmd/axis/main.go` — `WithMaxIterations(20)` → `WithMaxIterations(50)`

**效果：**
- E2E 任务可以从 20 轮提升到 50 轮
- 不同任务可通过 metadata 单独配置迭代预算
- 非法值（负数/0）回退到默认 50

### 3b. Synchronous Subagent Dispatch

**改动：**

- `internal/actor/spawn_executor.go` — `Execute()` 返回 `(map[string]any, error)` 替代 `error`，直接返回 worker 实际产出
- `internal/actor/llm_adapter.go` — `MaxTurns` 配置化（默认 15），`SpawnExecutorConfig` 新增 `WorkerMaxTurns`
- `internal/kernel/orchestrator/orchestrator.go` — `execFn` 透传 spawn 结果给调用方
- `cmd/axis/main.go` — 系统 prompt 追加分解指导

**效果：**
- Agent 调用 `spawn` 工具后，阻塞等待 worker 完成并返回实际结果
- Worker 有独立的迭代预算（默认 15 轮）
- Worker 无法递归 spawn（scope 排除了 spawn 工具）

---

## 验证结果

```
go build ./...                                   ✅
go test ./internal/agent/...                     ✅ (16.1s)
go test ./internal/actor/...                     ✅ (0.3s)
go test ./internal/model/multiturn/...            ✅ (0.2s)
go test ./internal/kernel/orchestrator/...        ✅ (0.9s)
go test ./internal/kernel/dispatcher/...          ✅ (6.8s)
go test ./internal/types/...                      ✅ (0.1s)
go test ./cmd/axis/...                            ✅ (13.4s)
```

**10 个包全部通过，零失败。**

---

## 文件改动清单

| 文件 | 改动类型 |
|------|----------|
| `internal/model/multiturn/loop.go` | 默认值变更 |
| `internal/agent/llm_executor.go` | 默认值变更 + metadata 覆盖逻辑 + import 追加 |
| `cmd/axis/main.go` | WithMaxIterations(50) + 系统 prompt 追加 |
| `internal/actor/spawn_executor.go` | Execute 签名变更 + 返回实际结果 |
| `internal/actor/llm_adapter.go` | MaxTurns 配置化 |
| `internal/actor/llm_adapter_test.go` | 测试适配新 Execute 签名 |
| `internal/kernel/orchestrator/orchestrator.go` | execFn 透传 spawn 结果 |
| `internal/kernel/budget/budget.go` | 移除死代码引用（ErrTokenBudgetExhausted） |
| `docs/validation/e2e-combat-001.md` | 战斗报告更新 |
| `reports/session/2026-05-16-e2e-rework-and-budget-improvements.md` | 本报告 |

---

## 剩余工作

1. **`internal/kernel/budget/` 目录**：不再被任何包 import，是死代码，待手动删除
2. **E2E 第二枪**（vigil-c742e4）：用改进后的系统重跑 bug fix 类 E2E
3. **DeepSeek V4 Flash API key**：配置后可获得更快的模型推理速度
4. **Agent 实际使用 spawn**：需验证 agent 在收到多文件任务时是否会主动调用 spawn 分解
