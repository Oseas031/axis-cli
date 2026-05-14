# Axis Agent Validation Report — 2026-05-14

## Executive Summary

Executed 25+ real tasks against LLMAgentExecutor with MiniMax-M2.7 provider. Validated multi-turn tool loop, error recovery, and self-judgement mechanism. Identified and fixed critical defects in quality judgement system.

---

## Round 1: Baseline (13 tasks)

Simple read/query tasks. All succeeded (excluding 429 rate limits).

| Metric | Value |
|--------|-------|
| Success rate (excl. rate limits) | 100% |
| Avg tool calls | 3.08 |
| Max tool calls | 11 |
| Primary failure mode | MiniMax 429 rate limit |

**Parameter calibration**: MaxIterations=20, MaxErrors=5, TurnTimeout=60s all adequate. No changes needed.

---

## Round 2: Self-Iteration (12 tasks)

Tasks designed to require error recovery and multi-step reasoning.

| Task | Tools | Status | Behavior |
|------|:---:|---|---|
| iter-01 | 3 | ✅ | Linear: read → analyze → write |
| iter-02 | 17 | ✅ | Heavy iteration, multiple strategy changes |
| iter-03 | 3 | ✅ | Write → read-back → verify |
| iter-05 | 3 | ✅ | Error → discover correct path → read |
| iter-07b | 12 | ✅ | **7 failures → strategy adaptation → success** |
| iter-08 | 9 | ✅ | Multi-file navigation |
| iter-09b | 2 | ✅ | Write → execute |
| iter-11 | 20 | ❌ | Budget exhausted |
| iter-12 | 20 | ❌ | Budget exhausted |

### Key Finding: iter-07b

Agent needed `go test` but bash environment lacked `go` in PATH:

```
Turn 1-2: go test → "command not found" (blind retry)
Turn 3: powershell → "command not found"
Turn 4: which go → confirms missing
Turn 5: cmd /c → "command not found"
Turn 6-7: PATH manipulation → still fails
Turn 8: ls (environment exploration)
Turn 10: cmd.exe /c "go test" → SUCCESS
Turn 11-12: write JSON result
```

**Verdict**: Agent demonstrates genuine strategy adaptation (not just retry). But turns 1-2 are identical — runaway detection would save 1 turn.

---

## Round 3: Self-Judgement Validation

### Problem Found

Initial `ToolTraceJudge` was a rubber stamp:
- `recalculate()` averaged only passed items → 1 pass = score 1.0
- 50% error threshold + "text > 10 chars" = mathematically impossible to fail
- `FollowUpHandler` wired but never triggered (dead code)

### Fix Applied

Replaced with `ExecutionJudge` (4 verifiable criteria):

| Criterion | Fails When |
|---|---|
| tool_error_rate | >30% of tool calls errored |
| no_trailing_failures | Last 3 tool calls all failed |
| intent_write_fulfilled | Task says "write/create" but no successful write |
| output_substance | Output is empty or only `<think>` blocks |

Fixed `recalculate()`: score = average of ALL items; passed = ALL criteria must pass.

Added retry-on-failure: judgement fail → inject correction feedback → retry once → re-judge.

### Unit Test Results

```
TestExecutionJudge_PassesGoodResult         PASS
TestExecutionJudge_FailsHighErrorRate       PASS (75% errors → rejected)
TestExecutionJudge_FailsTrailingErrors      PASS (last 3 failed → rejected)
TestExecutionJudge_FailsWriteIntentNotFulfilled  PASS (write task, no write → rejected)
TestExecutionJudge_FailsEmptyOutput         PASS (only <think> → rejected)
TestExecutionJudge_PassesWriteWithBash      PASS (bash echo succeeds → accepted)
```

### End-to-End Verification

| Task | Judgement | Event |
|---|---|---|
| jv-01 (read file) | score=1.00 | `judgement_passed` |
| jv-02 (create file) | score=1.00 | `judgement_passed` |
| jv-03 (write to /root) | score=1.00 | `judgement_passed` (Agent found alternative path) |

---

## Architecture Changes Made Today

| Change | Purpose |
|---|---|
| `ExecutionJudge` (4 criteria) | Replaces rubber-stamp ToolTraceJudge |
| `recalculate()` strict mode | ALL criteria must pass (not just average) |
| `Execute` → `executeOnce` + retry | Judgement failure triggers 1 retry with feedback |
| `PostExecutionJudge` interface | Pluggable judge, decoupled from executor |
| `SetFollowUpHandler` in dispatcher | Orchestrator auto-submits follow-up tasks |
| `JudgementResult.Summary()` | Human-readable failure reason |

---

## Honest Assessment

| Claim | Status |
|---|---|
| Multi-turn tool loop works | ✅ Verified (25+ tasks) |
| Agent adapts strategy on failure | ✅ Verified (iter-07b: 7 failures → success) |
| Self-judgement can reject bad output | ✅ Verified (6 unit tests prove failure paths) |
| Retry-on-judgement-failure works | ✅ Code path verified, not yet triggered in production |
| FollowUpTask generation works | ❌ Handler wired but LLMAgentExecutor never populates field |
| BootstrapOrchestrator loop works | ❌ Not tested |
| AutonomyTransition works | ❌ Not tested |

---

## Remaining Gaps

1. **FollowUpTasks never populated** — LLMAgentExecutor needs to parse Agent output for "next steps"
2. **Runaway detection missing** — Identical consecutive tool outputs not detected
3. **No production judgement_failed event yet** — Need adversarial tasks that truly can't be completed
4. **BashTool PATH issue** — Agent wastes 7 turns discovering `cmd.exe` works; should be in system prompt or PATH config
