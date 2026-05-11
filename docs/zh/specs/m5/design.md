# M5 Design: Bootstrap Loop

**Status**: Complete
**Last Updated**: 2026-05-10

## 1. Architecture Overview

```
                    ┌─────────────────────────────────────┐
                    │  BootstrapOrchestrator              │
                    │  (自循环任务调度)                    │
                    └──────────────┬──────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
     ┌────────▼────────┐  ┌───────▼────────┐  ┌───────▼────────┐
     │ AgentExecutor  │  │ Dispatcher     │  │ Scheduler       │
     │ (自因执行器)    │  │               │  │ (ready-set DAG) │
     └────────┬────────┘  └───────┬────────┘  └─────────────────┘
              │                   │
    ┌─────────▼─────────┐  ┌──────▼──────┐
    │ ModelProvider     │  │ Contract    │
    │ (Anthropic/OpenAI)│  │ Executor   │
    └───────────────────┘  └────────────┘
```

## 2. AgentExecutor Interface

```go
type AgentExecutor interface {
    Execute(ctx context.Context, req *AgentExecutionRequest) (*AgentExecutionResult, error)
    GetAutonomyLevel() AutonomyLevel
}

type AgentExecutionRequest struct {
    Task        *types.AgentTask
    SelfContext *SelfContext
    Contract    *types.AgentContract
    Autonomy    AutonomyLevel
}

type AgentExecutionResult struct {
    Output          map[string]any
    FollowUpTasks   []*types.AgentTask
    ValidationResult *ValidationSummary
    AutonomyDelta   AutonomyDelta
    Error           string
}
```

## 3. MockAgentExecutor

MockAgentExecutor 使用现有 ModelProvider 执行任务：

```go
type MockAgentExecutor struct {
    provider provider.ModelProvider
}

func (e *MockAgentExecutor) Execute(ctx context.Context, req *AgentExecutionRequest) (*AgentExecutionResult, error) {
    // 构建 prompt: "Analyze the following change request: {req.Task.Input}"
    // 调用 ModelProvider
    // 解析 response 生成 FollowUpTasks
    // 返回 AgentExecutionResult
}
```

## 4. SelfContext

```go
type SelfContext struct {
    TaskID          string
    TaskLineage     []string
    CodeSnapshot    *CodeSnapshot
    DocSnapshot     *DocSnapshot
    StateSnapshot   *StateSnapshot
    AutonomyLevel   AutonomyLevel
    CompetenceScore float64
}

type CodeSnapshot struct {
    ModifiedFiles []string
    SpecVersion  string
    TaskCount    int
    ToolCount    int
}
```

## 5. ContextBuilder

```go
type ContextBuilder struct {
    stateStore  sharedlayer.StateStore
    scheduler   *scheduler.Scheduler
    gitDir     string
}

func (b *ContextBuilder) BuildSelfContext(taskID string) (*SelfContext, error) {
    // 1. 获取任务状态
    // 2. 获取代码结构 (git diff --stat)
    // 3. 获取文档上下文
    // 4. 构建 TaskLineage
    // 5. 打包返回
}
```

## 6. Self-iteration Contracts

### 6.1 analyze-change-request

```
ContractID: "self/analyze-change-request"
Input: change_description, target_files, motivation
Output: impact_scope, risk_level, suggested_implementation_order
```

### 6.2 implement-change

```
ContractID: "self/implement-change"
Input: analysis_result, implementation_plan
Output: modified_files, new_contracts, implementation_notes
```

### 6.3 run-validation

```
ContractID: "self/run-validation"
Input: modified_files, test_scope
Output: validation_results, is_acceptable, blocking_issues
```

### 6.4 update-docs

```
ContractID: "self/update-docs"
Input: changed_files, validation_summary
Output: updated_docs, new_docs, doc_quality_score
```

### 6.5 review-result

```
ContractID: "self/review-result"
Input: implementation_result, validation_result, doc_result
Output: approval_status, review_notes, suggested_followups
```

### 6.6 spawn-followup-tasks

```
ContractID: "self/spawn-followup"
Input: review_result, current_task_id
Output: new_tasks[], loop_count, termination_reason
```

## 7. AutonomyTransition

```go
type AutonomyLevel int

const (
    AutonomyLevelNone   AutonomyLevel = 0
    AutonomyLevelLow    AutonomyLevel = 1
    AutonomyLevelMedium AutonomyLevel = 2
    AutonomyLevelHigh   AutonomyLevel = 3
    AutonomyLevelFull   AutonomyLevel = 4
)

type AutonomyTransition struct {
    From   AutonomyLevel
    To     AutonomyLevel
    Reason string
    BasedOn CompetenceEvidence
}

type CompetenceEvidence struct {
    TasksCompleted      int
    SuccessRate        float64
    ValidationPassRate float64
    AvgExecutionTime   time.Duration
}
```

## 8. BootstrapOrchestrator

```go
type BootstrapOrchestrator struct {
    *Orchestrator
    loopTracking map[string]int
    maxIterations int
}

func (bo *BootstrapOrchestrator) SubmitSelfIterationTask(task *types.AgentTask) error {
    // 检查 loop count
    // 注入 SelfContext
    // 提交到 Scheduler
}
```

## 9. File Structure

```
internal/agent/
  executor.go           # AgentExecutor interface
  mock_executor.go     # MockAgentExecutor
  runtime_adapter.go   # AgentRuntimeAdapter
  context.go           # SelfContext
  context_builder.go   # ContextBuilder
  followup.go         # FollowUpTaskGenerator
  autonomy.go         # AutonomyTransition model
  contracts/          # Self-iteration contracts
    analyze.go
    implement.go
    validate.go
    update_docs.go
    review.go
    spawn.go
internal/kernel/orchestrator/
  bootstrap.go        # BootstrapOrchestrator
```
