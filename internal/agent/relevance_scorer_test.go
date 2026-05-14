package agent

import "testing"

func TestKeywordRelevanceScorer(t *testing.T) {
	scorer := &KeywordRelevanceScorer{}

	taskIntent := "check provider config status"
	highChunk := "the provider config is loaded and status is healthy"
	lowChunk := "unrelated documentation about testing frameworks"

	highScore := scorer.Score(highChunk, taskIntent)
	lowScore := scorer.Score(lowChunk, taskIntent)

	if highScore <= lowScore {
		t.Errorf("expected high relevance chunk to score higher: high=%f low=%f", highScore, lowScore)
	}
	if highScore == 0 {
		t.Error("expected non-zero score for high relevance chunk")
	}
}

func TestKeywordRelevanceScorer_EmptyIntent(t *testing.T) {
	scorer := &KeywordRelevanceScorer{}
	if score := scorer.Score("anything", ""); score != 0 {
		t.Errorf("expected 0 for empty intent, got %f", score)
	}
}

func TestKeywordRelevanceScorer_FullMatch(t *testing.T) {
	scorer := &KeywordRelevanceScorer{}
	score := scorer.Score("deploy the service now", "deploy service")
	if score != 1.0 {
		t.Errorf("expected 1.0 for full keyword match, got %f", score)
	}
}
