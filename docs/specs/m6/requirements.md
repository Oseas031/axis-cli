# M6 Requirements: Self-Judgement (Self-Validation)

**Status**: Completed
**Last Updated**: 2026-05-10

## 1. Overview

M6 implements the Self-Judgement mechanism, enabling Agents to self-assess and validate their own execution results, thereby achieving a higher level of autonomy. This is a key complement to the Bootstrap Loop, laying the foundation for autogenesis "self-reflection" capability.

## 2. Goals

### 2.1 Self-Judgement Framework
- [x] JudgementCriteria data structure (validation criteria definition)
- [x] SelfJudgementEngine (self-judgement engine)
- [x] JudgementResult data structure (judgement results)
- [x] Integration with ValidationSummary

### 2.2 Built-in Judgement Strategies
- [x] SyntaxValidationStrategy (syntax checking)
- [x] SemanticValidationStrategy (semantic checking, LLM-based)
- [x] ContractValidationStrategy (contract output validation)
- [x] TestValidationStrategy (test result validation)
- [x] CoverageValidationStrategy (coverage validation)

### 2.3 Self-Judgement Contract
- [x] `self/judge-execution` contract
- [x] Input: execution_result, criteria, context
- [x] Output: judgement, confidence, suggested_corrections

### 2.4 Integration
- [x] AgentExecutionResult integrates JudgementResult
- [x] BootstrapOrchestrator supports self-judgement
- [x] CLI command support for judgement result display

## 3. Non-Goals

- Real LLM SDK integration (M4 already completed)
- Tool self-generation mechanism (M5 Bootstrap Loop foundation)
- Distributed judgement (single Agent self-judgement only)
- Auto-repair capability (judgement only, no automatic fixes)

## 4. Judgement Criteria Schema

```go
type JudgementCriteria struct {
    Name        string                 // Criteria name
    Type        JudgementType          // Judgement type
    Weight      float64                // Weight (0.0-1.0)
    Thresholds  map[string]float64     // Threshold configuration
    Enabled     bool                   // Whether enabled
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
    Passed          bool                // Whether passed
    Score           float64             // Total score (0.0-1.0)
    Judgements      []JudgementItem     // Per-dimension judgement results
    Confidence      float64             // Judgement confidence
    SuggestedFixes  []string            // Suggested fixes
    Metadata        map[string]any      // Additional metadata
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

- M5 Bootstrap Loop (T1-T18 completed)
- M4 Real LLM Integration (for SemanticValidationStrategy)
- M3 ModelProvider (already integrated)

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
