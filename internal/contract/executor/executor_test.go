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

func TestContractExecutor_ValidateEnum_IntType(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "int-enum",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "priority", Type: types.FieldTypeInt, Enum: []string{"1", "2", "3"}},
			},
		},
	}
	exec.RegisterContract(contract)

	// Valid int enum
	if err := exec.ValidateInput("int-enum", map[string]any{"priority": 2}); err != nil {
		t.Errorf("Valid int enum should pass: %v", err)
	}

	// Valid via JSON float64 (no fractional part)
	if err := exec.ValidateInput("int-enum", map[string]any{"priority": float64(3)}); err != nil {
		t.Errorf("Valid int enum via float64 should pass: %v", err)
	}

	// Int not in enum
	if err := exec.ValidateInput("int-enum", map[string]any{"priority": 99}); err == nil {
		t.Error("Int not in enum should fail")
	}
}

func TestContractExecutor_ValidateEnum_FloatPrecisionLoss(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "int-enum",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "priority", Type: types.FieldTypeInt, Enum: []string{"1", "2"}},
			},
		},
	}
	exec.RegisterContract(contract)

	// Float64 with fractional part
	err := exec.ValidateInput("int-enum", map[string]any{"priority": 1.5})
	if err == nil {
		t.Error("Float with fractional part should fail int enum validation")
	}
}

func TestContractExecutor_ValidateEnum_UnsupportedType(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "bool-enum",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "flag", Type: types.FieldTypeBool, Enum: []string{"true", "false"}},
			},
		},
	}
	exec.RegisterContract(contract)

	err := exec.ValidateInput("bool-enum", map[string]any{"flag": true})
	if err == nil {
		t.Error("Enum on unsupported type should fail")
	}
}

func TestContractExecutor_ValidateFieldType_ArrayAndObject(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "arr-obj",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "items", Type: types.FieldTypeArray},
				{Name: "meta", Type: types.FieldTypeObject},
			},
		},
	}
	exec.RegisterContract(contract)

	tests := []struct {
		name  string
		input map[string]any
		valid bool
	}{
		{"valid array and object", map[string]any{"items": []any{1, 2}, "meta": map[string]any{"k": "v"}}, true},
		{"invalid array", map[string]any{"items": "not-array", "meta": map[string]any{}}, false},
		{"invalid object", map[string]any{"items": []any{}, "meta": "not-object"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exec.ValidateInput("arr-obj", tt.input)
			if tt.valid && err != nil {
				t.Errorf("Expected valid: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected invalid")
			}
		})
	}
}

func TestContractExecutor_Execute_InvalidInput(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "name", Type: types.FieldTypeString, Required: true},
			},
		},
	}
	exec.RegisterContract(contract)

	result, err := exec.Execute("test", map[string]any{})
	if err == nil {
		t.Error("Execute with invalid input should return error")
	}
	if result != nil && result.Error == "" {
		t.Error("Execute with invalid input should set error in result")
	}
}

func TestContractExecutor_ValidateOutput_WrongType(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test",
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{Name: "result", Type: types.FieldTypeString, Required: true},
			},
		},
	}
	exec.RegisterContract(contract)

	err := exec.ValidateOutput("test", map[string]any{"result": 123})
	if err == nil {
		t.Error("Wrong type in output should fail validation")
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

func TestContractExecutor_ValidateEnum_IntOutOfRange(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "int-enum",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "priority", Type: types.FieldTypeInt, Enum: []string{"1", "2"}},
			},
		},
	}
	exec.RegisterContract(contract)

	// Float64 value out of int range
	err := exec.ValidateInput("int-enum", map[string]any{"priority": 1e20})
	if err == nil {
		t.Error("Float64 out of int range should fail int enum validation")
	}
}

func TestContractExecutor_ValidateEnum_IntFloat64NotInEnum(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "int-enum",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "priority", Type: types.FieldTypeInt, Enum: []string{"1", "2"}},
			},
		},
	}
	exec.RegisterContract(contract)

	// Float64 value in range but not in enum
	err := exec.ValidateInput("int-enum", map[string]any{"priority": float64(99)})
	if err == nil {
		t.Error("Float64 in range but not in enum should fail validation")
	}
}

func TestContractExecutor_RegisterContract_Duplicate(t *testing.T) {
	exec := NewContractExecutor()

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "name", Type: types.FieldTypeString, Required: true},
			},
		},
	}

	err := exec.RegisterContract(contract)
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	// Duplicate registration should fail
	err = exec.RegisterContract(contract)
	if err == nil {
		t.Error("Duplicate registration should fail")
	}
}

func TestContractExecutor_Execute_NonExistentContract(t *testing.T) {
	exec := NewContractExecutor()

	// Execute with a contract ID that does not exist
	result, err := exec.Execute("non-existent", map[string]any{"name": "test"})
	if err == nil {
		t.Error("Execute with non-existent contract should return error")
	}
	if result == nil {
		t.Error("Result should not be nil even on error")
	}
	if result.Error == "" {
		t.Error("Result.Error should contain error message on contract not found")
	}
}
