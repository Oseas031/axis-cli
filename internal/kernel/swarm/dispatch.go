package swarm

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/axis-cli/axis/internal/types"
)

// DispatchFn is the callback injected by Dispatcher to execute a single agent task.
type DispatchFn func(ctx context.Context, task *types.AgentTask, provider string) (map[string]any, error)

// SelectAgents picks agents from available providers satisfying config constraints.
func SelectAgents(available []string, cfg *SwarmConfig) ([]AgentSlot, error) {
	if len(available) == 0 {
		return nil, errors.New("swarm: no available providers")
	}
	if len(available) < cfg.MinSize {
		return nil, fmt.Errorf("swarm: need %d agents but only %d available", cfg.MinSize, len(available))
	}

	// For heterogeneous diversity, need >=2 distinct providers
	if cfg.Diversity == "heterogeneous" {
		unique := uniqueProviders(available)
		if len(unique) < 2 {
			return nil, errors.New("swarm: heterogeneous diversity requires >=2 distinct providers")
		}
	}

	// Select up to MaxSize agents
	count := cfg.MaxSize
	if count > len(available) {
		count = len(available)
	}

	slots := make([]AgentSlot, count)
	for i := 0; i < count; i++ {
		slots[i] = AgentSlot{
			AgentID:  fmt.Sprintf("swarm-agent-%d", i),
			Provider: available[i],
		}
	}
	return slots, nil
}

// Dispatch executes a task across multiple agents in parallel and aggregates results.
func Dispatch(ctx context.Context, task *types.AgentTask, cfg *SwarmConfig, agents []AgentSlot, fn DispatchFn) (*SwarmResult, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Shuffle if requested
	if cfg.Order == "shuffled" {
		shuffleSlots(agents)
	}

	// Parallel execution
	results := make([]SingleResult, len(agents))
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, slot := range agents {
		wg.Add(1)
		go func(idx int, s AgentSlot) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				results[idx] = SingleResult{AgentID: s.AgentID, Error: "context cancelled"}
				return
			default:
			}
			output, err := fn(ctx, task, s.Provider)
			if err != nil {
				results[idx] = SingleResult{AgentID: s.AgentID, Error: err.Error()}
			} else {
				results[idx] = SingleResult{AgentID: s.AgentID, Output: output}
			}
		}(i, slot)
	}
	wg.Wait()

	sr, err := Aggregate(results)
	if err != nil {
		return nil, err
	}
	sr.Agents = agents
	return sr, nil
}

func shuffleSlots(slots []AgentSlot) {
	for i := len(slots) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		slots[i], slots[j] = slots[j], slots[i]
	}
}

func uniqueProviders(providers []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range providers {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
