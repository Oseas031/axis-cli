# Semantic Boundaries

## Purpose

This document defines what each Axis concept is allowed to mean and do.

Good AI platform systems keep semantic boundaries explicit: a scheduler schedules, a provider generates model output, context prepares readiness, and permissions are not hidden inside unrelated layers.

## Boundary Table

| Concept | Owns | Must Not Own |
|---|---|---|
| `AgentTask` | task identity, contract ID, input, dependencies, status, metadata | execution logic, provider selection policy, hidden permissions |
| `AgentContract` | input/output shape and execution contract | scheduler policy, provider credentials, context retrieval |
| Scheduler | task readiness and status transitions | model calls, shell execution, provider config, natural language parsing |
| Orchestrator | module coordination and execution loop | provider profile storage, CLI rendering, hidden policy decisions |
| Dispatcher | route accepted tasks to execution path | admission policy, provider credential management |
| Provider | model request/response | task lifecycle, scheduler state, credentials persistence |
| ProviderConfig | project-local provider profile state | task execution, scheduler policy |
| Tool | bounded capability execution | global permission system, task scheduling |
| Intent Parser | user intent to structured Axis object | direct execution, context assembly, permission escalation |
| ContextBundle | task-specific readiness context | authority escalation, scheduler changes, task submission |
| EvolutionRun | isolated system-change proposal, atomic steps, verification evidence, promotion/discard decision | implicit main-tree mutation, hidden execution policy, automatic authority escalation |
| axis-up | external usability helper | internal imports, source mutation, core architecture |

## Core Rules

- Natural language produces structure; it does not execute by itself.
- Context improves action quality; it does not control or authorize action.
- Metadata adds audit and hints; it does not silently change core semantics.
- Providers answer model requests; they do not manage Axis state.
- CLI presents and invokes behavior; it should not become the domain layer.
- Draft evolution is not main system state.
- Verification is evidence for promotion; it is not promotion by itself.
- Discarded evolution remains audit history unless explicitly deleted by a documented cleanup path.

## Boundary Violation Signals

A change needs review if it:

- lets `intent` submit tasks without an explicit CLI/orchestrator path
- lets context packets expand tool/file/network access
- lets providers mutate scheduler state
- stores secrets outside provider config boundaries
- makes shell commands behave differently from CLI commands without a spec
- turns metadata into hidden policy enforcement
- lets an evolution workspace mutate the main tree before explicit promotion
- treats a successful verification command as automatic promotion

## Rule of Thumb

```text
If a module's name does not explain why it is allowed to perform an action, the action likely belongs elsewhere.
```
