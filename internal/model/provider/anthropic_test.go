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
			Content: []struct {
				Type string "json:\"type\""
				Text string "json:\"text,omitempty\""
			}{
				{Type: "text", Text: "Hello from Claude"},
			},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{
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
			Content: []struct {
				Type string "json:\"type\""
				Text string "json:\"text,omitempty\""
			}{
				{Type: "text", Text: "Response"},
			},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{
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
			Content: []struct {
				Type string "json:\"type\""
				Text string "json:\"text,omitempty\""
			}{
				{Type: "text", Text: "Response"},
			},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{
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
