package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// NewCheckCmd creates the "alty check" command.
func NewCheckCmd(app *composition.App) *cobra.Command {
	var gate string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run quality gates (lint, types, tests, fitness)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var gates []vo.QualityGate
			if gate != "" {
				g := vo.QualityGate(gate)
				valid := false
				for _, ag := range vo.AllQualityGates() {
					if ag == g {
						valid = true
						break
					}
				}
				if !valid {
					names := make([]string, 0)
					for _, ag := range vo.AllQualityGates() {
						names = append(names, string(ag))
					}
					return fmt.Errorf("invalid gate %q, valid: %v", gate, names)
				}
				gates = []vo.QualityGate{g}
			}

			report, err := app.QualityGateHandler.Check(context.Background(), gates)
			if err != nil {
				return fmt.Errorf("quality check: %w", err)
			}

			results := report.Results()
			for _, r := range results {
				status := "PASS"
				if !r.Passed() {
					status = "FAIL"
				}
				fmt.Printf("  [%s] %s (%dms)\n", status, r.Gate(), r.DurationMS())
				if !r.Passed() {
					fmt.Printf("         %s\n", r.Output())
				}
			}

			if report.Passed() {
				fmt.Printf("\nAll %d quality gate(s) passed.\n", len(results))
			} else {
				failed := 0
				for _, r := range results {
					if !r.Passed() {
						failed++
					}
				}
				return fmt.Errorf("%d quality gate(s) failed", failed)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&gate, "gate", "", "Run a specific gate: lint, types, tests, fitness")

	return cmd
}
