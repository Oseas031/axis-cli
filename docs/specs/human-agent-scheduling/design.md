---
type: spec
status: draft
created: 2026-08-26
last_verified: 2026-08-26
related:
  - "requirements.md"
  - "../actor-comm/requirements.md"
---

# Human-Agent Scheduling Design

> 原则：人机对称靠合约不靠特判；一切状态变更留痕；安全默认（preview/explicit/redaction）；诚实标注 v1 语义边界（park 不序列化、nudge 即审计）。

## 复用映射（不新建平行体系）

| 现有资产 | 本 Spec 中的角色 | 反调修正后的事实核查 |
|---|---|---|
| `comm.Message`（ID/From/To/Type/Payload/ReplyTo） | 收件箱消息单元，直接复用 | ✓ 已实现 |
| `comm.Mailbox`（JSONL） | 持久化底座；完成其文件头 transitional 注释的 per-actor 化 | ✓ 已实现，但读端有 64KB Scanner 上限缺陷，T2 加固 |
| `types.HumanCallRequest` | **扩展**而非替换：保留 Input/Metadata，新增字段 | dispatcher.go:668 现有调用方不破坏 |
| dispatcher `PermissionAsk` | 异步请求语义来源；接线即本 Spec T3 | permission.go:49 TODO |
| event log | 审计账本，新增 `human.request.*` 事件族 | ✓ |
| vigil / capability registry / Self-Judgement | **不动**（delegate/verify 后续阶段另行立项） | registry 实为存在性检查表，不做路由匹配 |

## D1: 数据合约

- per-actor 存储：`.axis/comm/<actor-id>.jsonl`，人类默认 actor-id 为 `human`。
- HumanCallRequest **纯扩展**（旧 JSON 可解析、旧调用方编译通过）：

```go
type HumanCallRequest struct {
    CallID   string            // 全局唯一请求 ID（≠TaskID，废除旧 CallID==TaskID 约定）
    TaskID   string
    Kind     string            // 决策性字段：decision | info（v1）；axis.request.kind metadata 仅作日志镜像
    Prompt   string
    Options  []string          // decision 类候选
    Context  map[string]any
    Input    map[string]any    // 保留（兼容）
    Metadata map[string]string // 保留（兼容）
    Deadline time.Time
}
```

- Message 编码规则（定死防分叉）：请求统一 `Type=query`；`ReplyTo` 引用**被响应消息的 Message.ID**；`inbox respond <call-id>` 先按 Payload 内 CallID 索引定位消息再取其 ID。

## D2: 等待与恢复语义（v1 = park，不序列化）

- dispatcher 命中 PermissionAsk → 产出 human request → 任务置 `waiting_human` → 执行 goroutine 挂起于调用点 channel。**worker 槽被占用是诚实的 v1 成本**，`max_waiting`（config，默认 2）防止池耗尽。
- 恢复：respond → 控制面 → dispatcher 按 CallID 唤醒对应 channel → payload 注入执行点。
- 跨重启：启动扫描将 `waiting_human` 一律标记 abandoned（原因 human_wait_interrupted）并留事件；收件箱请求保留、可查看、不再触发恢复。执行态 checkpoint 是未来独立 Spec 的前置。
- 单实例互斥：`axis start` 经 runtime.json PID 存活检查拒绝同项目根第二实例（防双写 mailbox/双恢复/事件重放）。
- 遗留路径迁移归 T3 所有：移除 `executeHumanTask` busy-poll（顺带消除 §2 违规），新路径一律走 inbox。

## D3: 对称 SLA（v1 只含 decision/info）

- config 新增 `human_sla{nudge_interval, max_nudges, degrade_policy}`，clock 抽象 + fake clock 测试。
- nudge = event log 审计记录 + 收件箱老化排序信号。**不是主动通知**（无守护进程、无推送通道，宪法 §1.2）——文档与 CLI 输出均如实表述。
- 降级链（fail-closed 优先）：decision→保守默认项+标记假设+任务挂起待确认；info→挂起上报。**不存在任何"超时视为通过"路径**——Ask 永不时钟自动升级 Allow（permission.go 不变量、宪法 §11）。
- 并发仲裁：调用生命周期单写者；respond 与 degrade 竞争时对 call status 做 CAS，**resolve wins over degrade**；败者幂等退出。
- 启动扫描补"已过期未答"分支：Deadline 已过的 pending 请求在启动时直接执行降级链。

## D4: 路由（v1 三元规则，无 registry 匹配）

- 仅 `axis.executor.preference ∈ {auto, require-human, require-agent}`：auto/require-agent→agent 执行器；require-human→inbox。auto 默认且不隐式扩权（§11）。capability 匹配推迟至独立 Spec（现有 registry 是存在性检查表，非路由器——反调发现 #10）。

## D5: CLI 面（对齐 actor-comm FR5）

```
axis inbox list [--json]           # pending 列表（老化排序：超期在前）
axis inbox show <call-id>          # 详情
axis inbox respond <call-id> --choice B | --text "..."   # decision 点选 / info 回复
```

输出契约按宪法 §8。done 子命令随 delegate 类后续阶段一并引入。

## D6: GUI Inbox 视图

- axis-gui 新增 `/inbox`：读 `.axis/comm/human.jsonl` 渲染；respond 经控制面代理 POST。此为对宪法 §14 Observatory 只读定位的显式修订，过渡条款：失效条件=Inbox 写代理被证实滥用。

## D7: 控制面端点

- `GET /v1/inbox`、`POST /v1/inbox/{call-id}/respond`
- snake_case 响应体；事件先落账本再返回主 ID。

## 语义边界自查（§6）

AgentTask 无执行逻辑 ✓；Scheduler 无 model 调用 ✓；PermissionAsk 语义不变仅载体变异步 ✓；ContextBundle 不升级权限 ✓；收件箱为 pull 模型非 push context ✓；无隐式自主权扩张（human 可达性由人自己配置）✓。

## 安全

- ~~沿用 provider 层脱敏~~（反调发现 #2：该层不存在）。v1 具体方案：新增最小 scrubber（denylist key 匹配，作用于 Context/Metadata 序列化前），respond 自由文本按不透明数据处理、长度受限；secret 防护的系统性方案仍以 memory/immunity 注释指出的缺口为准，另行立项。
