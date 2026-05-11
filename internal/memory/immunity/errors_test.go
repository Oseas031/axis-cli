package immunity

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrors_Distinct(t *testing.T) {
	all := []error{
		ErrSourceTaskIDRequired,
		ErrCauseRequired,
		ErrPromotedByRequired,
		ErrUnknownFailureClass,
		ErrTaskNotTerminal,
		ErrTaskNotFailed,
		ErrImmunityNotFound,
	}
	for i, a := range all {
		for j, b := range all {
			if i == j {
				continue
			}
			if errors.Is(a, b) {
				t.Errorf("errors should be distinct: %v Is %v", a, b)
			}
		}
	}
}

func TestErrors_WrappedIsDetectable(t *testing.T) {
	wrapped := fmt.Errorf("promote failed: %w", ErrCauseRequired)
	if !errors.Is(wrapped, ErrCauseRequired) {
		t.Errorf("errors.Is should detect wrapped sentinel; got false")
	}
	if errors.Is(wrapped, ErrTaskNotFailed) {
		t.Errorf("errors.Is should not match unrelated sentinel")
	}
}

func TestErrors_HaveImmunityPrefix(t *testing.T) {
	all := []error{
		ErrSourceTaskIDRequired,
		ErrCauseRequired,
		ErrPromotedByRequired,
		ErrUnknownFailureClass,
		ErrTaskNotTerminal,
		ErrTaskNotFailed,
		ErrImmunityNotFound,
	}
	for _, e := range all {
		if !contains(e.Error(), "immunity:") {
			t.Errorf("error %q missing 'immunity:' prefix", e.Error())
		}
	}
}
