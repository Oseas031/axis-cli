// Package rank provides multi-strategy packet ranking.
package rank

import (
	"github.com/axis-cli/axis/internal/contextpack/model"
)

// Strategy defines a ranking strategy interface.
type Strategy interface {
	// Name returns the strategy name.
	Name() string

	// Score scores a packet's relevance to the goal.
	Score(packet model.ContextPacket, goal string) float64

	// Weight returns the strategy's weight in combined scoring.
	Weight() float64
}

// Ranker defines the interface for multi-strategy ranking.
type Ranker interface {
	// Rank orders packets by combined strategy scores.
	Rank(packets []model.ContextPacket, goal string) []model.ContextPacket

	// AddStrategy adds a ranking strategy.
	AddStrategy(strategy Strategy)

	// Update adjusts strategy weights based on feedback.
	Update(strategy string, score float64)
}
