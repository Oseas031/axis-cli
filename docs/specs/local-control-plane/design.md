# Local Control Plane Design

## Overview

The Local Control Plane turns Axis from a set of process-local CLI commands into a project-local runtime with a stable local interface.

Current failure mode:

```text
process A: axis ask --submit "..."  -> task exists in process A memory
process B: axis status <task-id>     -> process B has a new empty orchestrator
```

Target shape:

```text
axis start
  -> owns orchestrator runtime

axis ask --submit "..."
  -> sends AgentTask to local runtime

axis status <task-id>
  -> queries local runtime
```

The control plane is not a new execution engine. It is a local interface over the existing orchestrator, scheduler, dispatcher, contracts, providers, and tools.

## Design Principles

### Local runtime, not hidden controller

The runtime owns process lifecycle and IPC. It must not own permission policy, provider credentials, contract semantics, or evolution decisions.

### One task path

All task submissions must still become `types.AgentTask` and pass the same admission and scheduling path.

### Queryable state

Status and events should be inspectable by humans, shell scripts, `axis-up`, and future agents.

### Explicit lifecycle

P0 should require explicit `axis start`. Commands should not silently spawn hidden background services.

### Shell-native protocol

The first interface should be easy to test with local commands and stable JSON.

## Architecture

### Components

```text
LocalRuntime
LocalControlServer
LocalControlClient
RuntimeLocator
TaskEventLog
Orchestrator
Scheduler
Dispatcher
```

### P0 Data Flow

```text
axis start
  -> init provider profile
  -> init orchestrator
  -> start local control server
  -> start orchestrator loop
  -> write runtime locator

axis ask --submit <prompt>
  -> parse prompt into AgentTask
  -> send submit request to local control server
  -> server calls Orchestrator.SubmitTask
  -> scheduler persists process-local state and event log records submission

axis status <task-id>
  -> send status request to local control server
  -> server calls Orchestrator.GetTaskStatus
  -> response rendered by CLI
```

### Local Control Boundary

The control server may expose P0 endpoints similar to:

```text
POST /v1/tasks
GET  /v1/tasks/<task-id>/status
GET  /v1/health
```

This is an example shape. The implementation may use another local IPC mechanism if it preserves the same semantic boundary.

### Runtime Locator

A project-local locator should allow short-lived CLI clients to find the running runtime.

Candidate layout:

```text
.axis/runtime.json
```

Candidate fields:

```text
pid
protocol
address
started_at
project_root
```

The locator is discovery metadata, not authority.

### Task Event Log

A project-local event log should record important transitions.

Candidate layout:

```text
.axis/events/tasks.jsonl
```

Candidate event fields:

```text
event_id
task_id
event_type
created_at
actor
status
message
```

P0 may start with submission and terminal events only if full transition logging is too large for the first implementation.

Implemented P0 starts with local control boundary events:

```text
submitted
status_requested
```

The event log is append-only JSONL at `.axis/events/tasks.jsonl`. It records task IDs, event type, actor, status, message, timestamp, and event ID. It must not record provider credentials, authorization headers, API keys, prompt secrets, or full task input payloads.

## Detailed Design

### LocalRuntime

Owns:

- orchestrator initialization
- provider resolution
- lifecycle context
- graceful shutdown
- local control server startup

Must not own:

- provider credential storage
- contract validation policy
- scheduler semantics
- prompt construction

### LocalControlServer

Owns:

- decoding local requests
- calling orchestrator methods
- encoding stable responses
- returning actionable errors

Must not bypass:

- admission validation
- scheduler submission
- dispatcher execution
- contract executor

### LocalControlClient

Owns:

- loading runtime locator
- sending local requests
- translating connection errors into CLI guidance

### CLI Integration

Commands should behave as follows:

```text
axis start
  foreground local runtime

axis ask --submit <prompt>
  requires local runtime in P0
  submits parsed AgentTask to runtime

axis status <task-id>
  requires local runtime in P0
  queries runtime
```

Dry-run remains local and does not require a runtime:

```text
axis ask <prompt>
```

### Shell Integration

Two compatible options exist:

1. Shell as a client to the local runtime.
2. Shell as an explicit in-process runtime session.

P0 chooses shell as an explicit in-process runtime session.

This means:

```text
axis shell
  -> starts and owns an in-process orchestrator session
  -> shell run / ask --submit submit into that session
  -> shell status reads from that session
```

The standalone CLI path remains:

```text
axis start
  -> starts project-local runtime and writes locator

axis ask --submit <prompt>
axis status <task-id>
  -> use local control client and require axis start
```

The shell must not silently attach to or spawn the project-local runtime in P0. This keeps lifecycle ownership explicit and avoids hidden control behavior. Tests document this boundary by verifying shell `run`/`status` share in-process session state and do not require `axis start`.

## File Structure

Candidate implementation layout:

```text
internal/control/
  client.go
  server.go
  locator.go
  events.go
  types.go

cmd/axis/
  control_runtime.go
  main.go
  ask_cmd.go
  shell_cmd.go
```

## Trade-offs

| Option | Decision | Rationale |
|---|---|---|
| Keep process-local memory only | Rejected | Does not satisfy cross-process status/query requirements |
| Full remote API server | Rejected | Too broad and violates local-first P0 scope |
| Localhost HTTP | Candidate | Simple, testable, works on Windows, easy for CLI and axis-up |
| Named pipe | Candidate | Strong local semantics but more platform-specific |
| File queue only | Candidate | Shell-native but harder to make responsive and safe |
| Auto-spawn daemon | Deferred | Useful later, but P0 should keep lifecycle explicit |

## Risks

| Risk | Mitigation |
|---|---|
| Hidden central control plane | Keep daemon local, explicit, and bounded to orchestrator API |
| Port conflicts | Use locator plus dynamic loopback port or documented configured port |
| Stale runtime locator | Validate health before use and show cleanup guidance |
| Credential leakage | Never write API keys to locator or event logs |
| Divergent shell behavior | Reuse control client in shell where possible |
| Execution semantics drift | Keep all submission paths routed through `Orchestrator.SubmitTask` |

## Acceptance Mapping

- FR1: LocalRuntime and `axis start`
- FR2: submit endpoint/client used by `axis ask --submit`
- FR3: status endpoint/client used by `axis status`
- FR4: TaskEventLog
- FR5: LocalControlServer and LocalControlClient
- FR6: RuntimeLocator health checks and CLI guidance
- FR7: shell client integration or explicit in-process documentation

## P0 Success Definition

The user's reported workflow should work:

```bash
axis start
axis ask --submit "Write a 100-word Axis introduction to the desktop"
axis status <task-id>
```

The status command should find the task through the running local runtime, and the runtime should execute submitted tasks through the existing Axis task path.
