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
