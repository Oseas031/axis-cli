# Axis Project Handover Document

> **Reference document.** For coding constraints see `CLAUDE.md`. For handover entry see `AGENT_INSTRUCTIONS.md`.

## Project Overview

**Project Name**: Axis (Agent-Native Scheduling System)
**Positioning**: Agent-native scheduling system; the early execution substrate for Agent autogenesis
**Core Capabilities**: Task scheduling, dependency management, contract admission, context supply, execution orchestration, verification and reflection infrastructure
**Tech Stack**: Go 1.26+
**Current Status**: M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Sandboxed Evolution ✅ | Local Control Plane ✅

**Important Notes**:
- **End Goal**: Agent-native scheduling system, progressively moving toward Agent autogenesis
- **CLI Positioning**: CLI is just one client of the scheduling system, not the core
- **Design Principle**: Occam's Razor — minimum viable, implement only the minimal feature set needed to validate core concepts

## Core Design Philosophy: More Context, More Action, Zero Control, Controllable Evolution

- **More Context**: Provide efficient query infrastructure so Agents actively build context per task needs
- **More Action**: Give Agents execution, composition, validation, correction, and follow-up task generation capabilities
- **Zero Control**: The system does not prescribe a single action path; it provides boundaries via contracts, permission tiers, isolation, and auditing
- **Controllable Evolution**: Agent self-bootstrapping, self-modification, and permission elevation must be observable, verifiable, and rollback-safe

This design philosophy stands in contrast to traditional scheduling systems' "Less Context, Less Action, More Control", reflecting the essential nature of Agent-native scheduling.

See [Agent-Native First Principles](docs/architecture/agent-native-first-principles.md)

## Autogenesis Positioning

Axis's direction is not automation, but autogenesis. The self-bootstrap origin has already occurred: external Agents are injecting ideas of observable solidification, execution, reflection, and evolution into Axis.

Four transitional structures must be understood as follows:

- **workflow** is temporary scaffolding
- **contract** is a growth boundary
- **permission rule** is a progressive autonomy mechanism
- **spec** is a seed

These structures are not permanent iron cages controlling Agents, but conditions that help Agents accumulate competence, earn autonomy, and eventually internalize, rewrite, and discard external structures into their own action structures.

M2 is not an ordinary parallel scheduling milestone, but the execution substrate for the future **Autogenesis Loop**.

M3 Phase 1 opened the execution path: Dispatcher no longer returns hardcoded stub results; tasks truly flow through ValidateInput → ModelProvider.Execute → ValidateOutput → TaskResult.

### Test Coverage

Coverage **87.1%** (2026-05-09), exceeding the 85% target:

| Module | Coverage |
|---|---|
| cmd/axis | 67.7% |
| contract/admission | 100.0% |
| contract/executor | 91.7% |
| human/executor | 100.0% |
| kernel/dispatcher | 88.4% |
| kernel/lifecycle | 100.0% |
| kernel/orchestrator | 84.4% |
| kernel/scheduler | 88.1% |
| kernel/sharedlayer | 100.0% |
| model/provider | 100.0% |
| model/tool | 97.0% |
| types | 100.0% |

## Interaction Design: bash is all you need

Axis's default interaction follows **"bash is all you need, simple but robust, composable and extensible"**:

- **CLI First**: Prioritize providing standard CLI commands
- **Shell Native**: Prioritize supporting bash / PowerShell / CI / Agent tool invocations
- **Composable**: Commands should be composable and scriptable
- **Minimal UI**: No default heavy Web UI or complex TUI
- **Robust**: Stay minimal but provide necessary fault tolerance, confirmation, rollback, and observability
- **Extensible**: Reserve extension points in interfaces so Axis can be called, orchestrated, and adapted by Agents

This is the concrete realization of **More Context, More Action, Zero Control, Controllable Evolution** at the interaction layer.

See [Bash is All You Need](docs/architecture/bash-is-all-you-need.md)

## Completed Work

### M1 Core Features (Completed)
- ✅ Core data structures (AgentTask, TaskStatus, TaskResult, FieldDef, etc.)
- ✅ State storage module (internal/kernel/sharedlayer/state_store)
- ✅ Lifecycle manager (internal/kernel/lifecycle)
- ✅ Scheduler (FIFO + dependency management) (internal/kernel/scheduler)
- ✅ Contract executor (input/output validation) (internal/contract/executor)
- ✅ Human executor (internal/human/executor)
- ✅ Dispatcher (internal/kernel/dispatcher)
- ✅ Orchestrator (internal/kernel/orchestrator)
- ✅ CLI client (cmd/axis)
- ✅ Unit tests (coverage ≥ 80%)

### M2 Parallel Scheduling (Completed)
- ✅ DAG parallel scheduling
- ✅ Contract admission rules
- ✅ SLA agreements (timeout/retries/failure_class)
- ✅ Structured error codes (9 codes)
- ✅ Parallel execution loop (5 workers)
- ✅ Retry and exhaustion wrapping

### M3 Execution Ecosystem (Completed)
- ✅ Phase 1: ModelProvider interface + MockModelProvider, execution path through Dispatcher → ContractExecutor → ModelProvider
- ✅ Phase 2: ModelProvider configurable (Functional Options), EchoModelProvider, HumanExecutor routing, DAG enhancements
- ✅ Phase 3: SLA strategy engine (failure_class routing, backoff strategies, priority scheduling), Tool layer (Tool interface + ToolRegistry + BashTool)

### M4 Real LLM Integration (Completed)
- ✅ Anthropic/OpenAI provider implementations
- ✅ Extended tool set (file read/write, HTTP client)
- ✅ Tool permission scopes
- ✅ Circuit breaker
- ✅ CLI `--provider` flag
- ✅ Shell `tools` command

### M5 Bootstrap Loop (Completed)
- ✅ AgentExecutor interface + MockAgentExecutor
- ✅ AgentRuntimeAdapter (external Agent CLI support)
- ✅ SelfContext data structure
- ✅ ContextBuilder implementation
- ✅ ContextCompressor (3 compression strategies)
- ✅ Self-iteration Contracts (analyze/implement/validate/update/review/spawn)
- ✅ BootstrapOrchestrator (self-loop task scheduling)
- ✅ FollowUpTaskGenerator
- ✅ AutonomyTransition data model (5-level autonomy)
- ✅ RuleEngine (competence evidence-based rule engine)

### Layered Memory Model — P0 (Completed)
- ✅ Spec-RDT: `docs/specs/layered-memory-model/` (requirements + design + tasks)
- ✅ Three explicit layers: Immediate / Working / Long-term
- ✅ `internal/memory/kv` — Indexed KV engine (history.jsonl + snapshot.bin + index.txt)
  - Pure Go stdlib, zero external deps, LF-only line terminators
  - Log is immutable source of truth; Compact only rebuilds snapshot+index
  - `compacted_history_offset` enables incremental replay on restart
  - Cross-platform atomic rename (Unix direct, Windows two-phase `.old` fallback)
  - Snapshot corruption → automatic full history replay
  - Index loss with valid snapshot → rebuild by scanning snapshot JSONL
  - Defensive bounds: maxKeyLen=256, maxValueLen=256 KiB
- ✅ `internal/memory/working` — Working Memory (key `wm:bundle:{id}`)
  - Retain / Release / List / Recall / Clear / Compact / GetBundle / UpdateBundle
  - P0 Recall = case-insensitive keyword match over goal + packet summaries
- ✅ `internal/memory/immediate` — ImmediateContext + ContextBuilder
  - UTF-8-safe 1024-byte head truncation
  - SHA-256 truncated to 128-bit (32 hex chars), zero deps, BLAKE3-128 interface parity
  - `.seen` single-line-per-entry file for cross-session file_changed tracking
  - Token estimation: rune-weighted (ASCII×0.25, CJK×1.0, other×0.5)
  - Budget degrades summary → path-only mode on exhaustion
- ✅ `internal/memory/longterm` — Event Store (append-only JSONL)
  - EventFilter: event_types, entity_id, time range, limit
  - Forgetting is soft-mark only; raw events never physically deleted
- ✅ CLI: `axis memory retain|release|list|inspect|compact|recall` (+ `--json`)
- ✅ Tests: 40 test cases, all pass (corruption recovery, boundary, concurrency-safe lock)

### M6 Self-Judgement (Completed)
- ✅ JudgementCriteria / JudgementResult / JudgementItem data structures
- ✅ SelfJudgementEngine (weighted scoring aggregation)
- ✅ 5 ValidationStrategies: Syntax, Semantic, Contract, Test, Coverage
- ✅ self/judge-execution contract
- ✅ JudgementResult integration into AgentExecutionResult
- ✅ BootstrapOrchestrator judgement support
- ✅ CLI judgement result display
- ✅ Unit and integration tests

### CI/CD Pipelines (Completed)
- ✅ CI Workflow (format, vet, staticcheck, test, build)
- ✅ CD Workflow (cross-platform build, Docker image, Release, signing)
- ✅ Security Workflow (SAST, SCA, Secret Scan, License Compliance)
- ✅ PR Quality Check Workflow (quality gates, code review)
- ✅ Monitoring Workflow (performance baseline, coverage trends, CI metrics)
- ✅ Dev Workflow (local development automation)
- ✅ Document Audit Workflow (format, links, content consistency, milestone alignment)
- ✅ Registry Validator Workflow (registry validation, index generation)

### Bug Fixes (Completed)
- ✅ Scheduler GetNextTask marking tasks as scheduled
- ✅ Dispatcher goroutine leak fix
- ✅ Orchestrator busy-wait pattern fix
- ✅ Contract executor thread safety fix
- ✅ CLI nil pointer risk fix
- ✅ Orchestrator Start logic error fix (state check and set logic reversed)
- ✅ Lifecycle check mutex protection
- ✅ Contract executor enum validation fix
- ✅ Dispatcher context shadowing fix
- ✅ Orchestrator Shutdown task cleanup (added task loop notification)
- ✅ Various workflow YAML fixes (registry paths, variable scoping, dead conditions)
- ✅ scheduler.go cyclic dependency detection algorithm fix
- ✅ state_store.go Load method zero value return fix
- ✅ lifecycle.go done channel double-close fix (sync.Once)
- ✅ dispatcher.go goroutine leak risk fix (timeoutCtx.Done() check)
- ✅ executor.go int type conversion precision loss fix
- ✅ orchestrator.go executeTask idempotency protection fix
- ✅ main.go global variable concurrency safety fix (sync.Once)
- ✅ axis shell interactive Shell implementation
- ✅ axis shell default contract registration fix
- ✅ M6 staticcheck warnings fix

### GitHub Infrastructure (Completed)
- ✅ GitHub CLI (gh v2.92.0) installed and authenticated as Oseas031
- ✅ Pre-commit hook fix: Windows Python compatibility, registry path updates, Unicode safe output
- ✅ CLAUDE.md creation: project architecture, commands, workflow routing, test coverage

## Key Technical Decisions

### Tech Stack Choice
- **Language**: Go 1.26+
- **Rationale**: Goroutine + Channel concurrency model, single static binary, zero external dependencies (Go standard library only for core modules)

### Core Constraints
- **Zero external dependencies**: Core modules depend only on Go standard library
- **Backward compatible**: Core interface changes must maintain backward compatibility
- **Occam's Razor**: Minimum viable, progressive enhancement

## Project Structure

```text
axis-cli/
├── cmd/
│   └── axis/              # Main CLI commands
├── internal/
│   ├── kernel/           # Scheduling kernel
│   │   ├── scheduler/    # Scheduler
│   │   ├── dispatcher/   # Dispatcher
│   │   ├── lifecycle/    # Lifecycle manager
│   │   ├── orchestrator/ # Orchestrator
│   │   └── sharedlayer/  # Shared state storage
│   ├── contract/         # Contract layer
│   │   ├── admission/    # Contract admission validation
│   │   └── executor/     # Contract executor
│   ├── human/            # Human-as-a-Function
│   │   └── executor/     # Human executor
│   ├── model/            # Model layer
│   │   ├── provider/     # ModelProvider interface + implementations
│   │   └── tool/         # Tool interface + ToolRegistry + BashTool
│   ├── agent/            # Agent layer
│   │   ├── contracts/    # Self-iteration contracts
│   │   └── judgement/    # Self-Judgement engine
│   ├── intent/           # Natural language intent parsing
│   ├── contextpack/      # Adaptive context assembly
│   ├── control/          # Local control plane
│   ├── evolution/        # Sandboxed evolution protocol
│   └── types/            # Core data types + error codes
├── scripts/              # Tool scripts
├── docs/                 # Documentation
│   └── specs/            # Milestone spec documents
├── configs/              # Configuration files
├── .github/              # GitHub configuration
│   └── workflows/        # GitHub Actions workflows
├── go.mod                # Go module definition
└── README.md             # Project README
```

## Known Issues

- ⚠️ `EnterWorktree` (and Agent `isolation: "worktree"`) creates worktree from default branch `main` HEAD, not current branch HEAD. For parallel development use manual worktree: `git worktree add -b <name> .claude/worktrees/<name> <commit>` + `EnterWorktree --path`
- ⚠️ sign-artifacts job unused — pending post-milestone handling

## Document Index

### Quick Navigation
- Coding constraints (constitution): `CLAUDE.md`
- Agent handover entry: `AGENT_INSTRUCTIONS.md`
- Session state (live, per-step): `docs/status/session-state.md`
- Current progress (milestone status): `docs/status/current-progress.md`
- Design principles: `docs/architecture/agent-native-first-principles.md`
- Workflow entry: `workflow/entry.md`
- Workflow index: `workflows/README.md`
- Roadmap: `docs/product/ROADMAP.md`
- Quick start: `docs/guides/QUICKSTART.md`
- Reports index: `reports/`

### Configuration Files
- Production config: `configs/config.yaml`
- Development config: `configs/config.dev.yaml`

### Deprecated Documents Warning
Do not read the following files (deprecated):
- `docs/deprecated/` — All files in this directory are archived and superseded

---

**Handover Date**: 2026-05-11
**Handover Status**: M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Sandboxed Evolution ✅ | Local Control Plane ✅
**M1 Acceptance**: ✅ Passed (2026-05-08)
**Next Steps**: See `docs/status/session-state.md` for the live state (last atomic step, next concrete action, blocking decisions). High-level direction: cross-process state persistence; Agent identity profiles; structured event log queries; dynamic model routing; execution feedback loop.
