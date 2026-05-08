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
