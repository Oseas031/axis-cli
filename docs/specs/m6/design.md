# M6 Design: Self-Judgement (Self-Validation)

**Status**: In Progress
**Last Updated**: 2026-05-10

## 1. Architecture Overview

```
                    ┌─────────────────────────────────────┐
                    │  SelfJudgementEngine                 │
                    │  (自评判引擎)                        │
                    └──────────────┬──────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
     ┌────────▼────────┐  ┌───────▼────────┐  ┌───────▼────────┐
     │ ValidationStrategy│  │ ModelProvider  │  │  ContractExec  │
     │ (多种策略)        │  │ (用于语义评判)  │  │  (用于契约验证) │
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

检查代码语法正确性：
- Go fmt 检查
- Go vet 检查
- AST 解析验证

### 4.2 SemanticValidationStrategy

使用 LLM 进行语义验证：
- 调用 ModelProvider 分析代码语义
- 检查逻辑一致性
- 评估代码质量

### 4.3 ContractValidationStrategy

验证契约输出：
- 输入验证 (ValidateInput)
- 输出验证 (ValidateOutput)
- Schema 合规性检查

### 4.4 TestValidationStrategy

验证测试结果：
- 测试通过率
- 失败测试分析
- 测试覆盖完整性

### 4.5 CoverageValidationStrategy

验证覆盖率：
- 语句覆盖率
- 分支覆盖率
- 与阈值比较

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
                    (基于 judgement 结果决定是否升级自主权)
```

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

Judgement confidence 基于：
- 各策略结果的权重平均
- 策略执行的一致性
- 输入数据的完整性

```go
func calculateConfidence(judgements []JudgementItem) float64 {
    // Confidence = weighted average of individual confidences
    // Higher when strategies agree, lower when they conflict
}
```
