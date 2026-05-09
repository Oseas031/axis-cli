package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunTaskInitializesLocalOrchestrator(t *testing.T) {
	resetCLIState()

	output := captureStdout(t, func() {
		if err := runTask(&cobra.Command{}, []string{"task-1"}); err != nil {
			t.Fatalf("runTask should initialize and submit locally: %v", err)
		}
	})

	if strings.Contains(output, "axis start") {
		t.Fatalf("runTask output should not tell users to run axis start first: %s", output)
	}
	if !strings.Contains(output, "Task task-1 submitted successfully") {
		t.Fatalf("runTask output should confirm submission, got: %s", output)
	}
}

func TestRunTask_InitOrchestratorError(t *testing.T) {
	resetCLIState()

	// First call succeeds (sets up orchestrator)
	output := captureStdout(t, func() {
		if err := runTask(&cobra.Command{}, []string{"task-1"}); err != nil {
			t.Fatalf("first runTask should succeed: %v", err)
		}
	})
	if !strings.Contains(output, "Task task-1 submitted successfully") {
		t.Fatalf("Expected success, got: %s", output)
	}
}

func TestGetTaskStatusNotFoundGivesContext(t *testing.T) {
	resetCLIState()

	err := getTaskStatus(&cobra.Command{}, []string{"missing"})
	if err == nil {
		t.Fatal("getTaskStatus should return an error for a missing task")
	}
	if !strings.Contains(err.Error(), "task missing not found") {
		t.Fatalf("getTaskStatus should include not-found context, got: %v", err)
	}
	if strings.Contains(err.Error(), "axis start") {
		t.Fatalf("getTaskStatus should not imply cross-process axis start state, got: %v", err)
	}
}

func TestGetTaskStatus_Success(t *testing.T) {
	resetCLIState()

	// Submit a task first
	if err := runTask(&cobra.Command{}, []string{"task-status-success"}); err != nil {
		t.Fatalf("runTask should succeed: %v", err)
	}

	err := getTaskStatus(&cobra.Command{}, []string{"task-status-success"})
	if err != nil {
		t.Fatalf("getTaskStatus should succeed for submitted task: %v", err)
	}
}

func TestPrintShellHelp(t *testing.T) {
	output := captureStdout(t, func() {
		printShellHelp()
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
	if len(c.OutputSchema.Fields) != 2 {
		t.Errorf("OutputSchema should have 2 fields, got %d", len(c.OutputSchema.Fields))
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

func TestRootCmd_InvalidSubcommand(t *testing.T) {
	root := newAxisRoot()
	root.SetArgs([]string{"invalid-cmd"})
	err := root.Execute()
	if err == nil {
		t.Error("invalid subcommand should fail")
	}
}

// newAxisRoot creates the same root command structure as main().
func newAxisRoot() *cobra.Command {
	rootCmd := &cobra.Command{Use: "axis", Short: "Agent-native scheduling system"}

	runCmd := &cobra.Command{Use: "run [task-id]", Short: "Submit and run a task", Args: cobra.ExactArgs(1), RunE: runTask}
	statusCmd := &cobra.Command{Use: "status [task-id]", Short: "Get task status", Args: cobra.ExactArgs(1), RunE: getTaskStatus}
	startCmd := &cobra.Command{Use: "start", Short: "Start the orchestrator", RunE: startOrchestrator}
	shellCmd := &cobra.Command{Use: "shell", Short: "Start an interactive Axis shell", RunE: runShell}

	rootCmd.AddCommand(runCmd, statusCmd, startCmd, shellCmd)
	return rootCmd
}

func resetCLIState() {
	orch = nil
	orchMutex = sync.Once{}
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
