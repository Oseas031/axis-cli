package executor

import (
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/types"
)

// ContractExecutor interface defines contract execution with validation
type ContractExecutor interface {
	Execute(contractID string, input map[string]any) (*types.ExecutionResult, error)
	ValidateInput(contractID string, input map[string]any) error
	ValidateOutput(contractID string, output map[string]any) error
	RegisterContract(contract *types.AgentContract) error
}

// ContractExecutorImpl implements contract execution
type ContractExecutorImpl struct {
	mu        sync.RWMutex
	contracts map[string]*types.AgentContract
}

// NewContractExecutor creates a new contract executor
func NewContractExecutor() *ContractExecutorImpl {
	return &ContractExecutorImpl{
		contracts: make(map[string]*types.AgentContract),
	}
}

// RegisterContract registers a contract
func (e *ContractExecutorImpl) RegisterContract(contract *types.AgentContract) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.contracts[contract.ContractID] = contract
	return nil
}

// Execute executes a contract with input validation
func (e *ContractExecutorImpl) Execute(contractID string, input map[string]any) (*types.ExecutionResult, error) {
	if err := e.ValidateInput(contractID, input); err != nil {
		return &types.ExecutionResult{
			Error: fmt.Sprintf("input validation failed: %v", err),
		}, err
	}

	// In milestone 1, we just validate and return a placeholder result
	// Actual execution will be handled by the dispatcher
	return &types.ExecutionResult{
		Output: map[string]any{"status": "validated"},
	}, nil
}

// ValidateInput validates input against the contract schema
func (e *ContractExecutorImpl) ValidateInput(contractID string, input map[string]any) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	contract, exists := e.contracts[contractID]
	if !exists {
		return fmt.Errorf("contract %s not found", contractID)
	}

	if contract.InputSchema == nil {
		return nil
	}

	for _, field := range contract.InputSchema.Fields {
		value, exists := input[field.Name]
		if field.Required && !exists {
			return fmt.Errorf("required field %s is missing", field.Name)
		}

		if exists {
			if err := e.validateFieldType(field.Name, value, field.Type); err != nil {
				return err
			}

			if len(field.Enum) > 0 {
				if field.Type != types.FieldTypeString {
					return fmt.Errorf("field %s has enum values but type is not string", field.Name)
				}
				if strValue, ok := value.(string); ok {
					if !e.isEnumValid(strValue, field.Enum) {
						return fmt.Errorf("field %s value %s is not in allowed enum values %v", field.Name, strValue, field.Enum)
					}
				}
			}
		}
	}

	return nil
}

// ValidateOutput validates output against the contract schema
func (e *ContractExecutorImpl) ValidateOutput(contractID string, output map[string]any) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	contract, exists := e.contracts[contractID]
	if !exists {
		return fmt.Errorf("contract %s not found", contractID)
	}

	if contract.OutputSchema == nil {
		return nil
	}

	for _, field := range contract.OutputSchema.Fields {
		value, exists := output[field.Name]
		if field.Required && !exists {
			return fmt.Errorf("required field %s is missing", field.Name)
		}

		if exists {
			if err := e.validateFieldType(field.Name, value, field.Type); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateFieldType validates a field's type
func (e *ContractExecutorImpl) validateFieldType(fieldName string, value any, expectedType types.FieldType) error {
	switch expectedType {
	case types.FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s must be string, got %T", fieldName, value)
		}
	case types.FieldTypeInt:
		if _, ok := value.(int); !ok {
			if _, ok := value.(float64); ok {
				// JSON numbers are float64 by default
				return nil
			}
			return fmt.Errorf("field %s must be int, got %T", fieldName, value)
		}
	case types.FieldTypeFloat:
		if _, ok := value.(float64); !ok {
			if _, ok := value.(int); ok {
				// Accept int as float
				return nil
			}
			return fmt.Errorf("field %s must be float, got %T", fieldName, value)
		}
	case types.FieldTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s must be bool, got %T", fieldName, value)
		}
	case types.FieldTypeArray:
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("field %s must be array, got %T", fieldName, value)
		}
	case types.FieldTypeObject:
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("field %s must be object, got %T", fieldName, value)
		}
	}
	return nil
}

// isEnumValid checks if a value is in the enum list
func (e *ContractExecutorImpl) isEnumValid(value string, enum []string) bool {
	for _, e := range enum {
		if e == value {
			return true
		}
	}
	return false
}
