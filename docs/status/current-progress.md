# Current Progress

**Updated**: 2026-05-15
**Milestones**: M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Coding Agent P0 ✅ | Hardening ✅ | Real-World Validation ✅

## Vigil Status

Total: 53 | Completed: 49 | Pending: 25 (3 P1, 17 P2, 4 P3)

## This Week (2026-05-15)

### GitHub Governance Overhaul
- Milestone tags m1-m6 + Release v0.1.0
- Branch protection (require CI, block force push)
- Issue templates (bug/feature/task)
- 250 private files removed from tracking (.claude/.devin/.swarm/)
- git-conventions.md: "Never Commit" section

### Swarm Topology (T1-T6)
- `internal/kernel/swarm/` — config, dispatch, aggregate (6 files)
- Parallel multi-agent execution with majority vote
- Dispatcher integration: swarm.* metadata → multi-agent path
- SwarmEvent emission for observability

### Agent Infrastructure (5 P1 tasks)
- FollowUpTask population (parse _next_steps from output)
- Interrupt ledger closure (synthetic tool_result on abort)
- Compact semantic recovery (RecoveryContext)
- Prompt layering (PromptAssembler, priority-based chain)
- Permission tri-state (ask/allow/deny + AutonomyLevel mapping)

### Cost Budget
- `AgentTask.CostBudget` field (float64, USD, 0=unlimited)
- `CostTracker` — per-task accumulation, 80% downgrade threshold
- Token usage callback wired: multiturn → executor → tracker
- Dispatcher enforcement: pre-execution budget check

### CI & Quality
- staticcheck 7 errors fixed
- Sandbox tests: testing.Short() skip for CI
- Vigil lock: Linux process detection fix (syscall.Kill)
- Provider type validation + task ID collision fix
- Blind testing: 11M fuzz executions, 0 crashes; race detector clean

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

- [Milestone Completions (M1-M6)](history/milestone-completions.md) — all completed task lists
- [Architecture Diagnosis](history/architecture-diagnosis.md) — 2026-05-11 strategic analysis
