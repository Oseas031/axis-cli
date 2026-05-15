# 研究报告：TencentDB-Agent-Memory

> 研究日期：2026-05-15
> 来源：https://github.com/Tencent/TencentDB-Agent-Memory (v0.3.4)
> 目的：提取对 Axis memory 模块的改进启示

## 1. 项目概述

TencentDB-Agent-Memory 是腾讯开源的 Agent 记忆系统，核心命题：

**符号化短期记忆 + 分层长期记忆**

- 短期：将冗长 tool 日志 offload 到外部文件，用 Mermaid 符号图压缩状态，token 降低 61%
- 长期：将碎片对话蒸馏为结构化 Persona/Scene，而非平坦向量堆

实测效果（与 OpenClaw 集成）：
- WideSearch pass rate +51.52%，token -61.38%
- SWE-bench pass rate +9.93%，token -33.09%
- PersonaMem accuracy 48% → 76%

## 2. 核心架构

### 2.1 四层语义金字塔（长期记忆）

```
L3 Persona    ← 用户画像/偏好/长期目标（persona.md）
L2 Scenario   ← 场景块/解决方案模式（scene_blocks/*.md）
L1 Atom       ← 原子事实/记忆条目（records/*.jsonl + vectors.db）
L0 Conversation ← 原始对话（conversations/*.jsonl）
```

**关键设计**：
- 上层承载判断和方向，下层承载证据和精度
- 完整可追溯链：Persona → Scenario → Atom → Conversation
- 异构存储：底层用 DB（全文检索），顶层用 Markdown（人类可读）

### 2.2 Context Offload 管线（短期记忆）

```
Tool Result → L1(摘要+评分) → L1.5(任务边界判断) → L2(Mermaid图构建) → 注入 context
```

**L1 处理**：
- 每个 tool call/result pair 生成 `OffloadEntry`
- LLM 生成 summary + replaceability score (0-10)
- 全文保存到 `refs/*.md`，摘要写入 `offload.jsonl`

**L1.5 任务边界判断**：
- 判断当前任务是否完成
- 判断新任务是否是历史任务的延续
- 区分 long task vs casual chat
- 决定使用哪个 MMD 文件

**L2 Mermaid 构建**：
- 从 offload.jsonl 中 node_id=null 的条目构建 Mermaid 流程图
- 每个节点有 id/label/status(done|doing|todo)/summary
- 触发条件：null 条目 ≥ 4 或超时 5 分钟

**L3 压缩**（三级策略）：
- **Mild** (50% context window)：替换非当前任务的 tool results（按 replaceability score 排序）
- **Aggressive** (85%)：删除尾部最旧消息的 40%，注入历史 MMD 作为补偿
- **Emergency** (95%)：删除到 60% 以下

### 2.3 混合检索

- BM25（关键词，支持中文 jieba 分词）
- Vector（embedding，支持 OpenAI 兼容 API）
- RRF（Reciprocal Rank Fusion）融合排序
- 超时 5s 降级：跳过注入，不阻塞对话

### 2.4 存储架构

```
~/.openclaw/memory-tdai/
├── conversations/     # L0 原始对话
├── records/           # L1 原子记忆
├── scene_blocks/      # L2 场景块
├── persona.md         # L3 用户画像
├── vectors.db         # SQLite + sqlite-vec
├── offload-{session}.jsonl  # 短期 offload 条目
├── refs/              # tool result 全文
├── mmds/              # Mermaid 流程图文件
└── state.json         # 插件状态
```

## 3. 与 Axis Memory 模块的映射对比

| TencentDB 概念 | Axis 对应 | 差距分析 |
|---|---|---|
| L0 Conversation | longterm (events.jsonl) | Axis 记录事件而非对话，粒度不同 |
| L1 Atom | working (ContextPacket) | Axis 按 bundle 组织，无自动提取 |
| L2 Scenario | 无直接对应 | Axis 无场景聚合层 |
| L3 Persona | horizon/principles | Axis 仅从失败事件提炼，来源单一 |
| Context Offload | 无 | Axis 无 tool result 压缩/外化机制 |
| Mermaid Canvas | 无 | Axis 无符号化任务状态表示 |
| Replaceability Score | 无 | Axis 无法判断哪些内容可安全替换 |
| L1.5 Task Boundary | 无 | Axis 无任务边界自动判断 |
| BM25+Vector+RRF | keyword only | Axis 仅 strings.Contains 匹配 |
| session-scoped state | immediate (per-execution) | Axis 粒度更细但无跨 turn 状态 |

## 4. 关键设计洞察

### 4.1 "可替换性评分"是压缩的关键

TencentDB 的 L1 为每个 tool result 生成 `score: 0-10`，表示"摘要能多好地替代原文"。这使得 mild compression 可以精确选择替换目标，而非盲目删除。

**对 Axis 的启示**：Axis 的 HistoryCompactor 目前是简单截断。引入 replaceability scoring 可以让 compaction 更智能。

### 4.2 任务边界判断是状态管理的核心

L1.5 的 `TaskJudgment` 回答三个问题：
1. 当前任务完成了吗？
2. 新任务是旧任务的延续吗？
3. 这是长任务还是闲聊？

这决定了：哪些内容属于"当前任务"（不可替换），哪些属于"历史任务"（可替换）。

**对 Axis 的启示**：Axis 的 multi-turn loop 没有任务边界概念。引入边界判断可以让 working memory 的 retain/release 更自动化。

### 4.3 符号化压缩 > 文本摘要

Mermaid 流程图比自然语言摘要更高效：
- 结构化：节点/边/状态一目了然
- 可追溯：node_id 链接到原始数据
- 高密度：几百 token 表达几万 token 的工作进度
- LLM 友好：Mermaid 是 LLM 训练数据中的常见格式

**对 Axis 的启示**：Axis 的 immediate memory 用 `path + summary(1024B) + hash`，但没有结构化的任务进度表示。

### 4.4 三级压缩策略的工程智慧

不是一刀切，而是渐进式：
- 50%：温和替换（不影响当前任务）
- 85%：激进删除（但注入历史 MMD 补偿信息损失）
- 95%：紧急清理（保命优先）

**对 Axis 的启示**：Axis 的 HistoryCompactor 是单一策略。分级压缩可以在保持任务连续性的同时控制 token 消耗。

### 4.5 白盒可审计 = Markdown 文件

所有中间产物都是人类可读的文件：
- persona.md、scene_blocks/*.md、offload.jsonl、refs/*.md、*.mmd
- 调试不需要查数据库，直接读文件

**对 Axis 的启示**：Axis 已经遵循这个原则（JSONL + Markdown），这是共同的设计哲学。

## 5. 不适用于 Axis 的设计

| 设计 | 不适用原因 |
|---|---|
| SQLite + sqlite-vec | Axis BOUNDARY.md 明确禁止外部依赖（无 CGO） |
| 后台 pipeline 自动触发 | Axis BOUNDARY.md 禁止后台 goroutine |
| 远程 embedding API | Axis 是本地优先，不依赖外部服务 |
| OpenClaw/Hermes 插件架构 | Axis 是独立 CLI，不是插件 |
| 自动注入 recall 到 prompt | Axis 禁止 push-based context injection |

## 6. 可落地的改进方向

### P1：Tool Result Offload（短期记忆压缩）

**问题**：Axis multi-turn loop 中 tool results 累积消耗大量 token，HistoryCompactor 只能简单截断。

**方案**：在 `internal/memory/` 新增 `offload/` 子模块：
- tool result 全文写入 `.axis/memory/refs/{timestamp}.md`
- 生成摘要 + replaceability score 写入 JSONL
- HistoryCompactor 按 score 选择性替换（而非盲目截断）
- 符合 BOUNDARY.md：无外部依赖，无后台任务，显式触发

### P2：任务状态符号化（Mermaid Canvas）

**问题**：长任务中 Agent 丢失方向感，重复已完成的工作。

**方案**：在 `internal/memory/working/` 扩展 WorkingBundle：
- 新增 `canvas` 字段存储 Mermaid 格式的任务进度图
- 每次 tool call 完成后更新 canvas（显式调用，非后台）
- Agent 可通过 syscall 查询当前 canvas
- 符合 BOUNDARY.md：queryable not injectable

### P3：分级 History Compaction

**问题**：当前 HistoryCompactor 是单一策略，要么不压缩要么全压缩。

**方案**：在 `internal/agent/` 的 HistoryCompactor 引入三级策略：
- Level 1 (50% budget)：替换 score 最高的非当前步骤 tool results
- Level 2 (80% budget)：删除最旧消息，保留最近 N 条
- Level 3 (95% budget)：紧急截断到 60%
- 需要 P1 的 replaceability score 作为前置

### P4：Working Memory 混合检索

**问题**：当前 Recall 仅用 `strings.Contains`，召回精度低。

**方案**：在 `internal/memory/kv/` 或 `working/` 引入 BM25：
- 纯 Go 实现 BM25（无外部依赖）
- 对 bundle.Goal + packet.Summary 建立倒排索引
- 保留 keyword fallback 作为降级路径
- 未来可选：本地 embedding（需评估 CGO 约束）

### P5：Horizon Dream 扩展（成功事件提炼）

**问题**：当前 Dream 仅处理失败事件，错过了从成功中学习的机会。

**方案**：扩展 `horizon/dream.go`：
- 新增 `DreamSuccess` 从成功事件提炼 best practices
- 类似 TencentDB 的 L2→L3 提炼：从多次成功中归纳场景模式
- 写入 `principles/` 而非 `patterns/`
- 触发条件：同类任务成功 ≥ 3 次

## 7. 优先级排序

| 优先级 | 改进 | 理由 |
|---|---|---|
| P1 | Tool Result Offload | 直接解决 token 消耗问题，ROI 最高 |
| P2 | 任务状态符号化 | 解决长任务方向感丢失，用户可感知 |
| P3 | 分级 History Compaction | 依赖 P1，但架构影响大 |
| P4 | Working Memory BM25 | 独立可做，提升召回质量 |
| P5 | Dream 扩展 | 增量改进，风险低 |

## 8. 约束提醒

所有改进必须遵守 `internal/memory/BOUNDARY.md`：
- 无外部依赖（纯 Go 标准库）
- 无后台 goroutine
- 无 push-based injection
- 无物理删除（append-only）
- LF-only 换行
- 显式触发（CLI 或 syscall）
