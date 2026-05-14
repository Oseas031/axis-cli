// Package executor provides contract execution with input/output validation.
package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// ContractExecutor interface defines contract execution with validation
type ContractExecutor interface {
	Execute(ctx context.Context, contractID string, input map[string]any) (*types.ExecutionResult, error)
	ValidateInput(contractID string, input map[string]any) error
	ValidateOutput(contractID string, output map[string]any) error
	RegisterContract(contract *types.AgentContract) error
}

// ContractExecutorImpl implements contract execution
type ContractExecutorImpl struct {
	mu                      sync.RWMutex
	contracts               map[string]*types.AgentContract
	provider                provider.ModelProvider
	toolRegistry            *tool.Registry
	skillsLoader            interface{ BuildSkillsPromptSection(context.Context) string }
	principlesLoader        interface{ BuildPrinciplesPromptSection() string }
	compactionPipeline      Compactor
	allowedScopes           []string
	circuitBreakerThreshold int
	maxTurns                int
	executionStarted        bool
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

// Deprecated: use NewContractExecutorWithConfig.
func (e *ContractExecutorImpl) SetProvider(p provider.ModelProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.executionStarted {
		panic("SetProvider called after execution started")
	}
	e.provider = p
}

// Deprecated: use NewContractExecutorWithConfig.
func (e *ContractExecutorImpl) SetToolRegistry(tr *tool.Registry) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.executionStarted {
		panic("SetToolRegistry called after execution started")
	}
	e.toolRegistry = tr
}

// Deprecated: use NewContractExecutorWithConfig.
func (e *ContractExecutorImpl) SetSkillsLoader(sl interface{ BuildSkillsPromptSection(context.Context) string }) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.executionStarted {
		panic("SetSkillsLoader called after execution started")
	}
	e.skillsLoader = sl
}

// Deprecated: use NewContractExecutorWithConfig.
func (e *ContractExecutorImpl) SetPrinciplesLoader(pl interface{ BuildPrinciplesPromptSection() string }) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.executionStarted {
		panic("SetPrinciplesLoader called after execution started")
	}
	e.principlesLoader = pl
}

// Deprecated: use NewContractExecutorWithConfig.
func (e *ContractExecutorImpl) SetCompactionPipeline(p Compactor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.executionStarted {
		panic("SetCompactionPipeline called after execution started")
	}
	e.compactionPipeline = p
}

// safeMarshal JSON-marshals a value with panic recovery.
// Returns an error if marshaling fails or if a panic occurs.
func safeMarshal(v any) ([]byte, error) {
	var result []byte
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic during JSON marshal: %v", r)
				result = nil
			}
		}()
		result, err = json.Marshal(v)
	}()
	return result, err
}

// Execute executes a contract: validates input, runs the provider (with optional
// multi-turn tool loop), and validates output.
func (e *ContractExecutorImpl) Execute(ctx context.Context, contractID string, input map[string]any) (*types.ExecutionResult, error) {
	e.mu.Lock()
	e.executionStarted = true
	e.mu.Unlock()

	if err := e.ValidateInput(contractID, input); err != nil {
		return &types.ExecutionResult{
			Error: fmt.Sprintf("input validation failed: %v", err),
		}, err
	}

	e.mu.RLock()
	p := e.provider
	tr := e.toolRegistry
	e.mu.RUnlock()

	if p != nil {
		req := &provider.ModelRequest{ContractID: contractID, Input: input}

		// Inject skills prompt if available
		if e.skillsLoader != nil {
			req.SystemPrompt = e.skillsLoader.BuildSkillsPromptSection(ctx)
		}
		// Inject derived principles (zero retrieval cost, always present)
		if e.principlesLoader != nil {
			req.SystemPrompt += e.principlesLoader.BuildPrinciplesPromptSection()
		}

		// Add tools if a registry is available with registered tools.
		hasTools := false
		if tr != nil {
			tools := tr.List()
			if len(tools) > 0 {
				req.Tools = tools
				hasTools = true
			}
		}

		// Multi-turn loop for tool-based execution.
		if hasTools {
			return e.executeMultiTurn(ctx, p, tr, req, contractID)
		}

		// Single-pass execution (backward compatible path).
		resp, err := p.Execute(ctx, req)
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

// executeMultiTurn runs a multi-turn tool-calling loop with a maximum of 10 turns.
// It implements a circuit breaker that aborts after 5 consecutive tool errors.
func (e *ContractExecutorImpl) executeMultiTurn(ctx context.Context, p provider.ModelProvider, tr *tool.Registry, req *provider.ModelRequest, contractID string) (*types.ExecutionResult, error) {
	var history []types.ModelMessage
	turnsSinceProgress := 0
	maxTurns := e.maxTurns
	if maxTurns == 0 {
		maxTurns = 10
	}
	consecutiveErrors := 0
	circuitBreakerThreshold := e.circuitBreakerThreshold
	if circuitBreakerThreshold == 0 {
		circuitBreakerThreshold = 5
	}

	var lastOutput map[string]any
	for turn := 0; turn < maxTurns; turn++ {
		req.History = history
		if turnsSinceProgress >= 5 {
			req.SystemPrompt += "\nReminder: you have not updated task progress in the last 5 turns. Consider checkpointing or recording progress."
			turnsSinceProgress = 0
		}
		if turn == maxTurns-1 {
			history = append(history, types.ModelMessage{
				Role:    "system",
				Content: "This is your final turn. You must produce a final answer now.",
			})
			req.History = history
		}
		resp, err := p.Execute(ctx, req)
		if err != nil {
			return &types.ExecutionResult{
				Error: fmt.Sprintf("provider execution failed: %v", err),
			}, err
		}

		if len(resp.ToolCalls) > 0 {
			assistantMsg := types.ModelMessage{
				Role:      "assistant",
				ToolCalls: resp.ToolCalls,
			}
			history = append(history, assistantMsg)

			for _, tc := range resp.ToolCalls {
				toolImpl, ok := tr.Get(tc.Name)
				if !ok {
					history = append(history, types.ModelMessage{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf("error: tool %s not found", tc.Name),
					})
					consecutiveErrors++
					if consecutiveErrors >= circuitBreakerThreshold {
						return &types.ExecutionResult{
							Error: fmt.Sprintf("circuit breaker triggered: %d consecutive tool errors, aborting", consecutiveErrors),
						}, fmt.Errorf("circuit breaker triggered: %d consecutive tool errors", consecutiveErrors)
					}
					continue
				}
				// Check permission scopes
				if !e.isScopeAllowed(tr, tc.Name) {
					history = append(history, types.ModelMessage{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf("error: tool %s requires scope %v which is not allowed", tc.Name, tr.GetScopes(tc.Name)),
					})
					consecutiveErrors++
					if consecutiveErrors >= circuitBreakerThreshold {
						return &types.ExecutionResult{
							Error: fmt.Sprintf("circuit breaker triggered: %d consecutive tool errors, aborting", consecutiveErrors),
						}, fmt.Errorf("circuit breaker triggered: %d consecutive tool errors", consecutiveErrors)
					}
					continue
				}
				result, execErr := toolImpl.Execute(ctx, tc.Input)
				if execErr != nil {
					history = append(history, types.ModelMessage{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf("error: %v", execErr),
					})
					consecutiveErrors++
					if consecutiveErrors >= circuitBreakerThreshold {
						return &types.ExecutionResult{
							Error: fmt.Sprintf("circuit breaker triggered: %d consecutive tool errors, aborting", consecutiveErrors),
						}, fmt.Errorf("circuit breaker triggered: %d consecutive tool errors", consecutiveErrors)
					}
					continue
				}
				// Successful execution resets the error counter
				consecutiveErrors = 0
				content, marshalErr := safeMarshal(result)
				if marshalErr != nil {
					history = append(history, types.ModelMessage{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf("error: failed to marshal tool result: %v", marshalErr),
					})
				} else {
					history = append(history, types.ModelMessage{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    string(content),
					})
				}
			}
			// Track progress tool usage
			progressMade := false
			for _, tc := range resp.ToolCalls {
				if tc.Name == "checkpoint" || tc.Name == "store_memory" {
					progressMade = true
					break
				}
			}
			if progressMade {
				turnsSinceProgress = 0
			} else {
				turnsSinceProgress++
			}
			// Compact history if pipeline is configured
			if e.compactionPipeline != nil {
				history = e.compactionPipeline.Compact(ctx, history)
			}
			continue
		}

		if resp.Output != nil {
			if err := e.ValidateOutput(contractID, resp.Output); err != nil {
				return &types.ExecutionResult{
					Error: fmt.Sprintf("output validation failed: %v", err),
				}, err
			}
			return &types.ExecutionResult{Output: resp.Output}, nil
		}
		lastOutput = resp.Output
	}

	return &types.ExecutionResult{
		Output: lastOutput,
		Error:  fmt.Sprintf("execution terminated: maximum turns (%d) reached without completion", maxTurns),
	}, fmt.Errorf("execution terminated: maximum turns (%d) reached without completion", maxTurns)
}

// isScopeAllowed checks whether all scopes required by a tool are in the executor's allowed list.
// If allowedScopes is empty, all tools are permitted (backward compatible).
func (e *ContractExecutorImpl) isScopeAllowed(tr *tool.Registry, toolName string) bool {
	if len(e.allowedScopes) == 0 {
		return true
	}
	for _, required := range tr.GetScopes(toolName) {
		found := false
		for _, allowed := range e.allowedScopes {
			if allowed == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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
