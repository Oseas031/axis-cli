package control

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRuntimeLocatorSaveAndLoad(t *testing.T) {
	root := t.TempDir()
	locator := NewRuntimeLocator(root)
	startedAt := time.Date(2026, 5, 10, 18, 0, 0, 0, time.UTC)
	record := RuntimeRecord{
		PID:         1234,
		Protocol:    "http",
		Address:     "127.0.0.1:4567",
		ProjectRoot: root,
		StartedAt:   startedAt,
	}

	if err := locator.Save(record); err != nil {
		t.Fatalf("save runtime locator: %v", err)
	}

	path := filepath.Join(root, ".axis", "runtime.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected runtime locator at %s: %v", path, err)
	}

	got, err := locator.Load()
	if err != nil {
		t.Fatalf("load runtime locator: %v", err)
	}
	if got.PID != 1234 || got.Protocol != "http" || got.Address != "127.0.0.1:4567" || got.ProjectRoot != root {
		t.Fatalf("unexpected runtime record: %#v", got)
	}
	if !got.StartedAt.Equal(startedAt) {
		t.Fatalf("expected started_at %s, got %s", startedAt, got.StartedAt)
	}
}

func TestRuntimeLocatorLoadMissing(t *testing.T) {
	_, err := NewRuntimeLocator(t.TempDir()).Load()
	if !errors.Is(err, ErrRuntimeLocatorNotFound) {
		t.Fatalf("expected ErrRuntimeLocatorNotFound, got %v", err)
	}
}

func TestRuntimeLocatorLoadMalformed(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".axis", "runtime.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("write malformed locator: %v", err)
	}

	_, err := NewRuntimeLocator(root).Load()
	if err == nil {
		t.Fatal("expected malformed locator error")
	}
	if errors.Is(err, ErrRuntimeLocatorNotFound) {
		t.Fatalf("malformed locator should not be reported as missing: %v", err)
	}
}

func TestRuntimeLocatorDoesNotWriteSecrets(t *testing.T) {
	root := t.TempDir()
	locator := NewRuntimeLocator(root)
	record := RuntimeRecord{
		PID:         1234,
		Protocol:    "http",
		Address:     "127.0.0.1:4567",
		ProjectRoot: root,
		StartedAt:   time.Now().UTC(),
	}
	if err := locator.Save(record); err != nil {
		t.Fatalf("save runtime locator: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".axis", "runtime.json"))
	if err != nil {
		t.Fatalf("read runtime locator: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal runtime locator: %v", err)
	}
	for _, forbidden := range []string{"api_key", "token", "secret", "bearer"} {
		if _, ok := raw[forbidden]; ok {
			t.Fatalf("runtime locator must not contain key %q: %#v", forbidden, raw)
		}
	}
}

func TestRuntimeLocatorDelete(t *testing.T) {
	root := t.TempDir()
	locator := NewRuntimeLocator(root)
	if err := locator.Save(RuntimeRecord{PID: 1234, Protocol: "http", Address: "127.0.0.1:4567", ProjectRoot: root, StartedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("save runtime locator: %v", err)
	}
	if err := locator.Delete(); err != nil {
		t.Fatalf("delete runtime locator: %v", err)
	}
	_, err := locator.Load()
	if !errors.Is(err, ErrRuntimeLocatorNotFound) {
		t.Fatalf("expected missing locator after delete, got %v", err)
	}
}
