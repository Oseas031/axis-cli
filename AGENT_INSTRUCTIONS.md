# AGENT HANDOVER — Single Entry Point

This is the **only** entry point for AI agents picking up the Axis project.
All coding constraints live in `CLAUDE.md` (the constitution). This file tells you **what to read and in what order**.

## Reading Order

1. **`CLAUDE.md`** — The constitution. All prohibitions, checklists, and engineering constraints.
2. **`docs/status/session-state.md`** — Live state (last step + next action). Read FIRST after CLAUDE.md to find out where work paused.
3. **`docs/status/current-progress.md`** — Ground truth for milestone status. (Do NOT rely on `HANDOVER.md` for status.)
4. **`docs/architecture/agent-native-first-principles.md`** — Six design principles.
5. **`workflow/entry.md`** — Route your task to the minimal workflow combination.
6. **`HANDOVER.md`** — Reference: project structure, completed work, known issues, next steps.

If editing `internal/kernel/`, `cmd/axis/`, `internal/contextpack/`, or `internal/agent/`, also read the adjacent `BOUNDARY.md`.

## Dev Loop

```bash
go build -o axis-dev.exe ./cmd/axis       # build
go test -race ./...                        # test (must pass)
gofmt -w . && go vet ./...                 # format + vet
staticcheck ./... && gosec ./...           # lint + security
```

## What NOT to Read

Anything in `docs/deprecated/` or files marked DRAFT / DEPRECATED.

## After Completing a Task

Synchronize docs by scope of impact:

| Changed | Update |
|---------|--------|
| Milestone progress | `docs/status/current-progress.md` |
| Known issues / next steps | `HANDOVER.md` |
| Doc index or directory structure | `docs/README.md` |
| Workflow status or routing | `workflows/README.md` |
