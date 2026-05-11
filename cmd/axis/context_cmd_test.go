package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axis-cli/axis/internal/contextpack"
	"github.com/axis-cli/axis/internal/intent"
)

func TestContextPreviewCommand_SelectsProviderContext(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"context", "preview", "fix", "minimax", "provider", "config", "--task-id", "ctx-test"})
	if err := root.Execute(); err != nil {
		t.Fatalf("context preview should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Context bundle preview:") {
		t.Fatalf("expected context bundle preview, got %q", output)
	}
	if !strings.Contains(output, "spec:model-provider") {
		t.Fatalf("expected provider context packet, got %q", output)
	}
	if !strings.Contains(output, "Not executed") {
		t.Fatalf("expected non-execution hint, got %q", output)
	}
}

func TestContextPreviewCommand_Stdin(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetIn(strings.NewReader("improve interactive shell help"))
	root.SetArgs([]string{"context", "preview", "--stdin", "--task-id", "ctx-stdin"})
	if err := root.Execute(); err != nil {
		t.Fatalf("context preview stdin should succeed: %v", err)
	}
	if !strings.Contains(out.String(), "spec:interactive-shell") {
		t.Fatalf("expected shell context packet, got %q", out.String())
	}
}

func TestContextPreviewCommand_BudgetShowsExclusions(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"context", "preview", "ask", "provider", "shell", "context", "scheduler", "axis-up", "--max-packets", "1"})
	if err := root.Execute(); err != nil {
		t.Fatalf("context preview with budget should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "\"truncated\": true") {
		t.Fatalf("expected truncated budget, got %q", output)
	}
	if !strings.Contains(output, "excluded by packet count budget") && !strings.Contains(output, "excluded by context budget") {
		t.Fatalf("expected budget exclusion reason, got %q", output)
	}
}

func TestContextPreviewCommand_MissingPrompt(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	root.SetArgs([]string{"context", "preview"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected missing prompt to fail")
	}
}

func TestContextInspectCommand_AfterAskSubmitWithContext(t *testing.T) {
	resetCLIState()
	if err := submitReadyContextTask("fix provider config", "inspect-context-task"); err != nil {
		t.Fatalf("prepare context task: %v", err)
	}
	bundleID := submittedTaskMetadata("inspect-context-task")["context.bundle_id"]
	if bundleID == "" {
		t.Fatal("expected context bundle id after submit")
	}

	inspectRoot := NewRootCommand(&App{providerName: "mock"})
	var inspectOut bytes.Buffer
	inspectRoot.SetOut(&inspectOut)
	inspectRoot.SetArgs([]string{"context", "inspect", bundleID})
	if err := inspectRoot.Execute(); err != nil {
		t.Fatalf("context inspect should succeed: %v", err)
	}
	output := inspectOut.String()
	if !strings.Contains(output, "Context readiness record:") {
		t.Fatalf("expected readiness record output, got %q", output)
	}
	if !strings.Contains(output, bundleID) {
		t.Fatalf("expected bundle id in inspect output, got %q", output)
	}
	if !strings.Contains(output, "spec:model-provider") {
		t.Fatalf("expected provider packet in inspect output, got %q", output)
	}
}

func TestContextInspectCommand_MissingRecord(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	root.SetArgs([]string{"context", "inspect", "ctx-missing"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected missing readiness record to fail")
	}
}

func submittedTaskMetadata(taskID string) map[string]string {
	for _, task := range orch.GetAllTasks() {
		if task.TaskID == taskID {
			return task.Metadata
		}
	}
	return nil
}

func submitReadyContextTask(prompt string, taskID string) error {
	initOrchestrator()
	result, err := intent.NewDeterministicParser().Parse(context.Background(), intent.Request{Prompt: prompt, ContractID: "default", TaskID: taskID})
	if err != nil {
		return err
	}
	bundle, err := assembleContextForTask(result.Task)
	if err != nil {
		return err
	}
	artifact, err := contextpack.DefaultRegistry.Register(bundle)
	if err != nil {
		return err
	}
	if err := contextpack.AttachReadinessMetadata(result.Task, artifact); err != nil {
		return err
	}
	return orch.SubmitTask(result.Task)
}

func TestContextPreflightCommand_ReadyAfterAskSubmitWithContext(t *testing.T) {
	resetCLIState()
	if err := submitReadyContextTask("fix provider config", "preflight-ready-task"); err != nil {
		t.Fatalf("prepare context task: %v", err)
	}

	preflightRoot := NewRootCommand(&App{providerName: "mock"})
	var preflightOut bytes.Buffer
	preflightRoot.SetOut(&preflightOut)
	preflightRoot.SetArgs([]string{"context", "preflight", "preflight-ready-task"})
	if err := preflightRoot.Execute(); err != nil {
		t.Fatalf("context preflight should succeed: %v", err)
	}
	output := preflightOut.String()
	if !strings.Contains(output, "Context readiness preflight:") {
		t.Fatalf("expected preflight output, got %q", output)
	}
	if !strings.Contains(output, "\"status\": \"ready\"") {
		t.Fatalf("expected ready status, got %q", output)
	}
	if !strings.Contains(output, "preflight-ready-task") {
		t.Fatalf("expected task id in preflight output, got %q", output)
	}
}

func TestContextPreflightCommand_MissingReadiness(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"run", "plain-task"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run should succeed: %v", err)
	}

	preflightRoot := NewRootCommand(&App{providerName: "mock"})
	var preflightOut bytes.Buffer
	preflightRoot.SetOut(&preflightOut)
	preflightRoot.SetArgs([]string{"context", "preflight", "plain-task"})
	if err := preflightRoot.Execute(); err != nil {
		t.Fatalf("context preflight should succeed: %v", err)
	}
	if !strings.Contains(preflightOut.String(), "\"status\": \"missing\"") {
		t.Fatalf("expected missing readiness, got %q", preflightOut.String())
	}
}

func TestContextPreflightCommand_MissingTask(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"context", "preflight", "missing-task"})
	if err := root.Execute(); err != nil {
		t.Fatalf("context preflight should succeed for missing task: %v", err)
	}
	if !strings.Contains(out.String(), "\"reason\": \"task is required\"") {
		t.Fatalf("expected missing task reason, got %q", out.String())
	}
}

func TestContextPreflightCommand_StrictReadyPasses(t *testing.T) {
	resetCLIState()
	if err := submitReadyContextTask("fix provider config", "preflight-strict-ready"); err != nil {
		t.Fatalf("prepare context task: %v", err)
	}

	preflightRoot := NewRootCommand(&App{providerName: "mock"})
	var preflightOut bytes.Buffer
	preflightRoot.SetOut(&preflightOut)
	preflightRoot.SetArgs([]string{"context", "preflight", "preflight-strict-ready", "--strict"})
	if err := preflightRoot.Execute(); err != nil {
		t.Fatalf("strict preflight should pass for ready task: %v", err)
	}
	if !strings.Contains(preflightOut.String(), "\"status\": \"ready\"") {
		t.Fatalf("expected ready status, got %q", preflightOut.String())
	}
}

func TestContextPreflightCommand_StrictMissingFails(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var runOut bytes.Buffer
	root.SetOut(&runOut)
	root.SetArgs([]string{"run", "preflight-strict-missing"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run should succeed: %v", err)
	}

	preflightRoot := NewRootCommand(&App{providerName: "mock"})
	var preflightOut bytes.Buffer
	preflightRoot.SetOut(&preflightOut)
	preflightRoot.SetArgs([]string{"context", "preflight", "preflight-strict-missing", "--strict"})
	if err := preflightRoot.Execute(); err == nil {
		t.Fatal("strict preflight should fail for missing readiness")
	}
	if !strings.Contains(preflightOut.String(), "\"status\": \"missing\"") {
		t.Fatalf("expected missing status output, got %q", preflightOut.String())
	}
}

func TestRenderContextBundle_ReturnsWriteError(t *testing.T) {
	bundle := &contextpack.ContextBundle{TaskID: "b1"}
	writeErr := errors.New("disk full")
	if err := renderContextBundle(bundle, &failingWriter{err: writeErr}); err == nil {
		t.Fatal("expected renderContextBundle to propagate write error")
	}
}

func TestRenderReadinessRecord_ReturnsWriteError(t *testing.T) {
	record := contextpack.ReadinessRecord{Artifact: contextpack.ReadinessArtifact{BundleID: "b1"}}
	writeErr := errors.New("disk full")
	if err := renderReadinessRecord(record, &failingWriter{err: writeErr}); err == nil {
		t.Fatal("expected renderReadinessRecord to propagate write error")
	}
}

func TestContextIndexCommand_Rebuild(t *testing.T) {
	resetCLIState()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("Axis CLI documentation"), 0644)

	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"context", "index", "--rebuild", "--root", dir})
	if err := root.Execute(); err != nil {
		t.Fatalf("context index rebuild should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Context index:") {
		t.Fatalf("expected index output, got %q", output)
	}
	if !strings.Contains(output, "rebuilt") {
		t.Fatalf("expected rebuilt message, got %q", output)
	}
}

func TestContextIndexCommand_Update(t *testing.T) {
	resetCLIState()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("hello world"), 0644)

	// First rebuild
	mgr := contextpack.NewIndexManager()
	if _, err := mgr.Rebuild(dir); err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}

	// Add a new file
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\n"), 0644)

	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"context", "index", "--update", "--root", dir})
	if err := root.Execute(); err != nil {
		t.Fatalf("context index update should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "updated") {
		t.Fatalf("expected updated message, got %q", output)
	}
}

func TestContextIndexCommand_Status(t *testing.T) {
	resetCLIState()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doc.md"), []byte("document content"), 0644)

	mgr := contextpack.NewIndexManager()
	if _, err := mgr.Rebuild(dir); err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}

	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"context", "index", "--status", "--root", dir})
	if err := root.Execute(); err != nil {
		t.Fatalf("context index status should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "healthy") {
		t.Fatalf("expected healthy status, got %q", output)
	}
	if !strings.Contains(output, "index loaded") {
		t.Fatalf("expected index loaded message, got %q", output)
	}
}

func TestContextIndexCommand_StatusMissing(t *testing.T) {
	resetCLIState()
	dir := t.TempDir()

	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"context", "index", "--status", "--root", dir})
	if err := root.Execute(); err != nil {
		t.Fatalf("context index status for missing should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "index not found") {
		t.Fatalf("expected index not found, got %q", output)
	}
}

func TestContextIndexCommand_MissingFlag(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	root.SetArgs([]string{"context", "index"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected missing flag to fail")
	}
}

func TestRenderPreflightResult_ReturnsWriteError(t *testing.T) {
	result := contextpack.PreflightResult{Status: contextpack.PreflightStatusReady}
	writeErr := errors.New("disk full")
	if err := renderPreflightResult(result, &failingWriter{err: writeErr}); err == nil {
		t.Fatal("expected renderPreflightResult to propagate write error")
	}
}
