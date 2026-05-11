package intent

import (
	"context"
	"testing"
)

func TestDeterministicParser_ParseDefaultTask(t *testing.T) {
	result, err := NewDeterministicParser().Parse(context.Background(), Request{Prompt: "check provider config"})
	if err != nil {
		t.Fatalf("Parse should succeed: %v", err)
	}
	if result.Task == nil {
		t.Fatal("expected task")
	}
	if result.Task.ContractID != "default" {
		t.Fatalf("expected default contract, got %s", result.Task.ContractID)
	}
	if result.Task.Input["message"] != "check provider config" {
		t.Fatalf("expected message input to preserve prompt, got %#v", result.Task.Input["message"])
	}
	if result.Task.Input["goal"] != "check provider config" {
		t.Fatalf("expected goal input to preserve prompt, got %#v", result.Task.Input["goal"])
	}
	// Check new namespaced keys per metadata-key-conventions.md
	if result.Task.Metadata["intent.source"] != "natural_language" {
		t.Fatalf("expected intent.source metadata, got %#v", result.Task.Metadata)
	}
	if result.Task.Metadata["intent.parser_mode"] != ParserModeDeterministic {
		t.Fatalf("expected intent.parser_mode metadata, got %#v", result.Task.Metadata)
	}
	// Verify backward compatibility with legacy keys during transition
	if result.Task.Metadata["source"] != "natural_language" {
		t.Fatalf("expected legacy source metadata for compatibility, got %#v", result.Task.Metadata)
	}
	if result.Task.Metadata["parser_mode"] != ParserModeDeterministic {
		t.Fatalf("expected legacy parser_mode metadata for compatibility, got %#v", result.Task.Metadata)
	}
}

func TestDeterministicParser_ParseExplicitContractAndTaskID(t *testing.T) {
	result, err := NewDeterministicParser().Parse(context.Background(), Request{Prompt: "write tests", ContractID: "code", TaskID: "task-123"})
	if err != nil {
		t.Fatalf("Parse should succeed: %v", err)
	}
	if result.Task.TaskID != "task-123" {
		t.Fatalf("expected explicit task ID, got %s", result.Task.TaskID)
	}
	if result.Task.ContractID != "code" {
		t.Fatalf("expected explicit contract, got %s", result.Task.ContractID)
	}
}

func TestDeterministicParser_ParseEmptyPrompt(t *testing.T) {
	_, err := NewDeterministicParser().Parse(context.Background(), Request{Prompt: "   "})
	if err == nil {
		t.Fatal("expected empty prompt to fail")
	}
}
