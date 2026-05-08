# DAG 并行调度策略设计

## 设计目标

构建 DAG 有向无环图并行调度模型，主 Agent 基于任务依赖关系生成全流程 DAG 图，无依赖的任务 100% 并行执行，有依赖的任务流水线化执行，总耗时从「各环节时长之和」压缩为「最长关键路径的单环节时长」。

## 1. 核心概念

### 1.1 DAG 调度原理

**传统串行调度**：
```
任务 A (10s) → 任务 B (5s) → 任务 C (8s) → 任务 D (3s)
总耗时 = 10 + 5 + 8 + 3 = 26s
```

**DAG 并行调度**：
```
任务 A (10s) ─┐
             ├→ 任务 C (8s) → 任务 D (3s)
任务 B (5s) ─┘
总耗时 = max(10, 5) + 8 + 3 = 21s
```

**性能提升**：
- 理论加速比 = 串行总耗时 / 关键路径耗时
- 实际加速比受限于并行度和资源约束

### 1.2 关键路径

**定义**：DAG 中从起点到终点的最长路径，决定了整个任务的最短完成时间。

**计算方法**：动态规划
```
dist[node] = max(dist[dep] + duration[node]) for all dep in dependencies
```

**用途**：
- 预估任务总耗时
- 识别瓶颈任务
- 优化资源分配

## 2. DAG 数据结构

### 2.1 DAG 图

```go
type DAGGraph struct {
    // 节点集合
    nodes map[string]*DAGNode

    // 边集合
    edges map[string][]*DAGEdge

    // 拓扑层级（缓存）
    levels [][]*DAGNode

    // 关键路径（缓存）
    criticalPath []*DAGNode

    // 读写锁
    mu sync.RWMutex
}

type DAGNode struct {
    // 节点 ID（唯一）
    ID string `json:"id"`

    // 关联的任务
    Task *AgentTask `json:"task"`

    // 依赖的节点 ID
    Dependencies []string `json:"dependencies"`

    // 依赖此节点的节点 ID
    Dependents []string `json:"dependents"`

    // 节点状态
    State NodeState `json:"state"`

    // 预估执行时长
    EstimatedDuration time.Duration `json:"estimated_duration"`

    // 实际执行时长
    ActualDuration time.Duration `json:"actual_duration,omitempty"`

    // 开始时间
    StartTime *time.Time `json:"start_time,omitempty"`

    // 结束时间
    EndTime *time.Time `json:"end_time,omitempty"`

    // 错误信息
    Error *ExecutionError `json:"error,omitempty"`

    // 元数据
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type DAGEdge struct {
    // 边 ID
    ID string `json:"id"`

    // 源节点 ID
    From string `json:"from"`

    // 目标节点 ID
    To string `json:"to"`

    // 边类型
    Type EdgeType `json:"type"`

    // 条件（可选，用于条件依赖）
    Condition *EdgeCondition `json:"condition,omitempty"`
}

type EdgeType string

const (
    EdgeTypeHard     EdgeType = "hard"     // 强依赖（必须等待）
    EdgeTypeSoft     EdgeType = "soft"     // 弱依赖（可跳过）
    EdgeTypeConditional EdgeType = "conditional" // 条件依赖
)

type EdgeCondition struct {
    // 条件表达式
    Expression string `json:"expression"`

    // 条件参数
    Params map[string]interface{} `json:"params"`
}

type NodeState string

const (
    NodeStatePending   NodeState = "pending"    // 等待中
    NodeStateReady     NodeState = "ready"      // 就绪（依赖已满足）
    NodeStateRunning   NodeState = "running"    // 执行中
    NodeStateCompleted NodeState = "completed"  // 已完成
    NodeStateFailed    NodeState = "failed"     // 失败
    NodeStateSkipped   NodeState = "skipped"    // 已跳过
    NodeStateCancelled NodeState = "cancelled"  // 已取消
)
```

### 2.2 DAG 构建器

```go
type DAGBuilder struct {
    graph *DAGGraph
}

func NewDAGBuilder() *DAGBuilder {
    return &DAGBuilder{
        graph: &DAGGraph{
            nodes: make(map[string]*DAGNode),
            edges: make(map[string][]*DAGEdge),
        },
    }
}

func (b *DAGBuilder) AddNode(task *AgentTask, estimatedDuration time.Duration) error {
    node := &DAGNode{
        ID:                 task.ID,
        Task:               task,
        Dependencies:       []string{},
        Dependents:         []string{},
        State:              NodeStatePending,
        EstimatedDuration:  estimatedDuration,
    }
    b.graph.nodes[task.ID] = node
    return nil
}

func (b *DAGBuilder) AddEdge(fromID, toID string, edgeType EdgeType, condition *EdgeCondition) error {
    // 检查节点是否存在
    if _, ok := b.graph.nodes[fromID]; !ok {
        return fmt.Errorf("node %s not found", fromID)
    }
    if _, ok := b.graph.nodes[toID]; !ok {
        return fmt.Errorf("node %s not found", toID)
    }

    // 添加边
    edge := &DAGEdge{
        ID:        fmt.Sprintf("%s->%s", fromID, toID),
        From:      fromID,
        To:        toID,
        Type:      edgeType,
        Condition: condition,
    }
    b.graph.edges[fromID] = append(b.graph.edges[fromID], edge)

    // 更新节点依赖关系
    b.graph.nodes[toID].Dependencies = append(b.graph.nodes[toID].Dependencies, fromID)
    b.graph.nodes[fromID].Dependents = append(b.graph.nodes[fromID].Dependents, toID)

    return nil
}

func (b *DAGBuilder) Build() (*DAGGraph, error) {
    // 检测循环依赖
    if err := b.detectCycles(); err != nil {
        return nil, err
    }

    // 计算拓扑层级
    if err := b.calculateLevels(); err != nil {
        return nil, err
    }

    // 计算关键路径
    b.calculateCriticalPath()

    return b.graph, nil
}
```

## 3. DAG 调度器

### 3.1 调度器接口

```go
type DAGScheduler interface {
    // 调度 DAG
    Schedule(graph *DAGGraph) (*ScheduleResult, error)

    // 取消调度
    Cancel(scheduleID string) error

    // 获取调度状态
    GetStatus(scheduleID string) (*ScheduleStatus, error)

    // 注册资源管理器
    RegisterResourceManager(rm ResourceManager) error

    // 设置失败策略
    SetFailureStrategy(strategy FailureStrategy) error
}
```

### 3.2 调度器实现

```go
type DAGSchedulerImpl struct {
    executor        TaskExecutor
    resourceMgr     ResourceManager
    failureStrategy FailureStrategy
    stateStore      StateStore
    metrics         MetricsCollector
    mu              sync.RWMutex
}

func NewDAGScheduler(executor TaskExecutor, rm ResourceManager) *DAGSchedulerImpl {
    return &DAGSchedulerImpl{
        executor:        executor,
        resourceMgr:     rm,
        failureStrategy: StrategyFailFast,
        stateStore:      NewMemoryStateStore(),
        metrics:         NewPrometheusMetrics(),
    }
}
```

### 3.3 核心调度算法

```go
func (s *DAGSchedulerImpl) Schedule(graph *DAGGraph) (*ScheduleResult, error) {
    scheduleID := generateScheduleID()
    result := &ScheduleResult{
        ScheduleID: scheduleID,
        StartTime:  time.Now(),
        Graph:      graph,
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 按拓扑层级调度
    for levelIdx, level := range graph.levels {
        s.metrics.LevelStarted(scheduleID, levelIdx, len(level))

        // 执行当前层级的所有节点
        levelResult := s.executeLevel(ctx, level)
        result.LevelResults = append(result.LevelResults, levelResult)

        // 检查失败策略
        if levelResult.HasFailure() {
            switch s.failureStrategy {
            case StrategyFailFast:
                cancel()
                result.Status = ScheduleStatusFailed
                result.Error = levelResult.FirstError()
                return result, result.Error
            case StrategyContinue:
                // 继续执行下一层级
            case StrategyBestEffort:
                // 跳过依赖失败节点的任务
                s.markSkippedNodes(graph, levelResult.FailedNodes())
            }
        }

        s.metrics.LevelCompleted(scheduleID, levelIdx, levelResult.Duration)
    }

    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Status = ScheduleStatusCompleted

    // 保存状态
    s.stateStore.Save(scheduleID, result)

    return result, nil
}

func (s *DAGSchedulerImpl) executeLevel(ctx context.Context, nodes []*DAGNode) *LevelResult {
    result := &LevelResult{
        StartTime: time.Now(),
    }

    // 资源感知的批次大小
    batchSize := s.calculateBatchSize(nodes)

    // 分批执行
    for i := 0; i < len(nodes); i += batchSize {
        end := min(i+batchSize, len(nodes))
        batch := nodes[i:end]

        batchResult := s.executeBatch(ctx, batch)
        result.Merge(batchResult)

        // 如果快速失败且有错误，立即返回
        if s.failureStrategy == StrategyFailFast && batchResult.HasFailure() {
            break
        }
    }

    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    return result
}

func (s *DAGSchedulerImpl) executeBatch(ctx context.Context, nodes []*DAGNode) *BatchResult {
    result := &BatchResult{
        StartTime: time.Now(),
    }

    var wg sync.WaitGroup
    errChan := make(chan error, len(nodes))
    resultChan := make(chan *NodeResult, len(nodes))

    for _, node := range nodes {
        wg.Add(1)
        go func(n *DAGNode) {
            defer wg.Done()

            // 检查上下文是否已取消
            if ctx.Err() != nil {
                errChan <- ctx.Err()
                return
            }

            // 执行节点
            nodeResult := s.executeNode(ctx, n)
            resultChan <- nodeResult

            if nodeResult.Error != nil {
                errChan <- nodeResult.Error
            }
        }(node)
    }

    wg.Wait()
    close(errChan)
    close(resultChan)

    // 收集结果
    for nodeResult := range resultChan {
        result.NodeResults = append(result.NodeResults, nodeResult)
    }

    // 收集错误
    for err := range errChan {
        result.Errors = append(result.Errors, err)
    }

    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    return result
}

func (s *DAGSchedulerImpl) executeNode(ctx context.Context, node *DAGNode) *NodeResult {
    startTime := time.Now()

    // 更新节点状态
    node.State = NodeStateRunning
    node.StartTime = &startTime

    // 执行任务
    taskResult, err := s.executor.Execute(ctx, node.Task)
    endTime := time.Now()

    // 更新节点状态
    node.EndTime = &endTime
    node.ActualDuration = endTime.Sub(startTime)

    if err != nil {
        node.State = NodeStateFailed
        node.Error = &ExecutionError{
            Code:    "EXECUTION_FAILED",
            Message: err.Error(),
        }
        return &NodeResult{
            NodeID:   node.ID,
            Success:  false,
            Error:    err,
            Duration: node.ActualDuration,
        }
    }

    node.State = NodeStateCompleted
    return &NodeResult{
        NodeID:   node.ID,
        Success:  true,
        Result:   taskResult,
        Duration: node.ActualDuration,
    }
}
```

### 3.4 资源感知调度

```go
func (s *DAGSchedulerImpl) calculateBatchSize(nodes []*DAGNode) int {
    // 获取可用资源
    availableSlots := s.resourceMgr.AvailableSlots()

    // 计算每个节点的资源需求
    totalRequired := 0
    for _, node := range nodes {
        totalRequired += s.estimateResourceRequirement(node)
    }

    // 批次大小受限于可用资源
    batchSize := min(len(nodes), availableSlots)

    // 如果总需求超过可用资源，按比例缩减
    if totalRequired > availableSlots {
        batchSize = (len(nodes) * availableSlots) / totalRequired
    }

    // 至少执行 1 个任务
    return max(batchSize, 1)
}

func (s *DAGSchedulerImpl) estimateResourceRequirement(node *DAGNode) int {
    // 基于任务类型和预估时长估算资源需求
    // 这里简化为固定值，实际可根据历史数据动态调整
    return 1
}
```

## 4. DAG 算法

### 4.1 拓扑排序

```go
func (b *DAGBuilder) topologicalSort() ([][]*DAGNode, error) {
    // 计算入度
    inDegree := make(map[string]int)
    for id := range b.graph.nodes {
        inDegree[id] = 0
    }
    for _, edges := range b.graph.edges {
        for _, edge := range edges {
            inDegree[edge.To]++
        }
    }

    // Kahn 算法
    queue := make([]*DAGNode, 0)
    for id, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, b.graph.nodes[id])
        }
    }

    levels := make([][]*DAGNode, 0)
    for len(queue) > 0 {
        level := make([]*DAGNode, len(queue))
        copy(level, queue)
        levels = append(levels, level)

        nextQueue := make([]*DAGNode, 0)
        for _, node := range queue {
            for _, edge := range b.graph.edges[node.ID] {
                inDegree[edge.To]--
                if inDegree[edge.To] == 0 {
                    nextQueue = append(nextQueue, b.graph.nodes[edge.To])
                }
            }
        }
        queue = nextQueue
    }

    // 检查是否所有节点都已处理
    if len(levels) != len(b.graph.nodes) {
        return nil, fmt.Errorf("cycle detected in DAG")
    }

    return levels, nil
}
```

### 4.2 循环依赖检测

```go
func (b *DAGBuilder) detectCycles() error {
    // 使用 DFS 检测循环
    visited := make(map[string]bool)
    recursionStack := make(map[string]bool)

    var hasCycle bool
    var cycleNodes []string

    var dfs func(nodeID string)
    dfs = func(nodeID string) {
        visited[nodeID] = true
        recursionStack[nodeID] = true

        for _, edge := range b.graph.edges[nodeID] {
            if !visited[edge.To] {
                dfs(edge.To)
            } else if recursionStack[edge.To] {
                hasCycle = true
                cycleNodes = append(cycleNodes, edge.To)
            }
        }

        recursionStack[nodeID] = false
    }

    for id := range b.graph.nodes {
        if !visited[id] {
            dfs(id)
        }
    }

    if hasCycle {
        return fmt.Errorf("cycle detected: %v", cycleNodes)
    }

    return nil
}
```

### 4.3 关键路径计算

```go
func (b *DAGBuilder) calculateCriticalPath() {
    // 动态规划计算最长路径
    dist := make(map[string]time.Duration)
    prev := make(map[string]string)

    // 初始化
    for id, node := range b.graph.nodes {
        dist[id] = node.EstimatedDuration
    }

    // 按拓扑层级更新
    for _, level := range b.graph.levels {
        for _, node := range level {
            for _, depID := range node.Dependencies {
                if dist[node.ID] < dist[depID]+node.EstimatedDuration {
                    dist[node.ID] = dist[depID] + node.EstimatedDuration
                    prev[node.ID] = depID
                }
            }
        }
    }

    // 找到最长路径的终点
    var maxID string
    maxDuration := time.Duration(0)
    for id, duration := range dist {
        if duration > maxDuration {
            maxDuration = duration
            maxID = id
        }
    }

    // 回溯关键路径
    var path []*DAGNode
    for id := maxID; id != ""; id = prev[id] {
        path = append([]*DAGNode{b.graph.nodes[id]}, path...)
    }

    b.graph.criticalPath = path
}
```

### 4.4 拓扑层级计算

```go
func (b *DAGBuilder) calculateLevels() error {
    levels, err := b.topologicalSort()
    if err != nil {
        return err
    }
    b.graph.levels = levels
    return nil
}
```

## 5. 失败处理策略

### 5.1 失败策略类型

```go
type FailureStrategy string

const (
    StrategyFailFast   FailureStrategy = "fail_fast"    // 快速失败：任一任务失败立即停止
    StrategyContinue   FailureStrategy = "continue"    // 继续执行：忽略失败继续执行
    StrategyBestEffort FailureStrategy = "best_effort" // 尽力而为：跳过依赖失败的任务
)
```

### 5.2 失败策略实现

```go
func (s *DAGSchedulerImpl) handleFailure(node *DAGNode, err error) {
    switch s.failureStrategy {
    case StrategyFailFast:
        // 取消整个调度
        s.cancelAllNodes()

    case StrategyContinue:
        // 仅标记当前节点失败，继续执行其他节点
        node.State = NodeStateFailed
        node.Error = &ExecutionError{
            Code:    "TASK_FAILED",
            Message: err.Error(),
        }

    case StrategyBestEffort:
        // 标记当前节点失败，并跳过依赖此节点的任务
        node.State = NodeStateFailed
        node.Error = &ExecutionError{
            Code:    "TASK_FAILED",
            Message: err.Error(),
        }
        s.skipDependentNodes(node)
    }
}

func (s *DAGSchedulerImpl) skipDependentNodes(failedNode *DAGNode) {
    // 递归跳过所有依赖此节点的任务
    queue := []*DAGNode{failedNode}

    for len(queue) > 0 {
        node := queue[0]
        queue = queue[1:]

        for _, depID := range node.Dependents {
            depNode := s.graph.nodes[depID]
            if depNode.State == NodeStatePending {
                depNode.State = NodeStateSkipped
                depNode.Error = &ExecutionError{
                    Code:    "SKIPPED_DUE_TO_DEPENDENCY_FAILURE",
                    Message: fmt.Sprintf("dependency %s failed", node.ID),
                }
                queue = append(queue, depNode)
            }
        }
    }
}
```

## 6. 调度结果

### 6.1 调度结果结构

```go
type ScheduleResult struct {
    // 调度 ID
    ScheduleID string `json:"schedule_id"`

    // 调度状态
    Status ScheduleStatus `json:"status"`

    // 开始时间
    StartTime time.Time `json:"start_time"`

    // 结束时间
    EndTime time.Time `json:"end_time"`

    // 总耗时
    Duration time.Duration `json:"duration"`

    // DAG 图
    Graph *DAGGraph `json:"graph"`

    // 层级结果
    LevelResults []*LevelResult `json:"level_results,omitempty"`

    // 错误信息
    Error error `json:"error,omitempty"`

    // 元数据
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ScheduleStatus string

const (
    ScheduleStatusPending   ScheduleStatus = "pending"
    ScheduleStatusRunning   ScheduleStatus = "running"
    ScheduleStatusCompleted ScheduleStatus = "completed"
    ScheduleStatusFailed    ScheduleStatus = "failed"
    ScheduleStatusCancelled ScheduleStatus = "cancelled"
)

type LevelResult struct {
    // 层级索引
    LevelIndex int `json:"level_index"`

    // 开始时间
    StartTime time.Time `json:"start_time"`

    // 结束时间
    EndTime time.Time `json:"end_time"`

    // 层级耗时
    Duration time.Duration `json:"duration"`

    // 批次结果
    BatchResults []*BatchResult `json:"batch_results,omitempty"`

    // 节点结果
    NodeResults []*NodeResult `json:"node_results,omitempty"`

    // 错误列表
    Errors []error `json:"errors,omitempty"`
}

type BatchResult struct {
    // 开始时间
    StartTime time.Time `json:"start_time"`

    // 结束时间
    EndTime time.Time `json:"end_time"`

    // 批次耗时
    Duration time.Duration `json:"duration"`

    // 节点结果
    NodeResults []*NodeResult `json:"node_results"`

    // 错误列表
    Errors []error `json:"errors,omitempty"`
}

type NodeResult struct {
    // 节点 ID
    NodeID string `json:"node_id"`

    // 是否成功
    Success bool `json:"success"`

    // 执行结果
    Result *TaskResult `json:"result,omitempty"`

    // 错误信息
    Error error `json:"error,omitempty"`

    // 执行耗时
    Duration time.Duration `json:"duration"`
}
```

## 7. 性能优化

### 7.1 DAG 缓存

```go
type DAGCache struct {
    cache map[string]*CachedDAG
    mu    sync.RWMutex
    ttl   time.Duration
}

type CachedDAG struct {
    Graph      *DAGGraph
    CachedAt   time.Time
    AccessTime time.Time
}

func (c *DAGCache) Get(key string) (*DAGGraph, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    cached, ok := c.cache[key]
    if !ok {
        return nil, false
    }

    // 检查 TTL
    if time.Since(cached.CachedAt) > c.ttl {
        delete(c.cache, key)
        return nil, false
    }

    cached.AccessTime = time.Now()
    return cached.Graph, true
}

func (c *DAGCache) Set(key string, graph *DAGGraph) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.cache[key] = &CachedDAG{
        Graph:      graph,
        CachedAt:   time.Now(),
        AccessTime: time.Now(),
    }
}
```

### 7.2 增量调度

```go
func (s *DAGSchedulerImpl) IncrementalSchedule(existingGraph *DAGGraph, newTasks []*AgentTask) (*ScheduleResult, error) {
    // 构建增量 DAG
    builder := NewDAGBuilder()

    // 复用已有节点
    for id, node := range existingGraph.nodes {
        if node.State == NodeStateCompleted {
            continue // 跳过已完成的节点
        }
        builder.AddNode(node.Task, node.EstimatedDuration)
    }

    // 添加新节点
    for _, task := range newTasks {
        builder.AddNode(task, s.estimateDuration(task))
    }

    // 添加新边（基于任务依赖）
    for _, task := range newTasks {
        for _, depID := range task.Dependencies {
            builder.AddEdge(depID, task.ID, EdgeTypeHard, nil)
        }
    }

    // 构建新 DAG
    newGraph, err := builder.Build()
    if err != nil {
        return nil, err
    }

    // 调度新 DAG
    return s.Schedule(newGraph)
}
```

### 7.3 动态重调度

```go
func (s *DAGSchedulerImpl) Reschedule(scheduleID string, newStrategy FailureStrategy) (*ScheduleResult, error) {
    // 获取原始调度结果
    original, err := s.stateStore.Load(scheduleID)
    if err != nil {
        return nil, err
    }

    // 重置未完成的节点
    for _, node := range original.Graph.nodes {
        if node.State == NodeStatePending || node.State == NodeStateRunning {
            node.State = NodeStatePending
            node.StartTime = nil
            node.EndTime = nil
            node.Error = nil
        }
    }

    // 更新失败策略
    s.failureStrategy = newStrategy

    // 重新调度
    return s.Schedule(original.Graph)
}
```

## 8. 可观测性

### 8.1 指标

```go
type DAGMetrics struct {
    // 调度指标
    ScheduleCount          int64
    ScheduleSuccessCount   int64
    ScheduleFailureCount   int64
    ScheduleDuration       prometheus.Histogram

    // 节点指标
    NodeCount              int64
    NodeSuccessCount       int64
    NodeFailureCount       int64
    NodeSkippedCount       int64
    NodeDuration           prometheus.Histogram

    // 层级指标
    LevelCount             int64
    LevelDuration          prometheus.Histogram

    // 资源指标
    ConcurrentExecutions   prometheus.Gauge
    ResourceUtilization    prometheus.Gauge
}
```

### 8.2 链路追踪

```go
func (s *DAGSchedulerImpl) executeNodeWithTracing(ctx context.Context, node *DAGNode) *NodeResult {
    // 创建 span
    ctx, span := tracer.Start(ctx, "execute_node",
        trace.WithAttributes(
            attribute.String("node_id", node.ID),
            attribute.String("task_type", node.Task.Type),
        ),
    )
    defer span.End()

    // 执行节点
    result := s.executeNode(ctx, node)

    // 记录结果
    if result.Error != nil {
        span.SetStatus(codes.Error, result.Error.Error())
        span.RecordError(result.Error)
    }

    return result
}
```

## 9. 与契约层的集成

### 9.1 从契约构建 DAG

```go
func ContractToDAG(contracts []*AgentContract) (*DAGGraph, error) {
    builder := NewDAGBuilder()

    // 添加节点
    for _, contract := range contracts {
        task := &AgentTask{
            ID:       contract.ContractID,
            Contract: contract,
        }
        builder.AddNode(task, contract.SLA.Timeout)
    }

    // 添加边（基于契约依赖）
    for _, contract := range contracts {
        for _, dep := range contract.Dependencies {
            builder.AddEdge(dep.DependencyAgentID, contract.ContractID, EdgeTypeHard, nil)
        }
    }

    // 构建 DAG
    return builder.Build()
}
```

### 9.2 DAG 执行结果映射到契约

```go
func DAGResultToContractExecution(result *ScheduleResult, contractID string) *ContractExecutionResult {
    node := result.Graph.nodes[contractID]

    return &ContractExecutionResult{
        ContractID:      contractID,
        ContractVersion: node.Task.Contract.Version,
        Status:          mapNodeStateToExecutionStatus(node.State),
        Output:          node.Task.Result,
        Error:           node.Error,
        Duration:        node.ActualDuration,
        Metadata: map[string]interface{}{
            "schedule_id": result.ScheduleID,
            "level_index": getNodeLevel(result.Graph, contractID),
        },
    }
}
```

## 10. 实施路线图

### 里程碑 1：DAG 基础调度
- [ ] 实现 DAG 图构建
- [ ] 实现拓扑排序
- [ ] 实现循环依赖检测
- [ ] 实现基础并行调度（层级并行）
- [ ] 实现快速失败策略
- [ ] 验证性能提升（串行 vs 并行对比）

### 里程碑 2：DAG 高级调度
- [ ] 实现关键路径计算
- [ ] 实现资源感知调度
- [ ] 实现多种失败策略（继续执行、尽力而为）
- [ ] 实现动态 DAG 调整
- [ ] 实现增量调度
- [ ] 实现 DAG 缓存

### 里程碑 3：DAG 生态成熟
- [ ] 实现 DAG 可视化（Web UI）
- [ ] 实现 DAG 性能分析工具
- [ ] 实现 DAG 模板库
- [ ] 实现智能调度（基于历史数据优化）
- [ ] 实现 DAG 预编译
- [ ] 实现分布式 DAG 调度（跨节点）
