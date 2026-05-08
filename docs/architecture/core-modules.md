# Axis CLI 核心模块架构设计

## 模块铁律
核心模块与可选模块实现 100% 解耦，可选模块的迭代不得修改核心内核的 Agent 原生设计，不得破坏内核的调度语义与函数调用规范。

## 1. Agent 原生调度内核 (internal/kernel)

### 1.0 管控者架构（主 Agent）
**职责**：主 Agent 角色升级为「管控者」，仅负责 3 件事

**核心职责**：
1. **契约制定与 DAG 编排**：定义不可突破的规则与流程
2. **异常兜底**：仅处理子 Agent 无法自动解决的超时、熔断、合规冲突
3. **人类节点调度**：触发必须人类介入的 P0 级审批卡点

**核心约束**：
- 绝对不介入子 Agent 的具体执行细节
- 不做指令中转
- 所有子 Agent 间的通信通过全局共享层点对点完成
- 无需主 Agent 转发

**关键组件**：
- 契约管理器
- DAG 编排器
- 异常处理器
- 人类节点调度器
- 全局共享层
- 观察者

**详细设计**：详见 `docs/architecture/orchestrator-architecture.md`

### 1.1 调度器 (scheduler)
**职责**：Agent 任务调度核心引擎

**核心功能**：
- Agent 任务队列管理
- 任务优先级调度算法
- Goroutine 并发调度（支持数万级并发）
- 任务生命周期管理（pending/running/completed/failed）
- 调度策略可插拔设计（FIFO、优先级、DAG 并行调度）

**关键接口**：
```go
type Scheduler interface {
    Submit(task *AgentTask) error
    SubmitBatch(tasks []*AgentTask) error
    Cancel(taskID string) error
    GetStatus(taskID string) TaskStatus
    ListTasks(filter TaskFilter) []*AgentTask
    SetStrategy(strategy SchedulingStrategy) error
}
```

**调度策略**：
- **FIFO**：先进先出，适用于无依赖任务
- **Priority**：基于优先级调度，适用于有优先级要求的任务
- **DAG Parallel**：基于 DAG 的并行调度，适用于有依赖关系的复杂任务流

**设计约束**：
- 无外部依赖（纯 Go 标准库）
- 调度语义不可被可选模块修改
- 支持热插拔调度策略
- DAG 调度策略与契约层依赖解析集成

### 1.2 分发器 (dispatcher)
**职责**：任务分发与执行路由

**核心功能**：
- 任务到执行器的路由映射
- 工具调用分发
- Human-as-a-Function 调用路由
- 超时控制与熔断机制
- 重试策略与幂等性保证

**关键接口**：
```go
type Dispatcher interface {
    Dispatch(task *AgentTask) (*TaskResult, error)
    RegisterTool(name string, handler ToolHandler) error
    RegisterHumanFunction(name string, handler HumanFunctionHandler) error
}
```

### 1.3 生命周期管理 (lifecycle)
**职责**：Agent 与任务全生命周期管控

**核心功能**：
- Agent 注册与注销
- 任务状态机管理
- 资源清理与回收
- 优雅关闭机制
- 崩溃恢复与状态持久化

**关键接口**：
```go
type LifecycleManager interface {
    RegisterAgent(agent *Agent) error
    UnregisterAgent(agentID string) error
    Shutdown(ctx context.Context) error
}
```

## 2. Human-as-a-Function 核心模块 (internal/human)

### 2.1 协议层 (protocol)
**职责**：定义 call human 标准化调用协议

**核心功能**：
- 调用协议定义（入参结构、出参格式）
- 超时管控机制
- 错误码体系定义
- 重试策略规范
- 幂等性保证机制

**协议规范**：
- 详见 `docs/protocols/call-human-spec.md`

### 2.2 执行器 (executor)
**职责**：人类调用履约执行引擎

**核心功能**：
- 调用请求排队
- 交互面板管理
- 履约状态跟踪
- 超时处理与取消
- 结果收集与验证

**关键接口**：
```go
type HumanExecutor interface {
    ExecuteCall(req *HumanCallRequest) (*HumanCallResult, error)
    CancelCall(callID string) error
    GetCallStatus(callID string) CallStatus
}
```

### 2.3 解析器 (parser)
**职责**：人类履约结果结构化解析

**核心功能**：
- 自然语言结果解析
- 结构化数据提取
- 结果验证与校验
- 错误识别与分类
- 解析策略可插拔

**关键接口**：
```go
type ResultParser interface {
    Parse(rawResult string) (*StructuredResult, error)
    Validate(result *StructuredResult, schema *ResultSchema) error
}
```

## 3. Agent 契约层 (internal/contract)

### 3.1 契约注册表 (registry)
**职责**：Agent 契约注册、版本管理与验证

**核心功能**：
- 契约注册与注销
- 契约版本管理
- 契约验证（Schema 验证、依赖检查）
- 契约查询与发现
- 契约状态管理（draft/active/deprecated/retired）

**关键接口**：
```go
type ContractRegistry interface {
    Register(contract *AgentContract) error
    Unregister(contractID string) error
    Get(contractID, version string) (*AgentContract, error)
    GetLatest(contractID string) (*AgentContract, error)
    List(filter ContractFilter) []*AgentContract
    Validate(contract *AgentContract) (*ValidationResult, error)
}
```

### 3.2 契约执行器 (executor)
**职责**：基于契约的 Agent 执行引擎

**核心功能**：
- 契约准入规则评估（本地校验 + 远程校验）
- 契约 SLA 执行（超时、重试、熔断）
- 输入 Schema 验证
- 输出 Schema 验证与验收
- 异常码处理与自动处置
- 与 call human 协议的同构映射

**关键接口**：
```go
type ContractExecutor interface {
    Execute(contractID, version string, input map[string]interface{}) (*ExecutionResult, error)
    ValidateInput(contractID, version string, input map[string]interface{}) error
    ValidateOutput(contractID, version string, output map[string]interface{}) (*ValidationResult, error)
}
```

### 3.3 准入引擎 (admission)
**职责**：契约准入规则评估

**核心功能**：
- 本地规则执行（Schema 验证、架构检查、权限检查）
- 远程规则执行（依赖 Agent 调用）
- 规则组合逻辑（AND/OR）
- 准入失败处理与错误码映射
- 循环依赖检测

**关键接口**：
```go
type AdmissionEngine interface {
    Evaluate(rules *AdmissionRules, input map[string]interface{}, context *ExecutionContext) (*AdmissionResult, error)
    RegisterLocalRule(name string, executor LocalRuleExecutor) error
    RegisterRemoteRule(name string, executor RemoteRuleExecutor) error
}
```

### 3.4 依赖解析器 (dependency)
**职责**：契约依赖管理与解析

**核心功能**：
- 依赖图构建
- 循环依赖检测
- 拓扑排序
- 依赖版本约束解析
- 依赖健康检查

**关键接口**：
```go
type DependencyResolver interface {
    BuildGraph(contracts []*AgentContract) (*DependencyGraph, error)
    DetectCycles() ([]Cycle, error)
    TopologicalSort() ([]string, error)
    ResolveDependencies(contractID string) ([]DependencyDeclaration, error)
}
```

## 4. 统一工具调用层 (internal/tools)

### 4.1 工具注册表 (registry)
**职责**：工具注册与发现

**核心功能**：
- 工具注册与注销
- 工具元数据管理
- 工具权限绑定
- 工具版本管理
- 工具依赖解析

**关键接口**：
```go
type ToolRegistry interface {
    Register(tool *Tool) error
    Unregister(name string) error
    Get(name string) (*Tool, error)
    List(filter ToolFilter) []*Tool
}
```

### 4.2 调用器 (invoker)
**职责**：工具调用执行引擎

**核心功能**：
- 工具调用路由
- 参数验证与转换
- 调用结果标准化
- 调用链追踪
- 调用缓存策略

**关键接口**：
```go
type ToolInvoker interface {
    Invoke(name string, params map[string]interface{}) (*ToolResult, error)
    InvokeChain(chain *ToolChain) (*ChainResult, error)
}
```

## 核心模块依赖关系

```
internal/kernel (调度内核)
    ├── orchestrator (管控者)
    │   ├── contract_manager (契约管理器)
    │   ├── dag_orchestrator (DAG 编排器)
    │   ├── exception_handler (异常处理器)
    │   ├── human_scheduler (人类节点调度器)
    │   └── observer (观察者)
    ├── shared_layer (全局共享层)
    │   ├── state_store (状态存储)
    │   ├── event_bus (事件总线)
    │   ├── data_pipeline (数据管道)
    │   └── lock_service (锁服务)
    ├── scheduler (调度器)
    ├── dispatcher (分发器)
    └── lifecycle (生命周期)

internal/human (Human-as-a-Function)
    ├── protocol (协议层)
    ├── executor (执行器)
    └── parser (解析器)

internal/contract (Agent 契约层)
    ├── registry (契约注册表)
    ├── executor (契约执行器)
    ├── admission (准入引擎)
    └── dependency (依赖解析器)

internal/tools (工具调用层)
    ├── registry (注册表)
    └── invoker (调用器)

依赖关系：
- orchestrator → contract_manager, dag_orchestrator, shared_layer
- dispatcher → tools.invoker, human.executor, contract.executor
- scheduler → lifecycle, shared_layer
- executor → protocol, parser
- invoker → registry
- contract.executor → contract.admission, contract.dependency, shared_layer
- contract.admission → contract.dependency
- contract.executor ↔ human.protocol (同构映射)
- dag_orchestrator → contract.dependency (依赖解析)
- sub_agent → shared_layer (点对点通信)
```

## Agent 契约层设计理念

**核心创新**：主 Agent 的核心职责从「发指令」变为「定契约」，所有子 Agent 封装为标准化无状态函数，与 call human 协议完全同构，彻底消除调度歧义与理解偏差。

**契约 5 要素**：
1. **输入 Schema**：强类型、无歧义的入参结构，仅保留完成该任务的最小必要信息
2. **输出 Schema**：标准化的出参结构，明确验收标准、格式要求、合规规则
3. **SLA 约定**：最长执行超时时间、重试次数、熔断阈值
4. **准入规则**：任务可执行的前置校验条件（架构规范校验、权限检查、依赖解析）
5. **异常码体系**：标准化的错误分类与自动处置规则

**与 call human 协议的同构性**：
- 契约输入 Schema ↔ Call Human Parameters
- 契约输出 Schema ↔ Call Human Result
- 契约 SLA ↔ Call Human Timeout + RetryConfig
- 契约准入规则 ↔ Call Human Context
- 契约异常码 ↔ Call Human ErrorCode + RetryAdvice

**详细设计**：详见 `docs/architecture/agent-contract-design.md`

## DAG 并行调度策略

**核心创新**：构建 DAG 有向无环图并行调度模型，主 Agent 基于任务依赖关系生成全流程 DAG 图，无依赖的任务 100% 并行执行，有依赖的任务流水线化执行，总耗时从「各环节时长之和」压缩为「最长关键路径的单环节时长」。

**DAG 调度原理**：
- 基于契约依赖关系构建 DAG 图
- 拓扑排序确定执行层级
- 同层级节点完全并行执行
- 关键路径计算预估总耗时
- 资源感知调度优化并发度

**性能提升**：
- 理论加速比 = 串行总耗时 / 关键路径耗时
- 实际加速比受限于并行度和资源约束
- 典型场景可达到 2-5 倍加速

**失败处理策略**：
- **Fail Fast**：任一任务失败立即停止
- **Continue**：忽略失败继续执行
- **Best Effort**：跳过依赖失败的任务

**与契约层集成**：
- 契约依赖声明自动映射为 DAG 边
- 契约 SLA 约定用于资源分配
- 契约准入规则用于节点执行前校验

**详细设计**：详见 `docs/architecture/dag-scheduling.md`

## 主 Agent 管控者架构

**核心创新**：主 Agent 角色从「执行者」升级为「管控者」，仅负责 3 件事，其余全流程自动化，彻底消除调度瓶颈。

**三大核心职责**：
1. **契约制定与 DAG 编排**：定义不可突破的规则与流程
2. **异常兜底**：仅处理子 Agent 无法自动解决的超时、熔断、合规冲突
3. **人类节点调度**：触发必须人类介入的 P0 级审批卡点

**绝对不介入原则**：
- 主 Agent 绝对不介入子 Agent 的具体执行细节
- 不做指令中转
- 所有子 Agent 间的通信通过全局共享层点对点完成
- 无需主 Agent 转发

**全局共享层**：
- 共享状态存储（子 Agent 间数据共享）
- 事件总线（发布-订阅模式）
- 数据管道（流式数据传输）
- 锁服务（并发控制）

**核心优势**：
- 消除调度瓶颈（主 Agent 不做指令中转）
- 职责清晰（管控者 vs 执行者）
- 可扩展性（子 Agent 点对点通信）
- 可观测性（观察者模式全链路追踪）

**详细设计**：详见 `docs/architecture/orchestrator-architecture.md`

## 核心模块技术约束

1. **零外部依赖**：核心模块只依赖 Go 标准库
2. **接口隔离**：每个模块通过清晰接口通信
3. **可插拔设计**：策略、解析器等支持热插拔
4. **并发安全**：所有公共接口必须线程安全
5. **可测试性**：核心模块必须有单元测试覆盖
6. **向后兼容**：核心接口变更必须保证向后兼容
7. **调度策略独立**：DAG 调度作为可选策略，不影响基础调度功能
8. **管控者不干预执行**：主 Agent 绝对不介入子 Agent 的具体执行细节
