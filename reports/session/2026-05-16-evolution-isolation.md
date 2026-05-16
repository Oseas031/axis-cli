# Session Report: Evolution Isolation & Agent Observability

**Date**: 2026-05-16
**Duration**: ~1.5 hours
**Theme**: 从"Agent 弄乱代码"到"物理隔离 + 可审计 + 可观察"的完整闭环

---

## 战略成果

本次会话解决了 Axis 自主执行中的**核心安全缺陷**：Agent 通过 `axis ask --submit` 执行任务时直接修改项目主树，无法回滚，每次测试后需手动恢复。

### 解决路径

```
问题发现 → 方案选择 → 实现 → E2E 验证
   ↓            ↓          ↓         ↓
Agent乱改   方案C(Evolution)  3步并行   file_write被拒
代码无法恢复  默认隔离+审计   scoped工具  bash cwd隔离
```

### 关键决策

| 决策 | 选择 | 原因 |
|------|------|------|
| 隔离方案 | Evolution Protocol（而非 git stash/tool.allowed_paths） | 符合 Axis 哲学（可控演化、审计而非审批）；基础设施已存在 |
| 隔离粒度 | bash cwd + file_write allowedDirs（不 copy 项目） | 最小实现成本；Agent 可读全项目但只写 workspace |
| 默认行为 | 隔离（需 `--direct` 跳过） | 安全默认原则；自迭代是例外而非常态 |
| Agent 自适应 | 不修改 Agent prompt | Agent 自行发现 file_write 被拒后改用 bash 写入 workspace |

---

## 技术变更清单

### 新文件
- `internal/model/tool/scoped.go` — ScopedBashTool + ScopedFileWriteTool + NewScopedRegistry

### 修改文件
| 文件 | 变更 |
|------|------|
| `internal/model/multiturn/loop.go` | +CostGuard, +OnTurnCompleted |
| `internal/agent/llm_executor.go` | +traceDir, +WithTraceDir, +Tools() getter, CostTracker wiring |
| `internal/kernel/dispatcher/dispatcher.go` | executeViaEvolution 创建 workspace + scope tools |
| `internal/kernel/orchestrator/orchestrator.go` | Unlock FeatureEvolution |
| `internal/model/tool/bash.go` | toWSLPath() 修复 Agent 路径混乱 |
| `cmd/axis/main.go` | WithTraceDir wiring, `axis status --trace` |
| `cmd/axis/ask_cmd.go` | `--direct` flag, 默认注入 evolution metadata |
| `cmd/axis/evolve_cmd.go` | promote 时 copy workspace → project root |
| `internal/types/costtracker.go` | 重建（被 Agent 删除后恢复） |
| `internal/types/types.go` | 恢复 CostBudget/SLAKeyCostBudget/ErrCostBudgetExceeded |

---

## E2E 验证结果

### Test 1: CostTracker E2E（DeepSeek）
- **任务**: "为 AgentTask 添加 cost_budget 字段..."
- **结果**: ✅ completed（50 轮内）
- **产出**: Agent 创建 costtracker.go（类型定义）
- **发现**: 旧二进制 20 轮限制 → 需重新编译

### Test 2: Reverse（移除 cost_budget）
- **任务**: "从 AgentTask 中移除 cost_budget..."
- **结果**: ❌ 50 轮用尽
- **发现**: 
  - Agent 路径混乱浪费 10+ 轮（已修复 toWSLPath）
  - Spawn 子 agent 也有轮次限制（1 个超时）
  - Agent 部分完成（删了 costtracker.go + CostGuard）但未完成全链路

### Test 3: Evolution 隔离验证
- **任务**: "创建 internal/types/hello.go..."
- **结果**: ✅ completed + 完全隔离
- **证据**:
  - `file_write` → "path is not in allowed directories" ← 拒绝
  - `bash cwd` → `.axis/evolution/run-.../workspace/` ← 隔离
  - Agent 自适应用 bash echo 写入 workspace ← 成功
  - `internal/types/hello.go` 不存在于项目主树 ← 确认

---

## 观察与教训

### Agent 行为模式
1. **路径混乱是首要效率杀手** — WSL bash 里 Windows 路径无意义，Agent 反复试错。toWSLPath 修复后应大幅改善。
2. **Agent 能自适应工具约束** — file_write 被拒后自动改用 bash 写文件。不需要改 prompt。
3. **Spawn 有效但有代价** — 每个 spawn 消耗独立轮次预算，超时时主 Agent 丢失子任务的进度。
4. **Memory 只在 judgement 失败时存 lesson** — budget exhaustion 不触发。

### 架构 Gap 更新
- **Gap E (工具边界)**: 部分解决 — Evolution 下动态 scope tools（ScopedRegistry）
- **新 Gap I (Budget exhaustion 无学习)**: budget 用尽时应存 lesson 记录"做到哪了"

---

## 下一步建议

1. **Budget exhaustion lesson** — 在 multiturn 结束时如果是 budget exhausted，存一条 lesson
2. **file_read 也 scope 到 workspace** — 当前 Agent 可以读全项目（设计意图），但如果要测试完全隔离的自迭代，需要选项
3. **DeepSeek 性价比验证** — 跑更多任务统计 token/成功率/轮次，与 Claude 对比
4. **Promote 流程增强** — promote 前显示 diff preview，类似 `git diff --stat`
