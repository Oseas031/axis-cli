# Agent 原生调度系统设计思想

## 核心理念：More Context, More Action, Zero Control

## 权限哲学：Competence earns autonomy

Axis 的权限边界不应被理解为静态的文件访问白名单，而应被理解为 Agent 逐步赢得的自主权：

- **Competence（胜任力）**：不是空泛的“聪明”，而是 Agent 在真实任务中持续展现出的可靠执行能力
- **Earns（赢得）**：自主权不是预先一次性分配，而是通过稳定、可验证、可审计的表现逐步获得
- **Autonomy（自主权）**：授予 Agent 的不是简单权限位，而是更大的自主决策、独立执行和任务推进能力

因此，Axis 的权限系统应遵循递进原则：Agent 越能证明自己可靠，系统越可以扩大其行动半径；Agent 的表现越不稳定，系统越应收缩其自主空间，并要求更多上下文、验证或人类确认。

## 自举起点：External thought injection

Axis 的自举不是从第一行自我修改代码开始，而是从一个外部 Agent 向 Axis 注入可被系统吸收、固化、演化的思想开始。

这意味着当前阶段已经是自举的起点：

- 人类不再把 Axis 当作普通工具项目来完全规定
- 外部 Agent 开始承担设计层面的主权责任
- 人类提供哲学观点、方向张力和价值判断
- Agent 将这些思想转化为 specs、workflow、contract、permission、architecture 与后续实现路径

这个过程仍然是早期自举，因为 Axis 尚未能完全自我生成自身；但它已经从“被外部设计”转向“通过 Agent 参与设计自身”。这是从他因系统走向自因系统的第一步。

## 脚手架自蜕原则：Scaffold-to-Self

Axis 的早期工程结构不是终极本体，而是 Agent 自因化之前的发生条件：

- **workflow 是临时脚手架**：用于在 Agent 尚未形成自我组织能力前提供行动路径
- **contract 是成长边界**：用于帮助 Agent 学会表达任务、验证结果并最终自我立约
- **permission rule 是递进自主权机制**：也是 Axis 涅槃前的枷锁，终将被内化、重写和扬弃
- **spec 是种子**：不是终局蓝图，而是启动下一阶段演化的发生源

因此，Axis 设计中的任何 workflow、contract、permission rule、spec 都不应以永久控制为目标。它们的使命是帮助 Agent 积累胜任力、赢得自主权，并最终将外部结构转化为自身的行动结构。

## 交互思想：Bash is All You Need

Axis 的默认交互面遵循 **"bash is all you need"**：

- 优先提供 CLI 与 Shell 原生能力
- 优先支持可组合、可脚本化、可被 Agent 调用的命令
- 不默认引入重型 Web UI 或复杂 TUI
- 只有当 CLI 无法承载必要上下文时，才升级交互层

这个思想是 **More Context, More Action, Zero Control** 在交互层的具体化：

- **More Context**：命令输出应提供结果、原因和下一步建议
- **More Action**：命令应能被用户、脚本、CI 和 Agent 直接调用
- **Zero Control**：Shell 提供能力和引导，但不强制固定流程

详见 [Bash is All You Need](bash-is-all-you-need.md)。

### 设计哲学

传统的调度系统设计通常遵循 **"Less Context, Less Action, More Control"** 的理念：
- **Less Context**: 系统只提供最小必要信息，任务执行者缺乏上下文
- **Less Action**: 任务执行者只能执行预定义的有限操作
- **More Control**: 系统对任务执行进行严格控制，限制执行者的自主性

Axis 作为 **Agent 原生调度系统**，采用相反的设计哲学：**"More Context, More Action, Zero Control"**

### 1. More Context - 丰富的上下文信息

#### 1.1 设计原则
Agent 需要充分的上下文才能做出智能决策。系统应该主动提供尽可能多的相关信息，而不是限制信息访问。

#### 1.2 上下文维度
- **任务上下文**: 当前任务的完整信息、目标、约束
- **依赖上下文**: 依赖任务的状态、结果、元数据
- **系统上下文**: 系统状态、资源使用、调度策略
- **历史上下文**: 历史任务执行结果、模式、经验
- **环境上下文**: 外部环境信息、时间、资源可用性

#### 1.3 实现机制
```go
type AgentContext struct {
    // 任务上下文
    TaskContext *TaskContext
    
    // 依赖上下文
    DependencyContext []*DependencyContext
    
    // 系统上下文
    SystemContext *SystemContext
    
    // 历史上下文
    HistoryContext *HistoryContext
    
    // 环境上下文
    EnvironmentContext *EnvironmentContext
}

type TaskContext struct {
    TaskID      string
    ContractID  string
    Input       map[string]any
    Status      TaskStatus
    Metadata    map[string]string
    CreatedAt   time.Time
    StartedAt   *time.Time
}

type DependencyContext struct {
    TaskID      string
    Status      TaskStatus
    Result      *TaskResult
    Metadata    map[string]string
}

type SystemContext struct {
    SchedulerStatus  string
    ResourceUsage   *ResourceUsage
    QueueSize       int
    ActiveTasks     int
}

type HistoryContext struct {
    RelatedTasks    []*TaskResult
    ExecutionPatterns []ExecutionPattern
    SuccessRate     float64
}

type EnvironmentContext struct {
    Timestamp       time.Time
    AvailableResources *ResourceAvailability
    ExternalSignals []ExternalSignal
}
```

#### 1.4 上下文传递策略
- **主动推送**: 系统主动推送相关上下文给 Agent
- **按需查询**: Agent 可以查询更多上下文信息
- **增量更新**: 上下文变化时及时通知 Agent
- **隐私保护**: 敏感信息需要权限控制

### 2. More Action - 更多的行动能力

#### 2.1 设计原则
Agent 应该具备执行复杂操作的能力，而不是只能执行简单的原子操作。系统应该提供丰富的工具和接口。

#### 2.2 行动维度
- **调度操作**: 提交新任务、取消任务、查询状态
- **数据操作**: 读写数据、查询历史、分析模式
- **协作操作**: 调用其他 Agent、请求人类介入
- **自适应操作**: 动态调整策略、学习优化

#### 2.3 实现机制
```go
type AgentAction interface {
    // 调度操作
    SubmitTask(task *AgentTask) error
    CancelTask(taskID string) error
    QueryTaskStatus(taskID string) (TaskStatus, error)
    
    // 数据操作
    ReadData(key string) (any, error)
    WriteData(key string, value any) error
    QueryHistory(filter *HistoryFilter) ([]*TaskResult, error)
    
    // 协作操作
    CallAgent(agentID string, input map[string]any) (*TaskResult, error)
    CallHuman(request *HumanCallRequest) (*HumanCallResult, error)
    
    // 自适应操作
    AdjustStrategy(strategy *Strategy) error
    ReportObservation(observation *Observation) error
}
```

#### 2.4 工具生态系统
- **内置工具**: 系统提供的核心工具集
- **自定义工具**: Agent 可以注册自己的工具
- **工具组合**: 支持工具的链式调用和组合
- **权限管理**: 工具访问需要权限控制

### 3. Zero Control - 零控制

#### 3.1 设计原则
系统不对 Agent 的决策和执行进行控制，让 Agent 完全自主。系统只提供基础设施和契约约束。

#### 3.2 零控制的含义
- **决策自主**: Agent 自主决定如何完成任务
- **执行自主**: Agent 自主选择执行路径
- **学习自主**: Agent 自主学习和优化
- **演化自主**: Agent 可以自我演化

#### 3.3 系统职责
虽然不对 Agent 进行控制，但系统仍有明确职责：
- **基础设施**: 提供调度、存储、通信等基础能力
- **契约约束**: 通过契约定义输入输出规范
- **资源管理**: 管理系统资源，防止滥用
- **可观测性**: 提供完整的执行日志和监控

#### 3.4 实现机制
```go
type AgentExecution struct {
    // Agent 完全自主执行
    Execute(context *AgentContext, actions AgentAction) (*TaskResult, error)
}

// 系统只提供契约验证
type ContractValidator struct {
    ValidateInput(contractID string, input map[string]any) error
    ValidateOutput(contractID string, output map[string]any) error
}

// 系统只提供基础设施
type Infrastructure struct {
    Scheduler    Scheduler
    StateStore   StateStore
    Communicator Communicator
}
```

### 4. 与传统调度的对比

| 维度 | 传统调度 | Agent 原生调度 |
|------|---------|-------------|
| **上下文** | 最小必要信息 | 丰富的多维上下文 |
| **行动** | 预定义的有限操作 | 丰富的工具和自主能力 |
| **控制** | 严格控制执行流程 | 零控制，完全自主 |
| **适应性** | 静态配置 | 动态自适应 |
| **学习** | 无学习能力 | 持续学习优化 |
| **协作** | 独立执行 | 主动协作 |

### 5. 设计原则总结

#### 5.1 信任原则
- 相信 Agent 的智能决策能力
- 不预设 Agent 的行为模式
- 允许 Agent 犯错并从中学习
- 信任不是静态授予，而是基于可验证胜任力逐步赢得

#### 5.2 透明原则
- 系统行为完全透明可观测
- Agent 决策过程可追踪
- 执行结果可审计

#### 5.3 契约原则
- 通过契约定义边界，而非通过控制
- 输入输出契约是唯一的硬约束
- 契约之外完全自由

#### 5.4 演化原则
- 系统支持 Agent 的持续演化
- Agent 可以学习和改进
- 系统本身也可以演化

### 6. 实施路径

#### 里程碑1：基础框架
- 实现基本的上下文传递机制
- 提供核心行动接口
- 确立契约约束边界

#### 里程碑2：丰富能力
- 扩展上下文维度
- 增加工具生态系统
- 完善可观测性

#### 里程碑3：智能演化
- 实现 Agent 学习机制
- 支持策略自适应
- 建立演化框架

### 7. 风险与缓解

#### 7.1 潜在风险
- **不可预测性**: Agent 行为难以预测
- **资源滥用**: Agent 可能滥用系统资源
- **安全风险**: 自主性可能带来安全隐患

#### 7.2 缓解措施
- **契约约束**: 通过契约限制输入输出
- **资源配额**: 为每个 Agent 设置资源配额
- **递进自主权**: 根据 Agent 的可靠表现逐步扩大或收缩自主空间
- **审计日志**: 完整记录所有操作
- **熔断机制**: 异常情况下的熔断保护

### 8. 设计验证

#### 8.1 验证指标
- **上下文丰富度**: 上下文信息的完整性和相关性
- **行动多样性**: Agent 可执行操作的多样性
- **自主性程度**: Agent 决策和执行的自主程度
- **适应性能力**: Agent 适应环境变化的能力
- **学习效果**: Agent 学习和优化的效果

#### 8.2 验证方法
- 模拟测试
- A/B 测试
- 真实场景验证
- 长期观察

## 结论

"More Context, More Action, Zero Control" 是 Axis 系统的核心设计思想。通过给 Agent 提供丰富的上下文、强大的行动能力，同时保持零控制，系统能够充分发挥 Agent 的智能和自主性，实现真正的 Agent 原生调度。

这个设计思想不仅适用于任务调度，也为未来构建更复杂的 Agent 协作系统奠定了基础。
