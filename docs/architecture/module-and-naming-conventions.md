# Module and Naming Conventions

## Purpose

This document defines Axis's module structure and naming conventions.

It is intended to guide all new development and to provide a future refactoring baseline for already implemented parts of the project.

The style is inspired by mature API/platform repositories: stable public entrypoints, clear internal boundaries, predictable names, small modules, and tests colocated with the behavior they protect.

## Core Principle

```text
Structure follows responsibility, not implementation convenience.
```

Axis modules should be grouped by what they are responsible for, not by temporary implementation details or by who created them.

## Design Goals

- Keep the project understandable for both Agents and humans.
- Make responsibilities visible from paths and names.
- Preserve the `internal` boundary for implementation details.
- Keep CLI code thin and orchestration code authoritative.
- Keep docs/specs aligned with code modules.
- Prefer additive reorganization over disruptive rewrites.

## Top-Level Layout

| Path | Responsibility |
|---|---|
| `cmd/axis/` | Main Axis CLI entrypoint and command wiring |
| `internal/` | Axis implementation packages not intended for external import |
| `docs/` | Product, architecture, specs, guides, status, and reference docs |
| `tools/` | External helper tools that operate through public CLI surfaces |
| `workflow/` | Active workflow rules and process guidance |
| `reports/` | Historical analysis, audits, strategy, and review reports |
| `scripts/` | Local helper scripts, if needed |

## Project-Local Runtime Directory (`.axis/`)

The `.axis/` directory stores project-local runtime state. It is gitignored.

| Path | Responsibility |
|---|---|
| `.axis/runtime.json` | Local runtime metadata (loopback server address) |
| `.axis/providers.json` | Project-local provider profiles |
| `.axis/events/tasks.jsonl` | Task event log |
| `.axis/skills/<name>/SKILL.md` | On-demand knowledge skills (see `docs/specs/skills-system/`) |

### `.axis/skills/` Rules

- Each skill is a directory named in kebab-case containing a `SKILL.md` file.
- Optional subdirectories: `scripts/`, `references/`.
- Skill names must match `^[a-z][a-z0-9-]*[a-z0-9]$`.
- No nested skills. No remote fetching. Pure local files.

## Internal Package Layout

### `internal/types`

Shared domain types used across modules.

Examples:

- `AgentTask`
- `AgentContract`
- `TaskResult`
- common metadata keys

Rules:

- Keep this package small.
- Do not put behavior-heavy services here.
- Add types here only when multiple modules need the same domain concept.

### `internal/kernel/<module>`

Core Axis runtime modules.

Current or expected modules:

```text
internal/kernel/orchestrator
internal/kernel/scheduler
internal/kernel/dispatcher
internal/kernel/lifecycle
internal/kernel/sharedlayer
```

Rules:

- Kernel packages own runtime state transitions.
- CLI must call kernel through explicit APIs, not mutate internals.
- Scheduler/orchestrator semantics must be protected by tests.

### `internal/contract/<module>`

Contract validation and execution path.

Current modules:

```text
internal/contract/admission
internal/contract/executor
```

Rules:

- Contract admission validates whether a task can enter execution.
- Contract execution handles execution behind an accepted contract.
- Do not mix provider profile management into contract packages.

### `internal/model/<module>`

Model provider and tool integration.

Current modules:

```text
internal/model/provider
internal/model/providerconfig
internal/model/tool
```

Rules:

- Provider implementations belong under `provider`.
- Project-local provider profile management belongs under `providerconfig`.
- Tool implementations belong under `tool`.
- Do not create provider-specific top-level packages unless a provider becomes large enough to justify it.

### `internal/agent/<module>`

Agent execution and judgement logic.

Current modules:

```text
internal/agent
internal/agent/contracts
internal/agent/judgement
internal/agent/judgement/strategies
```

Rules:

- Agent behavior belongs here, not in CLI.
- Judgement strategies should be explicit and testable.
- Avoid hiding policy-heavy control logic inside generic agent helpers.

### `internal/intent`

Natural language or user intent conversion into structured Axis objects.

Rules:

- Intent parsing ends at structured output such as `AgentTask`.
- It must not execute tasks directly except through CLI/orchestrator integration.
- It must preserve provenance metadata.

### `internal/contextpack`

Planned package for Adaptive Context Assembly.

Rules:

- Use `contextpack`, not `context`, to avoid conflict with Go's standard `context` package.
- It prepares context bundles; it does not grant permissions or execute tools.
- It should remain preview-first until execution injection semantics are approved.

## CLI Conventions

CLI code lives in:

```text
cmd/axis/
```

File naming:

```text
main.go
<feature>_cmd.go
<feature>_cmd_test.go
```

Examples:

```text
ask_cmd.go
ask_cmd_test.go
provider_cmd.go
provider_cmd_test.go
```

Command constructor naming:

```go
func newAskCommand() *cobra.Command
func newProviderCommand() *cobra.Command
```

Command behavior naming:

```go
func runTask(cmd *cobra.Command, args []string) error
func getTaskStatus(cmd *cobra.Command, args []string) error
```

Rules:

- CLI commands should be thin adapters.
- Business logic should live in `internal/*` packages.
- CLI output should be stable, clear, and shell-friendly.
- Prefer explicit flags over hidden behavior.
- Use `cmd.OutOrStdout()` and `cmd.InOrStdin()` in command files when testability matters.

## Tool Conventions

External helper tools live under:

```text
tools/<tool-name>/
```

Current tools:

```text
tools/axis-up/       # Onboarding helper
tools/axis-gui/      # Observation dashboard (React + Go HTTP server)
```

Rules:

- Tools must not import `github.com/axis-cli/axis/internal/...`.
- Tools must have their own `go.mod` (independent module).
- Tools should call public Axis binaries or read `.axis/` files only.
- Tools must not mutate Axis state or source code.
- Tools should optimize usability without invading Axis core.
- Tool source is tracked in git; build artifacts and `node_modules/` are gitignored.
- Each tool must have a `README.md` documenting purpose, build, and boundary rules.

## Documentation Conventions

### Architecture docs

Architecture references live under:

```text
docs/architecture/
```

Use descriptive kebab-case filenames:

```text
agent-native-first-principles.md
bash-is-all-you-need.md
module-and-naming-conventions.md
```

### Spec docs

Feature specs live under:

```text
docs/specs/<feature-name>/
```

Feature folders use kebab-case:

```text
natural-language-scheduling
adaptive-context-assembly
interactive-shell
model-provider
```

Every active feature spec should prefer this shape:

```text
requirements.md
design.md
tasks.md
```

Optional workflow binding:

```text
workflow-binding.md
```

Use workflow binding when a feature is large enough that implementation order, quality gates, or handoff synchronization must be explicit.

## Naming Style

### Folders

Use lowercase names.

Preferred:

```text
providerconfig
sharedlayer
contextpack
natural-language-scheduling
```

Rules:

- Go package folders should be lowercase and usually one word.
- Documentation folders should use kebab-case.
- Avoid vague folders like `common`, `utils`, `misc`, or `helpers` unless strongly justified.

### Go packages

Use short lowercase package names.

Preferred:

```go
package scheduler
package orchestrator
package providerconfig
package contextpack
```

Avoid:

```go
package axisScheduler
package provider_config
package helpers
```

### Go files

Use snake_case for multiword Go filenames.

Examples:

```text
provider_cmd.go
registry_test.go
ask_cmd_test.go
```

### Exported names

Export only stable concepts used across packages.

Examples:

```go
type AgentTask struct {}
type ModelProvider interface {}
func NewProvider(...) (...)
```

Unexport package-local helpers unless cross-package use is intentional.

### Constructors

Use `New<Type>` for domain constructors:

```go
func NewOrchestrator(opts ...OrchestratorOption) *Orchestrator
func NewDeterministicParser() *DeterministicParser
```

Use `new<Feature>Command` for Cobra command constructors:

```go
func newAskCommand() *cobra.Command
```

### Options

Use functional options when construction has optional behavior:

```go
type OrchestratorOption func(*Orchestrator)
func WithModelProvider(p provider.ModelProvider) OrchestratorOption
```

Do not introduce options for simple structs unless they improve clarity.

## Test Conventions

Tests should live beside the code they verify.

Naming:

```text
<file>_test.go
```

Test function names:

```go
func TestDeterministicParser_ParseDefaultTask(t *testing.T)
func TestAskCommand_DryRunByDefault(t *testing.T)
func TestScheduler_GetReadyTasks_ClaimsTasksOnce(t *testing.T)
```

Rules:

- Test behavior, not implementation trivia.
- Protect public semantics and high-risk invariants.
- Add targeted package tests before relying on `go test ./...`.
- Use regression tests for bugs that were fixed.

## Dependency Direction

Preferred dependency direction:

```text
cmd/axis
  -> internal/*

internal/kernel
  -> internal/contract
  -> internal/model
  -> internal/types

internal/intent
  -> internal/types

internal/contextpack
  -> internal/types
```

Rules:

- `internal/types` should not depend on higher-level packages.
- Kernel packages should not import CLI packages.
- Tools should not import `internal` packages.
- Specs should describe behavior but not become runtime dependencies.

## Feature Addition Checklist

When adding a new feature, decide first:

1. Is this runtime behavior, CLI surface, external helper, or documentation only?
2. Does it belong in an existing package?
3. Does it need a new package because it has a distinct responsibility?
4. Does it need `docs/specs/<feature>/requirements.md`, `design.md`, and `tasks.md`?
5. Does it change scheduler, orchestrator, contract, or provider semantics?
6. What tests protect the new behavior?
7. What docs must be synchronized?

## Refactoring Guidance for Existing Code

Future cleanup should use this document as a baseline.

Priorities:

1. Keep behavior stable.
2. Move logic out of CLI only when there is clear reusable domain behavior.
3. Preserve public command behavior unless intentionally changed.
4. Add tests before moving high-risk code.
5. Prefer small package-level moves over broad rewrites.
6. Update docs and specs in the same change.

## Anti-Patterns

Avoid:

- Large `main.go` files that contain domain logic.
- Generic `utils` packages with mixed responsibilities.
- Hidden permission or policy logic inside context, model, or CLI helpers.
- Provider-specific logic spread across unrelated packages.
- Specs that describe behavior no code can satisfy.
- Code changes that contradict active specs without updating them.
- External tools importing `internal` packages.

## Current Known Alignment Gaps

These are not immediate bugs. They are future refactoring targets:

- Shell behavior has been extracted to `cmd/axis/shell_cmd.go` (normalized per SWE1.6).
- Natural language scheduling now uses shared rendering helper `renderTaskProposal` in `cmd/axis/ask_cmd.go` (normalized per SWE1.6).
- Intent metadata keys in `internal/intent/parser.go` now use namespaced `intent.*` keys while maintaining legacy keys for backward compatibility (normalized per SWE1.6).
- Adaptive Context Assembly is implemented in `internal/contextpack` (accepted, T1-T9 completed).
- Existing docs/specs may not all include workflow binding; only larger features need it.

## Rule of Thumb

```text
If a name does not reveal responsibility, rename it.
If a package mixes responsibilities, split it.
If a feature needs planning, write the spec before the code.
If a helper becomes reusable domain behavior, move it out of CLI.
```
