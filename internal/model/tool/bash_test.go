package tool

import (
	"context"
	"testing"
)

func TestBashTool_Name(t *testing.T) {
	b := NewBashTool()
	if b.Name() != "bash" {
		t.Errorf("Expected name 'bash', got %q", b.Name())
	}
}

func TestBashTool_Schema(t *testing.T) {
	b := NewBashTool()
	s := b.Schema()
	if s.Name != "bash" {
		t.Errorf("Expected schema name 'bash', got %q", s.Name)
	}
	if len(s.Parameters) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(s.Parameters))
	}
	if s.Parameters[0].Name != "command" {
		t.Errorf("Expected parameter 'command', got %q", s.Parameters[0].Name)
	}
	if !s.Parameters[0].Required {
		t.Error("Expected command parameter to be required")
	}
}

func TestBashTool_Execute_Echo(t *testing.T) {
	b := NewBashTool()
	ctx := context.Background()
	input := map[string]any{"command": "echo hello"}

	result, err := b.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}

	stdout, ok := result["stdout"].(string)
	if !ok {
		t.Fatal("Expected stdout to be a string")
	}
	if stdout != "hello\n" && stdout != "hello\r\n" {
		t.Errorf("Expected stdout 'hello\\n', got %q", stdout)
	}

	exitCode, ok := result["exit_code"].(int)
	if !ok {
		t.Fatal("Expected exit_code to be an int")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit_code 0, got %d", exitCode)
	}
}

func TestBashTool_Execute_ExitCode(t *testing.T) {
	b := NewBashTool()
	ctx := context.Background()
	input := map[string]any{"command": "exit 42"}

	result, err := b.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}

	exitCode, ok := result["exit_code"].(int)
	if !ok {
		t.Fatal("Expected exit_code to be an int")
	}
	if exitCode != 42 {
		t.Errorf("Expected exit_code 42, got %d", exitCode)
	}
}

func TestBashTool_Execute_MissingCommand(t *testing.T) {
	b := NewBashTool()
	ctx := context.Background()
	input := map[string]any{}

	result, err := b.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute with missing command should not error: %v", err)
	}

	if result["error"] != "command is required and must be a string" {
		t.Errorf("Expected error message about missing command, got %v", result["error"])
	}
}

func TestBashTool_Execute_EmptyCommand(t *testing.T) {
	b := NewBashTool()
	ctx := context.Background()
	input := map[string]any{"command": ""}

	result, err := b.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute with empty command should not error: %v", err)
	}

	if result["error"] == nil {
		t.Error("Expected error for empty command")
	}
}

func TestBashTool_ImplementsInterface(t *testing.T) {
	var _ Tool = NewBashTool()
}
