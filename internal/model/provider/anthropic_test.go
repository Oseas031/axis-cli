package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestAnthropicProvider_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected x-api-key header 'test-key', got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version header '2023-06-01', got %s", r.Header.Get("anthropic-version"))
		}

		// Return mock response
		resp := anthropicResponse{
			ID:    "msg_123",
			Type:  "message",
			Role:  "assistant",
			Model: "claude-sonnet-4-5",
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Hello from Claude"},
			},
			Usage: anthropicUsage{
				InputTokens:  100,
				OutputTokens: 50,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "claude-sonnet-4-5",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newAnthropicProvider(cfg)

	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "hello"},
	}

	resp, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.Output["text"] != "Hello from Claude" {
		t.Errorf("Expected text output, got %v", resp.Output["text"])
	}
	if resp.InputTokens != 100 {
		t.Errorf("Expected InputTokens=100, got %d", resp.InputTokens)
	}
	if resp.OutputTokens != 50 {
		t.Errorf("Expected OutputTokens=50, got %d", resp.OutputTokens)
	}
}

func TestAnthropicProvider_Execute_WithTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			ID:    "msg_123",
			Type:  "message",
			Role:  "assistant",
			Model: "claude-sonnet-4-5",
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Response"},
			},
			Usage: anthropicUsage{
				InputTokens:  100,
				OutputTokens: 50,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "claude-sonnet-4-5",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newAnthropicProvider(cfg)

	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "hello"},
		Tools: []types.ToolDefinition{
			{
				Name:        "bash",
				Description: "Run a bash command",
				Parameters: []types.FieldDef{
					{Name: "command", Type: types.FieldTypeString, Required: true},
				},
			},
		},
	}

	resp, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.Output == nil {
		t.Error("Expected non-nil output")
	}
}

func TestAnthropicProvider_Execute_WithHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			ID:    "msg_123",
			Type:  "message",
			Role:  "assistant",
			Model: "claude-sonnet-4-5",
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Response"},
			},
			Usage: anthropicUsage{
				InputTokens:  150,
				OutputTokens: 75,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "claude-sonnet-4-5",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newAnthropicProvider(cfg)

	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "second"},
		History: []types.ModelMessage{
			{Role: "user", Content: "first message"},
			{Role: "assistant", Content: "first response"},
		},
	}

	resp, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.InputTokens != 150 {
		t.Errorf("Expected InputTokens=150, got %d", resp.InputTokens)
	}
}

func TestAnthropicProvider_Execute_ToolSchemaIncludesProperties(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		resp := anthropicResponse{
			ID:    "msg_123",
			Type:  "message",
			Role:  "assistant",
			Model: "claude-sonnet-4-5",
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Hello from Claude"},
			},
			Usage: anthropicUsage{InputTokens: 10, OutputTokens: 5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "claude-sonnet-4-5",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newAnthropicProvider(cfg)

	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "hello"},
		Tools: []types.ToolDefinition{
			{
				Name:        "bash",
				Description: "Run a bash command",
				Parameters: []types.FieldDef{
					{Name: "command", Type: types.FieldTypeString, Required: true, Description: "The command"},
					{Name: "timeout", Type: types.FieldTypeInt, Required: false, Description: "Timeout seconds"},
				},
			},
		},
	}

	_, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	toolsRaw, ok := capturedBody["tools"].([]any)
	if !ok || len(toolsRaw) == 0 {
		t.Fatal("expected tools array in request body")
	}
	tool0 := toolsRaw[0].(map[string]any)
	schema := tool0["input_schema"].(map[string]any)
	if schema["type"] != "object" {
		t.Fatalf("expected input_schema.type=object, got %v", schema["type"])
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected input_schema.properties to be present")
	}
	if _, hasCommand := props["command"]; !hasCommand {
		t.Fatal("expected properties to contain 'command'")
	}
	reqArr, ok := schema["required"].([]any)
	if !ok || len(reqArr) != 1 || reqArr[0] != "command" {
		t.Fatalf("expected required=[command], got %v", schema["required"])
	}
}

func TestAnthropicProvider_Execute_ToolUseResponseParsed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			ID:    "msg_123",
			Type:  "message",
			Role:  "assistant",
			Model: "claude-sonnet-4-5",
			Content: []anthropicContentBlock{
				{Type: "text", Text: "I'll run the command"},
				{Type: "tool_use", ID: "tu_01", Name: "bash", Input: json.RawMessage(`{"command":"echo hi"}`)},
			},
			Usage: anthropicUsage{InputTokens: 10, OutputTokens: 5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "claude-sonnet-4-5",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newAnthropicProvider(cfg)

	resp, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"msg": "hello"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "bash" {
		t.Errorf("expected tool name 'bash', got %s", resp.ToolCalls[0].Name)
	}
	if resp.ToolCalls[0].ID != "tu_01" {
		t.Errorf("expected tool id 'tu_01', got %s", resp.ToolCalls[0].ID)
	}
	cmd, ok := resp.ToolCalls[0].Input["command"].(string)
	if !ok || cmd != "echo hi" {
		t.Errorf("expected input command 'echo hi', got %v", resp.ToolCalls[0].Input["command"])
	}
}

func TestAnthropicProvider_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid request"}`))
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "claude-sonnet-4-5",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newAnthropicProvider(cfg)

	_, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"msg": "hello"},
	})
	if err == nil {
		t.Error("Expected error on API failure")
	}
}

func TestAnthropicProvider_Execute_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "claude-sonnet-4-5",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newAnthropicProvider(cfg)

	_, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"msg": "hello"},
	})
	if err == nil {
		t.Error("Expected error on invalid JSON response")
	}
}
