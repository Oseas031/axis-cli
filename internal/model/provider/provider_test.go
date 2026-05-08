package provider

import (
	"context"
	"testing"
)

func TestMockModelProvider_Execute(t *testing.T) {
	m := NewMockModelProvider()
	req := &ModelRequest{
		ContractID: "test",
		Input:      map[string]any{"message": "hello"},
	}

	resp, err := m.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	if resp.Output["message"] != "hello" {
		t.Errorf("Expected echoed message, got %v", resp.Output["message"])
	}
	if resp.Output["contract_id"] != "test" {
		t.Errorf("Expected contract_id=test, got %v", resp.Output["contract_id"])
	}
	if resp.Output["provider"] != "mock" {
		t.Errorf("Expected provider=mock, got %v", resp.Output["provider"])
	}
}

func TestMockModelProvider_Execute_EmptyInput(t *testing.T) {
	m := NewMockModelProvider()
	req := &ModelRequest{
		ContractID: "empty",
		Input:      map[string]any{},
	}

	resp, err := m.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}
	if resp.Output["status"] != "completed" {
		t.Errorf("Expected status=completed, got %v", resp.Output["status"])
	}
}

func TestMockModelProvider_ImplementsInterface(t *testing.T) {
	var _ ModelProvider = NewMockModelProvider()
}
