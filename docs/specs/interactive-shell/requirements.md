# Interactive Shell Requirements

## Summary

Add a lightweight interactive CLI shell to Axis, similar in interaction style to Claude Code but scoped to Milestone 1 capabilities. The shell is a client layer over the existing orchestrator and must not change the core scheduler architecture.

The shell also implements Axis's interaction philosophy: **bash is all you need**. It should be shell-native, composable, script-friendly, and minimal.

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

- [ ] `go build -o axis.exe cmd/axis/main.go` succeeds
- [ ] `axis shell` starts an interactive prompt
- [ ] `help` prints supported commands
- [ ] `run demo-task` submits a task or returns a clear contextual error
- [ ] `status demo-task` returns status or a clear not-found error
- [ ] invalid commands produce a non-fatal error message
- [ ] `exit` and `quit` shut down gracefully
- [ ] Ctrl+C exits cleanly

## Constraints

- Do not add Web UI
- Do not add a full TUI layout
- Do not change scheduler/orchestrator core architecture except for bug fixes required for shell operation
- Prefer Go standard library
- Keep the feature minimal and aligned with More Context, More Action, Zero Control
- Keep the feature aligned with bash is all you need: CLI first, shell native, composable, no heavy UI by default

## Non-Goals

- Natural language understanding
- LLM integration
- Persistent task database
- Multi-session shared orchestrator
- Browser dashboard
- Advanced command history/autocomplete

## Open Questions

- Should `run <task-id>` also accept JSON input later?
- Should shell expose task list once scheduler supports listing?
