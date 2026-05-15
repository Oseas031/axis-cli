---
type: spec
status: Draft
created: 2026-05-15
last_verified: 2026-05-15
related:
  - requirements.md
  - design.md
---

# Swarm Topology Tasks

**Status**: Draft
**Vigil**: vigil-ce7043

---

## T1: SwarmConfig + ParseFromMetadata

**File**: `internal/kernel/swarm/config.go`

- `SwarmConfig` struct（Pattern/MinSize/MaxSize/Diversity/Order）
- `ParseFromMetadata(map[string]string) *SwarmConfig` — 无 swarm.* key 返回 nil
- `Validate() error` — Pattern 必须是 "parallel_vote"，MinSize ≤ MaxSize，MinSize ≥ 2

**Done when**: `go test -race ./internal/kernel/swarm/...` passes; invalid config 返回 error; 无 swarm key 返回 nil。

## T2: Agent Selection + Diversity Validation

**File**: `internal/kernel/swarm/dispatch.go`

- `selectAgents(available []string, cfg *SwarmConfig) ([]AgentSlot, error)`
- v1: available = 已配置的 provider profile 名称列表
- diversity=heterogeneous 时要求 ≥2 distinct providers
- 可用 agent 不足时返回结构化错误

**Done when**: heterogeneous 拒绝单 provider; 不足 MinSize 返回 error。

## T3: Parallel Dispatch + Collect

**File**: `internal/kernel/swarm/dispatch.go`

- `Dispatch(ctx, task, cfg, dispatchFn) (*SwarmResult, error)`
- `dispatchFn` 是回调：`func(ctx, task, providerName) (map[string]any, error)` — 由 Dispatcher 注入
- 并行 goroutine 执行，WaitGroup 收集，context cancel 取消全部
- Order=shuffled 时 crypto/rand shuffle agent 列表

**Done when**: 3 agent 并行执行; 1 个失败不阻塞其他; ctx cancel 停止全部; shuffled 产生不同顺序。

## T4: Majority Vote Aggregation

**File**: `internal/kernel/swarm/aggregate.go`

- `Aggregate(results []SingleResult) *SwarmResult`
- Hash output → 分组 → 最大组胜出
- Confidence = largest_group / total; Unanimous = all same

**Done when**: 3 中 2 agree → confidence 0.67; 3 全 agree → unanimous; 全不同 → error。

## T5: Dispatcher Integration

**File**: `internal/kernel/dispatcher/dispatcher.go`（修改）

- `Dispatch()` 开头加 `swarm.ParseFromMetadata` 检查
- 非 nil 时调用 `swarm.Dispatch()`，传入 dispatchFn 包装现有单 Agent 路径
- dispatchFn 内部：临时切换 provider → 调用现有 `executeAgentTask`

**Done when**: 带 swarm metadata 的 task 走 swarm 路径; 不带的行为不变（现有测试全过）。

## T6: Event Recording

**File**: `internal/kernel/swarm/dispatch.go`（内嵌）

- Dispatch 完成后 append `swarm.executed` 事件到 event log
- 包含 agents/confidence/unanimous/pattern

**Done when**: 事件出现在 `.axis/events/tasks.jsonl`; JSON 格式正确。

---

## Definition of Done (全部)

- [ ] `go test -race ./internal/kernel/swarm/...` passes
- [ ] `go test -race ./internal/kernel/dispatcher/...` passes（回归）
- [ ] 无新外部依赖
- [ ] `go list -deps ./internal/kernel/swarm/ | grep -v "github.com/axis-cli/axis" | grep -v "^[a-z]"` 输出为空

## Verification Criteria

```bash
go test -race ./internal/kernel/swarm/...
go test -race ./internal/kernel/dispatcher/...
grep -rn "model/provider" internal/kernel/swarm/ --include="*.go" | grep -v "_test.go"
# Expected: 0 lines
```
