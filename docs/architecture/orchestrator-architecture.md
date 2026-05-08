# 主 Agent 管控者架构设计

## 设计目标

主 Agent 角色升级：从「执行者」变为「管控者」。主 Agent 仅负责 3 件事，其余全流程自动化，彻底消除调度瓶颈：
1. 契约制定与 DAG 编排，定义不可突破的规则与流程
2. 异常兜底：仅处理子 Agent 无法自动解决的超时、熔断、合规冲突
3. 人类节点调度：触发必须人类介入的 P0 级审批卡点

## 核心约束

**绝对不介入原则**：
- 主 Agent 绝对不介入子 Agent 的具体执行细节
- 不做指令中转
- 所有子 Agent 间的通信通过全局共享层点对点完成
- 无需主 Agent 转发

## 1. 主 Agent 管控者架构

### 1.1 管控者职责

```go
type Orchestrator struct {
    // 契约管理器
    contractManager *ContractManager

    // DAG 编排器
    dagOrchestrator *DAGOrchestrator

    // 异常处理器
    exceptionHandler *ExceptionHandler

    // 人类节点调度器
    humanScheduler *HumanScheduler

    // 全局共享层
    sharedLayer *SharedLayer

    // 观察者（仅观察，不干预）
    observer *Observer
}
```

### 1.2 三大核心职责

#### 职责 1：契约制定与 DAG 编排

```go
func (o *Orchestrator) Orchestrate(workflow *WorkflowDefinition) (*OrchestrationResult, error) {
    // 1. 制定契约
    contracts := o.contractManager.DefineContracts(workflow)

    // 2. 构建 DAG
    dag, err := o.dagOrchestrator.BuildDAG(contracts)
    if err != nil {
        return nil, err
    }

    // 3. 提交到调度器（不执行，只编排）
    result := &OrchestrationResult{
        WorkflowID:    workflow.ID,
        Contracts:     contracts,
        DAG:           dag,
        SharedLayerID: o.sharedLayer.Create(workflow.ID),
    }

    return result, nil
}
```

#### 职责 2：异常兜底

```go
func (o *Orchestrator) HandleException(event *ExceptionEvent) error {
    // 仅处理子 Agent 无法自动解决的异常
    switch event.Severity {
    case SeverityP0:
        // P0 级异常：需要人类介入
        return o.humanScheduler.Trigger(event)

    case SeverityP1:
        // P1 级异常：自动熔断
        return o.exceptionHandler.CircuitBreak(event)

    case SeverityP2:
        // P2 级异常：自动重试
        return o.exceptionHandler.Retry(event)

    default:
        // 其他异常：记录日志，不干预
        return o.observer.Record(event)
    }
}
```

#### 职责 3：人类节点调度

```go
func (o *Orchestrator) ScheduleHumanNode(approval *ApprovalRequest) error {
    // 仅调度 P0 级审批卡点
    if approval.Priority != PriorityP0 {
        return fmt.Errorf("only P0 approvals require orchestrator intervention")
    }

    // 触发人类节点
    return o.humanScheduler.Trigger(approval)
}
```

## 2. 全局共享层

### 2.1 设计目标

实现子 Agent 间点对点通信，无需主 Agent 转发。

### 2.2 共享层架构

```go
type SharedLayer struct {
    // 共享状态存储
    stateStore StateStore

    // 事件总线
    eventBus EventBus

    // 数据管道
    dataPipeline DataPipeline

    // 锁服务
    lockService LockService
}
```

### 2.3 共享状态存储

```go
type StateStore interface {
    // 写入状态
    Set(key string, value interface{}, ttl time.Duration) error

    // 读取状态
    Get(key string) (interface{}, error)

    // 删除状态
    Delete(key string) error

    // 原子操作
    AtomicSet(key string, oldValue, newValue interface{}) (bool, error)

    // 批量操作
    Batch(operations []Operation) error
}

type MemoryStateStore struct {
    data map[string]*StateItem
    mu   sync.RWMutex
    ttl  time.Duration
}

type StateItem struct {
    Value      interface{}
    Expiration time.Time
}
```

### 2.4 事件总线

```go
type EventBus interface {
    // 发布事件
    Publish(topic string, event *Event) error

    // 订阅事件
    Subscribe(topic string, handler EventHandler) error

    // 取消订阅
    Unsubscribe(topic string, handler EventHandler) error

    // 广播事件
    Broadcast(event *Event) error
}

type MemoryEventBus struct {
    subscribers map[string][]EventHandler
    mu          sync.RWMutex
}

type Event struct {
    ID        string
    Type      EventType
    Source    string
    Target    string // 空字符串表示广播
    Payload   interface{}
    Timestamp time.Time
}

type EventType string

const (
    EventTypeTaskCompleted   EventType = "task_completed"
    EventTypeTaskFailed      EventType = "task_failed"
    EventTypeDataAvailable   EventType = "data_available"
    EventTypeApprovalNeeded  EventType = "approval_needed"
    EventTypeException      EventType = "exception"
)
```

### 2.5 数据管道

```go
type DataPipeline interface {
    // 创建管道
    CreatePipeline(pipelineID string) error

    // 写入数据
    Write(pipelineID string, data interface{}) error

    // 读取数据
    Read(pipelineID string) (<-chan interface{}, error)

    // 关闭管道
    ClosePipeline(pipelineID string) error
}

type MemoryDataPipeline struct {
    pipelines map[string]*Pipeline
    mu        sync.RWMutex
}

type Pipeline struct {
    ID     string
    Buffer chan interface{}
    Closed bool
}
```

### 2.6 锁服务

```go
type LockService interface {
    // 获取锁
    Acquire(lockID string, ttl time.Duration) (bool, error)

    // 释放锁
    Release(lockID string) error

    // 尝试获取锁
    TryAcquire(lockID string, ttl time.Duration) (bool, error)
}

type MemoryLockService struct {
    locks map[string]*LockItem
    mu    sync.RWMutex
}

type LockItem struct {
    Owner      string
    Expiration time.Time
}
```

## 3. 子 Agent 间通信模式

### 3.1 点对点通信

```go
// Agent A 写入共享状态
func (a *SubAgent) ProduceResult(result interface{}) error {
    sharedLayerID := a.Context.SharedLayerID
    key := fmt.Sprintf("%s:agent_%s:result", sharedLayerID, a.ID)

    return a.sharedLayer.stateStore.Set(key, result, time.Hour)
}

// Agent B 读取共享状态
func (b *SubAgent) ConsumeDependency(agentID string) (interface{}, error) {
    sharedLayerID := b.Context.SharedLayerID
    key := fmt.Sprintf("%s:agent_%s:result", sharedLayerID, agentID)

    return b.sharedLayer.stateStore.Get(key)
}
```

### 3.2 事件驱动通信

```go
// Agent A 发布完成事件
func (a *SubAgent) NotifyCompletion() error {
    event := &Event{
        ID:        generateEventID(),
        Type:      EventTypeTaskCompleted,
        Source:    a.ID,
        Target:    "", // 广播
        Payload:   a.Result,
        Timestamp: time.Now(),
    }

    return a.sharedLayer.eventBus.Publish("task_events", event)
}

// Agent B 订阅事件
func (b *SubAgent) WaitForDependency(agentID string) error {
    handler := func(event *Event) {
        if event.Source == agentID && event.Type == EventTypeTaskCompleted {
            b.Dependencies[agentID] = event.Payload
        }
    }

    return b.sharedLayer.eventBus.Subscribe("task_events", handler)
}
```

### 3.3 流式数据传输

```go
// Agent A 流式输出数据
func (a *SubAgent) StreamData(dataStream <-chan interface{}) error {
    pipelineID := fmt.Sprintf("%s:agent_%s:stream", a.Context.SharedLayerID, a.ID)
    err := a.sharedLayer.dataPipeline.CreatePipeline(pipelineID)
    if err != nil {
        return err
    }

    go func() {
        for data := range dataStream {
            a.sharedLayer.dataPipeline.Write(pipelineID, data)
        }
        a.sharedLayer.dataPipeline.ClosePipeline(pipelineID)
    }()

    return nil
}

// Agent B 流式读取数据
func (b *SubAgent) ConsumeStream(agentID string) (<-chan interface{}, error) {
    pipelineID := fmt.Sprintf("%s:agent_%s:stream", b.Context.SharedLayerID, agentID)
    return b.sharedLayer.dataPipeline.Read(pipelineID)
}
```

## 4. 异常兜底机制

### 4.1 异常分类

```go
type ExceptionSeverity string

const (
    SeverityP0 ExceptionSeverity = "P0" // 需要人类介入
    SeverityP1 ExceptionSeverity = "P1" // 自动熔断
    SeverityP2 ExceptionSeverity = "P2" // 自动重试
    SeverityP3 ExceptionSeverity = "P3" // 记录日志
)
```

### 4.2 异常检测

```go
type ExceptionDetector struct {
    rules []ExceptionRule
}

type ExceptionRule struct {
    Condition string // 检测条件
    Severity  ExceptionSeverity
    Action    ExceptionAction
}

func (d *ExceptionDetector) Detect(event *Event) *Exception {
    for _, rule := range d.rules {
        if d.matchCondition(event, rule.Condition) {
            return &Exception{
                Event:    event,
                Severity: rule.Severity,
                Action:   rule.Action,
            }
        }
    }
    return nil
}
```

### 4.3 异常处理

```go
type ExceptionHandler struct {
    circuitBreaker *CircuitBreaker
    retryPolicy    *RetryPolicy
}

func (h *ExceptionHandler) Handle(exception *Exception) error {
    switch exception.Severity {
    case SeverityP1:
        return h.circuitBreaker.Trigger(exception.Event)

    case SeverityP2:
        return h.retryPolicy.Retry(exception.Event)

    default:
        return nil
    }
}
```

## 5. 人类节点调度

### 5.1 P0 级审批定义

```go
type ApprovalRequest struct {
    // 审批 ID
    ApprovalID string

    // 审批优先级
    Priority ApprovalPriority

    // 审批类型
    Type ApprovalType

    // 审批内容
    Content *ApprovalContent

    // 关联的契约
    ContractID string

    // 超时时间
    Timeout time.Duration
}

type ApprovalPriority string

const (
    PriorityP0 ApprovalPriority = "P0" // 必须人类介入
    PriorityP1 ApprovalPriority = "P1" // 可自动处理
)

type ApprovalType string

const (
    ApprovalTypeProdDeploy     ApprovalType = "prod_deploy"     // 生产部署
    ApprovalTypeSecurity      ApprovalType = "security"        // 安全审批
    ApprovalTypeCompliance    ApprovalType = "compliance"      // 合规审批
    ApprovalTypeBudget        ApprovalType = "budget"          // 预算审批
)
```

### 5.2 人类节点调度器

```go
type HumanScheduler struct {
    humanExecutor *HumanExecutor
    approvalQueue *ApprovalQueue
}

func (s *HumanScheduler) Trigger(approval *ApprovalRequest) error {
    // 仅处理 P0 级审批
    if approval.Priority != PriorityP0 {
        return fmt.Errorf("only P0 approvals require human intervention")
    }

    // 转换为 Human Call
    humanCall := &HumanCallRequest{
        CallID:   approval.ApprovalID,
        CallType: CallType(approval.Type),
        Parameters: map[string]interface{}{
            "content": approval.Content,
        },
        Timeout: approval.Timeout,
        Priority: 10, // 最高优先级
    }

    // 提交到人类执行器
    return s.humanExecutor.ExecuteCall(humanCall)
}
```

## 6. 观察者模式

### 6.1 观察者职责

```go
type Observer struct {
    logger    Logger
    metrics   MetricsCollector
    auditor   Auditor
}

func (o *Observer) Observe(event *Event) {
    // 记录日志
    o.logger.Log(event)

    // 收集指标
    o.metrics.Record(event)

    // 审计追踪
    if o.auditRequired(event) {
        o.auditor.Audit(event)
    }
}
```

### 6.2 观察范围

```go
func (o *Observer) auditRequired(event *Event) bool {
    // 需要审计的事件类型
    auditRequiredEvents := map[EventType]bool{
        EventTypeApprovalNeeded: true,
        EventTypeException:      true,
        EventTypeTaskFailed:      true,
    }

    return auditRequiredEvents[event.Type]
}
```

## 7. 工作流程

### 7.1 完整工作流

```
1. 主 Agent（管控者）
   ├─ 制定契约
   ├─ 构建 DAG
   ├─ 创建全局共享层
   └─ 提交到调度器

2. 调度器
   ├─ 执行 DAG 调度
   ├─ 启动子 Agent
   └─ 监控执行状态

3. 子 Agent（执行者）
   ├─ 从共享层读取依赖数据
   ├─ 执行任务
   ├─ 写入结果到共享层
   └─ 发布完成事件

4. 其他子 Agent
   ├─ 订阅事件
   ├─ 从共享层读取数据
   └─ 继续执行

5. 异常处理
   ├─ 检测异常
   ├─ P0 级：主 Agent 触发人类节点
   ├─ P1 级：自动熔断
   └─ P2 级：自动重试

6. 观察者
   ├─ 记录日志
   ├─ 收集指标
   └─ 审计追踪
```

### 7.2 通信流程图

```
子 Agent A                    全局共享层                    子 Agent B
    │                              │                              │
    │  写入结果                     │                              │
    ├────────────────────────────>│                              │
    │                              │                              │
    │  发布完成事件                 │                              │
    ├────────────────────────────>│                              │
    │                              │  通知订阅者                   │
    │                              ├────────────────────────────>│
    │                              │                              │
    │                              │  读取结果                     │
    │                              │<────────────────────────────┤
    │                              │                              │
```

## 8. 性能优化

### 8.1 共享层性能

- 使用内存存储（低延迟）
- 事件总线使用 Channel（高吞吐）
- 数据管道使用缓冲 Channel（流式传输）

### 8.2 并发控制

- 锁服务防止竞争条件
- 原子操作保证一致性
- 批量操作减少网络开销

### 8.3 资源管理

- TTL 自动清理过期数据
- 管道自动关闭
- 订阅自动取消

## 9. 与现有架构的集成

### 9.1 与契约层集成

```go
// 契约中定义共享层配置
type AgentContract struct {
    // ... 其他字段

    // 共享层配置
    SharedLayerConfig *SharedLayerConfig `json:"shared_layer_config,omitempty"`
}

type SharedLayerConfig struct {
    // 需要共享的数据字段
    SharedFields []string `json:"shared_fields"`

    // 事件订阅配置
    EventSubscriptions []EventSubscription `json:"event_subscriptions"`

    // 数据管道配置
    DataPipelineConfig *DataPipelineConfig `json:"data_pipeline_config,omitempty"`
}
```

### 9.2 与 DAG 调度集成

```go
// DAG 节点执行时自动创建共享层
func (s *DAGScheduler) executeNode(ctx context.Context, node *DAGNode) *NodeResult {
    // 获取或创建共享层
    sharedLayer := s.sharedLayerManager.GetOrCreate(node.Context.WorkflowID)

    // 注入到子 Agent
    node.Task.Context.SharedLayer = sharedLayer

    // 执行节点
    return s.executeNodeWithSharedLayer(ctx, node, sharedLayer)
}
```

## 10. 三层状态存储架构

### 10.1 设计目标

实现「单一可信源全局状态共享层」，消除信息不一致，同时保持核心模块零外部依赖。

**核心收益**：
- 彻底消除主 Agent 的上下文传递开销，调度瓶颈完全解除
- 所有子 Agent 拿到的永远是最新的、唯一的信息
- 子 Agent 间可点对点通信，无需主 Agent 中转

### 10.2 三层架构设计

为平衡零依赖和外部集成需求，采用分层架构：

#### Layer 1：核心内存存储（必选，里程碑 1）
- **实现**：纯内存状态存储
- **依赖**：零外部依赖
- **功能**：基础读写、TTL、原子操作
- **适用**：所有项目，满足基本需求

#### Layer 2：可插拔持久化适配器（可选，里程碑 2）
- **实现**：文件系统、SQLite 等本地持久化
- **依赖**：Go 标准库 + 可选轻量级依赖
- **功能**：进程重启后状态恢复
- **适用**：需要状态持久化的项目

#### Layer 3：外部仓库集成（可选，里程碑 3）
- **实现**：Harness、Git 仓库等外部集成
- **依赖**：外部 SDK（放在 pkg/）
- **功能**：与外部系统同步
- **适用**：企业级项目，需要跨系统协作

### 10.3 适配器模式设计

```go
// 核心接口（internal/kernel/shared_layer/state_store.go）
type StateStore interface {
    Set(key string, value interface{}, ttl time.Duration) error
    Get(key string) (interface{}, error)
    Delete(key string) error
    AtomicSet(key string, oldValue, newValue interface{}) (bool, error)
    Batch(operations []Operation) error
}

// 核心实现（Layer 1）
type MemoryStateStore struct {
    data map[string]*StateItem
    mu   sync.RWMutex
}

// 适配器接口（可选）
type StateStoreAdapter interface {
    StateStore
    SyncToRemote(key string) error
    SyncFromRemote(key string) error
}

// 文件系统适配器（Layer 2）
type FileStateStoreAdapter struct {
    local     *MemoryStateStore
    filePath  string
    syncPolicy SyncPolicy
}

// Harness 适配器（Layer 3）
type HarnessStateStoreAdapter struct {
    local         *MemoryStateStore
    harnessClient *HarnessClient  // 外部依赖
    syncPolicy    SyncPolicy
}
```

### 10.4 同步策略

#### Write-Through（写穿透）
- 写操作同时写本地和远程
- 一致性高，延迟高
- 适用于强一致性场景

#### Write-Back（写回）
- 写操作先写本地，异步同步到远程
- 性能高，一致性弱
- 适用于高并发场景

#### Write-Around（绕写）
- 写操作只写远程，读操作缓存到本地
- 适用于读多写少场景

```go
type SyncPolicy string

const (
    SyncPolicyWriteThrough SyncPolicy = "write_through"
    SyncPolicyWriteBack    SyncPolicy = "write_back"
    SyncPolicyWriteAround  SyncPolicy = "write_around"
)
```

### 10.5 降级策略

```go
type StateStoreManager struct {
    primary       StateStore
    fallback      StateStore
    healthChecker HealthChecker
}

func (m *StateStoreManager) Get(key string) (interface{}, error) {
    // 尝试主存储
    val, err := m.primary.Get(key)
    if err == nil {
        return val, nil
    }

    // 降级到备用存储
    if m.healthChecker.IsHealthy(m.fallback) {
        return m.fallback.Get(key)
    }

    return nil, ErrAllStoresUnavailable
}
```

### 10.6 配置化选择

```go
type SharedLayerConfig struct {
    // Layer 1 配置
    MemoryStoreEnabled bool `json:"memory_store_enabled"`
    MemoryStoreSize    int  `json:"memory_store_size_mb"`

    // Layer 2 配置
    FileStoreEnabled bool   `json:"file_store_enabled"`
    FileStorePath    string `json:"file_store_path"`
    SQLiteEnabled    bool   `json:"sqlite_enabled"`
    SQLitePath       string `json:"sqlite_path"`

    // Layer 3 配置
    HarnessEnabled    bool        `json:"harness_enabled"`
    HarnessEndpoint  string      `json:"harness_endpoint"`
    HarnessAPIKey    string      `json:"harness_api_key"`
    SyncPolicy       SyncPolicy  `json:"sync_policy"`

    // 降级配置
    FallbackEnabled   bool `json:"fallback_enabled"`
    HealthCheckInterval int `json:"health_check_interval_seconds"`
}
```

### 10.7 数据一致性保障

#### 版本号机制
每次状态更新携带版本号，冲突检测：

```go
type StateItem struct {
    Value      interface{}
    Version    int64
    Expiration time.Time
}

func (s *MemoryStateStore) SetWithVersion(key string, value interface{}, expectedVersion int64) (bool, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    item, exists := s.data[key]
    if exists && item.Version != expectedVersion {
        return false, ErrVersionConflict
    }

    newVersion := int64(0)
    if exists {
        newVersion = item.Version + 1
    }

    s.data[key] = &StateItem{
        Value:      value,
        Version:    newVersion,
        Expiration: time.Time{},
    }

    return true, nil
}
```

#### 最终一致性
接受短暂不一致，通过重试机制保证最终一致：

```go
type EventualConsistencyManager struct {
    retryQueue chan *SyncOperation
    maxRetries int
}

func (m *EventualConsistencyManager) SyncWithRetry(op *SyncOperation) error {
    for i := 0; i < m.maxRetries; i++ {
        err := op.Execute()
        if err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return ErrMaxRetriesExceeded
}
```

### 10.8 潜在风险与缓解

#### 风险 1：外部仓库集成违反零依赖
- **缓解**：外部集成放在 pkg/ 可选模块
- **缓解**：核心模块通过接口抽象，不依赖具体实现
- **缓解**：适配器模式保持解耦

#### 风险 2：数据一致性挑战
- **缓解**：版本号机制检测冲突
- **缓解**：最终一致性 + 重试机制
- **缓解**：TTL 自动清理过期数据
- **缓解**：降级策略保证可用性

#### 风险 3：权限管控复杂度
- **缓解**：基于契约的权限绑定
- **缓解**：细粒度 ACL（访问控制列表）
- **缓解**：审计日志记录所有访问

### 10.9 实施路线

#### 里程碑 1：Layer 1
- [ ] 实现内存状态存储
- [ ] 实现内存事件总线
- [ ] 实现内存数据管道
- [ ] 实现内存锁服务
- [ ] 验证子 Agent 点对点通信
- [ ] 不集成外部系统，保持零依赖

#### 里程碑 2：Layer 2
- [ ] 实现文件系统适配器
- [ ] 实现 SQLite 适配器（可选）
- [ ] 实现状态持久化
- [ ] 实现崩溃恢复
- [ ] 实现版本号机制

#### 里程碑 3：Layer 3
- [ ] 实现 Harness 适配器
- [ ] 实现同步策略（write-through/write-back）
- [ ] 实现降级策略
- [ ] 实现健康检查
- [ ] 实现最终一致性管理

## 11. 实施路线图

### 里程碑 1：共享层基础
- [ ] 实现内存状态存储
- [ ] 实现内存事件总线
- [ ] 实现内存数据管道
- [ ] 实现内存锁服务
- [ ] 验证子 Agent 间点对点通信

### 里程碑 2：管控者架构
- [ ] 实现主 Agent 管控者
- [ ] 实现异常检测与处理
- [ ] 实现 P0 级人类节点调度
- [ ] 实现观察者模式
- [ ] 验证管控者不介入执行细节

### 里程碑 3：生态成熟
- [ ] 实现持久化共享层（Redis）
- [ ] 实现分布式事件总线
- [ ] 实现分布式锁服务
- [ ] 实现共享层监控
- [ ] 实现智能异常预测

## 11. 核心优势

### 11.1 消除调度瓶颈
- 主 Agent 不做指令中转
- 子 Agent 点对点通信
- 并行度最大化

### 11.2 职责清晰
- 主 Agent：管控者
- 子 Agent：执行者
- 共享层：通信基础设施

### 11.3 可扩展性
- 新增子 Agent 无需修改主 Agent
- 共享层支持水平扩展
- 事件驱动松耦合

### 11.4 可观测性
- 观察者模式全链路追踪
- 异常自动检测
- 审计完整记录
