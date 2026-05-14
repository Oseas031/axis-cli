package agent

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// CandidatePool manages multiple solution candidates for a coding task.
type CandidatePool struct {
	Candidates []Candidate
	TestInputs []TestInput
}

type Candidate struct {
	ID     string
	Code   string
	Source string // which model/temperature produced it
}

type TestInput struct {
	Input    string
	Expected string // empty if unknown
}

type EquivalenceClass struct {
	Members []string // candidate IDs
	Output  string   // shared output for test inputs
	Size    int
}

// Partition runs all candidates against test inputs and groups by output equivalence.
func (cp *CandidatePool) Partition() []EquivalenceClass {
	groups := make(map[string]*EquivalenceClass)

	for _, c := range cp.Candidates {
		key := cp.outputKey(c)
		if g, ok := groups[key]; ok {
			g.Members = append(g.Members, c.ID)
			g.Size++
		} else {
			groups[key] = &EquivalenceClass{
				Members: []string{c.ID},
				Output:  cp.computeOutput(c),
				Size:    1,
			}
		}
	}

	result := make([]EquivalenceClass, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Size > result[j].Size
	})
	return result
}

// SelectDominant returns the largest equivalence class.
func (cp *CandidatePool) SelectDominant() *EquivalenceClass {
	classes := cp.Partition()
	if len(classes) == 0 {
		return nil
	}
	return &classes[0]
}

func (cp *CandidatePool) computeOutput(c Candidate) string {
	// Data-structure-only: use code content as simulated output.
	var parts []string
	for _, t := range cp.TestInputs {
		parts = append(parts, fmt.Sprintf("%s:%s", t.Input, c.Code))
	}
	return strings.Join(parts, "|")
}

func (cp *CandidatePool) outputKey(c Candidate) string {
	out := cp.computeOutput(c)
	h := sha256.Sum256([]byte(out))
	return fmt.Sprintf("%x", h)
}
