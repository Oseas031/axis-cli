package agent

import (
	"context"
	"strings"

	"github.com/axis-cli/axis/internal/agent/judgement"
	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
	"github.com/axis-cli/axis/internal/types"
)

// ExecutionJudge validates execution results using verifiable criteria.
// Unlike a rubber-stamp, this judge CAN and WILL fail tasks that don't
// meet quality thresholds.
type ExecutionJudge struct {
	MaxErrorRate float64 // default 0.3 (30%)
}

func (j *ExecutionJudge) maxErrorRate() float64 {
	if j.MaxErrorRate > 0 {
		return j.MaxErrorRate
	}
	return 0.3
}

func (j *ExecutionJudge) JudgeExecution(_ context.Context, task *types.AgentTask, result *AgentExecutionResult) *judgement.JudgementResult {
	if result == nil || result.Output == nil {
		return nil
	}

	traces, ok := result.Output["_tool_traces"]
	if !ok {
		return nil
	}
	traceList, ok := traces.([]ToolTrace)
	if !ok {
		return nil
	}

	jr := judgement.NewJudgementResult()

	// Criterion 1: Tool error rate must be below threshold
	j.judgeErrorRate(traceList, jr)

	// Criterion 2: Last N tool calls must not all be failures (Agent gave up)
	j.judgeTrailingFailures(traceList, jr)

	// Criterion 3: If task requires file creation, verify a write succeeded
	j.judgeIntentAlignment(task, traceList, jr)

	// Criterion 4: Output must have substantive content (not just <think> block)
	j.judgeOutputSubstance(result, jr)

	// Criterion 5: If Agent ran go test and it failed, the task is not done
	j.judgeTestResult(traceList, jr)

	return jr
}

func (j *ExecutionJudge) judgeErrorRate(traces []ToolTrace, jr *judgement.JudgementResult) {
	if len(traces) == 0 {
		return
	}
	errorCount := 0
	for _, t := range traces {
		if t.Error != "" {
			errorCount++
		}
	}
	rate := float64(errorCount) / float64(len(traces))
	passed := rate <= j.maxErrorRate()
	jr.AddJudgement(strategies.JudgementItem{
		CriteriaName: "tool_error_rate",
		Passed:       passed,
		Score:        1.0 - rate,
		Details:      formatRate(errorCount, len(traces), passed),
	})
}

func (j *ExecutionJudge) judgeTrailingFailures(traces []ToolTrace, jr *judgement.JudgementResult) {
	if len(traces) < 3 {
		return // not enough data
	}
	// Check last 3 tool calls
	tail := traces[len(traces)-3:]
	allFailed := true
	for _, t := range tail {
		if t.Error == "" {
			allFailed = false
			break
		}
	}
	jr.AddJudgement(strategies.JudgementItem{
		CriteriaName: "no_trailing_failures",
		Passed:       !allFailed,
		Score:        boolToScore(!allFailed),
		Details:      ternary(allFailed, "last 3 tool calls all failed — agent gave up", ""),
	})
}

func (j *ExecutionJudge) judgeIntentAlignment(task *types.AgentTask, traces []ToolTrace, jr *judgement.JudgementResult) {
	// Extract task prompt
	prompt := ""
	if task != nil && task.Input != nil {
		if msg, ok := task.Input["message"].(string); ok {
			prompt = strings.ToLower(msg)
		}
	}
	if prompt == "" {
		return // can't judge intent without prompt
	}

	// Check if task requires file creation
	requiresWrite := strings.Contains(prompt, "write") ||
		strings.Contains(prompt, "create") ||
		strings.Contains(prompt, "生成") ||
		strings.Contains(prompt, "写入")

	if !requiresWrite {
		return // not a write task, skip this criterion
	}

	// Verify at least one successful write operation
	hasSuccessfulWrite := false
	for _, t := range traces {
		if t.Error != "" {
			continue
		}
		if t.Name == "file_write" {
			hasSuccessfulWrite = true
			break
		}
		// bash with echo/cat/tee redirecting to file also counts
		if t.Name == "bash" && t.Output != "" {
			if strings.Contains(t.Output, "\"exit_code\":0") {
				hasSuccessfulWrite = true
				break
			}
		}
	}

	jr.AddJudgement(strategies.JudgementItem{
		CriteriaName: "intent_write_fulfilled",
		Passed:       hasSuccessfulWrite,
		Score:        boolToScore(hasSuccessfulWrite),
		Details:      ternary(!hasSuccessfulWrite, "task requires file creation but no successful write detected", ""),
	})
}

func (j *ExecutionJudge) judgeOutputSubstance(result *AgentExecutionResult, jr *judgement.JudgementResult) {
	text, ok := result.Output["text"]
	if !ok {
		jr.AddJudgement(strategies.JudgementItem{
			CriteriaName: "output_substance",
			Passed:       false,
			Score:        0,
			Details:      "no text output produced",
		})
		return
	}

	s, ok := text.(string)
	if !ok || len(s) == 0 {
		jr.AddJudgement(strategies.JudgementItem{
			CriteriaName: "output_substance",
			Passed:       false,
			Score:        0,
			Details:      "empty text output",
		})
		return
	}

	// Strip <think> blocks and check remaining content
	substance := stripThinkBlocks(s)
	hasSubstance := len(strings.TrimSpace(substance)) > 20

	jr.AddJudgement(strategies.JudgementItem{
		CriteriaName: "output_substance",
		Passed:       hasSubstance,
		Score:        boolToScore(hasSubstance),
		Details:      ternary(!hasSubstance, "output is only thinking/empty after stripping <think> blocks", ""),
	})
}

// stripThinkBlocks removes <think>...</think> content from LLM output.
func stripThinkBlocks(s string) string {
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			return s
		}
		end := strings.Index(s[start:], "</think>")
		if end == -1 {
			return s[:start] // unclosed think block, strip from start
		}
		s = s[:start] + s[start+end+len("</think>"):]
	}
}

// judgeTestResult checks if Agent ran go test and whether it passed.
// If the last bash tool call containing "go test" shows FAIL, the task failed.
// This is the authoritative oracle — compiler/tests decide, not LLM.
func (j *ExecutionJudge) judgeTestResult(traces []ToolTrace, jr *judgement.JudgementResult) {
	lastTestIdx := -1
	for i, t := range traces {
		if t.Name == "bash" && t.Error == "" && strings.Contains(t.Output, "go test") {
			lastTestIdx = i
		}
	}
	if lastTestIdx == -1 {
		return
	}
	output := traces[lastTestIdx].Output
	passed := strings.Contains(output, "PASS") && !strings.Contains(output, "FAIL")
	jr.AddJudgement(strategies.JudgementItem{
		CriteriaName: "test_oracle",
		Passed:       passed,
		Score:        boolToScore(passed),
		Details:      ternary(!passed, "go test reported FAIL in final test run", ""),
	})
}

func formatRate(errors, total int, passed bool) string {
	if passed {
		return ""
	}
	return strings.Join([]string{
		string(rune('0' + errors/10)), string(rune('0' + errors%10)),
		"/",
		string(rune('0' + total/10)), string(rune('0' + total%10)),
		" tool calls failed (>30% threshold)",
	}, "")
}

func boolToScore(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func ternary(cond bool, t, f string) string {
	if cond {
		return t
	}
	return f
}
