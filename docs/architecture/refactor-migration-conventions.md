# Refactor and Migration Conventions

## Purpose

Axis will normalize already developed parts against newer conventions. This must be done safely.

## Core Rule

```text
Refactor behavior last, structure first, and only with tests.
```

## Migration Principles

### Small Steps

Move one boundary at a time.

### Behavior Preservation

CLI behavior, task semantics, provider resolution, and scheduler state transitions should remain stable unless the migration explicitly changes them.

### Test Before Move

Add or confirm tests before moving high-risk logic.

### Specs Stay Synchronized

When a migration changes module responsibility, update architecture docs or feature specs in the same change.

## Refactor Checklist

Before refactoring:

1. Identify current behavior.
2. Identify target convention.
3. Add tests for behavior that must remain stable.
4. Move code in the smallest useful slice.
5. Run targeted tests.
6. Run `go test ./...` for code changes.
7. Update docs if paths or semantics changed.

## Safe Refactor Targets

Likely future targets:

- extract shell command handling from `cmd/axis/main.go`
- share `ask` rendering between CLI and shell
- align metadata keys to namespaced form
- introduce `internal/contextpack` for context assembly
- separate command construction from command behavior if command files grow large

## Prohibited Migration Style

Avoid:

- broad rewrites without tests
- moving code and changing behavior in the same diff
- renaming public CLI output casually
- adding new abstractions before duplication or responsibility drift is clear
- turning helpers into global control planes

## Completion Criteria

A migration is complete when:

- old behavior is tested or intentionally replaced
- new location follows module conventions
- docs/specs are updated
- no stale duplicate implementation remains
