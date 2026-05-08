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
		return fmt.Errorf("admission rejected: task ID is empty")
	}

	if task.ContractID == "" {
		return fmt.Errorf("admission rejected: contract ID is empty for task %s", task.TaskID)
	}

	if err := a.contractExecutor.ValidateInput(task.ContractID, task.Input); err != nil {
		return fmt.Errorf("admission rejected: contract %s validation failed for task %s: %w", task.ContractID, task.TaskID, err)
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
			return fmt.Errorf("admission rejected: %s=%q for task %s must be a positive integer", types.SLAKeyTimeoutMs, v, task.TaskID)
		}
	}
	if v, ok := task.Metadata[types.SLAKeyMaxRetries]; ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return fmt.Errorf("admission rejected: %s=%q for task %s must be a non-negative integer", types.SLAKeyMaxRetries, v, task.TaskID)
		}
	}
	return nil
}
