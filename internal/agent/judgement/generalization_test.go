package judgement

import "testing"

func TestRecordAcrossMultipleTaskTypes(t *testing.T) {
	gt := NewGeneralizationTracker()
	gt.Record("syntax-check", "build", true)
	gt.Record("syntax-check", "deploy", false)
	gt.Record("syntax-check", "test", true)

	score := gt.GetScore("syntax-check")
	if score == nil {
		t.Fatal("expected score, got nil")
	}
	if len(score.TaskTypes) != 3 {
		t.Fatalf("expected 3 task types, got %d", len(score.TaskTypes))
	}
	if score.TaskTypes["build"].PassRate != 1.0 {
		t.Errorf("expected build pass rate 1.0, got %f", score.TaskTypes["build"].PassRate)
	}
	if score.TaskTypes["deploy"].PassRate != 0.0 {
		t.Errorf("expected deploy pass rate 0.0, got %f", score.TaskTypes["deploy"].PassRate)
	}
}

func TestIsGeneralizedFalseWhenTooFewTypes(t *testing.T) {
	gt := NewGeneralizationTracker()
	gt.Record("syntax-check", "build", true)

	if gt.IsGeneralized("syntax-check", 3, 0.8) {
		t.Error("expected false when fewer than minTypes")
	}
}

func TestIsGeneralizedFalseWhenLowPassRate(t *testing.T) {
	gt := NewGeneralizationTracker()
	gt.Record("syntax-check", "build", true)
	gt.Record("syntax-check", "deploy", true)
	gt.Record("syntax-check", "test", false) // 0% pass rate for "test"

	if gt.IsGeneralized("syntax-check", 3, 0.8) {
		t.Error("expected false when one type has low pass rate")
	}
}

func TestIsGeneralizedTrueWhenAllConditionsMet(t *testing.T) {
	gt := NewGeneralizationTracker()
	for _, tt := range []string{"build", "deploy", "test"} {
		gt.Record("syntax-check", tt, true)
	}

	if !gt.IsGeneralized("syntax-check", 3, 0.8) {
		t.Error("expected true when all conditions met")
	}
}
