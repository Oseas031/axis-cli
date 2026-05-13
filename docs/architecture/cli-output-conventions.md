# CLI Output Conventions

> 展开自 CLAUDE.md §8（CLI 输出契约）


## Purpose

Axis is shell-native. CLI output is part of the product contract.

Outputs should be clear for humans and eventually stable for scripts.

## Output Modes

Default mode:

```text
human-readable
```

Future machine mode:

```text
--json
```

Do not introduce ad-hoc JSON fragments without documenting whether they are stable.

## Success Messages

Success messages should include:

- what happened
- primary ID
- suggested next command when useful

Example:

```text
Task ask-20260510-154000 submitted. Try: status ask-20260510-154000
```

## Preview Messages

Preview commands should clearly say they did not mutate state.

Example:

```text
Task proposal:
...
Not submitted. Use --submit to schedule this task.
```

## Error Messages

Errors should include:

- action that failed
- object ID when available
- concise cause
- next step when helpful

Avoid leaking secrets or raw provider credentials.

## Secret Redaction

Commands must not print:

- API keys
- bearer tokens
- provider secrets
- secret file contents

Provider `list` and `status` may show provider type, model, base URL, and active profile, but not API keys.

## Stability Rules

- Keep existing human output stable unless there is a reason to change it.
- For script-critical output, prefer adding `--json` instead of changing human text.
- Do not rely on color for meaning.
- Keep output line-oriented when possible.

## Command Categories

| Category | Default behavior |
|---|---|
| `run`, `ask --submit` | mutate state explicitly |
| `ask`, context preview | dry-run / preview |
| `status`, `provider status` | inspect state |
| `provider add/use/remove/archive` | mutate project-local provider config |
| `axis-up` commands | guide user through public CLI surfaces |

## JSON Output Future Shape

When added, JSON output should use stable snake_case fields:

```json
{
  "status": "ok",
  "task_id": "ask-123",
  "message": "submitted"
}
```

Errors should eventually support:

```json
{
  "status": "error",
  "code": "AXIS_INTENT_001",
  "message": "prompt is required"
}
```
