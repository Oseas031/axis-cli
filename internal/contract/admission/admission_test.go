package admission

import (
	"testing"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	"github.com/axis-cli/axis/internal/types"
)

func TestAdmissionValidator_Validate_ValidTask(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
	}

	if err := av.Validate(task); err != nil {
		t.Errorf("Valid task should pass admission: %v", err)
	}
}

func TestAdmissionValidator_Validate_EmptyTaskID(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	av := NewAdmissionValidator(ce)

	task := &types.AgentTask{
		TaskID:     "",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with empty TaskID should be rejected")
	}
}

func TestAdmissionValidator_Validate_EmptyContractID(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	av := NewAdmissionValidator(ce)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "",
		Input:      map[string]any{"message": "hello"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with empty ContractID should be rejected")
	}
}

func TestAdmissionValidator_Validate_UnknownContract(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	av := NewAdmissionValidator(ce)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "nonexistent",
		Input:      map[string]any{"message": "hello"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with unknown contract should be rejected")
	}
}

func TestAdmissionValidator_Validate_InvalidInput(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{}, // missing required "message" field
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with invalid input should be rejected")
	}
}

func TestAdmissionValidator_Validate_OptionalFieldMissing(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	contract := &types.AgentContract{
		ContractID: "optional-fields",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "required_field", Type: types.FieldTypeString, Required: true},
				{Name: "optional_field", Type: types.FieldTypeString, Required: false},
			},
		},
	}
	if err := ce.RegisterContract(contract); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "optional-fields",
		Input:      map[string]any{"required_field": "value"},
	}

	if err := av.Validate(task); err != nil {
		t.Errorf("Task missing optional field should still pass: %v", err)
	}
}

func TestAdmissionValidator_Validate_ValidSLA(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyTimeoutMs: "5000", types.SLAKeyMaxRetries: "2"},
	}

	if err := av.Validate(task); err != nil {
		t.Errorf("Task with valid SLA metadata should pass: %v", err)
	}
}

func TestAdmissionValidator_Validate_SLA_TimeoutNegative(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyTimeoutMs: "-100"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with negative sla.timeout_ms should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_TimeoutZero(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyTimeoutMs: "0"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with zero sla.timeout_ms should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_TimeoutNotInteger(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyTimeoutMs: "abc"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with non-integer sla.timeout_ms should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_MaxRetriesNegative(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyMaxRetries: "-1"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with negative sla.max_retries should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_MaxRetriesExceedsLimit(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyMaxRetries: "10"},
	}

	if err := av.Validate(task); err == nil {
		t.Errorf("Task with sla.max_retries > MaxRetryLimit (%d) should be rejected", types.MaxRetryLimit)
	}
}

func TestAdmissionValidator_Validate_SLA_FatalWithRetries(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyFailureClass: "fatal", types.SLAKeyMaxRetries: "2"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with failure_class=fatal and max_retries>0 should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_NoMetadata(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
	}

	if err := av.Validate(task); err != nil {
		t.Errorf("Task without SLA metadata should pass: %v", err)
	}
}

func TestAdmissionValidator_Validate_SLA_ValidFailureClass(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)

	// All three valid failure classes should pass
	for _, fc := range []string{types.FailureClassRetryable, types.FailureClassFatal, types.FailureClassDegradable} {
		task := &types.AgentTask{
			TaskID:     "task-1",
			ContractID: "default",
			Input:      map[string]any{"message": "hello"},
			Metadata:   map[string]string{types.SLAKeyFailureClass: fc},
		}
		if err := av.Validate(task); err != nil {
			t.Errorf("Task with failure_class=%s should pass: %v", fc, err)
		}
	}
}

func TestAdmissionValidator_Validate_SLA_InvalidFailureClass(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyFailureClass: "transient"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with invalid failure_class should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_EmptyFailureClass(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyFailureClass: ""},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with empty failure_class should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_ValidPriority(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyPriority: "128"},
	}

	if err := av.Validate(task); err != nil {
		t.Errorf("Task with valid sla.priority should pass: %v", err)
	}
}

func TestAdmissionValidator_Validate_SLA_PriorityOutOfRange(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)

	for _, val := range []string{"-1", "256", "999"} {
		task := &types.AgentTask{
			TaskID:     "task-1",
			ContractID: "default",
			Input:      map[string]any{"message": "hello"},
			Metadata:   map[string]string{types.SLAKeyPriority: val},
		}
		if err := av.Validate(task); err == nil {
			t.Errorf("Task with sla.priority=%s should be rejected (out of range)", val)
		}
	}
}

func TestAdmissionValidator_Validate_SLA_PriorityNotInteger(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyPriority: "abc"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with non-integer sla.priority should be rejected")
	}
}

func TestAdmissionValidator_Validate_SLA_ValidBackoff(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)

	for _, val := range []string{types.BackoffFixed, types.BackoffLinear, types.BackoffExponential} {
		task := &types.AgentTask{
			TaskID:     "task-1",
			ContractID: "default",
			Input:      map[string]any{"message": "hello"},
			Metadata:   map[string]string{types.SLAKeyBackoff: val},
		}
		if err := av.Validate(task); err != nil {
			t.Errorf("Task with sla.backoff=%s should pass: %v", val, err)
		}
	}
}

func TestAdmissionValidator_Validate_SLA_InvalidBackoff(t *testing.T) {
	ce := contractexec.NewContractExecutor()
	if err := ce.RegisterContract(testDefaultContract()); err != nil {
		t.Fatalf("Failed to register contract: %v", err)
	}

	av := NewAdmissionValidator(ce)
	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "default",
		Input:      map[string]any{"message": "hello"},
		Metadata:   map[string]string{types.SLAKeyBackoff: "unknown"},
	}

	if err := av.Validate(task); err == nil {
		t.Error("Task with invalid sla.backoff should be rejected")
	}
}

func testDefaultContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: "default",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "message", Type: types.FieldTypeString, Required: true},
			},
		},
	}
}
