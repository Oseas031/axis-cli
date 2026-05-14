# Quick Start

**[Chinese version / 中文版](../zh/guides/QUICKSTART.md)**

## What is Axis

Axis is an Agent-native scheduling system and the execution substrate for Agent autogenesis.

It is not an ordinary task queue, nor an LLM wrapper. Axis enables Agents to express work as tasks, obtain context, execute actions, validate results, reflect on failures, generate follow-up tasks, and earn greater autonomy through reliable performance.

## Core Principles

```text
More Context, More Action, Zero Control, Controllable Evolution
bash is all you need · Competence earns autonomy · Interface is existence
```

## Current Capabilities (M1-M6 ✅)

- **Task Scheduling**: FIFO + DAG parallel scheduling, dependency management, contract admission, SLA strategy engine
- **LLM Integration**: Anthropic / OpenAI / DeepSeek / MiniMax providers, token accounting, circuit breaker
- **Tool System**: BashTool, FileReadTool, FileWriteTool, HTTPClientTool
- **Natural Language Scheduling**: `axis ask` compiles prompts into AgentTask, dry-run preview / explicit submit
- **Adaptive Context Assembly**: ContextBundle / ReadinessArtifact / preflight / strict gate
- **Local Control Plane**: `axis start` launches loopback control server, cross-process submit/query
- **Staged Evolution Protocol**: Isolated workspace + atomic steps + explicit promote/discard gate
- **Self-Judgement Engine**: 5 validation strategies + BootstrapOrchestrator judgement integration
- **Bootstrap Loop**: BootstrapOrchestrator + FollowUpTaskGenerator + AutonomyTransition

## Build & Test

```bash
go test ./...
go build -o axis-dev.exe ./cmd/axis
```

> On Windows, build as `axis-dev.exe` to avoid overwriting an existing `axis.exe` in the project root.

## Basic Usage

### Local Runtime (Cross-Process Mode)

```powershell
# Terminal A: start the local runtime
.\axis-dev.exe start

# Terminal B: submit a natural-language task
.\axis-dev.exe ask "check provider config" --submit --task-id provider-check

# Terminal B: query task status
.\axis-dev.exe status provider-check
```

### Single-Process Mode

```powershell
.\axis-dev.exe run my-task
.\axis-dev.exe shell
```

### Provider Configuration

```powershell
.\axis-dev.exe provider add claude --type anthropic --api-key sk-ant-... --model claude-3-5-sonnet-20241022
.\axis-dev.exe provider use claude
.\axis-dev.exe provider status
```

### More Commands

```powershell
.\axis-dev.exe ask "prompt"              # dry-run preview
.\axis-dev.exe context preview "prompt"  # context preview
.\axis-dev.exe context preflight <id>    # readiness check
.\axis-dev.exe judge                     # self-judgement diagnostic
.\axis-dev.exe evolve inspect <run-id>   # evolution inspection
```

## New Users

If this is your first time with Axis, try the [axis-up](../tools/axis-up.md) guided onboarding tool:

```powershell
cd tools\axis-up
go build -o axis-up.exe .
.\axis-up.exe start
```

## Next Steps

- Read [Agent-Native First Principles](../architecture/agent-native-first-principles.md) — **read before coding**
- Check [Current Progress](../status/current-progress.md)
- Check [Project Roadmap](../product/ROADMAP.md)
