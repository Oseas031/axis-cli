# Model Provider Requirements

## Summary

Add a minimal model provider layer to Axis so task execution can produce model-like results without coupling the kernel to any specific LLM vendor. The first implementation must be a `MockModelProvider` that requires no API key and is suitable for local development and beginner onboarding.

This feature should extend the current Milestone 1 execution path while preserving the core architecture.

This feature is bound to the existing project workflows documented in [workflow-binding.md](workflow-binding.md).

## Design Philosophy

The model provider layer must follow Axis's design principles:

- **More Context**: model requests and responses should carry enough task context to understand what was executed
- **More Action**: task execution should produce a meaningful result rather than only validating input
- **Zero Control**: provider selection should be replaceable; Axis must not force one model vendor
- **Bash is All You Need**: the feature must be usable from `axis shell` and ordinary CLI commands

## Users

- Developers validating Axis locally
- New users following `docs/BEGINNER_GUIDE.md`
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

- [ ] `ModelProvider` interface exists
- [ ] `MockModelProvider` exists
- [ ] `go test ./...` passes
- [ ] `axis shell` can run a task without `contract default not found`
- [ ] task execution returns a mock model-like message
- [ ] docs explain that real model providers are not yet connected

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
