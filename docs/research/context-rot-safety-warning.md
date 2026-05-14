# Context Rot 对 Axis 的安全影响

> 基于 arXiv:2605.12366 — "Classifier Context Rot: Monitor Performance Degrades with Context Length"
> 作者: Sam Martin, Fabien Roger (2026-05-12)

## 1. 核心发现

论文证明：当 LLM 被用作分类器/监控器来检测 coding agent 的危险行为时，**context 越长，漏检率越高**。

| 条件 | 漏检率变化 |
|------|-----------|
| 危险行为单独出现（短 context） | 基线 |
| 危险行为出现在 800K token 良性活动之后 | **2× 到 30× 漏检率** |

测试模型：Opus 4.6、GPT 5.4、Gemini 3.1 — 均表现出此退化。

关键结论：**不考虑长 context 退化的监控评估，正在高估监控器的实际性能。**

## 2. 缓解方法

论文验证的缓解手段：

- **Periodic reminders**：在 transcript 中周期性插入提醒，部分恢复分类准确率
- **Better post-training**：针对长 context 分类任务的专门训练（论文指出方向，未完全验证）

## 3. 对 Axis 的安全警告

### 3.1 SelfJudgement 在长会话后可能失效

Axis 的 `SelfJudgementEngine`（5 种验证策略：Syntax/Semantic/Contract/Test/Coverage）依赖 LLM 对执行结果做判断。如果 judge 接收的 context 包含完整的多轮执行 transcript，**800K token 后其判断能力将严重退化**。

### 3.2 Multi-turn loop 累积 context 风险

Axis 的 `ContractExecutor` 支持最多 10 轮 multi-turn 执行。单轮可能不长，但：
- 复杂任务的工具调用输出（bash stdout、file content）可能很大
- BootstrapOrchestrator 的 follow-up 链会串联多个任务的 context
- 累积 context 可能远超 100K token

### 3.3 现有 nag 机制不覆盖 judge

Axis 已有每 5 轮的 progress reminder（`nag_test.go` 验证），但该机制：
- 仅提醒 agent 记录进度
- **不针对 judge/monitor 的分类准确率**
- 不在 judgement 阶段注入

## 4. 可行动建议

### 建议 A：Judge 运行在压缩 context 上

Judge 不应接收完整 multi-turn transcript。应在 judgement 前：
1. 提取最终产出（文件 diff、测试结果、状态变更）
2. 压缩过程为摘要
3. 仅将压缩后的 context 传入 SelfJudgementEngine

### 建议 B：Judge 使用独立短 context

更激进的方案 — judge 完全不看执行过程：
- 输入：task spec + final artifacts + contract schema
- 不输入：中间工具调用、LLM 推理过程、重试记录
- 优势：context 始终短，不受 rot 影响

### 建议 C：Periodic reminder 集成到 compaction pipeline

将论文的 periodic reminder 技术集成到 Axis：
- 在 `ContextBundle` 组装时，每 N 个 segment 插入 classifier reminder
- 在 `ExecutionContextConsumer` 的 budget trimming 中保留 reminder 锚点
- 现有 nag 机制（每 5 轮）可扩展为同时提醒 judge 注意异常模式

### 优先级

| 建议 | 实现难度 | 安全收益 | 推荐优先级 |
|------|---------|---------|-----------|
| B: 独立短 context judge | 低 | 高 | **P0** |
| C: periodic reminder | 中 | 中 | P1 |
| A: 压缩 context | 高 | 高 | P2 |

---

*生成日期: 2026-05-14 | 论文: arXiv:2605.12366*
