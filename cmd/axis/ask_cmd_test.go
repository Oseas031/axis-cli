package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/control"
	"github.com/axis-cli/axis/internal/types"
)

type failingWriter struct {
	err error
}

func (w *failingWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

func TestAskCommand_DryRunByDefault(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"ask", "check", "provider", "config", "--task-id", "ask-test"})
	if err := root.Execute(); err != nil {
		t.Fatalf("ask dry-run should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Task proposal:") {
		t.Fatalf("expected task proposal, got %q", output)
	}
	if !strings.Contains(output, "ask-test") {
		t.Fatalf("expected task ID in proposal, got %q", output)
	}
	if !strings.Contains(output, "original_prompt") {
		t.Fatalf("expected provenance metadata, got %q", output)
	}
	if !strings.Contains(output, "Not submitted") {
		t.Fatalf("expected dry-run not submitted hint, got %q", output)
	}
}

func TestAskCommand_SubmitRequiresLocalRuntime(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"ask", "check provider config", "--task-id", "ask-submit", "--submit"})
	err := root.Execute()
	if err == nil {
		t.Fatal("ask submit should require a local runtime")
	}
	if !strings.Contains(err.Error(), "axis start") {
		t.Fatalf("expected axis start guidance, got %v", err)
	}
}

func TestAskCommand_SubmitUsesLocalRuntime(t *testing.T) {
	resetCLIState()
	rootDir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- runLocalRuntime(ctx, rootDir, nil)
	}()
	locator := control.NewRuntimeLocator(rootDir)
	for i := 0; i < 50; i++ {
		if _, err := locator.Load(); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	resetCLIState()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("runtime did not stop")
		}
	})

	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"ask", "check provider config", "--task-id", "ask-runtime-submit", "--submit"})
	if err := root.Execute(); err != nil {
		t.Fatalf("ask submit should use local runtime: %v", err)
	}
	if !strings.Contains(out.String(), "Task ask-runtime-submit submitted") {
		t.Fatalf("expected submit confirmation, got %q", out.String())
	}

	client := control.NewClient(locator, http.DefaultClient)
	status, err := client.Status(context.Background(), "ask-runtime-submit")
	if err != nil {
		t.Fatalf("expected runtime status: %v", err)
	}
	if status.TaskID != "ask-runtime-submit" || status.Status == "" {
		t.Fatalf("unexpected runtime status: %#v", status)
	}
}

func TestAskCommand_WithContextPreview(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"ask", "fix", "minimax", "provider", "config", "--task-id", "ask-context", "--with-context"})
	if err := root.Execute(); err != nil {
		t.Fatalf("ask with context should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Task proposal:") {
		t.Fatalf("expected task proposal, got %q", output)
	}
	if !strings.Contains(output, "Context bundle preview:") {
		t.Fatalf("expected context bundle preview, got %q", output)
	}
	if !strings.Contains(output, "spec:model-provider") {
		t.Fatalf("expected provider context packet, got %q", output)
	}
	if !strings.Contains(output, "Not submitted") {
		t.Fatalf("expected dry-run not submitted hint, got %q", output)
	}
}

func TestAskCommand_WithContextSubmitAttachesReadinessMetadata(t *testing.T) {
	resetCLIState()
	rootDir := t.TempDir()
	var submitted control.SubmitTaskRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tasks" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&submitted); err != nil {
			t.Fatalf("decode submit request: %v", err)
		}
		controlResponse := control.SubmitTaskResponse{TaskID: submitted.Task.TaskID, Status: submitted.Task.Status}
		if err := json.NewEncoder(w).Encode(controlResponse); err != nil {
			t.Fatalf("encode submit response: %v", err)
		}
	}))
	defer server.Close()
	locator := control.NewRuntimeLocator(rootDir)
	if err := locator.Save(control.RuntimeRecord{Protocol: "http", Address: server.URL, ProjectRoot: rootDir, StartedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("save runtime locator: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"ask", "fix provider config", "--task-id", "ask-context-submit", "--with-context", "--submit"})
	if err := root.Execute(); err != nil {
		t.Fatalf("ask with context submit should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Task ask-context-submit submitted") {
		t.Fatalf("expected submit confirmation, got %q", output)
	}
	if strings.Contains(output, "Context bundle preview:") {
		t.Fatalf("submit path should not render context preview, got %q", output)
	}
	if submitted.Task == nil {
		t.Fatal("expected submitted task")
	}
	submittedMetadata := submitted.Task.Metadata
	if submittedMetadata["context.bundle_id"] == "" {
		t.Fatalf("expected context readiness bundle id, got %+v", submittedMetadata)
	}
	if submittedMetadata["context.assembly_mode"] != "rule_based" {
		t.Fatalf("expected rule-based context readiness metadata, got %+v", submittedMetadata)
	}
	if submittedMetadata["context.packet_count"] == "" {
		t.Fatalf("expected packet count metadata, got %+v", submittedMetadata)
	}
}

func TestAskCommand_Stdin(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetIn(strings.NewReader("inspect tools"))
	root.SetArgs([]string{"ask", "--stdin", "--task-id", "ask-stdin"})
	if err := root.Execute(); err != nil {
		t.Fatalf("ask stdin should succeed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "inspect tools") {
		t.Fatalf("expected stdin prompt in proposal, got %q", output)
	}
}

func TestRenderTaskProposal_ReturnsWriteError(t *testing.T) {
	task := &types.AgentTask{TaskID: "write-error-test"}
	writeErr := errors.New("disk full")
	if err := renderTaskProposal(task, &failingWriter{err: writeErr}); err == nil {
		t.Fatal("expected renderTaskProposal to propagate write error")
	}
}

func TestAskCommand_SubmitAndDryRunConflict(t *testing.T) {
	resetCLIState()
	root := NewRootCommand(&App{providerName: "mock"})
	root.SetArgs([]string{"ask", "check", "--submit", "--dry-run"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected --submit and --dry-run conflict")
	}
}
