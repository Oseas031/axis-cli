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
	if v.OutputOK {
		t.Error("expected OutputOK=false")
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
	if v.SideEffectOK {
		t.Error("expected SideEffectOK=false")
	}
}

func TestVerifiedBashTool_Execute(t *testing.T) {
	vt := NewVerifiedBashTool(nil)
	result, err := vt.Execute(context.Background(), map[string]any{"command": "echo verified"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ver, ok := result["verification"].(map[string]any)
	if !ok {
		t.Fatal("expected verification field in result")
	}
	if ver["passed"] != true {
		t.Errorf("expected verification passed, got: %v", ver)
	}
}

func TestVerifiedBashTool_FailingCommand(t *testing.T) {
	vt := NewVerifiedBashTool(&BashExpectation{ExitCode: 0})
	result, err := vt.Execute(context.Background(), map[string]any{"command": "exit 1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ver, ok := result["verification"].(map[string]any)
	if !ok {
		t.Fatal("expected verification field in result")
	}
	if ver["passed"] != false {
		t.Error("expected verification failed for exit 1")
	}
}

func TestVerifiedBashTool_ImplementsInterface(t *testing.T) {
	var _ Tool = NewVerifiedBashTool(nil)
}
