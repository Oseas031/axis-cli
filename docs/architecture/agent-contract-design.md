# Agent 契约模型详细设计

## 设计目标

主 Agent 的核心职责从「发指令」变为「定契约」，所有子 Agent 封装为标准化无状态函数，与 call human 协议完全同构，彻底消除调度歧义与理解偏差。

## 1. 契约核心要素

### 1.1 输入 Schema (Input Schema)

**设计原则**：
- 强类型、无歧义的入参结构
- 仅保留完成该任务的最小必要信息
- 禁止全量上下文透传
- 支持嵌套结构与枚举

**数据结构**：
```go
type InputSchema struct {
    // Schema ID（全局唯一）
    SchemaID string `json:"schema_id"`

    // Schema 版本
    Version string `json:"version"`

    // 字段定义
    Fields []FieldDef `json:"fields"`

    // 验证规则
    ValidationRules []ValidationRule `json:"validation_rules"`

    // 必需字段列表
    RequiredFields []string `json:"required_fields"`

    // 默认值
    Defaults map[string]interface{} `json:"defaults,omitempty"`

    // 元数据
    Metadata map[string]string `json:"metadata,omitempty"`
}

type FieldDef struct {
    // 字段名
    Name string `json:"name"`

    // 字段类型
    Type FieldType `json:"type"`

    // 字段描述
    Description string `json:"description"`

    // 是否必需
    Required bool `json:"required"`

    // 默认值
    Default interface{} `json:"default,omitempty"`

    // 验证规则
    Validators []Validator `json:"validators,omitempty"`

    // 枚举值（如果适用）
    Enum []interface{} `json:"enum,omitempty"`

    // 嵌套 Schema（如果类型是 object）
    NestedSchema *InputSchema `json:"nested_schema,omitempty"`
}

type FieldType string

const (
    TypeString   FieldType = "string"
    TypeInteger  FieldType = "integer"
    TypeFloat    FieldType = "float"
    TypeBoolean  FieldType = "boolean"
    TypeArray    FieldType = "array"
    TypeObject   FieldType = "object"
    TypeDateTime FieldType = "datetime"
    TypeUUID     FieldType = "uuid"
)
```

**示例**：
```json
{
  "schema_id": "code_review_input_v1",
  "version": "1.0.0",
  "fields": [
    {
      "name": "code",
      "type": "string",
      "description": "待评审的代码片段",
      "required": true,
      "validators": [
        {"type": "min_length", "value": 10},
        {"type": "max_length", "value": 10000}
      ]
    },
    {
      "name": "language",
      "type": "string",
      "description": "编程语言",
      "required": true,
      "enum": ["go", "python", "javascript", "rust", "java"]
    },
    {
      "name": "review_focus",
      "type": "array",
      "description": "评审关注点",
      "required": false,
      "default": ["security", "performance"],
      "nested_schema": {
        "fields": [
          {
            "name": "focus",
            "type": "string",
            "enum": ["security", "performance", "maintainability", "testing"]
          }
        ]
      }
    }
  ],
  "required_fields": ["code", "language"]
}
```

### 1.2 输出 Schema (Output Schema)

**设计原则**：
- 标准化的出参结构
- 明确验收标准
- 明确格式要求
- 明确合规规则

**数据结构**：
```go
type OutputSchema struct {
    // Schema ID
    SchemaID string `json:"schema_id"`

    // Schema 版本
    Version string `json:"version"`

    // 字段定义
    Fields []FieldDef `json:"fields"`

    // 验收标准
    AcceptanceCriteria []AcceptanceCriterion `json:"acceptance_criteria"`

    // 格式要求
    FormatRequirements []FormatRequirement `json:"format_requirements"`

    // 合规规则
    ComplianceRules []ComplianceRule `json:"compliance_rules"`

    // 置信度阈值（0-1）
    ConfidenceThreshold float64 `json:"confidence_threshold,omitempty"`
}

type AcceptanceCriterion struct {
    // 标准名称
    Name string `json:"name"`

    // 标准描述
    Description string `json:"description"`

    // 验证逻辑（表达式）
    ValidationExpression string `json:"validation_expression"`

    // 是否必需通过
    Required bool `json:"required"`

    // 失败时的错误码
    FailureErrorCode string `json:"failure_error_code,omitempty"`
}

type FormatRequirement struct {
    // 字段名
    FieldName string `json:"field_name"`

    // 格式类型
    FormatType FormatType `json:"format_type"`

    // 格式参数
    FormatParams map[string]interface{} `json:"format_params,omitempty"`
}

type ComplianceRule struct {
    // 规则名称
    Name string `json:"name"`

    // 规则类型
    RuleType ComplianceRuleType `json:"rule_type"`

    // 规则参数
    RuleParams map[string]interface{} `json:"rule_params"`

    // 违规时的错误码
    ViolationErrorCode string `json:"violation_error_code"`
}
```

**示例**：
```json
{
  "schema_id": "code_review_output_v1",
  "version": "1.0.0",
  "fields": [
    {
      "name": "issues",
      "type": "array",
      "description": "发现的问题列表",
      "required": true,
      "nested_schema": {
        "fields": [
          {
            "name": "severity",
            "type": "string",
            "enum": ["critical", "high", "medium", "low"]
          },
          {
            "name": "message",
            "type": "string",
            "required": true
          },
          {
            "name": "location",
            "type": "object",
            "nested_schema": {
              "fields": [
                {"name": "file", "type": "string"},
                {"name": "line", "type": "integer"}
              ]
            }
          }
        ]
      }
    },
    {
      "name": "suggestions",
      "type": "array",
      "description": "改进建议",
      "required": false
    },
    {
      "name": "overall_score",
      "type": "integer",
      "description": "总体评分（0-100）",
      "required": true,
      "validators": [
        {"type": "range", "min": 0, "max": 100}
      ]
    }
  ],
  "acceptance_criteria": [
    {
      "name": "issues_complete",
      "description": "必须包含所有发现的问题",
      "validation_expression": "len(issues) > 0 or overall_score == 100",
      "required": true,
      "failure_error_code": "AGENT_ERR_0001"
    }
  ],
  "format_requirements": [
    {
      "field_name": "overall_score",
      "format_type": "integer_range",
      "format_params": {"min": 0, "max": 100}
    }
  ],
  "compliance_rules": [
    {
      "name": "no_pii",
      "rule_type": "data_privacy",
      "rule_params": {"check_fields": ["issues", "suggestions"]},
      "violation_error_code": "AGENT_ERR_0301"
    }
  ],
  "confidence_threshold": 0.8
}
```

### 1.3 SLA 约定 (SLA Agreement)

**数据结构**：
```go
type SLAAgreement struct {
    // 最长执行超时时间（秒）
    Timeout int `json:"timeout"`

    // 重试配置
    RetryConfig *RetryConfig `json:"retry_config"`

    // 熔断配置
    CircuitBreakerConfig *CircuitBreakerConfig `json:"circuit_breaker_config"`

    // 性能期望
    PerformanceExpectations *PerformanceExpectations `json:"performance_expectations,omitempty"`

    // 资源限制
    ResourceLimits *ResourceLimits `json:"resource_limits,omitempty"`
}

type CircuitBreakerConfig struct {
    // 失败阈值（百分比）
    FailureThreshold float64 `json:"failure_threshold"`

    // 滚动窗口大小（请求数）
    RollingWindowSize int `json:"rolling_window_size"`

    // 熔断开启后的冷却时间（秒）
    CooldownPeriod int `json:"cooldown_period"`

    // 半开状态的测试请求数
    HalfOpenRequests int `json:"half_open_requests"`
}

type PerformanceExpectations struct {
    // P95 延迟（毫秒）
    P95LatencyMs int64 `json:"p95_latency_ms"`

    // P99 延迟（毫秒）
    P99LatencyMs int64 `json:"p99_latency_ms"`

    // 吞吐量（请求/秒）
    ThroughputRPS int `json:"throughput_rps"`
}

type ResourceLimits struct {
    // 最大内存使用（MB）
    MaxMemoryMB int `json:"max_memory_mb"`

    // 最大 CPU 使用（百分比）
    MaxCPUPercent float64 `json:"max_cpu_percent"`

    // 最大并发数
    MaxConcurrency int `json:"max_concurrency"`
}
```

### 1.4 准入规则 (Admission Rules)

**设计原则**：
- 分为本地校验和远程校验两级
- 避免循环依赖
- 支持规则组合

**数据结构**：
```go
type AdmissionRules struct {
    // 本地校验规则（无需外部调用）
    LocalRules []LocalRule `json:"local_rules"`

    // 远程校验规则（需要调用其他 Agent）
    RemoteRules []RemoteRule `json:"remote_rules,omitempty"`

    // 规则组合逻辑（AND/OR）
    CombinationLogic RuleCombinationLogic `json:"combination_logic"`

    // 准入失败时的错误码
    FailureErrorCode string `json:"failure_error_code"`
}

type LocalRule struct {
    // 规则名称
    Name string `json:"name"`

    // 规则类型
    RuleType LocalRuleType `json:"rule_type"`

    // 规则参数
    RuleParams map[string]interface{} `json:"rule_params"`

    // 是否必需通过
    Required bool `json:"required"`

    // 失败时的错误码
    FailureErrorCode string `json:"failure_error_code,omitempty"`
}

type LocalRuleType string

const (
    RuleTypeSchemaValidation   LocalRuleType = "schema_validation"
    RuleTypeArchitectureCheck LocalRuleType = "architecture_check"
    RuleTypePermissionCheck   LocalRuleType = "permission_check"
    RuleTypeQuotaCheck        LocalRuleType = "quota_check"
    RuleTypeDependencyCheck   LocalRuleType = "dependency_check"
)

type RemoteRule struct {
    // 规则名称
    Name string `json:"name"`

    // 依赖的 Agent ID
    DependencyAgentID string `json:"dependency_agent_id"`

    // 调用参数（从输入中提取）
    CallParameters map[string]string `json:"call_parameters"`

    // 期望的响应值
    ExpectedResponse map[string]interface{} `json:"expected_response"`

    // 是否必需通过
    Required bool `json:"required"`

    // 失败时的错误码
    FailureErrorCode string `json:"failure_error_code,omitempty"`

    // 超时时间（秒）
    Timeout int `json:"timeout"`
}

type RuleCombinationLogic string

const (
    LogicAND RuleCombinationLogic = "and"
    LogicOR  RuleCombinationLogic = "or"
)
```

**示例**：
```json
{
  "local_rules": [
    {
      "name": "schema_validation",
      "rule_type": "schema_validation",
      "required": true,
      "failure_error_code": "AGENT_ERR_0101"
    },
    {
      "name": "architecture_check",
      "rule_type": "architecture_check",
      "rule_params": {
        "check_style_guide": true,
        "check_naming_convention": true
      },
      "required": true,
      "failure_error_code": "AGENT_ERR_0102"
    },
    {
      "name": "permission_check",
      "rule_type": "permission_check",
      "rule_params": {
        "required_permission": "code:review"
      },
      "required": true,
      "failure_error_code": "AGENT_ERR_0103"
    }
  ],
  "remote_rules": [
    {
      "name": "requirement_alignment",
      "dependency_agent_id": "requirement_validator",
      "call_parameters": {
        "requirement_id": "$input.requirement_id"
      },
      "expected_response": {
        "aligned": true
      },
      "required": false,
      "failure_error_code": "AGENT_ERR_0104",
      "timeout": 30
    }
  ],
  "combination_logic": "and",
  "failure_error_code": "AGENT_ERR_0100"
}
```

### 1.5 异常码体系 (Exception Code System)

**数据结构**：
```go
type ExceptionCodeSystem struct {
    // 错误码定义
    ErrorCodes map[string]ErrorCodeDef `json:"error_codes"`

    // 默认错误码
    DefaultErrorCode string `json:"default_error_code"`

    // 错误码分类
    ErrorCategories map[string][]string `json:"error_categories"`
}

type ErrorCodeDef struct {
    // 错误码
    Code string `json:"code"`

    // 错误类别
    Category ErrorCategory `json:"category"`

    // 错误描述
    Description string `json:"description"`

    // 严重级别
    Severity ErrorSeverity `json:"severity"`

    // 可重试
    Retryable bool `json:"retryable"`

    // 重试策略
    RetryStrategy *RetryStrategy `json:"retry_strategy,omitempty"`

    // 自动处置规则
    AutoRemediation *AutoRemediation `json:"auto_remediation,omitempty"`

    // 人工介入条件
    HumanInterventionCondition *HumanInterventionCondition `json:"human_intervention_condition,omitempty"`
}

type ErrorCategory string

const (
    CategoryInput      ErrorCategory = "input"       // 输入错误
    CategoryValidation ErrorCategory = "validation"  // 验证错误
    CategoryExecution  ErrorCategory = "execution"   // 执行错误
    CategoryResource   ErrorCategory = "resource"    // 资源错误
    CategoryExternal   ErrorCategory = "external"    // 外部依赖错误
    CategoryCompliance ErrorCategory = "compliance"  // 合规错误
)

type ErrorSeverity string

const (
    SeverityCritical ErrorSeverity = "critical"  // 严重，立即停止
    SeverityHigh     ErrorSeverity = "high"      // 高，需要关注
    SeverityMedium   ErrorSeverity = "medium"    // 中等，记录日志
    SeverityLow      ErrorSeverity = "low"       // 低，忽略
)

type AutoRemediation struct {
    // 处置类型
    RemediationType RemediationType `json:"remediation_type"`

    // 处置参数
    RemediationParams map[string]interface{} `json:"remediation_params"`

    // 最大重试次数
    MaxRetries int `json:"max_retries"`
}

type RemediationType string

const (
    RemediationRetry      RemediationType = "retry"        // 重试
    RemediationFallback   RemediationType = "fallback"     // 降级
    RemediationSkip       RemediationType = "skip"         // 跳过
    RemediationTerminate  RemediationType = "terminate"    // 终止
)

type HumanInterventionCondition struct {
    // 触发条件
    TriggerCondition string `json:"trigger_condition"`

    // 调用 human 的参数
    HumanCallParams map[string]interface{} `json:"human_call_params"`
}
```

**错误码定义示例**：
```json
{
  "error_codes": {
    "AGENT_ERR_0001": {
      "code": "AGENT_ERR_0001",
      "category": "validation",
      "description": "输出验收标准未通过",
      "severity": "high",
      "retryable": false,
      "auto_remediation": {
        "remediation_type": "terminate",
        "remediation_params": {}
      }
    },
    "AGENT_ERR_0101": {
      "code": "AGENT_ERR_0101",
      "category": "validation",
      "description": "输入 Schema 验证失败",
      "severity": "high",
      "retryable": false,
      "auto_remediation": {
        "remediation_type": "terminate",
        "remediation_params": {}
      }
    },
    "AGENT_ERR_0201": {
      "code": "AGENT_ERR_0201",
      "category": "execution",
      "description": "Agent 执行超时",
      "severity": "high",
      "retryable": true,
      "retry_strategy": {
        "strategy": "exponential",
        "max_retries": 3,
        "initial_delay_ms": 1000,
        "max_delay_ms": 10000
      }
    },
    "AGENT_ERR_0301": {
      "code": "AGENT_ERR_0301",
      "category": "compliance",
      "description": "输出包含敏感信息",
      "severity": "critical",
      "retryable": false,
      "auto_remediation": {
        "remediation_type": "terminate",
        "remediation_params": {}
      },
      "human_intervention_condition": {
        "trigger_condition": "always",
        "human_call_params": {
          "call_type": "compliance_check",
          "severity": "critical"
        }
      }
    }
  },
  "default_error_code": "AGENT_ERR_9999",
  "error_categories": {
    "input": ["AGENT_ERR_01xx"],
    "validation": ["AGENT_ERR_01xx", "AGENT_ERR_03xx"],
    "execution": ["AGENT_ERR_02xx"],
    "resource": ["AGENT_ERR_04xx"],
    "external": ["AGENT_ERR_05xx"],
    "compliance": ["AGENT_ERR_03xx"]
  }
}
```

## 2. 契约定义

### 2.1 契约元数据

```go
type AgentContract struct {
    // 契约 ID（全局唯一）
    ContractID string `json:"contract_id"`

    // 契约版本
    Version string `json:"version"`

    // Agent 信息
    AgentInfo *AgentInfo `json:"agent_info"`

    // 输入 Schema
    InputSchema *InputSchema `json:"input_schema"`

    // 输出 Schema
    OutputSchema *OutputSchema `json:"output_schema"`

    // SLA 约定
    SLA *SLAAgreement `json:"sla"`

    // 准入规则
    AdmissionRules *AdmissionRules `json:"admission_rules"`

    // 异常码体系
    ExceptionCodes *ExceptionCodeSystem `json:"exception_codes"`

    // 依赖声明
    Dependencies []DependencyDeclaration `json:"dependencies,omitempty"`

    // 元数据
    Metadata map[string]string `json:"metadata,omitempty"`

    // 契约状态
    Status ContractStatus `json:"status"`

    // 创建时间
    CreatedAt int64 `json:"created_at"`

    // 更新时间
    UpdatedAt int64 `json:"updated_at"`
}

type AgentInfo struct {
    // Agent ID
    AgentID string `json:"agent_id"`

    // Agent 名称
    Name string `json:"name"`

    // Agent 描述
    Description string `json:"description"`

    // Agent 类型
    Type AgentType `json:"type"`

    // Agent 实现类型
    ImplementationType ImplementationType `json:"implementation_type"`
}

type AgentType string

const (
    AgentTypeTask      AgentType = "task"       // 任务型 Agent
    AgentTypeWorkflow  AgentType = "workflow"   // 工作流型 Agent
    AgentTypeService   AgentType = "service"    // 服务型 Agent
)

type ImplementationType string

const (
    ImplTypeLLM        ImplementationType = "llm"        // LLM 实现
    ImplTypeCode       ImplementationType = "code"       // 代码实现
    ImplTypeHybrid     ImplementationType = "hybrid"    // 混合实现
    ImplTypeHuman      ImplementationType = "human"     // Human-as-a-Function
)

type DependencyDeclaration struct {
    // 依赖的 Agent ID
    DependencyAgentID string `json:"dependency_agent_id"`

    // 依赖类型
    DependencyType DependencyType `json:"dependency_type"`

    // 依赖版本约束
    VersionConstraint string `json:"version_constraint,omitempty"`

    // 是否必需
    Required bool `json:"required"`
}

type DependencyType string

const (
    DepTypeAdmission  DependencyType = "admission"  // 准入依赖
    DepTypeExecution  DependencyType = "execution"  // 执行依赖
    DepTypeValidation DependencyType = "validation" // 验证依赖
)

type ContractStatus string

const (
    StatusDraft      ContractStatus = "draft"       // 草稿
    StatusActive     ContractStatus = "active"      // 激活
    StatusDeprecated ContractStatus = "deprecated"  // 废弃
    StatusRetired    ContractStatus = "retired"     // 退役
)
```

### 2.2 契约示例

```json
{
  "contract_id": "code_review_agent_v1",
  "version": "1.0.0",
  "agent_info": {
    "agent_id": "agent_code_review",
    "name": "Code Review Agent",
    "description": "执行代码评审任务",
    "type": "task",
    "implementation_type": "llm"
  },
  "input_schema": {
    "schema_id": "code_review_input_v1",
    "version": "1.0.0",
    "fields": [
      {
        "name": "code",
        "type": "string",
        "description": "待评审的代码片段",
        "required": true
      },
      {
        "name": "language",
        "type": "string",
        "description": "编程语言",
        "required": true,
        "enum": ["go", "python", "javascript", "rust", "java"]
      }
    ]
  },
  "output_schema": {
    "schema_id": "code_review_output_v1",
    "version": "1.0.0",
    "fields": [
      {
        "name": "issues",
        "type": "array",
        "required": true
      },
      {
        "name": "overall_score",
        "type": "integer",
        "required": true
      }
    ],
    "acceptance_criteria": [
      {
        "name": "issues_complete",
        "validation_expression": "len(issues) > 0 or overall_score == 100",
        "required": true
      }
    ]
  },
  "sla": {
    "timeout": 300,
    "retry_config": {
      "strategy": "exponential",
      "max_retries": 3
    },
    "circuit_breaker_config": {
      "failure_threshold": 0.5,
      "rolling_window_size": 10,
      "cooldown_period": 60
    }
  },
  "admission_rules": {
    "local_rules": [
      {
        "name": "schema_validation",
        "rule_type": "schema_validation",
        "required": true
      },
      {
        "name": "permission_check",
        "rule_type": "permission_check",
        "rule_params": {
          "required_permission": "code:review"
        },
        "required": true
      }
    ],
    "combination_logic": "and"
  },
  "exception_codes": {
    "error_codes": {
      "AGENT_ERR_0001": {
        "code": "AGENT_ERR_0001",
        "category": "validation",
        "description": "输出验收标准未通过",
        "severity": "high",
        "retryable": false
      }
    },
    "default_error_code": "AGENT_ERR_9999"
  },
  "dependencies": [],
  "status": "active",
  "created_at": 1715145600,
  "updated_at": 1715145600
}
```

## 3. 契约注册与管理

### 3.1 契约注册表

```go
type ContractRegistry struct {
    contracts map[string]*AgentContract
    versions   map[string][]string  // contract_id -> versions
    mu         sync.RWMutex
}

func (r *ContractRegistry) Register(contract *AgentContract) error
func (r *ContractRegistry) Unregister(contractID string) error
func (r *ContractRegistry) Get(contractID, version string) (*AgentContract, error)
func (r *ContractRegistry) GetLatest(contractID string) (*AgentContract, error)
func (r *ContractRegistry) List(filter ContractFilter) []*AgentContract
func (r *ContractRegistry) Validate(contract *AgentContract) error
```

### 3.2 契约验证

```go
type ContractValidator struct {
    schemaValidator SchemaValidator
    dependencyResolver DependencyResolver
}

func (v *ContractValidator) Validate(contract *AgentContract) (*ValidationResult, error)

type ValidationResult struct {
    Valid bool `json:"valid"`

    Errors []ValidationError `json:"errors"`

    Warnings []ValidationWarning `json:"warnings"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code"`
}
```

## 4. 契约执行

### 4.1 契约执行器

```go
type ContractExecutor struct {
    registry       *ContractRegistry
    admissionEngine *AdmissionEngine
    dispatcher     *Dispatcher
    circuitBreaker *CircuitBreaker
}

func (e *ContractExecutor) Execute(contractID, version string, input map[string]interface{}) (*ExecutionResult, error)

type ExecutionResult struct {
    // 契约 ID
    ContractID string `json:"contract_id"`

    // 契约版本
    ContractVersion string `json:"contract_version"`

    // 执行状态
    Status ExecutionStatus `json:"status"`

    // 输出数据
    Output map[string]interface{} `json:"output,omitempty"`

    // 验证结果
    ValidationResult *ValidationResult `json:"validation_result,omitempty"`

    // 错误信息
    Error *ExecutionError `json:"error,omitempty"`

    // 执行元数据
    Metadata *ExecutionMetadata `json:"metadata,omitempty"`
}

type ExecutionStatus string

const (
    ExecStatusPending    ExecutionStatus = "pending"
    ExecStatusAdmitting  ExecutionStatus = "admitting"
    ExecStatusExecuting  ExecutionStatus = "executing"
    ExecStatusValidating ExecutionStatus = "validating"
    ExecStatusCompleted  ExecutionStatus = "completed"
    ExecStatusFailed     ExecutionStatus = "failed"
)
```

### 4.2 准入引擎

```go
type AdmissionEngine struct {
    localRuleExecutor  LocalRuleExecutor
    remoteRuleExecutor RemoteRuleExecutor
    dependencyResolver DependencyResolver
}

func (e *AdmissionEngine) Evaluate(rules *AdmissionRules, input map[string]interface{}, context *ExecutionContext) (*AdmissionResult, error)

type AdmissionResult struct {
    // 是否通过
    Admitted bool `json:"admitted"`

    // 失败的规则
    FailedRules []FailedRule `json:"failed_rules,omitempty"`

    // 警告信息
    Warnings []string `json:"warnings,omitempty"`
}
```

## 5. 契约依赖解析

### 5.1 依赖图构建

```go
type DependencyGraph struct {
    nodes map[string]*DependencyNode
    edges map[string][]*DependencyEdge
    mu    sync.RWMutex
}

type DependencyNode struct {
    ContractID string
    Version    string
    Status     NodeStatus
}

type DependencyEdge struct {
    From        string
    To          string
    DepType     DependencyType
    Required    bool
}

func (g *DependencyGraph) Build(contracts []*AgentContract) error
func (g *DependencyGraph) DetectCycles() ([]Cycle, error)
func (g *DependencyGraph) TopologicalSort() ([]string, error)
```

### 5.2 循环依赖检测

```go
type CycleDetector struct {
    graph *DependencyGraph
}

func (d *CycleDetector) Detect() ([]Cycle, error)

type Cycle struct {
    // 循环中的契约 ID 列表
    ContractIDs []string `json:"contract_ids"`

    // 循环类型
    CycleType CycleType `json:"cycle_type"`

    // 严重级别
    Severity ErrorSeverity `json:"severity"`
}

type CycleType string

const (
    CycleTypeAdmission CycleType = "admission"  // 准入循环
    CycleTypeExecution CycleType = "execution"  // 执行循环
    CycleTypeMixed    CycleType = "mixed"       // 混合循环
)
```

## 6. 契约版本管理

### 6.1 版本兼容性

```go
type VersionCompatibility struct {
    FromVersion string
    ToVersion   string
    Compatible  bool
    BreakingChanges []BreakingChange
}

type BreakingChange struct {
    Field     string
    ChangeType ChangeType
    Description string
}

type ChangeType string

const (
    ChangeTypeRemoved    ChangeType = "removed"
    ChangeTypeModified   ChangeType = "modified"
    ChangeTypeAdded      ChangeType = "added"
    ChangeTypeRenamed    ChangeType = "renamed"
)
```

### 6.2 契约迁移

```go
type ContractMigrator struct {
    registry *ContractRegistry
}

func (m *ContractMigrator) MigrateInput(fromVersion, toVersion string, input map[string]interface{}) (map[string]interface{}, error)
func (m *ContractMigrator) MigrateOutput(fromVersion, toVersion string, output map[string]interface{}) (map[string]interface{}, error)
```

## 7. 契约工具链

### 7.1 契约代码生成

```go
type ContractCodeGenerator struct {
    templateEngine TemplateEngine
}

func (g *ContractCodeGenerator) GenerateGoStructs(contract *AgentContract) (string, error)
func (g *ContractCodeGenerator) GenerateValidator(contract *AgentContract) (string, error)
func (g *ContractCodeGenerator) GenerateTests(contract *AgentContract) (string, error)
```

### 7.2 契约文档生成

```go
type ContractDocGenerator struct {
    markdownEngine MarkdownEngine
}

func (g *ContractDocGenerator) GenerateMarkdown(contract *AgentContract) (string, error)
func (g *ContractDocGenerator) GenerateOpenAPI(contract *AgentContract) (string, error)
```

## 8. 与 Call Human 协议的映射

### 8.1 契约 → Call Human 请求

```go
func ContractToHumanCall(contract *AgentContract, input map[string]interface{}) *HumanCallRequest {
    return &HumanCallRequest{
        CallID: generateCallID(),
        CallType: CallType(contract.AgentInfo.Type),
        Caller: &CallerInfo{
            AgentID: contract.AgentInfo.AgentID,
            AgentType: string(contract.AgentInfo.Type),
            Timestamp: time.Now().Unix(),
        },
        Parameters: input,
        Timeout: contract.SLA.Timeout,
        Priority: 5,  // 默认优先级
        Context: &CallContext{
            TaskID: generateTaskID(),
        },
        Metadata: map[string]string{
            "contract_id": contract.ContractID,
            "contract_version": contract.Version,
        },
    }
}
```

### 8.2 Call Human 响应 → 契约验证

```go
func HumanCallToContractValidation(response *HumanCallResult, contract *AgentContract) *ValidationResult {
    validator := NewOutputSchemaValidator(contract.OutputSchema)
    return validator.Validate(response.Result.ParsedData)
}
```

## 9. 状态管理（可选，里程碑 3）

对于需要维护状态的 Agent，提供 Stateful Agent 模式：

```go
type StatefulAgentContract struct {
    *AgentContract

    // 状态 Schema
    StateSchema *InputSchema `json:"state_schema"`

    // 状态持久化配置
    StatePersistence *StatePersistenceConfig `json:"state_persistence"`

    // 状态迁移策略
    StateMigrationStrategy StateMigrationStrategy `json:"state_migration_strategy"`
}

type StatePersistenceConfig struct {
    // 存储类型
    StorageType StorageType `json:"storage_type"`

    // 存储配置
    StorageConfig map[string]interface{} `json:"storage_config"`

    // TTL（秒）
    TTL int `json:"ttl"`
}

type StorageType string

const (
    StorageTypeMemory  StorageType = "memory"
    StorageTypeRedis   StorageType = "redis"
    StorageTypeDatabase StorageType = "database"
)
```

## 10. 验收标准左移策略

### 10.1 设计目标

实现「验收标准左移」：测试、安全、合规 Agent 在需求定义阶段就介入，预定义的验收标准、合规规则、测试用例直接作为开发 Agent 的输入契约的一部分。

**核心收益**：
- 开发 Agent 从写第一行代码开始，就明确知道「做成什么样能通过验收」
- 消除「我以为符合要求，结果不符合」的偏差
- 消除开发完成后，测试、安全环节的等待时间

### 10.2 分级左移方案

为平衡可行性和复杂度，采用三级左移策略：

#### Level 1：架构 + 测试双前置（推荐先落地）
- **前置 Agent**：架构 Agent、测试 Agent
- **后置 Agent**：安全 Agent、合规 Agent
- **收益**：消除 60% 的需求理解偏差
- **复杂度**：低（2 个 Agent 协调）
- **适用场景**：所有项目

#### Level 2：三角色前置（核心项目）
- **前置 Agent**：架构 Agent、测试 Agent、安全 Agent
- **后置 Agent**：合规 Agent
- **收益**：消除 80% 的需求理解偏差
- **复杂度**：中（3 个 Agent 协调）
- **适用场景**：涉及敏感数据的核心项目

#### Level 3：全角色前置（高合规要求）
- **前置 Agent**：架构、测试、安全、合规
- **收益**：消除 95% 的需求理解偏差
- **复杂度**：高（4 个 Agent 协调）
- **适用场景**：金融、医疗等强合规领域

### 10.3 配置化设计

```go
type LeftShiftConfig struct {
    Level LeftShiftLevel  // L1/L2/L3
    RequiredPreAgents []string     // 指定必须前置的 Agent
    OptionalPreAgents []string     // 可选前置的 Agent
}

type LeftShiftLevel string

const (
    Level1 LeftShiftLevel = "level1"  // 架构 + 测试
    Level2 LeftShiftLevel = "level2"  // 架构 + 测试 + 安全
    Level3 LeftShiftLevel = "level3"  // 全角色
)
```

### 10.4 冲突解决机制

当多个前置 Agent 产生冲突时，按优先级解决：

```go
type ConflictResolution struct {
    Priority []AgentType  // [安全, 架构, 测试, 合规]
    FallbackStrategy string  // "strict" | "warning" | "ignore"
}

type AgentType string

const (
    AgentTypeSecurity   AgentType = "security"
    AgentTypeArchitecture AgentType = "architecture"
    AgentTypeTesting    AgentType = "testing"
    AgentTypeCompliance AgentType = "compliance"
)
```

### 10.5 模板化验收标准

为降低冷启动困境，提供行业通用模板：

- **Web API 模板**：标准的 API 契约模板
- **数据处理管道模板**：ETL 任务契约模板
- **微服务模板**：服务间通信契约模板
- **CLI 工具模板**：命令行工具契约模板

### 10.6 与现有架构的集成

#### 与准入规则引擎的集成
```go
type AdmissionRules struct {
    // ... 现有字段

    // 左移配置
    LeftShiftConfig *LeftShiftConfig `json:"left_shift_config,omitempty"`

    // 前置 Agent 契约引用
    PreAgentContracts []string `json:"pre_agent_contracts,omitempty"`
}
```

#### 与 DAG 调度的集成
- 在 DAG 构建阶段，根据 LeftShiftConfig 自动添加前置 Agent 节点
- 前置 Agent 并行执行，输出作为开发 Agent 的依赖输入
- 冲突检测在 DAG 拓扑排序阶段完成

### 10.7 实施路线

#### 里程碑 1：框架基础
- [ ] 定义 LeftShiftConfig 数据结构
- [ ] 在准入规则引擎中预留扩展点
- [ ] 不强制启用左移，保持传统流程

#### 里程碑 2：Level 1 落地
- [ ] 实现 Level 1（架构 + 测试前置）
- [ ] 实现基础冲突检测
- [ ] 提供 Web API 模板
- [ ] 验证可行性，收集反馈

#### 里程碑 3：Level 2/3 可选
- [ ] 实现 Level 2/3（可选启用）
- [ ] 完善冲突解决机制
- [ ] 扩展模板库
- [ ] 提供性能优化（缓存、预编译）

### 10.8 潜在风险与缓解

#### 风险 1：协调复杂度
- **缓解**：分级实施，从 Level 1 开始
- **缓解**：提供冲突检测和自动解决机制

#### 风险 2：需求变更的级联效应
- **缓解**：契约版本管理（已在架构中设计）
- **缓解**：变更影响分析工具
- **缓解**：强制契约版本一致性检查

#### 风险 3：冷启动困境
- **缓解**：提供模板库
- **缓解**：渐进式左移（从核心模块开始）
- **缓解**：支持降级到传统流程

## 11. 实施路线图

### 里程碑 1：契约基础框架
- [ ] 定义契约元数据结构
- [ ] 实现契约注册表
- [ ] 实现基础准入规则引擎（本地校验）
- [ ] 实现 1-2 个示例子 Agent 验证
- [ ] 与 call human 协议映射验证

### 里程碑 2：契约全量能力
- [ ] 完善准入规则引擎（支持远程校验）
- [ ] 实现契约依赖解析与循环检测
- [ ] 实现契约版本管理
- [ ] 实现契约代码生成工具
- [ ] 建立契约库（复用常见模式）

### 里程碑 3：契约生态成熟
- [ ] 实现契约市场（共享、复用）
- [ ] 实现契约审计与合规检查
- [ ] 实现契约性能优化（缓存、预编译）
- [ ] 实现 Stateful Agent 支持
- [ ] 建立契约治理机制
