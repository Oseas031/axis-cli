# Agent Design First Principles — Compensating LLM Deficiencies

> 展开自 CLAUDE.md §11（演化原则）+ §0（辩证方法论）

**Status**: Draft (2026-05-14)
**Core thesis**: Design the Agent around what LLM *structurally cannot do*, not around what it *can do*.

---

## 1. The Distinction

```
Industry approach:  LLM → (prompt engineering) → better output
                    Ceiling = LLM capability upper bound

Axis approach:      LLM → (structural deficiencies) → compensation mechanisms → systematic output
                    Ceiling = compensation mechanism design quality
```

LLM is not "the coder". LLM is the **intent translator**: it converts fuzzy human intent into structured action declarations. The system's deterministic components execute, verify, and correct.

---

## 2. LLM Deficiency Classification

Honest classification — not all are "structural":

### Architectural deficiencies (deep, slow to resolve)

| # | Deficiency | Why architectural | Compensation |
|---|---|---|---|
| D1 | **No execution verification** | Model generates tokens, cannot run code. Output correctness is unknowable to the model at generation time. | Forced execute-verify loop |
| D2 | **Unreliable causal reasoning** | Statistical correlation ≠ causation. Model knows "if usually follows condition" but not "why this condition must be this way". CoT narrows but doesn't close the gap. | Externalized causal signals |

### Current-form constraints (improving with model generations)

| # | Deficiency | Why not architectural | Compensation | Retirement condition |
|---|---|---|---|---|
| D3 | **No persistent state** | API design choice, not architecture limit. Models can have state; products choose stateless APIs. | External memory system | When models natively maintain cross-session project state |
| D4 | **Weak incremental editing** | Degree problem, rapidly improving. Models already do reasonable diffs. | Structured edit primitives | When model diff accuracy > 95% on real codebases |
| D5 | **Poor self-boundary awareness** | Calibration problem, improvable via RLHF. May have theoretical ceiling. | Externalized competence boundary | When model refusal accuracy > 90% on out-of-distribution tasks |

**Investment principle**: Heavy infrastructure for D1-D2. Lightweight wrappers for D3-D5 that degrade gracefully as models improve.

---

## 3. Compensation Mechanisms

### 3.1 Forced Execute-Verify Loop (compensates D1)

Every LLM output that claims to be code must pass through objective verification before being accepted.

**Layered verification** (not single gate):
- L0: Compilation (seconds, automatic, always)
- L1: Test suite (minutes, automatic, when tests exist)
- L2: Contract conformance (automatic, when contract defined)
- L3: Design review (requires higher-level judgement, may need human)

**Incompleteness acknowledgment**: When no contract/test exists, system explicitly marks output as "unverified beyond compilation". Does not pretend L0 = correctness.

**Convergence guarantee**: Diagnosis loop has iteration budget (`maxIterations`). Oscillation detection: if fix-A-breaks-B-fix-B-breaks-A pattern detected, escalate rather than continue.

### 3.2 Externalized Causal Signals (compensates D2)

Not a perfect causal graph (impossible to maintain). Multi-layer approximation:

- **Static**: import graph, type dependencies, module boundaries
- **Dynamic**: test failure signals ("changing A broke test for B")
- **Historical**: "last time A was modified, B also needed updating" (from event log)
- **Proactive association**: when Agent touches file X, system surfaces X's BOUNDARY.md, design decisions, historical change reasons — structured index, not raw text dump

### 3.3 External Memory (compensates D3, transitional)

Agent queries context on demand rather than receiving passive injection. But pure on-demand has an unknown-unknowns problem.

**Hybrid approach**:
- Agent declares what it needs (active query)
- System proactively surfaces structurally-related context (passive association)
- Neither is "push all context into prompt" — both are targeted

**Retirement path**: As models gain native persistent state, external memory simplifies to a sync layer.

### 3.4 Structured Edit Primitives (compensates D4, transitional)

LLM declares edit intent; system executes precise modification. But:

**Transactional operations**: For refactoring tasks, allow intermediate verification failures. Validate only the final state (staged evolution model).

**Retirement path**: As model diff accuracy improves, allow direct diff output with verification.

### 3.5 Externalized Competence Boundary (compensates D5, transitional)

System tracks what Agent can/cannot reliably do. Agent doesn't self-assess.

- Contract admission: reject tasks outside proven capability
- Autonomy levels: wider action radius earned through demonstrated reliability
- Competence scoring: based on historical success rate, not self-report

**Retirement path**: As model calibration improves, allow self-assessment with audit.

---

## 4. Execution Architecture

Not waterfall. **Progressive spiral**:

```
task
  → [intent parsing (LLM)] → rough action plan (will be wrong)
    → [boundary check (system)] → within capability?
      → no: reject or decompose
      → yes:
        → [execute (Tool)] → deterministic operation
          → [verify (system)] → layered check
            → pass: next step or complete
            → fail:
              → [diagnose (LLM)] → what went wrong?
                → [refine plan (LLM)] → corrected action plan
                  → back to execute (with iteration budget)
```

**Key insight**: LLM appears at three points:
1. **Start**: understand intent (will be imperfect)
2. **Failure**: diagnose what went wrong (pattern matching on error signals)
3. **Refinement**: adjust plan based on diagnosis

Everything between is deterministic system behavior.

### Exploration mode

When task is marked exploratory (unknown target):
- Relaxed verification constraints
- Multiple candidate generation allowed
- Human or higher-level judgement selects
- Maps to staged evolution's multi-branch explore + explicit promote/discard

---

## 5. The Contract Generation Problem

**Acknowledged incompleteness**: Verification depends on contracts. Contracts must be generated. This is self-referential.

Resolution (not solution — there is no complete solution):
1. When contract exists: full verification loop
2. When contract absent: degrade to compilation-only verification + explicit "unverified" marking
3. Contract generation itself: LLM generates draft → human approves (or higher-autonomy Agent with proven contract-writing track record)
4. Never pretend absence of contract = absence of bugs

---

## 6. LLM's Role — Precisely Defined

| Capability | Who | Why |
|---|---|---|
| Understand intent | LLM | Language understanding is core LLM strength |
| Generate candidates | LLM | Pattern matching produces "possible approaches" |
| Select approach | System (Contract + Judgement) | Needs objective criteria, cannot self-evaluate |
| Execute precisely | System (Tool + Bash) | Deterministic operations, no hallucination allowed |
| Verify results | System (Compiler + Test) | Needs ground truth |
| Remember history | System (Memory + EventLog) | Needs persistent state |
| Know boundaries | System (Autonomy + Admission) | LLM doesn't know what it doesn't know |
| Diagnose failures | LLM | Pattern matching on error signals |
| Explore alternatives | LLM | Divergent generation of options |

---

## 7. Transitional Structure Declaration

This entire framework is a transitional structure (CLAUDE.md §14 principle).

**Retirement conditions**:
- D1 compensation retires when: models can natively execute and verify code in-context
- D2 compensation retires when: models demonstrate reliable causal reasoning on novel codebases
- D3-D5 compensations retire when: their respective retirement conditions (§2 table) are met
- The framework itself retires when: LLM capability makes the "compensation" framing obsolete

**Anti-ossification rule**: Every 6 months (or when a new model generation ships), re-evaluate which deficiencies have moved from "architectural" to "current-form constraint" to "resolved". Demote or remove compensations accordingly.

---

## 8. What This Framework Is NOT

- NOT "LLM is bad at coding" — LLM is excellent at intent understanding and pattern generation
- NOT "never let LLM write code" — LLM generates candidates; system verifies
- NOT "permanent infrastructure" — all compensations are transitional
- NOT "first principles from Transformer math" — honest engineering framework based on observed deficiency patterns, designed to degrade gracefully as models improve
- NOT "complete" — acknowledges contract generation incompleteness and exploration mode as open problems
