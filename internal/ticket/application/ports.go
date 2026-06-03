// Package application defines ports for the Ticket bounded context.
package application

import (
	"context"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
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

// PortScanner scans Go source files for interface definitions via AST.
// Used by ticket validation to verify port signatures against actual code.
type PortScanner interface {
	// ScanPorts scans a directory for Go interface definitions and returns them keyed by name.
	ScanPorts(portsDir string) map[string]ticketdomain.ScannedPort
}

// BeadsGraphTicket is the minimal projection of a beads ticket the ripple
// flow needs: just the identifier and its open/closed status. Adapters
// translate the raw bd JSON shape into this VO so the handler stays
// ignorant of beads field names.
type BeadsGraphTicket struct {
	ID     string
	Status string
}

// IsOpen reports whether the ticket is in an actionable, not-yet-closed
// state. "open" and "in_progress" both qualify, mirroring the legacy bash
// ripple semantics at alto-scaffold/scripts/bd-ripple.
func (t BeadsGraphTicket) IsOpen() bool {
	return t.Status == "open" || t.Status == "in_progress"
}

// BeadsGraphReader reads dependency-graph data for a beads ticket. It is
// the narrow port the ripple-review handler depends on; adapters wrap
// `bd show --json` and `bd children --json` and translate the raw beads
// JSON to BeadsGraphTicket VOs. Returned slices include both open and
// closed tickets — the handler is responsible for filtering open ones and
// deduplicating across siblings/dependents/related.
type BeadsGraphReader interface {
	// ReadParent returns the parent ticket ID, or "" when the ticket has
	// no parent-child dependency. Adapters MUST NOT return an error for
	// the "no parent" case — only for transport / parse failures.
	ReadParent(ctx context.Context, ticketID string) (string, error)

	// ReadSiblings returns the children of parentID excluding selfID.
	// The raw list (open + closed) is returned; the handler filters.
	ReadSiblings(ctx context.Context, parentID, selfID string) ([]BeadsGraphTicket, error)

	// ReadDependents returns tickets that have a `blocks` dependency on
	// ticketID — i.e. tickets that ticketID blocks. The raw list (open +
	// closed) is returned; the handler filters.
	ReadDependents(ctx context.Context, ticketID string) ([]BeadsGraphTicket, error)

	// ReadRelated returns tickets linked by bidirectional `related`
	// edges (either direction). The raw list (open + closed) is returned;
	// the handler filters.
	ReadRelated(ctx context.Context, ticketID string) ([]BeadsGraphTicket, error)

	// ReadCloseContext returns the ticket's close_reason field, used as
	// the ContextDiff summary when no explicit override is supplied.
	// Returns "" when no close_reason is present.
	ReadCloseContext(ctx context.Context, ticketID string) (string, error)
}

// BeadsCommentWriter posts a free-form comment to a beads ticket. Used by
// the ripple-review handler to attach the structured "Ripple review
// needed" message produced by RippleReview.BuildRippleComment().
type BeadsCommentWriter interface {
	// AddComment posts comment as a new comment on ticketID. Adapters
	// MUST enforce a per-invocation timeout and refuse empty inputs.
	AddComment(ctx context.Context, ticketID, comment string) error
}
