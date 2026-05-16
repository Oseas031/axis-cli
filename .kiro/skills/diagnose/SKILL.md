---
name: diagnose
description: Disciplined diagnosis loop for hard bugs and performance regressions. Reproduce → minimise → hypothesise → instrument → fix → regression-test. Use when user says "diagnose"/"debug"/"为什么挂了"/"这个 bug", or reports something broken/throwing/failing.
tags: [engineering, debug, feedback-loop]
source: mattpocock/skills
source_version: 2026-05-15
---

# Diagnose

> 来源：mattpocock/skills (MIT)，适配 Axis 语境。

纪律化 debug 方法。跳过阶段必须显式说明理由。

## Phase 1 — 建立反馈循环

**这是核心技能。** 有了快速、确定性、可自动运行的 pass/fail 信号，bug 就 90% 解决了。

构建方式（按优先级）：

1. **失败测试** — unit/integration/e2e 任何能触达 bug 的 seam
2. **CLI 调用** — fixture 输入 + diff stdout vs known-good
3. **Throwaway harness** — 最小子系统 + mock deps + 单函数调用
4. **Property/fuzz loop** — 1000 随机输入找失败模式
5. **Bisection harness** — `git bisect run` 自动化
6. **Differential loop** — 新旧版本同输入 diff 输出

迭代优化循环本身：更快？信号更锐利？更确定性？

**非确定性 bug**：目标不是 clean repro 而是提高复现率。循环 100×、并行、加压、缩窄时间窗。

**无法建立循环时**：停下，列出已尝试的方法，向用户要：(a) 可复现的环境 (b) 捕获的 artifact (c) 临时 instrumentation 权限。**不要在没有循环的情况下进入 Phase 2。**

## Phase 2 — 复现

运行循环，确认：
- [ ] 产生的是用户描述的失败模式（不是附近的另一个 bug）
- [ ] 多次运行可复现
- [ ] 已捕获精确症状（error message / wrong output / slow timing）

## Phase 3 — 假设

生成 **3-5 个排序假设**，每个必须可证伪：

> "如果 X 是原因，那么 Y 操作会让 bug 消失 / Z 操作会让它恶化。"

展示给用户后再测试（他们常有领域知识可重排序）。

## Phase 4 — 仪器化

每个探针映射到 Phase 3 的具体预测。**一次只改一个变量。**

- 优先：debugger / REPL 断点
- 其次：带 `[DEBUG-xxxx]` 前缀的 targeted log
- 性能问题：先测量（timing harness / profiler），再修

## Phase 5 — 修复 + 回归测试

1. 将最小复现转为失败测试（在正确的 seam）
2. 看它失败
3. 应用修复
4. 看它通过
5. 重跑 Phase 1 循环验证原始场景

如果没有正确的 seam → 记录这个发现（架构问题）。

## Phase 6 — 清理 + 事后

- [ ] 原始复现不再触发
- [ ] 回归测试通过
- [ ] 所有 `[DEBUG-...]` instrumentation 已删除
- [ ] Throwaway 代码已删除
- [ ] commit message 中说明了正确的假设

**然后问：什么能预防这个 bug？** 如果答案涉及架构变更 → 交给 /improve-codebase-architecture。
