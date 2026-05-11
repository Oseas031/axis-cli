package contextpack

import (
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestAssemblerSelectsNaturalLanguageSchedulingSpec(t *testing.T) {
	bundle, err := NewAssembler().Assemble(taskWithGoal("use axis ask to convert a prompt into a task"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	assertSelected(t, bundle, "spec:natural-language-scheduling")
}

func TestAssemblerSelectsProviderSpec(t *testing.T) {
	bundle, err := NewAssembler().Assemble(taskWithGoal("fix minimax provider profile model config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	assertSelected(t, bundle, "spec:model-provider")
}

func TestAssemblerSelectsInteractiveShellSpec(t *testing.T) {
	bundle, err := NewAssembler().Assemble(taskWithGoal("normalize shell command behavior"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	assertSelected(t, bundle, "spec:interactive-shell")
}

func TestAssemblerBudgetExclusionsAreVisible(t *testing.T) {
	budget := ContextBudget{MaxPackets: 1, MaxBytes: 4096}
	bundle, err := NewAssembler(WithBudget(budget)).Assemble(taskWithGoal("ask prompt provider model shell context scheduler axis-up"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	if len(bundle.Packets) != 1 {
		t.Fatalf("expected one selected packet, got %d", len(bundle.Packets))
	}
	if !bundle.Budget.Truncated {
		t.Fatal("expected truncated budget")
	}
	if len(bundle.Trace.Excluded) == 0 {
		t.Fatal("expected excluded trace items")
	}
}

func TestAssemblerRejectsEmptyGoal(t *testing.T) {
	_, err := NewAssembler().Assemble(&types.AgentTask{TaskID: "t1", ContractID: "default", Input: map[string]any{}, CreatedAt: time.Now()})
	if err == nil {
		t.Fatal("expected empty goal to fail")
	}
}

func assertSelected(t *testing.T, bundle *ContextBundle, packetID string) {
	t.Helper()
	for _, packet := range bundle.Packets {
		if packet.ID == packetID {
			if packet.Source == "" || packet.Reason == "" {
				t.Fatalf("selected packet should include source and reason: %+v", packet)
			}
			return
		}
	}
	t.Fatalf("expected selected packet %s, got %+v", packetID, bundle.Packets)
}

func TestAssembler_HybridMode(t *testing.T) {
	chunks := []DocumentChunk{
		{Source: "docs/specs/model-provider/requirements.md", Content: "model provider config guide deepseek minimax", DocType: "doc"},
		{Source: "internal/model/provider/openai.go", Content: "package provider openai chat completions", DocType: "code"},
	}
	idx := &TFIDFIndex{}
	idx.Build(chunks)

	bundle, err := NewAssembler(WithIndex(idx)).Assemble(taskWithGoal("fix minimax provider profile model config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}

	// Should have rule packet
	assertSelected(t, bundle, "spec:model-provider")

	// Should have retrieval packet
	foundRetrieval := false
	for _, p := range bundle.Packets {
		if strings.HasPrefix(p.ID, "retrieval:") {
			foundRetrieval = true
			break
		}
	}
	if !foundRetrieval {
		t.Errorf("expected at least one retrieval packet in bundle, got %+v", bundle.Packets)
	}

	// Trace should note hybrid mode
	hasHybridNote := false
	for _, note := range bundle.Trace.Notes {
		if strings.Contains(note, "hybrid mode") {
			hasHybridNote = true
			break
		}
	}
	if !hasHybridNote {
		t.Errorf("expected hybrid mode trace note, got %v", bundle.Trace.Notes)
	}
}

func TestAssembler_RuleOnlyFallback(t *testing.T) {
	bundle, err := NewAssembler().Assemble(taskWithGoal("fix minimax provider profile model config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	assertSelected(t, bundle, "spec:model-provider")

	hasFallbackNote := false
	for _, note := range bundle.Trace.Notes {
		if strings.Contains(note, "rule-only fallback") {
			hasFallbackNote = true
			break
		}
	}
	if !hasFallbackNote {
		t.Errorf("expected fallback trace note, got %v", bundle.Trace.Notes)
	}
}

func TestAssembler_RetrievalBoostsRulePacket(t *testing.T) {
	chunks := []DocumentChunk{
		{Source: "docs/specs/model-provider/requirements.md", Content: "model provider config guide deepseek minimax", DocType: "doc"},
	}
	idx := &TFIDFIndex{}
	idx.Build(chunks)

	bundle, err := NewAssembler(WithIndex(idx)).Assemble(taskWithGoal("fix minimax provider profile model config"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}

	// The rule packet for model-provider should be present; its relevance may have been boosted.
	// We just verify it's selected and no duplicate retrieval packet for the same source exists.
	assertSelected(t, bundle, "spec:model-provider")

	dupeCount := 0
	for _, p := range bundle.Packets {
		if strings.Contains(p.Source, "model-provider") {
			dupeCount++
		}
	}
	if dupeCount > 1 {
		t.Errorf("expected deduplication to prevent duplicate model-provider packets, got %d", dupeCount)
	}
}

func TestAssemblerBudgetTruncatesRetrievalContent(t *testing.T) {
	longContent := "deepseek provider profile model config. " + strings.Repeat("This paragraph extends the content beyond any reasonable budget limit. ", 10)
	chunks := []DocumentChunk{
		{Source: "internal/provider/deepseek.go", Content: longContent, DocType: "code"},
	}
	idx := &TFIDFIndex{}
	idx.Build(chunks)

	budget := ContextBudget{MaxPackets: 2, MaxBytes: 350}
	bundle, err := NewAssembler(WithBudget(budget), WithIndex(idx)).Assemble(taskWithGoal("deepseek provider profile model config update"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}

	var retrieval *ContextPacket
	for i := range bundle.Packets {
		if strings.HasPrefix(bundle.Packets[i].ID, "retrieval:") {
			retrieval = &bundle.Packets[i]
			break
		}
	}
	if retrieval == nil {
		t.Fatal("expected a retrieval packet in bundle")
	}
	if !retrieval.IsPartial {
		t.Fatalf("expected retrieval packet IsPartial=true, got %+v", retrieval)
	}
	if retrieval.TruncatedAt == 0 {
		t.Fatalf("expected TruncatedAt > 0, got %d", retrieval.TruncatedAt)
	}
	if bundle.Budget.UsedBytes > bundle.Budget.MaxBytes {
		t.Fatalf("used bytes %d exceeds max bytes %d", bundle.Budget.UsedBytes, bundle.Budget.MaxBytes)
	}

	foundTrace := false
	for _, item := range bundle.Trace.Selected {
		if strings.HasPrefix(item.PacketID, "retrieval:") && strings.Contains(item.Reason, "truncated") {
			foundTrace = true
			break
		}
	}
	if !foundTrace {
		t.Fatalf("expected trace to note truncation, got selected: %+v", bundle.Trace.Selected)
	}
}

func TestAssemblerBudgetDropsPacketWhenFixedOverheadExceedsBudget(t *testing.T) {
	longContent := "deepseek provider profile model config. " + strings.Repeat("filler text here. ", 20)
	chunks := []DocumentChunk{
		{Source: "internal/provider/deepseek.go", Content: longContent, DocType: "code"},
	}
	idx := &TFIDFIndex{}
	idx.Build(chunks)

	budget := ContextBudget{MaxPackets: 2, MaxBytes: 250}
	bundle, err := NewAssembler(WithBudget(budget), WithIndex(idx)).Assemble(taskWithGoal("deepseek provider profile model config update"))
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}

	// model-provider packet ~187 bytes; remaining ~63 bytes < retrieval fixed ~107 bytes.
	for _, p := range bundle.Packets {
		if strings.HasPrefix(p.ID, "retrieval:") {
			t.Fatalf("expected retrieval packet to be excluded when fixed overhead exceeds budget, got %+v", p)
		}
	}
	if !bundle.Budget.Truncated {
		t.Fatal("expected budget to be marked truncated")
	}
	foundExcluded := false
	for _, item := range bundle.Trace.Excluded {
		if strings.HasPrefix(item.PacketID, "retrieval:") {
			foundExcluded = true
			break
		}
	}
	if !foundExcluded {
		t.Fatalf("expected retrieval packet in excluded trace, got: %+v", bundle.Trace.Excluded)
	}
}

func taskWithGoal(goal string) *types.AgentTask {
	return &types.AgentTask{TaskID: "t1", ContractID: "default", Input: map[string]any{"goal": goal}, CreatedAt: time.Now()}
}
