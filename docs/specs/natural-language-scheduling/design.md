# Natural Language Scheduling Design

## Overview

Natural language scheduling is a thin intent-to-task compiler layered above the existing Axis CLI and orchestrator. It accepts user text, creates an auditable task proposal, and optionally submits that proposal as a normal `types.AgentTask`.

It does not introduce a new execution engine. The scheduler, dispatcher, contract executor, and provider system remain authoritative.

## Architecture

```text
CLI / shell input
  -> intent parser
  -> task proposal
  -> dry-run renderer or orchestrator.SubmitTask
  -> existing scheduler/dispatcher/executor path
```

## Components

### CLI entrypoint

Add a future Cobra command:

```text
axis ask <prompt>
```

Recommended flags:

| Flag | Purpose |
|---|---|
| `--dry-run` | Render proposed task without submitting |
| `--submit` | Submit proposed task |
| `--contract` | Select target contract, default `default` |
| `--task-id` | Optional explicit task ID |
| `--stdin` | Read prompt from stdin for shell composition |

### Intent parser package

Recommended package:

```text
internal/intent
```

P0 parser should be deterministic:

```text
prompt -> AgentTask proposal
```

No LLM is required in P0. Later phases may add an LLM-backed parser behind the same interface.

### Core interface

```go
type Parser interface {
    Parse(ctx context.Context, req Request) (*Result, error)
}
```

Suggested data shapes:

```go
type Request struct {
    Prompt          string
    ContractID      string
    TaskID          string
    ParserMode      string
}

type Result struct {
    Task       *types.AgentTask
    Confidence string
    Notes      []string
}
```

### P0 task mapping

P0 should produce:

```text
task_id: generated or user-provided
contract_id: explicit contract or default
input.goal: original prompt
metadata.source: natural_language
metadata.original_prompt: original prompt
metadata.parser_mode: deterministic
```

### Shell integration

`axis shell` can later dispatch:

```text
ask <prompt>
```

This should call the same parser used by `axis ask`. The shell must not implement a separate parser.

## Safety and Non-Destructive Rules

- Dry-run should be the safest default behavior for generated structure.
- Submission must be explicit via `--submit` or a shell command that clearly states it submits.
- Natural language must not directly execute shell commands.
- Raw prompt, parsed task, and parser mode must remain observable.
- No scheduler or dispatcher behavior changes are required for P0.

## Evolution Path

### P0: Deterministic wrapper

Convert natural language into a default task with provenance metadata.

### P1: Contract suggestion

Suggest contract IDs based on local contract registry and simple rules. Still deterministic and dry-run friendly.

### P2: LLM parser

Use the active project-local model provider to produce structured JSON proposals. Validate before submission and fall back to deterministic mode on failure.

### P3: DAG planner

Convert complex natural language goals into multiple tasks with dependencies. This should only happen after scheduler DAG semantics are stable and observable.

## Relationship to Interactive Shell

The interactive shell remains a command loop. Natural language scheduling is a separate intent compiler that the shell may call. This preserves the original shell scope while making the extension path explicit.
