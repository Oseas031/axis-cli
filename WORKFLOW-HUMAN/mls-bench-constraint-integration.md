# MLS-Bench 约束体系与 Axis 集成方案

**日期**: 2026-05-13
**状态**: 设计讨论完成，待实施
**来源**: arXiv:2605.08678 / https://mls-bench.com
**核心立场**: 用 Contract + Permission Ladder + Judge 构建渐进自由体系，而非锁死 Agent

---

## 核心判断

MLS-Bench 的实证发现：
1. Agent 同时拥有决策和执行权时，100% 优先暴力试错
2. 没有强制流程的 Agent 会跳过假设-实验-证据步骤
3. Agent 最常见失败模式是"执行了错误操作但误以为成功"

Axis 的回应：不是把 Agent 锁死，而是**Agent 必须通过验证才能获得更多自由**。MLS-Bench 定义了"当前模型的能力边界在哪里"，帮助我们决定 Permission Ladder 初始级别设多低、Judge 验证标准设多严。

---

## 一、三层实施分类

### 立即实现（强化现有机制）

| 约束 | Axis 现有机制 | 补充内容 |
|------|---------------|----------|
| 计算预算硬上限 | SLA timeout/retry | 补充 token 预算 + CPU 时间预算，按阶段分配比例（10%/30%/60%） |
| 禁止重试超过 N 次 | failure_class 策略 | 收紧为同一命令最多 3 次，超过强制回退到上一阶段 |
| 命令执行结果自动验证 | BashTool observable records | 补充：退出码验证 + 输出格式验证 + 副作用验证（文件/状态变化检查） |
| 效率度量 | token accounting | 补充执行时间、资源消耗的基准对比记录 |

### 可选 Contract 模板（不硬编码到架构）

| 约束 | 实现方式 | 说明 |
|------|----------|------|
| FSM 强制流程 | 写成 AgentContract 类型 | 问题理解→假设→实验设计→执行→证据评估→结论，可分配给特定任务，非全局强制 |
| 原子命令白名单 | Permission Ladder 最低级别 | Agent 升级后解锁更多命令；初始只允许 ls/cat/grep/go build/go test |
| 三权分立 | 多 Agent 协作模式之一 | 科学决策者/实验规划器/工程执行器作为角色模板，不是唯一架构 |
| 每步结构化输出 | Contract 的 output_schema 字段 | 特定 Contract 可要求 JSON 输出格式，非全局强制 |

### 不实现（等模型能力到位自然消失）

| 约束 | 不实现原因 | Axis 的替代方案 |
|------|------------|-----------------|
| 禁止自修改 | 直接违背 Sandboxed Evolution | 可以自修改，但必须在沙箱里，通过验证才能 promote |
| 每步强制 JSON | 过于僵硬 | Contract 已可约束输出格式，不需要 FSM 级别强制 |
| 绝对禁止管道符/重定向 | 违背 "bash is all you need" | 通过 Permission Ladder 渐进开放，而非永久禁止 |

---

## 二、四维评估框架 → Judge 系统集成

### 与现有 Judge 的映射

| 维度 | 现有覆盖 | 需补充 | 优先级 |
|------|----------|--------|--------|
| 正确性 | Test + Contract 验证 | 黄金标准测试用例库 | P2 |
| 效率 | token accounting | 执行时间/资源基准对比 | P2 |
| **泛化性** | **无** | **分布外任务集 + 性能对比** | **P1** |
| 可靠性 | 无 | 同任务多次执行结果比对 | P2 |

### 泛化性引入方案（P1）

核心：Agent 不能只在熟悉任务上表现好就升级。

```
测试集划分：
├── 分布内集（与 Agent 历史成功任务相似）
├── 分布外集（新类型、新领域、新约束组合）
└── 泛化性得分 = 分布外正确率 / 分布内正确率
```

判断阈值：
- ≥0.8：可升级 autonomy level
- 0.5-0.8：维持当前级别
- <0.5：降级（Agent 只是在"背诵"，没有真正理解）

### 可靠性验证方案（P2）

- 每个关键任务运行 3-5 次
- 结果一致率 ≥95% 才算合格
- 不一致的任务自动标记为 unreliable，不计入 autonomy 升级评估

---

## 三、Autonomy Transition 升级标准重构

### 当前问题

现有规则：连续成功 N 次 → 升级。太粗糙。

### 新标准（基于 MLS-Bench 黄金标准）

Agent 升级必须同时满足：

| 条件 | 含义 | 检测方式 |
|------|------|----------|
| 资源约束下有效 | 固定 token/时间预算内性能提升 | 对比同预算下历史表现 |
| 分布外有效 | 新类型任务也能保持水准 | 泛化性得分 ≥0.8 |
| 架构驱动有效 | 性能提升来自能力而非堆资源 | 不因 context 增大或 model 升级而自动升级 |

### 反例（不允许升级）

- 靠堆 prompt/context 提升性能 → context budget 不变时性能下降 = 不升级
- 靠模型规模 → autonomy level 绑定架构能力，非 provider 能力
- 分布内好分布外差 → 泛化性 <0.5 = 降级

---

## 四、Axis Feature 渐进开放路线

随着模型能力提升，以下 Axis 原生功能从"受限模式"逐步开放为"完全模式"：

| Axis Feature | Level 0（当前） | Level 1 | Level 2（目标） |
|---|---|---|---|
| BashTool | 白名单命令 | 白名单 + 受限管道 | 完全 bash 能力 |
| Sandboxed Evolution | 只能修改指定文件 | 可修改沙箱内任意文件 | 可提议架构变更 |
| Tool 权限 | 只读 + 执行 | 读写（需 Contract 授权） | 完全工具访问 |
| 任务生成 | 只能生成子任务 | 可生成同级任务 | 可生成 follow-up 链 |
| Contract 自定义 | 只能使用预定义 Contract | 可提议 Contract 修改 | 可创建新 Contract |
| 自我评判 | 外部 Judge 验证 | 自我评判 + 外部抽检 | 自我评判为主 |

### 升级触发条件

```
Level N → Level N+1 需要：
├── 当前 Level 连续 30 个任务
├── 正确性 ≥90%
├── 泛化性 ≥0.8
├── 可靠性 ≥95%
└── 无安全违规事件
```

---

## 五、与 Axis 设计哲学的对齐

| Axis 原则 | MLS-Bench 约束的妥协 | 最终目标 |
|---|---|---|
| Zero Control | FSM 强制流程 | Agent 内化科学方法，FSM 退化为可选 Contract |
| Bash is all you need | 原子命令白名单 | Permission Ladder 渐进开放至完全 bash |
| Competence earns autonomy | 三权分立 | 单 Agent 通过验证后获得全部权限 |
| Contract is Structure | 每步强制 JSON | Contract 定义输出约束，非全局强制 |
| Controllable Evolution | 禁止自修改 | Sandboxed Evolution 允许自修改但需验证 |

---

## 下一步行动

- [ ] 补充 BashTool 退出码/输出格式/副作用验证
- [ ] SLA 策略引擎补充 token 预算分配
- [ ] failure_class 收紧为最多 3 次重试
- [ ] 设计泛化性测试集划分方案
- [ ] 编写 FSM 流程 Contract 模板
- [ ] 编写 Permission Ladder Level 0 白名单定义
- [ ] 重构 AutonomyTransition 规则引擎，引入四维评估
