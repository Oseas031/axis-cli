# Cost Budget: Requirements

> 实现 docs/architecture/agent-native-first-principles.md 原则 #3（可控演化）

## 目标

AgentTask 新增 cost_budget 约束，Dispatcher 在执行过程中检查累计 token 消耗，超出预算时自动降级到更便宜的模型。

## 需求

### R1: AgentTask.CostBudget 字段

- 在 `internal/types/types.go` 的 `AgentTask` struct 中新增 `CostBudget int` 字段（单位：token）
- JSON tag: `cost_budget`
- 零值表示无预算限制

### R2: Dispatcher 预算检查

- Dispatcher 在调用 AgentExecutor 前读取 `task.CostBudget`
- 执行过程中通过 provider 的 token accounting 累计消耗
- 当累计消耗 >= 80% budget 时，触发模型降级

### R3: 自动降级机制

- 降级策略：切换到 semantic layer = utility 的 provider（更便宜的模型）
- 降级后继续执行，不中断任务
- 降级事件写入 audit log：`cost.downgrade` event，包含 task_id、consumed、budget、new_model

### R4: 超出预算处理

- 当累计消耗 >= 100% budget 时，标记任务为 `failed`
- 失败原因：`cost_budget_exceeded`
- 不静默吞掉——必须在 TaskResult 中可观察

## 验收标准

1. `go test -race ./internal/types/ ./internal/kernel/dispatcher/` 通过
2. 新增测试：正常执行（budget 充足，不降级）
3. 新增测试：接近 budget（>=80%，触发降级，任务继续）
4. 新增测试：超出 budget（>=100%，任务失败）
5. 无外部依赖引入
6. Metadata key 使用 `cost.*` 命名空间

## 约束

- 不修改 Scheduler 语义
- 不修改 Provider 接口（通过现有 token accounting 读取消耗）
- 降级不改变 tool 权限（只改模型，不改能力）
