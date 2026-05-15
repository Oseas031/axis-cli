package agent

import "testing"

func TestPartition_TwoSameOneDifferent(t *testing.T) {
	cp := &CandidatePool{
		Candidates: []Candidate{
			{ID: "a", Code: "return x+1", Source: "claude"},
			{ID: "b", Code: "return x+1", Source: "gpt4"},
			{ID: "c", Code: "return x*2", Source: "deepseek"},
		},
		TestInputs: []TestInput{{Input: "5", Expected: "6"}},
	}

	dominant := cp.SelectDominant()
	if dominant == nil {
		t.Fatal("expected dominant class")
	}
	if dominant.Size != 2 {
		t.Fatalf("expected dominant size 2, got %d", dominant.Size)
	}
	if dominant.Members[0] != "a" || dominant.Members[1] != "b" {
		t.Fatalf("expected members [a b], got %v", dominant.Members)
	}
}

func TestPartition_AllSame(t *testing.T) {
	cp := &CandidatePool{
		Candidates: []Candidate{
			{ID: "a", Code: "return x+1", Source: "claude"},
			{ID: "b", Code: "return x+1", Source: "gpt4"},
			{ID: "c", Code: "return x+1", Source: "deepseek"},
		},
		TestInputs: []TestInput{{Input: "5", Expected: "6"}},
	}

	classes := cp.Partition()
	if len(classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(classes))
	}
	if classes[0].Size != 3 {
		t.Fatalf("expected size 3, got %d", classes[0].Size)
	}
}

func TestPartition_AllDifferent(t *testing.T) {
	cp := &CandidatePool{
		Candidates: []Candidate{
			{ID: "a", Code: "return x+1", Source: "claude"},
			{ID: "b", Code: "return x*2", Source: "gpt4"},
			{ID: "c", Code: "return x-1", Source: "deepseek"},
		},
		TestInputs: []TestInput{{Input: "5"}},
	}

	classes := cp.Partition()
	if len(classes) != 3 {
		t.Fatalf("expected 3 classes, got %d", len(classes))
	}
	dominant := cp.SelectDominant()
	if dominant.Size != 1 {
		t.Fatalf("expected dominant size 1 (no consensus), got %d", dominant.Size)
	}
}


func TestValidateDiversity_NoneAlwaysPasses(t *testing.T) {
	pool := &CandidatePool{
		Candidates: []Candidate{{ID: "1", Source: "gpt4"}},
		Diversity:  DiversityNone,
	}
	if err := pool.ValidateDiversity(); err != nil {
		t.Fatalf("DiversityNone should always pass, got: %v", err)
	}
}

func TestValidateDiversity_HeterogeneousPassesWithTwoSources(t *testing.T) {
	pool := &CandidatePool{
		Candidates: []Candidate{
			{ID: "1", Source: "gpt4"},
			{ID: "2", Source: "claude"},
		},
		Diversity: DiversityHeterogeneous,
	}
	if err := pool.ValidateDiversity(); err != nil {
		t.Fatalf("expected pass with 2 distinct sources, got: %v", err)
	}
}

func TestValidateDiversity_HeterogeneousFailsWithOneSource(t *testing.T) {
	pool := &CandidatePool{
		Candidates: []Candidate{
			{ID: "1", Source: "gpt4"},
			{ID: "2", Source: "gpt4"},
		},
		Diversity: DiversityHeterogeneous,
	}
	if err := pool.ValidateDiversity(); err == nil {
		t.Fatal("expected error with single source under heterogeneous policy")
	}
}

func TestDistinctSources(t *testing.T) {
	pool := &CandidatePool{
		Candidates: []Candidate{
			{ID: "1", Source: "gpt4"},
			{ID: "2", Source: "claude"},
			{ID: "3", Source: "gpt4"},
		},
	}
	sources := pool.DistinctSources()
	if len(sources) != 2 {
		t.Fatalf("expected 2 distinct sources, got %d", len(sources))
	}
}
