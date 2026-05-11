# M5 Requirements: Bootstrap Loop

**Status**: Complete
**Last Updated**: 2026-05-10

## 1. Overview

M5 实现 Bootstrap Loop 的最小闭环，让 Axis 具备自因化的行动结构。

## 2. Goals

### 2.1 AgentExecutor Infrastructure
- [x] AgentExecutor interface 定义
- [x] MockAgentExecutor 实现（使用现有 ModelProvider）
- [x] AgentRuntimeAdapter（支持外部 Agent CLI）
- [x] Orchestrator 支持 agent executor 路由

### 2.2 SelfContext / ContextBuilder
- [x] SelfContext 数据结构
- [x] ContextBuilder 实现
- [x] 上下文压缩

### 2.3 Self-iteration Contracts
- [x] analyze-change-request contract
- [x] implement-change contract
- [x] run-validation contract
- [x] update-docs contract
- [x] review-result contract
- [x] spawn-followup-tasks contract

### 2.4 Bootstrap Loop Integration
- [x] BootstrapOrchestrator 支持自循环任务
- [x] follow-up task generation
- [x] AutonomyTransition 规则引擎
- [x] 集成测试

## 3. Non-Goals

- 真实 LLM SDK 集成（保持 MockProvider）
- Web UI / TUI
- 分布式 worker
- 工具自生机制
- SelfJudgement（自评准）
- 数据库持久化（file-backed 足够）

## 4. Bootstrap Loop 最小闭环

```
analyze-change-request
    → implement-change
        → run-validation
            → update-docs
                → review-result
                    → spawn-followup-tasks
```

## 5. Dependencies

- M4 完成（M3-M4 是前置条件）
- AgentExecutor 依赖 ModelProvider（T1 之后）
- SelfContext 独立（T5 之后）
- Contracts 依赖 AgentExecutor（T8-T13 依赖 T1）
- BootstrapOrchestrator 依赖所有 contracts（T14 依赖 T13）
