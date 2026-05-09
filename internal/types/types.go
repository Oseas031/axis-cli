// Package types provides core data types for the agent system.
package types

import (
	"fmt"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// AgentTask represents an agent task to be scheduled
type AgentTask struct {
	TaskID       string            `json:"task_id"`
	ContractID   string            `json:"contract_id"`
	Input        map[string]any    `json:"input"`
	Dependencies []string          `json:"dependencies"`
	Status       TaskStatus        `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID    string         `json:"task_id"`
	Output    map[string]any `json:"output"`
	Error     string         `json:"error,omitempty"`
	Status    TaskStatus     `json:"status"`
	Completed time.Time      `json:"completed"`
}

// TaskState represents the stored state of a task
type TaskState struct {
	Task      *AgentTask  `json:"task"`
	Result    *TaskResult `json:"result,omitempty"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// FieldType represents the type of a field
type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeInt    FieldType = "int"
	FieldTypeFloat  FieldType = "float"
	FieldTypeBool   FieldType = "bool"
	FieldTypeArray  FieldType = "array"
	FieldTypeObject FieldType = "object"
)

// FieldDef defines a field in a schema
type FieldDef struct {
	Name        string    `json:"name"`
	Type        FieldType `json:"type"`
	Required    bool      `json:"required"`
	Description string    `json:"description,omitempty"`
	Enum        []string  `json:"enum,omitempty"`
}

// InputSchema defines the input contract
type InputSchema struct {
	Fields []FieldDef `json:"fields"`
}

// OutputSchema defines the output contract
type OutputSchema struct {
	Fields []FieldDef `json:"fields"`
}

// AgentContract defines an agent's input/output contract
type AgentContract struct {
	ContractID   string        `json:"contract_id"`
	InputSchema  *InputSchema  `json:"input_schema"`
	OutputSchema *OutputSchema `json:"output_schema"`
}

// SLA metadata keys stored in AgentTask.Metadata
const (
	SLAKeyTimeoutMs    = "sla.timeout_ms"
	SLAKeyMaxRetries   = "sla.max_retries"
	SLAKeyFailureClass = "sla.failure_class"
)

// TaskMetadataKeyExecutor selects the executor type for dispatch.
// Values: "model" (default) or "human".
const TaskMetadataKeyExecutor = "executor"

const (
	ExecutorTypeModel = "model"
	ExecutorTypeHuman = "human"
)

// ErrorCode is a stable machine-readable error identifier.
type ErrorCode string

const (
	ErrSchedulerNotRunning  ErrorCode = "SCHEDULER_NOT_RUNNING"
	ErrTaskNotFound         ErrorCode = "TASK_NOT_FOUND"
	ErrTaskAlreadyExists    ErrorCode = "TASK_ALREADY_EXISTS"
	ErrDependencyCycle      ErrorCode = "DEPENDENCY_CYCLE"
	ErrDependencyNotReady   ErrorCode = "DEPENDENCY_NOT_READY"
	ErrContractNotFound     ErrorCode = "CONTRACT_NOT_FOUND"
	ErrContractInputInvalid ErrorCode = "CONTRACT_INPUT_INVALID"
	ErrTaskTimeout          ErrorCode = "TASK_TIMEOUT"
	ErrTaskRetryExhausted   ErrorCode = "TASK_RETRY_EXHAUSTED"
)

// AgentError is a structured error with a stable error code.
type AgentError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *AgentError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AgentError) Unwrap() error {
	return e.Cause
}

func NewAgentError(code ErrorCode, message string) *AgentError {
	return &AgentError{Code: code, Message: message}
}

func NewAgentErrorWithCause(code ErrorCode, message string, cause error) *AgentError {
	return &AgentError{Code: code, Message: message, Cause: cause}
}

// ExecutionResult represents the result of a contract execution
type ExecutionResult struct {
	Output map[string]any `json:"output"`
	Error  string         `json:"error,omitempty"`
}

// HumanCallRequest represents a request to call a human
type HumanCallRequest struct {
	CallID   string            `json:"call_id"`
	TaskID   string            `json:"task_id"`
	Input    map[string]any    `json:"input"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// HumanCallResult represents the result of a human call
type HumanCallResult struct {
	CallID   string         `json:"call_id"`
	Output   map[string]any `json:"output"`
	Error    string         `json:"error,omitempty"`
	Status   CallStatus     `json:"status"`
	Resolved time.Time      `json:"resolved"`
}

// CallStatus represents the status of a human call
type CallStatus string

const (
	CallStatusPending   CallStatus = "pending"
	CallStatusCompleted CallStatus = "completed"
	CallStatusFailed    CallStatus = "failed"
)
