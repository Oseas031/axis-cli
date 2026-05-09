package provider

import (
	"context"
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
	p, err := NewProvider("anthropic", WithAPIKey("test-key"), WithModel("claude-sonnet-4-5"))
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
	p, err := NewProvider("openai", WithAPIKey("test-key"), WithModel("gpt-4"))
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
	p, err := NewProvider("anthropic", WithModel("claude-3-opus"))
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
	_, err := NewProvider("anthropic", WithBaseURL("https://custom.anthropic.com"))
	if err != nil {
		t.Fatalf("NewProvider with WithBaseURL should succeed: %v", err)
	}
}

func TestProviderOption_WithTimeout(t *testing.T) {
	_, err := NewProvider("anthropic", WithTimeout(60*time.Second))
	if err != nil {
		t.Fatalf("NewProvider with WithTimeout should succeed: %v", err)
	}
}

func TestProviderOption_WithMaxRetries(t *testing.T) {
	_, err := NewProvider("anthropic", WithMaxRetries(5))
	if err != nil {
		t.Fatalf("NewProvider with WithMaxRetries should succeed: %v", err)
	}
}

func TestProviderOption_Combined(t *testing.T) {
	p, err := NewProvider("openai",
		WithAPIKey("sk-xxx"),
		WithModel("gpt-4-turbo"),
		WithBaseURL("https://api.openai.com"),
		WithTimeout(120*time.Second),
		WithMaxRetries(10),
	)
	if err != nil {
		t.Fatalf("NewProvider with combined options should succeed: %v", err)
	}
	if p == nil {
		t.Fatal("Provider should not be nil")
	}
}
