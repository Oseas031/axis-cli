package skills

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorsAreSentinels(t *testing.T) {
	sentinels := []error{
		ErrSkillNotFound,
		ErrSkillNameRequired,
		ErrInvalidSkillName,
		ErrInvalidPath,
		ErrDescriptionRequired,
		ErrMissingSKILLMD,
	}
	for _, err := range sentinels {
		if err == nil {
			t.Error("sentinel error is nil")
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	wrapped := fmt.Errorf("loading skill pdf: %w", ErrSkillNotFound)
	if !errors.Is(wrapped, ErrSkillNotFound) {
		t.Error("wrapped error should match ErrSkillNotFound")
	}

	wrapped2 := fmt.Errorf("validate: %w", ErrInvalidSkillName)
	if !errors.Is(wrapped2, ErrInvalidSkillName) {
		t.Error("wrapped error should match ErrInvalidSkillName")
	}
}
