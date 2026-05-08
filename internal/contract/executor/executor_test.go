package executor

import (
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestContractExecutor_RegisterContract(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "name",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "result",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}

	err := exec.RegisterContract(contract)
	if err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}
}

func TestContractExecutor_ValidateInput(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "name",
					Type:     types.FieldTypeString,
					Required: true,
				},
				{
					Name: "age",
					Type: types.FieldTypeInt,
					Enum: []string{"25", "30", "35"},
				},
			},
		},
	}

	exec.RegisterContract(contract)

	// Valid input
	validInput := map[string]any{
		"name": "test",
		"age":  25,
	}
	err := exec.ValidateInput("test-contract", validInput)
	if err != nil {
		t.Errorf("Valid input should pass validation: %v", err)
	}

	// Missing required field
	invalidInput := map[string]any{
		"age": 25,
	}
	err = exec.ValidateInput("test-contract", invalidInput)
	if err == nil {
		t.Error("Missing required field should fail validation")
	}

	// Wrong type
	wrongTypeInput := map[string]any{
		"name": 123,
	}
	err = exec.ValidateInput("test-contract", wrongTypeInput)
	if err == nil {
		t.Error("Wrong type should fail validation")
	}
}

func TestContractExecutor_ValidateOutput(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "result",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}

	exec.RegisterContract(contract)

	// Valid output
	validOutput := map[string]any{
		"result": "success",
	}
	err := exec.ValidateOutput("test-contract", validOutput)
	if err != nil {
		t.Errorf("Valid output should pass validation: %v", err)
	}

	// Missing required field
	invalidOutput := map[string]any{}
	err = exec.ValidateOutput("test-contract", invalidOutput)
	if err == nil {
		t.Error("Missing required field should fail validation")
	}
}

func TestContractExecutor_Execute(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "name",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}

	exec.RegisterContract(contract)

	// Valid execution
	input := map[string]any{"name": "test"}
	result, err := exec.Execute("test-contract", input)
	if err != nil {
		t.Errorf("Execute should succeed: %v", err)
	}
	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestContractExecutor_ValidateInputAllTypes(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "str", Type: types.FieldTypeString},
				{Name: "int", Type: types.FieldTypeInt},
				{Name: "float", Type: types.FieldTypeFloat},
				{Name: "bool", Type: types.FieldTypeBool},
				{Name: "arr", Type: types.FieldTypeArray},
				{Name: "obj", Type: types.FieldTypeObject},
			},
		},
	}

	exec.RegisterContract(contract)

	validInput := map[string]any{
		"str":   "test",
		"int":   42,
		"float": 3.14,
		"bool":  true,
		"arr":   []any{1, 2, 3},
		"obj":   map[string]any{"key": "value"},
	}
	err := exec.ValidateInput("test-contract", validInput)
	if err != nil {
		t.Errorf("All valid types should pass: %v", err)
	}
}

func TestContractExecutor_ValidateNonExistentContract(t *testing.T) {
	exec := NewContractExecutor()

	input := map[string]any{"name": "test"}
	err := exec.ValidateInput("non-existent", input)
	if err == nil {
		t.Error("Non-existent contract should fail validation")
	}

	output := map[string]any{"result": "success"}
	err = exec.ValidateOutput("non-existent", output)
	if err == nil {
		t.Error("Non-existent contract should fail validation")
	}
}

func TestContractExecutor_ValidateNilSchema(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID:   "test-contract",
		InputSchema:  nil,
		OutputSchema: nil,
	}

	exec.RegisterContract(contract)

	// Should pass when schema is nil
	input := map[string]any{"name": "test"}
	err := exec.ValidateInput("test-contract", input)
	if err != nil {
		t.Errorf("Nil schema should pass validation: %v", err)
	}

	output := map[string]any{"result": "success"}
	err = exec.ValidateOutput("test-contract", output)
	if err != nil {
		t.Errorf("Nil schema should pass validation: %v", err)
	}
}

func TestContractExecutor_ValidateFieldType(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "str", Type: types.FieldTypeString},
				{Name: "int", Type: types.FieldTypeInt},
				{Name: "float", Type: types.FieldTypeFloat},
				{Name: "bool", Type: types.FieldTypeBool},
			},
		},
	}

	exec.RegisterContract(contract)

	tests := []struct {
		name  string
		field string
		value any
		valid bool
	}{
		{"valid string", "str", "test", true},
		{"invalid string", "str", 123, false},
		{"valid int", "int", 42, true},
		{"invalid int", "int", "test", false},
		{"valid float", "float", 3.14, true},
		{"invalid float", "float", "test", false},
		{"valid bool", "bool", true, true},
		{"invalid bool", "bool", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{tt.field: tt.value}
			err := exec.ValidateInput("test-contract", input)
			if tt.valid && err != nil {
				t.Errorf("Expected valid: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected invalid")
			}
		})
	}
}

func TestContractExecutor_ValidateEnum(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name: "status",
					Type: types.FieldTypeString,
					Enum: []string{"pending", "running", "completed"},
				},
			},
		},
	}

	exec.RegisterContract(contract)

	// Valid enum value
	validInput := map[string]any{"status": "pending"}
	err := exec.ValidateInput("test-contract", validInput)
	if err != nil {
		t.Errorf("Valid enum value should pass: %v", err)
	}

	// Invalid enum value
	invalidInput := map[string]any{"status": "invalid"}
	err = exec.ValidateInput("test-contract", invalidInput)
	if err == nil {
		t.Error("Invalid enum value should fail validation")
	}

	// Non-string value with enum (should not fail enum check but type check will fail)
	wrongTypeInput := map[string]any{"status": 123}
	err = exec.ValidateInput("test-contract", wrongTypeInput)
	if err == nil {
		t.Error("Wrong type should fail validation")
	}
}
