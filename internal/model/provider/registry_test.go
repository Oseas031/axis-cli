package provider

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewProvider_Mock(t *testing.T) {
	p, err := NewProvider("mock")
	if err != nil {
		t.Fatalf("NewProvider(\"mock\") should succeed: %v", err)
	}
	if p == nil {
		t.Fatal("NewProvider(\"mock\") should return non-nil provider")
	}
	resp, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"x": "y"},
	})
	if err != nil {
		t.Fatalf("Execute should succeed: %v", err)
	}
	if resp.Output["provider"] != "mock" {
		t.Errorf("Mock provider should set provider=mock, got %v", resp.Output["provider"])
	}
}

func TestNewProvider_Echo(t *testing.T) {
	p, err := NewProvider("echo")
	if err != nil {
		t.Fatalf("NewProvider(\"echo\") should succeed: %v", err)
	}
	resp, err := p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"x": "y"},
	})
	if err != nil {
		t.Fatalf("Execute should succeed: %v", err)
	}
	if resp.Output["provider"] != "echo" {
		t.Errorf("Echo provider should set provider=echo, got %v", resp.Output["provider"])
	}
}

func TestNewProvider_Anthropic(t *testing.T) {
	p, err := NewProvider("anthropic", WithAPIKey("test-key"), WithModel("claude-sonnet-4-5"), WithMaxRetries(0))
	if err != nil {
		t.Fatalf("NewProvider(\"anthropic\") should succeed: %v", err)
	}
	if p == nil {
		t.Fatal("NewProvider(\"anthropic\") should return non-nil provider")
	}
	// Execute will fail without valid API, but constructor should work
	_, _ = p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"x": "y"},
	})
}

func TestNewProvider_OpenAI(t *testing.T) {
	p, err := NewProvider("openai", WithAPIKey("test-key"), WithModel("gpt-4"), WithMaxRetries(0))
	if err != nil {
		t.Fatalf("NewProvider(\"openai\") should succeed: %v", err)
	}
	if p == nil {
		t.Fatal("NewProvider(\"openai\") should return non-nil provider")
	}
	// Execute will fail without valid API, but constructor should work
	_, _ = p.Execute(context.Background(), &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"x": "y"},
	})
}

func TestNewProvider_OpenAICompatibleProviders(t *testing.T) {
	for _, name := range []string{"deepseek", "minimax"} {
		p, err := NewProvider(name, WithAPIKey("test-key"), WithModel("test-model"))
		if err != nil {
			t.Fatalf("NewProvider(%q) should succeed: %v", name, err)
		}
		if p == nil {
			t.Fatalf("NewProvider(%q) should return non-nil provider", name)
		}
	}
}

func TestNewProvider_OpenAICompatibleDefaultBaseURLs(t *testing.T) {
	tests := map[string]string{
		"deepseek": "https://api.deepseek.com",
		"minimax":  "https://api.minimaxi.com/v1",
	}
	for name, expectedBaseURL := range tests {
		p, err := NewProvider(name, WithAPIKey("test-key"), WithModel("test-model"))
		if err != nil {
			t.Fatalf("NewProvider(%q) should succeed: %v", name, err)
		}
		openaiProvider, ok := p.(*OpenAIProvider)
		if !ok {
			t.Fatalf("NewProvider(%q) should use OpenAI-compatible provider, got %T", name, p)
		}
		if openaiProvider.config.baseURL != expectedBaseURL {
			t.Fatalf("expected %s baseURL %s, got %s", name, expectedBaseURL, openaiProvider.config.baseURL)
		}
	}
}

func TestNewProvider_Unknown(t *testing.T) {
	p, err := NewProvider("unknown")
	if err == nil {
		t.Fatal("NewProvider(\"unknown\") should return error")
	}
	if p != nil {
		t.Error("NewProvider(\"unknown\") should return nil provider")
	}
}

func TestProviderOption_WithModel(t *testing.T) {
	p, err := NewProvider("anthropic", WithModel("claude-3-opus"), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewProvider with WithModel should succeed: %v", err)
	}
	if p == nil {
		t.Fatal("Provider should not be nil")
	}
}

func TestProviderOption_WithAPIKey(t *testing.T) {
	_, err := NewProvider("anthropic", WithAPIKey("sk-ant-xxx"))
	if err != nil {
		t.Fatalf("NewProvider with WithAPIKey should succeed: %v", err)
	}
}

func TestProviderOption_WithBaseURL(t *testing.T) {
	_, err := NewProvider("anthropic", WithBaseURL("https://custom.anthropic.com"), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewProvider with WithBaseURL should succeed: %v", err)
	}
}

func TestProviderOption_WithTimeout(t *testing.T) {
	_, err := NewProvider("anthropic", WithTimeout(60*time.Second), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewProvider with WithTimeout should succeed: %v", err)
	}
}

func TestProviderOption_WithMaxRetries(t *testing.T) {
	_, err := NewProvider("anthropic", WithMaxRetries(5), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewProvider with WithMaxRetries should succeed: %v", err)
	}
}

func TestProviderOption_WithTemperature(t *testing.T) {
	p, err := NewProvider("openai", WithTemperature(0.7), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewProvider with WithTemperature should succeed: %v", err)
	}
	openaiProvider, ok := p.(*OpenAIProvider)
	if !ok {
		t.Fatalf("expected *OpenAIProvider, got %T", p)
	}
	if openaiProvider.config.temperature != 0.7 {
		t.Fatalf("expected temperature 0.7, got %f", openaiProvider.config.temperature)
	}
}

func TestProviderOption_WithMaxContext(t *testing.T) {
	p, err := NewProvider("anthropic", WithMaxContext(8192), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewProvider with WithMaxContext should succeed: %v", err)
	}
	anthropicProvider, ok := p.(*AnthropicProvider)
	if !ok {
		t.Fatalf("expected *AnthropicProvider, got %T", p)
	}
	if anthropicProvider.config.maxContext != 8192 {
		t.Fatalf("expected maxContext 8192, got %d", anthropicProvider.config.maxContext)
	}
}

func TestProviderOption_Combined(t *testing.T) {
	p, err := NewProvider("openai",
		WithAPIKey("sk-xxx"),
		WithModel("gpt-4-turbo"),
		WithBaseURL("https://api.openai.com"),
		WithTimeout(120*time.Second),
		WithMaxRetries(10),
		WithTemperature(0.5),
		WithMaxContext(16384),
	)
	if err != nil {
		t.Fatalf("NewProvider with combined options should succeed: %v", err)
	}
	if p == nil {
		t.Fatal("Provider should not be nil")
	}
	openaiProvider, ok := p.(*OpenAIProvider)
	if !ok {
		t.Fatalf("expected *OpenAIProvider, got %T", p)
	}
	if openaiProvider.config.temperature != 0.5 {
		t.Fatalf("expected temperature 0.5, got %f", openaiProvider.config.temperature)
	}
	if openaiProvider.config.maxContext != 16384 {
		t.Fatalf("expected maxContext 16384, got %d", openaiProvider.config.maxContext)
	}
}

func TestNewProvider_MissingAPIKey(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "anthropic"},
		{name: "openai"},
		{name: "deepseek"},
		{name: "minimax"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewProvider(tc.name, WithModel("test-model"))
			if err == nil {
				t.Fatalf("NewProvider(%q) without API key should return error", tc.name)
			}
			if !strings.Contains(err.Error(), "API key missing") {
				t.Fatalf("expected 'API key missing' in error, got: %v", err)
			}
		})
	}
}
