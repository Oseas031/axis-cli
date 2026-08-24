package contextpack

import (
	"testing"
	"time"
)

func TestRanker_Rank(t *testing.T) {
	ranker := NewRanker()

	candidates := []*Candidate{
		{
			ID:        "1",
			Content:   "context assembly specification",
			Source:    "docs/context.md",
			Relevance: 0.9,
			Timestamp: time.Now(),
		},
		{
			ID:        "2",
			Content:   "model provider configuration",
			Source:    "docs/provider.md",
			Relevance: 0.8,
			Timestamp: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        "3",
			Content:   "scheduler dag architecture",
			Source:    "docs/scheduler.md",
			Relevance: 0.85,
			Timestamp: time.Now().Add(-48 * time.Hour),
		},
	}

	ranked := ranker.Rank(candidates, "context assembly")

	if len(ranked) != 3 {
		t.Errorf("Rank() returned %d candidates, want 3", len(ranked))
	}

	if ranked[0].Score < ranked[1].Score {
		t.Error("Rank() did not sort by score descending")
	}
}

func TestFreshnessStrategy_Score(t *testing.T) {
	strategy := &FreshnessStrategy{}

	fresh := &Candidate{
		Timestamp: time.Now(),
	}

	old := &Candidate{
		Timestamp: time.Now().Add(-72 * time.Hour),
	}

	freshScore := strategy.Score(fresh, "")
	oldScore := strategy.Score(old, "")

	if freshScore <= oldScore {
		t.Error("FreshnessStrategy: fresh candidate should have higher score")
	}
}

func TestAdaptiveRanker_Update(t *testing.T) {
	ranker := NewAdaptiveRanker()

	initialWeights := make([]float64, len(ranker.Weights))
	copy(initialWeights, ranker.Weights)

	ranker.Update("relevance", 0.9)
	ranker.Update("importance", 0.7)

	if ranker.Weights[0] == initialWeights[0] {
		t.Error("AdaptiveRanker.Update() did not change weights")
	}
}
