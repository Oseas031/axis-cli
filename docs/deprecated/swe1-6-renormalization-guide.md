# SWE1.6 Renormalization Guide

## Purpose

This guide instructs SWE1.6 how to normalize existing Axis code against the current convention set.

It is not a feature implementation plan. It is a disciplined refactor and alignment protocol.

## Scope

SWE1.6 may normalize existing code, tests, and docs to follow:

- [axis-system-conventions.md](axis-system-conventions.md)
- [module-and-naming-conventions.md](module-and-naming-conventions.md)
- [semantic-boundaries.md](semantic-boundaries.md)
- [metadata-key-conventions.md](metadata-key-conventions.md)
- [cli-output-conventions.md](cli-output-conventions.md)
- [spec-lifecycle-conventions.md](spec-lifecycle-conventions.md)
- [data-model-evolution.md](data-model-evolution.md)
- [error-code-conventions.md](error-code-conventions.md)
- [external-tool-boundaries.md](external-tool-boundaries.md)
- [secret-handling.md](secret-handling.md)
- [refactor-migration-conventions.md](refactor-migration-conventions.md)
- [../specs/sandboxed-evolution/](../specs/sandboxed-evolution/)

## Non-Goals

SWE1.6 must not use renormalization as an excuse to add new product capability.

SWE1.6 must not treat high-impact system evolution as ordinary cleanup.

Do not implement:

- new scheduler semantics
- new provider features
- new LLM behavior
- Adaptive Context Assembly runtime
- permission systems
- new external services
- broad UX redesigns
- large rewrites without tests
- automatic promotion, discard, or verification of evolution work
- control-logic rewriting outside the Staged Evolution Protocol

## Core Rule

```text
Normalize structure and semantics without changing user-visible behavior.
```

If behavior must change, stop and create a separate spec or task.

## Operating Mode

SWE1.6 should work in small, independently reviewable slices.

Preferred slice size:

```text
one boundary, one package, or one CLI surface at a time
```

Avoid mixed diffs that combine:

- file moves
- behavior changes
- naming changes
- output changes
- new features

## Required Workflow

### Step 1: Baseline

Before changing code:

1. Run or inspect relevant tests.
2. Identify current behavior.
3. Identify the convention being applied.
4. List files likely to change.
5. State what must remain unchanged.

### Step 2: Classify the issue

Use one of these categories:

| Category | Meaning |
|---|---|
| `module-boundary` | code lives in the wrong package or layer |
| `naming` | file, package, function, or metadata name violates conventions |
| `semantic-boundary` | module owns behavior it should not own |
| `cli-output` | output is unclear, unstable, or unsafe |
| `metadata` | metadata key is unnamespaced or misused |
| `secret` | secret handling risk |
| `test-gap` | behavior is not protected before migration |
| `doc-drift` | docs and code disagree |
| `evolution-boundary` | change modifies system structure and should use Staged Evolution Protocol |

If an issue is classified as `evolution-boundary`, SWE1.6 should stop ordinary renormalization and create or update the relevant spec/task instead of directly changing core behavior.

### Step 3: Add or confirm tests

Before moving high-risk logic, ensure tests cover the behavior.

High-risk behavior includes:

- task submission
- task status transitions
- provider profile handling
- API key redaction
- shell command behavior
- `axis ask` dry-run and submit behavior
- scheduler claim semantics
- tool execution metadata
- staged evolution promotion/discard boundaries

### Step 4: Apply the smallest change

Examples of acceptable slices:

- extract shell command handling from `main.go` into `shell_cmd.go`
- move shared ask rendering into a local helper
- rename metadata keys while preserving legacy reads
- add missing docs index entries
- add redaction tests for provider output

### Step 5: Verify

Run targeted tests first.

Then run:

```bash
go test ./...
```

For docs-only changes, state that Go tests were not required.

### Step 6: Update docs

If paths, semantics, metadata keys, output contracts, or lifecycle states changed, update the relevant convention/spec docs.

If the change touches high-impact evolution surfaces, update or reference the Staged Evolution Protocol rather than normalizing the behavior directly.

## Priority Audit Targets

### P0: CLI and shell normalization

Current known target:

```text
cmd/axis/main.go
```

Potential actions:

- move shell command code to `shell_cmd.go`
- keep `main.go` focused on root command wiring and app initialization
- preserve shell behavior and tests

### P0: Ask command normalization

Current known files:

```text
cmd/axis/ask_cmd.go
cmd/axis/main.go
internal/intent/parser.go
```

Potential actions:

- share task proposal rendering between CLI ask and shell ask
- ensure intent metadata follows namespaced key conventions in a compatibility-safe way
- keep dry-run default and explicit submit behavior

### P0: Provider command safety

Current known files:

```text
cmd/axis/provider_cmd.go
internal/model/providerconfig/
```

Potential actions:

- verify no API keys are printed
- align provider output with CLI output conventions
- ensure provider config errors are clear

### P1: Metadata key alignment

Targets:

- intent metadata
- tool metadata
- future context metadata
- SLA metadata

Rules:

- prefer namespaced keys for new writes
- preserve legacy compatibility where already written
- add tests before changing observable metadata

### P1: External tool boundary audit

Target:

```text
tools/axis-up/
```

Check:

- no `internal` imports
- no source mutation by default
- safe handling of local binaries
- clear progressive disclosure

### P1: Spec status alignment

Targets:

```text
docs/specs/*/tasks.md
```

Check:

- task statuses reflect actual implementation
- paused features are marked `Paused`, not deprecated
- completed runtime tasks have verification notes when needed

### P1: Evolution boundary alignment

Targets:

```text
docs/specs/sandboxed-evolution/
docs/architecture/spec-lifecycle-conventions.md
docs/architecture/semantic-boundaries.md
docs/architecture/metadata-key-conventions.md
```

Check:

- structural changes are identified as evolution work when appropriate
- draft, verified, promoted, and discarded states remain distinct
- verification is not treated as automatic promotion
- `evolution.*` metadata remains provenance, not hidden control
- ordinary renormalization does not mutate main-system semantics through an evolution workspace

## Decision Rules

### When to move code

Move code when:

- responsibility is clear
- tests protect behavior
- the new location matches module conventions
- imports do not create cycles

Do not move code just to reduce file size.

### When to rename

Rename when:

- the current name hides responsibility
- the new name is convention-aligned
- references can be updated safely

Do not rename stable user-facing commands without a spec.

### When to add abstraction

Add abstraction only when:

- there are at least two real call sites
- the abstraction has a clear domain name
- it reduces semantic duplication

Do not create `utils`, `helpers`, or vague manager objects.

### When to stop

Stop and ask for direction if:

- a change alters CLI output relied on by users
- tests reveal existing behavior is inconsistent with specs
- a migration requires changing task semantics
- a convention conflicts with working code in a non-trivial way
- secrets may have been exposed
- a cleanup task would modify contract, workflow, context rule, permission, promotion, or self-modification semantics
- an evolution workspace or verification command would be needed to make the change safely

## Output Expected from SWE1.6

Each renormalization pass should produce a short report:

```text
Scope:
Convention applied:
Evolution boundary check:
Files changed:
Behavior preserved:
Tests run:
Remaining gaps:
```

## Acceptance Criteria

A SWE1.6 renormalization task is complete only when:

- changes are limited to the declared scope
- behavior is preserved or intentionally documented
- tests pass or docs-only rationale is stated
- no new feature scope is introduced
- relevant docs/specs are synchronized
- remaining gaps are listed clearly

## Final Rule

```text
Do not make Axis bigger while making it cleaner.
```

Renormalization should reduce ambiguity, not expand scope.
