---
type: spec
status: draft
created: 2026-08-26
last_verified: 2026-08-26
related:
  - "requirements.md"
  - "design.md"
  - "../actor-comm/tasks.md"
---

# Human-Agent Scheduling Tasks

**Status**: In Progress（2026-08-26: T1/T2/T4 完成，T3 内核+控制面+启动接线完成，T5 scrubber+GUI 完成，T6 完成；T7 文档同步完成中）
**Implements**: requirements.md + design.md（v1 范围 = US1/US2/US5；US3/US4 后续阶段另行立项）
**前置依赖**: actor-comm P0 的 comm 包（已落地）
**反调修订**: 2026-08-26 依 adversarial review 12 条发现修订

每个任务独立可编译、bisect-safe；破坏性用例覆盖 bad JSON/超时/空 body/超限 payload（§10）。

## T1: 类型与合约扩展 ✅

- `internal/types`：HumanCallRequest **纯扩展**——保留 Input/Metadata，新增 Kind/Prompt/Options/Context/Deadline（design.md D1 定稿 schema）；新增 TaskStatusWaitingHuman / CallStatusDegraded / axis.executor.preference 常量。
- 回归：旧 JSON 反序列化不丢字段（human_request_test.go）。
- 验收：`go test ./internal/types/...` ✅

## T2: per-actor 收件箱存储 + 读加固 ✅

- `internal/comm/inbox.go`（新增）：per-actor `.axis/comm/<actor-id>.jsonl`；Peek 非破坏 / MutatePending CAS / MutatePendingAllowPending（nudge）；旧单文件 mailbox.jsonl 一次性迁移并改名 .imported。
- 读加固：bufio.Reader 手工行读取（无 Scanner 64KB 上限）+ Deliver 大小上限（DefaultMaxRecordBytes=256KB 可配），超限拒绝。
- 边界用例：损坏行跳过并计数、进程内并发写锁、重启 pending 保持、超限消息被拒且既有消息可读（inbox_test.go 7 用例）。
- 验收：`go test -race ./internal/comm/...` ✅

## T3: dispatcher 异步接线 + 状态机扩展 + 遗留迁移 ✅

- `internal/human/gate`（新增）：Request park 语义 / RespondCall / DegradeExpired 启动扫描 / nudge 审计 / fail-closed 降级链（gate.go+sla.go+redact.go，8 测试）。
- `internal/kernel/dispatcher`：HumanGate 接口 + executeViaHumanInbox（替代 busy-poll，Output 回传修复）+ PermissionAsk 接线（RequiresConfirmation 收窄激活范围保兼容）+ Dispatch 超时在 park 期间让位（human_gate.go）。
- `internal/kernel/orchestrator`：SetHumanGate 转发 + SetStatusObserver→scheduler.UpdateTaskStatus（waiting_human 可查询）。
- 单实例互斥：runtime.json 健康探测拒绝第二实例；启动过期降级扫描落地于 runLocalRuntime。
- 事件族：human.request.created/resolved/nudged/degraded → event log。
- 并发仲裁：Inbox.MutatePending CAS，resolve wins over degrade。
- 验收：dispatcher/gate -race 全绿 ✅；e2e 三用例（waiting_human 端到端/启动降级扫描/双实例拒绝）✅

## T4: CLI inbox 命令 ✅

- `cmd/axis/inbox_cmd.go`：list/show/respond；运行时存活走控制面代理，离线直读本地文件；--json snake_case；输出含 Next 提示。
- 用例 6 个：空箱/pending 展示/JSON 形状/单人自动匹配/重复 respond 拒绝/歧义要求 --call-id/show 渲染。
- 验收：`go test -race ./cmd/axis/...` ✅

## T5: 最小 scrubber + GUI Inbox 视图 ✅

- scrubber：gate/redact.go denylist key 过滤，Context/Input/Metadata 序列化前生效；泄漏负向断言 ✅。
- axis-gui：/api/inbox 只读 + /api/inbox/{id}/respond 写代理（不直写文件）；前端 Inbox 视图（decision 点选/info 文本回复）+ 路由导航；dist 已重建。

## T6: 对称 SLA（nudge + 降级链，仅 decision/info）✅

- Config{WaitTimeout,NudgeInterval,MaxNudges,MaxWaiting} 默认值常量化；Clock 抽象接口。
- 断言全绿：deadline 到期必降级；max_nudges 提前降级；respond/degrade 竞争 resolve wins；无"超时通过"路径（fail-closed 设计性排除）。

## T7: 端到端场景 + 文档同步 🔄

- 示例脚本：scripts/human-inbox-demo.ps1（US1 决策闭环）✅
- README CLI 表补 inbox 行 ✅；CHANGELOG/current-progress 同步中。
- 验收：demo 待人工跑通确认；docs lint error=0 ✅
## Definition of Done

- requirements.md 验收标准 1–5 全部满足
- 新增代码无新外部依赖；gofmt/vet/staticcheck 干净
- event log 中可回放一次完整双向闭环
- 反调发现的 3 个 blocker（恢复语义/redaction 幻影/verify fail-open）已在设计与任务中闭合
