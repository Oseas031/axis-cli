// Command swarm-validate-real validates whether multi-candidate majority vote
// reduces failure rate compared to single-candidate execution.
//
// Methodology:
// 1. Define a task that LLMs frequently get wrong (edge cases in interval scheduling)
// 2. Call the LLM N times, each time extracting Go code from the response
// 3. For each response, compile and run test cases to determine pass/fail
// 4. Compare: single-call failure rate vs 3-candidate majority-vote failure rate
//
// Usage: go run ./cmd/swarm-validate-real
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/providerconfig"
	"github.com/axis-cli/axis/internal/project"
)

const prompt = `Implement the following Go function. Return ONLY the function body, no package declaration, no imports, no main, no explanation, no markdown fences.

Function signature:
func maxNonOverlapping(intervals [][2]int) int

The function takes a list of intervals where each interval is [start, end) (half-open: includes start, excludes end). Return the maximum number of non-overlapping intervals that can be selected.

Rules:
- Intervals are non-overlapping if they don't share any point. [1,3) and [3,5) are non-overlapping.
- [1,3) and [2,4) ARE overlapping.
- Empty input returns 0.
- Intervals may be given in any order.
- Intervals where start >= end are invalid and should be skipped.
- Duplicate intervals count as separate choices (but you can only pick one copy).

Example: maxNonOverlapping([][2]int{{1,3},{2,4},{3,5}}) returns 2 (pick [1,3) and [3,5)).`

// testHarness wraps the LLM output in a compilable Go program with test cases.
const testHarness = `package main

import (
	"fmt"
	"os"
	"sort"
)

%s

func main() {
	_ = sort.Ints

	tests := []struct {
		input    [][2]int
		expected int
	}{
		{[][2]int{}, 0},
		{[][2]int{{1, 3}}, 1},
		{[][2]int{{1, 3}, {3, 5}}, 2},
		{[][2]int{{1, 3}, {2, 4}, {3, 5}}, 2},
		{[][2]int{{1, 10}, {2, 5}, {3, 7}}, 1},
		{[][2]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}}, 4},
		{[][2]int{{5, 3}, {1, 2}, {3, 4}}, 2},
		{[][2]int{{1, 1}, {2, 3}}, 1},
		{[][2]int{{1, 3}, {3, 5}, {5, 7}}, 3},
		{[][2]int{{1, 10}, {2, 3}, {4, 5}, {6, 7}}, 3},
		{[][2]int{{1, 3}, {1, 3}, {1, 3}}, 1},
		{[][2]int{{1, 100}, {1, 2}, {2, 3}, {3, 4}}, 3},
	}

	passed := 0
	for i, tc := range tests {
		got := maxNonOverlapping(tc.input)
		if got == tc.expected {
			passed++
		} else {
			fmt.Fprintf(os.Stderr, "FAIL test %%d: input=%%v expected=%%d got=%%d\n", i, tc.input, tc.expected, got)
		}
	}
	fmt.Printf("%%d/%%d", passed, len(tests))
	if passed < len(tests) {
		os.Exit(1)
	}
}
`

func main() {
	root := project.MustResolveRoot()
	store := providerconfig.NewStore(root)
	cfg, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// Use active profile
	p, ok := cfg.Active()
	if !ok {
		fmt.Fprintf(os.Stderr, "no active provider profile\n")
		os.Exit(1)
	}

	opts := p.ProviderOptions()
	opts = append(opts, provider.WithTimeout(300*time.Second))
	opts = append(opts, provider.WithMaxRetries(1))
	mp, err := provider.NewProvider(p.Provider, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create provider: %v\n", err)
		os.Exit(1)
	}

	const totalRuns = 5
	const candidatesPerVote = 3

	fmt.Println("=== Swarm Prerequisite Validation (Real) ===")
	fmt.Printf("Model: %s/%s\n", p.Provider, p.Model)
	fmt.Printf("Task: maxNonOverlapping (interval scheduling with edge cases)\n")
	fmt.Printf("Runs: %d single + %d majority-vote (x%d candidates each)\n\n", totalRuns, totalRuns, candidatesPerVote)

	totalCalls := totalRuns + totalRuns*candidatesPerVote
	fmt.Printf("Making %d LLM calls...\n", totalCalls)

	type callResult struct {
		code   string
		passed bool
		score  string
		err    string
	}

	results := make([]callResult, totalCalls)
	for i := 0; i < totalCalls; i++ {
		if i > 0 {
			time.Sleep(5 * time.Second)
		}
		code, err := callLLM(mp, prompt)
		if err != nil {
			results[i] = callResult{err: err.Error()}
			fmt.Printf("  [%d/%d] LLM error: %v\n", i+1, totalCalls, err)
			continue
		}
		passed, score, testErr := runTests(code)
		results[i] = callResult{code: code, passed: passed, score: score, err: testErr}
		status := "✓"
		if !passed {
			status = "✗"
		}
		fmt.Printf("  [%d/%d] %s %s\n", i+1, totalCalls, status, score)
	}

	// Analyze
	singlePass := 0
	for i := 0; i < totalRuns; i++ {
		if results[i].passed {
			singlePass++
		}
	}

	votePass := 0
	for v := 0; v < totalRuns; v++ {
		base := totalRuns + v*candidatesPerVote
		passCount := 0
		for j := 0; j < candidatesPerVote; j++ {
			if results[base+j].passed {
				passCount++
			}
		}
		if passCount > candidatesPerVote/2 {
			votePass++
		}
	}

	fmt.Println("\n=== Results ===")
	fmt.Printf("Single-call:    %d/%d passed (%.0f%% success rate)\n", singlePass, totalRuns, float64(singlePass)/float64(totalRuns)*100)
	fmt.Printf("Majority-vote:  %d/%d passed (%.0f%% success rate)\n", votePass, totalRuns, float64(votePass)/float64(totalRuns)*100)

	singleRate := float64(singlePass) / float64(totalRuns)
	voteRate := float64(votePass) / float64(totalRuns)

	fmt.Println("\n=== Conclusion ===")
	if singleRate >= 0.9 {
		fmt.Println("⚠ Single-call success rate already ≥90%. Task may be too easy.")
	} else if voteRate > singleRate+0.1 {
		fmt.Printf("✓ Majority vote improved success rate by +%.0f%%.\n", (voteRate-singleRate)*100)
		fmt.Println("  Swarm topology implementation is justified.")
	} else {
		fmt.Println("✗ Majority vote did NOT significantly improve success rate.")
	}
}

func callLLM(mp provider.ModelProvider, prompt string) (string, error) {
	req := &provider.ModelRequest{
		Input:        map[string]any{"message": prompt},
		SystemPrompt: "You are a Go programmer. Output only the function implementation. No package, no imports, no main, no markdown, no explanation.",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	resp, err := mp.Execute(ctx, req)
	if err != nil {
		return "", err
	}
	text := fmt.Sprintf("%v", resp.Output)
	return extractFunction(text), nil
}

func extractFunction(raw string) string {
	raw = strings.TrimPrefix(raw, "map[text:")
	raw = strings.TrimSuffix(raw, "]")

	// Strip <think>...</think> (MiniMax thinking model)
	if idx := strings.Index(raw, "</think>"); idx >= 0 {
		raw = raw[idx+len("</think>"):]
	}

	re := regexp.MustCompile("(?s)```(?:go)?\\s*(.*?)```")
	if m := re.FindStringSubmatch(raw); len(m) > 1 {
		raw = m[1]
	}

	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "func") {
		if idx := strings.Index(raw, "func"); idx >= 0 {
			raw = raw[idx:]
		}
	}
	if !strings.Contains(raw, "maxNonOverlapping") {
		raw = "func maxNonOverlapping(intervals [][2]int) int {\n" + raw + "\n}"
	}
	return raw
}

func runTests(code string) (passed bool, score string, errMsg string) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("swarm-validate-%d-%d", os.Getpid(), rand.Int63()))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, "0/0", err.Error()
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module swarmtest\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(fmt.Sprintf(testHarness, code)), 0644)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(output))

	if err != nil {
		if strings.Contains(outStr, "/") {
			for _, l := range strings.Split(outStr, "\n") {
				if strings.Contains(l, "/") && !strings.Contains(l, "FAIL") {
					return false, strings.TrimSpace(l), outStr
				}
			}
		}
		return false, "compile_error", outStr
	}
	return true, outStr, ""
}
