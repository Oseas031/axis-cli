# Data Model Evolution

> 展开自 CLAUDE.md §11（演化原则）


## Purpose

Axis data models must evolve without breaking task execution, CLI use, or future agent integrations.

Core models include:

- `AgentTask`
- `AgentContract`
- `TaskResult`
- provider profiles
- context packets and bundles

## Evolution Principles

### Additive First

Prefer adding optional fields over changing existing field meanings.

### Typed Fields for Core Semantics

Use typed fields when a value affects validation, scheduling, execution, or public APIs.

### Metadata for Audit and Experimental Hints

Use metadata for provenance, optional hints, and experimental values.

### Backward Compatibility

When changing a model:

1. Keep reading old shape if possible.
2. Write new shape.
3. Document migration.
4. Add regression tests.

## Field Naming

JSON fields use snake_case:

```json
{
  "task_id": "t1",
  "contract_id": "default"
}
```

Go fields use exported PascalCase when cross-package:

```go
TaskID string `json:"task_id"`
```

## Metadata vs Typed Field Decision

Use metadata if:

- optional
- audit-only
- experimental
- not needed by core validation

Use typed field if:

- required by scheduler or admission
- required by contract execution
- required by stable CLI/API output
- used by multiple packages as core behavior

## Removal Rule

Never remove an observed field without:

- migration note
- tests
- compatibility period
- spec update

## Examples

`intent.original_prompt` can remain metadata because it is provenance.

A future `TimeoutMs` may become typed if scheduler/executor enforce it.
