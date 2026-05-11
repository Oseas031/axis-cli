package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestOpenAIProvider_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}

		// Return mock response
		resp := openaiResponse{
			ID:      "chatcmpl_123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role      string `json:"role"`
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					}{
						Role:    "assistant",
						Content: "Hello from GPT-4",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "gpt-4",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)

	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "hello"},
	}

	resp, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.Output["text"] != "Hello from GPT-4" {
		t.Errorf("Expected text output, got %v", resp.Output["text"])
	}
	if resp.InputTokens != 100 {
		t.Errorf("Expected InputTokens=100, got %d", resp.InputTokens)
	}
	if resp.OutputTokens != 50 {
		t.Errorf("Expected OutputTokens=50, got %d", resp.OutputTokens)
	}
}

func TestOpenAIChatCompletionsURL(t *testing.T) {
	tests := map[string]string{
		"https://api.openai.com":       "https://api.openai.com/v1/chat/completions",
		"https://api.openai.com/":      "https://api.openai.com/v1/chat/completions",
		"https://api.minimaxi.com/v1":  "https://api.minimaxi.com/v1/chat/completions",
		"https://api.minimaxi.com/v1/": "https://api.minimaxi.com/v1/chat/completions",
	}

	for baseURL, expected := range tests {
		if got := openAIChatCompletionsURL(baseURL); got != expected {
			t.Fatalf("expected %s for base URL %s, got %s", expected, baseURL, got)
		}
	}
}

func TestOpenAIProvider_Execute_WithTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openaiResponse{
			ID:      "chatcmpl_123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role      string `json:"role"`
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					}{
						Role:    "assistant",
						Content: "Using tool",
						ToolCalls: []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						}{
							{
								ID:   "call_123",
								Type: "function",
								Function: struct {
									Name      string `json:"name"`
									Arguments string `json:"arguments"`
								}{
									Name:      "bash",
									Arguments: `{"command":"ls"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "gpt-4",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)

	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "use bash"},
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
	if len(resp.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "bash" {
		t.Errorf("Expected tool call name 'bash', got %s", resp.ToolCalls[0].Name)
	}
}

func TestOpenAIProvider_Execute_WithHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openaiResponse{
			ID:      "chatcmpl_123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role      string `json:"role"`
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					}{
						Role:    "assistant",
						Content: "Second response",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     150,
				CompletionTokens: 75,
				TotalTokens:      225,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "gpt-4",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)

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

func TestOpenAIProvider_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "gpt-4",
		apiKey:     "invalid-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)

	_, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"msg": "hello"},
	})
	if err == nil {
		t.Error("Expected error on API failure")
	}
}

func TestOpenAIProvider_Execute_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "gpt-4",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)

	_, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"msg": "hello"},
	})
	if err == nil {
		t.Error("Expected error on invalid JSON response")
	}
}

func TestOpenAIProvider_Execute_HistoryIncludesToolCallID(t *testing.T) {
	var capturedReq []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq, _ = io.ReadAll(r.Body)
		resp := openaiResponse{
			ID:     "chatcmpl_123",
			Object: "chat.completion",
			Model:  "gpt-4",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role      string `json:"role"`
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					}{
						Role:    "assistant",
						Content: "Done",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens: 50, CompletionTokens: 25, TotalTokens: 75,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model: "gpt-4", apiKey: "test-key", baseURL: server.URL,
		timeout: 30, maxRetries: 3, httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)

	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "second"},
		History: []types.ModelMessage{
			{Role: "assistant", ToolCalls: []types.ToolCall{{ID: "call-abc", Name: "bash", Input: map[string]any{"command": "ls"}}}},
			{Role: "tool", ToolCallID: "call-abc", Content: `{"stdout":"file.txt"}`},
		},
	}
	_, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(capturedReq, &parsed); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}
	messages := parsed["messages"].([]any)
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}
	toolMsg := messages[1].(map[string]any)
	if toolMsg["role"] != "tool" {
		t.Fatalf("expected message[1] role=tool, got %v", toolMsg["role"])
	}
	if toolMsg["tool_call_id"] != "call-abc" {
		t.Fatalf("expected tool_call_id=call-abc in request, got %v", toolMsg["tool_call_id"])
	}
}

func TestOpenAIProvider_Execute_GeneratesSyntheticIDForEmptyToolCallID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openaiResponse{
			ID:     "chatcmpl_123",
			Object: "chat.completion",
			Model:  "gpt-4",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role      string `json:"role"`
						Content   string `json:"content"`
						ToolCalls []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					}{
						Role:    "assistant",
						Content: "Using tool",
						ToolCalls: []struct {
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						}{
							{
								ID:   "",
								Type: "function",
								Function: struct {
									Name      string `json:"name"`
									Arguments string `json:"arguments"`
								}{Name: "bash", Arguments: `{"command":"ls"}`},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model: "gpt-4", apiKey: "test-key", baseURL: server.URL,
		timeout: 30, maxRetries: 3, httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)
	req := &ModelRequest{
		ContractID: "test-contract",
		Input:      map[string]any{"message": "use bash"},
		Tools: []types.ToolDefinition{
			{Name: "bash", Description: "Run a bash command", Parameters: []types.FieldDef{{Name: "command", Type: types.FieldTypeString, Required: true}}},
		},
	}
	resp, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ID != "call-1" {
		t.Fatalf("expected synthetic id call-1 for empty tool call id, got %q", resp.ToolCalls[0].ID)
	}
}

func TestOpenAIProvider_Execute_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openaiResponse{
			ID:      "chatcmpl_123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{}, // Empty choices
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &providerConfig{
		model:      "gpt-4",
		apiKey:     "test-key",
		baseURL:    server.URL,
		timeout:    30,
		maxRetries: 3,
		httpClient: server.Client(),
	}
	p := newOpenAIProvider(cfg)

	_, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"msg": "hello"},
	})
	if err == nil {
		t.Error("Expected error on empty choices")
	}
}
