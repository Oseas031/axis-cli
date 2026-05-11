# axis-up

**[Chinese version / 中文版](../zh/tools/axis-up.md)**

`axis-up` is an external human onboarding tool for Axis.

It is not part of the Axis core, nor a core module of the scheduler, contract system, provider system, or permission system. Its responsibility is to help first-time human users quickly complete environment checks, builds, zero-config demos, and common issue fixes.

## Positioning

```text
axis    = Agent-native execution core
axis-up = Human onboarding companion
```

`axis-up` exists to avoid putting beginner usability logic into the Axis core, while keeping Axis's CLI-first, shell-native, composable direction.

## Design Principles

- **Commands by user intent**: Commands express what the user wants to do, not technical steps.
- **One entry covers first use**: `axis-up start` covers most first-time scenarios.
- **Smart detection, technical transparency**: Automatically detects environment state while explaining why each step matters.
- **Progressive disclosure**: Default to running with mock provider first, then guide toward real provider configuration.
- **External tool**: Does not import Axis `internal` packages, does not modify Axis source code.
- **Public CLI boundary**: Interacts with Axis through `axis-dev.exe` / `axis` public commands.

## Commands

```bash
axis-up start
axis-up check
axis-up demo
axis-up fix
```

## First-Time Path

Recommended for new users:

```bash
cd path/to/axis-cli
cd tools/axis-up
go build -o axis-up.exe .
.\axis-up.exe start
```

`start` runs a guided flow:

```text
Detect Axis repo → Detect Go → Build axis-dev.exe if needed → Use mock provider → Run demo
```

## Why Mock Provider by Default

First-time experience should not depend on:

- API keys
- External network access
- Model billing
- Provider compatibility

Therefore `axis-up` defaults to mock provider, letting users first understand Axis's task submission and execution mental model.

Real provider configuration is the next level of experience, achievable through Axis public commands:

```bash
axis-dev.exe provider add <name> --type <provider> --api-key <key> --model <model>
axis-dev.exe provider use <name>
```

## Tool Documentation

Detailed implementation notes are kept in the tool directory:

- [`tools/axis-up/README.md`](../../tools/axis-up/README.md)
- [`tools/axis-up/DESIGN.md`](../../tools/axis-up/DESIGN.md)

This way, if `axis-up` is ever split into an independent repository, the tool docs can migrate directly.

## Boundaries

`axis-up` can:

- Check local environment
- Build `axis-dev.exe`
- Call Axis public CLI commands
- Explain next steps
- Fix safe onboarding issues

`axis-up` should NOT:

- Import `github.com/axis-cli/axis/internal/...`
- Modify Axis source code
- Become a new Axis main entry point
- Introduce Web UI / heavy TUI
- Silently overwrite user provider configuration
