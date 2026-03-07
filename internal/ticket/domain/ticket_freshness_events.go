package domain

// TicketFlagged is emitted when a ticket is flagged for review due to an upstream change.
type TicketFlagged struct {
	reviewID    string
	ticketID    string
	contextDiff ContextDiff
	flaggedAt   string
}

// NewTicketFlagged creates a TicketFlagged event.
func NewTicketFlagged(reviewID, ticketID string, contextDiff ContextDiff, flaggedAt string) TicketFlagged {
	return TicketFlagged{
		reviewID:    reviewID,
		ticketID:    ticketID,
		contextDiff: contextDiff,
		flaggedAt:   flaggedAt,
	}
}

// ReviewID returns the review identifier.
func (e TicketFlagged) ReviewID() string { return e.reviewID }

// TicketID returns the flagged ticket identifier.
func (e TicketFlagged) TicketID() string { return e.ticketID }

// ContextDiff returns the context diff that triggered the flag.
func (e TicketFlagged) ContextDiff() ContextDiff { return e.contextDiff }

// FlaggedAt returns the timestamp when the ticket was flagged.
func (e TicketFlagged) FlaggedAt() string { return e.flaggedAt }

// FlagCleared is emitted when a freshness flag is cleared after explicit review.
type FlagCleared struct {
	reviewID  string
	ticketID  string
	clearedAt string
}

// NewFlagCleared creates a FlagCleared event.
func NewFlagCleared(reviewID, ticketID, clearedAt string) FlagCleared {
	return FlagCleared{reviewID: reviewID, ticketID: ticketID, clearedAt: clearedAt}
}

// ReviewID returns the review identifier.
func (e FlagCleared) ReviewID() string { return e.reviewID }

// TicketID returns the ticket whose flag was cleared.
func (e FlagCleared) TicketID() string { return e.ticketID }

// ClearedAt returns the timestamp when the flag was cleared.
func (e FlagCleared) ClearedAt() string { return e.clearedAt }
