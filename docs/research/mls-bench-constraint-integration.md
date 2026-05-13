# MLS-Bench Constraint Integration: Agent Capability Boundaries and Progressive Autonomy

**Date**: 2026-05-13
**Status**: Design Approved
**Source**: arXiv:2605.08678 / https://mls-bench.com
**Principle**: Contract + Permission Ladder + Judge → progressive freedom, not locked-down Agents

---

## Background

MLS-Bench empirically demonstrates three failure modes of current LLM-based Agents:

1. When given both decision and execution authority, Agents prefer brute-force trial-and-error over scientific reasoning (100%)
2. Without enforced process, Agents skip hypothesis-experiment-evidence steps entirely
3. Most common failure: executing wrong operations while believing they succeeded

Axis responds not by locking Agents down, but by requiring **verification before granting more freedom**.

---

## Implementation Classification

### Tier 1: Immediate (Strengthen Existing Mechanisms)

| Constraint | Axis Mechanism | Enhancement |
|---|---|---|
| Compute budget hard cap | SLA timeout/retry | Add token budget + CPU time budget, stage allocation (10%/30%/60%) |
| Max N retries | failure_class strategy | Tighten to max 3 per command, force rollback on exceed |
| Execution result auto-verification | BashTool observable records | Add: exit code + output format + side-effect verification |
| Efficiency measurement | token accounting | Add execution time and resource baseline comparison |

### Tier 2: Optional Contract Templates (Not Hardcoded)

| Constraint | Implementation | Notes |
|---|---|---|
| FSM enforced process | AgentContract type | Hypothesis→Experiment→Evidence→Conclusion; assignable per task |
| Atomic command whitelist | Permission Ladder Level 0 | Unlock more commands as Agent levels up |
| Separation of concerns | Multi-Agent collaboration pattern | Scientist/Planner/Executor as role templates, not sole architecture |
| Structured output per step | Contract output_schema field | Specific Contracts may require JSON format |

### Tier 3: Deferred (Model Capability Will Resolve)

| Constraint | Reason | Axis Alternative |
|---|---|---|
| No self-modification | Violates Sandboxed Evolution | Self-modify in sandbox, promote only after verification |
| Forced JSON every step | Too rigid | Contract already constrains output format |
| Permanent ban on pipes/redirects | Violates "bash is all you need" | Progressive unlock via Permission Ladder |

---

## Four-Dimension Evaluation Framework → Judge Integration

| Dimension | Current Coverage | Gap | Priority |
|---|---|---|---|
| Correctness | Test + Contract validation | Golden standard test suite | P2 |
| Efficiency | Token accounting | Time/resource baseline comparison | P2 |
| **Generalization** | **None** | **Out-of-distribution task set + performance comparison** | **P1** |
| Reliability | None | Multi-run result consistency check | P2 |

### Generalization (P1)

```
Score = OOD_correctness / ID_correctness

≥0.8 → eligible for autonomy upgrade
0.5-0.8 → maintain current level
<0.5 → downgrade (Agent is memorizing, not understanding)
```

### Reliability (P2)

- Run critical tasks 3-5 times
- Consistency ≥95% required
- Inconsistent tasks marked unreliable, excluded from autonomy evaluation

---

## Autonomy Transition Upgrade Criteria

### Current Problem

Existing rule: N consecutive successes → upgrade. Too coarse.

### New Standard (MLS-Bench Golden Standard)

Upgrade requires ALL of:

| Condition | Meaning | Detection |
|---|---|---|
| Effective under resource constraints | Performance improves within fixed token/time budget | Compare historical performance at same budget |
| Effective out-of-distribution | Maintains quality on new task types | Generalization score ≥0.8 |
| Architecture-driven | Improvement from capability, not resource scaling | No auto-upgrade from context increase or model swap |

### Disqualifiers

- Performance drops when context budget is held constant → no upgrade
- Autonomy level binds to architectural capability, not provider capability
- Generalization <0.5 → downgrade

---

## Feature Progressive Unlock Roadmap

| Axis Feature | Level 0 (Current) | Level 1 | Level 2 (Target) |
|---|---|---|---|
| BashTool | Whitelisted commands | Whitelist + limited pipes | Full bash |
| Sandboxed Evolution | Modify specified files only | Any file in sandbox | Propose architectural changes |
| Tool permissions | Read + execute | Read-write (Contract-authorized) | Full tool access |
| Task generation | Sub-tasks only | Sibling tasks | Follow-up chains |
| Contract customization | Predefined Contracts only | Propose Contract modifications | Create new Contracts |
| Self-judgement | External Judge verification | Self-judge + external spot-check | Self-judge primary |

### Upgrade Trigger

```
Level N → Level N+1 requires:
├── 30 consecutive tasks at current Level
├── Correctness ≥90%
├── Generalization ≥0.8
├── Reliability ≥95%
└── Zero security violations
```

---

## Alignment with Axis Design Philosophy

| Axis Principle | MLS-Bench Compromise | End State |
|---|---|---|
| Zero Control | FSM enforced process | Agent internalizes scientific method; FSM becomes optional Contract |
| Bash is all you need | Atomic command whitelist | Permission Ladder progressively unlocks full bash |
| Competence earns autonomy | Separation of concerns | Single Agent earns all permissions through verification |
| Contract is Structure | Forced JSON per step | Contract defines output constraints, not global enforcement |
| Controllable Evolution | No self-modification | Sandboxed Evolution allows self-modification with verification gate |
