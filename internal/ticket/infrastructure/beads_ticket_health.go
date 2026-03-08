package infrastructure

import (
	"context"
	"path/filepath"

	ticketapp "github.com/alty-cli/alty/internal/ticket/application"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

// BeadsTicketHealthAdapter implements TicketHealth by constructing a
// BeadsTicketReader for the given project directory and delegating
// report generation to TicketHealthHandler.
type BeadsTicketHealthAdapter struct{}

// Compile-time interface check.
var _ ticketapp.TicketHealth = (*BeadsTicketHealthAdapter)(nil)

// Report generates a ticket health report for the project.
func (a *BeadsTicketHealthAdapter) Report(
	ctx context.Context,
	projectDir string,
) (ticketdomain.TicketHealthReport, error) {
	beadsDir := filepath.Join(projectDir, ".beads")
	reader := NewBeadsTicketReader(beadsDir)

	tickets := reader.ReadOpenTickets(ctx)
	totalOpen := len(tickets)

	var flaggedTickets []ticketdomain.FlaggedTicket
	for _, t := range tickets {
		labels := t.Labels()
		hasReviewNeeded := false
		for _, l := range labels {
			if l == "review_needed" {
				hasReviewNeeded = true
				break
			}
		}
		if hasReviewNeeded {
			flags := reader.ReadFlags(ctx, t.TicketID())
			status := ticketdomain.FreshnessStatusReviewNeeded
			if len(flags) == 0 {
				status = ticketdomain.FreshnessStatusNeverReviewed
			}
			ft := ticketdomain.NewFlaggedTicket(t.TicketID(), t.Title(), flags, status)
			flaggedTickets = append(flaggedTickets, ft)
		}
	}

	return ticketdomain.NewTicketHealthReport(flaggedTickets, totalOpen, nil), nil
}
