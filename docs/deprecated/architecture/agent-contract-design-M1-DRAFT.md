# Agent 契约模型设计（里程碑1简化版）

## 设计目标

主 Agent 的核心职责从「发指令」变为「定契约」，所有子 Agent 封装为标准化无状态函数，与 call human 协议完全同构，彻底消除调度歧义与理解偏差。

## 里程碑1设计原则（奥卡姆剃刀）
- **最小可行**：只实现输入输出 Schema 验证
- **渐进增强**：SLA、准入规则、异常码体系在后续里程碑添加
- **聚焦核心**：里程碑1聚焦于"契约定义 + 输入输出验证"的可行性验证

## 1. 契约核心要素（里程碑1最小集）

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

    // 字段定义
    Fields []FieldDef `json:"fields"`

    // 必需字段列表
    RequiredFields []string `json:"required_fields"`
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
)
```

**示例**：
```json
{
  "schema_id": "code_review_input_v1",
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
  ],
  "required_fields": ["code", "language"]
}
```

### 1.2 输出 Schema (Output Schema)

**设计原则**：
- 标准化的出参结构
- 明确字段类型和结构

**数据结构**：
```go
type OutputSchema struct {
    // Schema ID
    SchemaID string `json:"schema_id"`

    // 字段定义
    Fields []FieldDef `json:"fields"`

    // 必需字段列表
    RequiredFields []string `json:"required_fields"`
}
```

**示例**：
```json
{
  "schema_id": "code_review_output_v1",
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
          }
        ]
      }
    },
    {
      "name": "overall_score",
      "type": "integer",
      "description": "总体评分（0-100）",
      "required": true
    }
  ],
  "required_fields": ["issues", "overall_score"]
}
```

**里程碑1暂不包含**：
- ❌ 验收标准（AcceptanceCriteria）
- ❌ 格式要求（FormatRequirements）
- ❌ 合规规则（ComplianceRules）
- ❌ 置信度阈值（ConfidenceThreshold）
- ❌ 复杂验证规则（ValidationRules）
- ❌ SLA 约定（超时、重试、熔断）
- ❌ 准入规则（本地/远程校验）
- ❌ 异常码体系

## 2. 契约定义（里程碑1最小集）

### 2.1 契约元数据

**数据结构**：
```go
type AgentContract struct {
    // 契约 ID（全局唯一）
    ContractID string `json:"contract_id"`

    // Agent 信息
    AgentInfo *AgentInfo `json:"agent_info"`

    // 输入 Schema
    InputSchema *InputSchema `json:"input_schema"`

    // 输出 Schema
    OutputSchema *OutputSchema `json:"output_schema"`
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
}

type AgentType string

const (
    AgentTypeTask AgentType = "task" // 任务型 Agent
)
```

**示例**：
```json
{
  "contract_id": "code_review_agent_v1",
  "agent_info": {
    "agent_id": "agent_code_review",
    "name": "Code Review Agent",
    "description": "执行代码评审任务",
    "type": "task"
  },
  "input_schema": {
    "schema_id": "code_review_input_v1",
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
    ],
    "required_fields": ["code", "language"]
  },
  "output_schema": {
    "schema_id": "code_review_output_v1",
    "fields": [
      {
        "name": "issues",
        "type": "array",
        "description": "发现的问题列表",
        "required": true
      },
      {
        "name": "overall_score",
        "type": "integer",
        "description": "总体评分（0-100）",
        "required": true
      }
    ],
    "required_fields": ["issues", "overall_score"]
  }
}
```

**里程碑1暂不包含**：
- ❌ 契约版本管理
- ❌ 依赖声明
- ❌ 契约状态管理（draft/active/deprecated/retired）
- ❌ 创建时间和更新时间

## 3. 契约执行（里程碑1简化版）

### 3.1 契约执行器

**职责**：基于契约的 Agent 执行引擎

**核心功能**：
- 输入 Schema 验证
- 输出 Schema 验证

**关键接口**：
```go
type ContractExecutor struct {
    registry *ContractRegistry
}

func (e *ContractExecutor) Execute(contractID string, input map[string]interface{}) (*ExecutionResult, error)
func (e *ContractExecutor) ValidateInput(contractID string, input map[string]interface{}) error
func (e *ContractExecutor) ValidateOutput(contractID string, output map[string]interface{}) error
```

### 3.2 契约注册表（简化版）

**职责**：契约注册与查询

**核心功能**：
- 契约注册
- 契约查询

**关键接口**：
```go
type ContractRegistry struct {
    contracts map[string]*AgentContract
    mu        sync.RWMutex
}

func (r *ContractRegistry) Register(contract *AgentContract) error
func (r *ContractRegistry) Get(contractID string) (*AgentContract, error)
```

## 4. 与 Call Human 协议的映射（里程碑1简化版）

### 4.1 契约 → Call Human 请求

```go
func ContractToHumanCall(contract *AgentContract, input map[string]interface{}) *HumanCallRequest {
    return &HumanCallRequest{
        CallID: generateCallID(),
        Caller: &CallerInfo{
            AgentID: contract.AgentInfo.AgentID,
            AgentType: string(contract.AgentInfo.Type),
        },
        Parameters: input,
    }
}
```

### 4.2 Call Human 响应 → 契约验证

```go
func HumanCallToContractValidation(response *HumanCallResult, contract *AgentContract) error {
    validator := NewOutputSchemaValidator(contract.OutputSchema)
    return validator.Validate(response.Result.ParsedData)
}
```

## 5. 实施路线图

### 里程碑1：契约基础框架
- [ ] 定义契约元数据结构（输入输出 Schema）
- [ ] 实现契约注册表
- [ ] 实现输入输出 Schema 验证
- [ ] 实现 1-2 个示例子 Agent 验证
- [ ] 与 call human 协议映射验证

### 里程碑2：契约全量能力
- [ ] 完善准入规则引擎（支持远程校验）
- [ ] 实现契约依赖解析与循环检测
- [ ] 实现契约版本管理
- [ ] 实现契约代码生成工具
- [ ] 建立契约库（复用常见模式）

### 里程碑3：契约生态成熟
- [ ] 实现契约文档生成
- [ ] 实现契约测试框架
- [ ] 实现契约市场
