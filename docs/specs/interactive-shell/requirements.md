# Interactive Shell Requirements

> 实现 bash-is-all-you-need.md（CLI First）


## Summary

Add a lightweight interactive CLI shell to Axis, similar in interaction style to Claude Code but scoped to Milestone 1 capabilities. The shell is a client layer over the existing orchestrator and must not change the core scheduler architecture.

The shell also implements Axis's interaction philosophy: **bash is all you need, simple but robust, composable and extensible**. It should be shell-native, composable, script-friendly, minimal, fault-tolerant, and extensible.

## Users

- Developers testing Axis locally
- Agents or operators who need a conversational command loop for task submission and status inspection

## Functional Requirements

### FR1: Shell command

Axis must provide a new CLI command:

```bash
axis shell
```

The command starts an interactive prompt and initializes the orchestrator for the session.

### FR2: Prompt loop

The shell must display a prompt and accept commands repeatedly until the user exits.

Example:

```text
axis> help
axis> run task-1
axis> status task-1
axis> exit
```

### FR3: Built-in commands

The shell must support these commands:

- `help` — show available commands
- `run <task-id>` — submit a task using the existing default contract placeholder
- `status <task-id>` — show task status
- `exit` / `quit` — gracefully shut down the orchestrator and exit

### FR4: Context-first output

Shell responses should provide enough context for the user to understand what happened:

- submitted task ID
- task status
- error reason when a command fails
- suggested next command when useful

### FR5: Non-blocking interaction

The prompt must remain usable after command errors. Invalid commands must not crash the shell.

### FR6: Graceful shutdown

On `exit`, `quit`, Ctrl+C, or EOF, the shell must call orchestrator shutdown and return to the terminal cleanly.

## Acceptance Criteria

- [x] `go build -o axis.exe cmd/axis/main.go` succeeds
- [x] `axis shell` starts an interactive prompt
- [x] `help` prints supported commands
- [x] `run demo-task` submits a task or returns a clear contextual error
- [x] `status demo-task` returns status or a clear not-found error
- [x] invalid commands produce a non-fatal error message
- [x] `exit` and `quit` shut down gracefully
- [x] Ctrl+C exits cleanly

## Constraints

- Do not add Web UI
- Do not add a full TUI layout
- Do not change scheduler/orchestrator core architecture except for bug fixes required for shell operation
- Prefer Go standard library
- Keep the feature minimal and aligned with More Context, More Action, Zero Control, Controllable Evolution
- Keep the feature aligned with bash is all you need, simple but robust, composable and extensible: CLI first, shell native, composable, no heavy UI by default, with clear errors and extension points

## Non-Goals

- Natural language understanding
- LLM integration
- Persistent task database
- Multi-session shared orchestrator
- Browser dashboard
- Advanced command history/autocomplete

## Open Questions

- ~~Should `run <task-id>` also accept JSON input later?~~ (Resolved: axis ask handles rich input via --stdin)
- ~~Should shell expose task list once scheduler supports listing?~~ (Resolved: axis status handles listing)

## Planned Extension: Natural Language Scheduling

Natural language understanding remains out of scope for the original interactive shell milestone. A future extension is tracked separately in `docs/specs/natural-language-scheduling/`.

That extension should treat natural language as an intent-to-task compiler:

- `axis ask "<prompt>"` creates or previews a structured `AgentTask`
- future `axis shell` may support `ask <prompt>`
- parsed tasks must preserve provenance metadata
- scheduler, dispatcher, and contract execution semantics must remain unchanged for the first implementation
