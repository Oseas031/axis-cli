# Current Progress

**Updated**: 2026-11-15
**Milestones**: M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Coding Agent P0 ✅ | Hardening ✅ | Real-World Validation ✅

## Vigil Status

Total: 114 | Completed: 89 | Pending: 25 (stale entries pending W0 audit reconciliation)

## This Week (2026-11-15)

### Workflow Alignment & Cleanup (2026-11-15)
- 修正开发流为“人守 Spec 与概念完整性 + 冻结新子系统扩张 + 确定性 DAG 门控与 `human/gate` 终审”，已写回 `CLAUDE.md`
- 确立 3 步收敛落地路线：① 跑通本地实测闭环与任务基准数据校准 → ② 顺滑 `axis inbox` 并发与人机协同体验 → ③（可选）纳管外部 CLI Agent 作为 `AgentExecutor`
- 修复 `internal/human/gate` `degrade()` 吞错导致的 CAS 竞态（`TestGate_ResolveWinsOverDegrade` 压测通过）
- 清理 `cmd/axis/main.go` SystemPrompt 重复段落与仓库追踪残留 `nu_`，对齐 `README.md` Go 1.26+ 与状态描述

### Positioning Pivot (2026-08-26)
- 核心命题锁定：人机互调度新范式——同合约双执行者，双向任务流
- New spec: docs/specs/human-agent-scheduling/（反调 12 条发现已闭合）
- **v1 实现落地**：per-actor 持久收件箱（CAS/大小上限/迁移）、human gate（park/nudge/fail-closed 降级/scrubber）、dispatcher 异步接线（busy-poll 移除、PermissionAsk 启用、park 让位超时）、waiting_human 状态可查询、控制面 /v1/inbox 端点、CLI inbox 命令族、GUI Inbox 视图、单实例互斥、启动过期扫描、human.request.* 事件族
- 测试：gate 8 + comm 7 + dispatcher 新增 7 + CLI 6 + e2e 3，全仓 `go test -race -short` 绿
- Docs lint tooling: --check 范围过滤、目录链接 orphan 识别、占位符跳过；存量债清偿 153→0

### Streamlining (2026-08-25, positioning locked: A+B)
- Removed dead subsystems: kernel/audit, safego, contextpack shims (-1491 LOC)
- Removed docs/zh mirror (7k lines); 9 stale docs marked historical
- Constitution v1.3 (407L) → v2.0 slim charter; README honest positioning
- Windows half-open conn hardening shipped; CI green (staticcheck now local-gated)

### June Wave Landed (was uncommitted since 2026-06-15)
- contextpack rework: chunker/ranker/rule-engine/assemble/consumer/index/model; packet.go removed
- agent executor interfaces, judgement isolation, tool-trace updates
- working-memory engine, immediate adapter, kv interface extraction
- compactor semantic recovery; kernel audit package; swarm boundary/fuzz coverage
- storeutil shared JSONL/atomic-write utilities; token estimator
- research: Context Recommendation System survey + spec design.md (req/tasks pending)

### Windows Half-Open Connection Hardening (this session)
Root cause: pooled loopback keep-alive conns silently swallowed writes against
the local runtime / httptest servers → multi-minute indefinite hangs.
- `control.LocalHTTPClient()` (keep-alives off) + 30s fallback deadline in Client.do
- provider Execute now honors configured timeout via ctx deadline (was dead config)
- HTTPClientTool de-keep-alived; provider/tool tests use non-pooled clients
- Regression tests: TestClientDoFallbackDeadline, TestLocalHTTPClientContract

### Hygiene
- internal/tmp scratch package removed (30 files)
- personal/interview files moved out of repo; `.amp/` ignored; `nul` deleted
- 9 bisect-safe commits; gofmt normalized; race clean on kernel/control/memory/multiturn/cmd

## Capability Summary

| Layer | Key Capabilities |
|-------|-----------------|
| **Scheduling** | FIFO + DAG parallel, 5-worker orchestrator, SLA timeout/retry/failure_class |
| **LLM** | Anthropic/OpenAI/DeepSeek/MiniMax, token accounting, circuit breaker, escalation, semantic layering |
| **Tools** | BashTool (L0/L1/Unrestricted), SandboxedBashTool (Docker), FileRead/Write, HTTP, permission scopes |
| **Agent** | LLMAgentExecutor, multi-turn loop, circuit breaker, HistoryCompactor, EventEmitter |
| **Context** | ContextBundle, ReadinessRegistry, preflight, budget trimming, relevance scoring |
| **Evolution** | Isolated workspace, atomic steps, trace ledger, verification, promote/discard |
| **Judge** | 5 strategies, context isolation, two-pass escalation, generalization scoring |
| **Memory** | Horizon/Immediate/Immunity/KV/Longterm/Working layers |
| **Multi-Agent** | Subagent isolation, JSONL mailbox, multi-candidate differential testing |
| **Autonomy** | Feature gate, dispatcher resolver, capability registry, transition rules |
| **Control** | Local HTTP server, cross-process submit/query, event log, orphan recovery |

## Current Limitations

- Single-machine only (no distributed scheduling)
- No multi-tenant support
- Docker sandbox requires Docker
- Self-Judgement is advisory (LLM judging LLM has same-source bias)
- No production validation at scale

## Architecture Gaps (Top Priority)

| # | Gap | Status |
|---|-----|--------|
| A | Cross-process context fracture | Partial (Local Control Plane done) |
| B | No inter-Agent collaboration primitives | Open |
| E | Tool boundaries static, not dynamic ladders | Open |
| F | No dynamic model routing | Partial (cost tracking done) |
| G | No Agent identity/capability profile | Open |
| H | Execution feedback loop broken | Open |

## History

- [2026-05-15 Week](history/week-2026-05-15.md) — GitHub governance, swarm T1-T6, agent infra P1s, cost budget, CI quality
- [Milestone Completions (M1-M6)](history/milestone-completions.md) — all completed task lists
- [Architecture Diagnosis](history/architecture-diagnosis.md) — 2026-05-11 strategic analysis
