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
		// Default: only check exit code == 0
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

// VerifiedBashTool wraps BashTool with automatic verification.
type VerifiedBashTool struct {
	inner  *BashTool
	expect *BashExpectation
}

// NewVerifiedBashTool creates a BashTool that auto-verifies results.
// If expect is nil, defaults to exit_code == 0 check only.
func NewVerifiedBashTool(expect *BashExpectation) *VerifiedBashTool {
	return &VerifiedBashTool{inner: NewBashTool(), expect: expect}
}

func (t *VerifiedBashTool) Name() string                  { return "bash" }
func (t *VerifiedBashTool) Schema() types.ToolDefinition  { return t.inner.Schema() }

// Execute runs the command and appends verification results.
func (t *VerifiedBashTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	result, err := t.inner.Execute(ctx, input)
	if err != nil {
		return result, err
	}
	// Skip verification if the inner call returned an error field (e.g. missing command)
	if _, hasErr := result["error"]; hasErr {
		return result, nil
	}
	v := VerifyBashResult(result, t.expect)
	result["verification"] = map[string]any{
		"passed":         v.Passed,
		"exit_code_ok":   v.ExitCodeOK,
		"output_ok":      v.OutputOK,
		"side_effect_ok": v.SideEffectOK,
		"failures":       v.Failures,
	}
	return result, nil
}
