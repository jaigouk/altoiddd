// Package domain implements the Ticket Pipeline bounded context domain layer.
package domain

import (
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// GeneratedEpic is an epic grouping tickets for one bounded context.
type GeneratedEpic struct {
	epicID             string
	title              string
	description        string
	boundedContextName string
	classification     vo.SubdomainClassification
}

// NewGeneratedEpic creates a GeneratedEpic value object.
func NewGeneratedEpic(
	epicID, title, description, boundedContextName string,
	classification vo.SubdomainClassification,
) GeneratedEpic {
	return GeneratedEpic{
		epicID:             epicID,
		title:              title,
		description:        description,
		boundedContextName: boundedContextName,
		classification:     classification,
	}
}

// EpicID returns the epic identifier.
func (e GeneratedEpic) EpicID() string { return e.epicID }

// Title returns the epic title.
func (e GeneratedEpic) Title() string { return e.title }

// Description returns the epic description.
func (e GeneratedEpic) Description() string { return e.description }

// BoundedContextName returns which bounded context this epic covers.
func (e GeneratedEpic) BoundedContextName() string { return e.boundedContextName }

// Classification returns the subdomain classification.
func (e GeneratedEpic) Classification() vo.SubdomainClassification { return e.classification }

// GeneratedTicket is a generated beads ticket for an aggregate.
type GeneratedTicket struct {
	ticketID           string
	title              string
	description        string
	detailLevel        vo.TicketDetailLevel
	epicID             string
	boundedContextName string
	aggregateName      string
	dependencies       []string
	depth              int
}

// NewGeneratedTicket creates a GeneratedTicket value object.
func NewGeneratedTicket(
	ticketID, title, description string,
	detailLevel vo.TicketDetailLevel,
	epicID, boundedContextName, aggregateName string,
	dependencies []string,
	depth int,
) GeneratedTicket {
	deps := make([]string, len(dependencies))
	copy(deps, dependencies)
	return GeneratedTicket{
		ticketID:           ticketID,
		title:              title,
		description:        description,
		detailLevel:        detailLevel,
		epicID:             epicID,
		boundedContextName: boundedContextName,
		aggregateName:      aggregateName,
		dependencies:       deps,
		depth:              depth,
	}
}

// TicketID returns the ticket identifier.
func (t GeneratedTicket) TicketID() string { return t.ticketID }

// Title returns the ticket title.
func (t GeneratedTicket) Title() string { return t.title }

// Description returns the rendered ticket body.
func (t GeneratedTicket) Description() string { return t.description }

// DetailLevel returns the generation depth.
func (t GeneratedTicket) DetailLevel() vo.TicketDetailLevel { return t.detailLevel }

// EpicID returns the parent epic ID.
func (t GeneratedTicket) EpicID() string { return t.epicID }

// BoundedContextName returns which bounded context this ticket targets.
func (t GeneratedTicket) BoundedContextName() string { return t.boundedContextName }

// AggregateName returns which aggregate this ticket implements.
func (t GeneratedTicket) AggregateName() string { return t.aggregateName }

// Depth returns the dependency depth.
func (t GeneratedTicket) Depth() int { return t.depth }

// Dependencies returns a defensive copy of dependency IDs.
func (t GeneratedTicket) Dependencies() []string {
	out := make([]string, len(t.dependencies))
	copy(out, t.dependencies)
	return out
}

// DependencyOrder is a topologically sorted ticket execution order.
type DependencyOrder struct {
	orderedIDs []string
}

// NewDependencyOrder creates a DependencyOrder value object.
func NewDependencyOrder(orderedIDs []string) DependencyOrder {
	ids := make([]string, len(orderedIDs))
	copy(ids, orderedIDs)
	return DependencyOrder{orderedIDs: ids}
}

// OrderedIDs returns a defensive copy of ordered ticket IDs.
func (d DependencyOrder) OrderedIDs() []string {
	out := make([]string, len(d.orderedIDs))
	copy(out, d.orderedIDs)
	return out
}
