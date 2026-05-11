# M6 Requirements: Self-Judgement (Self-Validation)

**Status**: Completed
**Last Updated**: 2026-05-10

## 1. Overview

M6 实现 Self-Judgement 机制，让 Agent 能够对自身执行结果进行自我评估和验证，从而实现更高级别的自主性。这是 Bootstrap Loop 的关键补充，为自因化的"自我反思"能力奠定基础。

## 2. Goals

### 2.1 Self-Judgement Framework
- [x] JudgementCriteria 数据结构（验证标准定义）
- [x] SelfJudgementEngine（自我评判引擎）
- [x] JudgementResult 数据结构（评判结果）
- [x] 与 ValidationSummary 集成

### 2.2 Built-in Judgement Strategies
- [x] SyntaxValidationStrategy（语法检查）
- [x] SemanticValidationStrategy（语义检查，基于 LLM）
- [x] ContractValidationStrategy（契约输出验证）
- [x] TestValidationStrategy（测试结果验证）
- [x] CoverageValidationStrategy（覆盖率验证）

### 2.3 Self-Judgement Contract
- [x] `self/judge-execution` contract
- [x] Input: execution_result, criteria, context
- [x] Output: judgement, confidence, suggested_corrections

### 2.4 Integration
- [x] AgentExecutionResult 集成 JudgementResult
- [x] BootstrapOrchestrator 支持自评判
- [x] CLI 命令行支持 judgement 结果展示

## 3. Non-Goals

- Real LLM SDK 集成（M4 已完成）
- 工具自生机制（M5 Bootstrap Loop 基础）
- 分布式评判（单一 Agent 自评判）
- 自动修复能力（仅评判，不自动修复）

## 4. Judgement Criteria Schema

```go
type JudgementCriteria struct {
    Name        string                 // 评判标准名称
    Type        JudgementType          // 评判类型
    Weight      float64                // 权重 (0.0-1.0)
    Thresholds  map[string]float64     // 阈值配置
    Enabled     bool                   // 是否启用
}

type JudgementType string

const (
    JudgementTypeSyntax     JudgementType = "syntax"
    JudgementTypeSemantic    JudgementType = "semantic"
    JudgementTypeContract    JudgementType = "contract"
    JudgementTypeTest       JudgementType = "test"
    JudgementTypeCoverage   JudgementType = "coverage"
    JudgementTypeCustom     JudgementType = "custom"
)
```

## 5. Judgement Result Schema

```go
type JudgementResult struct {
    Passed          bool                // 是否通过
    Score           float64             // 总分 (0.0-1.0)
    Judgements      []JudgementItem     // 各维度评判结果
    Confidence      float64             // 评判置信度
    SuggestedFixes  []string            // 建议修复项
    Metadata        map[string]any      // 额外元数据
}

type JudgementItem struct {
    CriteriaName    string
    Passed          bool
    Score           float64
    Details         string
    Error           string
}
```

## 6. Self-Judgement Contract

```
ContractID: "self/judge-execution"
Input: execution_result, criteria[], context
Output: judgement, confidence, suggested_corrections
```

## 7. Dependencies

- M5 Bootstrap Loop（T1-T18 完成）
- M4 Real LLM Integration（用于 SemanticValidationStrategy）
- M3 ModelProvider（已集成）

## 8. File Structure

```
internal/agent/
  judgement/
    criteria.go        # JudgementCriteria
    engine.go          # SelfJudgementEngine
    result.go          # JudgementResult
    strategies/
      strategy.go      # ValidationStrategy interface
      syntax.go        # SyntaxValidationStrategy
      semantic.go      # SemanticValidationStrategy
      contract.go      # ContractValidationStrategy
      test.go          # TestValidationStrategy
      coverage.go      # CoverageValidationStrategy
  contracts/
    judge.go           # self/judge-execution contract
```
