# M4 Requirements: Real LLM Integration + Extended Tools

**Status**: Draft
**Last Updated**: 2026-05-09

## 1. Overview

M4打通真实LLM作为推理后端，同时扩展工具集（文件读写、HTTP client），为Agent提供更完整的工作能力。

## 2. Goals

### 2.1 Real LLM Integration
- [ ] Anthropic provider (Claude family)
- [ ] OpenAI provider (GPT family)
- [ ] Provider configuration via functional options or config file
- [ ] Token accounting (input/output count per request)
- [ ] Streaming support (for long outputs)

### 2.2 Extended Tools
- [ ] FileReadTool — read files with path validation
- [ ] FileWriteTool — write files with path validation
- [ ] HTTPClientTool — make HTTP requests
- [ ] Tool permission scopes (read-only, write-only, network, etc.)

### 2.3 Security
- [ ] Tool execution sandboxing (allowlists, deny lists)
- [ ] Circuit breaker for runaway tool loops
- [ ] Safe JSON serialization in tool results

## 3. Non-Goals

- Real sandbox (gVisor, WASM) — deferred to M5
- Token truncation/summarization — deferred to M5
- Streaming cancellation — deferred to M5
- Multiple concurrent providers — one active at a time

## 4. Interface Boundaries

### ModelProvider
The interface stays the same:
```go
type ModelProvider interface {
    Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error)
}
```

New implementations:
- `AnthropicModelProvider` — calls Anthropic API
- `OpenAIModelProvider` — calls OpenAI API

### Tool
The interface stays the same:
```go
type Tool interface {
    Name() string
    Schema() types.ToolDefinition
    Execute(ctx context.Context, input map[string]any) (map[string]any, error)
}
```

New tools: FileReadTool, FileWriteTool, HTTPClientTool

## 5. Configuration

M4 introduces a provider configuration mechanism:
```go
// Example
provider, err := NewProvider("anthropic",
    WithModel("claude-sonnet-4-5"),
    WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
    WithBaseURL("https://api.anthropic.com"),
)
```

## 6. Dependencies

- Go stdlib only for core modules
- External HTTP client: standard `net/http` (no SDK dependencies)
- No external LLM SDKs (direct API calls via net/http)
