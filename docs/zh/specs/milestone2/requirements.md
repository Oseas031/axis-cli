# Milestone 2 Requirements

**Status**: Completed

## Summary

Milestone 2 extends Axis from Milestone 1's serial FIFO scheduling into a minimal parallel scheduling system for dependency graphs. The goal is to improve execution throughput while preserving Axis's Agent-native design philosophy and the existing Milestone 1 task/contract/state model.

Milestone 2 must remain spec-first. This document defines what to build before implementation begins.

This feature is bound to the existing Axis workflow system through [workflow-binding.md](workflow-binding.md). Implementation must follow the workflow route declared there.

## Users

- Agents that submit multiple dependent tasks and expect ready tasks to run concurrently
- Developers validating Axis locally through CLI and shell-native workflows
- Future API/SDK clients that need predictable task scheduling semantics

## Design Philosophy

Milestone 2 must follow:

- **More Context**: expose enough dependency, readiness, and failure context for agents to decide next actions
- **More Action**: allow multiple independent tasks to proceed without artificial serial bottlenecks
- **Zero Control**: avoid central policy-heavy control planes; keep scheduling semantics explicit and replaceable
- **Controllable Evolution**: keep scheduler changes observable, testable, and safe to roll back
- **Bash is All You Need, simple but robust, composable and extensible**: keep validation and usage accessible from ordinary CLI commands and Go tests

## Functional Requirements

### FR1: DAG dependency model

Axis must treat task dependencies as a directed acyclic graph where each task can depend on zero or more prior tasks.

The existing `AgentTask.Dependencies []string` field remains the source of dependency edges.

### FR2: Parallel-ready scheduling

The scheduler must be able to return all currently ready pending tasks, not only one task at a time.

A task is ready when:

- its status is `pending`
- all declared dependencies exist
- all declared dependencies are `completed`
- selecting it does not violate scheduler lifecycle state

### FR3: Backward compatibility

Existing Milestone 1 behavior must continue to work:

- `Submit`
- `Cancel`
- `GetStatus`
- `GetNextTask`
- `UpdateTaskStatus`
- existing CLI commands
- existing orchestrator single-task execution path until replaced by parallel loop

### FR4: Contract admission rules

Axis must add a minimal admission phase before task submission is accepted for execution.

Admission must validate:

- referenced contract exists
- task input satisfies contract input schema
- task dependencies do not create cycles
- task dependency IDs are either already known or explicitly allowed as future submissions by design decision

### FR5: SLA metadata

Axis must support minimal SLA metadata on tasks without implementing a full policy engine.

The first SLA scope is:

- task timeout
- retry count
- failure classification placeholder

SLA must be represented in a way that does not require external dependencies or persistent databases.

### FR6: Error code foundation

Axis must define a small internal error code vocabulary for scheduler, contract admission, and execution failures.

Error codes must be stable enough for CLI and future API/SDK clients to inspect, but must not introduce a large exception framework.

### FR7: Observability for parallel execution

Parallel execution must make task state transitions inspectable via existing state store mechanisms.

At minimum, tests must be able to verify:

- ready tasks become `running`
- completed tasks unblock dependents
- failed dependencies prevent dependent execution
- timeout/retry behavior is visible in task result or status context

## Acceptance Criteria

- [x] `workflow-binding.md` declares upstream workflows and completion criteria
- [x] User confirms `requirements.md`, `design.md`, `tasks.md`, and `workflow-binding.md` before implementation
- [x] Existing `go test ./...` passes before implementation starts
- [x] Scheduler has tests for returning multiple ready tasks
- [x] Scheduler has tests for DAG dependency unblocking
- [x] Orchestrator has tests proving independent tasks can execute concurrently
- [x] Contract admission rejects unknown or invalid contracts before scheduling
- [x] SLA timeout behavior is covered by tests
- [x] Error codes are documented and used in at least scheduler/admission paths
- [x] Existing Milestone 1 CLI behavior still works
- [x] No Web UI, TUI, external database, or external model SDK is introduced

## Constraints

- Keep core modules on Go standard library only
- Preserve existing public semantics where possible
- Prefer additive interfaces over invasive rewrites
- Do not make CLI the core architecture
- Do not implement Milestone 3 features such as tool calling, contract library, full SDK, or global event bus
- Keep implementation minimal and test-driven

## Non-Goals

- Distributed scheduling
- Persistent database-backed state
- Cross-process worker pools
- Full policy engine
- Real LLM provider integration
- Tool calling layer
- Browser dashboard
- Multi-tenant authorization

## Open Questions

- Should dependencies on not-yet-submitted task IDs be rejected immediately or allowed for future submission?
- Should retries re-enter the scheduler queue or stay inside task execution?
- Should timeout be stored in `AgentTask.Metadata` initially or promoted to a typed field?
- Should `GetReadyTasks(limit int)` be added to `Scheduler`, or should `GetNextTask` be repeatedly called by a worker pool?
