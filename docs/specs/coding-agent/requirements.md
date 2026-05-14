# CodingAgent: Requirements

> 展开自 CLAUDE.md §0（辩证开发方法论 — Phase II Determinateness）
> 实现 first-principles.md

**Date**: 2026-05-14
**Status**: Draft

---

## Problem Statement

当前 LLM 编码存在三个结构性缺陷：

1. **暴力试错**：拥有执行权后跳过思考，反复修改直到测试通过，不理解根因。
2. **跳过验证**：生成代码后直接宣称成功，不运行测试，不检查副作用。
3. **全知假设**：假装已读过全部代码，在不完整上下文下做出错误推断。

这三个问题的共同根源：LLM 缺乏**认识论纪律**——不区分"猜想"和"知识"。

---

## Functional Requirements

### FR1: Explore-Think-Act 工具调用顺序约束

Agent 的工具调用必须遵循默认路径：Explore → Think → Act → Verify。系统在 multi-turn loop 中追踪当前 phase，若 Agent 跳过某阶段，必须在输出中声明 `skip_phase` + `reason`。未声明的跳过视为违约。

### FR2: 结构化输出

每次代码变更必须附带结构化产物，包含：`hypothesis`（假说）、`prediction`（可证伪预测）、`exploration`（已读取的上下文）、`code_changes`（变更清单）、`verification`（验证结果）。Contract OutputSchema 强制校验这些字段。

### FR3: Phase 4 (Verify) 不可跳过

Verify 是唯一的硬约束。Agent 不得在未执行验证的情况下产出最终结果。验证手段包括 `verify_bash`、`bash(go test)`、或等价的机器可检查操作。跳过 Verify 的输出将被 ContractExecutor 拒绝。

### FR4: 未知问题必须分解再执行

当 Agent 遇到无法直接映射到已知解题路径的问题时，必须先分解为子问题。分解产物为显式的子任务列表，每个子任务独立验证。禁止对未知问题直接套模板。

### FR5: 失败记录进入 long-horizon memory

每次验证失败必须调用 `store_memory` 记录失败信息（失败路径 + 根因 + 排除的假设）。这些记录以 `patterns` 类别存储，供后续 `recall_memory` 检索，避免重复踩坑。

---

## Non-Goals

- **不自主修改 system prompt**：CodingAgent 不得修改自身的行为指令层。
- **不绕过 contract**：所有输出必须通过 ContractExecutor 的 schema 校验，无例外。
- **不自评质量**：Judge 是外部机制（SelfJudgementEngine / 人类），Agent 不做"我的代码质量如何"的判断。

---

## Acceptance Criteria

| # | 条件 | 验证方式 |
|---|------|----------|
| AC1 | Agent 在未执行 Explore 工具的情况下直接 Act，输出包含 `skip_phase` 声明 | 单元测试：mock provider 返回无 explore 的 tool_calls，检查输出格式 |
| AC2 | 最终输出缺少 `verification` 字段时，ContractExecutor 返回 validation error | 单元测试：OutputSchema 校验 |
| AC3 | Agent 跳过 Verify phase 时，执行循环拒绝产出 | 集成测试：multi-turn loop 中无 verify_bash/go test 调用时不接受 final output |
| AC4 | 验证失败后 `store_memory` 被调用 | 集成测试：mock memory store 检查写入 |
| AC5 | 结构化输出 JSON 符合 OutputSchema 定义 | Contract schema 校验测试 |
