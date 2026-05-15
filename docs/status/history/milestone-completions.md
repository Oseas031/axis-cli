# Milestone Completions (M1-M6)

> Historical record. All tasks completed. See [current-progress](../current-progress.md) for active status.

## M1 — Project Foundation
- Fix staticcheck ST1003, contract executor enum, CI workflow godoc
- Workflow improvement plan, document audit, cleanup
- Claude Code workflow continuity guide
- Folder reorganization, workflow deprecated check
- Milestone 1 acceptance passed + report generated

## M2 — DAG Parallel Scheduling
- Spec skeleton (DAG, contract admission, SLA, error codes)
- T1-T7: scheduler ready-set API, contract admission, SLA parsing, orchestrator parallel loop, error codes, CLI/docs
- Coverage raised to 75.7%
- GitHub CLI installed, pre-commit hooks fixed, CI workflows fixed

## M3 Phase 1 — ModelProvider
- ModelProvider interface + MockModelProvider
- Dispatcher → ContractExecutor → ModelProvider execution path
- ErrDependencyNotReady, failed dependency handling
- Coverage raised to 88.8%

## M3 Phase 2 — Configurability
- ModelProvider functional options, EchoModelProvider, NewProvider factory
- DAG enhancements (GetAllTasks, GetDependencyGraph)
- HumanExecutor routing + polling wait
- Shell dag/resolve commands

## M3 Phase 3 — SLA + Tools
- SLA failure_class routing + backoff strategies + priority ordering
- Tool interface + ToolRegistry + BashTool
- ModelRequest/Response extension (Tools/History/ToolCalls)
- Multi-turn execution loop (max 10 turns)
- Coverage 87.1%

## M4 — Real LLM Integration

### Original (T1-T18)
- ProviderConfig + functional options
- AnthropicProvider + OpenAIProvider (with MiniMax endpoint)
- FileReadTool / FileWriteTool / HTTPClientTool + permission scopes
- Circuit breaker, CLI provider commands
- Coverage 93.7%

### Gap Fix (T19-T22)
- CLI env fallback (ANTHROPIC/OPENAI/DEEPSEEK/MINIMAX API keys)
- Default model corrections
- Empty API key early detection

### Hardening (T23-T28)
- `axis provider test` diagnostic
- Exponential backoff retry (5xx only)
- Tool output truncation (64 KiB)
- Provider structured logging (JSON lines)
- Token cost tracking (CostEstimateUSD)

## M5 — Bootstrap Loop
- AgentExecutor interface + AgentRuntimeAdapter
- SelfContext + ContextBuilder + ContextCompressor
- Self-iteration contracts (6 types)
- BootstrapOrchestrator + FollowUpTaskGenerator
- AutonomyTransition (5-level) + RuleEngine
- Staged Evolution T1-T10 fully completed
- Agent Context Query Model (context.requested_sources)

## M6 — Integration + Memory
- Integration testing + documentation (T14-T18)
- Agent memory system assessment and enhancement

## Coding Agent P0 (2026-05-14)
- LLMAgentExecutor (258 lines, 7 unit tests)
- FileWriteTool path validation hardening
- BashTool Permission Ladder (L0/L1/Unrestricted)
- Guarantee Registry (Hard/Soft promises)
- axis-gui frontend rewrite (5 pages)

## Skills System (2026-05-12)
- internal/skills/: Discover, Load, Validate, parseFrontmatter
- CLI: axis skills list/show/validate/create
- load_skill tool + Layer 1 system prompt injection

## Three-Layer Context Compaction (2026-05-12)
- EstimateTokens, ToolResultCompaction, SummarizationCompaction, TruncationCompaction
- CompactionPipeline integrated into executeMultiTurn
- compact tool registered

## Kernel Abstraction Model (2026-05-12)
- 9 syscall primitives: submit_task, query_state, acquire_context, request_capability, compact, spawn, introspect, yield, checkpoint
- Actor Model + JSONL Mailbox + Router

## axis-gui Toolchain
- Connection fix, proxy fix, font CDN fix, contract ID support
- Scheduler crash recovery, orchestrator busy-poll removal
- WebSocket real-time, task timeline, dark mode
