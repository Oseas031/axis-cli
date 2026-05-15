package provider

import (
	"strings"
	"testing"
)

func TestPromptAssembler_Empty(t *testing.T) {
	pa := NewPromptAssembler()
	if got := pa.Build(); got != "" {
		t.Fatal("empty assembler should return empty string")
	}
}

func TestPromptAssembler_Ordering(t *testing.T) {
	pa := NewPromptAssembler()
	pa.AddLayer("append", "APPEND", PriorityAppend)
	pa.AddLayer("default", "DEFAULT", PriorityDefault)
	pa.AddLayer("project", "PROJECT", PriorityProject)
	pa.AddLayer("task", "TASK", PriorityTask)

	result := pa.Build()
	defIdx := strings.Index(result, "DEFAULT")
	projIdx := strings.Index(result, "PROJECT")
	taskIdx := strings.Index(result, "TASK")
	appIdx := strings.Index(result, "APPEND")

	if defIdx > projIdx || projIdx > taskIdx || taskIdx > appIdx {
		t.Fatalf("wrong order: def=%d proj=%d task=%d app=%d", defIdx, projIdx, taskIdx, appIdx)
	}
}

func TestPromptAssembler_SkipsEmpty(t *testing.T) {
	pa := NewPromptAssembler()
	pa.AddLayer("default", "BASE", PriorityDefault)
	pa.AddLayer("project", "", PriorityProject) // empty, should be skipped
	pa.AddLayer("task", "TASK", PriorityTask)

	result := pa.Build()
	if strings.Contains(result, "\n\n\n") {
		t.Fatal("should not have triple newlines from empty layer")
	}
}

func TestPromptAssembler_ProjectOverridesDefault(t *testing.T) {
	pa := NewPromptAssembler()
	pa.AddLayer("default", "You are a helpful assistant.", PriorityDefault)
	pa.AddLayer("project", "You are an Axis agent. Follow CLAUDE.md.", PriorityProject)

	result := pa.Build()
	// Project comes after default (higher priority = later in output)
	if strings.Index(result, "helpful") > strings.Index(result, "Axis agent") {
		t.Fatal("project should come after default")
	}
}
