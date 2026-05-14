package agent

import (
	"context"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestExecutionJudge_PassesGoodResult(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "Here is the answer to your question.",
			"_tool_traces": []ToolTrace{
				{Name: "file_read", Output: `{"content":"hello"}`},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "read the file"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if !jr.Passed {
		t.Errorf("expected pass, got fail: score=%.2f, summary=%s", jr.Score, jr.Summary())
	}
}

func TestExecutionJudge_FailsHighErrorRate(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "I tried but failed.",
			"_tool_traces": []ToolTrace{
				{Name: "bash", Error: "command not found"},
				{Name: "bash", Error: "command not found"},
				{Name: "bash", Error: "permission denied"},
				{Name: "bash", Output: `{"exit_code":0}`},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "run the tests"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if jr.Passed {
		t.Errorf("expected fail (75%% error rate), got pass: score=%.2f", jr.Score)
	}
}

func TestExecutionJudge_FailsTrailingErrors(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "Done.",
			"_tool_traces": []ToolTrace{
				{Name: "bash", Output: `{"exit_code":0}`},
				{Name: "bash", Output: `{"exit_code":0}`},
				{Name: "bash", Error: "timeout"},
				{Name: "bash", Error: "timeout"},
				{Name: "bash", Error: "timeout"},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "check status"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if jr.Passed {
		t.Errorf("expected fail (trailing 3 errors), got pass: score=%.2f", jr.Score)
	}
}

func TestExecutionJudge_FailsWriteIntentNotFulfilled(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "I could not write the file due to permissions.",
			"_tool_traces": []ToolTrace{
				{Name: "bash", Error: "permission denied"},
				{Name: "file_write", Error: "path not allowed"},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "Create a file test.txt with hello"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if jr.Passed {
		t.Errorf("expected fail (write intent not fulfilled), got pass: score=%.2f", jr.Score)
	}
}

func TestExecutionJudge_FailsEmptyOutput(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "<think>\nLet me think about this...\n</think>\n",
			"_tool_traces": []ToolTrace{
				{Name: "bash", Output: `{"exit_code":0}`},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "what is 2+2"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if jr.Passed {
		t.Errorf("expected fail (only think block, no substance), got pass: score=%.2f", jr.Score)
	}
}

func TestExecutionJudge_PassesWriteWithBash(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "File created successfully.",
			"_tool_traces": []ToolTrace{
				{Name: "bash", Output: `{"exit_code":0,"stdout":"","command":"echo hello > test.txt"}`},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "Write hello to test.txt"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if !jr.Passed {
		t.Errorf("expected pass (bash write succeeded), got fail: score=%.2f, summary=%s", jr.Score, jr.Summary())
	}
}


func TestExecutionJudge_FailsWhenGoTestFails(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "I fixed the code.",
			"_tool_traces": []ToolTrace{
				{Name: "file_read", Output: `{"content":"..."}`},
				{Name: "file_write", Output: `{"success":true}`},
				{Name: "bash", Output: `{"command":"go test ./internal/tmp/ -run TestAdd -count=1","exit_code":1,"stdout":"--- FAIL: TestAdd (0.00s)\nFAIL"}`},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "fix the bug"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if jr.Passed {
		t.Errorf("expected fail (go test FAIL), got pass: score=%.2f", jr.Score)
	}
}

func TestExecutionJudge_PassesWhenGoTestPasses(t *testing.T) {
	j := &ExecutionJudge{}
	result := &AgentExecutionResult{
		Output: map[string]any{
			"text": "Fixed the bug and verified with go test. All tests pass now.",
			"_tool_traces": []ToolTrace{
				{Name: "file_write", Output: `{"success":true}`},
				{Name: "bash", Output: `{"command":"go test ./internal/tmp/ -run TestAdd -count=1","exit_code":0,"stdout":"ok  github.com/axis-cli/axis/internal/tmp\nPASS"}`},
			},
		},
	}
	task := &types.AgentTask{Input: map[string]any{"message": "fix the bug"}}
	jr := j.JudgeExecution(context.Background(), task, result)
	if jr == nil {
		t.Fatal("expected judgement result")
	}
	if !jr.Passed {
		t.Errorf("expected pass (go test PASS), got fail: score=%.2f, summary=%s", jr.Score, jr.Summary())
	}
}
