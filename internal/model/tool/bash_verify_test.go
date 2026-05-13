package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyBashResult_ExitCodePass(t *testing.T) {
	result := map[string]any{"exit_code": 0, "stdout": "ok\n"}
	v := VerifyBashResult(result, nil)
	if !v.Passed {
		t.Errorf("expected pass, got failures: %v", v.Failures)
	}
}

func TestVerifyBashResult_ExitCodeFail(t *testing.T) {
	result := map[string]any{"exit_code": 1, "stdout": ""}
	v := VerifyBashResult(result, &BashExpectation{ExitCode: 0})
	if v.Passed {
		t.Error("expected fail for non-zero exit code")
	}
	if v.ExitCodeOK {
		t.Error("expected ExitCodeOK=false")
	}
}

func TestVerifyBashResult_ExitCodeSkip(t *testing.T) {
	result := map[string]any{"exit_code": 42, "stdout": ""}
	v := VerifyBashResult(result, &BashExpectation{ExitCode: -1})
	if !v.Passed {
		t.Errorf("expected pass when exit code check skipped, got: %v", v.Failures)
	}
}

func TestVerifyBashResult_OutputContains(t *testing.T) {
	result := map[string]any{"exit_code": 0, "stdout": "hello world\n"}
	v := VerifyBashResult(result, &BashExpectation{OutputContains: []string{"hello", "world"}})
	if !v.Passed {
		t.Errorf("expected pass, got: %v", v.Failures)
	}
}

func TestVerifyBashResult_OutputContainsFail(t *testing.T) {
	result := map[string]any{"exit_code": 0, "stdout": "hello\n"}
	v := VerifyBashResult(result, &BashExpectation{OutputContains: []string{"world"}})
	if v.Passed {
		t.Error("expected fail for missing output")
	}
}

func TestVerifyBashResult_OutputNotContains(t *testing.T) {
	result := map[string]any{"exit_code": 0, "stdout": "error: something failed\n"}
	v := VerifyBashResult(result, &BashExpectation{OutputNotContains: []string{"error:"}})
	if v.Passed {
		t.Error("expected fail for forbidden output")
	}
}

func TestVerifyBashResult_SideEffectFileCreated(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "out.txt")
	os.WriteFile(f, []byte("x"), 0644)

	result := map[string]any{"exit_code": 0, "stdout": ""}
	v := VerifyBashResult(result, &BashExpectation{FilesCreated: []string{f}})
	if !v.Passed {
		t.Errorf("expected pass, got: %v", v.Failures)
	}
}

func TestVerifyBashResult_SideEffectFileMissing(t *testing.T) {
	result := map[string]any{"exit_code": 0, "stdout": ""}
	v := VerifyBashResult(result, &BashExpectation{FilesCreated: []string{"/nonexistent/file.txt"}})
	if v.Passed {
		t.Error("expected fail for missing file")
	}
}

func TestVerifyBashTool_Name(t *testing.T) {
	vt := NewVerifyBashTool()
	if vt.Name() != "verify_bash" {
		t.Errorf("expected name verify_bash, got %q", vt.Name())
	}
}

func TestVerifyBashTool_Execute_Pass(t *testing.T) {
	vt := NewVerifyBashTool()
	result, err := vt.Execute(context.Background(), map[string]any{
		"exit_code": 0,
		"stdout":    "hello world",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["passed"] != true {
		t.Errorf("expected passed=true, got %v", result)
	}
}

func TestVerifyBashTool_Execute_Fail(t *testing.T) {
	vt := NewVerifyBashTool()
	result, err := vt.Execute(context.Background(), map[string]any{
		"exit_code":          float64(1),
		"stdout":             "some output",
		"expected_exit_code": float64(0),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["passed"] != false {
		t.Error("expected passed=false for exit code mismatch")
	}
}

func TestVerifyBashTool_Execute_OutputCheck(t *testing.T) {
	vt := NewVerifyBashTool()
	result, err := vt.Execute(context.Background(), map[string]any{
		"exit_code":           float64(0),
		"stdout":              "build successful",
		"output_contains":     "successful",
		"output_not_contains": "error",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["passed"] != true {
		t.Errorf("expected passed=true, got %v", result)
	}
}

func TestVerifyBashTool_ImplementsInterface(t *testing.T) {
	var _ Tool = NewVerifyBashTool()
}
