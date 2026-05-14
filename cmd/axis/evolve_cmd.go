package main

import (
	"encoding/json"
	"fmt"

	"github.com/axis-cli/axis/internal/evolution"
	"github.com/spf13/cobra"
)

func newEvolveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evolve",
		Short: "Sandboxed evolution protocol for Axis self-modification",
	}
	cmd.AddCommand(newEvolveInspectCommand(), newEvolvePromoteCommand(), newEvolveDiscardCommand())
	return cmd
}

func newEvolveInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect [run-id]",
		Short: "Inspect an evolution run (read-only)",
		Args:  cobra.ExactArgs(1),
		RunE:  evolveInspect,
	}
}

func evolveInspect(cmd *cobra.Command, args []string) error {
	runID := args[0]
	store, err := evolution.NewStore(defaultApp.resolvedRoot())
	if err != nil {
		return fmt.Errorf("open evolution store: %w", err)
	}

	result := make(map[string]any)

	intent, err := store.ReadIntent(runID)
	if err == nil {
		result["intent"] = intent
	} else {
		result["intent_error"] = err.Error()
	}

	run, err := store.ReadRun(runID)
	if err == nil {
		result["run"] = run
	} else {
		result["run_error"] = err.Error()
	}

	ledger := evolution.NewLedger(store.RunDir(runID))
	steps, ledgerErrs, err := ledger.ReadSteps()
	if err == nil {
		result["steps"] = steps
		if len(ledgerErrs) > 0 {
			var errs []string
			for _, e := range ledgerErrs {
				errs = append(errs, e.Error())
			}
			result["ledger_errors"] = errs
		}
	} else {
		result["steps_error"] = err.Error()
	}

	verifier := evolution.NewVerifier(store)
	record, err := verifier.ReadVerification(runID)
	if err == nil {
		result["verification"] = record
	} else {
		result["verification_error"] = err.Error()
	}

	decision, err := store.ReadDecision(runID)
	if err == nil {
		result["decision"] = decision
	} else {
		result["decision_error"] = err.Error()
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal inspect result: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func newEvolvePromoteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "promote [run-id]",
		Short: "Promote an evolution run to the main project tree",
		Args:  cobra.ExactArgs(1),
		RunE:  evolvePromote,
	}
}

func evolvePromote(cmd *cobra.Command, args []string) error {
	runID := args[0]
	store, err := evolution.NewStore(defaultApp.resolvedRoot())
	if err != nil {
		return fmt.Errorf("open evolution store: %w", err)
	}
	gate := evolution.NewDecisionGate(store)
	decision, err := gate.Promote(runID, ".", "cli", "promoted via axis evolve promote")
	if err != nil {
		return err
	}
	fmt.Printf("Run %s promoted successfully\n", decision.RunID)
	return nil
}

func newEvolveDiscardCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "discard [run-id]",
		Short: "Discard an evolution run without promoting its changes",
		Args:  cobra.ExactArgs(1),
		RunE:  evolveDiscard,
	}
}

func evolveDiscard(cmd *cobra.Command, args []string) error {
	runID := args[0]
	store, err := evolution.NewStore(defaultApp.resolvedRoot())
	if err != nil {
		return fmt.Errorf("open evolution store: %w", err)
	}
	gate := evolution.NewDecisionGate(store)
	decision, err := gate.Discard(runID, "cli", "discarded via axis evolve discard")
	if err != nil {
		return err
	}
	fmt.Printf("Run %s discarded successfully\n", decision.RunID)
	return nil
}
