package contextpack

import "testing"

func TestNewExecutionContextSummaryCopiesSources(t *testing.T) {
	sources := []string{"docs/specs/model-provider/design.md"}
	summary := NewExecutionContextSummary(PreflightResult{Status: PreflightStatusReady, BundleID: "ctx-1", PacketCount: 1, SourceDigest: "abc"}, ConsumptionModeSummary, sources)
	sources[0] = "mutated"
	if summary.Sources[0] != "docs/specs/model-provider/design.md" {
		t.Fatalf("expected copied sources, got %+v", summary.Sources)
	}
	if summary.ConsumptionMode != ConsumptionModeSummary {
		t.Fatalf("expected summary mode, got %q", summary.ConsumptionMode)
	}
}

func TestExecutionContextSummaryModes(t *testing.T) {
	modes := []ConsumptionMode{ConsumptionModeNone, ConsumptionModeObserved, ConsumptionModeSummary, ConsumptionModePromptAugmented}
	for _, mode := range modes {
		if mode == "" {
			t.Fatal("consumption mode should not be empty")
		}
	}
}
