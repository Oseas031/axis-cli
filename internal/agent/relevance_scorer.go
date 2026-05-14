package agent

import "strings"

// RelevanceScorer scores how relevant a context chunk is to the current task.
type RelevanceScorer interface {
	Score(chunk string, taskIntent string) float64
}

// KeywordRelevanceScorer scores by keyword overlap between chunk and task intent.
type KeywordRelevanceScorer struct{}

// Score returns the ratio of task intent keywords found in the chunk.
func (s *KeywordRelevanceScorer) Score(chunk string, taskIntent string) float64 {
	keywords := strings.Fields(strings.ToLower(taskIntent))
	if len(keywords) == 0 {
		return 0
	}
	lower := strings.ToLower(chunk)
	matched := 0
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			matched++
		}
	}
	return float64(matched) / float64(len(keywords))
}
