# Local Control Plane Requirements

> 实现 agent-native-first-principles.md P1（Interface is Existence）


## Summary

Add a local Axis control plane so separate CLI processes, `axis-up`, the interactive shell, and future agents can submit tasks, query status, and observe execution through one shared local runtime.

The feature fixes the current process-bound task state problem:

```text
axis ask --submit <prompt>
axis status <task-id>
```

should not fail only because the two commands run in different OS processes.

The control plane is a local execution substrate, not a hidden central controller.

## Philosophy Alignment

- **Interface is existence**: CLI, shell, `axis-up`, and future agents should call the same local task interface.
- **Query is context**: task status, event history, and execution state should be queryable instead of trapped in process memory.
- **Layered isolation is collaboration**: coordination is shared, while task execution and future workspaces remain isolated.
- **Contract is structure**: all submissions must still materialize as `types.AgentTask` and pass contract/admission paths.
- **More Context, More Action, Zero Control**: users gain more observable action without hidden permission or execution semantics.
- **Bash is all you need**: the first interface must remain local, CLI-native, scriptable, and machine-readable.

## Users

- Developers using `axis ask --submit` followed by `axis status` in separate commands.
- Operators running a long-lived local Axis runtime.
- `axis-up` and other external tools that need a stable local control interface.
- Future agents that need task submission and status inspection without embedding Axis internals.

## Functional Requirements

### FR1: Long-lived local runtime

Axis must provide a local runtime process that owns the orchestrator execution loop for the current project.

Example shape:

```bash
axis start
```

The runtime must be project-local and must not require a remote service.

### FR2: Cross-process task submission

A separate CLI process must be able to submit an ordinary `types.AgentTask` to the running local runtime.

Example:

```bash
axis ask "write 100 words about Axis to desktop" --submit
```

The submitted task must pass existing parsing, admission, scheduler, dispatcher, and contract paths.

### FR3: Cross-process status query

A separate CLI process must be able to query task status from the running local runtime.

Example:

```bash
axis status <task-id>
```

It must not depend on the submitting process still being alive.

### FR4: Observable local events

The runtime should record enough local evidence to explain task lifecycle transitions:

- submitted
- queued
- running
- completed
- failed
- cancelled when supported

P0 may use JSON files or JSON Lines under `.axis/`.

### FR5: Local IPC boundary

The control plane must expose a local-only interface suitable for CLI clients.

P0 may choose one implementation:

- localhost HTTP bound to loopback
- named pipe
- project-local file queue with lock discipline

The selected mechanism must be documented and tested.

### FR6: Safe fallback behavior

If no local runtime is running, commands that require it must fail with actionable guidance.

Example:

```text
No local Axis runtime found. Start one with: axis start
```

P0 must not silently create hidden background daemons unless explicitly specified.

### FR7: Shell compatibility

`axis shell` should either:

- use the same local runtime protocol, or
- clearly document that it owns an in-process runtime session.

It must not diverge semantically from CLI task submission.

## Acceptance Criteria

- [x] `axis start` starts a project-local runtime with an orchestrator execution loop.
- [x] `axis ask --submit ...` can submit to that runtime from a separate process.
- [x] `axis status <task-id>` can query status from a separate process.
- [x] Submitted tasks are executed by the runtime when dependencies and contracts allow.
- [x] The current task-not-found failure caused only by process isolation is eliminated.
- [x] Runtime communication is local-only.
- [x] Provider credentials remain under provider configuration boundaries.
- [x] Contract, scheduler, dispatcher, provider, and context semantics are not bypassed.
- [x] Tests cover no-runtime guidance, submit-to-runtime, status-from-runtime, and shell compatibility.

## Constraints

- P0 must be local-first.
- P0 must avoid remote network dependencies.
- P0 must not introduce a Web UI or TUI.
- P0 must not require a database server.
- P0 must not store API keys in task events.
- P0 must not execute natural language directly without task materialization.
- P0 must preserve dry-run by default for `axis ask` unless `--submit` is explicit.

## Non-Goals

- No distributed multi-node scheduler.
- No cloud-hosted control plane.
- No browser UI.
- No autonomous permission escalation.
- No provider credential management changes.
- No contract semantic changes.
- No prompt injection or provider request mutation.
- No sandboxed evolution promotion logic.

## P0 Decisions

1. P0 uses localhost HTTP bound to loopback with a project-local `.axis/runtime.json` locator.
2. P0 `axis start` blocks in the foreground; explicit daemonize mode is deferred.
3. P0 CLI commands detect a running runtime through `.axis/runtime.json`.
4. P0 task event logs are append-only JSONL under `.axis/events/tasks.jsonl`.
