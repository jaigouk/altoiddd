package domain

import (
	"fmt"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ReviewChecklistTemplate is the standard review checklist for ripple reviews.
const ReviewChecklistTemplate = `**Review checklist:**
- [ ] Read the description -- does it still match the new context?
- [ ] Check acceptance criteria -- are any obsolete, incomplete, or contradicted?
- [ ] Verify DDD alignment -- do bounded-context boundaries still hold?
- [ ] Update or dismiss -- apply changes, or mark as unchanged if still valid.`

// RippleReview is the aggregate root for managing freshness flags.
type RippleReview struct {
	reviewID         string
	closedTicketID   string
	contextDiff      ContextDiff
	flaggedTicketIDs []string
	events           []any
	completed        bool
}

// NewRippleReview creates a RippleReview aggregate root.
func NewRippleReview(reviewID, closedTicketID string, contextDiff ContextDiff) *RippleReview {
	return &RippleReview{
		reviewID:       reviewID,
		closedTicketID: closedTicketID,
		contextDiff:    contextDiff,
	}
}

// ReviewID returns the review identifier.
func (r *RippleReview) ReviewID() string { return r.reviewID }

// ClosedTicketID returns the ID of the ticket whose closure triggered this review.
func (r *RippleReview) ClosedTicketID() string { return r.closedTicketID }

// ContextDiff returns the context diff describing what changed.
func (r *RippleReview) ContextDiff() ContextDiff { return r.contextDiff }

// FlaggedTickets returns a defensive copy of currently flagged ticket IDs.
func (r *RippleReview) FlaggedTickets() []string {
	out := make([]string, len(r.flaggedTicketIDs))
	copy(out, r.flaggedTicketIDs)
	return out
}

// Events returns a defensive copy of domain events.
func (r *RippleReview) Events() []any {
	out := make([]any, len(r.events))
	copy(out, r.events)
	return out
}

// FlagTicket flags a ticket for review. Only open tickets can be flagged.
func (r *RippleReview) FlagTicket(ticketID string, isOpen bool, flaggedAt string) error {
	if !isOpen {
		return fmt.Errorf("only open tickets can be flagged; '%s' is not open: %w",
			ticketID, domainerrors.ErrInvariantViolation)
	}
	r.flaggedTicketIDs = append(r.flaggedTicketIDs, ticketID)
	r.events = append(r.events, NewTicketFlagged(r.reviewID, ticketID, r.contextDiff, flaggedAt))
	return nil
}

// BuildRippleComment builds a structured ripple review comment.
func (r *RippleReview) BuildRippleComment() string {
	return fmt.Sprintf(
		"**Ripple review needed** -- `%s` was closed.\n\n**What changed:** %s\n\n%s",
		r.closedTicketID,
		r.contextDiff.Summary(),
		ReviewChecklistTemplate,
	)
}

// ClearFlag clears a freshness flag after explicit review.
func (r *RippleReview) ClearFlag(ticketID string, clearedAt string) error {
	for i, id := range r.flaggedTicketIDs {
		if id == ticketID {
			r.flaggedTicketIDs = append(r.flaggedTicketIDs[:i], r.flaggedTicketIDs[i+1:]...)
			r.events = append(r.events, NewFlagCleared(r.reviewID, ticketID, clearedAt))
			return nil
		}
	}
	return fmt.Errorf("ticket '%s' is not flagged and cannot be cleared: %w",
		ticketID, domainerrors.ErrInvariantViolation)
}

// Complete finalizes the ripple review and emits RippleReviewCreated event.
// Can only be called once per review.
func (r *RippleReview) Complete() error {
	if r.completed {
		return fmt.Errorf("review '%s' is already completed: %w",
			r.reviewID, domainerrors.ErrInvariantViolation)
	}
	r.completed = true
	event, err := NewRippleReviewCreated(r.reviewID, r.closedTicketID, len(r.flaggedTicketIDs))
	if err != nil {
		return fmt.Errorf("creating RippleReviewCreated event: %w", err)
	}
	r.events = append(r.events, event)
	return nil
}
