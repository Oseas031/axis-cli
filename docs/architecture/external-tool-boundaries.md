# External Tool Boundaries

## Purpose

External tools improve usability without invading Axis core.

Current external tools:

```text
tools/axis-up/   — Human onboarding companion
tools/axis-gui/  — Local Web Dashboard (connects to Local Control Plane)
```

## Core Rule

```text
External tools call Axis from the outside.
They do not become Axis internals.
```

## Allowed

External tools may:

- call public Axis binaries or commands
- inspect public files such as README, docs, and config files
- guide users through setup
- build local development binaries when explicit
- write tool-local artifacts
- write documented project-local config when that is their purpose

## Not Allowed

External tools must not:

- import `github.com/axis-cli/axis/internal/...`
- mutate Axis source code as a side effect of normal use
- bypass public CLI behavior
- depend on private in-memory runtime state
- print secrets
- overwrite `axis.exe` unless explicitly approved

## axis-up Specific Boundary

`axis-up` should remain:

- user-intent grouped
- progressive disclosure oriented
- safe for first-time users
- mock-provider friendly by default
- transparent about commands it runs

## Command Safety

External tools should classify actions as:

- check-only
- preview
- local build
- config mutation
- execution

Mutating actions should be explicit.

## Dependency Rule

A tool can depend on its own module dependencies, but must not force new dependencies into Axis core.
