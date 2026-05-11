package contextpack

import "testing"

func TestReadinessRegistryRegisterAndInspect(t *testing.T) {
	registry := NewReadinessRegistry()
	bundle, err := NewAssembler().Assemble(taskWithGoal("fix provider config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	artifact, err := registry.Register(bundle)
	if err != nil {
		t.Fatalf("register should succeed: %v", err)
	}
	record, err := registry.Inspect(artifact.BundleID)
	if err != nil {
		t.Fatalf("inspect should succeed: %v", err)
	}
	if record.Artifact.BundleID != artifact.BundleID {
		t.Fatalf("expected artifact %s, got %s", artifact.BundleID, record.Artifact.BundleID)
	}
	if record.Bundle.Goal != bundle.Goal {
		t.Fatalf("expected bundle goal %q, got %q", bundle.Goal, record.Bundle.Goal)
	}
	if len(record.Bundle.Packets) == 0 {
		t.Fatal("expected inspected bundle packets")
	}
}

func TestReadinessRegistryInspectMissing(t *testing.T) {
	registry := NewReadinessRegistry()
	if _, err := registry.Inspect("ctx-missing"); err == nil {
		t.Fatal("expected missing readiness record to fail")
	}
}

func TestReadinessRegistryWithStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(tmpDir)
	registry, err := NewReadinessRegistryWithStore(store)
	if err != nil {
		t.Fatalf("NewReadinessRegistryWithStore should succeed: %v", err)
	}

	bundle, err := NewAssembler().Assemble(taskWithGoal("fix provider config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	artifact, err := registry.Register(bundle)
	if err != nil {
		t.Fatalf("register should succeed: %v", err)
	}

	// Simulate a new process by creating a fresh registry from the same store.
	registry2, err := NewReadinessRegistryWithStore(store)
	if err != nil {
		t.Fatalf("second NewReadinessRegistryWithStore should succeed: %v", err)
	}
	record, err := registry2.Inspect(artifact.BundleID)
	if err != nil {
		t.Fatalf("inspect from second process should succeed: %v", err)
	}
	if record.Artifact.BundleID != artifact.BundleID {
		t.Fatalf("expected artifact %s, got %s", artifact.BundleID, record.Artifact.BundleID)
	}
	if record.Bundle.Goal != bundle.Goal {
		t.Fatalf("expected bundle goal %q, got %q", bundle.Goal, record.Bundle.Goal)
	}
}

func TestReadinessRegistryWithStore_ResetDeletesFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(tmpDir)
	registry, err := NewReadinessRegistryWithStore(store)
	if err != nil {
		t.Fatalf("NewReadinessRegistryWithStore should succeed: %v", err)
	}
	bundle, _ := NewAssembler().Assemble(taskWithGoal("test goal"))
	artifact, _ := registry.Register(bundle)

	registry.Reset()

	registry2, _ := NewReadinessRegistryWithStore(store)
	if _, err := registry2.Inspect(artifact.BundleID); err == nil {
		t.Fatal("expected inspect to fail after reset")
	}
}
