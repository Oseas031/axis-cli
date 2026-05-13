# Secret Handling

> 展开自 CLAUDE.md §1.7（禁止输出密钥）


## Purpose

Axis supports real model providers, so secret handling must be explicit.

## Secret Types

Secrets include:

- API keys
- bearer tokens
- private keys
- provider credentials
- raw authorization headers
- secret file contents

## Storage Rules

Project-local provider credentials may be stored in:

```text
.axis/providers.json
```

Rules:

- Do not store secrets in source files.
- Do not store secrets in tests.
- Do not include real keys in docs or examples.
- Do not print secrets in CLI output.
- Prefer local project config over global shell mutation.

## Output Redaction

Commands such as provider `list` and `status` may show:

- active profile
- provider type
- model
- base URL
- updated time

They must not show:

- API key
- token prefix unless explicitly designed and safe
- raw credential values

## Test Rules

Tests must use:

- fake keys
- environment-provided keys for manual tests
- skipped integration tests when external credentials are unavailable

Manual test files must not contain real keys.

## Git Rules

Generated credential files should not be committed unless explicitly safe.

If a secret is accidentally committed:

1. Remove it from code immediately.
2. Rotate the credential.
3. Add regression checks if possible.
4. Document the fix if relevant.

## Provider Endpoint Notes

Provider endpoint mistakes can look like auth failures. Diagnose endpoint correctness before assuming a key is invalid.

MiniMax Token Plan keys must use:

```text
https://api.minimaxi.com/v1
```
