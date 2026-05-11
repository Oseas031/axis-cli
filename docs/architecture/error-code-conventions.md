# Error Code Conventions

## Purpose

Error codes give Agents, scripts, and CLI users stable failure categories without creating a heavy exception framework.

## Format

Preferred format:

```text
AXIS_<DOMAIN>_<NUMBER>
```

Examples:

```text
AXIS_INTENT_001
AXIS_CONTRACT_001
AXIS_SCHEDULER_001
AXIS_PROVIDER_001
AXIS_CONTEXT_001
```

## Domains

| Domain | Scope |
|---|---|
| `INTENT` | natural language / intent parsing |
| `CONTRACT` | contract validation and execution contract errors |
| `SCHEDULER` | readiness, dependency, state transition errors |
| `ORCHESTRATOR` | coordination and lifecycle errors |
| `PROVIDER` | model provider creation and request errors |
| `TOOL` | tool execution errors |
| `CONTEXT` | adaptive context assembly errors |
| `CONFIG` | project-local config/profile errors |

## Rules

- Codes should be stable once exposed.
- Human messages may change; codes should not change casually.
- Use codes for actionable or script-observable failures.
- Do not assign codes to every internal error prematurely.

## CLI Shape

Human mode may show:

```text
Error AXIS_INTENT_001: prompt is required
```

Future JSON mode should show:

```json
{
  "status": "error",
  "code": "AXIS_INTENT_001",
  "message": "prompt is required"
}
```

## Registry Rule

When codes are introduced in code, maintain a documented registry in this file or a dedicated error-code registry.

## Initial Reserved Codes

| Code | Meaning |
|---|---|
| `AXIS_INTENT_001` | prompt is required |
| `AXIS_CONTRACT_001` | contract not found |
| `AXIS_CONTRACT_002` | task input does not satisfy contract |
| `AXIS_SCHEDULER_001` | dependency cycle detected |
| `AXIS_PROVIDER_001` | provider profile is invalid |
| `AXIS_CONTEXT_001` | context assembly found no usable goal |
