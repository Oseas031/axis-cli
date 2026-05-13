package tool

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

// BashVerification holds the result of verifying a bash command execution.
type BashVerification struct {
	Passed       bool     `json:"passed"`
	ExitCodeOK   bool     `json:"exit_code_ok"`
	OutputOK     bool     `json:"output_ok"`
	SideEffectOK bool     `json:"side_effect_ok"`
	Failures     []string `json:"failures,omitempty"`
}

// BashExpectation defines what a bash command execution should produce.
type BashExpectation struct {
	// ExitCode expected (default 0). Set to -1 to skip exit code check.
	ExitCode int
	// OutputContains checks stdout contains all these substrings.
	OutputContains []string
	// OutputNotContains checks stdout does NOT contain these substrings.
	OutputNotContains []string
	// FilesCreated checks these paths exist after execution.
	FilesCreated []string
	// FilesModified checks these paths exist after execution (existence check only in v1).
	FilesModified []string
}

// VerifyBashResult checks a BashTool result against expectations.
func VerifyBashResult(result map[string]any, expect *BashExpectation) *BashVerification {
	v := &BashVerification{ExitCodeOK: true, OutputOK: true, SideEffectOK: true}

	if expect == nil {
		expect = &BashExpectation{ExitCode: 0}
	}

	// Exit code verification
	if expect.ExitCode != -1 {
		exitCode, _ := result["exit_code"].(int)
		if exitCode != expect.ExitCode {
			v.ExitCodeOK = false
			v.Failures = append(v.Failures, fmt.Sprintf("exit_code: got %d, want %d", exitCode, expect.ExitCode))
		}
	}

	// Output verification
	stdout, _ := result["stdout"].(string)
	for _, s := range expect.OutputContains {
		if !strings.Contains(stdout, s) {
			v.OutputOK = false
			v.Failures = append(v.Failures, fmt.Sprintf("output missing: %q", s))
		}
	}
	for _, s := range expect.OutputNotContains {
		if strings.Contains(stdout, s) {
			v.OutputOK = false
			v.Failures = append(v.Failures, fmt.Sprintf("output contains forbidden: %q", s))
		}
	}

	// Side-effect verification
	for _, path := range expect.FilesCreated {
		if _, err := os.Stat(path); err != nil {
			v.SideEffectOK = false
			v.Failures = append(v.Failures, fmt.Sprintf("file not created: %s", path))
		}
	}
	for _, path := range expect.FilesModified {
		if _, err := os.Stat(path); err != nil {
			v.SideEffectOK = false
			v.Failures = append(v.Failures, fmt.Sprintf("file not found: %s", path))
		}
	}

	v.Passed = v.ExitCodeOK && v.OutputOK && v.SideEffectOK
	return v
}

// VerifyBashTool is an optional tool the Agent can call to verify a previous
// bash execution result. The Agent decides when verification is needed.
// This respects Zero Control: the system does not auto-judge for the Agent.
type VerifyBashTool struct{}

func NewVerifyBashTool() *VerifyBashTool { return &VerifyBashTool{} }

func (t *VerifyBashTool) Name() string { return "verify_bash" }

func (t *VerifyBashTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "verify_bash",
		Description: "Verify a bash command result. Agent calls this after bash to check exit code, output content, and file side-effects.",
		Parameters: []types.FieldDef{
			{Name: "exit_code", Type: types.FieldTypeInt, Required: true, Description: "The exit code from the bash execution"},
			{Name: "stdout", Type: types.FieldTypeString, Required: false, Description: "The stdout from the bash execution"},
			{Name: "expected_exit_code", Type: types.FieldTypeInt, Required: false, Description: "Expected exit code (default 0, -1 to skip)"},
			{Name: "output_contains", Type: types.FieldTypeString, Required: false, Description: "Comma-separated substrings that stdout must contain"},
			{Name: "output_not_contains", Type: types.FieldTypeString, Required: false, Description: "Comma-separated substrings that stdout must NOT contain"},
			{Name: "files_created", Type: types.FieldTypeString, Required: false, Description: "Comma-separated file paths that must exist"},
		},
	}
}

func (t *VerifyBashTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	exitCode := 0
	if v, ok := input["exit_code"].(float64); ok {
		exitCode = int(v)
	} else if v, ok := input["exit_code"].(int); ok {
		exitCode = v
	}

	stdout, _ := input["stdout"].(string)

	expectedExit := 0
	if v, ok := input["expected_exit_code"].(float64); ok {
		expectedExit = int(v)
	} else if v, ok := input["expected_exit_code"].(int); ok {
		expectedExit = v
	}

	var outputContains, outputNotContains, filesCreated []string
	if v, ok := input["output_contains"].(string); ok && v != "" {
		outputContains = strings.Split(v, ",")
	}
	if v, ok := input["output_not_contains"].(string); ok && v != "" {
		outputNotContains = strings.Split(v, ",")
	}
	if v, ok := input["files_created"].(string); ok && v != "" {
		filesCreated = strings.Split(v, ",")
	}

	result := map[string]any{"exit_code": exitCode, "stdout": stdout}
	expect := &BashExpectation{
		ExitCode:          expectedExit,
		OutputContains:    outputContains,
		OutputNotContains: outputNotContains,
		FilesCreated:      filesCreated,
	}

	v := VerifyBashResult(result, expect)
	return map[string]any{
		"passed":         v.Passed,
		"exit_code_ok":   v.ExitCodeOK,
		"output_ok":      v.OutputOK,
		"side_effect_ok": v.SideEffectOK,
		"failures":       v.Failures,
	}, nil
}
