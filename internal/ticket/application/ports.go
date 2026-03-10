// Package application defines ports for the Ticket bounded context.
package application

import (
	"context"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

// BeadsWriter writes beads tickets and epics to the issue tracker.
type BeadsWriter interface {
	// WriteEpic writes an epic to the issue tracker and returns the assigned ID.
	WriteEpic(ctx context.Context, epic ticketdomain.GeneratedEpic) (string, error)

	// WriteTicket writes a ticket to the issue tracker and returns the assigned ID.
	WriteTicket(ctx context.Context, ticket ticketdomain.GeneratedTicket) (string, error)

	// SetDependency sets a dependency between two tickets.
	SetDependency(ctx context.Context, ticketID string, dependsOnID string) error
}

// TicketGeneration generates dependency-ordered beads tickets from DDD artifacts
// with complexity-budget-driven detail levels.
type TicketGeneration interface {
	// Generate generates beads tickets from a domain model.
	Generate(ctx context.Context, model *ddd.DomainModel, outputDir string) error
}

// TicketHealth reports on ticket staleness and ripple review status
// across the project backlog.
type TicketHealth interface {
	// Report generates a ticket health report.
	Report(ctx context.Context, projectDir string) (ticketdomain.TicketHealthReport, error)
}

// CommandRunner executes verification commands and returns their output.
// Implementations must enforce security controls (allowlist, timeout, no shell expansion).
type CommandRunner interface {
	// Run executes a command and returns its stdout output.
	// Returns error if command fails, times out, or is not in allowlist.
	Run(ctx context.Context, command string) (string, error)
}

// TicketContentReader reads ticket markdown content for claim verification.
type TicketContentReader interface {
	// ReadTicketContent reads the full description/content of a ticket.
	ReadTicketContent(ctx context.Context, ticketID string) (string, error)
}

// BeadsLabelWriter manages labels on beads tickets.
type BeadsLabelWriter interface {
	// AddLabel adds a label to a ticket.
	AddLabel(ctx context.Context, ticketID, label string) error

	// RemoveLabel removes a label from a ticket.
	RemoveLabel(ctx context.Context, ticketID, label string) error
}
