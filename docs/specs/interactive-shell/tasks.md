# Interactive Shell Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)

## Progress Tracking

| Task | Status |
|---|---|
| T1: Add shell command | Completed |
| T2: Implement shell loop | Completed |
| T3: Add graceful shutdown handling | Completed |
| T4: Verify build and basic shell commands | Completed |
| T5: Update HANDOVER.md | Completed |
| T6: Track natural language scheduling extension | Planned |

---

## T1: Add shell command

**Goal**: Register `axis shell` in the existing Cobra CLI.

**Files**:

- `cmd/axis/main.go`

**Design References**:

- `design.md` — Shell command registration

**Acceptance Criteria**:

- `axis shell --help` is available
- Existing `run`, `status`, and `start` commands remain unchanged

**Depends on**: None

---

## T2: Implement shell loop

**Goal**: Add a minimal prompt loop using Go standard library.

**Files**:

- `cmd/axis/main.go`

**Design References**:

- `design.md` — Shell runner
- `design.md` — Command parser

**Acceptance Criteria**:

- Shell prints a welcome message
- Shell accepts repeated input
- `help`, `run <task-id>`, `status <task-id>`, `exit`, and `quit` are handled
- Invalid commands print guidance and keep the shell alive

**Depends on**: T1

---

## T3: Add graceful shutdown handling

**Goal**: Ensure the shell exits cleanly on command exit, EOF, or Ctrl+C.

**Files**:

- `cmd/axis/main.go`

**Design References**:

- `design.md` — Orchestrator lifecycle

**Acceptance Criteria**:

- `exit` shuts down the orchestrator cleanly
- `quit` shuts down the orchestrator cleanly
- Ctrl+C exits without panic
- EOF exits without panic

**Depends on**: T2

---

## T4: Verify build and basic shell commands

**Goal**: Confirm the feature compiles and the basic interaction works.

**Files**:

- None, verification only

**Commands**:

```bash
go build -o axis.exe cmd/axis/main.go
```

Manual checks:

```text
axis shell
axis> help
axis> run demo-task
axis> status demo-task
axis> unknown
axis> exit
```

**Acceptance Criteria**:

- Build succeeds
- Shell starts
- Help displays commands
- Commands do not crash the shell

**Depends on**: T3

---

## T5: Update HANDOVER.md

**Goal**: Record the interactive shell addition and its design scope.

**Files**:

- `HANDOVER.md`

**Acceptance Criteria**:

- HANDOVER mentions `axis shell`
- HANDOVER states this is a lightweight CLI client layer, not core scheduler architecture

**Depends on**: T4

---

## T6: Track natural language scheduling extension

**Goal**: Record natural language task scheduling as a non-destructive extension of the shell, not part of the original shell milestone.

**Files**:

- `docs/specs/interactive-shell/requirements.md`
- `docs/specs/interactive-shell/design.md`
- `docs/specs/natural-language-scheduling/`

**Acceptance Criteria**:

- Shell scope remains command-based for the completed milestone
- Future `ask` integration points to the dedicated natural language scheduling spec
- No runtime code changes are required

**Depends on**: T5
