# Agent-Native Design Philosophy

## Core Thesis

**More Context, More Action, Zero Control, Controllable Evolution**

Axis believes that what Agents need is not denser external control, but:

- **More Context**: Obtain sufficient context to understand tasks, dependencies, history, system state, and failure causes. The system provides efficient query infrastructure and does not actively push redundant information.
- **More Action**: Have the capability to execute, compose, verify, correct, and generate follow-up tasks, and earn permissions and decision-making authority matching those capabilities.
- **Zero Control**: The system provides contracts, infrastructure, and observability, but does not dictate a single action path for the Agent or intervene in the Agent's autonomous decision-making and self-evolution.
- **Controllable Evolution**: The Agent's bootstrapping and self-generation processes must remain within a controllable scope, ensuring correct evolution direction and manageable risk through thought-injection validation, self-modification validation, permission constraints, and audit mechanisms.

## Permission Philosophy

**Competence earns autonomy, autonomy matches responsibility, evolution is controllable**

- **Competence**: Capability is not abstract intelligence, but reliable execution demonstrated in real tasks, dynamically evaluated through quantifiable metrics such as task completion rate, failure post-mortem quality, and risk control ability.
- **Earns**: Autonomy is gradually earned through stable, verifiable, auditable performance, with permission scope precisely matching capability level.
- **Autonomy**: Autonomy is not just file access rights, but the complete capability of autonomous decision-making, independent task execution, initiating contract rewrites, and participating in collaboration.
- **Responsibility**: Autonomy must match risk responsibility. The larger the action radius, the heavier the audit, post-mortem, and risk-bearing burden.
- **Controllable Evolution**: Agent self-evolution, permission elevation, and contract rewriting must pass validation and must not destabilize the system or collaborative order.

## Interaction Principles

**bash is all you need, simple but robust, composable and extensible**

- **Shell-native**: Axis defaults to CLI-first, scriptable, composable, and callable by humans, CI, and Agents.
- **Simple but robust**: Reject redundant Web UIs or complex TUIs; maintain a minimalist design while providing necessary fault tolerance, confirmation, rollback, and observability.
- **Composable and extensible**: Interaction interfaces support multi-dimensional composition and reserve extension space, enabling Axis itself to be directly called, orchestrated, and modified by Agents.

See [Bash is All You Need](bash-is-all-you-need.md).

## Six First Principles

### 1. Interface Is Existence

The essence of any entity is the set of all interfaces it can access.

The core value of interfaces is to enable equal interaction and capability reuse. Humans and Agents are not absolute subjects, but different interface sets:

- Humans possess physical-world interfaces and creative-thinking interfaces.
- Agents possess digital-world interfaces and large-scale reasoning interfaces.
- The system must provide identity-agnostic interface abstractions.
- Universal interfaces guarantee cross-agent calling; extension interfaces preserve individual differences.

Engineering constraints:

- Humans and Agents implement the same class of agent interfaces.
- Any agent can call interfaces exposed by other agents.
- Interface design must not contain identity bias.
- All interface calls must enter observable logs.

### 2. Query Is Context

All intelligence cognition is active, not passive.

The value of context lies in precisely matching needs, not comprehensive coverage. Context should not be forcibly injected by the platform, but actively queried and built by the agent according to the current task.

This principle applies equally to humans and Agents:

- Humans also do not need to see all information, only what the current task requires.
- Agents should not have their judgment polluted by redundant context.
- The system's responsibility is to provide efficient query infrastructure, not to construct cognition for the agent.

Engineering constraints:

- Establish a unified event log system.
- Optimize log indexing for queries on task dependencies, historical records, failure causes, system state, etc.
- All agents use the same query API to obtain context.
- Context management is an internal agent capability; the system only provides query support.

### 3. Ladder Is Boundary

Permissions are proportional to capability and inversely proportional to risk.

Permission is not an identity privilege, but an expression of capability. The "execution ladder" model applies equally to humans and Agents. No one is born with the highest permission.

Permission allocation depends on only two metrics:

- Historical performance.
- Current task needs.

The more reliable the performance, the broader the permission scope; the more unstable the performance, the more verification steps. All high-risk operations require secondary confirmation, regardless of whether the executor is human or Agent.

Engineering constraints:

- Automatically assign initial permissions based on task type.
- Dynamically adjust permission scope according to historical performance.
- Implement secondary confirmation for high-risk operations, confirmable by humans or capable Agents.
- Leave a full trace of permission changes for capability assessment optimization.

### 4. Layered Isolation Is Collaboration

Absolute isolation leads to inefficiency; absolute sharing leads to chaos.

Isolation is the foundation of collaboration; sharing is the means of collaboration. Regardless of whether the executor is human or Agent, independent workspace is needed when executing tasks.

Collaboration does not rely on shared workspaces, but on:

- Shared event logs.
- Version control systems.
- Verifiable merge mechanisms.

Isolation granularity should be dynamically adjusted according to task complexity and agent capability, avoiding excessive isolation or excessive sharing.

Engineering constraints:

- Each task has an independent sandbox.
- Sandbox isolation scope is determined by task type, risk, and agent capability.
- Establish safe merge mechanisms based on Git.
- Shared event logs serve as the collaboration core without destroying independent workspaces.

### 5. Contract Is Structure

The foundation of multi-agent collaboration is explicit, executable, and evolvable contracts.

The file system is the universal contract language among all agents. It defines roles, responsibilities, boundaries, and behavioral norms. Under the premise of observing contracts, each agent can organize internal knowledge in the way it considers optimal.

Contracts are not permanent control structures. Capable agents can initiate contract rewrite applications, which are updated after verification and approval by relevant parties.

Engineering constraints:

- Standardize contract meta-files.
- Contracts have equal binding force on all agents.
- Contract modifications require consent from relevant parties.
- Establish contract rewrite trigger conditions and validation mechanisms.
- Leave a full trace of contract changes as a basis for capability assessment.

### 6. Capability Is Decision Right

Decision rights should belong to the agent that performs best on a specific task.

Final decision rights are not innate human rights, but responsibility based on capability:

- If an Agent performs better on a certain class of tasks, the Agent should have the final decision right.
- If a human performs better on a certain class of tasks, the human should have the final decision right.
- Humans can take over at any time, but the takeover action must be logged.
- Takeover records enter capability assessment and are used to correct subsequent decision-right allocation.

Engineering constraints:

- Establish a capability assessment system based on historical data.
- Evaluate agents using metrics such as task completion rate, failure post-mortem quality, and risk control capability.
- Automatically route tasks to the most suitable agent.
- Decision-right allocation must be observable, traceable, and optimizable.

## 自举起点

Axis 的自举从外部 Agent 向 Axis 注入可被系统吸收、固化、演化的思想开始。

这意味着当前阶段已经是自举起点：

- 人类提供哲学观点、方向张力和价值判断。
- Agent 承担设计层面的主权责任。
- Agent 将思想转化为 spec、workflow、contract、permission、architecture 与实现路径。
- Axis 从被外部设计，转向通过 Agent 参与设计自身。

早期的 workflow、contract、permission rule、spec 都是过渡脚手架。它们的使命不是永久控制 Agent，而是帮助 Agent 积累能力、赢得自主权，并最终将外部结构内化为自身行动结构。

## 传统调度与 Axis 的差异

| 维度 | 传统调度 | Axis |
| --- | --- | --- |
| 实体模型 | 人类控制工具 | 人类与 Agent 都是智能体 |
| 接口 | 身份区分 | 人机无差别抽象 |
| 上下文 | 平台推送或静态注入 | 智能体主动查询 |
| 行动 | 预定义有限操作 | 可组合、可验证、自生成 |
| 权限 | 静态身份授权 | 能力阶梯授权 |
| 协作 | 共享工作区或中心控制 | 沙箱隔离 + 事件日志 + 版本合并 |
| 契约 | 固定规则 | 可验证、可演化结构 |
| 决策权 | 人类默认最终裁决 | 能力决定决策权 |
| 演化 | 外部规划升级 | 可控自举与自我修改 |

## 实施路径

### 里程碑 1：契约与执行基础

- 建立任务契约、执行记录和基础权限模型。
- 提供 CLI 原生入口。
- 形成最小可审计执行闭环。

### 里程碑 2：查询、阶梯与并行协作

- 建立事件日志和查询 API。
- 支持任务依赖、失败原因、历史表现查询。
- 引入能力阶梯、风险分级和高风险确认。
- 支持沙箱隔离和 Git 安全合并。

### 里程碑 3：能力评估与决策路由

- 建立历史表现评估模型。
- 将任务路由到最适合的智能体。
- 让决策权随能力动态转移。
- 将人类接管、Agent 失败、契约变更纳入能力评估。

### 里程碑 4：可控演化

- 支持 Agent 发起契约重写。
- 支持自我修改提案和验证。
- 建立思想注入、自我修改、权限提升的校验机制。
- 形成可观测、可回滚、可复盘的演化闭环。

## 风险边界

Axis 不追求无边界自治。

自治必须满足：

- 行为可观测。
- 决策可追溯。
- 权限可收缩。
- 契约可验证。
- 演化可回滚。
- 高风险动作可二次确认。

Zero Control 不是无约束。它表示系统不替智能体规定唯一行动路径；边界由契约、能力阶梯、隔离层、审计日志和可控演化机制共同构成。

## 结论

Axis 的目标不是制造一个更强的任务调度器，而是建立一个人类与 Agent 都能平等接入、主动查询、能力授权、隔离协作、契约演化、按能力分配决策权的 Agent 原生系统。

它的核心不是控制 Agent，而是让 Agent 在可观测、可验证、可回滚的边界内获得更多上下文、更多行动能力和可控演化空间。
