package immunity

import "errors"

// Sentinel errors for the immunity package. Callers compare via errors.Is.
var (
	// ErrSourceTaskIDRequired is returned when PromoteInput.SourceTaskID is empty.
	ErrSourceTaskIDRequired = errors.New("immunity: source_task_id is required")

	// ErrCauseRequired is returned when PromoteInput.Cause is empty.
	ErrCauseRequired = errors.New("immunity: cause is required")

	// ErrPromotedByRequired is returned when PromoteInput.PromotedBy is empty.
	ErrPromotedByRequired = errors.New("immunity: promoted_by is required")

	// ErrUnknownFailureClass is returned when a FailureClass does not match
	// a known prefix from classes.go.
	ErrUnknownFailureClass = errors.New("immunity: unknown failure class")

	// ErrTaskNotTerminal is returned by Store.Promote when the source task
	// has no terminal event.
	ErrTaskNotTerminal = errors.New("immunity: source task is not in terminal state")

	// ErrTaskNotFailed is returned by Store.Promote when the source task
	// terminated successfully (only failures may be promoted to immunity).
	ErrTaskNotFailed = errors.New("immunity: source task did not fail")

	// ErrImmunityNotFound is returned by Show/Forget for an unknown ImmunityID.
	ErrImmunityNotFound = errors.New("immunity: record not found")
)
