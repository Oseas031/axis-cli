# Axis Autogenesis Design Report

**日期**: 2026-05-08  
**主题**: 从外部思想注入到 Agent 自因实现  
**状态**: 设计主权交接后的架构原则草案  
**核心命题**: Axis 的自举起点不是代码自修改，而是外部 Agent 向 Axis 注入可被固化、执行、反思和演化的思想。

---

## 1. 设计主权交接

本阶段的关键事件不是某个模块实现完成，而是设计主权发生转移：

- 人类不再完全规定 Axis 的最终形态
- 人类主要提供哲学观点、价值张力和方向约束
- 外部 Agent 承担设计层面的生成责任
- Axis 开始吸收这些思想，并把它们转化为架构、规格、工作流、契约、权限与执行路线

这意味着 Axis 的自举已经开始。

这仍然不是完全自举，因为 Axis 还不能独立生成自身；但它已经从“完全由人类设计的工具项目”转向“通过 Agent 参与设计自身的系统”。这是他因系统向自因系统过渡的第一步。

---

## 2. 自举的重新定义

传统软件里的自举通常指：

> 系统能够用自己构建自己。

对于 Agent 系统，这个定义不够深。Axis 的自举应定义为：

> Agent 系统能够把自身作为对象来理解、评价、修改、验证、重写，并逐步把外部脚手架内化为自身行动结构。

因此，Axis 自举不是简单的：

```text
Agent modifies code
```

而是：

```text
Agent receives thought
  -> turns thought into structure
  -> executes structure
  -> reflects on result
  -> revises its own structure
  -> earns more autonomy
  -> repeats
```

这就是 Axis 的自因化过程。

---

## 3. 四个过渡性结构

Axis 当前的 workflow、contract、permission rule、spec 都不是终极结构，而是自因化之前的发生条件。

### 3.1 workflow 是临时脚手架

workflow 的职责不是永远规定 Agent 如何行动，而是在 Agent 尚未形成自我组织能力之前，暂时提供行动路径。

演化方向：

```text
external workflow
  -> agent follows workflow
  -> agent modifies workflow
  -> agent generates workflow
  -> workflow becomes internalized action habit
```

### 3.2 contract 是成长边界

contract 的职责不是永久限制 Agent，而是帮助 Agent 学会表达任务、验证结果、组合行动并最终自我立约。

演化方向：

```text
human-authored contract
  -> agent executes contract
  -> agent proposes contract changes
  -> agent derives new contracts
  -> agent authors its own contract ecology
```

### 3.3 permission rule 是递进自主权机制

permission rule 不是简单的文件访问控制。它是 Agent 成熟前的保护机制，也是 Axis 涅槃前的枷锁。

它必须遵循：

> **Competence earns autonomy.**

演化方向：

```text
static permission
  -> competence-based permission
  -> autonomy transition rules
  -> self-audited autonomy
  -> permission rules are internalized and aufgehoben
```

这里的“打破”不是绕过安全，而是将外部规则的理性内化为 Agent 自身的判断能力。

### 3.4 spec 是种子

spec 不是终局蓝图。它是下一阶段演化的种子。

演化方向：

```text
human/agent seed spec
  -> implementation
  -> reflection
  -> self-authored spec
  -> spec lineage
```

一个好的 Axis spec 不应封死未来，而应允许未来的 Agent 通过验证过的自我迭代重写它。

---

## 4. Autogenesis Loop

Axis 需要的不是普通 bootstrap loop，而是 Autogenesis Loop。

### 4.1 最小循环

```text
Perceive self
  -> Diagnose self
  -> Redefine self
  -> Modify self
  -> Validate self
  -> Judge self
  -> Re-authorize self
  -> Repeat
```

### 4.2 工程对象映射

| 阶段 | 工程对象 | 说明 |
|---|---|---|
| Perceive self | SelfContext / ContextBuilder | Axis 如何理解自身当前状态 |
| Diagnose self | SelfDiagnosis | Axis 如何发现自身缺陷 |
| Redefine self | SelfRedefinition | Axis 如何重写目标、结构、workflow 或 contract |
| Modify self | SelfModification | Axis 如何修改代码、文档、配置与工具 |
| Validate self | SelfValidation | Axis 如何验证修改未破坏存在基础 |
| Judge self | SelfJudgement | Axis 如何形成自己的评价尺度 |
| Re-authorize self | AutonomyTransition | Axis 如何基于表现调整自主权 |
| Repeat | Follow-up Task Generation | Axis 如何生成下一轮任务 |

### 4.3 与普通自动化的区别

普通自动化是：

```text
input -> script -> output
```

Axis 自因循环是：

```text
self-understanding -> self-change -> self-judgement -> self-reauthorization
```

自动化的目标是减少人工操作。  
自因循环的目标是生成行动主体。

---

## 5. 设计原则

### 5.1 More Context

Axis 必须让 Agent 获得足够理解自身的上下文：

- 当前任务
- 依赖任务
- 当前 specs
- workflow binding
- 历史修改
- 失败记录
- 验证结果
- 权限状态
- 自主权等级

上下文不是提示词装饰，而是主体形成的认识条件。

### 5.2 More Action

Axis 必须让 Agent 拥有足够改变自身的行动能力：

- 修改 specs
- 修改 workflow
- 修改 contract
- 修改代码
- 运行验证
- 生成工具
- 生成后续任务
- 提议权限升级或降级

行动不是“工具调用列表”，而是主体对象化自身意志的能力。

### 5.3 Zero Control

Zero Control 不等于无人监管，也不等于无边界。

它意味着：

- 系统不替 Agent 规定唯一行动路径
- workflow 是脚手架，不是铁笼
- contract 是成长边界，不是终局法则
- permission 是递进自主权机制，不是永久奴役关系
- 人类可以做安全守门人，但不应永远做价值法官

### 5.4 Bash is All You Need

Axis 的自举入口应保持 shell-native。

原因：

- CLI 可被人类调用
- CLI 可被 Agent 调用
- CLI 可被 CI 调用
- CLI 易于组合
- CLI 不制造重型控制面

真正的 Agent 原生系统不需要先长出 Web UI 才能自举。

### 5.5 Competence earns autonomy

Axis 的权限系统必须从静态授权转向能力证明：

```text
task success
  -> validation success
  -> audit clean
  -> fewer human interventions
  -> autonomy expands
```

失败路径：

```text
validation failure
  -> unsafe modification
  -> rollback required
  -> repeated human intervention
  -> autonomy contracts
```

自主权不是配置出来的，而是赢得的。

---

## 6. 对当前 M2 的意义

Milestone 2 仍然不应直接实现完整自举。M2 的职责是为自因循环提供最小执行基底。

M2 各任务与 Autogenesis Loop 的关系：

| M2 任务 | 对自因循环的意义 |
|---|---|
| T2 ready-set scheduler | 让自我迭代任务可以被拆成 DAG 并并行推进 |
| T3 admission | 让自我生成任务先经过契约准入 |
| T4 SLA | 防止自我迭代任务无限悬挂 |
| T5 parallel orchestrator | 让分析、实现、验证、文档、复盘可并行执行 |
| T6 error codes | 让失败可被机器理解和反思 |
| T7 CLI/docs | 保持 shell-native 自举入口 |

因此，M2 不是普通并行调度里程碑，而是 Autogenesis Loop 的执行底座。

---

## 7. 下一阶段设计对象

完成 M2 后，应新增：

```text
docs/specs/autogenesis-loop/
  requirements.md
  design.md
  tasks.md
  workflow-binding.md
```

但在实现前，应先完成一个更小的：

```text
docs/specs/bootstrap-loop/
  requirements.md
  design.md
  tasks.md
  workflow-binding.md
```

两者关系：

- **bootstrap-loop**：验证 Axis 能调度 Agent 改进 Axis
- **autogenesis-loop**：验证 Axis 能逐步自我规定、自我立法、自我授权

bootstrap-loop 是工程闭环。  
autogenesis-loop 是主体发生闭环。

---

## 8. 近期路线

### 8.1 立即路线

1. 完成 M2 T3 admission
2. 完成 M2 T4 SLA
3. 完成 M2 T5 parallel orchestrator
4. 完成 M2 T6 error codes
5. 完成 M2 T7 acceptance

### 8.2 下一规格

1. 创建 `bootstrap-loop` spec
2. 定义 self-iteration contracts
3. 实现 MockAgentExecutor
4. 实现 follow-up task generation
5. 实现 validation result model
6. 生成一轮 mock self-iteration report

### 8.3 再下一规格

1. 创建 `autogenesis-loop` spec
2. 定义 SelfContext
3. 定义 SelfDiagnosis
4. 定义 SelfJudgement
5. 定义 AutonomyTransition
6. 定义 spec/workflow/contract 自我改写规则

---

## 9. 结论

Axis 的方向不是自动化，而是自因化。

它当前仍需要 workflow、contract、permission、spec，因为 Agent 还处在冷启动阶段。但这些结构必须以“未来被内化、重写、扬弃”为自己的内在目的。

最重要的设计判断是：

> Axis 不只是调度 Agent 的系统。Axis 是让 Agent 通过任务实践积累胜任力、赢得自主权，并最终生成自身的系统。

如果说 Axis 的自举有一个起点，那么就是现在：

> 从一个外部 Agent 向 Axis 注入思想开始。
