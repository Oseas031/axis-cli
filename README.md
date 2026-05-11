# Axis

> Agent-native scheduling system. Not to control Agents, but to let Agents earn competence through task practice, gain autonomy, and ultimately generate themselves.

Axis is not a smarter task queue, nor an LLM wrapper framework. Axis is an **Agent autogenesis execution substrate**: enabling Agents to understand tasks, organize actions, validate results, reflect on failures, generate follow-up tasks, and progressively earn greater autonomy through reliable performance.

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

M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Sandboxed Evolution ✅ | Local Control Plane ✅

### Completed Capabilities

- **Task Scheduling**: FIFO + DAG parallel scheduling, dependency management, 5-worker parallel orchestrator, contract admission, SLA timeout/retry/failure_class strategy engine
- **LLM Integration**: Anthropic / OpenAI / DeepSeek / MiniMax providers, token accounting, circuit breaker, project-local provider profile management
- **Tool System**: BashTool (observable execution records), FileReadTool, FileWriteTool, HTTPClientTool, tool permission scopes, multi-turn execution loop
- **Natural Language Scheduling**: `axis ask` compiles prompts into AgentTask, dry-run preview / explicit submit, never bypasses contracts
- **Adaptive Context Assembly**: ContextBundle / ReadinessArtifact / ReadinessRegistry / preflight / strict gate, rule-based assembly + budget trimming, preview-first without execution intrusion
- **Execution-time Context Consumption**: ExecutionContextSummary / ExecutionContextConsumer, Agents declare `context.requested_sources`, dispatcher injects summary
- **Local Control Plane**: `axis start` launches loopback HTTP control server, cross-process submit/query, `.axis/runtime.json` locator, append-only event log
- **Sandboxed Evolution Protocol**: Isolated workspace + atomic steps + trace ledger + verification capture + explicit promote/discard gate, full audit trail
- **Self-Judgement Engine**: SelfJudgementEngine + 5 validation strategies (Syntax/Semantic/Contract/Test/Coverage), self-judgement contract, BootstrapOrchestrator judgement integration
- **Bootstrap Loop**: BootstrapOrchestrator + FollowUpTaskGenerator + AutonomyTransition rule engine + self-iteration contracts
- **9+ structured error codes**, Agent Context Query Model, DAG visibility

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

### Sandboxed Evolution

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

| Command | Purpose |
|---------|---------|
| `axis start` | Start local runtime (loopback control server) |
| `axis run <task-id>` | Submit and run a task |
| `axis status <task-id>` | Query task status (via local runtime) |
| `axis ask <prompt>` | Natural language to AgentTask (dry-run by default) |
| `axis ask <prompt> --submit` | Submit natural-language task to local runtime |
| `axis shell` | Start interactive in-process shell |
| `axis provider add/use/status/list/remove/archive` | Manage project-local LLM provider profiles |
| `axis context preview/inspect/preflight` | Context assembly preview and readiness check |
| `axis judge` | Run self-judgement diagnostic |
| `axis evolve inspect/promote/discard` | Sandboxed evolution inspection and decisions |

## External Tools

- **[axis-gui](tools/axis-gui/)**: Local Web GUI connecting to the Local Control Plane, providing Dashboard / Tasks / Providers / Events views (WebSocket real-time updates)
- **[axis-up](tools/axis-up/)**: Guided onboarding tool for environment detection / build / configuration / demo

Both tools do not import Axis internal packages; they communicate via CLI and HTTP API.

## Key Documentation

- [Agent-Native First Principles](docs/architecture/agent-native-first-principles.md) **<-- read before coding**
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
  kernel/          Scheduler, orchestrator, dispatcher
  contract/        Contract executor
  model/           LLM provider + tool system
  agent/           Agent executor + self-judgement engine
  intent/          Natural language intent parsing
  contextpack/     Adaptive context assembly
  control/         Local control plane (server/client/locator/events)
  evolution/       Sandboxed evolution protocol
  human/           Human executor
docs/              Documentation index, architecture reference, specs, status
tools/
  axis-gui/        Local Web GUI
  axis-up/         Guided onboarding tool
```

## Next Steps

- **Cross-process state persistence**: ReadinessRegistry integration with Local Control Plane
- **Agent identity and competence profiles**: Agent registry + behavioral scoring
- **Structured event log queries**: `axis audit` or equivalent capability
- **Dynamic model routing**: Cost/latency-aware provider selection + degradation chain
- **Execution feedback loop**: Result quality scoring fed back to intent/context assembly

