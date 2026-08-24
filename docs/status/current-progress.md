# Current Progress

**Updated**: 2026-08-24
**Milestones**: M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Coding Agent P0 ✅ | Hardening ✅ | Real-World Validation ✅

## Vigil Status

Total: 114 | Completed: 89 | Pending: 25 (stale entries pending W0 audit reconciliation)

## This Week (2026-08-24)

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
