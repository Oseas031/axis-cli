# Axis System Conventions

> 展开自 CLAUDE.md §7（代码与架构风格）


## Purpose

This is the convention map for Axis.

It turns scattered engineering expectations into a small system of stable rules. These conventions are inspired by mature AI platform engineering practices: clear API boundaries, safe defaults, auditable behavior, explicit lifecycle states, stable machine-readable outputs, and incremental migration.

Axis should use these documents as the baseline for future implementation and for normalizing already developed modules.

## Convention Map

| Area | Document | Priority |
|---|---|---|
| Module layout and naming | [module-and-naming-conventions.md](module-and-naming-conventions.md) | P0 |
| Semantic boundaries | [semantic-boundaries.md](semantic-boundaries.md) | P0 |
| Metadata keys | [metadata-key-conventions.md](metadata-key-conventions.md) | P0 |
| CLI output | [cli-output-conventions.md](cli-output-conventions.md) | P0 |
| Spec lifecycle | [spec-lifecycle-conventions.md](spec-lifecycle-conventions.md) | P0 |
| Data model evolution | [data-model-evolution.md](data-model-evolution.md) | P1 |
| Error codes | [error-code-conventions.md](error-code-conventions.md) | P1 |
| External tool boundaries | [external-tool-boundaries.md](external-tool-boundaries.md) | P1 |
| Secret handling | [secret-handling.md](secret-handling.md) | P1 |
| Refactor migration | [refactor-migration-conventions.md](refactor-migration-conventions.md) | P1 |

## Global Principles

### Stable Surfaces, Replaceable Internals

Public commands, task structures, metadata keys, and specs should be stable. Internal implementations can evolve behind them.

### Safety Defaults

Dry-run, preview, redaction, validation, and explicit submit are preferred defaults for high-impact actions.

### Auditable by Design

Important decisions should leave a trace: selected context, parser mode, provider profile, metadata source, error code, or task transition.

### Machine-Friendly Without Losing Human Clarity

Axis should keep human-readable CLI output, but important operations should have a path toward stable machine-readable output.

### Small Contracts Over Large Control Planes

Axis should prefer explicit task, contract, metadata, and context structures over hidden global controllers.

### Progressive Evolution

Move from deterministic and local behavior to richer adaptive behavior only after traces, tests, and compatibility are stable.

## Usage

When adding or changing a feature:

1. Check this map.
2. Follow the relevant convention documents.
3. Update specs before changing core behavior.
4. Add tests for changed semantics.
5. Preserve compatibility or document the migration.
