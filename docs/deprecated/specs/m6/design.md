# M6 Design: Self-Judgement (Self-Validation)

**Status**: Completed
**Last Updated**: 2026-05-10

## 1. Architecture Overview

```
                    ┌─────────────────────────────────────┐
                    │  SelfJudgementEngine                 │
                    │  (self-judgement engine)              │
                    └──────────────┬──────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
     ┌────────▼────────┐  ┌───────▼────────┐  ┌───────▼────────┐
     │ ValidationStrategy│  │ ModelProvider  │  │  ContractExec  │
     │ (multiple         │  │ (for semantic  │  │  (for contract │
     │  strategies)      │  │  judgement)    │  │   validation)  │
     └─────────────────┘  └────────────────┘  └─────────────────┘
```

## 2. Core Components

### 2.1 JudgementCriteria

```go
type JudgementCriteria struct {
    Name        string
    Type        JudgementType
    Weight      float64
    Thresholds  map[string]float64
    Enabled     bool
}
```

### 2.2 SelfJudgementEngine

```go
type SelfJudgementEngine struct {
    strategies map[JudgementType]ValidationStrategy
    logger     func(string, ...interface{})
}

func (e *SelfJudgementEngine) Judge(
    result *AgentExecutionResult,
    criteria []JudgementCriteria,
) (*JudgementResult, error)

func (e *SelfJudgementEngine) RegisterStrategy(
    judgementType JudgementType,
    strategy ValidationStrategy,
)
```

### 2.3 ValidationStrategy Interface

```go
type ValidationStrategy interface {
    Validate(
        input any,
        criteria JudgementCriteria,
    ) (*JudgementItem, error)
    CanHandle(criteria JudgementCriteria) bool
}
```

## 3. Judgement Flow

```
AgentExecutionResult
    │
    ▼
SelfJudgementEngine.Judge()
    │
    ├─► SyntaxValidationStrategy  ──► JudgementItem
    │
    ├─► SemanticValidationStrategy ─► JudgementItem
    │        │
    │        └─► ModelProvider (LLM)
    │
    ├─► ContractValidationStrategy ─► JudgementItem
    │        │
    │        └─► ContractExecutor
    │
    ├─► TestValidationStrategy ─► JudgementItem
    │
    └─► CoverageValidationStrategy ─► JudgementItem

    │
    ▼
JudgementResult (aggregated)
```

## 4. Strategy Implementations

### 4.1 SyntaxValidationStrategy

Checks code syntax correctness:
- Go fmt check
- Go vet check
- AST parse validation

### 4.2 SemanticValidationStrategy

Uses LLM for semantic validation:
- Calls ModelProvider to analyze code semantics
- Checks logical consistency
- Evaluates code quality

### 4.3 ContractValidationStrategy

Validates contract output:
- Input validation (ValidateInput)
- Output validation (ValidateOutput)
- Schema compliance check

### 4.4 TestValidationStrategy

Validates test results:
- Test pass rate
- Failed test analysis
- Test coverage completeness

### 4.5 CoverageValidationStrategy

Validates coverage:
- Statement coverage
- Branch coverage
- Comparison against thresholds

## 5. Integration with Bootstrap Loop

```
analyze-change-request
    → implement-change
        → run-validation
            → update-docs
                → review-result
                    → spawn-followup-tasks
                            │
                            ▼
                    self/judge-execution
                            │
                            ▼
                    (decide whether to upgrade autonomy based on judgement result)
```

### 5.1 BootstrapOrchestrator Judgement API

`BootstrapOrchestrator` injects `judgement.Engine` via `WithJudgementEngine` option:

```go
engine := judgement.NewEngine()
bo := NewBootstrapOrchestrator(scheduler, maxIterations, WithJudgementEngine(engine))
```

Key methods:

- `JudgeExecutionResult(result *AgentExecutionResult) (*JudgementResult, error)` — Performs self-judgement on execution results; returns nil when engine is not configured.
- `CalculateAutonomyDelta(jr *JudgementResult) AutonomyDelta` — Calculates autonomy adjustment based on judgement results:
  - score >= 0.95 && confidence >= 0.90 → Delta +2 (excellent)
  - passed → Delta +1 (pass)
  - !passed && score >= 0.50 → Delta 0 (borderline)
  - !passed && score < 0.50 → Delta -1 (failure)
- `EvaluateAndDecide(result *AgentExecutionResult) error` — End-to-end judgement that writes `result.JudgementResult` and `result.AutonomyDelta` in place.

Default judgement strategy combination (`defaultJudgementCriteria`):

| Strategy | Weight | Threshold |
|----------|--------|-----------|
| syntax | 0.20 | min_pass_rate = 1.0 |
| test | 0.40 | min_pass_rate = 0.90 |
| coverage | 0.40 | min_coverage = 0.85 |

### 5.2 CLI Display

Standalone diagnostic command:

```bash
axis judge                    # Run default judgement and display results
axis shell> judge             # Same diagnostic within shell
```

Output includes Passed / Score / Confidence summary and per-strategy details.

## 6. Judgement Contract Schema

```go
// Input Schema
{
    "execution_result": {
        "output": map[string]any,
        "validation_result": {
            "tests_passed": int,
            "tests_failed": int,
            "coverage": float64,
            "is_acceptable": bool
        },
        "autonomy_delta": {
            "delta": int,
            "reason": string
        }
    },
    "criteria": [
        {
            "name": string,
            "type": string,  // syntax|semantic|contract|test|coverage
            "weight": float64,
            "thresholds": map[string]float64
        }
    ],
    "context": {
        "task_id": string,
        "self_context": SelfContext
    }
}

// Output Schema
{
    "judgement": {
        "passed": bool,
        "score": float64,
        "judgements": [
            {
                "criteria_name": string,
                "passed": bool,
                "score": float64,
                "details": string
            }
        ],
        "confidence": float64,
        "suggested_fixes": [string]
    }
}
```

## 7. Judgement Thresholds

```go
const (
    DefaultMinCoverage     = 0.85  // 85% minimum coverage
    DefaultMinTestPassRate = 0.90  // 90% test pass rate
    DefaultMinConfidence    = 0.70  // 70% judgement confidence
    DefaultPassingScore     = 0.75  // 75% total score to pass
)
```

## 8. Confidence Calculation

Judgement confidence is based on:
- Weighted average of individual strategy results
- Consistency of strategy execution
- Completeness of input data

```go
func calculateConfidence(judgements []JudgementItem) float64 {
    // Confidence = weighted average of individual confidences
    // Higher when strategies agree, lower when they conflict
}
```
