package provider

import (
	"context"
	"testing"
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

func TestNewProvider_Unknown(t *testing.T) {
	p, err := NewProvider("unknown")
	if err == nil {
		t.Fatal("NewProvider(\"unknown\") should return error")
	}
	if p != nil {
		t.Error("NewProvider(\"unknown\") should return nil provider")
	}
}
