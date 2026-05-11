# Interactive Shell Design

## Overview

The interactive shell is a thin explicit in-process session over the existing Axis orchestrator. It provides a loop-based command interface without introducing Web UI, TUI frameworks, hidden daemon attachment, or new core scheduling abstractions.

Design philosophy:

- **More Context**: every command result includes clear context and suggested next steps
- **More Action**: users can submit tasks and inspect status without restarting the CLI
- **Zero Control**: invalid commands produce guidance, not hard exits
- **Bash is all you need**: CLI is the primitive interface; the shell is a lightweight composition layer, not a separate product surface

## Architecture

```text
cmd/axis/main.go
  ├── cobra command: shell
  └── shell runner
        ├── starts Orchestrator
        ├── reads stdin line by line
        ├── parses command tokens
        ├── calls existing Orchestrator methods
        └── gracefully shuts down
```

The shell is intentionally not the same lifecycle mode as the Local Control Plane runtime:

```text
axis shell
  -> owns its own in-process orchestrator session

axis start
  -> owns the project-local runtime used by separate CLI invocations
```

In P0, shell commands do not silently attach to or spawn `axis start`. Shell `run`, shell `ask --submit`, and shell `status` share state inside the same shell session.

## Components

### Shell command registration

Add a new Cobra command in `cmd/axis/main.go`:

```text
axis shell
```

It calls `runShell(cmd, args)`.

### Shell runner

`runShell` owns the interactive session lifecycle:

1. Create or reuse an orchestrator
2. Start it with a cancellable context
3. Print welcome/help hint
4. Enter scanner loop
5. Dispatch commands
6. Shutdown on exit/EOF/Ctrl+C

### Command parser

Keep parsing minimal:

- trim whitespace
- split by spaces
- first token is command
- remaining tokens are arguments

No shell quoting or JSON input in Milestone 1.

### Command handlers

Supported commands:

| Command | Behavior |
|---|---|
| `help` | print commands |
| `run <task-id>` | submit default pending task |
| `status <task-id>` | call `GetTaskStatus` on the shell session orchestrator |
| `exit` / `quit` | gracefully stop session |

### Orchestrator lifecycle

The shell starts the orchestrator in the background:

```text
ctx, cancel := context.WithCancel(context.Background())
go orch.Start(ctx)
```

On exit:

```text
cancel()
orch.Shutdown(context.Background())
```

### Error handling

Errors are printed with context and the shell continues unless the user exits.

Examples:

```text
axis> status missing
Task missing not found. Try: run missing
```

## File Structure

```text
cmd/axis/main.go              # add shell command and shell runner
```

No new packages are required for the first version.

## Trade-offs

| Option | Decision | Rationale |
|---|---|---|
| Full TUI | Rejected | Too heavy for Milestone 1 |
| Web UI | Rejected | Requires server/frontend, outside current scope |
| Interactive shell | Chosen | Minimal, CLI-native, close enough to Claude Code interaction loop |
| External readline package | Rejected | Avoid new dependency |
| Standard bufio.Scanner | Chosen | Simple and reliable |
| Browser-first UI | Rejected | Violates bash is all you need for the current milestone |

## Extension Point: Natural Language Scheduling

The shell command parser may later route `ask <prompt>` to the natural language scheduling layer described in `docs/specs/natural-language-scheduling/`.

This keeps the shell as a thin command loop while allowing natural language input to be compiled into ordinary Axis tasks. The shell must not become a separate chatbot runtime or bypass contracts, scheduler state, or task metadata.

## Risks

| Risk | Mitigation |
|---|---|
| Orchestrator startup races | Start once per shell session |
| Existing global `orch` state confusion | Reuse existing global but keep shell lifecycle explicit |
| Default contract may not exist | Return contextual error and keep shell alive |
| Ctrl+C behavior | Capture signal and trigger graceful shutdown |

## Acceptance Mapping

- FR1: Cobra command registration
- FR2: scanner prompt loop
- FR3: command switch
- FR4: contextual print messages
- FR5: command errors do not return from shell loop
- FR6: defer shutdown and signal handling
