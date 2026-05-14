package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestProviderCommand_AddUseStatusDoesNotPrintSecret(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	dir := t.TempDir()
	if err := os.MkdirAll(dir+"/.axis", 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer os.Chdir(cwd)

	// Reset defaultApp to use tempdir as root (provider_cmd uses defaultApp.resolvedRoot())
	defaultApp = &App{providerName: "mock", root: dir}

	root := NewRootCommand(&App{providerName: "mock"})
	root.SetArgs([]string{"provider", "add", "openai-local", "--type", "openai", "--api-key", "secret-key", "--model", "gpt-4o-mini", "--base-url", "https://api.openai.com"})
	if err := root.Execute(); err != nil {
		t.Fatalf("provider add failed: %v", err)
	}

	root = NewRootCommand(&App{providerName: "mock"})
	root.SetArgs([]string{"provider", "use", "openai-local"})
	if err := root.Execute(); err != nil {
		t.Fatalf("provider use failed: %v", err)
	}

	var out bytes.Buffer
	root = NewRootCommand(&App{providerName: "mock"})
	root.SetOut(&out)
	root.SetArgs([]string{"provider", "status"})
	if err := root.Execute(); err != nil {
		t.Fatalf("provider status failed: %v", err)
	}
	status := out.String()
	if !strings.Contains(status, "active_profile: openai-local") {
		t.Fatalf("expected active profile in status, got %q", status)
	}
	if strings.Contains(status, "secret-key") {
		t.Fatalf("status must not print API keys: %q", status)
	}

	data, err := os.ReadFile(".axis/providers.json")
	if err != nil {
		t.Fatalf("expected project-local providers.json: %v", err)
	}
	if !strings.Contains(string(data), "secret-key") {
		t.Fatal("expected API key to be stored in project-local providers.json")
	}
}
