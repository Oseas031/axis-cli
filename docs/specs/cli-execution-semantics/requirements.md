---
type: spec
status: active
created: 2026-05-14
last_verified: 2026-08-26
---

# CLI Execution Semantics Requirements

> 实现 agent-native-first-principles.md P1（CLI First）+ CLAUDE.md §1.2（No hidden daemons）

**Status**: In Progress
**Date**: 2026-05-14
**Decision**: Path B — in-process default, runtime as optional enhancement

## Problem Statement

Axis CLI has two execution worlds that are invisible to the user:

1. **In-process orchestrator** — used by `axis run` and `axis shell`
2. **HTTP runtime** — used by `axis start`, `axis status`, `axis ask --submit`

These worlds do not share state. A task submitted via `axis run` cannot be queried via `axis status`. The user receives no indication of which world they are operating in.

Additionally, `axis run` calls `initOrchestrator()` + `submitTask()` but never calls `orch.Start()`, so submitted tasks are never executed.

## Design Decision: Path B

**In-process is the default. Runtime is an optional enhancement for cross-process coordination.**

Rationale:
- "bash is all you need" — `axis run` should work like `go test`: run, report, exit
- No hidden daemons (§1.2) — requiring `axis start` for basic operations violates this
- Progressive complexity — single-command use cases should not pay the cost of multi-process architecture

## Functional Requirements

### FR1: `axis run` is synchronous one-shot execution

- Start in-process orchestrator
- Submit task
- Wait for task to reach terminal state (Completed / Failed)
- Print result
- Exit with appropriate code (0 = success, 1 = failure)
- Timeout: respect `sla.timeout` if set, default 60s

### FR2: `axis run` accepts user-provided input

- `axis run <task-id> --input '{"key": "value"}'` — JSON input
- `axis run <task-id> --prompt "do something"` — natural language (parsed via intent)
- Without either flag: error with usage hint (no hardcoded fake input)

### FR3: `axis status` requires runtime

- If no runtime is running, return clear error: "No local runtime. Start one with: axis start"
- Exit code 1 when runtime unreachable

### FR4: `axis start` is the only way to enter persistent mode

- Starts HTTP runtime with orchestrator
- Enables cross-process submit (`ask --submit`) and query (`status`)
- Writes `.axis/runtime.json` locator

### FR5: `axis shell` is self-contained in-process

- Starts its own orchestrator (already does this correctly)
- `run`/`status` within shell share the same orchestrator instance
- Does NOT connect to HTTP runtime
- Does NOT write `.axis/runtime.json`

### FR6: Unified project root resolution

- All commands that access `.axis/` use `project.ResolveRoot()` from cwd
- No hardcoded `"."` as root
- If `.axis/` not found, commands that require it error clearly

### FR7: Command mode documentation

- `--help` for each command states whether it requires runtime
- README CLI table has a "Requires Runtime" column

## Non-Functional Requirements

### NFR1: No behavior change for `axis shell`

Shell already works correctly (calls `orch.Start`). This spec must not break it.

### NFR2: No new dependencies

Fix uses existing orchestrator/scheduler/dispatcher infrastructure.

### NFR3: Cross-platform

`axis run` must work on Windows (no signals for timeout — use context cancellation).

## Acceptance Criteria

- `axis run <id> --prompt "hello"` executes task synchronously and prints result
- `axis run <id>` without input flags prints usage error
- `axis status <id>` without runtime prints "No local runtime" error
- `axis shell` continues to work as before
- `go test -race ./cmd/axis/...` passes
- No hardcoded `"."` remains in cmd/axis/ for project root
