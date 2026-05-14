# Coding Agent — Requirements

> 实现 `docs/architecture/agent-design-first-principles.md` 全文

**Status**: In Progress
**Spec-RDT ID**: coding-agent

## Problem Statement

Axis has complete task scheduling infrastructure (M1-M6) but no LLM-driven executor that can autonomously perform coding tasks. The `MockAgentExecutor` and `SimulationAgentExecutor` validate the pipeline but cannot do real work.

## Requirements

### R1: Multi-turn LLM ↔ Tool execution loop

The executor MUST implement a loop where:
- LLM receives task + history → emits tool calls or final output
- Tools execute deterministically → results append to history
- Loop continues until termination condition met or budget exhausted

### R2: Pluggable termination condition

Termination MUST NOT be solely determined by LLM (LLM doesn't know when code is correct). The system MUST support external termination functions that can inspect history and decide: Continue / Complete / Failed / NeedHuman.

### R3: Circuit breaker

Consecutive tool errors MUST trigger automatic abort to prevent infinite retry loops and token waste. Threshold MUST be configurable.

### R4: History compaction interface

The executor MUST support a compaction hook to manage context window growth during long-running tasks. V1 may be no-op but the interface MUST exist.

### R5: Tool registry injection

Tools MUST be injected via `*tool.Registry`, not hardcoded. Different agents may have different tool sets.

### R6: Provider injection

The LLM provider MUST be injected, not self-selected. Future dynamic routing depends on this.

### R7: Iteration budget externalization

Max iterations MUST be passed in from outside, not hardcoded internally. Future autonomy system will dynamically adjust budgets.

### R8: Complete execution trace

Every tool invocation MUST be recorded with: tool name, input, output/error, duration. AgentID MUST be populated in results.

### R9: System prompt injection point

The executor MUST accept a system prompt that defines its role. Different agent types (coding, review, ops) are differentiated by system prompt.

### R10: Backward compatibility

`NewBashTool()` and existing orchestrator wiring MUST continue to work unchanged. New functionality is additive.
