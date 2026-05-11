package main

import (
	"fmt"

	"github.com/axis-cli/axis/internal/agent/judgement"
	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
	"github.com/spf13/cobra"
)

func newJudgeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "judge",
		Short: "Run self-judgement and display the result",
		Long: `Run a self-judgement evaluation using the built-in validation strategies
and display the aggregated result. This is a read-only diagnostic command.`,
		RunE: runJudge,
	}
}

func runJudge(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	engine := judgement.NewEngine()
	criteria := defaultBootstrapCriteria()

	// Build a synthetic execution result from any available validation data.
	// P0: run judgement with a minimal synthetic input to verify the pipeline.
	input := map[string]any{
		"tests_passed": 1,
		"tests_failed": 0,
		"coverage":     1.0,
	}

	result, err := engine.Judge(input, criteria)
	if err != nil {
		return fmt.Errorf("judgement failed: %w", err)
	}

	fmt.Fprintln(out, "=== Self-Judgement Result ===")
	fmt.Fprintf(out, "Passed:     %t\n", result.Passed)
	fmt.Fprintf(out, "Score:      %.2f\n", result.Score)
	fmt.Fprintf(out, "Confidence: %.2f\n", result.Confidence)
	fmt.Fprintln(out, "-----------------------------")
	if len(result.Judgements) > 0 {
		fmt.Fprintln(out, "Details:")
		for _, item := range result.Judgements {
			status := "PASS"
			if !item.Passed {
				status = "FAIL"
			}
			fmt.Fprintf(out, "  [%s] %s (score: %.2f) %s\n", status, item.CriteriaName, item.Score, item.Details)
			if item.Error != "" {
				fmt.Fprintf(out, "       error: %s\n", item.Error)
			}
		}
	}
	if len(result.SuggestedFixes) > 0 {
		fmt.Fprintln(out, "Suggested fixes:")
		for _, fix := range result.SuggestedFixes {
			fmt.Fprintf(out, "  - %s\n", fix)
		}
	}
	fmt.Fprintln(out, "=============================")
	return nil
}

func defaultBootstrapCriteria() []strategies.JudgementCriteria {
	return []strategies.JudgementCriteria{
		{
			Name:    "test_pass_rate",
			Type:    strategies.JudgementTypeTest,
			Weight:  0.40,
			Enabled: true,
			Thresholds: map[string]float64{
				"min_pass_rate": judgement.DefaultJudgementThresholds.MinTestPassRate,
			},
		},
		{
			Name:    "coverage_threshold",
			Type:    strategies.JudgementTypeCoverage,
			Weight:  0.40,
			Enabled: true,
			Thresholds: map[string]float64{
				"min_coverage": judgement.DefaultJudgementThresholds.MinCoverage,
			},
		},
		{
			Name:    "syntax_check",
			Type:    strategies.JudgementTypeSyntax,
			Weight:  0.20,
			Enabled: true,
			Thresholds: map[string]float64{
				"min_pass_rate": 1.0,
			},
		},
	}
}
