# LLM 提供商无关架构设计

## 设计目标

实现 LLM 提供商无关的统一接口，支持 OpenAI、Anthropic、本地模型等多种提供商，确保核心模块不依赖特定 LLM 实现。

## 1. 核心抽象层

### 1.1 提供商接口

```go
type LLMProvider interface {
    // 同步补全
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

    // 流式补全
    StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan CompletionChunk, error)

    // 获取模型信息
    GetModelInfo(model string) (*ModelInfo, error)

    // 列出可用模型
    ListModels() ([]*ModelInfo, error)

    // 健康检查
    HealthCheck(ctx context.Context) error
}
```

### 1.2 请求结构

```go
type CompletionRequest struct {
    // 模型标识
    Model string `json:"model"`

    // 提示词
    Messages []Message `json:"messages"`

    // 温度（0-2）
    Temperature float64 `json:"temperature,omitempty"`

    // 最大 token 数
    MaxTokens int `json:"max_tokens,omitempty"`

    // Top P（0-1）
    TopP float64 `json:"top_p,omitempty"`

    // 停止序列
    Stop []string `json:"stop,omitempty"`

    // 频率惩罚（-2 到 2）
    FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`

    // 存在惩罚（-2 到 2）
    PresencePenalty float64 `json:"presence_penalty,omitempty"`

    // 元数据
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

### 1.3 响应结构

```go
type CompletionResponse struct {
    // 提供商
    Provider string `json:"provider"`

    // 模型
    Model string `json:"model"`

    // 生成的消息
    Message Message `json:"message"`

    // 使用统计
    Usage *Usage `json:"usage"`

    // 完成原因
    FinishReason string `json:"finish_reason"`

    // 时间戳
    Timestamp int64 `json:"timestamp"`
}
```

### 1.4 流式响应

```go
type CompletionChunk struct {
    // 提供商
    Provider string `json:"provider"`

    // 模型
    Model string `json:"model"`

    // 增量内容
    Delta string `json:"delta"`

    // 是否完成
    Done bool `json:"done"`

    // 完成原因
    FinishReason string `json:"finish_reason,omitempty"`

    // 使用统计（仅在完成时提供）
    Usage *Usage `json:"usage,omitempty"`
}
```

## 2. 提供商实现

### 2.1 OpenAI 提供商

**配置结构**：
```go
type OpenAIConfig struct {
    // API Key
    APIKey string `json:"api_key"`

    // 组织 ID（可选）
    OrganizationID string `json:"organization_id,omitempty"`

    // 基础 URL（支持自定义端点）
    BaseURL string `json:"base_url,omitempty"`

    // 默认模型
    DefaultModel string `json:"default_model"`

    // 超时配置
    Timeout time.Duration `json:"timeout"`
}
```

**模型映射**：
- GPT-4
- GPT-4 Turbo
- GPT-3.5 Turbo
- 其他 OpenAI 模型

### 2.2 Anthropic 提供商

**配置结构**：
```go
type AnthropicConfig struct {
    // API Key
    APIKey string `json:"api_key"`

    // 基础 URL
    BaseURL string `json:"base_url,omitempty"`

    // 默认模型
    DefaultModel string `json:"default_model"`

    // 版本（默认 2023-06-01）
    Version string `json:"version,omitempty"`

    // 超时配置
    Timeout time.Duration `json:"timeout"`
}
```

**模型映射**：
- Claude 3 Opus
- Claude 3 Sonnet
- Claude 3 Haiku
- 其他 Anthropic 模型

### 2.3 本地模型提供商

**配置结构**：
```go
type LocalModelConfig struct {
    // 模型路径
    ModelPath string `json:"model_path"`

    // 模型类型（llama.cpp, vllm, ollama 等）
    ModelType string `json:"model_type"`

    // 服务地址（如果使用远程服务）
    ServiceURL string `json:"service_url,omitempty"`

    // GPU 配置
    GPUConfig *GPUConfig `json:"gpu_config,omitempty"`

    // 超时配置
    Timeout time.Duration `json:"timeout"`
}
```

**支持的后端**：
- llama.cpp
- vLLM
- Ollama
- LocalAI
- 其他本地推理框架

### 2.4 提供商注册表

```go
type ProviderRegistry struct {
    providers map[string]LLMProvider
    configs   map[string]interface{}
    mu        sync.RWMutex
}

func (r *ProviderRegistry) Register(name string, provider LLMProvider, config interface{}) error
func (r *ProviderRegistry) Unregister(name string) error
func (r *ProviderRegistry) Get(name string) (LLMProvider, error)
func (r *ProviderRegistry) List() []string
func (r *ProviderRegistry) GetDefault() (LLMProvider, error)
```

## 3. 统一消息格式

### 3.1 消息结构

```go
type Message struct {
    // 角色
    Role MessageRole `json:"role"`

    // 内容
    Content string `json:"content"`

    // 名称（可选，用于 function call）
    Name string `json:"name,omitempty"`

    // 工具调用（可选）
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`

    // 工具调用 ID（可选）
    ToolCallID string `json:"tool_call_id,omitempty"`
}
```

### 3.2 消息角色

```go
type MessageRole string

const (
    RoleSystem    MessageRole = "system"
    RoleUser      MessageRole = "user"
    RoleAssistant MessageRole = "assistant"
    RoleTool      MessageRole = "tool"
)
```

## 4. 模型能力适配

### 4.1 模型信息

```go
type ModelInfo struct {
    // 模型 ID
    ID string `json:"id"`

    // 提供商
    Provider string `json:"provider"`

    // 模型名称
    Name string `json:"name"`

    // 上下文窗口大小
    ContextWindow int `json:"context_window"`

    // 最大输出 token
    MaxOutputTokens int `json:"max_output_tokens"`

    // 支持的功能
    Capabilities []ModelCapability `json:"capabilities"`

    // 成本信息
    Pricing *Pricing `json:"pricing,omitempty"`

    // 元数据
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

### 4.2 模型能力

```go
type ModelCapability string

const (
    CapabilityText       ModelCapability = "text"
    CapabilityVision     ModelCapability = "vision"
    CapabilityFunction   ModelCapability = "function"
    CapabilityStreaming  ModelCapability = "streaming"
    CapabilityJSON       ModelCapability = "json"
)
```

### 4.3 成本信息

```go
type Pricing struct {
    // 输入价格（每 1K tokens）
    InputPrice float64 `json:"input_price"`

    // 输出价格（每 1K tokens）
    OutputPrice float64 `json:"output_price"`

    // 货币单位
    Currency string `json:"currency"`
}
```

## 5. 配置管理

### 5.1 配置文件结构

```yaml
llm:
  # 默认提供商
  default_provider: "openai"

  # 提供商配置
  providers:
    openai:
      api_key: "${OPENAI_API_KEY}"
      base_url: "https://api.openai.com/v1"
      default_model: "gpt-4-turbo-preview"
      timeout: 60s

    anthropic:
      api_key: "${ANTHROPIC_API_KEY}"
      base_url: "https://api.anthropic.com"
      default_model: "claude-3-opus-20240229"
      timeout: 60s

    local:
      model_path: "/models/llama-2-7b.gguf"
      model_type: "llama-cpp"
      service_url: "http://localhost:8080"
      timeout: 120s

  # 模型别名
  model_aliases:
    gpt4: "openai:gpt-4-turbo-preview"
    claude: "anthropic:claude-3-opus-20240229"
    local: "local:llama-2-7b"
```

### 5.2 环境变量支持

- `AXIS_LLM_DEFAULT_PROVIDER`
- `AXIS_LLM_OPENAI_API_KEY`
- `AXIS_LLM_ANTHROPIC_API_KEY`
- `AXIS_LLM_LOCAL_MODEL_PATH`

## 6. 错误处理

### 6.1 统一错误类型

```go
type LLMError struct {
    // 错误码
    Code string `json:"code"`

    // 错误消息
    Message string `json:"message"`

    // 提供商
    Provider string `json:"provider"`

    // 原始错误
    OriginalError error `json:"-"`

    // 可重试
    Retryable bool `json:"retryable"`
}
```

### 6.2 错误码定义

| 错误码 | 描述 | 可重试 |
|--------|------|--------|
| `LLM_0001` | 配置错误 | 否 |
| `LLM_0002` | 认证失败 | 否 |
| `LLM_0003` | 模型不存在 | 否 |
| `LLM_0101` | 速率限制 | 是 |
| `LLM_0102` | 配额不足 | 否 |
| `LLM_0201` | 网络错误 | 是 |
| `LLM_0202` | 超时 | 是 |
| `LLM_0301` | 提供商错误 | 否 |
| `LLM_0302` | 模型错误 | 否 |

## 7. 使用统计

### 7.1 使用信息

```go
type Usage struct {
    // 提示词 token 数
    PromptTokens int `json:"prompt_tokens"`

    // 完成 token 数
    CompletionTokens int `json:"completion_tokens"`

    // 总 token 数
    TotalTokens int `json:"total_tokens"`

    // 成本（如果可用）
    Cost float64 `json:"cost,omitempty"`
}
```

### 7.2 配额管理

```go
type QuotaManager struct {
    // 配额限制
    limits map[string]QuotaLimit

    // 使用统计
    usage map[string]QuotaUsage

    mu sync.RWMutex
}

type QuotaLimit struct {
    // 时间窗口
    Window time.Duration

    // 最大请求数
    MaxRequests int

    // 最大 token 数
    MaxTokens int
}

type QuotaUsage struct {
    // 当前请求数
    Requests int

    // 当前 token 数
    Tokens int

    // 窗口开始时间
    WindowStart time.Time
}
```

## 8. 缓存策略

### 8.1 请求缓存

```go
type CacheConfig struct {
    // 缓存启用
    Enabled bool `json:"enabled"`

    // TTL
    TTL time.Duration `json:"ttl"`

    // 最大条目数
    MaxEntries int `json:"max_entries"`

    // 缓存键生成策略
    KeyStrategy CacheKeyStrategy `json:"key_strategy"`
}
```

### 8.2 缓存键策略

```go
type CacheKeyStrategy string

const (
    CacheKeyExact    CacheKeyStrategy = "exact"     // 精确匹配
    CacheKeySemantic CacheKeyStrategy = "semantic"  // 语义相似（未来）
)
```

## 9. 可观测性

### 9.1 指标

- 请求数（按提供商、模型、状态）
- 延迟（P50, P95, P99）
- Token 使用量
- 成本
- 错误率
- 缓存命中率

### 9.2 链路追踪

每个请求包含：
- Trace ID
- Span ID
- 提供商
- 模型
- 请求/响应元数据

## 10. 安全性

### 10.1 API 密钥管理

- 支持环境变量
- 支持配置文件（加密）
- 支持密钥管理服务（AWS Secrets Manager、HashiCorp Vault）

### 10.2 内容过滤

- 可选的内容安全过滤
- PII 检测与脱敏
- 敏感信息拦截

## 11. 测试策略

### 11.1 Mock 提供商

```go
type MockProvider struct {
    responses map[string]*CompletionResponse
    errors    map[string]error
    latency   time.Duration
}
```

### 11.2 集成测试

- 每个提供商的集成测试
- 使用测试 API Key
- 限制测试配额

## 12. 性能优化

### 12.1 连接池

- HTTP 连接复用
- 连接池大小可配置
- 连接超时管理

### 12.2 批量请求

- 支持批量请求（如果提供商支持）
- 批量大小可配置
- 批量超时管理

### 12.3 异步请求

- 异步请求队列
- 优先级调度
- 结果回调
