package contextpack

import (
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestNewReadinessArtifactIsReproducible(t *testing.T) {
	bundle, err := NewAssembler().Assemble(taskWithGoal("fix provider config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	first, err := NewReadinessArtifact(bundle)
	if err != nil {
		t.Fatalf("artifact should be created: %v", err)
	}
	second, err := NewReadinessArtifact(bundle)
	if err != nil {
		t.Fatalf("artifact should be created again: %v", err)
	}
	if first.BundleID == "" || first.BundleID != second.BundleID {
		t.Fatalf("expected reproducible bundle id, got %q and %q", first.BundleID, second.BundleID)
	}
	if first.AssemblyMode != "rule_based" {
		t.Fatalf("expected rule_based assembly mode, got %q", first.AssemblyMode)
	}
	if first.PacketCount != len(bundle.Packets) {
		t.Fatalf("expected packet count %d, got %d", len(bundle.Packets), first.PacketCount)
	}
	if first.SourceDigest == "" {
		t.Fatal("expected source digest")
	}
}

func TestAttachReadinessMetadata(t *testing.T) {
	bundle, err := NewAssembler().Assemble(taskWithGoal("fix provider config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	artifact, err := NewReadinessArtifact(bundle)
	if err != nil {
		t.Fatalf("artifact should be created: %v", err)
	}
	task := &types.AgentTask{TaskID: "t1", ContractID: "default", Input: map[string]any{"goal": "fix provider config"}, CreatedAt: time.Now()}
	if err := AttachReadinessMetadata(task, artifact); err != nil {
		t.Fatalf("attach metadata should succeed: %v", err)
	}
	if task.Metadata[MetadataBundleID] != artifact.BundleID {
		t.Fatalf("expected bundle id metadata, got %+v", task.Metadata)
	}
	if task.Metadata[MetadataAssemblyMode] != "rule_based" {
		t.Fatalf("expected assembly mode metadata, got %+v", task.Metadata)
	}
	if _, ok := task.Metadata["context.bundle"]; ok {
		t.Fatalf("should not attach full bundle metadata, got %+v", task.Metadata)
	}
}
