package contextpack

import "testing"

func anchorTestPackets(ids []string, sources []string, relevances []float64) []ContextPacket {
	out := make([]ContextPacket, len(ids))
	for i := range ids {
		out[i] = ContextPacket{ID: ids[i], Source: sources[i], Relevance: relevances[i], Reason: "test"}
	}
	return out
}

func packetIDs(packets []ContextPacket) []string {
	out := make([]string, len(packets))
	for i, p := range packets {
		out[i] = p.ID
	}
	return out
}

func strSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAnchorNone_PreservesOrder(t *testing.T) {
	packets := anchorTestPackets(
		[]string{"a", "b", "c"},
		[]string{"s1", "s2", "s3"},
		[]float64{0.9, 0.8, 0.7},
	)
	result := ApplyAnchorMitigation(packets, AnchorNone, 42)
	if !strSliceEq(packetIDs(result), []string{"a", "b", "c"}) {
		t.Fatalf("expected unchanged order, got %v", packetIDs(result))
	}
}

func TestAnchorNone_DoesNotModifyInput(t *testing.T) {
	packets := anchorTestPackets([]string{"a", "b"}, []string{"s", "s"}, []float64{0.5, 0.5})
	orig := make([]ContextPacket, len(packets))
	copy(orig, packets)
	_ = ApplyAnchorMitigation(packets, AnchorShuffle, 99)
	if !strSliceEq(packetIDs(packets), packetIDs(orig)) {
		t.Fatal("input slice was modified")
	}
}

func TestAnchorShuffle_FixedSeed(t *testing.T) {
	packets := anchorTestPackets(
		[]string{"a", "b", "c", "d"},
		[]string{"s", "s", "s", "s"},
		[]float64{0.8, 0.8, 0.8, 0.8},
	)
	r1 := ApplyAnchorMitigation(packets, AnchorShuffle, 42)
	r2 := ApplyAnchorMitigation(packets, AnchorShuffle, 42)
	if !strSliceEq(packetIDs(r1), packetIDs(r2)) {
		t.Fatal("same seed should produce same order")
	}
}

func TestAnchorShuffle_PreservesDifferentRelevance(t *testing.T) {
	packets := anchorTestPackets(
		[]string{"high", "mid", "low"},
		[]string{"s", "s", "s"},
		[]float64{0.9, 0.5, 0.1},
	)
	result := ApplyAnchorMitigation(packets, AnchorShuffle, 42)
	if !strSliceEq(packetIDs(result), []string{"high", "mid", "low"}) {
		t.Fatalf("different-relevance packets should keep order, got %v", packetIDs(result))
	}
}

func TestAnchorCredibilitySort_Interleaves(t *testing.T) {
	packets := anchorTestPackets(
		[]string{"a1", "a2", "a3", "b1", "b2"},
		[]string{"A", "A", "A", "B", "B"},
		[]float64{0.9, 0.8, 0.7, 0.85, 0.75},
	)
	result := ApplyAnchorMitigation(packets, AnchorCredibilitySort, 0)
	got := packetIDs(result)
	expected := []string{"a1", "b1", "a2", "b2", "a3"}
	if !strSliceEq(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestEdge_Empty(t *testing.T) {
	for _, s := range []AnchorMitigationStrategy{AnchorNone, AnchorShuffle, AnchorCredibilitySort} {
		result := ApplyAnchorMitigation(nil, s, 0)
		if result != nil {
			t.Fatalf("strategy %d: expected nil for nil input, got %v", s, result)
		}
	}
}

func TestEdge_Single(t *testing.T) {
	packets := anchorTestPackets([]string{"x"}, []string{"s"}, []float64{1.0})
	for _, s := range []AnchorMitigationStrategy{AnchorNone, AnchorShuffle, AnchorCredibilitySort} {
		result := ApplyAnchorMitigation(packets, s, 0)
		if len(result) != 1 || result[0].ID != "x" {
			t.Fatalf("strategy %d: expected [x], got %v", s, packetIDs(result))
		}
	}
}
