---
type: architecture
status: active
created: 2026-05-11
last_verified: 2026-05-14
related:
  - agent-native-first-principles.md
  - module-and-naming-conventions.md
---

# Semantic Boundaries

> 展开自 CLAUDE.md §4（目录边界）+ §6（语义边界）

## Purpose

> 本体论："Interface is existence"。语义边界定义的不是"谁负责什么"，而是"什么通过什么接口存在"。越界 = 让一个概念通过不属于它的接口存在 = 本体论错误。

This document defines what each Axis concept is allowed to mean and do. The boundary table is a **否定表**（逻辑学："否定是核心动作"）——每个概念通过"Must Not Own"获得其 Determinateness。

## Boundary Table

| Concept | Owns | Must Not Own |
|---|---|---|
| `AgentTask` | task identity, contract ID, input, dependencies, status, metadata | execution logic, provider selection policy, hidden permissions |
| `AgentContract` | input/output shape and execution contract | scheduler policy, provider credentials, context retrieval |
| Scheduler | task readiness and status transitions | model calls, shell execution, provider config, natural language parsing |
| Orchestrator | module coordination and execution loop | provider profile storage, CLI rendering, hidden policy decisions |
| Dispatcher | route accepted tasks to execution path | admission policy, provider credential management |
| Provider | model request/response | task lifecycle, scheduler state, credentials persistence |
| ProviderConfig | project-local provider profile state | task execution, scheduler policy |
| Tool | bounded capability execution | global permission system, task scheduling |
| Skills (Loader) | on-demand knowledge discovery and loading from `.axis/skills/` | automatic prompt injection, scheduler/contract modification, background work, network access |
| Intent Parser | user intent to structured Axis object | direct execution, context assembly, permission escalation |
| ContextBundle | task-specific readiness context | authority escalation, scheduler changes, task submission |
| EvolutionRun | isolated system-change proposal, atomic steps, verification evidence, promotion/discard decision | implicit main-tree mutation, hidden execution policy, automatic authority escalation |
| axis-up | external usability helper | internal imports, source mutation, core architecture |

## Core Rules

- Natural language produces structure; it does not execute by itself.
- Context improves action quality; it does not control or authorize action.
- Metadata adds audit and hints; it does not silently change core semantics.
- Providers answer model requests; they do not manage Axis state.
- CLI presents and invokes behavior; it should not become the domain layer.
- Draft evolution is not main system state.
- Verification is evidence for promotion; it is not promotion by itself.
- Discarded evolution remains audit history unless explicitly deleted by a documented cleanup path.
- Kernel syscalls are fixed primitives; their semantics cannot be overridden by Agents or userland tools.
- Userland tools extend Agent capabilities; they do not modify kernel scheduling, isolation, or state semantics.
- Capability granting flows from Kernel to Agent; Agents cannot self-escalate without Kernel verification.

## Boundary Violation Signals

A change needs review if it:

- lets `intent` submit tasks without an explicit CLI/orchestrator path
- lets context packets expand tool/file/network access
- lets providers mutate scheduler state
- stores secrets outside provider config boundaries
- makes shell commands behave differently from CLI commands without a spec
- turns metadata into hidden policy enforcement
- lets an evolution workspace mutate the main tree before explicit promotion
- lets a userland tool modify syscall semantics (scheduling, isolation, state transitions)
- lets an Agent bypass `request_capability` to access tools directly
- treats a successful verification command as automatic promotion

## Rule of Thumb

```text
If a module's name does not explain why it is allowed to perform an action, the action likely belongs elsewhere.
```

## 演化声明

> 历史观："没有最终架构"。本文档的边界表反映当前矛盾的最优解。当模块职责因实践需要发生变化时，先修改本文档（L2），再修改代码（L3）。边界变更 = 架构决策，必须经过 Phase I→II→III。
