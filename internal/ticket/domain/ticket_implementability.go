package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// FindingSeverity is the severity of an implementability finding.
type FindingSeverity string

// Finding severity constants.
const (
	FindingSeverityCritical FindingSeverity = "critical"
	FindingSeverityMajor    FindingSeverity = "major"
	FindingSeverityMinor    FindingSeverity = "minor"
)

// ImplementabilityFinding is a single implementability issue found during validation.
type ImplementabilityFinding struct {
	severity    FindingSeverity
	location    string
	description string
}

// NewImplementabilityFinding creates an ImplementabilityFinding with validation.
func NewImplementabilityFinding(severity FindingSeverity, location, description string) (ImplementabilityFinding, error) {
	if strings.TrimSpace(description) == "" {
		return ImplementabilityFinding{}, fmt.Errorf("ImplementabilityFinding description must not be empty: %w", domainerrors.ErrInvariantViolation)
	}
	return ImplementabilityFinding{severity: severity, location: location, description: description}, nil
}

// Severity returns the finding severity level.
func (f ImplementabilityFinding) Severity() FindingSeverity { return f.severity }

// Location returns where in the ticket the finding was detected.
func (f ImplementabilityFinding) Location() string { return f.location }

// Description returns the finding description.
func (f ImplementabilityFinding) Description() string { return f.description }

// InterfaceMismatch is a port/method signature mismatch between ticket sections.
type InterfaceMismatch struct {
	sectionA    string
	sectionB    string
	description string
}

// NewInterfaceMismatch creates an InterfaceMismatch value object.
func NewInterfaceMismatch(sectionA, sectionB, description string) InterfaceMismatch {
	return InterfaceMismatch{sectionA: sectionA, sectionB: sectionB, description: description}
}

// SectionA returns the first section involved in the mismatch.
func (m InterfaceMismatch) SectionA() string { return m.sectionA }

// SectionB returns the second section involved in the mismatch.
func (m InterfaceMismatch) SectionB() string { return m.sectionB }

// Description returns the mismatch description.
func (m InterfaceMismatch) Description() string { return m.description }

// UnresolvedDependency is a reference to a port/library that does not exist.
type UnresolvedDependency struct {
	portName    string
	location    string
	description string
}

// NewUnresolvedDependency creates an UnresolvedDependency with validation.
func NewUnresolvedDependency(portName, location, description string) (UnresolvedDependency, error) {
	if strings.TrimSpace(portName) == "" {
		return UnresolvedDependency{}, fmt.Errorf("UnresolvedDependency port_name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}
	return UnresolvedDependency{portName: portName, location: location, description: description}, nil
}

// PortName returns the unresolved port or library name.
func (u UnresolvedDependency) PortName() string { return u.portName }

// Location returns where the unresolved dependency was found.
func (u UnresolvedDependency) Location() string { return u.location }

// Description returns the dependency issue description.
func (u UnresolvedDependency) Description() string { return u.description }

// TicketSection is a parsed section of a ticket description.
type TicketSection struct {
	heading string
	content string
}

// NewTicketSection creates a TicketSection value object.
func NewTicketSection(heading, content string) TicketSection {
	return TicketSection{heading: heading, content: content}
}

// Heading returns the section heading.
func (s TicketSection) Heading() string { return s.heading }

// Content returns the section content.
func (s TicketSection) Content() string { return s.content }

// TicketStructure is a full parsed ticket as a collection of TicketSections.
type TicketStructure struct {
	sections []TicketSection
}

// NewTicketStructure creates a TicketStructure value object.
func NewTicketStructure(sections []TicketSection) TicketStructure {
	s := make([]TicketSection, len(sections))
	copy(s, sections)
	return TicketStructure{sections: s}
}

// GetSection looks up a section by heading. Returns nil if not found.
func (ts TicketStructure) GetSection(heading string) *TicketSection {
	for _, s := range ts.sections {
		if s.heading == heading {
			return &s
		}
	}
	return nil
}

// DesignTraceResult is a structured validation result with findings list.
type DesignTraceResult struct {
	ticketID string
	findings []ImplementabilityFinding
}

// NewDesignTraceResult creates a DesignTraceResult value object.
func NewDesignTraceResult(ticketID string, findings []ImplementabilityFinding) DesignTraceResult {
	f := make([]ImplementabilityFinding, len(findings))
	copy(f, findings)
	return DesignTraceResult{ticketID: ticketID, findings: f}
}

// TicketID returns the ticket identifier.
func (r DesignTraceResult) TicketID() string { return r.ticketID }

// IsValid returns whether the ticket passed all implementability checks.
func (r DesignTraceResult) IsValid() bool { return len(r.findings) == 0 }

// CriticalCount returns the number of CRITICAL findings.
func (r DesignTraceResult) CriticalCount() int {
	count := 0
	for _, f := range r.findings {
		if f.severity == FindingSeverityCritical {
			count++
		}
	}
	return count
}

// Findings returns a defensive copy of findings.
func (r DesignTraceResult) Findings() []ImplementabilityFinding {
	out := make([]ImplementabilityFinding, len(r.findings))
	copy(out, r.findings)
	return out
}
