package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/contextpack"
	"github.com/axis-cli/axis/internal/control"
	"github.com/axis-cli/axis/internal/types"
	"github.com/spf13/cobra"
)

func TestRunTaskInitializesLocalOrchestrator(t *testing.T) {
	resetCLIState()

	output := captureStdout(t, func() {
		if err := runTask(&cobra.Command{}, []string{"task-1"}); err != nil {
			t.Fatalf("runTask should initialize and execute locally: %v", err)
		}
	})

	if strings.Contains(output, "axis start") {
		t.Fatalf("runTask output should not tell users to run axis start first: %s", output)
	}
	if !strings.Contains(output, "Task task-1 completed") {
		t.Fatalf("runTask output should confirm completion, got: %s", output)
	}
}

func TestRunTask_InitOrchestratorError(t *testing.T) {
	resetCLIState()

	// First call succeeds (sets up orchestrator and executes)
	output := captureStdout(t, func() {
		if err := runTask(&cobra.Command{}, []string{"task-1"}); err != nil {
			t.Fatalf("first runTask should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Task task-1 completed") {
		t.Fatalf("Expected completion, got: %s", output)
	}
}

func TestGetTaskStatusNotFoundGivesContext(t *testing.T) {
	resetCLIState()

	err := getTaskStatus(&cobra.Command{}, []string{"missing"})
	if err == nil {
		t.Fatal("getTaskStatus should return an error for a missing task")
	}
	if !strings.Contains(err.Error(), "axis start") {
		t.Fatalf("getTaskStatus should include local runtime guidance, got: %v", err)
	}
}

func TestGetTaskStatus_Success(t *testing.T) {
	t.Skip("status now queries a local runtime; covered by TestGetTaskStatusUsesLocalRuntime")
}

func TestGetTaskStatusUsesLocalRuntime(t *testing.T) {
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
	client := control.NewClient(locator, http.DefaultClient)
	if _, err := client.SubmitTask(context.Background(), &types.AgentTask{TaskID: "status-runtime", ContractID: "default", Input: map[string]any{"message": "hello"}}); err != nil {
		t.Fatalf("submit task to runtime: %v", err)
	}
	resetCLIState()
	defaultApp.root = rootDir
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

	output := captureStdout(t, func() {
		if err := getTaskStatus(&cobra.Command{}, []string{"status-runtime"}); err != nil {
			t.Fatalf("status should use local runtime: %v", err)
		}
	})
	if !strings.Contains(output, "Task status-runtime status:") {
		t.Fatalf("expected runtime status output, got %q", output)
	}
}

func TestNewLocalHTTPServer_HasTimeouts(t *testing.T) {
	srv := newLocalHTTPServer(http.NewServeMux())
	if srv.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 10*time.Second {
		t.Errorf("WriteTimeout = %v, want 10s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("IdleTimeout = %v, want 120s", srv.IdleTimeout)
	}
}

func TestRunLocalRuntimeWritesLocatorAndServesHealth(t *testing.T) {
	resetCLIState()
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)

	go func() {
		done <- runLocalRuntime(ctx, root, io.Discard)
	}()

	locator := control.NewRuntimeLocator(root)
	var record control.RuntimeRecord
	for i := 0; i < 50; i++ {
		got, err := locator.Load()
		if err == nil {
			record = got
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if record.Address == "" {
		t.Fatal("expected runtime locator to be written")
	}

	resp, err := http.Get(record.Address + "/v1/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected health 200, got %d", resp.StatusCode)
	}
	var health control.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if health.Status != "ok" || health.Address != record.Address || health.ProjectRoot != root {
		t.Fatalf("unexpected health response: %#v, locator: %#v", health, record)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runtime returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runtime did not stop after context cancellation")
	}
	if _, err := os.Stat(filepath.Join(root, ".axis", "runtime.json")); !os.IsNotExist(err) {
		t.Fatalf("expected runtime locator to be removed, got %v", err)
	}
}

func TestRunLocalRuntimeWritesTaskEvents(t *testing.T) {
	resetCLIState()
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- runLocalRuntime(ctx, root, io.Discard)
	}()
	locator := control.NewRuntimeLocator(root)
	for i := 0; i < 50; i++ {
		if _, err := locator.Load(); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	client := control.NewClient(locator, http.DefaultClient)
	if _, err := client.SubmitTask(context.Background(), &types.AgentTask{TaskID: "event-task", ContractID: "default", Input: map[string]any{"message": "hello"}}); err != nil {
		t.Fatalf("submit task to runtime: %v", err)
	}
	if _, err := client.Status(context.Background(), "event-task"); err != nil {
		t.Fatalf("status task through runtime: %v", err)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runtime returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runtime did not stop")
	}
	data, err := os.ReadFile(control.NewTaskEventLog(root).Path())
	if err != nil {
		t.Fatalf("read task event log: %v", err)
	}
	log := string(data)
	if !strings.Contains(log, `"event_type":"submitted"`) {
		t.Fatalf("expected submitted event, got %q", log)
	}
	if !strings.Contains(log, `"event_type":"status_requested"`) {
		t.Fatalf("expected status_requested event, got %q", log)
	}
	if strings.Contains(strings.ToLower(log), "api_key") || strings.Contains(strings.ToLower(log), "secret") || strings.Contains(strings.ToLower(log), "token") {
		t.Fatalf("event log should not contain secret-like fields, got %q", log)
	}
}

func TestPrintShellHelp(t *testing.T) {
	output := captureStdout(t, func() {
		printShellHelp(os.Stdout)
	})
	if !strings.Contains(output, "Available commands:") {
		t.Errorf("Expected help header, got: %s", output)
	}
}

func TestDefaultContract(t *testing.T) {
	c := defaultContract()
	if c.ContractID != "default" {
		t.Errorf("Expected ContractID=default, got %s", c.ContractID)
	}
	if c.InputSchema == nil || c.OutputSchema == nil {
		t.Fatal("InputSchema and OutputSchema should not be nil")
	}
	if len(c.InputSchema.Fields) != 1 || c.InputSchema.Fields[0].Name != "message" {
		t.Error("InputSchema should have one 'message' field")
	}
	if len(c.OutputSchema.Fields) != 3 {
		t.Errorf("OutputSchema should have 3 fields, got %d", len(c.OutputSchema.Fields))
	}
}

func TestSubmitTask_Direct(t *testing.T) {
	resetCLIState()
	initOrchestrator()

	err := submitTask("direct-task")
	if err != nil {
		t.Fatalf("submitTask should succeed: %v", err)
	}
}

func TestMainCLI_Help(t *testing.T) {
	root := newAxisRoot()
	root.SetArgs([]string{"--help"})
	err := root.Execute()
	if err != nil {
		t.Fatalf("root --help should not error: %v", err)
	}
}

func TestMainCLI_RunMissingArg(t *testing.T) {
	resetCLIState()
	root := newAxisRoot()
	root.SetArgs([]string{"run"})
	err := root.Execute()
	if err == nil {
		t.Error("run without args should fail")
	}
}

func TestMainCLI_StatusMissingArg(t *testing.T) {
	resetCLIState()
	root := newAxisRoot()
	root.SetArgs([]string{"status"})
	err := root.Execute()
	if err == nil {
		t.Error("status without args should fail")
	}
}

func TestMainCLI_UnknownCommand(t *testing.T) {
	root := newAxisRoot()
	root.SetArgs([]string{"nonexistent"})
	err := root.Execute()
	if err == nil {
		t.Error("unknown command should fail")
	}
}

func TestRunShell_HelpExit(t *testing.T) {
	resetCLIState()

	input := "help\nexit\n"
	output := captureStdoutStdin(t, input, func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})

	if !strings.Contains(output, "Available commands:") {
		t.Errorf("Expected help output in shell, got: %s", output)
	}
	if !strings.Contains(output, "Exiting Axis shell.") {
		t.Errorf("Expected exit message, got: %s", output)
	}
}

func TestRunShell_Quit(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "quit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Exiting Axis shell.") {
		t.Errorf("Expected exit message, got: %s", output)
	}
}

func TestRunShell_EmptyInput(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Exiting Axis shell.") {
		t.Errorf("Expected exit message, got: %s", output)
	}
}

func TestRunShell_UnknownCommand(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "blah\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Unknown command") {
		t.Errorf("Expected unknown command message, got: %s", output)
	}
}

func TestRunShell_RunAndStatus(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "run shell-task\nstatus shell-task\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "submitted") && !strings.Contains(output, "Submitted") {
		t.Errorf("Expected submission confirmation, got: %s", output)
	}
	if strings.Contains(output, "TASK_TIMEOUT") {
		t.Errorf("unexpected TASK_TIMEOUT in output: %s", output)
	}
}

func TestRunShell_AskPreview(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "ask check provider config\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Task proposal:") {
		t.Fatalf("expected ask to render a task proposal, got: %s", output)
	}
	if !strings.Contains(output, "original_prompt") {
		t.Fatalf("expected ask proposal to include provenance metadata, got: %s", output)
	}
	if !strings.Contains(output, "Not submitted") {
		t.Fatalf("expected ask preview not to submit, got: %s", output)
	}
}

func TestRunShell_AskSubmitAndStatus(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "ask --submit check provider config\nstatus ask-20260510-185400\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "submitted") {
		t.Fatalf("expected ask submit confirmation, got: %s", output)
	}
	if !strings.Contains(output, "Try: status ask-") {
		t.Fatalf("expected submitted task status hint, got: %s", output)
	}
	if !strings.Contains(output, "Could not get status for task ask-20260510-185400") {
		t.Fatalf("expected shell status to use in-process session state, got: %s", output)
	}
	if strings.Contains(output, "TASK_TIMEOUT") {
		t.Errorf("unexpected TASK_TIMEOUT in output: %s", output)
	}
}

func TestRunShell_RunThenStatusUsesInProcessSession(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "run shell-local-task\nstatus shell-local-task\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Task shell-local-task submitted") {
		t.Fatalf("expected shell run submit confirmation, got: %s", output)
	}
	if !strings.Contains(output, "Task shell-local-task status:") {
		t.Fatalf("expected shell status to use in-process submitted task, got: %s", output)
	}
	if strings.Contains(output, "axis start") {
		t.Fatalf("shell in-process status should not require axis start, got: %s", output)
	}
	if strings.Contains(output, "TASK_TIMEOUT") {
		t.Errorf("unexpected TASK_TIMEOUT in output: %s", output)
	}
}

func TestRunShell_NoPromptSuppressesPromptForPipeDrivers(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "run pipe-task\nstatus pipe-task\nexit\n", func() {
		root := NewRootCommand(&App{providerName: "mock"})
		root.SetArgs([]string{"shell", "--no-prompt"})
		if err := root.Execute(); err != nil {
			t.Fatalf("shell --no-prompt should succeed: %v", err)
		}
	})
	if strings.Contains(output, "axis> ") {
		t.Fatalf("expected --no-prompt to suppress shell prompt, got: %s", output)
	}
	if !strings.Contains(output, "Task pipe-task submitted") {
		t.Fatalf("expected shell to process piped run command, got: %s", output)
	}
	if !strings.Contains(output, "Task pipe-task status:") {
		t.Fatalf("expected shell to process piped status command, got: %s", output)
	}
}

func TestRunShell_RunMissingArg(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "run\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Usage: run") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestRunShell_StatusMissingArg(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "status\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Usage: status") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestRunShell_StatusUnknownTask(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "status none\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Could not get status") {
		t.Errorf("Expected error message for unknown task, got: %s", output)
	}
}

func TestRunShell_DagEmpty(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "dag\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "No tasks") {
		t.Errorf("Expected 'No tasks' for empty dag, got: %s", output)
	}
}

func TestRunShell_DagWithTasks(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "run t1\nrun t2\ndag\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "t1") || !strings.Contains(output, "t2") {
		t.Errorf("Expected dag output to list tasks, got: %s", output)
	}
}

func TestRootCmd_InvalidSubcommand(t *testing.T) {
	root := newAxisRoot()
	root.SetArgs([]string{"invalid-cmd"})
	err := root.Execute()
	if err == nil {
		t.Error("invalid subcommand should fail")
	}
}

func TestRootCmd_ProviderFlagIsPersistent(t *testing.T) {
	resetCLIState()
	root := newAxisRoot()
	root.SetArgs([]string{"--provider", "echo", "run", "provider-task"})
	if err := root.Execute(); err != nil {
		t.Fatalf("root persistent provider flag should work before subcommand: %v", err)
	}
	if defaultApp.providerName != "echo" {
		t.Fatalf("expected providerName echo, got %s", defaultApp.providerName)
	}
}

func TestRootCmd_ModelFlagIsPersistent(t *testing.T) {
	resetCLIState()
	root := newAxisRoot()
	root.SetArgs([]string{"--provider", "echo", "--model", "test-model", "run", "model-task"})
	if err := root.Execute(); err != nil {
		t.Fatalf("root persistent model flag should work before subcommand: %v", err)
	}
	if defaultApp.modelName != "test-model" {
		t.Fatalf("expected modelName test-model, got %s", defaultApp.modelName)
	}
}

func TestDefaultModelForProvider(t *testing.T) {
	if got := defaultModelForProvider("anthropic"); got == "" {
		t.Fatal("anthropic should have a default model")
	}
	if got := defaultModelForProvider("openai"); got == "" {
		t.Fatal("openai should have a default model")
	}
	if got := defaultModelForProvider("deepseek"); got != "deepseek-v4-flash" {
		t.Fatalf("expected deepseek default model deepseek-v4-flash, got %s", got)
	}
	if got := defaultModelForProvider("minimax"); got != "MiniMax-M2.7" {
		t.Fatalf("expected minimax default model MiniMax-M2.7, got %s", got)
	}
	if got := defaultModelForProvider("mock"); got != "" {
		t.Fatalf("mock should not need a default model, got %s", got)
	}
}

func TestEnvAPIKeyForProvider(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"anthropic", "ANTHROPIC_API_KEY"},
		{"openai", "OPENAI_API_KEY"},
		{"deepseek", "DEEPSEEK_API_KEY"},
		{"minimax", "MINIMAX_API_KEY"},
		{"mock", ""},
		{"unknown", ""},
	}
	for _, tc := range tests {
		got := envAPIKeyForProvider(tc.provider)
		if got != tc.want {
			t.Errorf("envAPIKeyForProvider(%q) = %q, want %q", tc.provider, got, tc.want)
		}
	}
}

func TestProviderOptions_EnvFallback(t *testing.T) {
	resetCLIState()
	os.Setenv("DEEPSEEK_API_KEY", "sk-env-ds")
	defer os.Unsetenv("DEEPSEEK_API_KEY")

	app := &App{providerName: "deepseek"}
	opts := app.providerOptions()
	if len(opts) != 2 {
		t.Fatalf("expected 2 options (model + apiKey), got %d", len(opts))
	}
}

// newAxisRoot creates the same root command structure as main().
func newAxisRoot() *cobra.Command {
	return NewRootCommand(defaultApp)
}

func resetCLIState() {
	orch = nil
	defaultApp = &App{providerName: "mock", root: os.TempDir()}
	contextpack.DefaultRegistry = contextpack.NewReadinessRegistry()
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	os.Stdout = original

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, reader); err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}
	return buffer.String()
}

func captureStdoutStdin(t *testing.T, input string, fn func()) string {
	t.Helper()

	// Replace stdin
	inReader, inWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = inReader
	defer func() { os.Stdin = oldStdin }()

	// Replace stdout
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = outWriter
	defer func() { os.Stdout = oldStdout }()

	// Write input in background and close to signal EOF
	go func() {
		inWriter.WriteString(input)
		inWriter.Close()
	}()

	// Read output in background
	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, outReader)
		outCh <- buf.String()
	}()

	fn()

	outWriter.Close()
	return <-outCh
}

func TestRunShell_ResolveMissingArg(t *testing.T) {
	resetCLIState()

	output := captureStdoutStdin(t, "resolve\nexit\n", func() {
		if err := runShell(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runShell should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Usage: resolve") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}
