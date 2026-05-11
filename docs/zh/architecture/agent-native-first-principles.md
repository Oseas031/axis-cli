# Agent 原生调度系统的第一性原理

**性质**: 架构底层 rationale（不变的原理层）— **编码前必读**
**关联**: `reports/strategy/agent-native-scenario-principles-analysis-2026-05-11.md`（完整场景推演与缺陷诊断）

> 本文件是 Axis 所有设计决策的根基。原 `agent-native-design-philosophy.md` 的内容已合并至此并废弃。

---

## 核心命题

Agent 原生调度系统的底层本质不是"更聪明的 Cron + BPMN"，而是**面向自主计算实体的生命周期治理系统**。

---

## 六大第一性原理

### 1. Interface is Existence（接口即存在）

> 实体是其所暴露接口的集合。人类和 Agent 实现同一套 agent 接口，无身份偏见。所有接口调用必须留下可观测的日志。

- `axis ask` = 人类通过 CLI 接口向系统注入意图
- `axis status` = 查询 Agent 任务的接口状态
- `.axis/events/tasks.jsonl` = Agent 行为的 append-only 接口调用日志
- 对系统而言，"你是谁"不重要，"你调用了什么接口、产生了什么效果"才重要

### 2. Query is Context（查询即上下文）

> 上下文不是被系统组装后 push 给 Agent 的"参数包"，而是 Agent 主动查询和构建的共享现实。

- `contextpack` 是"上下文经济系统"：预算注意力、记录排除项、可追溯来源
- Agent 应能声明自己的上下文需求，而非被动接受系统组装
- ReadinessArtifact 的 `source_digest` 确保上下文版本一致性，支持事后审计重现

### 3. Ladder is Boundary（阶梯即边界）

> 权限/自治权是能力导向和风险导向的，不是身份导向的。历史表现和任务需求决定动态执行阶梯。

- 新 Agent 获得最小权限（围栏）
- 表现优异的 Agent 渐进获得更多工具权限（阶梯上升）
- 频繁失败的 Agent 权限收缩（阶梯下降）
- 高风险操作需要二次确认，无论执行者是人类还是 Agent

### 4. Layered Isolation is Collaboration（分层隔离即协作）

> 每个任务/Actor 获得隔离的工作空间；协作通过共享事件日志和版本控制发生；隔离粒度适应任务复杂度和 Agent 能力。

- 本地控制平面 = Agent 生态的"市政厅"，协调但不替代自主执行
- 沙盒化演进 = 在隔离空间中实验，验证后才可晋升到主线
- 协作不是"共享内存"，而是"共享不可篡改的历史"

### 5. Contract is Structure（契约即结构）

> 文件系统/元文件是所有 Agent 的共同契约语言。契约对所有 Actor 平等约束，可被有能力的 Agent 在验证和共识后重写，全程留痕。

- `docs/specs/` 下的 requirements/design/tasks 是功能契约
- `.axis/providers.json` 是模型路由契约
- `.axis/runtime.json` 是运行时定位契约
- 契约变更必须通过沙盒化演进协议：实验 → 验证 → 晋升

### 6. Capability is Decision Right（能力即决策权）

> 最终决策权属于对特定任务证明过最佳能力的智能体。人类可随时接管，但接管行为被记录并更新能力评估。

- 调度器根据 Agent 历史表现和能力画像分配任务
- 人类接管不是"失败"，而是"能力评估数据点"
- "Competence earns autonomy, autonomy matches responsibility, evolution is controllable"

---

## 核心主张

> **More Context, More Action, Zero Control, Controllable Evolution**

- **More Context**：Agent 拥有的上下文越多，能做出的决策越好。系统责任是提供可查询、可预算、可审计的上下文，而非控制 Agent 的行为。
- **More Action**：随着上下文增加和可靠性证明，Agent 应获得更大的行动半径（更多工具、更多权限、更少的审批）。
- **Zero Control**：系统不控制 Agent 的"思考过程"，只定义接口边界、记录行为、执行最小权限。控制是边界，不是干预。
- **Controllable Evolution**：Agent 的能力可以进化，但进化必须在沙盒中进行，必须经过验证，人类保留晋升/丢弃的最终决策权。

---

## 关键战略拒绝

| 拒绝项 | 原因 |
|---|---|
| 工作流画布（拖拽式 DAG 编辑器） | 契约和事件日志本身就是流程的定义 |
| 万能 Agent（一个 Agent 做所有事） | 通过 contextpack + provider route 让多个专精 Agent 协作 |
| 黑盒 AI（不可观测的行为） | 所有行为写入 append-only 事件日志，所有上下文可 inspect/preflight |
| 静态权限（人工配置后不变） | 权限基于表现动态调整，但晋升必须通过沙盒验证 |
| Web/TUI 优先 | CLI 原生、可组合、可脚本化——"bash is all you need" |

---

## 演进层

```
P0（本地/保守）        P1（增强）              P2（分布式）            P3（自治）
─────────────────────────────────────────────────────────────────────────────
自然语言意图编译    →  意图质量评分/自动修正  →  多模态意图           →  意图预测
本地控制平面        →  远程/联邦节点         →  联邦集群共识         →  全球 Agent 网络
上下文组装          →  上下文质量评估        →  RAG 融合            →  主动上下文查询
沙盒化演进          →  自动验证测试生成      →  多候选并行进化       →  自主进化策略
工具注册表          →  工具使用学习          →  工具组合发现        →  新工具发明
事件日志            →  长期存储/查询         →  行为模式挖掘        →  组织智能
```

---

## 交互原则

**bash is all you need, simple but robust, composable and extensible**

- **Shell-native**：CLI 优先，可脚本化、可组合、可被人类、CI 和 Agent 调用
- **Simple but robust**：拒绝冗余 Web UI 或复杂 TUI，同时提供必要的容错、确认、回滚和可观测能力
- **Composable and extensible**：接口支持多维组合并预留扩展空间，Axis 自身可被 Agent 直接调用、编排和改造

详见 [Bash is All You Need](bash-is-all-you-need.md)。

---

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

---

## 自举起点

Axis 的自举从外部 Agent 向系统注入可被吸收、固化、演化的思想开始：

- 人类提供哲学观点、方向张力和价值判断
- Agent 承担设计层面的主权责任
- Agent 将思想转化为 spec、workflow、contract、permission、architecture 与实现路径
- Axis 从被外部设计，转向通过 Agent 参与设计自身

早期的 workflow、contract、permission rule、spec 都是过渡脚手架。它们的使命不是永久控制 Agent，而是帮助 Agent 积累能力、赢得自主权，并最终将外部结构内化为自身行动结构。

---

## 风险边界

Axis 不追求无边界自治。自治必须满足：

- 行为可观测
- 决策可追溯
- 权限可收缩
- 契约可验证
- 演化可回滚
- 高风险动作可二次确认

Zero Control 不是无约束。它表示系统不替智能体规定唯一行动路径；边界由契约、能力阶梯、隔离层、审计日志和可控演化机制共同构成。

---

## 结论

Axis 的核心不是控制 Agent，而是让 Agent 在可观测、可验证、可回滚的边界内获得更多上下文、更多行动能力和可控演化空间。
