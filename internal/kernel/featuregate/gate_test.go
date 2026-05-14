package featuregate

import "testing"

func TestNewGateDefaults(t *testing.T) {
	g := NewGate()
	if !g.IsUnlocked(FeatureBashTool) {
		t.Error("expected bash_tool unlocked by default")
	}
	if !g.IsUnlocked(FeatureFileWrite) {
		t.Error("expected file_write unlocked by default")
	}
	if g.IsUnlocked(FeatureNetwork) {
		t.Error("expected network locked by default")
	}
	if g.IsUnlocked(FeatureEvolution) {
		t.Error("expected evolution locked by default")
	}
	if g.IsUnlocked(FeatureSpawn) {
		t.Error("expected spawn locked by default")
	}
}

func TestUnlockAndLock(t *testing.T) {
	g := NewGate()
	g.Unlock(FeatureNetwork)
	if !g.IsUnlocked(FeatureNetwork) {
		t.Error("expected network unlocked after Unlock")
	}
	g.Lock(FeatureNetwork)
	if g.IsUnlocked(FeatureNetwork) {
		t.Error("expected network locked after Lock")
	}
}

func TestUnlockedFeatures(t *testing.T) {
	g := NewGate()
	features := g.UnlockedFeatures()
	if len(features) != 2 {
		t.Fatalf("expected 2 default features, got %d", len(features))
	}
	found := map[Feature]bool{}
	for _, f := range features {
		found[f] = true
	}
	if !found[FeatureBashTool] || !found[FeatureFileWrite] {
		t.Error("expected bash_tool and file_write in UnlockedFeatures")
	}
}
