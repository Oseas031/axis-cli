# Dialectical Development Methodology

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
