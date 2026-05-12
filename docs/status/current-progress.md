# Current Progress

**[Chinese version / 中文版](../zh/status/current-progress.md)**

**Updated**: 2026-05-12
**Current Milestones**: M1 ✅ | M2 ✅ | M3 (Phase 1-3) ✅ | M4 ✅ (Original T1-T18 + Gap Fix T19-T22 + Hardening T23-T28) | M5 ✅ | M6 ✅

## Current Design Positioning

Axis is no longer defined as a generic Agent scheduling platform, but as the early execution substrate for Agent autogenesis.

Key observations:

- The bootstrap origin has already occurred: external Agents are injecting ideas of observable solidification, execution, reflection, and evolution into Axis
- M2 is not an ordinary parallel scheduling milestone, but the execution substrate for the future Autogenesis Loop
- workflow is temporary scaffolding, contract is a growth boundary, permission rule is a progressive autonomy mechanism, spec is a seed
- M2 fully completed
- M3 Phase 1 completed: ModelProvider execution path opened, coverage 88.8%, DAG/SLA supplemented

## Completed Tasks (M1)
- [x] Fix staticcheck ST1003 error (shared_layer → sharedlayer)
- [x] Fix contract executor enum validation logic (support int type)
- [x] Fix CI workflow godoc -html deprecated parameter
- [x] Create workflow improvement plan and fix high-priority issues
- [x] Document audit and cleanup (moved 4 obsolete docs to deprecated)
- [x] Create document audit workflow (document-audit.yml)
- [x] Create Claude Code workflow continuity guide
- [x] Workflow cleanup (update workflow registry, create workflow index)
- [x] Folder reorganization (create reports/ and docs/deprecated/workflows/)
- [x] Workflow deprecated content check and risk assessment
- [x] Delete unused docs job
- [x] Workflow experience summary and improvements (create registry-validator.yml)
- [x] Daily retrospective report
- [x] Milestone 1 acceptance passed
- [x] Generate Milestone 1 acceptance report

## Completed Tasks (M2)
- [x] Create M2 spec document skeleton (DAG parallel scheduling, contract admission rules, SLA, error codes)
- [x] Supplement M2 workflow binding (bind wf-doc-004 + wf-pr-check + wf-ci + wf-doc-006)
- [x] M2 workflow binding confirmed
- [x] T1 baseline verification completed (local CI-equivalent coverage 62.8%, exceeding 60% gate)
- [x] T2 scheduler ready-set API completed (`GetReadyTasks(limit int)`, CI-equivalent coverage 63.6%)
- [x] Install and run GitHub Actions equivalent tools: `staticcheck`, `gosec`, `govulncheck`, `markdownlint`
- [x] T2.5 standard CLI Bash-first semantic fix completed (CI-equivalent coverage 67.3%)
- [x] T3 contract admission layer implementation completed (CI-equivalent coverage 68.1%)
- [x] T4 SLA parsing and execution timeout implementation completed (CI-equivalent coverage 69.2%)
- [x] T5 orchestrator parallel execution loop implementation completed (CI-equivalent coverage 69.3%)
- [x] T6 error code basic implementation completed
- [x] T7 CLI/docs update and acceptance completed
- [x] Test coverage raised to 75.7%
- [x] **Milestone 2 fully completed**
- [x] All work categorized by single upstream workflow, experience audited and hardened into workflow rules
- [x] Core documentation rewritten with autogenesis / Autogenesis design positioning
- [x] Created CLAUDE.md for Claude Code integration (includes full project context, build commands, architecture overview)
- [x] GitHub CLI (gh v2.92.0) installed and authenticated as Oseas031
- [x] Pre-commit hook fix: Windows Python compatibility (bash wrapper), registry path updates, Unicode safe output
- [x] Registry fix: registered wf-entry, fixed wf-release file references, dependency chain consistency
- [x] CI Workflow fix: registry-validator scope bug, ci.yml dead conditions, document-audit M2 semantics, CODING_STANDARDS update
- [x] PR Quality Check fix: documentation-check git diff shallow clone failure → all 4 jobs passing
- [x] Monitoring fault diagnosis: 3 job failures root-caused, fixes on milestone1-acceptance branch pending PR merge
- [x] lmh-harness-v1 engineering methodology integrated
- [x] Project memory system initialized (GitHub CLI first preference)

## Completed Tasks (M3 Phase 1)
- [x] ModelProvider interface + MockModelProvider (provider.go, mock.go)
- [x] Dispatcher → ContractExecutor → ModelProvider execution path opened
- [x] `ErrDependencyNotReady` error code + `sla.failure_class` SLA constants
- [x] Failed dependency handling (failed = done, no longer permanently blocking downstream)
- [x] `types_test.go` covering AgentError/ErrorCode/SLA/FieldType/TaskStatus
- [x] Orchestrator retry exhaustion test (output validation failure triggers → retry → exhaust → Failed)
- [x] cmd/axis shell stdin simulation test (help/exit/unknown/run/status/empty/quit)
- [x] Dispatcher parent context cancel + errChan path test
- [x] Executor SetProvider + Execute with provider + ValidateOutput test
- [x] Admission empty/valid failure_class SLA test
- [x] Test coverage raised to 88.8% (exceeding 85% target)
- [x] Worktree isolation defect investigation (EnterWorktree based on default branch main HEAD, not current branch HEAD)
- [x] Manual worktree parallel development Plan B verified (git worktree add -b + EnterWorktree --path)

## Completed Tasks (M3 Phase 2)
- [x] ModelProvider configurability (Functional Options Pattern: WithModelProvider)
- [x] EchoModelProvider (distinct from MockModelProvider)
- [x] NewProvider factory function (supports "mock", "echo")
- [x] DAG enhancements: GetAllTasks, GetDependencyGraph (scheduler + orchestrator)
- [x] Shell dag command (readable dependency graph output)
- [x] HumanExecutor routing: TaskMetadataKeyExecutor + dispatcher executeHumanTask
- [x] HumanExecutor polling wait + timeout mechanism
- [x] Orchestrator ResolveCall exposed + Shell resolve command
- [x] Test coverage: provider registry, scheduler DAG, dispatcher human routing, shell dag/resolve
- [x] Coverage maintained at 86.8% (exceeding 85% target)

## Completed Tasks (M3 Phase 3)
- [x] SLA type constants: failure_class (retryable/fatal/degradable), priority, backoff
- [x] Failure class routing + backoff strategies (orchestrator: parseSLA extension, backoffDelay, fatal/degradable branches)
- [x] Priority ordering (scheduler: GetReadyTasks sorted by priority descending, same-priority FIFO)
- [x] SLA admission validation extension (priority 0-255, backoff enum, failure_class enum)
- [x] Tool interface + ToolRegistry (pluggable tool registration)
- [x] BashTool (os/exec execution, 30s timeout, returns stdout/exit_code)
- [x] ModelRequest/ModelResponse extension (Tools + History + ToolCalls fields)
- [x] MockModelProvider tool-aware (simulates multi-turn tool-use)
- [x] Multi-turn execution loop (ContractExecutor: provider → tool → provider, max 10 turns)
- [x] Orchestrator assembly (ToolRegistry + BashTool injection)
- [x] Test coverage 24+ new cases, coverage 87.1%
- [x] axis-dev.exe compilation passed

## Completed Tasks (M4 - Real LLM Integration + Extended Tools)

### Original M4 (T1-T18)
- [x] ProviderConfig + functional options (`WithModel`, `WithAPIKey`, `WithBaseURL`, `WithTimeout`, `WithMaxRetries`)
- [x] AnthropicProvider (`/v1/messages`, `x-api-key`, tool_use parsing, token accounting)
- [x] OpenAIProvider (`/v1/chat/completions`, SSE, function calling, MiniMax endpoint fix)
- [x] FileReadTool / FileWriteTool / HTTPClientTool with permission scopes
- [x] Tool permission metadata (`tool.allowed_tools`, `tool.allowed_paths`, `tool.allowed_hosts`)
- [x] Circuit breaker in ContractExecutor (consecutive tool error threshold)
- [x] CLI `--provider` / `--model` flags with project-local profile resolution
- [x] `axis provider add/list/use/status/remove/archive` commands
- [x] Provider tests with httptest (coverage 91.8%)
- [x] Tool tests with path escape regression coverage (coverage 93.7%)

### M4 Phase 4.5: CLI Usability Gap Fix (T19-T22)
- [x] T19: CLI env fallback — `providerOptions()` falls back to `ANTHROPIC_API_KEY` / `OPENAI_API_KEY` / `DEEPSEEK_API_KEY` / `MINIMAX_API_KEY` when no active profile
- [x] T20: Default model correction — `deepseek-chat` (deprecated) → `deepseek-v4-flash`; `MiniMax-Text-01` (incompatible endpoint) → `MiniMax-M2.7`
- [x] T21: Empty API key early detection — `NewProvider` returns actionable error at construction time instead of 401 at request time
- [x] T22: Regression tests for env fallback, default models, and missing key scenarios

### M4 Phase 4.6: Non-Destructive Provider & Tool Hardening (T23-T28)
- [x] T23: `axis provider test` diagnostic command — lightweight ping to verify API key, network reachability, and model name acceptance
- [x] T24: Exponential backoff retry — `providerConfig.maxRetries` now enforced in AnthropicProvider and OpenAIProvider; retries only on 5xx / network errors, not 4xx
- [x] T25: Tool output truncation — FileReadTool and HTTPClientTool capped at 64 KiB (matching BashTool), returns `truncated: true` + `total_bytes`
- [x] T26: Provider structured logging — `AXIS_PROVIDER_LOG=1` outputs JSON lines with `provider/method/url/status/duration_ms/input_tokens/output_tokens/cost_usd` (no apiKey in logs)
- [x] T27: Token cost tracking — `ModelResponse.CostEstimateUSD` field with lightweight per-model pricing table (anthropic/openai/deepseek/minimax)
- [x] T28: Provider status enhancement — `axis provider status` now prompts user to run `axis provider test` for connectivity verification

## Completed Tasks (M5 - Bootstrap Loop)
- [x] AgentExecutor interface + MockAgentExecutor implementation
- [x] AgentRuntimeAdapter (external Agent CLI support)
- [x] Orchestrator AgentExecutor injection (WithAgentExecutor option)
- [x] SelfContext data structure (TaskID, TaskLineage, CodeSnapshot, DocSnapshot, StateSnapshot)
- [x] ContextBuilder implementation (BuildSelfContext method)
- [x] ContextCompressor (context compression, 3 strategies)
- [x] Self-iteration Contracts (analyze/implement/validate/update/review/spawn)
- [x] BootstrapOrchestrator (self-loop task scheduling + loop tracking)
- [x] FollowUpTaskGenerator (generate follow-up tasks from execution results)
- [x] AutonomyTransition data model (5-level autonomy)
- [x] RuleEngine (competence evidence-based rule engine)
- [x] Integration tests (Full DAG workflow, concurrent tracking)
- [x] M5 documentation updates (requirements.md, design.md, tasks.md marked Complete)
- [x] Phase 2: Sandboxed Evolution T1-T10 fully completed (data model/storage/workspace/ledger/verification/inspect/promote/discard/tests/docs)
- [x] Agent Context Query Model T8 completed: `context.requested_sources` metadata key, `ExecutionContextSummary` 3-field extension, `AgentExecutionRequest.RequestedSources`, dispatcher duplicate parsing elimination, review fixes

## Completed Tasks (M6)
- [x] M6 Phase 6.4-6.5: Integration + Testing + Documentation (T14-T18 all completed)
- [x] Phase 3: Agent memory system assessment and enhancement

## Completed Tasks (axis-gui toolchain)
- [x] axis-gui connection fix: absolute path resolution for Go 1.19+ exec security restrictions
- [x] axis-gui proxy fix: header-first write order fix for HTTP 500
- [x] axis-gui font CDN fix: replaced broken fontsource URL
- [x] axis-gui contract ID support: frontend submitTask added contract_id field
- [x] axis-gui error enhancement: extract message field from backend response for accurate display
- [x] T1: scheduler crash recovery test + implementation (stale Running → Failed)
- [x] T2: orchestrator busy-poll removal (time.After replaced with taskSubmitted channel signal)
- [x] T3: full regression test passed go test ./...
- [x] T5: TasksPage WebSocket real-time integration (5s polling → WebSocket driven + 30s fallback)
- [x] T5: WebSocket connection status visualization (live badge dynamically reflects connection state)
- [x] T6: task timeline aggregation (group events by task_id, expand to show full event timeline)
- [x] T7: dark mode system preference listener (auto-follows prefers-color-scheme when not manually set)

## Issues Encountered
- ✅ staticcheck ST1003 — Fixed (commit 1d9aaef, 37f23c0)
- ✅ godoc -html deprecated parameter — Fixed (commit 457b30a)
- ✅ Enum validation not supporting int type — Fixed (commit 5c4231f)
- ✅ Outdated documentation — Cleaned (commit b323b7d)
- ✅ Missing document audit workflow — Created (commit bb2045f)
- ✅ Inconsistent workflow registry — Cleaned (commit f1fde53)
- ✅ Unused content — Partially fixed (docs job deleted, commit 27b94c5)
- ✅ release.yml and cd-workflow duplicate — Handled (release.yml deleted, registry marked deprecated)
- ⚠️ sign-artifacts job unused — Pending (post-milestone)
- ✅ T1 GitHub CI-equivalent coverage gate met: total coverage 62.8%
- ✅ T2 GitHub CI-equivalent coverage gate still met: total coverage 63.6%
- ✅ `staticcheck ./...` local pass
- ✅ `gosec ./...` local pass, Issues: 0
- ✅ `govulncheck ./...` local pass
- ✅ T2.5 through T5 CI-equivalent coverage gates all met (67.3% → 69.3%)
- ✅ Test coverage raised to 75.7% (exceeding 75% target)
- ✅ Test coverage further raised to 88.8% (exceeding 85% target)
- ⚠️ Isolation worktree based on stale commit (main HEAD) not current branch HEAD → Root-caused, using manual worktree Plan B
- ⚠️ Windows does not support programmatic signal sending → SIGINT-related tests removed
- ⚠️ `markdownlint "**/*.md"` local found existing Markdown style issues; consistent with document-audit.yml, this check is currently non-blocking advisory
- ✅ Workflow retrospective appended to `reports/daily/workflow-system-retrospective-2026-05-08.md`
- ✅ Retrospective experience hardened into `workflow/entry.md`, `workflow/meta-workflow-management.md`, `workflow/occams-razor-architecture-simplification.md`
- ✅ PR Quality Check git diff shallow clone failure — Fixed (commit f9962de, added fetch-depth:0 + || true)
- ✅ Monitoring 3 job failures — Fixed on milestone1-acceptance branch, pending PR merge to main
- ✅ M4 T19-T22 gap fix: CLI env fallback, default model correction, empty key early detection, regression tests
- ✅ M4 T23-T28 non-destructive hardening: provider test, retry, truncation, structured logging, cost tracking, status enhancement

## Completed (2026-05-12)

### Skills System (full implementation)
- [x] internal/skills/: Discover, Load, Validate, parseFrontmatter
- [x] CLI: axis skills list/show/validate/create
- [x] load_skill tool registered in ToolRegistry
- [x] Layer 1 system prompt injection (BuildSkillsPromptSection)
- [x] Boundary enforcement tests (scheduler isolation, path safety)

### Three-Layer Context Compaction
- [x] EstimateTokens (4-char heuristic)
- [x] ToolResultCompaction (truncate old tool results)
- [x] SummarizationCompaction (LLM-based summarization)
- [x] TruncationCompaction (sliding window backstop)
- [x] CompactionPipeline integrated into executeMultiTurn
- [x] compact tool registered

### Structural Fixes
- [x] internal/project/root.go: ResolveRoot (walk-up .axis/ discovery)
- [x] Replaced 6 scattered os.Getwd() + .axis path constructions

### Kernel Abstraction Model
- [x] docs/architecture/kernel-abstraction-model.md (syscall layer, 4 core abstractions, infra layer)
- [x] Axis repositioned as "OS for Agents"
- [x] All 9 syscall primitives implemented: submit_task, query_state, acquire_context, request_capability, compact, spawn, introspect, yield, checkpoint

### Actor Model & Communication Layer (P0)
- [x] internal/actor/: Actor interface, Message, MessageType, ActorStatus
- [x] internal/comm/: JSONL Mailbox (Send/Receive/Peek/Ack) + Router
- [x] yield tool: voluntary execution pause
- [x] checkpoint tool: intermediate state persistence
- [x] spawn tool: isolated subtask creation (full/shared isolation)

### Architecture Decisions
- [x] Homogeneous Actor model chosen (人机同构)
- [x] Actor-Comm spec triplet created (docs/specs/actor-comm/)
- [x] Unified Actor Model & Communication Layer planned (P0-P3 phases)

## Important Reminders
- M1 ✅ | M2 ✅ | M3 Phase 1-3 ✅ | **M4 ✅ (Original T1-T18 + Gap Fix T19-T22 + Hardening T23-T28)** | M5 ✅ | M6 ✅ Completed
- Coverage stable, all packages passing `go test ./...`
- SLA strategy engine: supports failure_class routing (retryable/fatal/degradable), backoff strategies, priority ordering
- Tool layer: Tool interface + BashTool + FileReadTool + FileWriteTool + HTTPClientTool, all with 64 KiB truncation
- Provider layer: Anthropic + OpenAI + DeepSeek + MiniMax, with env fallback, retry, cost tracking, and `axis provider test` diagnostic
- Follow Occam's Razor principle
- Continue maintaining CLI-first / shell-native, no Web UI or heavy TUI
- All work progress must be recorded in documentation
- Handover checklist must be completed before handover
- worktree isolation has known defect (based on main HEAD), use manual worktree (Plan B) for parallel development

## Current Spec Documents
- Milestone 2 Requirements: `docs/specs/milestone2/requirements.md`
- Milestone 2 Design: `docs/specs/milestone2/design.md`
- Milestone 2 Tasks: `docs/specs/milestone2/tasks.md`
- Milestone 2 Workflow Binding: `docs/specs/milestone2/workflow-binding.md`
- M4 Requirements: `docs/specs/m4/requirements.md`
- M4 Design: `docs/specs/m4/design.md`
- M4 Tasks: `docs/specs/m4/tasks.md`

## Architecture Diagnosis & Strategic Direction (2026-05-11)

**Full analysis**: `reports/strategy/agent-native-scenario-principles-analysis-2026-05-11.md`
**First principles reference**: `docs/architecture/agent-native-first-principles.md`

### Top 8 Core Gaps Identified

| # | Gap | Severity | Impact on Scenarios | Status |
|---|---|---|---|---|
| A | Cross-process context fracture: ReadinessRegistry is in-process only; no local-control-plane awareness | **Critical** | Breaks multi-terminal workflow, cluster ops, enterprise audit | Partial: Local Control Plane T1-T8 completed |
| B | Orchestrator is pseudo-parallel single-thread loop; no inter-Agent collaboration primitives | **Critical** | Blocks AI-native startup pipelines, digital workforce chaining | Open |
| C | Event log is append-only but lacks structured query API or feedback loop | **High** | Prevents competence-based autonomy, organizational intelligence | Open |
| D | ~~Sandboxed Evolution is spec-only; zero implementation~~ | **Critical** | ~~"Controllable Evolution" = "No Evolution"~~ | **RESOLVED** — T2-T10 fully implemented (2026-05-11) |
| E | Tool boundaries are static fences, not dynamic ladders | **High** | "Competence earns autonomy" remains unimplemented | Open |
| F | Model routing is manual gearbox; no latency/cost-aware dynamic scheduling or fallback | **High** | Token cost uncontrolled, no auto-degradation on provider failure | Partial: T27 token cost tracking implemented (`ModelResponse.CostEstimateUSD`); full dynamic routing remains planned |
| G | No Agent identity or capability profile; only AgentTask exists | **High** | Cannot route tasks to best-fit Agent; "Capability is decision right" needs identity | Open |
| H | Execution feedback loop is fully broken; no quality assessment or system improvement from results | **High** | Same errors repeat; system is open-loop, not closed-loop | Open |

### Design Philosophy Assessment

- **Still fully applicable**: "More Context, More Action, Zero Control", "bash is all you need", "Interface is existence", "Contract is structure"
- **Partially resolved**: "Query is context" — T8 implemented: Agent declares needs via `context.requested_sources`, system resolves against readiness registry and reports satisfied/missing. Full Agent-driven demand (P1+) remains planned.
- **Needs refinement**: "Ladder is boundary" (static fences vs dynamic ladders)
- **Resolved**: "Controllable Evolution" — Sandboxed Evolution P0 fully implemented with isolated workspaces, atomic steps, verification gates, and explicit promote/discard decisions
- **M4 validated**: "Competence earns autonomy" — T24 retry, T25 truncation, T26 logging, T27 cost tracking all demonstrate system-level competence improvement without user control increase. T23 `axis provider test` embodies "bash is all you need" diagnostic philosophy.
- **Fundamental challenge**: Autonomy-reliability tension requires graduated autonomy (P0 high-control, P1 partial, P2 full-in-sandbox)

### Recommended Next Priority Order

1. ~~**Sandboxed Evolution P0 implementation**~~ — **COMPLETED** (2026-05-11)
2. **Cross-process state persistence**: Make ReadinessRegistry local-control-plane-aware
3. **Agent identity & capability profile**: Introduce Agent registry and behavioral scoring
4. **Event log structured query**: Add `axis audit` or equivalent for log consumption
5. **Dynamic model routing**: Cost/latency-aware provider selection with fallback chains
6. **Execution feedback loop**: Result quality scoring feeding back into intent/context assembly
