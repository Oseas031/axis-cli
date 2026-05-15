# Axis

> Infrastructure for Objectification — providing the conditions for intent to become objective existence through AI, without prescribing the path of transformation.

Axis is not a task scheduler, nor an LLM wrapper framework. Axis is **Objectification infrastructure**: the substrate that enables human intent (and eventually Agent intent) to be transformed into objective, verifiable, evolvable artifacts — code, documents, systems — through the dialectical process of Construct (对象化), Constraint (规定性), and Judge (扬弃).

**[Chinese version / 中文版](docs/zh/ROOT-README.md)**

## Core Thesis

```text
More Context, More Action, Zero Control, Controllable Evolution
```

- **More Context**: The system provides query infrastructure; Agents actively query and assemble context rather than passively receiving redundant pushes
- **More Action**: Execution, composition, validation, correction, and follow-up task generation capabilities, with permissions matched to competence
- **Zero Control**: The system provides contracts, infrastructure, and observability, but does not prescribe a single action path for the Agent
- **Controllable Evolution**: Self-bootstrapping, self-generation, and self-modification must remain within observable, verifiable, and rollback-safe boundaries

## Design Principles

```text
bash is all you need · Competence earns autonomy · Interface is existence
```

- **CLI First**: Scriptable, composable, callable by humans/CI/Agents; no default Web UI or complex TUI
- **Progressive Autonomy**: The more reliable, the wider the action radius; high-risk operations are not exempt based on executor identity
- **Interface is Existence**: Humans and Agents implement the same agent interface, with no identity bias
- **Contract is Structure**: File system / meta-files are the shared contract language for all Agents
- **Transitional Structures**: workflow/contract/permission/spec are seeds and scaffolding, eventually to be internalized, rewritten, and discarded by Agents

## Current Status

M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Staged Evolution ✅ | Local Control Plane ✅

### Completed Capabilities

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
go test ./...
go build -o axis-dev.exe ./cmd/axis
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
| `axis gui [--port 3000]` | Launch observation dashboard (Web UI) | No |

## External Tools

- **[axis-gui](tools/axis-gui/)**: Local Web GUI connecting to the Local Control Plane, providing Dashboard / Tasks / Providers / Events / Chat views (WebSocket real-time updates)
- **[axis-up](tools/axis-up/)**: Guided onboarding tool for environment detection / build / configuration / demo

Both tools do not import Axis internal packages; they communicate via CLI and HTTP API.

## Key Documentation

- [Dialectical Development Methodology](docs/architecture/dialectical-development-methodology.md) **<-- read first: how decisions are made**
- [Agent-Native First Principles](docs/architecture/agent-native-first-principles.md) **<-- read before coding**
- [Kernel Abstraction Model](docs/architecture/kernel-abstraction-model.md) — syscall layer, core abstractions, infrastructure
- [Bash is All You Need](docs/architecture/bash-is-all-you-need.md)
- [System Conventions](docs/architecture/axis-system-conventions.md)
- [Current Progress](docs/status/current-progress.md)
- [Documentation Index](docs/README.md)
- [Agent-Native Scenarios Whitepaper](docs/product/axis-native-scenarios-whitepaper.md)
- [Autogenesis Design Report](reports/strategy/axis-autogenesis-design-2026-05-08.md)

## Tech Stack

- **Go 1.21+**, core modules prefer the standard library
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

- **Autonomy negative feedback**: Auto-downgrade on consecutive failures, permission scope enforcement
- **Runaway detection**: Semantic progress tracking (repeated tool output → abort)
- **Real-world validation**: Run 10+ coding tasks end-to-end, calibrate all theoretical parameters
- **Cross-process state persistence**: ReadinessRegistry integration with Local Control Plane
- **Structured event log queries**: `axis audit` or equivalent capability
- **Dynamic model routing**: Cost/latency-aware provider selection + degradation chain

