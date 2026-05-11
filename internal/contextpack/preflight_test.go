package contextpack

import (
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestPreflightReady(t *testing.T) {
	registry := NewReadinessRegistry()
	task := &types.AgentTask{TaskID: "t1", ContractID: "default", Input: map[string]any{"goal": "fix provider config"}, CreatedAt: time.Now()}
	bundle, err := NewAssembler().Assemble(task)
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	artifact, err := registry.Register(bundle)
	if err != nil {
		t.Fatalf("register should succeed: %v", err)
	}
	if err := AttachReadinessMetadata(task, artifact); err != nil {
		t.Fatalf("attach readiness metadata should succeed: %v", err)
	}
	result := Preflight(task, registry)
	if result.Status != PreflightStatusReady {
		t.Fatalf("expected ready preflight, got %+v", result)
	}
	if result.BundleID != artifact.BundleID {
		t.Fatalf("expected bundle id %s, got %s", artifact.BundleID, result.BundleID)
	}
}

func TestPreflightMissingBundleID(t *testing.T) {
	task := &types.AgentTask{TaskID: "t1", Metadata: map[string]string{}}
	result := Preflight(task, NewReadinessRegistry())
	if result.Status != PreflightStatusMissing {
		t.Fatalf("expected missing preflight, got %+v", result)
	}
}

func TestPreflightUntraceableMissingRegistryRecord(t *testing.T) {
	task := &types.AgentTask{TaskID: "t1", Metadata: map[string]string{MetadataBundleID: "ctx-missing", MetadataPacketCount: "1", MetadataSourceDigest: "abc"}}
	result := Preflight(task, NewReadinessRegistry())
	if result.Status != PreflightStatusUntraceable {
		t.Fatalf("expected untraceable preflight, got %+v", result)
	}
}

func TestPreflightNegativePacketCount(t *testing.T) {
	registry := NewReadinessRegistry()
	task := &types.AgentTask{TaskID: "t1", ContractID: "default", Input: map[string]any{"goal": "fix provider config"}, CreatedAt: time.Now()}
	bundle, err := NewAssembler().Assemble(task)
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	artifact, err := registry.Register(bundle)
	if err != nil {
		t.Fatalf("register should succeed: %v", err)
	}
	if err := AttachReadinessMetadata(task, artifact); err != nil {
		t.Fatalf("attach metadata should succeed: %v", err)
	}
	// Tamper with packet count to negative
	task.Metadata[MetadataPacketCount] = "-3"
	result := Preflight(task, registry)
	if result.Status != PreflightStatusMissing {
		t.Fatalf("expected missing preflight for negative packet count, got %+v", result)
	}
}
