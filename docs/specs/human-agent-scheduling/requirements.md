---
type: spec
status: draft
created: 2026-08-26
last_verified: 2026-08-26
related:
  - "../actor-comm/requirements.md"
  - "../../CLAUDE.md"
---

# Human-Agent Scheduling Requirements

> 范式命题：同一份任务合约下，人与 Agent 互为可调度的执行者。人能让 Agent 做什么，Agent 也能让人做什么——接口对称，能力各异。

## Problem

当前 Axis 只有人→Agent 的单向调度（`ask/run`）。Agent 遇到只有人能处理的事务时，只能阻塞等待（PermissionAsk 未接线）或失败。同时，人类侧的工作追踪（vigil）与 Agent 任务队列是两个割裂的世界。

本 Spec 把"人"升级为调度系统中的一等执行者：**双向任务流、同一合约语言、同一审计账本**。

## 与 actor-comm 的关系

[actor-comm](../actor-comm/requirements.md) 定义了传输层（Actor 接口、Message、Mailbox、Router），其 P0 的 comm 包已落地。本 Spec **不重复传输层**，定义其之上的**调度语义与人类侧闭环**：

- actor-comm 回答"消息怎么送"
- 本 Spec 回答"任务怎么派给人、人怎么回、等不到人怎么办"

## User Stories

### v1 范围

#### US1: 决策请求（decision）
> 开发者让 Axis 重构一个模块。Agent 执行中发现两个可行方案，向人的收件箱发一条带上下文的决策请求并暂停等待。人在 GUI 收件箱点选方案 B，任务恢复执行。

#### US2: 信息请求（info）
> 用户让 Axis 准备一份报告。Agent 发现缺少"目标读者是谁"这一只有人知道的信息，发出信息请求并暂停。人回复后 Agent 继续。

#### US5: 对称 SLA（审计语义）
> 人长期未响应时，系统按 nudge 策略在事件账本记录提醒并使收件箱条目老化置顶；超过硬期限走降级链，全程可审计。注意：v1 无推送通道，nudge 是审计记录与排序信号，不是主动通知。

### 后续阶段（本 Spec 不实现，仅占位）

- **US3 委派任务（delegate）**: vigil item 联动、Agent 起草内容回填——依赖 vigil 语义扩展，另行立项。
- **US4 验收请求（verify）**: 由 Self-Judgement 引擎主动发出——依赖判定引擎改造，另行立项。

## Functional Requirements

- **FR1 收件箱持久化**: 人类收件箱基于 per-actor JSONL mailbox（`.axis/comm/<actor-id>.jsonl`），进程重启后 pending 请求不丢失；单条消息有大小上限，超限拒绝而非写入（防读端 brick）。
- **FR2 请求类型**: v1 实现 decision / info 两类，以 `HumanCallRequest.Kind` 字段为准（`axis.request.kind` metadata 仅作日志过滤镜像）；payload 含完整决策上下文。
- **FR3 等待语义**: Agent 发出请求后任务置 `waiting_human` 子状态，执行 goroutine 挂起于调用点、**继续占用 worker 槽**（诚实成本），并发等待数受 `max_waiting` 上限保护。**跨重启不恢复**：runtime 重启时 waiting_human 任务标记 abandoned 并留事件，收件箱请求保留但不再触发恢复——真正的执行态序列化/checkpoint 是未来独立 Spec。
- **FR4 CLI 收件箱**: `axis inbox list/show/respond/done`，输出符合宪法 §8（行导向、--json snake_case、成功含主 ID 与下一步建议）。
- **FR5 GUI 收件箱视图**: axis-gui 新增 Inbox 视图，读 `.axis/comm/` 展示 pending 项，写操作经控制面代理。此为对 §14 Observatory 只读定位的显式修订（过渡条款：失效条件=Inbox 被证实滥用写路径）。
- **FR6 对称 SLA**: 人型等待适用 nudge 策略（可配置间隔与上限），不吃 retry；nudge 为 event log 审计记录 + 收件箱老化排序。超硬期限走降级链：decision→保守默认项+标记假设并挂起待确认；info→挂起上报。verify 类降级为 fail-closed（见 US4 占位）。参数入 config。
- **FR7 路由偏好**: 任务 metadata `axis.executor.preference ∈ {auto, require-human, require-agent}`；v1 仅此三元规则（auto/require-agent→agent 执行器，require-human→inbox）。capability registry 匹配是未来 Spec，不在本期。
- **FR8 全程留痕**: 请求创建/响应/nudge/降级均为 event log 事件（`human.request.*`），可审计。

## Non-Goals

- 不做手机推送（通知通道只留抽象接口）
- 不做分布式/多租户；同一项目根同时只允许一个 runtime（启动互斥）
- 不做自然语言回复理解（respond 的结构化字段优先，自由文本原样透传给任务上下文）
- 不做执行态序列化/checkpoint（跨重启恢复依赖它，明确推迟）
- 不改 vigil 自身语义；不动 Self-Judgement 引擎（delegate/verify 后续阶段另行立项）

## Acceptance Criteria

1. Agent 提交的 decision 请求在 `axis inbox list` 可见，`respond` 后任务从 `waiting_human` 恢复至终态，event log 有 created+resolved 两事件。
2. 杀掉 runtime 后重启：pending 请求仍在收件箱中；对应任务被标记 abandoned 且留痕（P0 不做跨重启恢复）。
3. nudge 次数达到上限后任务进入降级路径而非无限等待；并发 respond 与降级竞争时 resolve wins（fake clock + CAS 断言）。
4. 超限大小的 respond 被拒绝且不影响既有消息可读性。
5. `go test -race -short ./internal/comm/... ./internal/kernel/... ./cmd/axis/...` 全绿。
