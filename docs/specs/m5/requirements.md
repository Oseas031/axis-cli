# M5 Requirements: Bootstrap Loop

**Status**: Complete
**Last Updated**: 2026-05-10

## 1. Overview

M5 implements the minimal closed loop of the Bootstrap Loop, giving Axis the action structure for autogenesis.

## 2. Goals

### 2.1 AgentExecutor Infrastructure
- [x] AgentExecutor interface definition
- [x] MockAgentExecutor implementation (uses existing ModelProvider)
- [x] AgentRuntimeAdapter (supports external Agent CLI)
- [x] Orchestrator supports agent executor routing

### 2.2 SelfContext / ContextBuilder
- [x] SelfContext data structure
- [x] ContextBuilder implementation
- [x] Context compression

### 2.3 Self-iteration Contracts
- [x] analyze-change-request contract
- [x] implement-change contract
- [x] run-validation contract
- [x] update-docs contract
- [x] review-result contract
- [x] spawn-followup-tasks contract

### 2.4 Bootstrap Loop Integration
- [x] BootstrapOrchestrator supports self-loop tasks
- [x] follow-up task generation
- [x] AutonomyTransition rule engine
- [x] Integration tests

## 3. Non-Goals

- Real LLM SDK integration (keep MockProvider)
- Web UI / TUI
- Distributed workers
- Tool self-generation mechanism
- SelfJudgement (self-validation) — deferred to M6
- Database persistence (file-backed is sufficient)

## 4. Bootstrap Loop Minimal Closed Loop

```
analyze-change-request
    → implement-change
        → run-validation
            → update-docs
                → review-result
                    → spawn-followup-tasks
```

## 5. Dependencies

- M4 completed (M3-M4 are prerequisites)
- AgentExecutor depends on ModelProvider (after T1)
- SelfContext is independent (after T5)
- Contracts depend on AgentExecutor (T8-T13 depend on T1)
- BootstrapOrchestrator depends on all contracts (T14 depends on T13)
