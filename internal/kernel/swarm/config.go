package swarm

import (
	"errors"
	"strconv"
)

// SwarmConfig is parsed from task metadata. Minimal, no file-system backing.
type SwarmConfig struct {
	Pattern   string // "parallel_vote" (only supported in v1)
	MinSize   int    // default 2
	MaxSize   int    // default 3
	Diversity string // "none" | "heterogeneous"
	Order     string // "fixed" | "shuffled"
}

// ParseFromMetadata extracts SwarmConfig from task metadata.
// Returns nil if no swarm.* keys present.
func ParseFromMetadata(meta map[string]string) *SwarmConfig {
	if meta == nil {
		return nil
	}
	pattern, ok := meta["swarm.pattern"]
	if !ok || pattern == "" {
		return nil
	}
	cfg := &SwarmConfig{
		Pattern:   pattern,
		MinSize:   2,
		MaxSize:   3,
		Diversity: "heterogeneous",
		Order:     "shuffled",
	}
	if v, ok := meta["swarm.min_size"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MinSize = n
		}
	}
	if v, ok := meta["swarm.max_size"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxSize = n
		}
	}
	if v, ok := meta["swarm.diversity"]; ok {
		cfg.Diversity = v
	}
	if v, ok := meta["swarm.order"]; ok {
		cfg.Order = v
	}
	return cfg
}

// Validate checks SwarmConfig constraints.
func (c *SwarmConfig) Validate() error {
	if c.Pattern != "parallel_vote" {
		return errors.New("swarm: unsupported pattern: " + c.Pattern)
	}
	if c.MinSize < 2 {
		return errors.New("swarm: min_size must be >= 2")
	}
	if c.MaxSize < c.MinSize {
		return errors.New("swarm: max_size must be >= min_size")
	}
	if c.Diversity != "none" && c.Diversity != "heterogeneous" {
		return errors.New("swarm: diversity must be 'none' or 'heterogeneous'")
	}
	if c.Order != "fixed" && c.Order != "shuffled" {
		return errors.New("swarm: order must be 'fixed' or 'shuffled'")
	}
	return nil
}
