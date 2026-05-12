# Kernel Abstraction Model

## Positioning

Axis is an operating system for Agents.

It solves the problem: **how to let a non-deterministic, autonomous entity exist safely, observably, and evolvably in a complex world.**

Axis does not control Agents. It provides the substrate conditions for their existence — scheduling, context, capabilities, isolation, and observability — so they can act, learn, and evolve within verifiable boundaries.

## Architecture

```
┌─────────────────────────────────────────────────┐
│  System Call Layer (fixed primitives)            │
│  The only interface between Agent and Kernel     │
├─────────────────────────────────────────────────┤
│  Core Abstraction Layer                         │
│  ┌──────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐
│  │ Schedule │ │ Context  │ │ Capability│ │ Isolation│
│  │ Lifecycle│ │ Budget   │ │ Registry │ │ Boundary │
│  │ Dependency│ │ Introspec│ │ Granting │ │ Sandbox  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘
├─────────────────────────────────────────────────┤
│  Infrastructure Layer                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │ WAL      │ │ Snapshot │ │ Index    │        │
│  │ EventStream│ │ State  │ │ Retrieval│        │
│  └──────────┘ └──────────┘ └──────────┘        │
└─────────────────────────────────────────────────┘
```

## System Call Layer

Kernel syscalls are the fixed primitives that define what an Agent can say to the Kernel. Their semantics are defined by the Kernel and cannot be rewritten by Agents.

| Syscall | Semantics | Current Implementation |
|---------|-----------|----------------------|
| `submit_task` | Submit a task for scheduling | `orch.SubmitTask()` |
| `query_state` | Query task/system state | `status` command, StateStore |
| `acquire_context` | Request context loading | contextpack Assembler |
| `request_capability` | Request access to a tool/skill | load_skill, permission ladder |
| `compact` | Request context compaction | CompactionPipeline |
| `spawn` | Create subtask/sub-agent with isolation | `spawn` tool (full/shared isolation) |
| `introspect` | Get self-state snapshot | SelfContext, ContextBuilder |
| `yield` | Voluntarily yield execution | `yield` tool |
| `checkpoint` | Persist intermediate state | `checkpoint` tool |

### Syscall vs Userland

**Syscall** = what Agent can say to Kernel (fixed, semantic, cannot be overridden).

**Userland tool** = what Agent can do to the world (pluggable, registered, can be added/removed).

Examples of userland tools: `bash`, `file_read`, `file_write`, `http_request`, `load_skill`.

An Agent obtains userland tools through `request_capability`. The Kernel decides whether to grant based on autonomy level and permission boundaries.

## Core Abstraction Layer

### Scheduling

Manages task lifecycle and execution ordering.

- **State machine**: pending → ready → running → completed/failed
- **DAG dependency resolution**: tasks declare dependencies, scheduler resolves readiness
- **Priority + SLA**: failure class routing, backoff strategies, timeout enforcement
- **Preemption**: the Kernel's only "control" — can suspend low-priority tasks

Maps to: `internal/kernel/scheduler/`, `internal/kernel/orchestrator/`, `internal/kernel/lifecycle/`

### Context

Manages what information is available to an Agent during execution.

- **Budget**: each task has a token budget; compaction runs automatically within budget
- **Introspection**: Agent can query its own task lineage, competence score, autonomy level
- **Assembly**: retrieves relevant context packets based on task goal (TF-IDF, rules)
- **Lifecycle**: context is created with task, destroyed with task, does not leak across tasks

Maps to: `internal/contextpack/`, `internal/agent/context.go`, `internal/contract/executor/history_compact.go`

### Capability

Manages what an Agent is allowed to do.

- **Registry**: declarative directory of all available tools and skills
- **Granting**: Agent's current capability set is determined by autonomy level
- **Discovery**: Agent queries available capabilities via syscall, not pre-injected
- **Upgrade path**: stable performance → competence increase → unlock more capabilities

Maps to: `internal/model/tool/`, `internal/skills/`, `internal/agent/` (autonomy)

### Isolation

Ensures one Agent's failure or misbehavior does not affect others.

- **Task isolation**: each task has independent history, context, tool scope
- **Agent isolation**: sub-agents execute in sandbox, return summary to parent
- **Evolution isolation**: system modification proposals verified in branch before promotion
- **Failure isolation**: one task's panic does not affect other tasks' scheduling

Maps to: `internal/evolution/`, Sandboxed Evolution Protocol, dispatcher routing

## Infrastructure Layer

### WAL + Event Stream

- All state changes write WAL first, then update memory
- `tasks.jsonl` is a materialized view of the event stream
- Supports replay: rebuild any point-in-time system state from WAL

Maps to: `internal/control/events.go`, `.axis/events/`

### Snapshot + State

- StateStore: persistent task execution state
- Checkpoint: intermediate state save for long tasks (Agent-triggered via syscall)
- Used for crash recovery and task migration

Maps to: `internal/kernel/sharedlayer/`, `.axis/`

### Index + Retrieval

- TF-IDF / semantic index: retrieval backend for context assembly
- Skills directory index
- Code/doc index supporting `acquire_context`

Maps to: `internal/contextpack/index*.go`, `internal/skills/discover.go`

## Design Principles Applied

| Principle | Kernel Expression |
|-----------|------------------|
| More Context | Kernel provides query infrastructure (index, retrieval, skills); does not push |
| More Action | Syscall table + capability registry define action space; Agent decides what to use |
| Zero Control | Kernel schedules and isolates but never dictates Agent's next action |
| Controllable Evolution | All mutations go through sandbox → verify → promote pipeline |

## What This Model Is Not

- Not a 1:1 Linux clone. Linux manages deterministic processes; Axis manages non-deterministic agents.
- Not a framework. Frameworks call your code; an OS provides substrate for your existence.
- Not a control plane. Control planes enforce policy; Axis provides conditions for autonomous action.

## Evolution Path

| Phase | Focus |
|-------|-------|
| Current | Syscall semantics stabilization, core abstractions implemented |
| Next | Unified Actor model + Communication Layer (`docs/specs/actor-comm/`) |
| Next+1 | `yield` + `checkpoint` primitives, spawn with full isolation |
| Future | Multi-actor scheduling, capability marketplace, network communication |

## Relationship to Other Documents

- `agent-native-design-philosophy.md` — **why** (principles)
- `kernel-abstraction-model.md` — **what** (this document: structure)
- `semantic-boundaries.md` — **must not** (constraints)
- `module-and-naming-conventions.md` — **how** (code organization)
