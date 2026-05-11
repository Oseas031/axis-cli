package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestJudgeCommand(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := newJudgeCommand()
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("judge command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Self-Judgement Result") {
		t.Errorf("expected output to contain 'Self-Judgement Result', got:\n%s", out)
	}
	if !strings.Contains(out, "Passed:") {
		t.Errorf("expected output to contain 'Passed:', got:\n%s", out)
	}
	if !strings.Contains(out, "Score:") {
		t.Errorf("expected output to contain 'Score:', got:\n%s", out)
	}
	if !strings.Contains(out, "Confidence:") {
		t.Errorf("expected output to contain 'Confidence:', got:\n%s", out)
	}
}

func TestJudgeCommand_OutputFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := newJudgeCommand()
	cmd.SetOut(buf)

	if err := runJudge(cmd, nil); err != nil {
		t.Fatalf("runJudge failed: %v", err)
	}

	out := buf.String()
	lines := strings.Split(out, "\n")
	var hasHeader, hasFooter bool
	for _, line := range lines {
		if strings.Contains(line, "=== Self-Judgement Result ===") {
			hasHeader = true
		}
		if strings.Contains(line, "=============================") {
			hasFooter = true
		}
	}
	if !hasHeader {
		t.Error("expected header separator in output")
	}
	if !hasFooter {
		t.Error("expected footer separator in output")
	}
}

func TestDefaultBootstrapCriteria(t *testing.T) {
	criteria := defaultBootstrapCriteria()
	if len(criteria) == 0 {
		t.Fatal("expected non-empty criteria")
	}

	for _, c := range criteria {
		if c.Name == "" {
			t.Error("expected criteria name to be non-empty")
		}
		if c.Weight < 0 || c.Weight > 1 {
			t.Errorf("expected weight between 0 and 1, got %f", c.Weight)
		}
		if !c.Enabled {
			t.Errorf("expected criteria %s to be enabled", c.Name)
		}
	}
}
