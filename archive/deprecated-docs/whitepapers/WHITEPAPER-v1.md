# Axis Whitepaper

## Abstract

Axis is an Agent-native scheduling system. Its long-term goal is not to provide a stronger task queue, but to construct an execution substrate for Agent autogenesis: enabling Agents to accumulate competence through task practice, earn autonomy, and ultimately gain the ability to understand, modify, validate, judge, and re-authorize themselves.

Axis is currently in an early stage: Milestone 1 is complete, and Milestone 2 is building DAG parallel scheduling, contract admission, SLA, and error codes. These are not isolated features, but foundational organs for the future Autogenesis Loop.

## Core Philosophy

### More Context, More Action, Zero Control, Controllable Evolution

Axis does not gain security by reducing Agent capabilities, but by providing queryable context, more action capabilities, contract boundaries, permission ladders, and controllable evolution, enabling Agents to take on more complex tasks.

### bash is all you need, simple but robust, composable and extensible

Axis defaults to shell-native. CLI, scripts, CI, and Agents should all be able to call Axis directly without relying on heavy UIs. The interaction surface must remain simple, reliable, composable, and extensible.

### Competence earns autonomy, autonomy matches responsibility, evolution is controllable

Axis's permission system is not a static whitelist, but a progressive autonomy mechanism. Agents earn a larger action radius through reliable performance; failures, incorrect modifications, and validation failures lead to autonomy contraction. Autonomy must match risk responsibility, and permission elevation, self-evolution, and contract rewriting must pass verification.

### Scaffold-to-Self

Axis's current workflow, contract, permission rule, and spec are all transitional structures:

- workflow is temporary scaffolding
- contract is a growth boundary
- permission rule is a progressive autonomy mechanism
- spec is a seed

Their mission is not to permanently control the Agent, but to help the Agent eventually internalize, rewrite, and supersede these external structures.

## Bootstrap Definition

Axis's bootstrap is not simple code self-modification.

True bootstrap is:

```text
Agent receives thought
  -> turns thought into structure
  -> executes structure
  -> reflects on result
  -> revises its own structure
  -> earns more autonomy
  -> repeats
```

The bootstrap starting point has already occurred: external Agents are injecting ideas into Axis that can be solidified, executed, reflected upon, and evolved.

## Autogenesis Loop

Axis's long-term loop is:

```text
Perceive self
  -> Diagnose self
  -> Redefine self
  -> Modify self
  -> Validate self
  -> Judge self
  -> Re-authorize self
  -> Repeat
```

Corresponding future engineering objects:

- SelfContext
- SelfDiagnosis
- SelfRedefinition
- SelfModification
- SelfValidation
- SelfJudgement
- AutonomyTransition
- Follow-up Task Generation

## Current Architecture

```text
AgentTask
  -> admission
  -> scheduler
  -> orchestrator
  -> dispatcher
  -> executor
  -> state store
```

Milestone 1 has achieved the basic scheduling closed loop. Milestone 2 expands it into a minimal DAG parallel scheduling system.

## Milestone Roadmap

### Milestone 1: Basic Scheduling

Completed:

- FIFO scheduling
- Simple dependency management
- Input/output validation
- Basic state storage
- Basic CLI

### Milestone 2: Autogenesis Execution Substrate

In progress:

- ready-set DAG scheduling
- contract admission
- SLA timeout / retry
- parallel orchestrator
- stable error codes
- CLI/docs acceptance

### Bootstrap Loop

Follow-up specs:

- self-iteration contracts
- MockAgentExecutor
- validation result model
- follow-up task generation
- mock self-iteration DAG

### Autogenesis Loop

Longer-term specs:

- SelfContext
- SelfDiagnosis
- SelfJudgement
- AutonomyTransition
- tool self-generation
- self-authored specs / contracts / workflows

## Non-Goals

Not in current stage:

- Web UI
- Complex TUI
- External databases
- Real LLM SDK binding
- Distributed workers
- Global event bus

## Summary

Axis is not just a system for scheduling Agents. Axis is a system that enables Agents to accumulate competence through task practice, earn autonomy, and ultimately generate themselves.
