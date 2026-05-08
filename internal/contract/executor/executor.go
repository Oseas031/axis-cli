// Package executor provides contract execution with input/output validation.
package executor

import (
	"context"
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/model/provider"
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
	provider  provider.ModelProvider
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

	// Check if contract already exists
	if _, exists := e.contracts[contract.ContractID]; exists {
		return fmt.Errorf("contract %s already exists", contract.ContractID)
	}

	e.contracts[contract.ContractID] = contract
	return nil
}

// SetProvider sets the model provider for execution.
func (e *ContractExecutorImpl) SetProvider(p provider.ModelProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.provider = p
}

// Execute executes a contract: validates input, runs the provider, validates output.
func (e *ContractExecutorImpl) Execute(contractID string, input map[string]any) (*types.ExecutionResult, error) {
	if err := e.ValidateInput(contractID, input); err != nil {
		return &types.ExecutionResult{
			Error: fmt.Sprintf("input validation failed: %v", err),
		}, err
	}

	e.mu.RLock()
	p := e.provider
	e.mu.RUnlock()

	if p != nil {
		req := &provider.ModelRequest{ContractID: contractID, Input: input}
		resp, err := p.Execute(context.Background(), req)
		if err != nil {
			return &types.ExecutionResult{
				Error: fmt.Sprintf("provider execution failed: %v", err),
			}, err
		}
		if err := e.ValidateOutput(contractID, resp.Output); err != nil {
			return &types.ExecutionResult{
				Error: fmt.Sprintf("output validation failed: %v", err),
			}, err
		}
		return &types.ExecutionResult{
			Output: resp.Output,
		}, nil
	}

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
		return types.NewAgentError(types.ErrContractNotFound, fmt.Sprintf("contract %s not found", contractID))
	}

	if contract.InputSchema == nil {
		return nil
	}

	for _, field := range contract.InputSchema.Fields {
		value, exists := input[field.Name]
		if field.Required && !exists {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("required field %s is missing", field.Name))
		}

		if exists {
			if err := e.validateFieldType(field.Name, value, field.Type); err != nil {
				return types.NewAgentErrorWithCause(types.ErrContractInputInvalid, "input validation failed", err)
			}

			if len(field.Enum) > 0 {
				if err := e.validateEnum(field.Name, value, field.Type, field.Enum); err != nil {
					return types.NewAgentErrorWithCause(types.ErrContractInputInvalid, "input validation failed", err)
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
		return types.NewAgentError(types.ErrContractNotFound, fmt.Sprintf("contract %s not found", contractID))
	}

	if contract.OutputSchema == nil {
		return nil
	}

	for _, field := range contract.OutputSchema.Fields {
		value, exists := output[field.Name]
		if field.Required && !exists {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("required field %s is missing", field.Name))
		}

		if exists {
			if err := e.validateFieldType(field.Name, value, field.Type); err != nil {
				return types.NewAgentErrorWithCause(types.ErrContractInputInvalid, "input validation failed", err)
			}

			if len(field.Enum) > 0 {
				if err := e.validateEnum(field.Name, value, field.Type, field.Enum); err != nil {
					return types.NewAgentErrorWithCause(types.ErrContractInputInvalid, "input validation failed", err)
				}
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

// validateEnum checks if a value is in the enum list
func (e *ContractExecutorImpl) validateEnum(fieldName string, value any, fieldType types.FieldType, enum []string) error {
	switch fieldType {
	case types.FieldTypeString:
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("field %s enum validation expects string, got %T", fieldName, value)
		}
		if !e.isStringEnumValid(strValue, enum) {
			return fmt.Errorf("field %s value %s is not in allowed enum values %v", fieldName, strValue, enum)
		}
	case types.FieldTypeInt:
		intValue, ok := value.(int)
		if !ok {
			// JSON numbers are float64 by default
			if floatVal, ok := value.(float64); ok {
				// Check if the float value is within int range
				if floatVal < float64(-1<<63) || floatVal > float64(1<<63-1) {
					return fmt.Errorf("field %s value %f is out of int range", fieldName, floatVal)
				}
				intValue = int(floatVal)
				// Check for precision loss
				if float64(intValue) != floatVal {
					return fmt.Errorf("field %s value %f has fractional part, cannot convert to int", fieldName, floatVal)
				}
			} else {
				return fmt.Errorf("field %s enum validation expects int, got %T", fieldName, value)
			}
		}
		if !e.isIntEnumValid(intValue, enum) {
			return fmt.Errorf("field %s value %d is not in allowed enum values %v", fieldName, intValue, enum)
		}
	default:
		return fmt.Errorf("field %s has enum values but type %s is not supported for enum", fieldName, fieldType)
	}
	return nil
}

// isStringEnumValid checks if a string value is in the enum list
func (e *ContractExecutorImpl) isStringEnumValid(value string, enum []string) bool {
	for _, e := range enum {
		if e == value {
			return true
		}
	}
	return false
}

// isIntEnumValid checks if an int value is in the enum list (enum values are stored as strings)
func (e *ContractExecutorImpl) isIntEnumValid(value int, enum []string) bool {
	for _, e := range enum {
		if e == fmt.Sprintf("%d", value) {
			return true
		}
	}
	return false
}
