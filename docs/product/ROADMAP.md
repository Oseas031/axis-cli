# Project Evolution Roadmap

**[Chinese version / 中文版](../zh/product/ROADMAP.md)**

**Last Updated**: 2026-05-11

## Evolution Stages

```
M1 Basic Sched → M2 Parallel Sched → M3 Exec Ecosystem → M4 Real LLM → M5 Bootstrap Loop → M6 Self-Judgement
  Completed        Completed            Completed           Completed      Completed           Completed
```

## Milestone 1: Basic Scheduling ✅ Completed

**Core Capabilities**
- FIFO task scheduling
- Simple dependency management
- Input/output validation
- Basic state storage
- Basic CLI

**Verification Criteria**
- Task submit → schedule → execute → result return closed loop
- Dependent tasks execute only after dependencies complete
- Input/output schema validation passes

## Milestone 2: Parallel Scheduling ✅ Completed

**New Capabilities**
- DAG parallel scheduling (`GetReadyTasks` ready-set API)
- Contract admission rules (`internal/contract/admission`)
- SLA agreements (timeout, retry, `failure_class`, backoff strategies, priority ordering)
- Error code system (`AgentError` / `ErrorCode`)
- Parallel execution loop (worker pool + bounded goroutines)

**Performance Improvements**
- Independent tasks execute in parallel
- Total duration compressed from "sum of all stages" to "critical path duration"

## Milestone 3: Execution Ecosystem ✅ Completed

**Phase 1 — Execution Path**
- ModelProvider interface + Mock/Echo implementations
- Dispatcher → ContractExecutor → ModelProvider execution chain
- `ErrDependencyNotReady` + `sla.failure_class`

**Phase 2 — Ecosystem Maturity**
- Provider configurability (functional options)
- HumanExecutor routing
- DAG enhancements (`GetAllTasks` / `GetDependencyGraph`)

**Phase 3 — SLA Strategy Engine + Tool Layer**
- `failure_class` routing (retryable / fatal / degradable)
- Backoff strategies (fixed / linear / exponential)
- Priority ordering (`sla.priority` 0-255)
- Tool interface + ToolRegistry + BashTool
- ModelRequest/ModelResponse multi-turn extension
- ContractExecutor multi-turn tool-use loop (max 10 turns)

## Milestone 4: Real LLM Integration ✅ Completed

**New Capabilities**
- Anthropic Provider (Claude series, direct HTTP calls)
- OpenAI Provider (GPT series, compatible with DeepSeek / MiniMax)
- Provider configuration (functional options: `WithModel` / `WithAPIKey` / `WithBaseURL`)
- Token accounting (`InputTokens` / `OutputTokens`)
- Safe JSON serialization

**Extended Tools**
- `file_read` — file reading with path validation
- `file_write` — file writing with path validation
- `http_request` — HTTP client with host allowlist
- Tool permission scopes
- Circuit breaker (consecutive error threshold)

## Milestone 5: Bootstrap Loop ✅ Completed

**Core Components**
- AgentExecutor interface + MockAgentExecutor
- AgentRuntimeAdapter
- SelfContext data structure
- ContextBuilder / ContextCompressor
- Self-iteration contracts (analyze / implement / validate / update / review / spawn)
- BootstrapOrchestrator (self-loop task scheduling)
- FollowUpTaskGenerator
- AutonomyTransition (5-level autonomy hierarchy)
- RuleEngine (competence evidence-based rule engine)

**Sandboxed Evolution Protocol**
- Isolated workspace (`.axis/evolution/<run-id>/`)
- Atomic evolution steps + append-only trace ledger
- Explicit verify / promote / discard gates
- Full audit trail

## Milestone 6: Self-Judgement ✅ Completed

**Core Components**
- JudgementCriteria / JudgementResult / JudgementItem
- SelfJudgementEngine
- 5 built-in validation strategies: Syntax / Semantic / Contract / Test / Coverage
- `self/judge-execution` contract
- BootstrapOrchestrator judgement integration
- CLI `axis judge` diagnostic command

## Evolution Principles

- Each milestone is independently verifiable
- Subsequent milestones do not modify core scheduling semantics
- Maintain backward compatibility
- Occam's Razor: minimum viable, progressive enhancement
- **More Context, More Action, Zero Control, Controllable Evolution**
