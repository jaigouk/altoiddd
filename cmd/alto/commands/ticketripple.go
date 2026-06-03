package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
)

// NewTicketRippleCmd creates the "alto ticket-ripple <id>" command. It
// productizes the bash `alto-scaffold/scripts/bd-ripple` workflow as a Go
// subcommand backed by the Ticket bounded context's RippleHandler.
func NewTicketRippleCmd(app *composition.App) *cobra.Command {
	var contextOverride string

	cmd := &cobra.Command{
		Use:   "ticket-ripple <ticket-id>",
		Short: "Flag open dependents of a closed ticket for review",
		Long: `Walks the dependency graph of <ticket-id> (siblings, dependents, related) and
applies the review_needed label + a structured ripple comment to every open
neighbour, so a downstream agent or operator notices that the upstream change
may have invalidated their assumptions.

The context summary used in the ripple comment comes from --context when
supplied; otherwise the closed ticket's close_reason is used. An empty
summary aborts with an error — the DDD invariant on ContextDiff forbids
whitespace-only inputs.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ticketID := args[0]
			var override *string
			if cmd.Flags().Changed("context") {
				v := contextOverride
				override = &v
			}

			result, err := app.RippleHandler.Handle(cmd.Context(), ticketID, override)
			if err != nil {
				return fmt.Errorf("ripple review: %w", err)
			}

			if result.FlaggedCount == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No open siblings or dependents found for %s.\n", ticketID)
				return nil
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Flagged %d ticket(s) for review (events: %d).\n",
				result.FlaggedCount, result.EventCount)
			return nil
		},
	}

	cmd.Flags().StringVar(&contextOverride, "context", "",
		"Override the ContextDiff summary (defaults to the closed ticket's close_reason)")

	return cmd
}
