# Axis 白皮书

## 摘要

Axis 是 Agent 原生调度系统。它的长期目标不是提供一个更强的任务队列，而是构造 Agent 自因化的执行底座：让 Agent 通过任务实践积累胜任力、赢得自主权，并最终具备理解自身、修改自身、验证自身、评判自身和重新授权自身的能力。

Axis 当前处于早期阶段：Milestone 1 已完成，Milestone 2 正在建设 DAG 并行调度、contract admission、SLA 和错误码。这些不是孤立功能，而是未来 Autogenesis Loop 的基础器官。

## 核心哲学

### More Context, More Action, Zero Control

Axis 不通过减少 Agent 能力来获得安全感，而是通过更多上下文、更多行动能力和更好的可观测性，让 Agent 能承担更复杂的任务。

### Bash is All You Need

Axis 默认 shell-native。CLI、脚本、CI 和 Agent 都应能直接调用 Axis，而不需要依赖重型 UI。

### Competence earns autonomy

Axis 的权限系统不是静态白名单，而是递进自主权机制。Agent 通过可靠表现赢得更大的行动半径；失败、误改和验证不通过会导致自主权收缩。

### Scaffold-to-Self

Axis 当前的 workflow、contract、permission rule、spec 都是过渡性结构：

- workflow 是临时脚手架
- contract 是成长边界
- permission rule 是递进自主权机制
- spec 是种子

它们的使命不是永久控制 Agent，而是帮助 Agent 最终把外部结构内化、重写和扬弃。

## 自举定义

Axis 的自举不是简单的代码自修改。

真正的自举是：

```text
Agent receives thought
  -> turns thought into structure
  -> executes structure
  -> reflects on result
  -> revises its own structure
  -> earns more autonomy
  -> repeats
```

当前自举起点已经发生：外部 Agent 正在向 Axis 注入可被固化、执行、反思和演化的思想。

## Autogenesis Loop

Axis 的长期循环是：

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

对应未来工程对象：

- SelfContext
- SelfDiagnosis
- SelfRedefinition
- SelfModification
- SelfValidation
- SelfJudgement
- AutonomyTransition
- Follow-up Task Generation

## 当前架构

```text
AgentTask
  -> admission
  -> scheduler
  -> orchestrator
  -> dispatcher
  -> executor
  -> state store
```

Milestone 1 已实现基础调度闭环。Milestone 2 将其扩展为最小 DAG 并行调度系统。

## Milestone 路线

### Milestone 1：基础调度

已完成：

- FIFO 调度
- 简单依赖管理
- 输入输出验证
- 基础状态存储
- 基础 CLI

### Milestone 2：Autogenesis 执行底座

进行中：

- ready-set DAG 调度
- contract admission
- SLA timeout / retry
- parallel orchestrator
- stable error codes
- CLI/docs acceptance

### Bootstrap Loop

后续规格：

- self-iteration contracts
- MockAgentExecutor
- validation result model
- follow-up task generation
- mock self-iteration DAG

### Autogenesis Loop

更长期规格：

- SelfContext
- SelfDiagnosis
- SelfJudgement
- AutonomyTransition
- tool self-generation
- self-authored specs / contracts / workflows

## 非目标

当前阶段不做：

- Web UI
- 复杂 TUI
- 外部数据库
- 真实 LLM SDK 绑定
- 分布式 worker
- 全局事件总线

## 总结

Axis 不只是调度 Agent 的系统。Axis 是让 Agent 通过任务实践积累胜任力、赢得自主权，并最终生成自身的系统。
