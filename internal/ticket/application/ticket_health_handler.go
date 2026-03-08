package application

import (
	"context"
	"fmt"

	"github.com/alty-cli/alty/internal/ticket/domain"
)

// TicketReader is a handler-local interface for reading ticket data.
// Defined where consumed per Go convention.
type TicketReader interface {
	// ReadOpenTickets returns all open tickets.
	ReadOpenTickets(ctx context.Context) ([]domain.OpenTicketData, error)

	// ReadFlags returns freshness flags for a specific ticket.
	ReadFlags(ctx context.Context, ticketID string) ([]domain.FreshnessFlag, error)
}

// TicketHealthHandler queries ticket freshness and builds a health report.
// It reads all open tickets, identifies those needing review, gathers
// their freshness flags, and computes the oldest last-reviewed date.
type TicketHealthHandler struct {
	reader TicketReader
}

// NewTicketHealthHandler creates a new TicketHealthHandler.
func NewTicketHealthHandler(reader TicketReader) *TicketHealthHandler {
	return &TicketHealthHandler{reader: reader}
}

// Report builds a ticket health report.
func (h *TicketHealthHandler) Report(ctx context.Context) (domain.TicketHealthReport, error) {
	openTickets, err := h.reader.ReadOpenTickets(ctx)
	if err != nil {
		return domain.TicketHealthReport{}, fmt.Errorf("reading open tickets: %w", err)
	}

	var flagged []domain.FlaggedTicket
	for _, ticket := range openTickets {
		if hasLabel(ticket.Labels(), "review_needed") {
			flags, err := h.reader.ReadFlags(ctx, ticket.TicketID())
			if err != nil {
				return domain.TicketHealthReport{}, fmt.Errorf("reading flags for ticket %s: %w", ticket.TicketID(), err)
			}
			flagged = append(flagged, domain.NewFlaggedTicket(
				ticket.TicketID(),
				ticket.Title(),
				flags,
				domain.FreshnessStatusReviewNeeded,
			))
		}
	}

	// Find the oldest last_reviewed across ALL open tickets.
	var oldest *string
	for _, ticket := range openTickets {
		lr := ticket.LastReviewed()
		if lr != nil {
			if oldest == nil || *lr < *oldest {
				s := *lr
				oldest = &s
			}
		}
	}

	return domain.NewTicketHealthReport(flagged, len(openTickets), oldest), nil
}

// hasLabel checks if a label is present in a slice.
func hasLabel(labels []string, target string) bool {
	for _, l := range labels {
		if l == target {
			return true
		}
	}
	return false
}
