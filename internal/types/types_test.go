package types

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestAgentError_Error(t *testing.T) {
	e := NewAgentError(ErrTaskTimeout, "timed out")
	expected := "[TASK_TIMEOUT] timed out"
	if e.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, e.Error())
	}
}

func TestAgentError_Error_WithCause(t *testing.T) {
	cause := errors.New("connection refused")
	e := NewAgentErrorWithCause(ErrTaskTimeout, "timed out", cause)
	expected := "[TASK_TIMEOUT] timed out: connection refused"
	if e.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, e.Error())
	}
}

func TestAgentError_Unwrap(t *testing.T) {
	cause := errors.New("inner")
	e := NewAgentErrorWithCause(ErrTaskRetryExhausted, "exhausted", cause)
	if !errors.Is(e, cause) {
		t.Error("Unwrap should return the cause error")
	}

	e2 := NewAgentError(ErrTaskNotFound, "not found")
	if e2.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause")
	}
}

func TestNewAgentError(t *testing.T) {
	e := NewAgentError(ErrContractNotFound, "contract test-1 not found")
	if e.Code != ErrContractNotFound {
		t.Errorf("Code = %s, want %s", e.Code, ErrContractNotFound)
	}
	if e.Message != "contract test-1 not found" {
		t.Errorf("Message = %s", e.Message)
	}
	if e.Cause != nil {
		t.Error("Cause should be nil")
	}
}

func TestNewAgentErrorWithCause(t *testing.T) {
	cause := fmt.Errorf("schema error")
	e := NewAgentErrorWithCause(ErrContractInputInvalid, "input invalid", cause)
	if e.Code != ErrContractInputInvalid {
		t.Errorf("Code = %s", e.Code)
	}
	if e.Cause != cause {
		t.Error("Cause not set correctly")
	}
}

func TestErrorCodeConstants(t *testing.T) {
	codes := map[ErrorCode]string{
		ErrSchedulerNotRunning:  "SCHEDULER_NOT_RUNNING",
		ErrTaskNotFound:         "TASK_NOT_FOUND",
		ErrTaskAlreadyExists:    "TASK_ALREADY_EXISTS",
		ErrDependencyCycle:      "DEPENDENCY_CYCLE",
		ErrDependencyNotReady:   "DEPENDENCY_NOT_READY",
		ErrContractNotFound:     "CONTRACT_NOT_FOUND",
		ErrContractInputInvalid: "CONTRACT_INPUT_INVALID",
		ErrTaskTimeout:          "TASK_TIMEOUT",
		ErrTaskRetryExhausted:   "TASK_RETRY_EXHAUSTED",
	}
	for code, expected := range codes {
		if string(code) != expected {
			t.Errorf("Code %s has wrong string value: %s", expected, string(code))
		}
	}
}

func TestSLAKeyConstants(t *testing.T) {
	if SLAKeyTimeoutMs != "sla.timeout_ms" {
		t.Errorf("SLAKeyTimeoutMs = %s", SLAKeyTimeoutMs)
	}
	if SLAKeyMaxRetries != "sla.max_retries" {
		t.Errorf("SLAKeyMaxRetries = %s", SLAKeyMaxRetries)
	}
	if SLAKeyFailureClass != "sla.failure_class" {
		t.Errorf("SLAKeyFailureClass = %s", SLAKeyFailureClass)
	}
}

func TestTaskStatusConstants(t *testing.T) {
	if string(TaskStatusPending) != "pending" {
		t.Error("TaskStatusPending wrong")
	}
	if string(TaskStatusRunning) != "running" {
		t.Error("TaskStatusRunning wrong")
	}
	if string(TaskStatusCompleted) != "completed" {
		t.Error("TaskStatusCompleted wrong")
	}
	if string(TaskStatusFailed) != "failed" {
		t.Error("TaskStatusFailed wrong")
	}
}

func TestFieldTypeConstants(t *testing.T) {
	types := map[FieldType]string{
		FieldTypeString: "string",
		FieldTypeInt:    "int",
		FieldTypeFloat:  "float",
		FieldTypeBool:   "bool",
		FieldTypeArray:  "array",
		FieldTypeObject: "object",
	}
	for ft, expected := range types {
		if string(ft) != expected {
			t.Errorf("FieldType %s = %s", expected, string(ft))
		}
	}
}

func TestAgentTask_Fields(t *testing.T) {
	now := time.Now()
	task := &AgentTask{
		TaskID:       "t1",
		ContractID:   "c1",
		Input:        map[string]any{"k": "v"},
		Dependencies: []string{"d1"},
		Status:       TaskStatusPending,
		CreatedAt:    now,
		Metadata:     map[string]string{SLAKeyTimeoutMs: "5000"},
	}
	if task.TaskID != "t1" || task.ContractID != "c1" {
		t.Error("Task fields not set correctly")
	}
	if task.Metadata[SLAKeyTimeoutMs] != "5000" {
		t.Error("Metadata not set correctly")
	}
}

func TestTaskResult_Fields(t *testing.T) {
	now := time.Now()
	result := &TaskResult{
		TaskID:    "t1",
		Output:    map[string]any{"result": "ok"},
		Error:     "",
		Status:    TaskStatusCompleted,
		Completed: now,
	}
	if result.Status != TaskStatusCompleted {
		t.Error("Result status wrong")
	}
}

func TestAgentContract_Fields(t *testing.T) {
	contract := &AgentContract{
		ContractID: "c1",
		InputSchema: &InputSchema{
			Fields: []FieldDef{{Name: "msg", Type: FieldTypeString, Required: true}},
		},
	}
	if contract.ContractID != "c1" || len(contract.InputSchema.Fields) != 1 {
		t.Error("Contract fields not set correctly")
	}
}
