package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
)

// NewTicketVerifyCmd creates the "alto ticket verify" command.
func NewTicketVerifyCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "ticket-verify <ticket-id>",
		Short: "Verify quantitative claims in a ticket",
		Long: `Parses a ticket for quantitative claims (e.g., "14 findings") and
verifies them by running associated commands.

Claims are detected from bold number patterns (e.g., **14 issues**) and
verified using commands from code blocks in the ticket.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ticketID := args[0]

			results, err := app.TicketVerifyHandler.Verify(context.Background(), ticketID)
			if err != nil {
				return fmt.Errorf("verifying ticket: %w", err)
			}

			if len(results) == 0 {
				fmt.Println("No verifiable claims found in ticket.")
				return nil
			}

			var mismatches int
			for _, r := range results {
				claim := r.Claim()
				if r.Match() {
					fmt.Printf("✓ VERIFIED: %s = %s\n", claim.ClaimText(), r.ActualValue())
				} else {
					fmt.Printf("✗ MISMATCH: %s\n", claim.ClaimText())
					fmt.Printf("  %s\n", r.Discrepancy())
					mismatches++
				}
			}

			fmt.Printf("\nTotal: %d claims, %d mismatches\n", len(results), mismatches)

			if mismatches > 0 {
				return fmt.Errorf("%d claim(s) failed verification", mismatches)
			}

			return nil
		},
	}
}
