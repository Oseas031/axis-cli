package contextpack

import (
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestExecutionContextConsumerSummarizeReadyTask(t *testing.T) {
	registry := NewReadinessRegistry()
	task := taskForExecutionSummary("t-ready", "fix provider config")
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
	summary := NewExecutionContextConsumer(registry).Summarize(task)
	if summary.Status != PreflightStatusReady {
		t.Fatalf("expected ready summary, got %+v", summary)
	}
	if summary.ConsumptionMode != ConsumptionModeSummary {
		t.Fatalf("expected summary consumption mode, got %+v", summary)
	}
	if summary.PacketCount != len(bundle.Packets) {
		t.Fatalf("expected packet count %d, got %d", len(bundle.Packets), summary.PacketCount)
	}
	if len(summary.Sources) == 0 {
		t.Fatalf("expected summary sources, got %+v", summary)
	}
	for _, packet := range bundle.Packets {
		for _, source := range summary.Sources {
			if source == packet.Content {
				t.Fatalf("summary should not expose packet content, got %+v", summary.Sources)
			}
		}
	}
	if summary.RequestedSources != nil {
		t.Fatalf("expected nil RequestedSources without metadata, got %+v", summary.RequestedSources)
	}
	if summary.SatisfiedSources != nil {
		t.Fatalf("expected nil SatisfiedSources without metadata, got %+v", summary.SatisfiedSources)
	}
	if summary.MissingSources != nil {
		t.Fatalf("expected nil MissingSources without metadata, got %+v", summary.MissingSources)
	}
}

func TestExecutionContextConsumerSummarizeMissingReadiness(t *testing.T) {
	task := taskForExecutionSummary("t-missing", "fix provider config")
	summary := NewExecutionContextConsumer(NewReadinessRegistry()).Summarize(task)
	if summary.Status != PreflightStatusMissing {
		t.Fatalf("expected missing status, got %+v", summary)
	}
	if summary.ConsumptionMode != ConsumptionModeNone {
		t.Fatalf("expected none consumption mode, got %+v", summary)
	}
	if summary.RequestedSources != nil {
		t.Fatalf("expected nil RequestedSources without metadata, got %+v", summary.RequestedSources)
	}
	if summary.SatisfiedSources != nil {
		t.Fatalf("expected nil SatisfiedSources without metadata, got %+v", summary.SatisfiedSources)
	}
	if summary.MissingSources != nil {
		t.Fatalf("expected nil MissingSources without metadata, got %+v", summary.MissingSources)
	}
}

func TestExecutionContextConsumerSummarizeUntraceableReadiness(t *testing.T) {
	task := taskForExecutionSummary("t-untraceable", "fix provider config")
	task.Metadata = map[string]string{
		MetadataBundleID:     "ctx-missing",
		MetadataPacketCount:  "1",
		MetadataSourceDigest: "abc",
	}
	summary := NewExecutionContextConsumer(NewReadinessRegistry()).Summarize(task)
	if summary.Status != PreflightStatusUntraceable {
		t.Fatalf("expected untraceable status, got %+v", summary)
	}
	if summary.ConsumptionMode != ConsumptionModeNone {
		t.Fatalf("expected none consumption mode, got %+v", summary)
	}
	if summary.RequestedSources != nil {
		t.Fatalf("expected nil RequestedSources without metadata, got %+v", summary.RequestedSources)
	}
	if summary.SatisfiedSources != nil {
		t.Fatalf("expected nil SatisfiedSources without metadata, got %+v", summary.SatisfiedSources)
	}
	if summary.MissingSources != nil {
		t.Fatalf("expected nil MissingSources without metadata, got %+v", summary.MissingSources)
	}
}

func TestExecutionContextConsumerSummarizeUntraceableWithRequests(t *testing.T) {
	task := taskForExecutionSummary("t-untraceable-req", "fix provider config")
	task.Metadata = map[string]string{
		MetadataBundleID:         "ctx-missing",
		MetadataPacketCount:      "1",
		MetadataSourceDigest:     "abc",
		MetadataRequestedSources: "docs/specs/missing.md",
	}
	summary := NewExecutionContextConsumer(NewReadinessRegistry()).Summarize(task)
	if summary.Status != PreflightStatusUntraceable {
		t.Fatalf("expected untraceable status, got %+v", summary)
	}
	if summary.ConsumptionMode != ConsumptionModeNone {
		t.Fatalf("expected none consumption mode, got %+v", summary)
	}
	if len(summary.RequestedSources) != 1 {
		t.Fatalf("expected 1 requested source, got %+v", summary.RequestedSources)
	}
	if len(summary.MissingSources) != 1 {
		t.Fatalf("expected 1 missing source, got %+v", summary.MissingSources)
	}
	if len(summary.SatisfiedSources) != 0 {
		t.Fatalf("expected 0 satisfied sources, got %+v", summary.SatisfiedSources)
	}
}

func TestExecutionContextConsumerSummarizeRequestedSources(t *testing.T) {
	registry := NewReadinessRegistry()
	task := taskForExecutionSummary("t-requested", "fix provider config")
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
	// Request one source that exists and one that does not
	task.Metadata[MetadataRequestedSources] = bundle.Packets[0].Source + ", missing-source.md"

	summary := NewExecutionContextConsumer(registry).Summarize(task)
	if summary.Status != PreflightStatusReady {
		t.Fatalf("expected ready summary, got %+v", summary)
	}
	if len(summary.RequestedSources) != 2 {
		t.Fatalf("expected 2 requested sources, got %+v", summary.RequestedSources)
	}
	if len(summary.SatisfiedSources) != 1 {
		t.Fatalf("expected 1 satisfied source, got %+v", summary.SatisfiedSources)
	}
	if summary.SatisfiedSources[0] != bundle.Packets[0].Source {
		t.Fatalf("expected satisfied source %s, got %s", bundle.Packets[0].Source, summary.SatisfiedSources[0])
	}
	if len(summary.MissingSources) != 1 {
		t.Fatalf("expected 1 missing source, got %+v", summary.MissingSources)
	}
	if summary.MissingSources[0] != "missing-source.md" {
		t.Fatalf("expected missing source missing-source.md, got %s", summary.MissingSources[0])
	}
}

func TestExecutionContextConsumerSummarizeMissingReadinessWithRequests(t *testing.T) {
	task := taskForExecutionSummary("t-missing-req", "fix provider config")
	task.Metadata = map[string]string{
		MetadataRequestedSources: "docs/specs/missing.md",
	}
	summary := NewExecutionContextConsumer(NewReadinessRegistry()).Summarize(task)
	if summary.Status != PreflightStatusMissing {
		t.Fatalf("expected missing status, got %+v", summary)
	}
	if len(summary.RequestedSources) != 1 {
		t.Fatalf("expected 1 requested source, got %+v", summary.RequestedSources)
	}
	if len(summary.MissingSources) != 1 {
		t.Fatalf("expected 1 missing source, got %+v", summary.MissingSources)
	}
	if len(summary.SatisfiedSources) != 0 {
		t.Fatalf("expected 0 satisfied sources, got %+v", summary.SatisfiedSources)
	}
}

func taskForExecutionSummary(taskID string, goal string) *types.AgentTask {
	return &types.AgentTask{TaskID: taskID, ContractID: "default", Input: map[string]any{"goal": goal}, CreatedAt: time.Now()}
}
