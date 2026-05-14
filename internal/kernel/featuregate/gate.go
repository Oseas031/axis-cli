package featuregate

import "sync"

// Feature represents a gated capability.
type Feature string

const (
	FeatureBashTool  Feature = "bash_tool"
	FeatureFileWrite Feature = "file_write"
	FeatureNetwork   Feature = "network"
	FeatureEvolution Feature = "evolution"
	FeatureSpawn     Feature = "spawn"
)

// Gate controls which features are unlocked.
type Gate struct {
	mu       sync.RWMutex
	unlocked map[Feature]bool
}

func NewGate() *Gate {
	return &Gate{unlocked: map[Feature]bool{
		FeatureBashTool:  true, // always available
		FeatureFileWrite: true, // always available
	}}
}

func (g *Gate) IsUnlocked(f Feature) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.unlocked[f]
}

func (g *Gate) Unlock(f Feature) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.unlocked[f] = true
}

func (g *Gate) Lock(f Feature) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.unlocked, f)
}

func (g *Gate) UnlockedFeatures() []Feature {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []Feature
	for f := range g.unlocked {
		result = append(result, f)
	}
	return result
}
