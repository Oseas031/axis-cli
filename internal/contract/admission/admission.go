// Package admission provides pre-scheduling task validation.
package admission

import (
	"fmt"
	"strconv"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	"github.com/axis-cli/axis/internal/types"
)

// AdmissionValidator validates tasks before they enter the scheduler.
type AdmissionValidator interface {
	Validate(task *types.AgentTask) error
}

// AdmissionValidatorImpl implements AdmissionValidator using the contract registry.
type AdmissionValidatorImpl struct {
	contractExecutor contractexec.ContractExecutor
}

// NewAdmissionValidator creates a new admission validator.
func NewAdmissionValidator(ce contractexec.ContractExecutor) *AdmissionValidatorImpl {
	return &AdmissionValidatorImpl{contractExecutor: ce}
}

// Validate checks a task before scheduling. It verifies the task has a non-empty
// TaskID, the referenced contract exists, and the input satisfies the contract schema.
func (a *AdmissionValidatorImpl) Validate(task *types.AgentTask) error {
	if task.TaskID == "" {
		return types.NewAgentError(types.ErrContractInputInvalid, "admission rejected: task ID is empty")
	}

	if task.ContractID == "" {
		return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: contract ID is empty for task %s", task.TaskID))
	}

	if err := a.contractExecutor.ValidateInput(task.ContractID, task.Input); err != nil {
		return types.NewAgentErrorWithCause(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: contract %s validation failed for task %s", task.ContractID, task.TaskID), err)
	}

	if err := validateSLA(task); err != nil {
		return err
	}

	return nil
}

// validateSLA checks SLA metadata values for validity.
func validateSLA(task *types.AgentTask) error {
	if v, ok := task.Metadata[types.SLAKeyTimeoutMs]; ok {
		ms, err := strconv.Atoi(v)
		if err != nil || ms <= 0 {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: %s=%q for task %s must be a positive integer", types.SLAKeyTimeoutMs, v, task.TaskID))
		}
	}
	if v, ok := task.Metadata[types.SLAKeyMaxRetries]; ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: %s=%q for task %s must be a non-negative integer", types.SLAKeyMaxRetries, v, task.TaskID))
		}
		if n > types.MaxRetryLimit {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: %s=%d for task %s exceeds MaxRetryLimit (%d)", types.SLAKeyMaxRetries, n, task.TaskID, types.MaxRetryLimit))
		}
	}
	if v, ok := task.Metadata[types.SLAKeyFailureClass]; ok {
		if v != types.FailureClassRetryable && v != types.FailureClassFatal && v != types.FailureClassDegradable {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: %s=%q for task %s must be one of: retryable, fatal, degradable", types.SLAKeyFailureClass, v, task.TaskID))
		}
	}
	if v, ok := task.Metadata[types.SLAKeyPriority]; ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 255 {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: %s=%q for task %s must be an integer in [0,255]", types.SLAKeyPriority, v, task.TaskID))
		}
	}
	if v, ok := task.Metadata[types.SLAKeyBackoff]; ok {
		if v != types.BackoffFixed && v != types.BackoffLinear && v != types.BackoffExponential {
			return types.NewAgentError(types.ErrContractInputInvalid, fmt.Sprintf("admission rejected: %s=%q for task %s must be one of: fixed, linear, exponential", types.SLAKeyBackoff, v, task.TaskID))
		}
	}
	return nil
}
