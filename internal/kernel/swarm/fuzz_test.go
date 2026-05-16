package swarm

import (
	"testing"
)

func FuzzParseFromMetadata(f *testing.F) {
	f.Add("parallel_vote", "2", "3", "heterogeneous", "shuffled")
	f.Add("", "", "", "", "")
	f.Add("unknown", "-1", "0", "bad", "bad")
	f.Add("parallel_vote", "999999", "1", "none", "fixed")
	f.Fuzz(func(t *testing.T, pattern, minS, maxS, div, order string) {
		meta := map[string]string{
			"swarm.pattern":   pattern,
			"swarm.min_size":  minS,
			"swarm.max_size":  maxS,
			"swarm.diversity": div,
			"swarm.order":     order,
		}
		cfg := ParseFromMetadata(meta)
		if cfg != nil {
			_ = cfg.Validate()
		}
	})
}

func FuzzAggregate(f *testing.F) {
	f.Add(3, true)
	f.Add(1, false)
	f.Add(10, true)
	f.Fuzz(func(t *testing.T, count int, allSame bool) {
		if count <= 0 || count > 100 {
			return
		}
		results := make([]SingleResult, count)
		for i := range results {
			if allSame {
				results[i] = SingleResult{AgentID: "a", Output: map[string]any{"x": "y"}}
			} else {
				results[i] = SingleResult{AgentID: "a", Output: map[string]any{"x": i}}
			}
		}
		_, _ = Aggregate(results)
	})
}
