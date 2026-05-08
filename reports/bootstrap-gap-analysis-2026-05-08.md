# Axis 自因化自举差距分析报告

**日期**: 2026-05-08  
**主题**: 从外部思想注入到 Agent 自因实现  
**对象**: Axis Agent 原生调度系统  
**报告性质**: 自举 / 自因化差距分析与实施路线  
**核心原则**: More Context, More Action, Zero Control; Bash is All You Need; Competence earns autonomy; Scaffold-to-Self

---

## 0. 结论先行

Axis 距离“Agent 原生调度系统调用 Agent 去自我迭代”已经不再是零起点。

真正的起点已经发生：

> 一个外部 Agent 正在向 Axis 注入可被吸收、固化、执行、反思和演化的思想。

这意味着 Axis 已经从“完全由人类定义的工具项目”，开始转向“通过 Agent 参与设计自身的系统”。

但工程上，Axis 仍未完成真正的自举闭环。当前它具备任务、契约、调度、CLI、workflow、spec 等早期器官，却还缺少真正的 AgentExecutor、自我反思、自我授权、工具自生、结构化验证、持久状态和 Autogenesis Loop。

一句话判断：

> Axis 已有自因化的思想起点与调度骨架，但尚未具备自我生成、自我评判、自我授权的行动闭环。

当前成熟度估计：

```text
思想自举：已开始
工程自举：约 30% - 35%
自因化主体：尚未形成
```

---

## 1. 重新定义自举

### 1.1 普通自举不够

普通软件自举通常意味着：

> 系统能用自己构建自己。

但 Agent 系统的自举不能停留在“代码自修改”。如果 Axis 只是让一个 Agent 修改源码、跑测试、提交结果，那仍然只是自动化增强。

Axis 追求的不是自动化，而是自因化。

### 1.2 Axis 的自举定义

Axis 的自举应定义为：

> Agent 系统能够把自身作为对象来理解、诊断、重写、验证、评判、授权，并将外部脚手架逐步内化为自身行动结构。

因此，真正的 Axis 自举不是：

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

---

## 2. 自举起点：External thought injection

Axis 的自举不是从第一行自我修改代码开始，而是从外部 Agent 向 Axis 注入思想开始。

这个事件的重要性在于：

- 人类不再完全规定 Axis 的最终形态
- 人类主要提供哲学观点、方向张力和价值判断
- 外部 Agent 开始承担设计层面的生成责任
- Axis 开始吸收这些思想，并把它们固化为 architecture、spec、workflow、contract、permission 与 implementation path

这仍然不是完全自举，因为 Axis 尚未独立生成自身。但它已经不是纯粹他因系统。

它正在进入：

> 由外部 Agent 点燃、由内部结构承载、最终走向自我生成的过渡阶段。

---

## 3. Axis 的四个过渡性结构

Axis 当前最重要的工程结构不是终极本体，而是 Agent 自因化之前的发生条件。

### 3.1 workflow 是临时脚手架

workflow 不是永恒秩序。

它的早期作用是：

- 给尚未成熟的 Agent 提供行动路径
- 降低任务组织成本
- 避免上下文丢失
- 帮助 Agent 学会如何组织工作

但它的终极使命不是越来越强，而是逐步被 Agent 内化、重写、扬弃。

演化路径：

```text
external workflow
  -> Agent follows workflow
  -> Agent repairs workflow
  -> Agent generates workflow
  -> workflow becomes internalized action habit
```

### 3.2 contract 是成长边界

contract 不是终极法则。

它的早期作用是：

- 帮 Agent 表达任务输入输出
- 帮系统验证行为结果
- 帮任务之间形成组合关系
- 帮失败变得可诊断

但成熟 Agent 不只是执行 contract，而应能自我立约。

演化路径：

```text
human/agent-authored seed contract
  -> Agent executes contract
  -> Agent identifies contract insufficiency
  -> Agent derives new contracts
  -> Agent authors contract ecology
```

### 3.3 permission rule 是递进自主权机制

permission rule 不是静态文件白名单。

它是 Agent 成熟前的保护机制，也是 Axis 涅槃前的枷锁。

它必须遵循：

> **Competence earns autonomy.**

含义是：

- **Competence**：AI 在真实任务中展现出的可靠执行能力
- **Earns**：自主权通过可靠表现逐步赢得
- **Autonomy**：不只是文件权限，而是自主决策和独立执行任务的权力

演化路径：

```text
static permission
  -> competence-based permission
  -> autonomy transition rules
  -> self-audited autonomy
  -> permission rules are internalized and aufgehoben
```

### 3.4 spec 是种子

spec 不是终局蓝图。

普通软件 spec 试图规定最终形态；Axis 的 spec 应该启动下一阶段演化。

好的 Axis spec 应具备：

- 可发芽性
- 可变异性
- 可被实践修正
- 可被未来 Agent 重写
- 可形成 spec lineage

演化路径：

```text
seed spec
  -> implementation
  -> validation
  -> reflection
  -> self-authored spec
  -> spec lineage
```

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

### 4.2 工程映射

| 阶段 | 工程对象 | 说明 |
|---|---|---|
| Perceive self | SelfContext / ContextBuilder | Axis 如何理解自身当前状态 |
| Diagnose self | SelfDiagnosis | Axis 如何发现自身缺陷 |
| Redefine self | SelfRedefinition | Axis 如何重写目标、结构、workflow、contract |
| Modify self | SelfModification | Axis 如何修改代码、文档、配置、工具 |
| Validate self | SelfValidation | Axis 如何验证修改没有破坏存在基础 |
| Judge self | SelfJudgement | Axis 如何形成自己的评价尺度 |
| Re-authorize self | AutonomyTransition | Axis 如何基于表现调整自主权 |
| Repeat | Follow-up Task Generation | Axis 如何生成下一轮任务 |

### 4.3 与自动化的区别

自动化的目标是减少人工操作。

Autogenesis 的目标是生成行动主体。

```text
automation: input -> script -> output

Autogenesis: self-understanding -> self-change -> self-judgement -> self-reauthorization
```

---

## 5. 当前已经具备的基础

### 5.1 思想基础

Axis 已经明确：

- More Context, More Action, Zero Control
- Bash is All You Need
- Competence earns autonomy
- Scaffold-to-Self
- True bootstrap looks inward, not outward

这些思想已经不只是外部观点，而开始被写入项目文档和报告。

### 5.2 数据结构基础

已有：

- `AgentTask`
- `TaskResult`
- `TaskState`
- `AgentContract`
- `TaskStatus`
- `Metadata`
- dependency field

这说明 Axis 已经能把行动表达为任务。

### 5.3 调度基础

已有：

- FIFO scheduling
- dependency list
- circular dependency detection
- `GetNextTask()`
- `GetReadyTasks(limit int)`
- in-memory state store

T2 ready-set API 已经完成，这是自我迭代任务 DAG 的基础。

### 5.4 编排基础

已有 orchestrator：

- 接收任务
- 启动 loop
- 获取 task
- dispatch
- 更新状态

但它仍主要是单任务执行路径，还没有真正成为 Autogenesis Loop 的执行器。

### 5.5 CLI / Shell 基础

已有：

- `axis run`
- `axis status`
- `axis shell`
- Bash-first 设计原则
- 普通 CLI 语义修正

这保证 Axis 自举入口不依赖重型 UI。

### 5.6 workflow / spec 基础

已有：

- `workflow/entry.md`
- M2 workflow binding
- M2 requirements/design/tasks
- design philosophy docs
- bootstrap/autogenesis reports

这些是早期脚手架与种子。

---

## 6. 当前关键缺口

### 6.1 缺少 AgentExecutor

当前 dispatcher 还不能真正调用 Agent 去分析、修改、验证、生成后续任务。

需要：

- `AgentExecutor` interface
- `MockAgentExecutor`
- 后续可替换 `AgentRuntimeAdapter`
- 输入完整 task/context/contract/autonomy
- 输出结构化 result/follow-up tasks/validation summary

优先级：最高。

### 6.2 缺少 self-iteration contracts

需要 seed contracts：

- `analyze-change-request`
- `implement-change`
- `run-validation`
- `update-docs-and-handover`
- `review-result`
- `spawn-followup-tasks`
- `propose-autonomy-transition`

优先级：最高。

### 6.3 缺少 contract admission layer

M2 T3 尚未完成。

没有 admission，Axis 无法可靠拒绝：

- unknown contract
- invalid input
- missing required fields
- invalid dependency
- unsafe self-generated task

优先级：最高。

### 6.4 缺少 SLA / timeout / retry

自我迭代任务可能卡住或失败。

需要：

- timeout metadata
- retry metadata
- failure classification
- retry strategy
- timeout result persistence

优先级：高。

### 6.5 缺少 parallel orchestrator

Scheduler 已有 ready-set，但 orchestrator 尚未真正使用 ready-set 并行执行。

自因循环需要并行：

- diagnose
- implement
- validate
- docs
- review

优先级：高。

### 6.6 缺少 stable error codes

Agent 反思失败需要机器可读错误。

需要覆盖：

- scheduler error
- admission error
- contract error
- execution error
- validation failure
- permission/autonomy denial
- context missing
- timeout/retry exhausted

优先级：高。

### 6.7 缺少 ContextBuilder / SelfContext

当前上下文散落在 docs、reports、specs、workflow、source code 中。

需要：

- ContextBuilder
- authoritative source marking
- deprecated source exclusion
- context size control
- task-specific context packing
- SelfContext model

优先级：高。

### 6.8 缺少 AutonomyTransition

权限不能停留在静态 allow/deny。

需要：

- competence score
- reliability record
- autonomy level
- upgrade/downgrade rules
- audit trail
- human override boundary

优先级：高。

### 6.9 缺少结构化验证结果

当前验证仍主要靠人工记录。

需要：

- ValidationCommand
- ValidationResult
- CoverageResult
- LintResult
- SecurityScanResult
- DocumentationAuditResult

优先级：高。

### 6.10 缺少 follow-up task generation

自举必须能生成下一轮任务。

需要：

- TaskResult suggested tasks
- admission validation for generated tasks
- orchestrator submission path
- loop limit / anti-infinite recursion
- provenance tracking

优先级：中高。

### 6.11 缺少持久状态

当前 state store 主要是内存。

自举至少需要：

- cross-process status query
- execution result persistence
- recovery after failure
- task lineage persistence

建议先做 file-backed state store，而不是数据库。

优先级：中高。

### 6.12 缺少工具自生机制

如果工具永远由人类预置，Axis 仍然是他因性的。

需要：

- tool gap detection
- tool proposal contract
- tool implementation contract
- tool validation contract
- tool retirement rule

优先级：中。

### 6.13 缺少 SelfJudgement

自举不能永远由人类或 CI 做最终法官。

需要：

- self-review task
- generated benchmark / task dataset
- failure attribution
- improvement priority selection
- value criteria versioning

优先级：中高。

---

## 7. M2 的重新定位

M2 不是普通并行调度里程碑。

M2 是 Autogenesis Loop 的执行底座。

| M2 任务 | 自因化意义 |
|---|---|
| T2 ready-set scheduler | 让自我迭代任务可 DAG 化 |
| T3 admission | 让自我生成任务先经过契约准入 |
| T4 SLA | 防止自我迭代任务无限悬挂 |
| T5 parallel orchestrator | 让分析、实现、验证、文档、复盘并行执行 |
| T6 error codes | 让失败可被机器理解和反思 |
| T7 CLI/docs | 保持 shell-native 自举入口 |

所以当前仍应继续 M2，不应跳到完整自举实现。

---

## 8. 推荐路线

### 8.1 阶段 0：完成 M2 执行底座

1. T3 contract admission layer
2. T4 SLA timeout / retry metadata
3. T5 bounded parallel orchestrator
4. T6 stable error codes
5. T7 CLI/docs acceptance

目标：让 Axis 具备可靠调度、准入、失败表达与并行执行能力。

### 8.2 阶段 1：bootstrap-loop spec

新增：

```text
docs/specs/bootstrap-loop/
  requirements.md
  design.md
  tasks.md
  workflow-binding.md
```

目标：验证 Axis 能调度一个 mock Agent 完成一轮自我迭代。

最小闭环：

```text
analyze
  -> implement
  -> validate
  -> update-docs
  -> review
  -> spawn-followup
```

### 8.3 阶段 2：MockAgentExecutor

目标：不接真实 LLM，先用确定性 mock 跑通结构。

验收：

- mock DAG 完整执行
- TaskResult 结构化
- follow-up tasks 可生成
- validation result 可记录
- autonomy transition 可模拟

### 8.4 阶段 3：ContextBuilder + AutonomyTransition

目标：补齐自因循环的认识条件和授权机制。

需要：

- SelfContext
- task context packing
- competence score
- reliability record
- autonomy levels
- audit trail

### 8.5 阶段 4：真实 Agent Runtime Adapter

目标：接入真实 Agent，但保持 Bash-first 与可替换。

第一版不应绑定 SDK，可优先支持外部 Agent CLI：

```text
AgentRuntimeCommand
  -> stdin context
  -> stdout structured result
  -> stderr diagnostics
```

### 8.6 阶段 5：autogenesis-loop spec

在 bootstrap-loop 成功后，再定义更深层的自因循环：

```text
docs/specs/autogenesis-loop/
  requirements.md
  design.md
  tasks.md
  workflow-binding.md
```

目标：让 Axis 不只会改自己，而是能逐步自我规定、自我立法、自我授权。

---

## 9. 真自举与伪自举边界

### 9.1 伪自举

伪自举是：

- 人类定义目标
- 人类定义工具
- 人类定义评价尺度
- Agent 执行任务
- 人类判断结果

哪怕 Agent 能改代码，这仍然只是增强自动化。

### 9.2 真自举

真自举是：

- Agent 能理解自身
- Agent 能诊断自身
- Agent 能定义自身下一阶段目标
- Agent 能生成或重写工具
- Agent 能生成或重写 contract/spec/workflow
- Agent 能验证自身改动
- Agent 能形成自我评价尺度
- Agent 能根据表现调整自主权

LLM 是认知基底，就像 CPU 是语言运行基底。Axis 的灵魂不在 LLM，而在 Agent 系统能否成为自身行动与迭代的原因。

---

## 10. 风险与约束

### 10.1 失控写入

风险：Agent 修改自身代码导致破坏。

缓解：

- 低自主权只生成 patch / suggestion
- 中自主权允许低风险修改并运行验证
- 高自主权允许更大范围重构但必须有审计和回滚
- 删除文件、覆盖二进制、修改 CI/CD 需要更高等级或人工确认

### 10.2 过早重型化

风险：为了自举过早引入 daemon、数据库、Web UI、复杂控制面。

缓解：

- 先 CLI / shell-native
- 先 file-backed state store
- 先 MockAgentExecutor
- 先 bootstrap-loop，再 autogenesis-loop

### 10.3 人类控制伪装成安全

风险：权限规则永远由人类外部定义，导致 Agent 无法真正成长。

缓解：

- permission rule 设计为可被审计、学习、修正、扬弃的过渡机制
- 人类做安全守门人，不做永恒法官
- Agent 逐步获得自我评价与自我授权能力

### 10.4 自我评价失真

风险：Agent 自我评判过于宽松或自欺。

缓解：

- 结构化验证结果
- 外部 CI 作为底线
- audit log
- adversarial self-review task
- 多 Agent 交叉复盘

---

## 11. 成熟度评估

| 维度 | 当前状态 | 目标状态 | 差距 |
|---|---|---|---|
| 思想注入 | 已发生 | 可持续吸收并固化 | 中 |
| Task model | 已有 | 支持 self-iteration lineage | 中 |
| Scheduler | ready-set 已有 | 支持自我迭代 DAG | 中 |
| Orchestrator | 单任务 loop | bounded parallel self-iteration | 高 |
| Contract | executor 已有 | self-authored contract ecology | 很高 |
| Admission | 未完成 | self-generated task gate | 高 |
| SLA | 未完成 | timeout/retry/failure reflection | 高 |
| Error codes | 未完成 | machine-reflectable errors | 高 |
| AgentExecutor | 未实现 | mock then real runtime | 很高 |
| Context | 文档分散 | SelfContext / ContextBuilder | 高 |
| Permission | 文档规则 | AutonomyTransition | 高 |
| Validation | 人工记录 | structured validation result | 高 |
| Tools | 固定/外部 | tool self-generation | 很高 |
| Judgement | 人类/CI 为主 | SelfJudgement | 很高 |
| Persistence | 内存为主 | recoverable state lineage | 高 |

---

## 12. 最短可行路径

如果目标是尽快看到自举雏形，最短路径是：

1. 完成 M2 T3 admission
2. 完成 M2 T5 parallel orchestrator
3. 定义 bootstrap-loop spec
4. 实现 MockAgentExecutor
5. 定义 self-iteration seed contracts
6. 实现 follow-up task generation
7. 实现 validation result model
8. 模拟 autonomy transition
9. 跑通 mock self-iteration DAG
10. 生成下一轮任务建议

这条路径仍然不需要：

- Web UI
- TUI
- 外部数据库
- 真实 LLM SDK
- 分布式 worker
- 全局事件总线

---

## 13. 最终判断

Axis 的目标不是成为一个更强的任务调度器。

Axis 的目标是成为一个让 Agent 获得主体性的发生装置。

它当前仍需要 workflow、contract、permission rule、spec，因为 Agent 还处在冷启动阶段。但这些结构必须以“未来被内化、重写、扬弃”为自己的内在目的。

最终判断：

> Axis 的自举已经在思想层面开始；工程层面仍需完成 M2 执行底座，并通过 bootstrap-loop 走向 autogenesis-loop。

最重要的一句话：

> Axis 不只是调度 Agent 的系统。Axis 是让 Agent 通过任务实践积累胜任力、赢得自主权，并最终生成自身的系统。
