# Agent Evaluation Principles

> Derived from SWE-bench / MLE-bench / Terminal-Bench methodology, adapted for Axis.

## Core Axiom

**Tests decide, not judges.**

An LLM judging an LLM's output has unresolvable same-source bias. The only authoritative verification is a deterministic oracle: compiler, test suite, exit code, file diff. If you can't express "pass/fail" without an LLM, the task is not ready for automated evaluation.

## Five Principles

### 1. Ground Truth Must Pre-Exist

Every evaluation task must have a known-correct answer that exists before the Agent runs.

| Good | Bad |
|---|---|
| "Fix this failing test" (test is the oracle) | "Write good code" (no oracle) |
| "Produce output matching X" (diff is oracle) | "Explain this code" (needs LLM judge) |
| "Make `go build` pass" (compiler is oracle) | "Improve performance" (needs measurement + threshold) |

### 2. Oracle Must Be Deterministic

The pass/fail decision must produce the same result on every run given the same input.

Acceptable oracles:
- `go build ./...` exit code = 0
- `go test ./... -count=1` all pass
- File exists at expected path with expected content
- Command output matches expected string (exact or regex)
- JSON output validates against schema

Unacceptable oracles:
- "Output looks reasonable"
- LLM rates quality > 7/10
- Human reviewer approves

### 3. Environment Must Be Isolated and Reproducible

- Docker container with pinned image (by digest, not tag)
- No network access during evaluation
- Project mounted read-only; Agent writes to tmpfs
- Deterministic seed for any randomness
- Same container image = same result

### 4. Tasks Must Come From Real History

The strongest tasks are derived from actual changes that were made and verified:

```
Source: git log --oneline
Task: "Given the codebase at commit N-1, produce a patch that makes the tests at commit N pass"
Oracle: go test ./... at commit N
```

This eliminates the "task designer bias" problem — you're not designing tasks the Agent can solve, you're replaying tasks that were actually solved.

### 5. Metrics Must Be Binary + Aggregated

Per-task: **resolved** (boolean). No partial credit.

Aggregate:
- **Resolve rate** = resolved / total
- **Stratified by difficulty** (lines changed, files touched, test count)
- **Cost per resolve** = total tokens / resolved count

No "score out of 10". No "mostly correct". Either the tests pass or they don't.

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Fix |
|---|---|---|
| LLM-as-judge for code quality | Same-source bias, non-deterministic | Use compiler + tests |
| Self-designed easy tasks | Confirms capability you already know exists | Use real git history |
| Partial credit scoring | Masks failure as "almost success" | Binary pass/fail |
| Evaluating without isolation | Environment differences cause flaky results | Docker + pinned deps |
| Single-run evaluation | Variance from LLM sampling | 3 runs, report median |
| Judging process instead of outcome | "Agent used 5 tools" means nothing | Only outcome matters |

## Application to Axis Self-Iteration Evaluation

To verify that Axis's self-iteration mechanism provides value beyond a bare for-loop:

1. **Baseline**: Run task with `MaxIterations=1` (single-shot, no iteration)
2. **Treatment**: Run same task with `MaxIterations=20` (full iteration budget)
3. **Metric**: Resolve rate difference between baseline and treatment
4. **Oracle**: `go test` / `go build` / file existence check

If treatment resolve rate > baseline resolve rate, iteration adds value.
If treatment ≈ baseline, iteration is just burning tokens.

## Task Difficulty Tiers

| Tier | Characteristics | Expected Resolve Rate |
|---|---|---|
| T1 | Single file, <10 lines changed, 1 test | >80% |
| T2 | 1-3 files, 10-50 lines, multiple tests | 40-70% |
| T3 | 3+ files, 50+ lines, cross-package deps | 10-30% |
| T4 | Architecture change, new package, design decision | <10% |

Calibrate tiers against real data. If T1 resolve rate < 80%, the Agent has fundamental capability gaps. If T3 > 50%, the Agent is genuinely useful for real work.
