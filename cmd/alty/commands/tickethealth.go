package commands

import (
	"context"
	"fmt"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/spf13/cobra"
)

// NewTicketHealthCmd creates the "alty ticket-health" command.
func NewTicketHealthCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "ticket-health",
		Short: "Show ripple review report for tickets needing attention",
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := app.TicketHealthHandler.Report(context.Background())
			if err != nil {
				return fmt.Errorf("ticket health: %w", err)
			}

			fmt.Printf("Ticket Freshness: %.1f%% (%s)\n",
				report.FreshnessPct(), report.FreshnessLabel())
			fmt.Printf("Open: %d  Flagged: %d\n",
				report.TotalOpen(), report.ReviewNeededCount())

			flagged := report.FlaggedTickets()
			if len(flagged) > 0 {
				fmt.Println("\nFlagged tickets:")
				for _, ft := range flagged {
					fmt.Printf("  %s - %s\n", ft.TicketID(), ft.Title())
				}
			}

			if report.HasIssues() {
				return fmt.Errorf("tickets need review")
			}

			return nil
		},
	}
}
