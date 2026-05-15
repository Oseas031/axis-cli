# DAG 并行调度策略设计（里程碑1简化版）

## 设计目标

构建 DAG 有向无环图并行调度模型，主 Agent 基于任务依赖关系生成全流程 DAG 图，无依赖的任务 100% 并行执行，有依赖的任务流水线化执行，总耗时从「各环节时长之和」压缩为「最长关键路径的单环节时长」。

## 里程碑1设计原则（奥卡姆剃刀）
- **最小可行**：只实现基础DAG构建 + 拓扑排序 + 简单层级并行
- **渐进增强**：条件依赖、软依赖、资源感知调度在后续里程碑添加
- **聚焦核心**：里程碑1聚焦于"DAG依赖管理 + 基础并行"的可行性验证

## 1. 核心概念（里程碑1简化版）

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

**里程碑1暂不包含**：
- ❌ 关键路径计算
- ❌ 资源感知调度

## 2. DAG 数据结构（里程碑1简化版）

### 2.1 DAG 图

```go
type DAGGraph struct {
    // 节点集合
    nodes map[string]*DAGNode

    // 边集合
    edges map[string][]*DAGEdge

    // 拓扑层级（缓存）
    levels [][]*DAGNode

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

    // 实际执行时长
    ActualDuration time.Duration `json:"actual_duration,omitempty"`

    // 开始时间
    StartTime *time.Time `json:"start_time,omitempty"`

    // 结束时间
    EndTime *time.Time `json:"end_time,omitempty"`

    // 错误信息
    Error *ExecutionError `json:"error,omitempty"`
}

type DAGEdge struct {
    // 边 ID
    ID string `json:"id"`

    // 源节点 ID
    From string `json:"from"`

    // 目标节点 ID
    To string `json:"to"`
}

type NodeState string

const (
    NodeStatePending   NodeState = "pending"    // 等待中
    NodeStateReady     NodeState = "ready"      // 就绪（依赖已满足）
    NodeStateRunning   NodeState = "running"    // 执行中
    NodeStateCompleted NodeState = "completed"  // 已完成
    NodeStateFailed    NodeState = "failed"     // 失败
)
```

**里程碑1暂不包含**：
- ❌ 预估执行时长（EstimatedDuration）
- ❌ 边类型（EdgeType：硬依赖、软依赖、条件依赖）
- ❌ 边条件（EdgeCondition）
- ❌ 节点状态 Skipped、Cancelled

### 2.2 DAG 构建器（简化版）

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

func (b *DAGBuilder) AddNode(task *AgentTask) error {
    node := &DAGNode{
        ID:           task.ID,
        Task:         task,
        Dependencies: []string{},
        Dependents:   []string{},
        State:        NodeStatePending,
    }
    b.graph.nodes[task.ID] = node
    return nil
}

func (b *DAGBuilder) AddEdge(fromID, toID string) error {
    // 检查节点是否存在
    if _, ok := b.graph.nodes[fromID]; !ok {
        return fmt.Errorf("node %s not found", fromID)
    }
    if _, ok := b.graph.nodes[toID]; !ok {
        return fmt.Errorf("node %s not found", toID)
    }

    // 添加边
    edge := &DAGEdge{
        ID:   fmt.Sprintf("%s->%s", fromID, toID),
        From: fromID,
        To:   toID,
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

    return b.graph, nil
}
```

**里程碑1简化**：
- AddNode 移除 estimatedDuration 参数
- AddEdge 移除 edgeType 和 condition 参数
- Build 移除 calculateCriticalPath 调用

## 3. DAG 调度器（里程碑1简化版）

### 3.1 调度器接口

```go
type DAGScheduler interface {
    // 调度 DAG
    Schedule(graph *DAGGraph) (*ScheduleResult, error)

    // 取消调度
    Cancel(scheduleID string) error
}
```

### 3.2 调度器实现（简化版）

```go
type DAGSchedulerImpl struct {
    executor   TaskExecutor
    stateStore StateStore
    mu         sync.RWMutex
}

func NewDAGScheduler(executor TaskExecutor) *DAGSchedulerImpl {
    return &DAGSchedulerImpl{
        executor:   executor,
        stateStore: NewMemoryStateStore(),
    }
}
```

**里程碑1简化**：
- 移除 ResourceManager
- 移除 FailureStrategy（默认快速失败）
- 移除 MetricsCollector
- 移除 GetStatus、RegisterResourceManager、SetFailureStrategy 方法

### 3.3 核心调度算法（简化版）

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
        // 执行当前层级的所有节点
        levelResult := s.executeLevel(ctx, level)
        result.LevelResults = append(result.LevelResults, levelResult)

        // 快速失败：任一任务失败立即停止
        if levelResult.HasFailure() {
            cancel()
            result.Status = ScheduleStatusFailed
            result.Error = levelResult.FirstError()
            return result, result.Error
        }
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

**里程碑1简化**：
- 移除资源感知调度（calculateBatchSize）
- 移除分批执行（executeBatch）
- 移除多种失败策略（只保留快速失败）
- 移除指标收集（metrics）

## 4. DAG 算法（里程碑1简化版）

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

### 4.3 拓扑层级计算

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

**里程碑1暂不包含**：
- ❌ 关键路径计算（calculateCriticalPath）

## 5. 调度结果（里程碑1简化版）

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
}

type ScheduleStatus string

const (
    ScheduleStatusCompleted ScheduleStatus = "completed"
    ScheduleStatusFailed    ScheduleStatus = "failed"
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

    // 节点结果
    NodeResults []*NodeResult `json:"node_results,omitempty"`

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

**里程碑1暂不包含**：
- ❌ BatchResult（批次结果）
- ❌ Metadata（元数据）
- ❌ ScheduleStatus Pending、Running、Cancelled

## 6. 实施路线图（里程碑1简化版）

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
