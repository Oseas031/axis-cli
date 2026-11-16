# Axis

> **Single-Binary Multi-Harness AI Agent Control Tower & Local Coding Orchestrator** (Go, Zero-Framework).  
> Axis unifies native LLM agents and external coding CLI harnesses (`claude`, `codex`, `gemini`, `opencode`, etc.) under one local control plane: **single-command Web workbench (`axis gui`), git worktree isolation, inline human-in-the-loop approvals, and cross-harness pipelines**.

## Core Business Scenario

When developers move from single-turn chat coding to **concurrent, multi-model AI development**, three real-world bottlenecks appear:

1. **Multi-model cost & reliability chaos**: Frontier models burn budgets on trivial tasks, while single-provider blips break long pipelines → solved by **DAG scheduling, per-task USD budget caps, circuit breakers, and model escalation**.
2. **Broken human-agent handoff**: When an Agent hits ambiguity in a background run, it either guesses blindly across 20 files or crashes → solved by **symmetric human-agent scheduling (`human/gate` + `axis inbox`)**, parking tasks until a human responds or safely degrading.
3. **Unsafe & opaque repository mutations**: Blindly stuffing prompts causes Context Rot, and direct writes risk breaking `main` → solved by **opt-in `contextpack` (`preflight --strict`)** and **staged sandbox evolution** (deterministic `go test`/`go vet` gates before explicit `promote`).

## Problem-Driven Architecture (Why Axis?)

Every core subsystem in Axis exists to solve a concrete failure mode in real-world AI coding workflows:

| Real-World Pain Point | Architectural Solution | Forward-Looking Design |
|---|---|---|
| **Ambiguity stalls or ruins runs**: Agents either guess blindly across 20 files or crash when requirements are unclear. | **Park / Nudge / Fail-Closed Gate** (`internal/human/gate`) backed by a persistent JSONL CAS inbox (`internal/comm`, `axis inbox`). | **Human-Agent Symmetric Scheduling**: Humans and Agents act as peer executors under one contract, enabling bidirectional task flow. |
| **Unbounded cost & fragile APIs**: Frontier models burn budgets on trivial steps; single-provider blips break long pipelines. | **DAG `scheduler` + 5-worker `orchestrator`**, per-task USD budget caps, multi-provider circuit breaker, and model escalation. | **Reliability Lab**: Swarm multi-candidate majority voting and measurable cost/quality trade-offs. |
| **Context Rot & blind prompt stuffing**: Dumping full repos into prompts wastes tokens and degrades reasoning quality. | **Opt-in `contextpack`**: relevance scoring, token-budget trimming, and `context preflight --strict` readiness gates. | **Pull-over-Push Context**: Queryable, auditable context assembly rather than hidden prompt injection. |
| **Unsafe self-modification**: Letting Agents directly mutate the main tree or tool permissions risks silent regressions. | **Staged Evolution + Permission Ladder** (Bash L0/L1/Docker sandbox) + deterministic `go test`/`go vet` verification. | **Controllable Evolution**: Isolated atomic steps with trace ledgers and explicit `promote`/`discard` gates. |

## Research Thesis & Design Philosophy

Axis is also an experimental **Objectification infrastructure**: transforming human (and Agent) intent into verifiable, evolvable artifacts through **Construct — Constraint — Judge**.

```text
More Context, More Action, Zero Control, Controllable Evolution
```

- **More Context**: Queryable context infrastructure; Agents assemble context on demand rather than receiving redundant pushes
- **More Action**: Composable execution, validation, and follow-up generation with competence-matched permissions
- **Zero Control**: Contracts and observability without prescribing a single rigid reasoning path
- **Controllable Evolution**: Self-modification strictly bounded by isolated workspaces, deterministic tests, and human sign-off

```text
bash is all you need · Competence earns autonomy · Interface is existence
```

- **CLI First**: Scriptable, composable single binary; no Web/TUI framework in core
- **Progressive Autonomy**: Reliability earns wider action radius; no identity-based exemptions
- **Transitional Structures**: Specs, workflows, and permissions act as inspectable scaffolding rather than opaque black boxes

## Current Status

M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Staged Evolution ✅ | Local Control Plane ✅ | Human-Agent Scheduling (v1) ✅ | Multi-Harness Control Tower & Single-Binary UI ✅

### Completed Capabilities

- **Single-Binary Multi-Harness Control Tower (`axis gui` / `axis web`)**: Zero-bundler 3-pane Web workbench embedded via `go:embed` with SSE real-time stream, `@harness` routing, `Stop` / task cancellation, inline human approvals (with masked secret input), `git worktree` isolation & diff review, Save-as-Skill, Routines, Memory notes, and cross-harness Pipelines (`{{prev_output}}` / `{{prev_diff}}`)

- **Task Scheduling**: FIFO + DAG parallel scheduling, dependency management, 5-worker parallel orchestrator, contract admission, SLA timeout/retry/failure_class strategy engine
- **LLM Integration**: Anthropic / OpenAI / DeepSeek / MiniMax providers, token accounting, circuit breaker (configurable), project-local provider profile management, quality-gated model escalation, semantic layering (primary/utility routing)
- **Tool System**: BashTool (observable execution records), SandboxedBashTool (Docker-based process/network/filesystem isolation), FileReadTool, FileWriteTool, HTTPClientTool, tool permission scopes, multi-turn execution loop (shared `multiturn.Run` with configurable cap + graceful termination), syscall tools (compact/yield/checkpoint)
- **Agent Executor**: LLMAgentExecutor with multi-turn tool loop, circuit breaker, pluggable TerminationFn, HistoryCompactor, EventEmitter, per-turn timeout, AgentID tracking
- **Natural Language Scheduling**: `axis ask` compiles prompts into AgentTask, dry-run preview / explicit submit, never bypasses contracts
- **Adaptive Context Assembly**: ContextBundle / ReadinessArtifact / ReadinessRegistry / preflight / strict gate, rule-based assembly + budget trimming, task-aware relevance scoring, preview-first without execution intrusion
- **Execution-time Context Consumption**: ExecutionContextSummary / ExecutionContextConsumer, Agents declare `context.requested_sources`, dispatcher injects summary
- **Local Control Plane**: `axis start` launches loopback HTTP control server, cross-process submit/query, `.axis/runtime.json` locator, append-only event log, orphaned task recovery on restart
- **Staged Evolution Protocol**: Isolated workspace + atomic steps + trace ledger + verification capture + explicit promote/discard gate, full audit trail
- **Self-Judgement Engine**: SelfJudgementEngine + 5 validation strategies (Syntax/Semantic/Contract/Test/Coverage), context isolation (Context Rot prevention), two-pass escalating judge (lightweight first), generalization scoring, self-judgement contract, BootstrapOrchestrator judgement integration
- **Bootstrap Loop**: BootstrapOrchestrator + FollowUpTaskGenerator + AutonomyTransition rule engine (configurable thresholds) + self-iteration contracts
- **Multi-Agent Infrastructure**: Subagent context isolation (IsolationPolicy), JSONL mailbox (send/receive/mark-read), multi-candidate differential testing (CandidatePool), Swarm Topology (parallel_vote pattern, heterogeneous diversity, majority vote aggregation)
- **Progressive Autonomy**: Feature gate (progressive unlock), dispatcher autonomy resolver (metadata-driven, reads `agent.autonomy_level` from task metadata), capability registry, permission tri-state (ask/allow/deny), cost budget enforcement (per-task USD limit + auto-downgrade)
- **Dispatcher**: Audit log (in-memory + external audit callback), configurable timeout, tool filter per agent profile
- **Security**: BashTool Permission Ladder (L0/L1/Unrestricted), SandboxedBashTool (Docker container isolation), FileWriteTool path validation hardened
- **Guarantee Registry**: Hard/Soft level system promises, Register/Verify/List
- **9+ structured error codes**, Agent Context Query Model, DAG visibility

### Current Limitations

> Honesty over aspiration. These are known architectural boundaries, not bugs.

- **Single-machine only**: All state is in-memory, control plane is loopback HTTP. No distributed scheduling, no leader election, no horizontal scaling.
- **No multi-tenant support**: Single-user, single-project design. No user isolation, no priority preemption.
- **Docker sandbox requires Docker**: SandboxedBashTool falls back to direct execution (with warning) if Docker is unavailable. Without Docker, there is no OS-level isolation.
- **Self-Judgement is advisory**: LLM judging LLM has unresolvable same-source bias. Only compiler/test results are authoritative verification.
- **No production validation**: All parameters (iteration caps, circuit breaker thresholds, context budgets) are theoretical values pending real-world calibration.

## Quick Start

```bash
# Build & verify without any API key (uses built-in mock provider)
go build -o axis-dev.exe ./cmd/axis
go test -race -short ./internal/kernel/... ./internal/human/... ./cmd/axis
.\axis-dev.exe ask "check provider config"       # dry-run intent -> AgentTask preview
.\axis-dev.exe context preview "check config"    # opt-in context assembly preview
```

> On Windows, output to `axis-dev.exe` to avoid overwriting or locking an existing `axis.exe` in the project root.

### Local Runtime

Cross-command submit and query requires explicitly starting a local runtime:

```powershell
# Terminal A: start the project-local runtime
.\axis-dev.exe start

# Terminal B: submit a natural-language task
.\axis-dev.exe ask "check provider config" --submit --task-id provider-check

# Terminal B: query task status
.\axis-dev.exe status provider-check
```

- `axis start` writes `.axis/runtime.json`, exposes a loopback control server, and appends events to `.axis/events/tasks.jsonl`
- `axis ask <prompt>` defaults to dry-run preview and does not require a runtime
- `axis shell` is an in-process session; `run`/`ask --submit`/`status` within the shell share session state and do not silently attach to `axis start`

### Provider Management

```powershell
# Add project-local provider profiles
.\axis-dev.exe provider add claude --type anthropic --api-key sk-ant-... --model claude-3-5-sonnet-20241022
.\axis-dev.exe provider add gpt --type openai --api-key sk-... --model gpt-4o-mini
.\axis-dev.exe provider add ds --type deepseek --api-key sk-... --model deepseek-chat
.\axis-dev.exe provider add mm --type minimax --api-key ... --model MiniMax-Text-01

# Switch / inspect / list
.\axis-dev.exe provider use claude
.\axis-dev.exe provider status
.\axis-dev.exe provider list
```

Profiles are stored in `.axis/providers.json` and do not modify shell environment variables or system configuration.

### Context Preview and Readiness Check

```powershell
# Preview context assembly result for a task
.\axis-dev.exe context preview "check provider config"

# Check context readiness
.\axis-dev.exe context inspect <bundle-id>
.\axis-dev.exe context preflight <task-id>
.\axis-dev.exe context preflight <task-id> --strict
```

### Staged Evolution

```powershell
# Inspect evolution run details
.\axis-dev.exe evolve inspect <run-id>

# Promote or discard evolution results
.\axis-dev.exe evolve promote <run-id>
.\axis-dev.exe evolve discard <run-id>
```

### Self-Judgement

```powershell
# Run self-judgement diagnostic
.\axis-dev.exe judge
```

## CLI Commands

| Command | Purpose | Requires Runtime |
|---------|---------|:---:|
| `axis run <task-id>` | Execute a task synchronously (in-process) | No |
| `axis run <task-id> --background` | Submit task to runtime, return immediately | **Yes** |
| `axis start` | Start local runtime (loopback control server) | N/A (creates it) |
| `axis status <task-id>` | Query task status (via local runtime) | **Yes** |
| `axis ask <prompt>` | Natural language to AgentTask (dry-run by default) | No |
| `axis ask <prompt> --submit` | Submit natural-language task to local runtime | **Yes** |
| `axis shell` | Start interactive in-process shell | No |
| `axis provider add/use/status/list/remove/archive` | Manage project-local LLM provider profiles | No |
| `axis context preview/inspect/preflight` | Context assembly preview and readiness check | No |
| `axis judge` | Run self-judgement diagnostic | No |
| `axis evolve inspect/promote/discard` | Staged evolution inspection and decisions | No |
| `axis skills list/show/validate/create` | Manage on-demand knowledge skills | No |
| `axis vigil resume/list/add/start/done/show/triage` | Cross-session work tracking | No |
| `axis inbox list/show/respond` | Human inbox: answer agent decision/info requests (human-agent-scheduling) | Optional |
| `axis gui [--port 3000]` | Launch observation dashboard (Web UI) | No |

## External Tools

- **[axis-gui](tools/axis-gui/)**: Local Web GUI connecting to the Local Control Plane, providing Dashboard / Tasks / Providers / Events / Inbox / Chat views (WebSocket real-time updates)
- **[axis-up](tools/axis-up/)**: Guided onboarding tool for environment detection / build / configuration / demo

Both tools do not import Axis internal packages; they communicate via CLI and HTTP API.

## Key Documentation

- [CLAUDE.md](CLAUDE.md) **<-- authoritative charter, workflow & module boundaries**
- [Current Progress](docs/status/current-progress.md) — milestone & cleanup ground truth
- [Semantic Boundaries](docs/architecture/semantic-boundaries.md) — what each module must NOT do
- [Git Conventions](docs/architecture/git-conventions.md) — streamlined single-operator commit workflow
- [Documentation Index](docs/README.md) — full docs router (specs, guides, historical reports)

## Tech Stack

- **Go 1.26+**, core modules prefer the standard library
- **Single binary CLI**, shell-native workflow
- **Cobra** CLI framework
- **Project-local state**: `.axis/` directory (providers.json / runtime.json / events/ / evolution/)

## Project Structure

```text
cmd/axis/          CLI entry and command definitions
internal/
  types/           Core data types (AgentTask, AgentContract, ErrorCode...)
  kernel/          Scheduler, orchestrator, dispatcher, feature gate, capability registry
  contract/        Contract executor (permission scopes, circuit breaker, compaction)
  model/           LLM provider (escalation, layering) + tool system + multiturn loop
  agent/           Agent executor + self-judgement engine + candidates + relevance scoring
  intent/          Natural language intent parsing
  contextpack/     Adaptive context assembly
  control/         Local control plane (server/client/locator/events)
  evolution/       Staged evolution protocol
  comm/            Multi-agent communication (JSONL mailbox)
  skills/          On-demand knowledge skills (loader + metadata)
  memory/          Memory subsystems (horizon/immediate/immunity/kv/longterm/working)
  guarantee/       System guarantee registry (Hard/Soft promises)
  human/           Human executor
  vigil/           Cross-session work tracking
docs/              Documentation index, architecture reference, specs, status
tools/
  axis-gui/        Local Web GUI (Observatory)
  axis-up/         Guided onboarding tool
```

## Next Steps

- **Freeze horizontal expansion**: No new independent subsystems until the existing core loop is validated end-to-end
- **Minimal Spec → DAG + Deterministic Gate loop**: Keep human-owned Spec/boundaries; wire deterministic `go test`/`go vet` gates and `human/gate` (`axis inbox`) final sign-off
- **Real-world validation**: Run 10+ coding tasks end-to-end, calibrate timeouts, circuit breakers, and context budgets
- **Debt cleanup**: Reconcile pending vigil items and finish remaining `human-agent-scheduling` polish

