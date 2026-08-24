package contextpack

import (
	"math"
	"sort"
	"strings"
	"time"
)

// Ranker 多策略排序器
type Ranker struct {
	Strategies []RankingStrategy
	Weights    []float64
}

// RankingStrategy 排序策略接口
type RankingStrategy interface {
	Name() string
	Score(candidate *Candidate, query string) float64
}

// Candidate 候选项
type Candidate struct {
	ID        string
	Content   string
	Source    string
	Relevance float64
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// RankedCandidate 排序后的候选项
type RankedCandidate struct {
	*Candidate
	Score    float64
	Strategy string
	Rank     int
}

// NewRanker 创建排序器
func NewRanker() *Ranker {
	return &Ranker{
		Strategies: []RankingStrategy{
			&RelevanceStrategy{},
			&ImportanceStrategy{},
			&FreshnessStrategy{},
		},
		Weights: []float64{0.5, 0.3, 0.2},
	}
}

// Rank 多策略排序
func (r *Ranker) Rank(candidates []*Candidate, query string) []*RankedCandidate {
	if len(candidates) == 0 {
		return nil
	}

	var ranked []*RankedCandidate

	for _, candidate := range candidates {
		totalScore := 0.0
		strategyName := ""

		for i, strategy := range r.Strategies {
			score := strategy.Score(candidate, query)
			totalScore += score * r.Weights[i]
			strategyName += strategy.Name() + "+"
		}

		ranked = append(ranked, &RankedCandidate{
			Candidate: candidate,
			Score:     totalScore,
			Strategy:  strategyName[:len(strategyName)-1],
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	for i, r := range ranked {
		r.Rank = i + 1
	}

	return ranked
}

// RelevanceStrategy 相关性策略
type RelevanceStrategy struct{}

func (s *RelevanceStrategy) Name() string {
	return "relevance"
}

func (s *RelevanceStrategy) Score(candidate *Candidate, query string) float64 {
	query = strings.ToLower(query)
	content := strings.ToLower(candidate.Content)

	score := 0.0
	queryWords := strings.Fields(query)

	matchedWords := 0
	for _, word := range queryWords {
		if strings.Contains(content, word) {
			matchedWords++
		}
	}

	if len(queryWords) > 0 {
		score = float64(matchedWords) / float64(len(queryWords))
	}

	return 0.5*score + 0.5*candidate.Relevance
}

// ImportanceStrategy 重要性策略
type ImportanceStrategy struct{}

func (s *ImportanceStrategy) Name() string {
	return "importance"
}

func (s *ImportanceStrategy) Score(candidate *Candidate, query string) float64 {
	content := candidate.Content

	entropy := calculateEntropy(content)

	lengthScore := 1.0
	if len(content) < 100 {
		lengthScore = float64(len(content)) / 100.0
	} else if len(content) > 10000 {
		lengthScore = 10000.0 / float64(len(content))
	}

	return 0.6*entropy + 0.4*lengthScore
}

// calculateEntropy 计算信息熵
func calculateEntropy(text string) float64 {
	if len(text) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, r := range text {
		freq[r]++
	}

	entropy := 0.0
	length := float64(len(text))

	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	maxEntropy := math.Log2(float64(len(freq)))
	if maxEntropy > 0 {
		entropy /= maxEntropy
	}

	return entropy
}

// FreshnessStrategy 时效性策略
type FreshnessStrategy struct{}

func (s *FreshnessStrategy) Name() string {
	return "freshness"
}

func (s *FreshnessStrategy) Score(candidate *Candidate, query string) float64 {
	age := time.Since(candidate.Timestamp)

	halfLife := 24.0 * time.Hour
	decay := math.Exp(-age.Hours() / halfLife.Hours() * math.Log(2))

	return decay
}

// AdaptiveRanker 自适应排序器
type AdaptiveRanker struct {
	*Ranker
	History []RankingPerformance
}

// RankingPerformance 排序性能记录
type RankingPerformance struct {
	Strategy  string
	Score     float64
	Timestamp time.Time
}

// NewAdaptiveRanker 创建自适应排序器
func NewAdaptiveRanker() *AdaptiveRanker {
	return &AdaptiveRanker{
		Ranker: NewRanker(),
	}
}

// Update 根据反馈更新策略权重
func (r *AdaptiveRanker) Update(strategy string, score float64) {
	r.History = append(r.History, RankingPerformance{
		Strategy:  strategy,
		Score:     score,
		Timestamp: time.Now(),
	})

	r.recalculateWeights()
}

func (r *AdaptiveRanker) recalculateWeights() {
	strategyScores := make(map[string][]float64)

	for _, perf := range r.History {
		strategyScores[perf.Strategy] = append(strategyScores[perf.Strategy], perf.Score)
	}

	avgScores := make(map[string]float64)
	for strategy, scores := range strategyScores {
		sum := 0.0
		for _, score := range scores {
			sum += score
		}
		avgScores[strategy] = sum / float64(len(scores))
	}

	totalScore := 0.0
	for _, score := range avgScores {
		totalScore += score
	}

	if totalScore > 0 {
		for i, strategy := range r.Strategies {
			if score, ok := avgScores[strategy.Name()]; ok {
				r.Weights[i] = score / totalScore
			}
		}
	}
}
