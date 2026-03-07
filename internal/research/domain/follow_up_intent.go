package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// FollowUpIntent is a concrete ticket idea discovered during a spike.
type FollowUpIntent struct {
	title       string
	description string
}

// NewFollowUpIntent creates a validated FollowUpIntent.
func NewFollowUpIntent(title, description string) (FollowUpIntent, error) {
	if strings.TrimSpace(title) == "" {
		return FollowUpIntent{}, fmt.Errorf(
			"followUpIntent title must not be empty or whitespace-only: %w",
			domainerrors.ErrInvariantViolation)
	}
	return FollowUpIntent{title: title, description: description}, nil
}

// Title returns the follow-up intent title.
func (f FollowUpIntent) Title() string { return f.title }

// Description returns the follow-up intent description.
func (f FollowUpIntent) Description() string { return f.description }

// FollowUpAuditResult is the result of auditing spike follow-ups against created tickets.
type FollowUpAuditResult struct {
	spikeID          string
	reportPath       string
	definedIntents   []FollowUpIntent
	matchedTicketIDs []string
	orphanedIntents  []FollowUpIntent
}

// NewFollowUpAuditResult creates a FollowUpAuditResult.
func NewFollowUpAuditResult(
	spikeID, reportPath string,
	definedIntents []FollowUpIntent,
	matchedTicketIDs []string,
	orphanedIntents []FollowUpIntent,
) FollowUpAuditResult {
	di := make([]FollowUpIntent, len(definedIntents))
	copy(di, definedIntents)
	mt := make([]string, len(matchedTicketIDs))
	copy(mt, matchedTicketIDs)
	oi := make([]FollowUpIntent, len(orphanedIntents))
	copy(oi, orphanedIntents)
	return FollowUpAuditResult{
		spikeID:          spikeID,
		reportPath:       reportPath,
		definedIntents:   di,
		matchedTicketIDs: mt,
		orphanedIntents:  oi,
	}
}

// SpikeID returns the spike identifier.
func (r FollowUpAuditResult) SpikeID() string { return r.spikeID }

// ReportPath returns the spike report path.
func (r FollowUpAuditResult) ReportPath() string { return r.reportPath }

// DefinedIntents returns a defensive copy.
func (r FollowUpAuditResult) DefinedIntents() []FollowUpIntent {
	out := make([]FollowUpIntent, len(r.definedIntents))
	copy(out, r.definedIntents)
	return out
}

// MatchedTicketIDs returns a defensive copy.
func (r FollowUpAuditResult) MatchedTicketIDs() []string {
	out := make([]string, len(r.matchedTicketIDs))
	copy(out, r.matchedTicketIDs)
	return out
}

// OrphanedIntents returns a defensive copy.
func (r FollowUpAuditResult) OrphanedIntents() []FollowUpIntent {
	out := make([]FollowUpIntent, len(r.orphanedIntents))
	copy(out, r.orphanedIntents)
	return out
}

// DefinedCount returns the number of follow-up intents defined in the spike report.
func (r FollowUpAuditResult) DefinedCount() int { return len(r.definedIntents) }

// OrphanedCount returns the number of intents with no corresponding ticket.
func (r FollowUpAuditResult) OrphanedCount() int { return len(r.orphanedIntents) }

// HasOrphans returns whether any defined intents are missing corresponding tickets.
func (r FollowUpAuditResult) HasOrphans() bool { return r.OrphanedCount() > 0 }
