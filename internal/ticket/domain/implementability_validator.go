package domain

import (
	"fmt"
	"regexp"
	"strings"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Section headings expected per detail level.
var (
	FullSections = []string{
		"## Goal",
		"## DDD Alignment",
		"## Design",
		"## SOLID Mapping",
		"## TDD Workflow",
		"## Steps",
		"## Acceptance Criteria",
		"## Edge Cases",
	}
	StandardSections = []string{
		"## Goal",
		"## DDD Alignment",
		"## Steps",
		"## Acceptance Criteria",
	}
)

// Patterns that suggest unspecified dependencies.
var vagueActionRE = regexp.MustCompile(
	`(?i)\b(?:adapter|service|handler)\s+` +
		`(?:performs?|does|executes?|runs?|handles?)\s+` +
		`(?:iterative\s+)?(?:web\s+)?(?:search|lookup|query|fetch|scan|call)`,
)

// AC checkbox pattern.
var acCheckboxRE = regexp.MustCompile(`(?m)^- \[ \]`)

// ValidateImplementability validates a single ticket's implementability.
func ValidateImplementability(ticket GeneratedTicket) DesignTraceResult {
	if ticket.DetailLevel() == vo.TicketDetailStub {
		return NewDesignTraceResult(ticket.TicketID(), nil)
	}

	var findings []ImplementabilityFinding
	desc := ticket.Description()

	findings = append(findings, checkSectionPresence(ticket)...)
	findings = append(findings, checkUnspecifiedDependencies(desc)...)
	if ticket.DetailLevel() == vo.TicketDetailFull {
		findings = append(findings, checkEmptyInvariants(desc)...)
	}
	findings = append(findings, checkEmptyAcceptanceCriteria(desc)...)

	return NewDesignTraceResult(ticket.TicketID(), findings)
}

// ValidateImplementabilityPlan validates all tickets and returns one result per ticket.
func ValidateImplementabilityPlan(tickets []GeneratedTicket) []DesignTraceResult {
	results := make([]DesignTraceResult, len(tickets))
	for i, t := range tickets {
		results[i] = ValidateImplementability(t)
	}
	return results
}

func checkSectionPresence(ticket GeneratedTicket) []ImplementabilityFinding {
	expected := FullSections
	if ticket.DetailLevel() != vo.TicketDetailFull {
		expected = StandardSections
	}
	var findings []ImplementabilityFinding
	for _, heading := range expected {
		if !strings.Contains(ticket.Description(), heading) {
			f, _ := NewImplementabilityFinding(
				FindingSeverityMajor,
				heading,
				fmt.Sprintf("Missing section: %s", heading),
			)
			findings = append(findings, f)
		}
	}
	return findings
}

func checkUnspecifiedDependencies(description string) []ImplementabilityFinding {
	var findings []ImplementabilityFinding
	for _, match := range vagueActionRE.FindAllString(description, -1) {
		f, _ := NewImplementabilityFinding(
			FindingSeverityCritical,
			"Design",
			fmt.Sprintf("Unspecified dependency: '%s' — which port or library implements this?", match),
		)
		findings = append(findings, f)
	}
	return findings
}

func checkEmptyInvariants(description string) []ImplementabilityFinding {
	if !strings.Contains(description, "### Invariants") {
		f, _ := NewImplementabilityFinding(
			FindingSeverityMajor,
			"## Design",
			"No invariant subsection found in Design — FULL tickets should specify domain invariants",
		)
		return []ImplementabilityFinding{f}
	}
	return nil
}

func checkEmptyAcceptanceCriteria(description string) []ImplementabilityFinding {
	if !acCheckboxRE.MatchString(description) {
		f, _ := NewImplementabilityFinding(
			FindingSeverityMajor,
			"## Acceptance Criteria",
			"No acceptance criteria checkboxes found — tickets must have testable AC items (- [ ] ...)",
		)
		return []ImplementabilityFinding{f}
	}
	return nil
}
