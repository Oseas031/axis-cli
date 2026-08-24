# 上下文推荐系统（Context Recommendation System）在 AI Agent 领域的应用研究

> 日期：2026-06-15
> 研究员：AI Agent (Axis CLI)
> 状态：研究报告

---

## 摘要

"上下文推荐系统"（Context Recommendation System）是一个正在 AI Agent 领域快速形成的研究方向。它借鉴传统推荐系统的信息过滤思想，解决 AI Agent 在有限上下文窗口（context window）中如何选择、排序、压缩和刷新上下文内容的核心问题。本报告从定义、技术方法、应用场景、研究论文、开源实现等维度进行系统性梳理。

---

## 1. 定义与核心概念

### 1.1 什么是上下文推荐系统

上下文推荐系统是一种专门服务于 AI Agent（LLM-based Agent）的信息管理机制，其核心任务是：**在每次 LLM 推理前，从海量可用信息中推荐（选择+排序+压缩）最合适的上下文子集，以最大化 Agent 在当前任务中的表现**。

与传统推荐系统推荐"商品给用户"不同，上下文推荐系统推荐"信息给 LLM"。其核心命题：

- **有限注意力预算**（Attention Budget）：LLM 的上下文窗口是稀缺资源，每个 token 都消耗注意力
- **信息边际收益递减**（Diminishing Returns）：更多上下文不等于更好表现，"context rot" 会降低推理质量
- **动态需求**：不同任务阶段需要不同的上下文组合

### 1.2 关键术语

| 术语 | 定义 |
|------|------|
| **Context Window** | LLM 单次推理能处理的最大 token 数 |
| **Attention Budget** | 模型在窗口内有效注意力的隐性限制 |
| **Context Selection** | 从候选信息池中挑选哪些内容进入上下文 |
| **Context Ranking/Ordering** | 确定已选内容在上下文中的排列顺序 |
| **Context Compression** | 将冗长内容压缩为更短的等效表示 |
| **Context Refresh** | 何时以及如何更新已缓存的上下文 |
| **Context Bundle** | 为特定任务组装的上下文包（Axis 术语） |
| **Context Poisoning** | 错误/幻觉信息进入上下文导致级联错误 |
| **Context Distraction** | 过多历史信息导致 Agent 偏离当前任务 |
| **Context Confusion** | 无关工具/文档干扰模型决策 |
| **Context Clash** | 上下文中的矛盾信息导致 Agent 卡死 |

### 1.3 与上下文工程（Context Engineering）的关系

Anthropic 在 2025 年 9 月将"上下文工程"定义为"curating and maintaining the optimal set of tokens during LLM inference"。上下文推荐系统是上下文工程的**核心引擎**——上下文工程定义了"做什么"，上下文推荐系统解决"怎么做"。

```
Context Engineering（全局框架）
  └── Context Recommendation System（核心引擎）
        ├── Selection（选择什么）
        ├── Ranking（如何排列）
        ├── Compression（如何精简）
        └── Refresh（何时更新）
```

---

## 2. 与传统推荐系统的区别

### 2.1 技术对比表格

| 维度 | 传统推荐系统 | 上下文推荐系统 |
|------|------------|--------------|
| **推荐对象** | 商品/内容给人类用户 | 信息片段给 LLM |
| **用户模型** | 人类偏好模型（协同过滤等） | LLM 注意力模型（U 型曲线等） |
| **核心约束** | 展示位数量有限 | Token 预算有限（8K-2M tokens） |
| **质量信号** | 点击率、转化率、停留时间 | 任务完成率、推理准确度、token 效率 |
| **实时性** | 可异步预计算 | 必须在推理前实时组装 |
| **反馈循环** | 用户行为日志 | Agent 执行结果 + 任务成败 |
| **排序依据** | 用户画像 + 物品特征 | 任务目标 + 权威层级 + 时效性 + 语义相关性 |
| **冷启动问题** | 新用户/新物品 | 新任务类型、无历史 |
| **评估指标** | Precision@K, NDCG, MAP | Token 效率比、任务成功率、信息损失率 |
| **数据规模** | 百万级商品库 | 千-万级上下文候选片段 |

### 2.2 核心范式差异

**传统推荐：User-Item 二元关系**
```
User Profile → Item Candidates → Scoring → Top-K → Display
```

**上下文推荐：Task-Context 多维关系**
```
Task Goal + Agent State → Context Candidates → Multi-signal Scoring → Budget-constrained Selection → Ordered Assembly → LLM
```

### 2.3 LLM 特有的"注意力特性"

传统推荐系统处理的是人类注意力（可以自由扫视），而上下文推荐必须处理 LLM 的**位置偏差**：

- **Lost-in-the-Middle 现象**（Liu et al., 2023）：LLM 对上下文开头和结尾的信息注意力高，对中间信息严重遗忘
- **U 型注意力曲线**：开头 85-95% 准确率 → 中间 76-82% → 结尾 85-93%
- **因果注意力掩码**（Causal Attention Masking）：每个 token 只能看到之前的 token，导致位置不对称
- **位置编码插值**（Position Encoding Interpolation）：长上下文中位置信息衰减

---

## 3. 主要技术方法

### 3.1 基于相关性的推荐（Relevance-based）

**核心思想**：根据当前任务目标，计算每个候选上下文片段与任务的语义相关性。

| 方法 | 原理 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|---------|
| **向量相似度检索** | 将任务和文档编码为向量，计算余弦相似度 | 语义理解强，支持模糊匹配 | 需要嵌入模型，计算开销 | RAG 系统 |
| **TF-IDF 检索** | 统计词频-逆文档频率 | 无需外部模型，确定性强 | 不理解语义 | 本地轻量系统 |
| **混合检索** | 向量 + TF-IDF 结合 | 兼顾语义和关键词 | 复杂度增加 | 生产级系统 |
| **Cross-Encoder 重排序** | 联合编码 query-document 对 | 精度高（提升 15-30%） | 计算开销大，延迟高 | 高精度场景 |
| **Query-Aware 摘要** | 根据查询动态生成摘要 | 精准压缩 | 需要额外 LLM 调用 | 文档摘要任务 |

**关键研究**：
- Cross-Encoder reranking 在多文档 QA 任务中 30x 压缩比下可提升准确率 10 个百分点
- 提取式压缩（Extractive Compression）在 10x 压缩比下可保持最小精度损失

### 3.2 基于重要性的推荐（Importance-based）

**核心思想**：根据内容本身的重要性、权威性、稀缺性进行排序。

| 方法 | 原理 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|---------|
| **权威层级排序** | 按来源权威性分层（用户请求 > 代码 > spec > 记忆） | 确定性强，可审计 | 需预定义层级 | Axis 的 FR6 方案 |
| **Token 重要性评分** | 训练 scorer 评估每个 token 的重要性 | 精细到 token 级 | 需要训练 | KV 缓存压缩 |
| **AST 感知压缩** | 代码结构感知的选择性压缩 | 保留语义完整性 | 仅限代码 | 代码分析场景 |
| **工具结果缓存** | 过期工具输出降权或清除 | 简单有效 | 可能丢失有用信息 | Agent 循环执行 |

**关键论文**：
- **KV-Distill**（ICLR 2025）：训练 scorer 保留最重要的 context tokens，在 1000x 压缩比下仍保持连贯生成
- **ICAE**（ICLR 2024, Microsoft）：In-context Autoencoder 将长上下文压缩为短的 memory slots

### 3.3 基于时效性的推荐（Recency-based）

**核心思想**：越新的信息越可能与当前任务相关。

| 方法 | 原理 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|---------|
| **滑动窗口** | 只保留最近 N 条消息 | 简单高效 | 丢弃历史关键信息 | 短对话场景 |
| **指数衰减记忆** | 记忆按时间指数衰减重要性 | 自然遗忘曲线 | 可能遗忘长期关键信息 | 持久记忆系统 |
| **最近访问优先** | 最近使用的文件/工具结果优先 | 符合工作局部性 | 忽略全局重要性 | 编码 Agent |
| **时间戳加权** | 综合 recency + relevance | 平衡新旧 | 权重调优困难 | 混合场景 |

**实践数据**：
- Claude Code 的 compaction 策略：保留最近 5 个访问的文件 + 压缩的历史摘要
- 有效上下文长度（Effective Context Length）远小于标称窗口，通常在 60-70% 处开始退化

### 3.4 基于用户/任务行为的推荐（Behavior-based）

**核心思想**：根据 Agent 的历史行为模式和任务特征动态调整推荐。

| 方法 | 原理 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|---------|
| **任务引导选择** | 分析任务目标动态选择相关上下文 | 针对性强 | 需要任务理解能力 | Agent 任务规划 |
| **渐进式披露** | Agent 探索过程中逐步发现相关上下文 | 自然，类似人类认知 | 慢，可能走弯路 | 代码探索 |
| **Sub-agent 分工** | 子 Agent 处理子任务，返回压缩结果 | 并行高效 | 架构复杂 | 复杂研究任务 |
| **错误驱动刷新** | Agent 执行失败时触发上下文更新 | 按需更新 | 被动 | 交互式任务 |
| **反思评分** | Agent "反思"每条记忆的重要性后才存储 | 质量高 | 需要额外推理 | 持久记忆系统 |

### 3.5 综合对比矩阵

| 方法类别 | 实时性 | 精度 | 计算开销 | 可审计性 | 适用 Agent 类型 |
|---------|--------|------|---------|---------|----------------|
| 相关性（向量检索） | 高 | 中高 | 中 | 中 | RAG Agent |
| 相关性（Cross-Encoder） | 低 | 高 | 高 | 高 | 高精度研究 |
| 重要性（Token Scorer） | 中 | 高 | 中 | 中 | 长上下文 Agent |
| 重要性（权威层级） | 高 | 中 | 低 | 高 | 确定性系统 |
| 时效性（滑动窗口） | 极高 | 低 | 极低 | 低 | 简单对话 |
| 行为（任务引导） | 中 | 中高 | 中 | 高 | 复杂 Agent |
| 混合策略 | 中 | 高 | 中 | 高 | 生产系统 |

---

## 4. 在 AI Agent 中的应用场景

### 4.1 上下文选择（Context Selection）

**问题**：哪些内容应该进入 prompt？

**核心策略**：

1. **Just-in-Time 检索**（Anthropic 推荐）
   - 维护轻量级引用（文件路径、查询、链接）
   - 在推理时动态加载数据
   - 优点：避免过时索引，支持渐进式发现

2. **预算约束选择**
   - 设定最大 token 预算
   - 按权威层级 + 相关性 + 时效性综合排序
   - 在预算内选择 top-K

3. **工具感知选择**
   - 只选择与当前工具集相关的文档
   - 过滤与当前任务阶段无关的上下文

**Axis 实现参考**：`Adaptive Context Assembly`（FR1-FR7）采用规则驱动的 P0 实现，后续 P5 引入 TF-IDF 混合检索。

### 4.2 上下文排序（Context Ranking）

**问题**：已选内容如何排列？

**核心发现**：Lost-in-the-Middle 的对策

```
最佳排列策略（基于 U 型注意力曲线）：

[System Prompt]        ← 位置 0-5%（高注意力）
[最相关文档]           ← 位置 5-20%（高注意力）
[次相关文档]           ← 位置 20-40%（中注意力）  ← 危险区
[其他文档]             ← 位置 40-70%（低注意力）  ← 最危险
[次相关文档]           ← 位置 70-90%（中注意力）
[最相关文档]           ← 位置 90-100%（高注意力）
[当前任务/查询]        ← 位置末尾（最高注意力）
```

**排序技术**：
- **两阶段检索架构**：向量检索（高召回）→ Cross-Encoder 重排序（高精度）
- **注意力排序**（Attention Sorting）：基于注意力分数的迭代重排
- **Multi-Scale Positional Encoding**（Ms-PoE）：缓解位置偏差的架构方案

### 4.3 上下文压缩（Context Compression）

**问题**：如何精简内容以适应预算？

**压缩层次**：

```
Level 1: 工具结果清除（最安全）
  → 过期工具调用结果直接删除

Level 2: 截断（Truncation）
  → 保留头部/尾部，丢弃中间

Level 3: 摘要压缩（Summarization）
  → LLM 生成摘要替代原始内容

Level 4: 提取式压缩（Extractive）
  → 保留关键句子/段落

Level 5: 隐式压缩（Latent）
  → KV 缓存压缩、In-context Autoencoder

Level 6: 架构级压缩
  → 滑动窗口注意力 + 全局注意力交替
```

**关键研究**：

| 论文 | 方法 | 压缩比 | 性能保持 | 特点 |
|------|------|--------|---------|------|
| **ACON** (ICLR 2026) | Agent Context Optimization | 26-54% token 节省 | 95%+ 精度 | 无需微调，自然语言空间优化 |
| **KV-Distill** (ICLR 2025) | KV 缓存蒸馏 | 1000x | 连贯生成 | 参数高效适配器 |
| **ICAE** (ICLR 2024) | In-context Autoencoder | 可变 | 良好 | 将长上下文压缩为 memory slots |
| **Prompt Compression** (Jha et al.) | 多方法对比 | 10x+ | 最小损失 | 提取式压缩在 10x 下表现最强 |
| **Headroom** | 工具输出压缩 | 70-95% token 节省 | 相同答案 | 实用工具，支持多 Agent 框架 |

### 4.4 上下文刷新（Context Refresh）

**问题**：何时以及如何更新已缓存的上下文？

**触发条件**：
1. **Token 预算耗尽** → Compaction（压缩+重启）
2. **任务阶段转换** → 替换阶段相关上下文
3. **Agent 执行失败** → 刷新相关上下文
4. **时间衰减** → 旧上下文降权或移除
5. **外部数据变化** → 增量更新

**刷新策略**：

| 策略 | 描述 | 适用场景 |
|------|------|---------|
| **Compaction** | 压缩历史 → 新窗口 | 长会话 |
| **结构化笔记** | Agent 写笔记到外部存储 | 迭代开发 |
| **Sub-agent 分工** | 子 Agent 返回压缩结果 | 复杂研究 |
| **记忆衰减** | 指数衰减 + 去重 | 持久记忆 |
| **增量更新** | mtime 检测 + 增量索引 | 代码变更 |

**Anthropic Claude Code 实践**：
- Compaction：将对话历史传递给模型进行摘要，保留架构决策、未解决 bug、实现细节
- 工具结果清除：最早实现的轻量级 compaction 形式
- 结构化笔记：NOTES.md 模式，Agent 跨 session 追踪进度

---

## 5. 相关研究论文

### 5.1 核心论文

| 论文 | 年份 | 核心贡献 | 与上下文推荐的关系 |
|------|------|---------|------------------|
| **Lost in the Middle** (Liu et al., Stanford/UC Berkeley) | 2023 | 证实 LLM 对中间位置信息的系统性忽视 | 位置排序的理论基础 |
| **ICAE: In-context Autoencoder** (Microsoft, ICLR 2024) | 2024 | 将长上下文压缩为 memory slots | 上下文压缩 |
| **KV-Distill** (ICLR 2025) | 2025 | Token 级 KV 缓存压缩，1000x 压缩比 | 隐式上下文压缩 |
| **ACON: Agent Context Optimization** (ICLR 2026) | 2026 | 面向长期 Agent 的上下文压缩框架 | Agent 专用压缩 |
| **Characterizing Prompt Compression** (Jha et al., UC Berkeley) | 2025 | 系统评估多种压缩方法 | 压缩方法对比 |
| **ContextSim** (Bougie et al., Woven by Toyota, 2026) | 2026 | 上下文感知的推荐系统评估 | Agent 评估方法 |
| **LLM-based Recommender Systems Survey** (Zhao et al.) | 2024 | LLM 与推荐系统融合综述 | 范式参考 |

### 5.2 技术博客与工程实践

| 来源 | 发布日期 | 核心内容 |
|------|---------|---------|
| **Anthropic: Effective Context Engineering** | 2025-09 | 上下文工程六层架构、Compaction、JIT 检索 |
| **Weaviate: Context Engineering** | 2025-12 | 六大支柱、记忆架构、MCP 工具集成 |
| **NVIDIA: Reimagining LLM Memory** | 2026-01 | Test-Time Training、上下文压缩为权重 |
| **Redis: Context Window Management** | 2026-02 | 实用管理策略：截断、摘要、缓存、滑动窗口 |
| **Microsoft: AI Agents for Beginners - Ch.12** | 2026 | 上下文工程教学框架 |

---

## 6. 开源实现和工具

### 6.1 上下文压缩工具

| 工具 | GitHub Stars | 语言 | 核心功能 | 特点 |
|------|-------------|------|---------|------|
| **Headroom** | - | Python | 工具输出压缩（70-95% token 节省） | 支持 Claude/Codex/Cursor/Aider/LangChain/CrewAI |
| **LLMLingua** | 3k+ | Python | Prompt 压缩 | 微软出品，学术研究基础 |
| **context-engine** (Emmimal) | 191 | Python | 检索+重排序+记忆衰减+token 预算 | 纯 Python，零外部依赖 |
| **Awesome-Context-Compression-LLMs** | 76 | - | 论文集合 | 分类：显式/隐式/KV 压缩 |

### 6.2 上下文管理框架

| 工具 | GitHub Stars | 功能 | 与上下文推荐的关系 |
|------|-------------|------|------------------|
| **LangChain** | 100k+ | RAG pipeline、Retriever 抽象 | 提供检索+重排序基础设施 |
| **LlamaIndex** | 40k+ | 文档索引、RAG 查询引擎 | 文档级上下文管理 |
| **LangGraph** | 30k+ | 有状态 Agent 编排 | 上下文状态持久化 |
| **MemGPT/Letta** | 10k+ | 操作系统式内存管理 | 分级上下文管理 |
| **Weaviate Engram** | - | Agent 记忆 GA | 语义记忆检索 |
| **OpenContext** | 585 | 跨 Agent 知识复用 | 跨 session 上下文管理 |

### 6.3 Agent 记忆系统

| 工具 | 功能 | 上下文推荐相关特性 |
|------|------|------------------|
| **CortexKit Magic-Context** (834 stars) | "海马体"式记忆 | 自管理上下文、无界上下文 |
| **context-mem** (13 stars) | 跨 session 记忆 | 98%+ 检索准确率，44 MCP 工具 |
| **comb-ai** (10 stars) | Context-Optimized Memory Bank | token 使用优化、缓存感知读取 |
| **mex** (744 stars) | 持久项目记忆 | 结构化脚手架 + 漂移检测 |
| **kanwas** (699 stars) | 共享上下文板 | 团队/Agent 间上下文协作 |

### 6.4 Axis 相关实现

| Spec | 状态 | 与上下文推荐的关系 |
|------|------|------------------|
| **Adaptive Context Assembly** | 已有 requirements/design/tasks | Axis 的上下文推荐系统 P0 实现 |
| **Context Compaction** | 已有 requirements/design/tasks | 上下文压缩的 Axis 实现 |
| **Skills System** | 已有 requirements/design | 通过 skills 管理上下文注入 |

---

## 7. 优缺点分析

### 7.1 上下文推荐系统的整体优势

1. **提升 Agent 可靠性**：正确的上下文选择可显著降低幻觉和错误
2. **降低 token 成本**：精确的上下文选择避免浪费（GPT-5.5 输入 $10/M tokens）
3. **改善延迟**：更少 token = 更快首 token 时间（TTFT）
4. **支持长期任务**：通过 compaction 和记忆管理支持小时级 Agent 执行
5. **可审计**：每条上下文选择都有来源、原因、相关性分数

### 7.2 主要挑战与局限

| 挑战 | 描述 | 当前缓解方案 |
|------|------|------------|
| **Lost-in-the-Middle** | 中间位置信息被忽视 | 位置感知排序、多轮推理 |
| **压缩信息损失** | 压缩可能丢失关键细节 | 保留原文引用、渐进式展开 |
| **实时性要求** | 每次推理前需实时组装 | 缓存、增量更新、预计算 |
| **评估困难** | 上下文质量难以量化 | 任务成功率 proxy、人工评估 |
| **冷启动** | 新任务无历史可参考 | 规则 fallback、少样本示例 |
| **Context Poisoning** | 错误信息级联放大 | 来源验证、权威层级 |
| **工具集膨胀** | 工具描述占用大量 token | 工具发现、按需加载 |

---

## 8. 适用场景矩阵

| 场景 | 推荐方法 | 压缩策略 | 刷新策略 | 典型工具 |
|------|---------|---------|---------|---------|
| **代码 Agent（Claude Code 等）** | JIT 检索 + 工具结果缓存 | AST 感知压缩 + 工具结果清除 | Compaction + 笔记 | Headroom, ACON |
| **RAG 问答系统** | 向量检索 + Cross-Encoder 重排 | 提取式压缩 10x | 按查询刷新 | LangChain, LlamaIndex |
| **多 Agent 协作** | Sub-agent 分工返回摘要 | Agent 级摘要压缩 | 任务级刷新 | LangGraph, CrewAI |
| **长期研究 Agent** | 渐进式披露 + 结构化笔记 | Compaction + 摘要 | 里程碑刷新 | Claude Code, Feynman |
| **对话 Agent** | 滑动窗口 + 最近 N 轮 | 窗口截断 | 滑动刷新 | 简单实现 |
| **工具调用密集型 Agent** | 工具结果优先 + 错误驱动刷新 | 工具输出压缩 | 错误触发 | MCP 工具集成 |

---

## 9. 实现建议

### 9.1 架构分层建议

```
Layer 0: 候选收集（Candidate Collection）
  ├── 静态规则（系统 prompt、工具描述、文件路径）
  ├── 动态检索（向量/TF-IDF/JIT 工具调用）
  └── 记忆召回（短期/长期/工作记忆）

Layer 1: 多信号评分（Multi-signal Scoring）
  ├── 相关性分数（语义相似度）
  ├── 权威性分数（来源层级）
  ├── 时效性分数（时间衰减）
  └── 重要性分数（token 重要性 / 任务相关性）

Layer 2: 预算约束选择（Budget-constrained Selection）
  ├── Token 预算分配（system/history/docs 各占多少）
  ├── Lost-in-the-Middle 位置感知
  └── 冲突检测与解决（Context Clash 防御）

Layer 3: 组装与压缩（Assembly & Compression）
  ├── 内容压缩（截断/摘要/提取式/隐式）
  ├── 位置排序（高注意力位置放高优先级内容）
  └── 审计追踪（记录选择原因和遗漏）

Layer 4: 刷新与维护（Refresh & Maintenance）
  ├── Compaction 触发检测
  ├── 记忆衰减与去重
  └── 增量索引更新
```

### 9.2 实现优先级

| 优先级 | 功能 | 复杂度 | 价值 |
|--------|------|--------|------|
| **P0** | 规则驱动的上下文选择 + 权威层级排序 | 低 | 高 |
| **P0** | Token 预算约束 | 低 | 高 |
| **P1** | 工具结果清除（轻量 compaction） | 低 | 中高 |
| **P1** | Lost-in-the-Middle 位置感知排序 | 中 | 高 |
| **P2** | TF-IDF 混合检索 | 中 | 中 |
| **P2** | 摘要压缩 | 中 | 中 |
| **P3** | Cross-Encoder 重排序 | 高 | 中 |
| **P3** | Token 级重要性评分 | 高 | 中 |
| **P4** | 隐式压缩（KV 缓存） | 极高 | 高 |

### 9.3 与 Axis 项目的映射

Axis 的 `Adaptive Context Assembly` spec 已经建立了上下文推荐系统的框架：

- **FR1** Context Bundle Model → 候选数据结构
- **FR2** Assembly Trace → 审计追踪
- **FR3** Rule-based P0 → Layer 0 + Layer 1 的确定性实现
- **FR4** Budgeted Context → Layer 2 预算约束
- **FR5** Preview before execution → 安全观察层
- **FR6** Authority Hierarchy → Layer 1 权威性评分
- **FR8-FR11** Retrieval-backed → Layer 0 的 TF-IDF 增强

**建议的增量改进**：
1. 在 FR6 权威层级中引入时效性衰减因子
2. 在 FR4 预算中加入位置感知的预算分配策略
3. 参考 ACON 论文实现 agent-specific 的压缩指南优化
4. 利用 Headroom 的 ContentRouter 思想进行内容类型感知压缩

---

## 10. 未来方向

1. **Agent 自主上下文管理**：让 Agent 自己决定什么信息值得保留（如 Claude Code 的 NOTES.md 模式）
2. **多模态上下文推荐**：图像、音频、视频内容的上下文选择与压缩
3. **上下文质量评估**：建立标准化的上下文质量 benchmark
4. **Test-Time Training**（NVIDIA）：将上下文压缩为模型权重而非 token 序列
5. **跨 Agent 上下文市场**：多 Agent 系统中的上下文共享与交换机制
6. **上下文推荐的 RLHF**：基于人类反馈优化上下文选择策略

---

## 参考文献

1. Liu, N.F. et al. "Lost in the Middle: How Language Models Use Long Contexts." arXiv:2307.03172, 2023.
2. Anthropic. "Effective context engineering for AI agents." Engineering Blog, Sep 2025.
3. Weaviate. "Context Engineering - LLM Memory and Retrieval for AI Agents." Blog, Dec 2025.
4. Ge, T. et al. "In-context Autoencoder for Context Compression in a Large Language Model." ICLR 2024.
5. KV-Distill. "Nearly Lossless Context Compression." ICLR 2025.
6. Kang, M. et al. "ACON: Optimizing Context Compression for Long-horizon LLM Agents." ICLR 2026.
7. Jha, S. et al. "Characterizing Prompt Compression Methods for Long Context Inference." 2025.
8. NVIDIA. "Reimagining LLM Memory: Using Context as Training Data." Blog, Jan 2026.
9. Bougie, N. et al. "ContextSim: Context-Aware Agent Simulation for Recommender System Evaluation." arXiv:2604.09549, 2026.
10. Microsoft. "AI Agents for Beginners - Chapter 12: Context Engineering." 2026.
11. Headroom. "Compress everything your AI agent reads." GitHub, 2026.
12. Emmimal. "context-engine: Pure-Python context management layer." GitHub, 2026.
