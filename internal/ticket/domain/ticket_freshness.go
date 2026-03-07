package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ContextDiff captures what changed when a ticket was closed.
type ContextDiff struct {
	summary            string
	triggeringTicketID string
	producedAt         string
}

// NewContextDiff creates a ContextDiff with validation.
func NewContextDiff(summary, triggeringTicketID, producedAt string) (ContextDiff, error) {
	if strings.TrimSpace(summary) == "" {
		return ContextDiff{}, fmt.Errorf("ContextDiff summary must not be empty or whitespace-only: %w", domainerrors.ErrInvariantViolation)
	}
	return ContextDiff{
		summary:            summary,
		triggeringTicketID: triggeringTicketID,
		producedAt:         producedAt,
	}, nil
}

// Summary returns the change summary text.
func (d ContextDiff) Summary() string { return d.summary }

// TriggeringTicketID returns the ID of the ticket that triggered this diff.
func (d ContextDiff) TriggeringTicketID() string { return d.triggeringTicketID }

// ProducedAt returns when the diff was produced.
func (d ContextDiff) ProducedAt() string { return d.producedAt }

// FreshnessFlag is a single flag indicating a ticket needs review.
type FreshnessFlag struct {
	contextDiff ContextDiff
	flaggedAt   string
}

// NewFreshnessFlag creates a FreshnessFlag value object.
func NewFreshnessFlag(contextDiff ContextDiff, flaggedAt string) FreshnessFlag {
	return FreshnessFlag{contextDiff: contextDiff, flaggedAt: flaggedAt}
}

// ContextDiff returns the context diff associated with this flag.
func (f FreshnessFlag) ContextDiff() ContextDiff { return f.contextDiff }

// FlaggedAt returns when the flag was raised.
func (f FreshnessFlag) FlaggedAt() string { return f.flaggedAt }

// TicketFreshnessStatus is the freshness status of a ticket.
type TicketFreshnessStatus string

// Freshness status constants.
const (
	FreshnessStatusFresh         TicketFreshnessStatus = "fresh"
	FreshnessStatusReviewNeeded  TicketFreshnessStatus = "review_needed"
	FreshnessStatusNeverReviewed TicketFreshnessStatus = "never_reviewed"
)

// FlaggedTicket is a ticket with pending freshness flags.
type FlaggedTicket struct {
	ticketID string
	title    string
	status   TicketFreshnessStatus
	flags    []FreshnessFlag
}

// NewFlaggedTicket creates a FlaggedTicket value object.
func NewFlaggedTicket(ticketID, title string, flags []FreshnessFlag, status TicketFreshnessStatus) FlaggedTicket {
	f := make([]FreshnessFlag, len(flags))
	copy(f, flags)
	return FlaggedTicket{ticketID: ticketID, title: title, flags: f, status: status}
}

// TicketID returns the ticket identifier.
func (ft FlaggedTicket) TicketID() string { return ft.ticketID }

// Title returns the ticket title.
func (ft FlaggedTicket) Title() string { return ft.title }

// Status returns the freshness status.
func (ft FlaggedTicket) Status() TicketFreshnessStatus { return ft.status }

// FlagCount returns the number of pending flags.
func (ft FlaggedTicket) FlagCount() int { return len(ft.flags) }

// Flags returns a defensive copy of freshness flags.
func (ft FlaggedTicket) Flags() []FreshnessFlag {
	out := make([]FreshnessFlag, len(ft.flags))
	copy(out, ft.flags)
	return out
}

// TicketHealthReport is an aggregate report on ticket freshness.
type TicketHealthReport struct {
	oldestLastReviewed *string
	flaggedTickets     []FlaggedTicket
	totalOpen          int
}

// NewTicketHealthReport creates a TicketHealthReport value object.
func NewTicketHealthReport(flaggedTickets []FlaggedTicket, totalOpen int, oldestLastReviewed *string) TicketHealthReport {
	ft := make([]FlaggedTicket, len(flaggedTickets))
	copy(ft, flaggedTickets)
	var olr *string
	if oldestLastReviewed != nil {
		s := *oldestLastReviewed
		olr = &s
	}
	return TicketHealthReport{flaggedTickets: ft, totalOpen: totalOpen, oldestLastReviewed: olr}
}

// ReviewNeededCount returns the number of tickets needing review.
func (r TicketHealthReport) ReviewNeededCount() int { return len(r.flaggedTickets) }

// HasIssues returns whether any tickets need review.
func (r TicketHealthReport) HasIssues() bool { return r.ReviewNeededCount() > 0 }

// TotalOpen returns the total number of open tickets.
func (r TicketHealthReport) TotalOpen() int { return r.totalOpen }

// OldestLastReviewed returns the oldest last-reviewed date, or nil.
func (r TicketHealthReport) OldestLastReviewed() *string {
	if r.oldestLastReviewed == nil {
		return nil
	}
	s := *r.oldestLastReviewed
	return &s
}

// FreshnessPct returns the percentage of open tickets that are fresh.
func (r TicketHealthReport) FreshnessPct() float64 {
	if r.totalOpen == 0 {
		return 100.0
	}
	return float64(r.totalOpen-r.ReviewNeededCount()) / float64(r.totalOpen) * 100
}

// FreshnessLabel returns a human-readable threshold label.
func (r TicketHealthReport) FreshnessLabel() string {
	pct := r.FreshnessPct()
	if pct >= 90.0 {
		return "healthy"
	}
	if pct >= 70.0 {
		return "acceptable"
	}
	return "action needed"
}

// FlaggedTickets returns a defensive copy of flagged tickets.
func (r TicketHealthReport) FlaggedTickets() []FlaggedTicket {
	out := make([]FlaggedTicket, len(r.flaggedTickets))
	copy(out, r.flaggedTickets)
	return out
}

// EpicHealthSummary is an epic-level breakdown of ticket freshness.
type EpicHealthSummary struct {
	epicID       string
	totalTickets int
	freshCount   int
	staleCount   int
}

// NewEpicHealthSummary creates an EpicHealthSummary with validation.
func NewEpicHealthSummary(epicID string, totalTickets, freshCount, staleCount int) (EpicHealthSummary, error) {
	if totalTickets < 0 || freshCount < 0 || staleCount < 0 {
		return EpicHealthSummary{}, fmt.Errorf("counts must be non-negative: %w", domainerrors.ErrInvariantViolation)
	}
	if freshCount+staleCount != totalTickets {
		return EpicHealthSummary{}, fmt.Errorf(
			"fresh_count (%d) + stale_count (%d) must equal total_tickets (%d): %w",
			freshCount, staleCount, totalTickets, domainerrors.ErrInvariantViolation,
		)
	}
	return EpicHealthSummary{epicID: epicID, totalTickets: totalTickets, freshCount: freshCount, staleCount: staleCount}, nil
}

// EpicID returns the epic identifier.
func (s EpicHealthSummary) EpicID() string { return s.epicID }

// TotalTickets returns the total number of tickets in this epic.
func (s EpicHealthSummary) TotalTickets() int { return s.totalTickets }

// FreshCount returns the number of fresh tickets.
func (s EpicHealthSummary) FreshCount() int { return s.freshCount }

// StaleCount returns the number of stale tickets.
func (s EpicHealthSummary) StaleCount() int { return s.staleCount }

// FreshnessPct returns the percentage of tickets that are fresh.
func (s EpicHealthSummary) FreshnessPct() float64 {
	if s.totalTickets == 0 {
		return 100.0
	}
	return float64(s.freshCount) / float64(s.totalTickets) * 100
}

// OpenTicketData is raw data for an open ticket used by the reader protocol.
type OpenTicketData struct {
	lastReviewed *string
	ticketID     string
	title        string
	labels       []string
}

// NewOpenTicketData creates an OpenTicketData value object.
func NewOpenTicketData(ticketID, title string, labels []string, lastReviewed *string) OpenTicketData {
	l := make([]string, len(labels))
	copy(l, labels)
	var lr *string
	if lastReviewed != nil {
		s := *lastReviewed
		lr = &s
	}
	return OpenTicketData{ticketID: ticketID, title: title, labels: l, lastReviewed: lr}
}

// TicketID returns the ticket identifier.
func (o OpenTicketData) TicketID() string { return o.ticketID }

// Title returns the ticket title.
func (o OpenTicketData) Title() string { return o.title }

// Labels returns a defensive copy of labels.
func (o OpenTicketData) Labels() []string {
	out := make([]string, len(o.labels))
	copy(out, o.labels)
	return out
}

// LastReviewed returns the last review date, or nil.
func (o OpenTicketData) LastReviewed() *string {
	if o.lastReviewed == nil {
		return nil
	}
	s := *o.lastReviewed
	return &s
}
