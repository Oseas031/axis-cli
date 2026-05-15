# Call Human 标准化调用协议规范 V1.0

## 协议概述

定义 Agent 调用 Human-as-a-Function 的标准化协议，包括入参结构、出参格式、超时管控、错误码体系、重试策略与幂等性保证。

## 1. 调用请求结构

### 1.1 请求头

```go
type HumanCallRequest struct {
    // 调用 ID（全局唯一）
    CallID string `json:"call_id"`

    // 调用类型（枚举）
    // code_review: 代码评审
    // prod_approval: 生产环境高权限操作审批
    // offline_config: 线下环境配置执行
    // requirement_confirm: 需求边界与变更确认
    // compliance_check: 合规性校验
    // physical_execution: 物理世界信息采集与执行
    CallType CallType `json:"call_type"`

    // 调用者信息
    Caller *CallerInfo `json:"caller"`

    // 调用参数
    Parameters map[string]interface{} `json:"parameters"`

    // 超时设置（秒）
    Timeout int `json:"timeout"`

    // 优先级（0-10，10 最高）
    Priority int `json:"priority"`

    // 幂等性键（用于重试去重）
    IdempotencyKey string `json:"idempotency_key,omitempty"`

    // 上下文信息
    Context *CallContext `json:"context,omitempty"`

    // 元数据
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

### 1.2 调用者信息

```go
type CallerInfo struct {
    // Agent ID
    AgentID string `json:"agent_id"`

    // Agent 类型
    AgentType string `json:"agent_type"`

    // 调用时间戳
    Timestamp int64 `json:"timestamp"`

    // 调用链 ID（用于链路追踪）
    TraceID string `json:"trace_id,omitempty"`
}
```

### 1.3 调用上下文

```go
type CallContext struct {
    // 关联任务 ID
    TaskID string `json:"task_id,omitempty"`

    // 前置调用结果
    PreviousResults []*HumanCallResult `json:"previous_results,omitempty"`

    // 共享上下文数据
    SharedData map[string]interface{} `json:"shared_data,omitempty"`
}
```

## 2. 调用响应结构

### 2.1 响应头

```go
type HumanCallResult struct {
    // 调用 ID
    CallID string `json:"call_id"`

    // 调用状态
    Status CallStatus `json:"status"`

    // 调用结果
    Result *StructuredResult `json:"result,omitempty"`

    // 错误信息
    Error *CallError `json:"error,omitempty"`

    // 执行元数据
    Execution *ExecutionMetadata `json:"execution,omitempty"`

    // 时间戳
    Timestamp int64 `json:"timestamp"`
}
```

### 2.2 调用状态枚举

```go
type CallStatus string

const (
    StatusPending    CallStatus = "pending"     // 待执行
    StatusQueued     CallStatus = "queued"      // 已排队
    StatusExecuting  CallStatus = "executing"   // 执行中
    StatusCompleted  CallStatus = "completed"   // 已完成
    StatusFailed     CallStatus = "failed"      // 失败
    StatusCancelled  CallStatus = "cancelled"   // 已取消
    StatusTimeout    CallStatus = "timeout"     // 超时
)
```

### 2.3 结构化结果

```go
type StructuredResult struct {
    // 原始结果（自然语言）
    RawResult string `json:"raw_result"`

    // 解析后的结构化数据
    ParsedData map[string]interface{} `json:"parsed_data,omitempty"`

    // 置信度（0-1）
    Confidence float64 `json:"confidence,omitempty"`

    // 解析器信息
    Parser *ParserInfo `json:"parser,omitempty"`
}
```

### 2.4 执行元数据

```go
type ExecutionMetadata struct {
    // 队列等待时间（毫秒）
    QueueDuration int64 `json:"queue_duration"`

    // 执行时间（毫秒）
    ExecutionDuration int64 `json:"execution_duration"`

    // 重试次数
    RetryCount int `json:"retry_count"`

    // 执行者信息
    Executor *ExecutorInfo `json:"executor,omitempty"`
}
```

## 3. 错误码体系

### 3.1 错误码结构

```go
type CallError struct {
    // 错误码
    Code ErrorCode `json:"code"`

    // 错误消息
    Message string `json:"message"`

    // 错误详情
    Details map[string]interface{} `json:"details,omitempty"`

    // 重试建议
    RetryAdvice *RetryAdvice `json:"retry_advice,omitempty"`
}
```

### 3.2 错误码定义

| 错误码 | 类别 | 描述 | 可重试 |
|--------|------|------|--------|
| `ERR_0001` | 协议错误 | 请求参数格式错误 | 否 |
| `ERR_0002` | 协议错误 | 缺少必需参数 | 否 |
| `ERR_0003` | 协议错误 | 参数类型不匹配 | 否 |
| `ERR_0101` | 权限错误 | 调用者无权限调用 | 否 |
| `ERR_0102` | 权限错误 | 超出调用配额 | 否 |
| `ERR_0201` | 执行错误 | 人类节点不可用 | 是 |
| `ERR_0202` | 执行错误 | 人类节点拒绝执行 | 否 |
| `ERR_0203` | 执行错误 | 执行超时 | 是 |
| `ERR_0204` | 执行错误 | 执行中断 | 否 |
| `ERR_0301` | 解析错误 | 结果解析失败 | 否 |
| `ERR_0302` | 解析错误 | 结果验证失败 | 否 |
| `ERR_0401` | 系统错误 | 内部服务错误 | 是 |
| `ERR_0402` | 系统错误 | 资源不足 | 是 |
| `ERR_0403` | 系统错误 | 状态不一致 | 否 |

### 3.3 重试建议

```go
type RetryAdvice struct {
    // 是否可重试
    Retryable bool `json:"retryable"`

    // 重试策略
    Strategy RetryStrategy `json:"strategy,omitempty"`

    // 建议重试延迟（毫秒）
    DelayMs int64 `json:"delay_ms,omitempty"`

    // 最大重试次数
    MaxRetries int `json:"max_retries,omitempty"`
}
```

## 4. 超时管控机制

### 4.1 超时层级

```go
type TimeoutConfig struct {
    // 协议层默认超时
    ProtocolDefault int `json:"protocol_default"`

    // 协议层最大超时
    ProtocolMax int `json:"protocol_max"`

    // 执行层默认超时
    ExecutorDefault int `json:"executor_default"`

    // 执行层最大超时
    ExecutorMax int `json:"executor_max"`

    // 解析层超时
    ParserTimeout int `json:"parser_timeout"`
}
```

### 4.2 超时处理流程

1. **请求超时**：调用请求超过 `ProtocolMax` 未被处理 → 返回 `ERR_0203`
2. **执行超时**：执行超过 `ExecutorMax` → 自动取消，返回 `ERR_0203`
3. **解析超时**：解析超过 `ParserTimeout` → 返回 `ERR_0301`
4. **级联超时**：上层超时自动触发下层超时

## 5. 重试策略

### 5.1 重试策略类型

```go
type RetryStrategy string

const (
    RetryNone       RetryStrategy = "none"        // 不重试
    RetryLinear     RetryStrategy = "linear"      // 线性退避
    RetryExponential RetryStrategy = "exponential" // 指数退避
    RetryImmediate  RetryStrategy = "immediate"    // 立即重试
)
```

### 5.2 重试配置

```go
type RetryConfig struct {
    // 重试策略
    Strategy RetryStrategy `json:"strategy"`

    // 最大重试次数
    MaxRetries int `json:"max_retries"`

    // 初始延迟（毫秒）
    InitialDelayMs int64 `json:"initial_delay_ms"`

    // 最大延迟（毫秒）
    MaxDelayMs int64 `json:"max_delay_ms"`

    // 退避因子（指数退避）
    BackoffFactor float64 `json:"backoff_factor"`

    // 抖动启用（避免惊群效应）
    JitterEnabled bool `json:"jitter_enabled"`
}
```

### 5.3 重试决策

- 仅错误码标记为 `Retryable: true` 时才重试
- 超过 `MaxRetries` 后停止重试
- 幂等性键相同的调用不会重复执行
- 重试次数计入 `ExecutionMetadata.RetryCount`

## 6. 幂等性保证

### 6.1 幂等性键

```go
type IdempotencyKey struct {
    // 键值（由调用者生成，建议 UUID）
    Key string `json:"key"`

    // 有效期（秒）
    TTL int `json:"ttl"`
}
```

### 6.2 幂等性保证机制

1. **调用去重**：相同 `IdempotencyKey` 的调用直接返回缓存结果
2. **结果缓存**：成功调用的结果缓存 TTL 时间
3. **状态去重**：执行中的调用重复请求返回当前状态
4. **缓存失效**：失败调用不缓存，允许重试

### 6.3 幂等性约束

- 幂等性键必须全局唯一
- 幂等性键有效期结束后自动清理
- 幂等性键仅对成功调用有效
- 调用者负责生成幂等性键

## 7. 6 大核心技术落地场景

### 7.1 代码评审 (code_review)

**参数结构**：
```json
{
  "code": "待评审代码",
  "language": "编程语言",
  "review_focus": ["安全性", "性能", "可维护性"]
}
```

**结果结构**：
```json
{
  "issues": ["问题列表"],
  "suggestions": ["改进建议"],
  "overall_score": 0-100
}
```

### 7.2 生产环境高权限操作审批 (prod_approval)

**参数结构**：
```json
{
  "operation": "操作描述",
  "environment": "生产环境",
  "risk_level": "风险等级",
  "rollback_plan": "回滚方案"
}
```

**结果结构**：
```json
{
  "approved": true/false,
  "conditions": ["审批条件"],
  "approver": "审批人"
}
```

### 7.3 线下环境配置执行 (offline_config)

**参数结构**：
```json
{
  "config_target": "配置目标",
  "config_changes": "配置变更",
  "verification_steps": ["验证步骤"]
}
```

**结果结构**：
```json
{
  "executed": true/false,
  "verification_results": ["验证结果"],
  "notes": "执行备注"
}
```

### 7.4 需求边界与变更确认 (requirement_confirm)

**参数结构**：
```json
{
  "requirement": "需求描述",
  "proposed_changes": ["提议变更"],
  "impact_analysis": "影响分析"
}
```

**结果结构**：
```json
{
  "confirmed": true/false,
  "clarifications": ["澄清点"],
  "accepted_changes": ["接受的变更"]
}
```

### 7.5 合规性校验 (compliance_check)

**参数结构**：
```json
{
  "subject": "校验对象",
  "compliance_type": "合规类型",
  "regulatory_framework": "监管框架"
}
```

**结果结构**：
```json
{
  "compliant": true/false,
  "violations": ["违规项"],
  "remediation_steps": ["整改步骤"]
}
```

### 7.6 物理世界信息采集与执行 (physical_execution)

**参数结构**：
```json
{
  "action": "物理动作",
  "location": "位置",
  "safety_requirements": ["安全要求"]
}
```

**结果结构**：
```json
{
  "executed": true/false,
  "observations": ["观察结果"],
  "photos": ["照片URL"],
  "safety_verified": true/false
}
```

## 8. 协议版本管理

### 8.1 版本号格式

```
主版本号.次版本号.修订号
例：1.0.0
```

### 8.2 版本兼容性

- 主版本号变更：不兼容的 API 修改
- 次版本号变更：向后兼容的功能性新增
- 修订号变更：向后兼容的问题修正

### 8.3 版本协商

调用请求需包含协议版本：
```json
{
  "protocol_version": "1.0.0"
}
```

服务端返回支持的协议版本：
```json
{
  "supported_versions": ["1.0.0", "1.1.0"],
  "preferred_version": "1.1.0"
}
```
