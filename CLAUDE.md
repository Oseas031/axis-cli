# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Identity

Axis is an **Agent-native scheduling system** — the execution substrate for Agent autogenesis. The goal is not a task queue or LLM framework; it's to let Agents understand tasks, organize actions, verify results, reflect on failures, and generate subsequent tasks, progressively earning autonomy through demonstrated competence.

**Core proposition**: More Context, More Action, Zero Control.

**Interaction principle**: Bash is All You Need — CLI-first, scriptable, composable, no Web UI or complex TUI.

## Current Status

- **Milestone 1**: Complete and accepted (2026-05-08). FIFO scheduling, dependency management, contract validation, state storage, orchestrator, CLI/shell.
- **Milestone 2**: In progress. Tasks T0-T2.5 complete; **T3 (contract admission layer) is the next pending task**. T4-T7 remain.

## Build, Test, Lint

```bash
# Build (Windows local dev — use axis-dev.exe to avoid overwriting existing binary)
go build -o axis-dev.exe cmd/axis/main.go

# Run all tests with race detection
go test -race ./...

# Run tests for a specific package
go test -race ./internal/kernel/scheduler/...

# Run a single test
go test -race ./internal/kernel/scheduler/... -run TestScheduler_GetReadyTasks

# Format, vet, staticcheck
gofmt -w .
go vet ./...
staticcheck ./...

# Cyclomatic complexity check (threshold 15)
gocyclo -over 15 .

# Security scan
gosec ./...
```

CI enforces: `gofmt`, `go vet`, `staticcheck`, `go test -race` with ≥60% coverage, `gocyclo`, `gosec`, and cross-platform build (linux/windows/darwin).

## Architecture

### Module: `github.com/axis-cli/axis` (Go 1.26)

Single external dependency: `github.com/spf13/cobra` (CLI only). All core modules use Go stdlib exclusively.

### Composition Root

`internal/kernel/orchestrator/orchestrator.go` is the top-level wiring point. `NewOrchestrator()` constructs and injects every module:

```
Orchestrator
  ├── StateStore        (MemoryStateStore — in-memory map + RWMutex)
  ├── LifecycleManager  (running flag, graceful shutdown via context cancel + done chan)
  ├── Scheduler         (FIFO queue, dependency resolution, cycle detection)
  │     depends on: StateStore, LifecycleManager (as LifecycleChecker)
  ├── Dispatcher        (routes tasks to executors, 30-min timeout)
  │     depends on: ContractExecutor, HumanExecutor
  ├── ContractExecutor  (schema validation against registered contracts)
  └── HumanExecutor     (human-in-the-loop call lifecycle — wired but not yet invoked)
```

### Core Data Types (`internal/types/types.go`)

- `AgentTask` — the unit of work: TaskID, ContractID, Input (map[string]any), Dependencies, Status, timestamps, Metadata
- `TaskStatus` — pending → running → completed | failed
- `AgentContract` — InputSchema + OutputSchema ([]FieldDef), each field has Name, Type (string/int/float/bool/array/object), Required, Enum
- `TaskState` — persisted snapshot: Task + Result + UpdatedAt

### Key Interfaces

| Interface | Package | Purpose |
|---|---|---|
| `StateStore` | `sharedlayer` | Save/Load/Delete task state |
| `LifecycleManager` | `lifecycle` | Shutdown / IsRunning |
| `Scheduler` | `scheduler` | Submit, Cancel, GetStatus, GetNextTask, GetReadyTasks, UpdateTaskStatus |
| `Dispatcher` | `dispatcher` | Dispatch(ctx, task) → TaskResult |
| `ContractExecutor` | `contract/executor` | Execute, ValidateInput, ValidateOutput, RegisterContract |
| `HumanExecutor` | `human/executor` | ExecuteCall, GetCallStatus, ResolveCall |

Every package follows the interface+impl pattern — interfaces are exported, implementations are unexported structs, making the system fully mockable.

### Runtime Flow

1. CLI creates/gets singleton Orchestrator via `sync.Once`
2. `Orchestrator.Start()` launches background `runTaskLoop` goroutine
3. `SubmitTask(task)` → Scheduler.Submit (cycle check, state store persist) → non-blocking notify on `taskSubmitted` channel
4. `runTaskLoop` polls `GetNextTask()` (delegates to `GetReadyTasks(1)` in M2), dispatches each ready task
5. `executeTask`: idempotency check → status to running → dispatcher.Dispatch → status to completed/failed
6. On shutdown: running flag cleared → task loop signalled → lifecycleManager.Shutdown (idempotent via sync.Once)

## CLI

Four cobra subcommands: `run <task-id>`, `status <task-id>`, `start`, `shell`.

`axis run` and `axis status` work standalone — they internally create a local orchestrator. No daemon or `axis start` prerequisite. Everything is in-process.

`axis shell` provides an interactive REPL (run, status, exit, help).

## Workflow System

**Every task starts at `workflow/entry.md`** — the single authoritative routing table. It maps work types to minimal workflow combinations:

- Feature/Bug Fix → wf-pr-check + wf-ci + wf-doc-006
- New Feature/Spec → wf-doc-004 + wf-occams + wf-pr-check + wf-ci + wf-doc-006
- Docs/Design/Workflow change → wf-doc-004 + wf-doc-006 + wf-occams

Workflows are registered in `.github/config/registry.yml` with status, dependencies, and file references. Active GH Actions workflows: ci, pr-check, cd, security, monitoring, dev, document-audit.

Meta-workflow governance: `workflow/meta-workflow-management.md` (wf-doc-004). Occam's razor guard: `workflow/occams-razor-architecture-simplification.md` (wf-occams).

## Critical Constraints

- **Zero external dependencies** in core modules (Go stdlib only; cobra is CLI-only)
- **Backward compatibility** for all existing M1 public APIs
- **Additive interfaces** — no rewrites of existing types
- **Occam's Razor**: build only what the current milestone needs. No Web UI, no TUI, no external DB, no daemons, no real LLM SDKs, no Prometheus/Grafana
- **CLI is a client**, not the core. The scheduler kernel is the core
- **Do not read deprecated files** in `docs/deprecated/`

## Test Conventions

- White-box testing (`package` matches source package)
- Fresh instances per test (no shared state, no TestMain, no fixtures)
- Real in-memory implementations (no mocks for internal modules)
- Naming: `Test<Type>_<Scenario>`
- Assertions: `t.Fatalf` for hard stops, `t.Errorf` for soft failures; no third-party assertion library
- Status polling in integration tests: ticker loop with deadline (pattern in orchestrator_test.go)

## Key Docs

- `workflow/entry.md` — start here for any task
- `docs/current-progress.md` — latest state
- `HANDOVER.md` — full project handoff context
- `docs/QUICKSTART.md` — build/run instructions and constraints
- `docs/specs/milestone2/` — M2 requirements, design, tasks, workflow binding
- `reports/daily/` — daily retrospectives
