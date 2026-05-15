package swarm

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
)

// AgentSlot represents one participant in the swarm execution.
type AgentSlot struct {
	AgentID  string
	Provider string
}

// SingleResult is one agent's output.
type SingleResult struct {
	AgentID string
	Output  map[string]any
	Error   string
}

// SwarmResult is the output of a swarm execution.
type SwarmResult struct {
	Agents     []AgentSlot
	Results    []SingleResult
	Winner     *SingleResult
	Confidence float64
	Unanimous  bool
}

// Aggregate performs majority vote on results by hashing outputs.
func Aggregate(results []SingleResult) (*SwarmResult, error) {
	if len(results) == 0 {
		return nil, errors.New("swarm: no results to aggregate")
	}

	// Hash each output
	groups := make(map[string][]int) // hash -> indices
	for i, r := range results {
		if r.Error != "" {
			continue
		}
		h := hashOutput(r.Output)
		groups[h] = append(groups[h], i)
	}

	if len(groups) == 0 {
		return nil, errors.New("swarm: all agents failed")
	}

	// Find largest group
	var bestHash string
	var bestSize int
	for h, indices := range groups {
		if len(indices) > bestSize {
			bestHash = h
			bestSize = len(indices)
		}
	}

	winnerIdx := groups[bestHash][0]
	winner := results[winnerIdx]
	total := 0
	for _, indices := range groups {
		total += len(indices)
	}

	return &SwarmResult{
		Results:    results,
		Winner:     &winner,
		Confidence: float64(bestSize) / float64(total),
		Unanimous:  bestSize == total,
	}, nil
}

func hashOutput(output map[string]any) string {
	data, err := json.Marshal(output)
	if err != nil {
		return fmt.Sprintf("marshal-error-%p", output)
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}
