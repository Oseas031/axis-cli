# Metadata Key Conventions

> 展开自 CLAUDE.md §1.6（命名空间元数据键）


## Purpose

Axis uses metadata for lightweight audit, provenance, and optional hints. Metadata must remain predictable and namespaced.

## Key Format

Preferred format:

```text
namespace.key
```

Examples:

```text
intent.original_prompt
intent.parser_mode
context.bundle_id
context.assembly_mode
context.requested_sources
evolution.run_id
evolution.step_id
tool.allowed_paths
sla.timeout_ms
provider.profile
```

Legacy un-namespaced keys may exist, but new keys should use namespaces.

## Namespaces

| Namespace | Purpose |
|---|---|
| `intent.*` | natural language and intent parser provenance |
| `context.*` | context assembly bundle and trace metadata |
| `evolution.*` | sandboxed evolution run, step, verification, and decision provenance |
| `tool.*` | task-local tool boundaries and audit hints |
| `sla.*` | timeout, retry, and service-level hints |
| `provider.*` | provider profile and model selection provenance |
| `axis.*` | Axis runtime-generated metadata |

## Rules

- Metadata must not store secrets.
- Metadata must not silently grant permissions.
- Metadata must not silently promote, discard, or verify evolution work.
- Metadata keys should be stable once external users can observe them.
- Prefer typed fields when a value becomes required for core semantics.
- Prefer metadata when a value is optional, experimental, or audit-oriented.

## Promotion Rule

Move a metadata key to a typed field when:

- multiple core modules require it
- validation depends on it
- tests need stable structured access
- CLI/API consumers rely on it as a primary field

## Redaction Rule

Never place these in metadata:

- API keys
- bearer tokens
- private keys
- raw credentials
- local secret file contents

## Compatibility

When replacing a legacy key:

1. Write the new namespaced key.
2. Read both old and new keys during transition.
3. Document the deprecation.
4. Remove the old key only after a migration window.
