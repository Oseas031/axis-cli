package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func withTempCwd(t *testing.T) func() {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return func() { _ = os.Chdir(old) }
}

func execCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	rootCmd := NewRootCommand(&App{providerName: "mock"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestMemoryRetainList(t *testing.T) {
	defer withTempCwd(t)()

	out, err := execCLI(t, "memory", "retain", "ctx-test-1", "--reason", "test run")
	if err != nil {
		t.Fatalf("retain: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ctx-test-1") {
		t.Fatalf("retain output missing bundle id: %s", out)
	}

	out, err = execCLI(t, "memory", "list")
	if err != nil {
		t.Fatalf("list: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ctx-test-1") {
		t.Fatalf("list output missing bundle id: %s", out)
	}
}

func TestMemoryRelease(t *testing.T) {
	defer withTempCwd(t)()

	_, _ = execCLI(t, "memory", "retain", "ctx-release", "--reason", "tmp")
	out, err := execCLI(t, "memory", "release", "ctx-release")
	if err != nil {
		t.Fatalf("release: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Released") {
		t.Fatalf("expected Released in output, got: %s", out)
	}
}

func TestMemoryCompact(t *testing.T) {
	defer withTempCwd(t)()

	_, _ = execCLI(t, "memory", "retain", "ctx-c", "--reason", "r")
	out, err := execCLI(t, "memory", "compact")
	if err != nil {
		t.Fatalf("compact: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Compacted") {
		t.Fatalf("expected Compacted in output, got: %s", out)
	}
}

func TestMemoryListJSON(t *testing.T) {
	defer withTempCwd(t)()

	_, _ = execCLI(t, "memory", "retain", "ctx-json", "--reason", "json-test")
	out, err := execCLI(t, "memory", "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v\n%s", err, out)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if _, ok := decoded["items"]; !ok {
		t.Fatalf("JSON output missing 'items': %s", out)
	}
}

func TestMemoryRetain_RequiresReason(t *testing.T) {
	defer withTempCwd(t)()

	_, err := execCLI(t, "memory", "retain", "ctx-noreason")
	if err == nil {
		t.Fatal("expected error when --reason is missing")
	}
}
