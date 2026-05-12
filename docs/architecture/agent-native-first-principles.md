# First Principles of Agent-Native Scheduling

**Nature**: Foundational architectural rationale (immutable principle layer) — **read before coding**
**Related**: `reports/strategy/agent-native-scenario-principles-analysis-2026-05-11.md` (full scenario analysis and defect diagnosis)

> This document is the foundation of all Axis design decisions. The former `agent-native-design-philosophy.md` has been merged here and deprecated.

**[Chinese version / 中文版](../../docs/zh/architecture/agent-native-first-principles.md)**

---

## Core Thesis

The underlying nature of an Agent-native scheduling system is not "a smarter Cron + BPMN", but a **lifecycle governance system for autonomous computational entities**.

More precisely: Axis is an **operating system for Agents**. It solves the problem of how to let a non-deterministic, autonomous entity exist safely, observably, and evolvably in a complex world.

Like Linux provides processes with scheduling, memory, filesystem, and IPC — Axis provides Agents with scheduling, context, capabilities, and isolation. But unlike Linux (which manages deterministic processes), Axis manages entities whose next action cannot be predicted. This demands a fundamentally different relationship: not control, but substrate.

See `docs/architecture/kernel-abstraction-model.md` for the structural expression of this thesis.

---

## Six First Principles

### 1. Interface is Existence

> An entity is the set of interfaces it exposes. Humans and Agents implement the same agent interface, with no identity bias. All interface calls must leave observable logs.

- `axis ask` = human injecting intent into the system via CLI interface
- `axis status` = querying Agent task interface state
- `.axis/events/tasks.jsonl` = append-only interface call log of Agent behavior
- To the system, "who you are" doesn't matter; "what interfaces you called and what effects you produced" does

### 2. Query is Context

> Context is not a "parameter package" assembled by the system and pushed to Agents, but a shared reality that Agents actively query and construct.

- `contextpack` is a "context economy system": budgeted attention, recorded exclusions, traceable sources
- Agents should be able to declare their own context needs rather than passively accepting system assembly
- ReadinessArtifact's `source_digest` ensures context version consistency and supports post-hoc audit reproduction

### 3. Ladder is Boundary

> Permissions/autonomy are competence-driven and risk-driven, not identity-driven. Historical performance and task requirements determine the dynamic execution ladder.

- New Agents receive minimum permissions (fenced)
- High-performing Agents progressively gain more tool permissions (ladder ascent)
- Frequently failing Agents have permissions contracted (ladder descent)
- High-risk operations require secondary confirmation, regardless of whether the executor is human or Agent

### 4. Layered Isolation is Collaboration

> Each task/Actor receives an isolated workspace; collaboration occurs through shared event logs and version control; isolation granularity adapts to task complexity and Agent capability.

- Local control plane = the "town hall" of the Agent ecosystem, coordinating but not replacing autonomous execution
- Sandboxed evolution = experiment in isolated space, promote to mainline only after verification
- Collaboration is not "shared memory" but "shared immutable history"

### 5. Contract is Structure

> File system / meta-files are the shared contract language for all Agents. Contracts constrain all Actors equally, can be rewritten by capable Agents after verification and consensus, with full audit trail.

- requirements/design/tasks under `docs/specs/` are functional contracts
- `.axis/providers.json` is the model routing contract
- `.axis/runtime.json` is the runtime locator contract
- Contract changes must go through the sandboxed evolution protocol: experiment → verify → promote

### 6. Capability is Decision Right

> Ultimate decision rights belong to the agent that has demonstrated the best capability for a specific task. Humans can take over at any time, but takeover behavior is recorded and updates capability assessment.

- The scheduler assigns tasks based on Agent historical performance and capability profiles
- Human takeover is not "failure" but "capability assessment data point"
- "Competence earns autonomy, autonomy matches responsibility, evolution is controllable"

---

## Core Assertions

> **More Context, More Action, Zero Control, Controllable Evolution**

- **More Context**: The more context an Agent has, the better decisions it can make. The system's responsibility is to provide queryable, budgetable, auditable context, not to control Agent behavior.
- **More Action**: As context increases and reliability is proven, Agents should gain a wider action radius (more tools, more permissions, fewer approvals).
- **Zero Control**: The system does not control the Agent's "thinking process"; it only defines interface boundaries, records behavior, and enforces minimum permissions. Control is boundary, not intervention.
- **Controllable Evolution**: Agent capabilities can evolve, but evolution must occur in sandboxes, must be verified, and humans retain the final promote/discard decision right.

---

## Key Strategic Rejections

| Rejection | Reason |
|---|---|
| Workflow canvas (drag-and-drop DAG editor) | Contracts and event logs themselves define the process |
| Universal Agent (one Agent does everything) | Use contextpack + provider route to enable multiple specialized Agents to collaborate |
| Black-box AI (unobservable behavior) | All behavior written to append-only event log, all context can be inspect/preflight |
| Static permissions (unchanged after manual config) | Permissions dynamically adjusted based on performance, but promotion must go through sandbox verification |
| Web/TUI first | CLI native, composable, scriptable — "bash is all you need" |

---

## Evolution Layers

```
P0 (Local/Conservative)    P1 (Enhanced)              P2 (Distributed)           P3 (Autonomous)
────────────────────────────────────────────────────────────────────────────────────────────────
NL intent compilation   →  Intent quality scoring    →  Multimodal intent       →  Intent prediction
Local control plane     →  Remote/federated nodes    →  Federated cluster       →  Global Agent network
Context assembly        →  Context quality eval      →  RAG fusion             →  Active context query
Sandboxed evolution     →  Auto verification gen     →  Multi-candidate evolve  →  Autonomous evolution
Tool registry           →  Tool usage learning       →  Tool combo discovery   →  New tool invention
Event log               →  Long-term storage/query   →  Behavior pattern mining →  Organizational intelligence
```

---

## Interaction Principles

**bash is all you need, simple but robust, composable and extensible**

- **Shell-native**: CLI first, scriptable, composable, callable by humans, CI, and Agents
- **Simple but robust**: Reject redundant Web UI or complex TUI while providing necessary fault tolerance, confirmation, rollback, and observability
- **Composable and extensible**: Interfaces support multi-dimensional composition with reserved extension points; Axis itself can be directly called, orchestrated, and adapted by Agents

See [Bash is All You Need](bash-is-all-you-need.md) for details.

---

## Traditional Scheduling vs. Axis

| Dimension | Traditional Scheduling | Axis |
| --- | --- | --- |
| Entity model | Humans control tools | Both humans and Agents are intelligent entities |
| Interface | Identity-differentiated | Identity-agnostic abstraction |
| Context | Platform-pushed or statically injected | Agent-initiated queries |
| Action | Predefined limited operations | Composable, verifiable, self-generating |
| Permissions | Static identity-based authorization | Competence ladder authorization |
| Collaboration | Shared workspace or central control | Sandbox isolation + event log + version merge |
| Contract | Fixed rules | Verifiable, evolvable structures |
| Decision rights | Humans default to final arbitration | Capability determines decision rights |
| Evolution | Externally planned upgrades | Controllable bootstrapping and self-modification |

---

## Bootstrap Origin

Axis's bootstrapping begins with external Agents injecting ideas that can be absorbed, solidified, and evolved by the system:

- Humans provide philosophical viewpoints, directional tension, and value judgments
- Agents take sovereign responsibility at the design level
- Agents transform ideas into specs, workflows, contracts, permissions, architecture, and implementation paths
- Axis transitions from being externally designed to participating in designing itself through Agents

Early workflows, contracts, permission rules, and specs are all transitional scaffolding. Their mission is not to permanently control Agents, but to help Agents accumulate competence, earn autonomy, and ultimately internalize external structures into their own action structures.

---

## Risk Boundaries

Axis does not pursue unbounded autonomy. Autonomy must satisfy:

- Behavior is observable
- Decisions are traceable
- Permissions can be contracted
- Contracts are verifiable
- Evolution is rollback-safe
- High-risk actions require secondary confirmation

Zero Control does not mean no constraints. It means the system does not prescribe a single action path for intelligent entities; boundaries are jointly formed by contracts, competence ladders, isolation layers, audit logs, and controllable evolution mechanisms.

---

## Conclusion

Axis's core is not about controlling Agents, but about giving Agents more context, more action capabilities, and controllable evolution space within observable, verifiable, and rollback-safe boundaries.
