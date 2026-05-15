---
type: spec
status: Draft
created: 2026-05-15
last_verified: 2026-05-15
related:
  - ../actor-comm/requirements.md
  - ../../architecture/kernel-abstraction-model.md
  - ../../research/bystander-effect-multi-agent-reasoning-2026-05-15.md
---

# Swarm Topology Requirements

> 实现 agent-native-first-principles.md P4（Layered Isolation is Collaboration）

**Status**: Draft
**Vigil**: vigil-ce7043

## Problem Statement

Axis 的 CandidatePool 有 DiversityPolicy 和 output equivalence partitioning，但只在 Agent 内部使用，Dispatcher/Orchestrator 无法发起"让多个 Agent 并行执行同一任务并聚合结果"的操作。

论文 arXiv:2605.10698 证明：
- 异构 swarm 比同构 swarm 准确率高 23%
- 固定发言顺序导致 Lead Anchor Effect
- 规模超过阈值后 Agent 自主性指数衰减

当前缺失：Dispatcher 没有"多 Agent 并行执行 + 约束 + 聚合"的路径。

## Functional Requirements

### FR1: Topology Declaration

- Task metadata 可声明 `swarm.pattern`（当前仅支持 `parallel_vote`）
- 声明 `swarm.min_size` / `swarm.max_size`（默认 2/3）
- 声明 `swarm.diversity`：`none` | `heterogeneous`（默认 heterogeneous）
- 声明 `swarm.order`：`fixed` | `shuffled`（默认 shuffled）

### FR2: Topology-Aware Dispatch

- Dispatcher 检测 `swarm.*` metadata，存在时走多 Agent 路径
- 按 diversity 约束选择 Agent（不同 provider）
- 约束不可满足时返回结构化错误，不静默降级
- 无 `swarm.*` metadata 的任务行为完全不变

### FR3: Parallel Execution + Aggregation

- 并行启动 N 个 Agent 执行相同任务
- 所有完成后，按 majority vote 聚合（复用 CandidatePool.Partition 语义）
- 聚合结果包含 confidence（agreement_count / total）

### FR4: Order Shuffling

- `order: shuffled` 时，Agent 结果注入聚合的顺序随机化
- 目的：防止实现中任何隐式的位置偏好

### FR5: Observability

- 每次 swarm 执行产生 `swarm.executed` 事件
- 包含：参与 Agent 列表（ID + provider）、各自结果、聚合结果、confidence

## Non-Goals

- 不做 sequential_review（无真实场景验证需求）
- 不做 hierarchical_delegate（无 Lead Agent 基础设施）
- 不做 credibility scoring（无历史数据源）
- 不做 sovereignty_budget 自动计算（无 entropy 估算能力）
- 不做 `.axis/topologies/` 文件系统（inline metadata 足够）
- 不做 CLI 命令（`axis swarm *`）（无用户）
- 不做分布式 swarm

## Acceptance Criteria

1. Task 带 `swarm.pattern=parallel_vote` + `swarm.min_size=2` 时触发多 Agent 执行
2. `swarm.diversity=heterogeneous` 时拒绝全同 provider 的 Agent 集合
3. 无 `swarm.*` metadata 的 task dispatch 行为不变（回归）
4. `swarm.executed` 事件包含完整参与者和结果
5. `go test -race ./internal/kernel/swarm/...` 通过
6. 无新外部依赖

## Prerequisite Validation

> 本 spec 实现前必须先验证：CandidatePool 能在至少一个真实任务中产出有意义的差异结果。如果多候选本身无价值，Topology 层也无价值。验证方式：手动用 `axis run` 对同一任务跑 2 个不同 provider，观察输出差异。
