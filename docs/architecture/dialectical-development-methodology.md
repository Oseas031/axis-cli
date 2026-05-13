# Dialectical Development Methodology

> 展开自 CLAUDE.md §0（作者工作方法论）+ §13（矛盾治理框架）

**Status**: Active Principle
**Supersedes**: `docs/guides/YOUR_IMPLICIT_METHODOLOGY.md` (SRS Loop is the operational form; this is the ontological foundation)

---

## Axis Ontological Position

Axis is **Objectification infrastructure** — providing the conditions for intent to become objective existence through AI, without prescribing the path of transformation.

- Human → Agent Objectification: `axis ask` transforms human intent into AgentTask (objective artifact)
- Agent → Agent Objectification: Subagent/multi-Agent chains pass objectified outputs as inputs
- The motive force currently originates from humans (Level 0); as Agents develop Dream/follow-up capabilities, motive force partially transfers to Agents

Axis provides: scheduling, context, capabilities, isolation, memory, judgement. It does NOT provide: the specific path of transformation.

---

## Core Triad

| 原词 | 哲学严谨表达 | 矛盾论内涵 |
|------|-------------|-----------|
| **Construct** | 对象化（Objectification） | 人的意图通过 AI 转化为客观存在的实践过程，是主体见之于客体的飞跃 |
| **Constraint** | 规定性（Determinateness） | 给 AI 的无限可能性划定质的界限，解决事物"是什么不是什么"的根本矛盾 |
| **Judge** | 扬弃（Aufhebung/Sublation） | 不是简单评判对错，而是保留合理内核、抛弃错误偏差、提升到更高阶段的辩证否定 |

---

## 理论基础（Theoretical Foundation）

> 本文档理论基础层展开 CLAUDE.md §0（体系第一性原理）与 §13（Feedback Protocol），为 dialectical-development-methodology.md 提供认识论与方法论根基，不新增任何 What 级约束。

> **认识论地位声明**：本层命题来自三重实践的结晶：（1）作者的思考与行动经验，（2）人类文明知识的积累（LLM 内化），（3）本对话中的认识迭代（犯错→否定→修正）。其中标注为"公理"的命题是价值选择，不是可证伪的声明。所有命题在后续实践中可被修正或深化。
>
> **受众声明**：本层面向人类设计者，不面向 Agent。Agent 通过 CLAUDE.md（L1）间接受约束，不需要理解哲学来源。

### 内容边界（三不原则）

本层仅包含：
- ✅ 从 Axis 实践中提炼的认识论原理
- ✅ 经过验证的方法论框架
- ✅ 理论参照（毛选/黑格尔）作为旁证，不作为论据

本层不包含：
- ❌ 具体操作步骤
- ❌ 新约束定义
- ❌ 工程化细节

---

### 从 Axis 实践中提炼的五条认识论原理

**一、规则是实践的结晶，不是实践的前提**

Axis 的每一条有效规则都来自具体的失败或成功。Phase 声明三行格式来自"AI 跳过边界确认直接执行"的反复失败；CI 集成测试隔离来自 600s 超时的具体事故；§13 治理框架来自"约束跨对话丢失"的真实痛点。

推论：先写规则再实践 = 教条。先实践再提炼规则 = 正确路径。规则写入 L1 的前提是它已经在 L3 被验证过。

**二、每个阶段只有一个主要矛盾值得解决**

Axis 开发中反复验证：同时解决多个问题 = 每个都解决不好。Phase 声明要求写"主要矛盾侧面"，不是形式主义——它强制 AI 做取舍。今天的 harness 重构就是例证：先解决"约束丢失"（主要矛盾），横向协调和理论实践预判是次生矛盾，在主要矛盾解决后才浮现。

推论：如果一个 Phase 声明里写了两个"主要矛盾"，说明还没想清楚。

**三、理论超前实践时，标记而非强制**

Axis 有多条规则是"写了但还做不到"的：完整的 Crystal Unit 系统、跨项目 Agent 协作、动态模型路由。正确做法不是删除（丢失方向）也不是强制（卡死当前工作），而是标记 `aspirational` 并声明激活条件。

推论：一条规则如果从未被实践触发过，它要么是 aspirational（正确但超前），要么是死规则（应该删除）。区分标准：它保护的对象是否存在。

**四、绕过规则 ≥ 3 次 = 规则有问题，不是执行者有问题**

Axis 实践中验证：当 AI 反复需要绕过某条约束才能完成工作，问题出在约束本身。AGENT_INSTRUCTIONS.md 被绕过（AI 直接读 CLAUDE.md）→ 说明它是多余的中间层 → 删除。`YOUR_IMPLICIT_METHODOLOGY.md` 从未被 AI 主动读取 → 说明它的形式不适配 AI → deprecated。

证据类型不限于计数：绕过次数是量化信号；"从未被触发"是存在性证据；分析性判断（如"功能重叠"）也有效——但必须显式声明证据类型，不能混淆"我觉得"和"数据表明"。

推论：规则的权威性来自它的有效性，不来自它的存在时间。

**五、从实践涌现的模式比预设的规则更可靠**

Phase 声明格式不是设计出来的，是在多次对话中自然涌现、被用户认可、然后才写入 L1 的。这条路径（L3 涌现 → 验证 → L2 提炼 → L1 固化）比"先设计完美规则再执行"更可靠，因为它经过了实践检验。

推论：§13.4 模式 C 是最重要的规则进化路径。大多数好规则不是设计出来的，是长出来的。

---

---

### 本体论：Axis 世界中什么是"存在"

**核心命题：Interface is existence — 没有接口的东西不存在。**

在 Axis 的世界里，"存在"不是指代码被写出来了，而是指它通过接口可被其他实体（人/Agent/CI）观察和交互。一个没有 CLI 命令暴露的内部函数，对 Agent 来说不存在。一个没有 event log 记录的状态变更，对审计来说不存在。

Axis 实践证据：
- `Contract is Structure`：契约不是文档，是结构性存在——Agent 必须通过它才能执行
- 文件系统是共享本体：`.axis/` 目录、spec 文件、event log 是所有实体的共同现实
- Tool 注册 = 存在声明：未注册的 tool 对 Agent 不可见，等于不存在

推论：设计一个功能时，先问"它通过什么接口存在？"如果答不出来，它就不应该被实现。

**核心命题：Objectification 是存在的生成方式。**

事物不是"被发现"的，是"被对象化"的。人的意图通过 `axis ask` 变成 AgentTask——这不是记录意图，是让意图获得客观存在。代码通过 `git commit` 获得存在。规则通过写入 CLAUDE.md 获得存在。口头约定不是存在。

---

### 价值论：Axis 如何判断"好"

**核心命题：有效性是唯一价值标准。**（公理——这是价值选择，不是可证伪的发现。）

Axis 不问"这个方案优雅吗"、"这个架构流行吗"、"这个模式经典吗"。只问：它在实践中有效吗？

Axis 实践证据：
- 规则的权威性来自有效性，不来自存在时间（理论基础原理四）
- Judge 系统的 5 种验证策略全部是可执行的（Syntax/Semantic/Contract/Test/Coverage），没有"主观评审"
- Crystal Unit 存储的是"解题路径 + 闭合证明"，不是"好看的代码"

推论：
- "最佳实践"不是价值来源。在 Axis 中被验证有效的实践才是。
- 审美判断（代码风格、命名偏好）从属于有效性判断，不能凌驾其上。
- 一个丑陋但通过所有验证的方案，优于一个优雅但未验证的方案。

**核心命题：可验证性 > 正确性。**

一个"正确但不可验证"的声明，在 Axis 中没有价值。因为无法验证 = 无法扬弃 = 无法演化。§5 Spec-First 要求验证标准必须是 machine-checkable，正是这个价值论的操作化。

---

### 历史观：Axis 如何理解自身演化

**核心命题：演化是螺旋的，不是线性的。**

Axis 不是从 M1 到 M6 线性进步的。每个里程碑都包含对前一个里程碑的否定：M3 否定了 M2 的"调度器直接调用 provider"假设；M5 否定了"人类必须在循环中"的假设；今天的 harness 重构否定了"多文件分层"的假设。

Axis 实践证据：
- Sandboxed Evolution 的 promote/discard 是演化的原子操作——不是"加新功能"，是"验证后替换旧结构"
- deprecated/ 目录的存在本身就是历史观的体现：旧结构不是被删除，是被扬弃（保留在历史中，精华已提取）
- §13.1 三分类（永久/渐进/过渡）承认规则本身也有生命周期

推论：
- 没有"最终架构"。当前架构是当前矛盾的最优解，矛盾变化时架构必须跟着变。
- "Stable surfaces, replaceable internals"（§11）是演化的具体策略：对外承诺稳定，对内保留替换自由。
- 向后兼容是对历史的尊重，不是对历史的束缚。当兼容成本超过重写成本时，breaking change 是正确选择。

**核心命题：发展的动力是内部矛盾，不是外部需求。**

外部需求可以**触发**工作，但工作的方向和内容由内部矛盾决定。Axis 的每次重大演化都来自内部矛盾激化：
- Long Horizon memory 来自"Agent 跨对话丢失经验"的内部矛盾
- 治理框架来自"约束跨对话丢失"的内部矛盾
- harness 重构来自"文档层级混乱"的内部矛盾

推论：roadmap 不是"用户想要什么"的清单，是"当前主要矛盾是什么"的诊断。

---

### 逻辑学：Axis 中有效推理的规则

**核心命题：时序不可逆 — Phase I → II → III 是逻辑依赖，不是时间分离。**

II 的输入必须是 I 的输出，III 的输入必须是 II 的输出。这是依赖关系，不要求时间间隔——可以在同一呼吸中完成 I→II→III，只要逻辑依赖被满足。不能先执行再确定边界（会产生无边界的垃圾）。不能先确定边界再外化意图（会约束不存在的东西）。

Axis 实践证据：
- 每次"跳过 Phase II 直接执行"都产生了需要返工的结果（§0 rule #1 的来源）
- 每次"没有 Phase I 就开始 Phase II"都产生了"约束了错误的东西"

**核心命题：分解先于综合 — 未知问题必须先拆解为已知子问题。**

这是 CodingAgent P4 的逻辑学基础。不是"分解比较好"，是"对未知整体直接操作在逻辑上不可能产生可靠结果"——因为你无法验证你不理解的东西。

**核心命题：否定是推理的核心动作，不是肯定。**

Axis 的推理模式不是"证明 X 是对的"，而是"排除 X 不是什么"。Determinateness 的本质就是否定——划定边界就是说"不是这个、不是那个"。CodingAgent P1（Conjecture & Refutation）直接来自这个逻辑学立场。

推论：
- 测试的价值不在于"证明代码正确"，在于"排除已知的错误模式"
- Code review 的价值不在于"确认好"，在于"发现不好"
- spec 的 Non-Goals 比 Goals 更重要——它们是 Determinateness 的逻辑表达

---

### 术语对照（理论参照）

| Axis 实践原理 | 理论参照 | 说明 |
|--------------|----------|------|
| 规则是实践的结晶 | 《实践论》实践-认识-再实践 | Axis 验证了这个循环的有效性 |
| 每阶段一个主要矛盾 | 《矛盾论》主要矛盾 | Phase 声明是这个原理的操作化 |
| 标记而非强制 | 对立统一：稳定 vs 发展 | §13.1 三分类是 Axis 的具体解法 |
| 绕过 ≥3 次 = 规则有问题 | 《反对本本主义》反对教条 | Axis 用计数代替主观判断 |
| 涌现比预设更可靠 | 从群众中来到群众中去 | §13.4 模式 C 是 Axis 的实现 |

---

## Fundamental Propositions

**Objectification is free.**

Objectification 近乎无限。AI 可以无限生成，对象化的成本趋近于零。

**Show me your Determisublation.**

Determisublation = 规定性 + 扬弃。这是开发者的真正工作——不是生成，而是界定和扬弃。

---

## Dialectical Movement

### 矛盾一：无边界生成 vs 有边界交付

**Objectification 前确立 Determinateness**，解决 AI 无边界生成与项目有边界交付的主要矛盾。

在 Axis 中的体现：
- Contract 是 Determinateness 的具体形式
- Permission Ladder 是 Determinateness 的渐进开放
- Admission Validator 是 Determinateness 的边界执行

### 矛盾二：偶然生成 vs 必然要求

**Objectification 后完成 Sublation**，解决 AI 生成的偶然与交付的必然要求的新主要矛盾。

在 Axis 中的体现：
- Judge 系统是 Sublation 的机制化
- 复盘四栏（保留/修正/剔除/沉淀）是 Sublation 的操作化
- Sandboxed Evolution 的 promote/discard 是 Sublation 的原子操作

### 统一：螺旋迭代

**Determinateness 与 Sublation 是 Unity of opposites**，通过 Negation of negation 的螺旋迭代，推动项目完成。

```
Objectification（生成）
    ↓
Determinateness（规定：这是什么，不是什么）
    ↓
Objectification（在规定内再次生成）
    ↓
Sublation（扬弃：保留内核，否定偏差，提升阶段）
    ↓
New Determinateness（更高阶段的新规定）
    ↑
    └── Negation of negation ──────────────────
```

---

## Mapping to Axis Architecture

| 辩证环节 | Axis 机制 | 说明 |
|----------|-----------|------|
| Objectification | `axis ask` / Agent 执行 / 代码生成 | 意图→客观存在 |
| Determinateness | Contract / Permission Ladder / Admission / SLA | 划定质的界限 |
| Sublation | Judge / Evolution promote-discard / 复盘四栏 | 辩证否定 |

---

## Mapping to SRS Loop

SRS Loop 是本方法论的操作形式。8 步归入三个辩证阶段：

### Phase I: Objectification（生成）

| 步骤 | 操作 | 退出条件 |
|------|------|----------|
| A0 Posture | 确立方向 | 知道"这次要对象化什么" |
| A1 Externalize | 思想外化为文本 | 想法已成为可被规定的客观物 |

### Phase II: Determinateness（规定）

| 步骤 | 操作 | 退出条件 |
|------|------|----------|
| A2 Inventory | 盘点已有规定 | — |
| A3 Diagnose | 识别缺口 | — |
| A4 Realign | 映射到原则 | — |
| A5 Minimize | 选最小单元 | 下一步行动的边界清晰到不可误解 |

### Phase III: Sublation（扬弃）

| 步骤 | 操作 | 退出条件 |
|------|------|----------|
| A6 Execute | 在规定内生成（代码/文档） | — |
| A7 Distill | 保留/修正/剔除/沉淀 | — |
| A8 Sublate | 写回规则，提升阶段 | 新规则已固化，旧结构已被扬弃或确认 |

**螺旋性**：Phase III 的 A6 是一次受约束的微型 Objectification。A8 完成后产生新的 Determinateness，回到 Phase I。

### 失败回退

- A6 执行失败 → 回退到 Phase II（A5 重新选择最小单元）
- A7 评判发现方向错误 → 回退到 Phase I（A1 重新外化）
- A8 发现原则本身需要修正 → 这是更高阶的 Sublation，更新 `docs/architecture/` 后重新进入循环

---

## Design Decisions Derived from This Methodology

1. **系统不替 Agent 做 Sublation** — Zero Control 原则的哲学根基：Sublation 必须由具有判断力的主体完成，不能被自动化为机械规则
2. **Contract 是 Determinateness 的唯一合法形式** — 硬编码的约束不是 Determinateness，因为它不可被扬弃
3. **Sandboxed Evolution 是 Sublation 的安全容器** — 允许否定（修改），但在验证前不允许否定的结果逃逸到主系统
4. **Progressive Autonomy 是 Determinateness 的动态化** — 规定性不是固定的，它随着 Agent 能力的证明而松绑

---

## When to Reference This Document

- 设计新机制时：问"这属于 Objectification、Determinateness 还是 Sublation？"
- 发现结构性冲突时：问"哪个辩证环节被违背了？"
- 评估是否需要某个功能时：问"它解决的是哪对矛盾？"

---

## Human-AI Collaboration Cognitive Model

（来源：`docs/guides/AI_DRIVEN_DEVELOPMENT_WORKFLOW.md` 精华提炼）

开发者在 AI 协作中的真实角色不是"写代码的人"，而是：

- **AI Workflow Architect**：设计契约、边界、演化协议
- **Quality Assurance Designer**：定义验证标准，不写验证代码
- **Experimenter**：验证"AI 能否通过结构化执行基质产生可靠工作"的假说

**核心洞察**：你的优势是**设计能力**而非编码经验。AI 的优势是**执行能力**而非设计判断。两者的分工不是"你指挥 AI 写代码"，而是"你设计约束空间，AI 在约束内自由生成"。

这就是为什么 Determinateness（规定性）是人类的核心贡献——不是写代码，而是让 AI 只能写出对的代码。

---

## Sublation in Practice: SWE1.6 Renormalization

（来源：`docs/architecture/swe1-6-renormalization-guide.md` 精华提炼）

SWE1.6 重正化是 Axis 历史上第一次大规模 Sublation 实践：

- **Objectification**：早期代码快速生成，积累了大量命名/结构不一致
- **Determinateness**：制定了 11 份规范文档（命名、语义边界、元数据键等）
- **Sublation**：用规范重写现有代码，保留功能（合理内核），否定不一致（错误偏差），提升到有序状态

**关键约束**：重正化不是新增功能的借口。这条约束本身就是 Determinateness 的体现——扬弃必须有边界，否则就是无限扩张。

---

## Operational Principles

1. **主要矛盾优先**：每次 A5 Minimize 时问"这解决的是主要矛盾还是次要矛盾？"主要矛盾未解决前，不分散精力到次要矛盾。
2. **实践检验**：任何设计如果不能在一个 SRS 循环内被 `go test` 验证，就是空谈。实践是检验认识的唯一标准。
3. **阶段论**：不用终态理想批判当前阶段的妥协。每个阶段有每个阶段的正确做法。Level 0 的严格约束不是"违背 Zero Control"，而是当前阶段的正确 Determinateness。
4. **对立统一**：架构权衡不是非此即彼，是统筹兼顾。自由与约束、简单与完备、速度与质量、集中与分散——对立面相互依存。
5. **调查研究**：A2 Inventory 和 A3 Diagnose 必须基于代码事实（读文件、跑测试、grep），不靠记忆推断。没有调查就没有发言权。
