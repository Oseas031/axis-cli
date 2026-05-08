# Axis CLI 核心模块架构设计（里程碑1简化版）

## 模块铁律
核心模块与可选模块实现 100% 解耦，可选模块的迭代不得修改核心内核的 Agent 原生设计，不得破坏内核的调度语义与函数调用规范。

## 里程碑1设计原则（奥卡姆剃刀）
- **最小可行**：只实现验证核心概念所需的最小功能集
- **渐进增强**：高级特性（如准入规则、SLA、条件依赖）在后续里程碑添加
- **聚焦核心**：里程碑1聚焦于"任务调度 + 简单依赖管理"的可行性验证

## 1. Agent 原生调度内核 (internal/kernel)

### 1.1 调度器 (scheduler)
**职责**：Agent 任务调度核心引擎

**核心功能**：
- Agent 任务队列管理（FIFO）
- 任务串行调度
- 简单依赖管理（依赖任务完成后才执行后续任务）
- 任务生命周期管理（pending/running/completed/failed）
- 循环依赖检测

**关键接口**：
```go
type Scheduler interface {
    Submit(task *AgentTask) error
    Cancel(taskID string) error
    GetStatus(taskID string) TaskStatus
}
```

**设计约束**：
- 无外部依赖（纯 Go 标准库）
- 调度语义不可被可选模块修改

### 1.2 分发器 (dispatcher)
**职责**：任务分发与执行路由

**核心功能**：
- 任务到执行器的路由映射
- Human-as-a-Function 调用路由
- 基础超时控制

**关键接口**：
```go
type Dispatcher interface {
    Dispatch(task *AgentTask) (*TaskResult, error)
}
```

### 1.3 生命周期管理 (lifecycle)
**职责**：Agent 与任务全生命周期管控

**核心功能**：
- 任务状态机管理
- 优雅关闭机制

**关键接口**：
```go
type LifecycleManager interface {
    Shutdown(ctx context.Context) error
}
```

### 1.4 共享状态存储 (shared_layer/state_store)
**职责**：任务状态持久化

**核心功能**：
- 任务状态存储
- 状态查询

**关键接口**：
```go
type StateStore interface {
    Save(taskID string, state TaskState) error
    Load(taskID string) (TaskState, error)
}
```

## 2. Agent 契约层 (internal/contract) - 里程碑1简化版

### 2.1 契约定义（里程碑1最小集）
**职责**：定义 Agent 的输入输出契约

**核心要素**：
1. **输入 Schema**：强类型、无歧义的入参结构
2. **输出 Schema**：标准化的出参结构

**数据结构**：
```go
type AgentContract struct {
    ContractID   string       `json:"contract_id"`
    InputSchema   *InputSchema  `json:"input_schema"`
    OutputSchema  *OutputSchema `json:"output_schema"`
}

type InputSchema struct {
    Fields []FieldDef `json:"fields"`
}

type OutputSchema struct {
    Fields []FieldDef `json:"fields"`
}
```

**里程碑1暂不包含**：
- SLA 约定（超时、重试、熔断）
- 准入规则（本地/远程校验）
- 异常码体系
- 依赖管理
- 版本管理

### 2.2 契约执行器（简化版）
**职责**：基于契约的 Agent 执行引擎

**核心功能**：
- 输入 Schema 验证
- 输出 Schema 验证

**关键接口**：
```go
type ContractExecutor interface {
    Execute(contractID string, input map[string]interface{}) (*ExecutionResult, error)
    ValidateInput(contractID string, input map[string]interface{}) error
    ValidateOutput(contractID string, output map[string]interface{}) error
}
```

## 3. Human-as-a-Function 核心模块 (internal/human) - 里程碑1简化版

### 3.1 执行器 (executor)
**职责**：人类调用履约执行引擎

**核心功能**：
- 调用请求排队
- 履约状态跟踪
- 结果收集

**关键接口**：
```go
type HumanExecutor interface {
    ExecuteCall(req *HumanCallRequest) (*HumanCallResult, error)
    GetCallStatus(callID string) CallStatus
}
```

**里程碑1暂不包含**：
- 复杂协议层（使用简单JSON协议）
- 解析器（使用简单字符串匹配）

## 4. 依赖关系（里程碑1简化版）

```
internal/kernel (调度内核)
    ├── scheduler (调度器) - FIFO + 简单依赖管理
    ├── dispatcher (分发器) - 基础任务分发
    ├── lifecycle (生命周期) - 基础生命周期管理
    └── shared_layer/state_store (状态存储)

internal/contract (Agent 契约层)
    └── executor (契约执行器) - 输入输出验证

internal/human (Human-as-a-Function)
    └── executor (执行器) - 基础任务队列

依赖关系：
- scheduler → lifecycle, state_store
- dispatcher → contract.executor, human.executor
```

## 核心模块技术约束（里程碑1）

1. **零外部依赖**：核心模块只依赖 Go 标准库
2. **接口隔离**：每个模块通过清晰接口通信
3. **并发安全**：所有公共接口必须线程安全
4. **可测试性**：核心模块必须有单元测试覆盖
5. **向后兼容**：核心接口变更必须保证向后兼容

## 里程碑1范围总结

**包含**：
- ✅ FIFO 任务调度
- ✅ 简单依赖管理（依赖完成才执行）
- ✅ 输入输出 Schema 验证
- ✅ 基础 Human Task Queue
- ✅ 简单状态存储

**不包含（延后到里程碑2+）**：
- ❌ DAG 并行调度
- ❌ 契约准入规则（本地/远程校验）
- ❌ SLA 约定（超时、重试、熔断）
- ❌ 异常码体系
- ❌ 契约版本管理
- ❌ 工具调用层
- ❌ 全局事件总线
- ❌ 管控者架构
