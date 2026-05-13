# Model Provider Requirements

> 实现 semantic-boundaries.md Provider 边界


## Summary

Add a minimal model provider layer to Axis so task execution can produce model-like results without coupling the kernel to any specific LLM vendor. The first implementation must be a `MockModelProvider` that requires no API key and is suitable for local development and beginner onboarding.

This feature should extend the current Milestone 1 execution path while preserving the core architecture.

This feature is bound to the existing project workflows documented in [workflow-binding.md](workflow-binding.md).

## Design Philosophy

The model provider layer must follow Axis's design principles:

- **More Context**: model requests and responses should carry enough task context to understand what was executed
- **More Action**: task execution should produce a meaningful result rather than only validating input
- **Zero Control**: provider selection should be replaceable; Axis must not force one model vendor or one execution path
- **Controllable Evolution**: provider changes must remain observable, testable, and reversible
- **Bash is All You Need, simple but robust, composable and extensible**: the feature must be usable from `axis shell` and ordinary CLI commands, with clear errors and extension points

## Users

- Developers validating Axis locally
- New users following `docs/guides/BEGINNER_GUIDE.md`
- Future agents that need a stable model execution abstraction

## Functional Requirements

### FR1: ModelProvider interface

Axis must define a model provider abstraction that can generate a response from a task prompt or input.

The abstraction must not depend on OpenAI, Claude, Gemini, or any external SDK.

### FR2: MockModelProvider

Axis must provide a mock provider that:

- requires no API key
- returns deterministic output
- includes task context in the output
- is suitable for tests and beginner demos

### FR3: Default execution path

The existing `default` contract execution path should use `MockModelProvider` so `axis shell` can demonstrate a complete execution path.

### FR4: Shell compatibility

The existing shell command:

```text
run <task-id>
```

must continue to work.

The output should make it clear that the result came from the mock provider, not a real model.

### FR5: No real model configuration yet

This feature must not require:

- API keys
- network calls
- provider config files
- external model SDKs

### FR6: Testability

The model provider must be testable with standard Go tests.

## Acceptance Criteria

- [x] `ModelProvider` interface exists
- [x] `MockModelProvider` exists
- [x] `go test ./...` passes
- [x] `axis shell` can run a task without `contract default not found`
- [x] task execution returns a mock model-like message
- [x] docs explain that real model providers are not yet connected

## Constraints

- Do not add real OpenAI / Claude / Gemini integration in this step
- Do not add API key handling yet
- Do not add Web UI
- Do not change scheduler semantics
- Prefer Go standard library
- Keep the provider layer small and replaceable

## Non-Goals

- Streaming model output
- Tool calling
- Conversation memory
- Provider routing
- Retry/backoff policies
- Cost tracking
- Prompt templates
- Real model provider implementation

## Open Questions

- Should future real providers be selected by environment variable or config file?
- Should shell add `ask <task-id> <prompt>` after MockModelProvider is complete?
- Should task input evolve from fixed `message: test` to user-provided prompt text?

## Provider Management Extension

Axis now manages real model providers through project-local files, not process environment mutation.

### FR7: Project-local provider profiles

Axis must store provider profiles under the current Agent project directory, using `.axis/providers.json` by default. The storage layer must not write system environment variables, registry keys, shell profiles, or user-global configuration.

Each profile must support:

- profile name
- provider type
- API key
- base URL
- default model
- temperature
- max context
- archived flag
- updated timestamp

### FR8: Task route mapping

Axis must represent task routes for:

- reasoning
- code generation
- writing
- tool calling

Each route maps to a profile name and optional model override.

### FR9: One-command switching

Axis must let a user switch active profile from the CLI. Switching must validate the target profile, create a backup, update active provider state, and leave existing profile data intact.

### FR10: Validation, backup, and recovery

Axis must validate provider config before and after writes. Invalid JSON or missing required fields must fail safely. Writes should be atomic enough for local file usage and create timestamped backups before destructive changes.

### FR11: Status visibility

Axis must expose current active profile, provider type, model, base URL, and config update time without printing secrets.

### FR12: Extensible provider adapters

Provider profile resolution must produce existing `provider.ProviderOption` values without embedding vendor-specific switching rules in the CLI command handlers.

## Provider Management Constraints

- No environment variable operations.
- No registry or system configuration mutation.
- No global user config.
- No web UI.
- No daemon.
- Prefer Go standard library and JSON for the first implementation.

