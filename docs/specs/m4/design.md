# M4 Design: Real LLM Integration + Extended Tools

**Status**: Draft
**Last Updated**: 2026-05-09

## 1. Architecture Overview

M4 extends the existing ModelProvider and Tool infrastructure without changing interfaces.

```
                    ┌─────────────────────┐
                    │  Orchestrator       │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │  ContractExecutor   │
                    │  (multi-turn loop)  │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼────────┐ ┌────▼────┐ ┌────────▼────────┐
     │ BashTool        │ │ FileTools│ │ HTTPClientTool  │
     └─────────────────┘ └─────────┘ └─────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │  ModelProvider      │
                    │  (pluggable)       │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼────────┐ ┌────▼────┐ ┌────────▼────────┐
     │ MockModelProvider│ │Anthropic│ │   OpenAI        │
     │ (existing)      │ │Provider  │ │   Provider      │
     └─────────────────┘ └─────────┘ └─────────────────┘
```

## 2. Provider Configuration

### 2.1 Functional Options Pattern

```go
type ProviderOption func(*providerConfig)

type providerConfig struct {
    model     string
    apiKey    string
    baseURL   string
    timeout   time.Duration
    maxRetries int
}

func WithModel(model string) ProviderOption { ... }
func WithAPIKey(key string) ProviderOption { ... }
func WithBaseURL(url string) ProviderOption { ... }
func WithTimeout(d time.Duration) ProviderOption { ... }
func WithMaxRetries(n int) ProviderOption { ... }
```

### 2.2 NewProvider Factory

```go
func NewProvider(name string, opts ...ProviderOption) (provider.ModelProvider, error) {
    switch name {
    case "mock":
        return NewMockModelProvider(), nil
    case "echo":
        return NewEchoModelProvider(), nil
    case "anthropic":
        return NewAnthropicProvider(opts...)
    case "openai":
        return NewOpenAIProvider(opts...)
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}
```

## 3. LLM Provider Implementations

### 3.1 AnthropicProvider

```go
type AnthropicProvider struct {
    config   providerConfig
    httpClient *http.Client
}

func (p *AnthropicProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
    // Build Anthropic API request
    // Use net/http only (no SDK)
    // Handle streaming via chunked responses
    // Return ModelResponse
}
```

**API Endpoint**: `POST /v1/messages`
**Auth**: `x-api-key` header

### 3.2 OpenAIProvider

```go
type OpenAIProvider struct {
    config   providerConfig
    httpClient *http.Client
}

func (p *OpenAIProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
    // Build OpenAI API request
    // Use net/http only (no SDK)
    // Handle streaming via SSE
    // Return ModelResponse
}
```

**API Endpoint**: `POST /v1/chat/completions`
**Auth**: `Authorization: Bearer` header

## 4. Extended Tools

### 4.1 FileReadTool

```go
type FileReadTool struct {
    allowedDirs []string  // permission scope
}

func (t *FileReadTool) Name() string { return "file_read" }

func (t *FileReadTool) Schema() types.ToolDefinition {
    return types.ToolDefinition{
        Name:        "file_read",
        Description: "Read contents of a file",
        Parameters: []types.FieldDef{
            {Name: "path", Type: types.FieldTypeString, Required: true, Description: "Absolute path to file"},
        },
    }
}

func (t *FileReadTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    path, _ := input["path"].(string)
    // Validate path is within allowedDirs
    // Read and return content
}
```

### 4.2 FileWriteTool

```go
type FileWriteTool struct {
    allowedDirs []string
}

func (t *FileWriteTool) Name() string { return "file_write" }

func (t *FileWriteTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    path, _ := input["path"].(string)
    content, _ := input["content"].(string)
    // Validate path within allowedDirs
    // Write content
}
```

### 4.3 HTTPClientTool

```go
type HTTPClientTool struct {
    allowedHosts []string  // permission scope
}

func (t *HTTPClientTool) Name() string { return "http_request" }

func (t *HTTPClientTool) Schema() types.ToolDefinition {
    return types.ToolDefinition{
        Name:        "http_request",
        Description: "Make an HTTP request",
        Parameters: []types.FieldDef{
            {Name: "method", Type: types.FieldTypeString, Required: true, Description: "GET, POST, PUT, DELETE"},
            {Name: "url", Type: types.FieldTypeString, Required: true, Description: "URL to request"},
            {Name: "headers", Type: types.FieldTypeObject, Required: false, Description: "HTTP headers"},
            {Name: "body", Type: types.FieldTypeString, Required: false, Description: "Request body"},
        },
    }
}
```

## 5. Security: Tool Permissions

Each tool has a permission scope that restricts what it can access:

| Scope | BashTool | FileReadTool | FileWriteTool | HTTPClientTool |
|-------|----------|--------------|---------------|----------------|
| filesystem:read | ✓ | ✓ | ✗ | ✗ |
| filesystem:write | ✓ | ✗ | ✓ | ✗ |
| network | ✓ | ✗ | ✗ | ✓ |
| subprocess | ✓ | ✗ | ✗ | ✗ |

Permissions are set at tool registration time:
```go
orchestrator := NewOrchestrator(
    WithToolPermissions(map[string][]string{
        "file_read":  {"filesystem:read"},
        "file_write": {"filesystem:write"},
        "http_request": {"network"},
    }),
)
```

## 6. Error Handling

### 6.1 Circuit Breaker

The ContractExecutor tracks tool errors per turn. If errors exceed a threshold (e.g., 5 consecutive errors), abort the execution:

```go
const maxConsecutiveErrors = 5

for turn := 0; turn < maxTurns; turn++ {
    result, err := tool.Execute(ctx, toolInput)
    if err != nil {
        consecutiveErrors++
        if consecutiveErrors >= maxConsecutiveErrors {
            return nil, fmt.Errorf("circuit breaker: too many tool errors")
        }
        history = append(history, ModelMessage{Role: "user", Content: err.Error()})
        continue
    }
    consecutiveErrors = 0
}
```

### 6.2 Safe JSON Serialization

Wrap tool result serialization to prevent panics:

```go
func safeMarshal(v any) ([]byte, error) {
    defer func() {
        if r := recover(); r != nil {
            // log and return error
        }
    }()
    return json.Marshal(v)
}
```

## 7. Token Accounting

Add token tracking to ModelResponse:

```go
type ModelResponse struct {
    Output         string
    ToolCalls      []ToolCall
    InputTokens    int
    OutputTokens   int
}
```

Providers calculate tokens based on model:
- Anthropic: use `usage` from API response
- OpenAI: use `usage` from API response
- Mock/Echo: estimate based on character count

## 8. Streaming Support

For long outputs, providers support chunked responses:

```go
type ModelResponse struct {
    Output         string
    ToolCalls      []ToolCall
    InputTokens    int
    OutputTokens   int
    Done           bool
}
```

Streaming is synchronous per turn — each turn waits for complete response. True streaming (progressive output) is deferred to M5.

## 9. File Structure

```
internal/model/
  provider/
    provider.go      # interface (unchanged)
    mock.go         # (unchanged)
    echo.go         # (unchanged)
    anthropic.go    # NEW
    openai.go       # NEW
    registry.go     # (unchanged)
  tool/
    tool.go        # interface (unchanged)
    bash.go        # (unchanged)
    file_read.go   # NEW
    file_write.go  # NEW
    http_client.go  # NEW
```
