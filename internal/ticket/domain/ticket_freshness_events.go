package domain

import (
	"encoding/json"
	"fmt"
)

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

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e TicketFlagged) MarshalJSON() ([]byte, error) {
	type proxy struct {
		ReviewID    string      `json:"review_id"`
		TicketID    string      `json:"ticket_id"`
		ContextDiff ContextDiff `json:"context_diff"`
		FlaggedAt   string      `json:"flagged_at"`
	}
	data, err := json.Marshal(proxy{
		ReviewID:    e.reviewID,
		TicketID:    e.ticketID,
		ContextDiff: e.contextDiff,
		FlaggedAt:   e.flaggedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling TicketFlagged: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *TicketFlagged) UnmarshalJSON(data []byte) error {
	type proxy struct {
		ReviewID    string      `json:"review_id"`
		TicketID    string      `json:"ticket_id"`
		ContextDiff ContextDiff `json:"context_diff"`
		FlaggedAt   string      `json:"flagged_at"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling TicketFlagged: %w", err)
	}
	e.reviewID = p.ReviewID
	e.ticketID = p.TicketID
	e.contextDiff = p.ContextDiff
	e.flaggedAt = p.FlaggedAt
	return nil
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e FlagCleared) MarshalJSON() ([]byte, error) {
	type proxy struct {
		ReviewID  string `json:"review_id"`
		TicketID  string `json:"ticket_id"`
		ClearedAt string `json:"cleared_at"`
	}
	data, err := json.Marshal(proxy{
		ReviewID:  e.reviewID,
		TicketID:  e.ticketID,
		ClearedAt: e.clearedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling FlagCleared: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *FlagCleared) UnmarshalJSON(data []byte) error {
	type proxy struct {
		ReviewID  string `json:"review_id"`
		TicketID  string `json:"ticket_id"`
		ClearedAt string `json:"cleared_at"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling FlagCleared: %w", err)
	}
	e.reviewID = p.ReviewID
	e.ticketID = p.TicketID
	e.clearedAt = p.ClearedAt
	return nil
}

// RippleReviewCreated is emitted when a ripple review completes, signaling ticket flagging is done.
type RippleReviewCreated struct {
	reviewID       string
	closedTicketID string
	flaggedCount   int
}

// NewRippleReviewCreated creates a RippleReviewCreated event. Returns error if IDs are empty.
func NewRippleReviewCreated(reviewID, closedTicketID string, flaggedCount int) (RippleReviewCreated, error) {
	if reviewID == "" {
		return RippleReviewCreated{}, fmt.Errorf("review ID cannot be empty")
	}
	if closedTicketID == "" {
		return RippleReviewCreated{}, fmt.Errorf("closed ticket ID cannot be empty")
	}
	return RippleReviewCreated{
		reviewID:       reviewID,
		closedTicketID: closedTicketID,
		flaggedCount:   flaggedCount,
	}, nil
}

// ReviewID returns the review identifier.
func (e RippleReviewCreated) ReviewID() string { return e.reviewID }

// ClosedTicketID returns the ID of the ticket whose closure triggered this review.
func (e RippleReviewCreated) ClosedTicketID() string { return e.closedTicketID }

// FlaggedCount returns the number of tickets flagged in this review.
func (e RippleReviewCreated) FlaggedCount() int { return e.flaggedCount }

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e RippleReviewCreated) MarshalJSON() ([]byte, error) {
	type proxy struct {
		ReviewID       string `json:"review_id"`
		ClosedTicketID string `json:"closed_ticket_id"`
		FlaggedCount   int    `json:"flagged_count"`
	}
	data, err := json.Marshal(proxy{
		ReviewID:       e.reviewID,
		ClosedTicketID: e.closedTicketID,
		FlaggedCount:   e.flaggedCount,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling RippleReviewCreated: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *RippleReviewCreated) UnmarshalJSON(data []byte) error {
	type proxy struct {
		ReviewID       string `json:"review_id"`
		ClosedTicketID string `json:"closed_ticket_id"`
		FlaggedCount   int    `json:"flagged_count"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling RippleReviewCreated: %w", err)
	}
	e.reviewID = p.ReviewID
	e.closedTicketID = p.ClosedTicketID
	e.flaggedCount = p.FlaggedCount
	return nil
}
