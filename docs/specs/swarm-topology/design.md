---
type: spec
status: Draft
created: 2026-05-15
last_verified: 2026-05-15
related:
  - requirements.md
  - tasks.md
  - ../../research/bystander-effect-multi-agent-reasoning-2026-05-15.md
---

# Swarm Topology Design

> 实现 kernel-abstraction-model.md — Topology 作为 Dispatcher 的可选执行路径

**Status**: Draft

## Core Decision

Swarm Topology 不是独立的 Kernel 抽象层。它是 Dispatcher 的一个**可选执行路径**：当 task metadata 包含 `swarm.*` 字段时，Dispatcher 走多 Agent 并行路径而非单 Agent 路径。

```
Dispatcher.Dispatch(task)
    │
    ├─ task.Metadata["swarm.pattern"] 存在？
    │   ├─ Yes → swarmDispatch(task)  ← NEW
    │   └─ No  → existing single-agent path (unchanged)
```

**为什么不是独立 Kernel 层**：Axis 从未执行过真实多 Agent 任务。在验证价值前，不值得引入新的 Kernel 抽象。Dispatcher 内部的可选路径足够，且可随时提升为独立模块。

## Data Model

```go
package swarm

// SwarmConfig is parsed from task metadata. Minimal, no file-system backing.
type SwarmConfig struct {
    Pattern   string // "parallel_vote" (only supported value in v1)
    MinSize   int    // default 2
    MaxSize   int    // default 3
    Diversity string // "none" | "heterogeneous"
    Order     string // "fixed" | "shuffled"
}

// ParseFromMetadata extracts SwarmConfig from task metadata.
// Returns nil if no swarm.* keys present.
func ParseFromMetadata(meta map[string]string) *SwarmConfig

// AgentSlot represents one participant in the swarm execution.
type AgentSlot struct {
    AgentID  string
    Provider string // provider profile name
}

// SwarmResult is the output of a swarm execution.
type SwarmResult struct {
    Agents      []AgentSlot
    Results     []SingleResult // one per agent
    Winner      *SingleResult  // majority vote winner
    Confidence  float64        // agreement_count / total
    Unanimous   bool
}

// SingleResult is one agent's output.
type SingleResult struct {
    AgentID string
    Output  map[string]any
    Error   string
}
```

## Module Placement

```
internal/kernel/swarm/
├── config.go        // SwarmConfig + ParseFromMetadata
├── config_test.go
├── dispatch.go      // SwarmDispatch function
├── dispatch_test.go
├── aggregate.go     // majority vote aggregation
└── aggregate_test.go
```

6 个文件。不设 BOUNDARY.md（规模不够）。

## Execution Flow (parallel_vote)

```
1. ParseFromMetadata(task.Metadata) → SwarmConfig
2. Validate: MinSize ≤ MaxSize, Pattern == "parallel_vote"
3. SelectAgents: 从可用 provider profiles 中选 N 个满足 diversity 约束的
4. Shuffle: 如果 Order == "shuffled"，随机化 agent 列表（crypto/rand）
5. Parallel Execute: 对每个 AgentSlot，clone task + 设置 provider → 调用单 Agent dispatch
6. Collect: 等待所有完成（context cancellation 取消全部）
7. Aggregate: hash outputs → majority vote → SwarmResult
8. Record: append swarm.executed event
```

## Diversity Enforcement

```go
func validateDiversity(slots []AgentSlot, diversity string) error {
    if diversity != "heterogeneous" {
        return nil
    }
    seen := map[string]bool{}
    for _, s := range slots {
        seen[s.Provider] = true
    }
    if len(seen) < 2 {
        return errors.New("heterogeneous diversity requires >=2 distinct providers")
    }
    return nil
}
```

## Agent Selection

v1 简化：从 `provider list` 中取所有已配置的 provider profile 作为可用 Agent 池。每个 provider profile = 一个 AgentSlot。如果可用 provider 数量 < MinSize 且 diversity=heterogeneous，返回错误。

## Aggregation (majority_vote)

复用 CandidatePool 的语义：hash 每个 agent 的 output，按 hash 分组，最大组胜出。

## Dispatcher Integration

```go
// In dispatcher.go
func (d *DispatcherImpl) Dispatch(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
    cfg := swarm.ParseFromMetadata(task.Metadata)
    if cfg != nil {
        return d.dispatchSwarm(ctx, task, cfg)
    }
    // existing path unchanged
    ...
}
```

## Event Schema

```json
{
  "type": "swarm.executed",
  "task_id": "task-123",
  "topology": {"pattern": "parallel_vote", "size": 3, "diversity": "heterogeneous"},
  "agents": [
    {"agent_id": "agent-1", "provider": "claude"},
    {"agent_id": "agent-2", "provider": "gpt"},
    {"agent_id": "agent-3", "provider": "deepseek"}
  ],
  "confidence": 0.67,
  "unanimous": false
}
```

## What This Design Explicitly Defers

| 概念 | 状态 | 激活条件 |
|------|------|----------|
| sequential_review pattern | 不实现 | 有真实场景证明"审查链"比"并行投票"更好 |
| hierarchical_delegate | 不实现 | Actor 接口实现 + Lead Agent 能力存在 |
| credibility scoring | 不实现 | 有 ≥50 次 swarm 执行历史数据 |
| sovereignty_budget | 不实现 | 有 D_L 校准数据 |
| topology 文件系统 | 不实现 | 有 ≥3 个不同 topology 配置被反复使用 |
| 独立 Kernel 抽象层 | 不提升 | swarm dispatch 路径被 ≥3 个模块依赖 |

## Semantic Boundary

**swarm 包 Must Not**:
- 直接 import `model/provider`（通过 dispatcher 路径）
- 修改 task 内容（只读 metadata）
- 存储状态（无状态，每次从 metadata 解析）
- 影响非 swarm 任务的 dispatch 路径
