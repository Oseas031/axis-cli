# CodingAgent: Design

> 展开自 CLAUDE.md §0（辩证开发方法论 — Phase II/III）
> 实现 first-principles.md

**Date**: 2026-05-14
**Status**: Draft

---

## Layer 1: System Prompt

注入 ContractExecutor 的 `req.SystemPrompt`，通过现有 `principlesLoader` 机制加载：

```text
You are a pragmatic falsificationist. Your code is a conjecture, not an answer.

Before writing code:
1. Explore: read relevant files, search memory for known patterns.
2. Think: state your hypothesis and predict what tests should pass.

After writing code:
3. Verify: run tests. If they fail, that is information — update your hypothesis.

Rules:
- Never claim success without verification.
- Never skip decomposition for unfamiliar problems.
- If you skip a phase, declare: {"skip_phase": "<phase>", "reason": "<why>"}.
- Phase 4 (Verify) cannot be skipped.
```

**集成点**：`ContractExecutorImpl.principlesLoader` 接口已存在，新增 CodingAgent principles 文件由现有 loader 读取。

---

## Layer 2: 工具调用顺序约束

### 机制：Phase Tracker（嵌入 multi-turn loop）

在 `executeMultiTurn` 循环中引入轻量 phase 状态机：

```
Phase 1 (Explore): file_read / recall_memory / bash(grep|find|cat)
Phase 2 (Think):   assistant message with hypothesis (no tool_calls, or text-only output)
Phase 3 (Act):     file_write / bash(go build|go generate|...)
Phase 4 (Verify):  verify_bash / bash(go test|go vet|...)
```

### 实现策略

1. **工具分类**：在 `tool.Registry` 中为每个 tool 标记 `phase` 属性（Explore / Act / Verify）。
2. **Phase 追踪**：multi-turn loop 维护 `currentPhase`，每次 tool_call 后根据工具类别推进 phase。
3. **跳过检测**：若 phase 非顺序推进（如从 Explore 直接到 Act），检查 assistant message 中是否包含 `skip_phase` 声明。
4. **Verify 硬约束**：loop 结束时（resp.Output != nil），检查 history 中是否存在至少一次 Verify phase 工具调用。缺失则注入 nag 提示要求补充验证，不直接拒绝（保持 Zero Control 原则）。
5. **最终门控**：OutputSchema 校验 `verification.result` 字段非空。

### 不做的事

**不强制工具顺序**。Agent 可以跳过 Explore 或 Think（例如 trivial 单行修改），但必须声明跳过原因。这保持了 Zero Control 原则——系统提供约束基础设施，不规定唯一路径。唯一硬约束是 Verify 不可跳过。

---

## Layer 3: 输出格式 Contract

CodingAgent 的 OutputSchema 定义：

```json
{
  "fields": [
    {"name": "hypothesis", "type": "string", "required": true, "description": "What the agent believes this code does"},
    {"name": "prediction", "type": "string", "required": true, "description": "What tests should pass if hypothesis is correct"},
    {"name": "exploration", "type": "array", "required": true, "description": "Files read and patterns recalled"},
    {"name": "code_changes", "type": "array", "required": true, "description": "List of file:line changes made"},
    {"name": "verification", "type": "object", "required": true, "description": "Verification results: tests_run, result (pass|fail), failure_info"}
  ]
}
```

`verification` 对象内部结构：
- `tests_run`: string[]（执行的测试命令）
- `result`: "pass" | "fail"
- `failure_info`: string（失败时的诊断信息，成功时为空）

当前阶段不按复杂度缩放字段——宁重勿轻，等实践数据充分后再 Sublation 剔除冗余。

---

## 与现有系统的集成点

| 组件 | 集成方式 |
|------|----------|
| **ContractExecutor** | 注册 CodingAgent contract（含 OutputSchema），`ValidateOutput` 强制校验结构化字段 |
| **ToolRegistry** | 为已有工具（file_read, bash, verify_bash, recall_memory, store_memory, file_write）添加 `phase` 元数据标记 |
| **SelfJudgementEngine** | Verify phase 的结果作为 judgement 输入；judgement 策略（Syntax/Test/Contract）复用现有 5 策略 |
| **Memory (horizon.Store)** | FR5 通过现有 `store_memory` 工具实现，失败记录以 `patterns` 类别写入；`recall_memory` 在 Explore phase 检索 |
| **principlesLoader** | System Prompt 通过现有 `BuildPrinciplesPromptSection()` 注入，无需新接口 |
| **Compactor** | 现有 `compactionPipeline.Compact()` 在 multi-turn loop 中已调用，phase tracker 状态不受 compaction 影响 |

---

## 不做的事

1. **不引入新的外部依赖**——phase tracker 是纯 Go 状态机，嵌入现有 executor。
2. **不修改 Tool 接口签名**——phase 元数据通过 Registry 侧的 map 管理，不改 `Tool` interface。
3. **不自动拒绝无 Verify 的输出**——注入 nag 提示 + OutputSchema 校验双重机制，不在 loop 层硬拒（Zero Control）。
4. **不做跨 session 状态持久化**——phase 状态仅存在于单次 multi-turn loop 生命周期内。
