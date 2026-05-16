# Current Progress

**Updated**: 2026-05-16
**Milestones**: M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Coding Agent P0 ✅ | Hardening ✅ | Real-World Validation ✅ | Evolution Isolation ✅

## Vigil Status

Total: 53 | Completed: 49 | Pending: 25 (3 P1, 17 P2, 4 P3)

## This Week (2026-05-16)

### Evolution Filesystem Isolation (E2E validated)
- `axis ask --submit` now defaults to Evolution Protocol (isolated workspace)
- `--direct` flag to bypass isolation for self-iteration testing
- `ScopedBashTool` — forces cwd to `.axis/evolution/<run>/workspace/`
- `ScopedFileWriteTool` — restricts allowedDirs to workspace only
- `NewScopedRegistry` — runtime tool swap for evolution execution
- `axis evolve promote <run-id>` copies workspace → project root
- `FeatureEvolution` unlocked by default in orchestrator gate
- **E2E verified**: Agent file_write rejected to project root, bash cwd isolated, Agent self-adapted

### History Dump (Trace System)
- `OnTurnCompleted` hook in multiturn loop — emits each ModelMessage
- Per-task trace files: `.axis/traces/<task-id>.jsonl` (incremental append)
- `axis status <task-id> --trace` — formatted conversation history viewer
- Full agent reasoning chain now observable post-mortem

### CostTracker Integration
- `CostGuard` callback in multiturn LoopConfig — per-turn cost check
- LLMAgentExecutor wires `CostTracker` from task budget → CostGuard
- 80% threshold → `cost_degraded` event; 100% → abort with `COST_BUDGET_EXCEEDED`
- Provider `CostEstimateUSD` flows through to tracker

### BashTool Path Fix
- `toWSLPath()` converts Windows cwd to `/mnt/<drive>/...` in bash output
- Eliminates Agent path confusion (was wasting 10+ turns per task on path issues)

### Iteration Budget Fix
- Default 20→50 confirmed working (first E2E run failed due to stale binary)
- `axis.max_iterations` metadata override verified

### Dead Code Cleanup
- `internal/kernel/budget/` directory deleted (orphaned, zero imports)

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
