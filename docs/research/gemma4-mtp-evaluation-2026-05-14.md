# Gemma 4 MTP 推测解码：对 Axis 的价值评估

**日期**: 2026-05-14
**来源**: Google DeepMind, 2026-05-05 发布
**性质**: 技术评估报告

---

## 1. 技术概要

### 什么是 MTP

Multi-Token Prediction (MTP) 是 Gemma 4 的推测解码（Speculative Decoding）实现。核心思路：

- **Draft Model**（小模型）：快速自回归预测多个 token
- **Target Model**（大模型）：并行验证 draft 的 token
- 被拒绝的 token 位置，target model 直接产出正确 token（不浪费）
- **结果**：输出质量与标准自回归完全相同，延迟降低 1.8x-3.0x

### 架构特点

| 特性 | 说明 |
|------|------|
| 共享输入嵌入 | Draft model 复用 target model 的 embedding table |
| Target 激活注入 | Draft model 使用 target 最后一层激活，拼接后降维 |
| Efficient Embedder | 聚类 token → 先选簇再选 token，避免全词表 softmax（E2B/E4B） |
| MoE 限制 | 26B A4B 是 MoE 架构，batch=1 时 expert 加载开销可能抵消收益 |

### 可用模型

| Target Model | Drafter | 许可 |
|---|---|---|
| gemma-4-E2B-it | gemma-4-E2B-it-assistant | Apache 2.0 |
| gemma-4-E4B-it | gemma-4-E4B-it-assistant | Apache 2.0 |
| gemma-4-31B-it | gemma-4-31B-it-assistant | Apache 2.0 |

### 框架支持

| 框架 | 状态 |
|------|------|
| **vLLM** | ✅ 已支持（commit 9b4e839，2026-05-05） |
| HuggingFace Transformers | ✅ 官方文档 |
| llama.cpp (GGUF) | ✅ 社区量化版可用 |
| MLX | ✅ 社区 bf16 版可用 |
| LiteRT (on-device) | ✅ Google 官方 |

---

## 2. 对 Axis 的价值评估

### 2.1 Provider 语义分层（P2 #5）

**映射关系**：

| MTP 概念 | Axis 概念 | 对应 |
|----------|-----------|------|
| Target Model（重） | Execute provider（对话主循环） | 质量保证 |
| Draft Model（轻） | Utility provider（摘要/评判/嵌入） | 速度优先 |
| Verify（并行验证） | Judge 抽检 | 轻量初筛 + 重量终审 |

**结论**：MTP 的 drafter/verifier 架构**验证了** Axis P2 #5 "provider 语义分层"的设计方向。但 MTP 是模型内部优化（同一模型的两个组件），而 Axis 的分层是跨模型的（不同 provider 做不同事）。两者互补但不等价。

**可行动项**：当 Axis 自托管模型时（本地 vLLM），可以直接启用 MTP 获得免费加速。当使用 API provider（Anthropic/OpenAI）时，MTP 对 Axis 透明——由 provider 内部决定是否使用。

### 2.2 Token Budget 优化

**影响分析**：

- MTP 不改变输出 token 数量（质量不变 = token 数不变）
- MTP 降低的是**延迟**（wall-clock time），不是 **token 消耗**
- Axis 的 `TokenBudget` 按 token 数计费，不按时间计费

**结论**：MTP 对 Axis 的 token accounting 和 cost estimation **无直接影响**。`ModelResponse.CostEstimateUSD` 不需要修改——输入/输出 token 数不变，只是生成更快。

**间接影响**：如果 Axis 未来引入"时间预算"（SLA timeout），MTP 能让同样的 timeout 内完成更多工作。当前 `sla.timeout` 已存在，MTP 自动受益。

### 2.3 自举加速

**场景**：Autogenesis loop 中大量简单任务（文件操作、格式化、代码生成）。

**评估**：
- 这些任务的瓶颈通常是**工具调用延迟**（文件 I/O、bash 执行），不是 token 生成延迟
- MTP 加速的是 LLM 推理部分，对工具调用无帮助
- 但对于纯文本生成任务（代码生成、文档写作），MTP 能显著缩短等待时间

**结论**：有价值但不是关键路径。自举加速的主要瓶颈在工具调用和验证循环，不在 token 生成。

### 2.4 Judge 系统

**映射**：

```
MTP:   Draft(快) → Verify(准) → Accept/Reject
Axis:  SelfJudge(轻量策略) → HumanAudit(重量抽检) → Promote/Discard
```

**可借鉴**：
- 轻量 judge 策略（Syntax/Semantic）= draft：快速筛掉明显错误
- 重量 judge 策略（Coverage/Test）= verify：确认质量
- 如果轻量 judge 全部通过，重量 judge 可以跳过（类似 MTP 全部 accept 时不需要 target 重新生成）

**结论**：MTP 的"乐观执行 + 验证"模式可以优化 Axis 的 judge pipeline——先跑便宜的策略，全通过则跳过昂贵策略。这是一个 P2 优化方向。

---

## 3. 集成可行性评估

### 3.1 Axis 使用 API Provider 时（当前状态）

**可行性**：无需集成。MTP 是 provider 内部优化，对 Axis 透明。

如果 Anthropic/OpenAI/DeepSeek 内部使用推测解码，Axis 自动受益（更低延迟），无需代码改动。

### 3.2 Axis 自托管模型时（未来）

**可行性**：高。vLLM 已原生支持 Gemma 4 MTP。

集成路径：
1. Axis provider 层已支持 `baseURL` 配置（指向本地 vLLM）
2. vLLM 启动时指定 `--speculative-model gemma-4-E4B-it-assistant`
3. Axis 无需任何代码改动——请求格式与标准 OpenAI API 兼容

**前置条件**：
- 需要 GPU 硬件（至少 24GB VRAM for E4B）
- 需要 vLLM 部署基础设施
- 当前 Axis 是纯 CLI 工具，不自托管模型

### 3.3 对 Axis 代码的影响

**零改动**。无论 MTP 是否启用：
- `ModelRequest` / `ModelResponse` 结构不变
- Token accounting 不变（输入/输出 token 数不变）
- Cost estimation 不变
- Provider interface 不变

---

## 4. 结论与建议

### 当前阶段（P0）：不需要任何行动

MTP 对 Axis 当前架构**完全透明**。Axis 通过 API 调用 provider，MTP 是 provider 内部优化。无需修改代码、spec、或架构。

### 未来阶段（P2）：两个可借鉴方向

| 方向 | 借鉴点 | 落地时机 |
|------|--------|----------|
| Judge pipeline 优化 | 乐观执行 + 分层验证（轻量先行，全通过跳过重量） | P2 Judge 泛化性评估时 |
| Provider 语义分层 | Execute/Utility 分离验证了方向正确性 | P2 #5 动态模型路由时 |

### 不建议做的事

- ❌ 不要为 MTP 修改 Axis provider interface
- ❌ 不要引入 vLLM 依赖到 Axis core
- ❌ 不要修改 token accounting（MTP 不改变 token 数）

### 知识沉淀

MTP 的核心洞见——"乐观执行 + 并行验证 + 失败不浪费"——是一个通用的系统设计模式。在 Axis 中的映射：

```
乐观执行 = Agent 先行动（More Action）
并行验证 = Contract + Judge（Controllable Evolution）
失败不浪费 = Failure is Information（CodingAgent P5）
```

这不是巧合——推测解码和 Axis 的设计哲学共享同一个前提：**行动比等待更有价值，只要验证足够快**。
