package contextpack

import "math/rand"

// v1: static strategies. TODO: adaptive strategy selection based on task type

// AnchorMitigationStrategy controls how packets are reordered to prevent
// the Lead Anchor Effect (arXiv:2605.10698). The first item in multi-source
// context disproportionately influences agent reasoning.
type AnchorMitigationStrategy int

const (
	// AnchorNone applies no reordering (default, backward compatible).
	AnchorNone AnchorMitigationStrategy = iota
	// AnchorShuffle randomizes packets with equal relevance scores.
	AnchorShuffle
	// AnchorCredibilitySort interleaves packets by source (round-robin).
	AnchorCredibilitySort
)

// ApplyAnchorMitigation returns a new slice with packets reordered per strategy.
// The input slice is never modified.
func ApplyAnchorMitigation(packets []ContextPacket, strategy AnchorMitigationStrategy, seed int64) []ContextPacket {
	if len(packets) <= 1 {
		return copyPackets(packets)
	}
	switch strategy {
	case AnchorShuffle:
		return applyShuffle(packets, seed)
	case AnchorCredibilitySort:
		return applyCredibilitySort(packets)
	default:
		return copyPackets(packets)
	}
}

func copyPackets(packets []ContextPacket) []ContextPacket {
	if packets == nil {
		return nil
	}
	out := make([]ContextPacket, len(packets))
	copy(out, packets)
	return out
}

func applyShuffle(packets []ContextPacket, seed int64) []ContextPacket {
	out := copyPackets(packets)
	rng := rand.New(rand.NewSource(seed))

	// Shuffle groups of packets with near-equal relevance.
	// Assumes input is sorted by descending relevance (assembler guarantees this).
	const tolerance = 0.01
	i := 0
	for i < len(out) {
		j := i + 1
		for j < len(out) && (out[i].Relevance-out[j].Relevance) < tolerance {
			// Guard: if out[j] has higher relevance than out[i], stop grouping.
			if out[j].Relevance > out[i].Relevance {
				break
			}
			j++
		}
		if j-i > 1 {
			rng.Shuffle(j-i, func(a, b int) {
				out[i+a], out[i+b] = out[i+b], out[i+a]
			})
		}
		i = j
	}
	return out
}

func applyCredibilitySort(packets []ContextPacket) []ContextPacket {
	// Group by source, preserving order within each source.
	order := make([]string, 0)
	groups := make(map[string][]ContextPacket)
	for _, p := range packets {
		if _, exists := groups[p.Source]; !exists {
			order = append(order, p.Source)
		}
		groups[p.Source] = append(groups[p.Source], p)
	}

	// Round-robin interleave.
	out := make([]ContextPacket, 0, len(packets))
	idx := make(map[string]int, len(order))
	for len(out) < len(packets) {
		for _, src := range order {
			if idx[src] < len(groups[src]) {
				out = append(out, groups[src][idx[src]])
				idx[src]++
			}
		}
	}
	return out
}
