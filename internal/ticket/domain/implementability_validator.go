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

// portMethodRE matches InterfaceName.MethodName( in ticket description text.
var portMethodRE = regexp.MustCompile(`\b([A-Z][a-zA-Z]+)\.([A-Z][a-zA-Z]+)\(`)

func checkPortSignatures(description string, scannedPorts map[string]ScannedPort) []ImplementabilityFinding {
	var findings []ImplementabilityFinding
	if len(scannedPorts) == 0 {
		return findings
	}

	matches := portMethodRE.FindAllStringSubmatch(description, -1)
	checked := make(map[string]bool)

	for _, match := range matches {
		portName := match[1]
		methodName := match[2]
		key := portName + "." + methodName
		if checked[key] {
			continue
		}
		checked[key] = true

		port, found := scannedPorts[portName]
		if !found {
			f, _ := NewImplementabilityFinding(
				FindingSeverityMajor,
				"Design",
				fmt.Sprintf("Port '%s' referenced in ticket but not found in scanned codebase", portName),
			)
			findings = append(findings, f)
			continue
		}

		methodFound := false
		for _, m := range port.Methods() {
			if m.Name() == methodName {
				methodFound = true
				break
			}
		}
		if !methodFound {
			f, _ := NewImplementabilityFinding(
				FindingSeverityCritical,
				"Design",
				fmt.Sprintf("Method '%s.%s' referenced in ticket but not found on scanned port", portName, methodName),
			)
			findings = append(findings, f)
		}
	}

	return findings
}

// layerViolationRE matches patterns suggesting domain-layer imports of infrastructure concerns.
var layerViolationRE = regexp.MustCompile(
	`(?i)domain\s+(?:layer\s+)?(?:imports?|depends?\s+on|references?)\s+` +
		`(?:infrastructure|http|database|sql|grpc|external|third.party)`,
)

func checkLayerCompliance(description string) []ImplementabilityFinding {
	var findings []ImplementabilityFinding
	for _, match := range layerViolationRE.FindAllString(description, -1) {
		f, _ := NewImplementabilityFinding(
			FindingSeverityCritical,
			"DDD Alignment",
			fmt.Sprintf("Layer violation: '%s' — domain must not depend on infrastructure", match),
		)
		findings = append(findings, f)
	}
	return findings
}

// pascalCaseTypeRE matches PascalCase identifiers that look like type names.
var pascalCaseTypeRE = regexp.MustCompile(`\b([A-Z][a-z]+(?:[A-Z][a-z]+)+)\b`)

func checkGlossaryAlignment(description string, glossaryTerms []string) []ImplementabilityFinding {
	var findings []ImplementabilityFinding
	if len(glossaryTerms) == 0 {
		return findings
	}

	termSet := make(map[string]bool, len(glossaryTerms))
	for _, term := range glossaryTerms {
		termSet[strings.ToLower(term)] = true
	}

	matches := pascalCaseTypeRE.FindAllString(description, -1)
	checked := make(map[string]bool)

	for _, typeName := range matches {
		lower := strings.ToLower(typeName)
		if checked[lower] {
			continue
		}
		checked[lower] = true

		if !termSet[lower] {
			f, _ := NewImplementabilityFinding(
				FindingSeverityMajor,
				"DDD Alignment",
				fmt.Sprintf("Type '%s' not found in ubiquitous language glossary", typeName),
			)
			findings = append(findings, f)
		}
	}

	return findings
}

// ValidateImplementabilityWithScanner runs all implementability checks including
// AST-based port verification, layer compliance, and glossary alignment.
// scannedPorts may be nil (AST checks skipped). glossaryTerms may be empty (glossary checks skipped).
func ValidateImplementabilityWithScanner(
	ticket GeneratedTicket,
	scannedPorts map[string]ScannedPort,
	glossaryTerms []string,
) DesignTraceResult {
	if ticket.DetailLevel() == vo.TicketDetailStub {
		return NewDesignTraceResult(ticket.TicketID(), nil)
	}

	var findings []ImplementabilityFinding
	desc := ticket.Description()

	// Existing text-based checks
	findings = append(findings, checkSectionPresence(ticket)...)
	findings = append(findings, checkUnspecifiedDependencies(desc)...)
	if ticket.DetailLevel() == vo.TicketDetailFull {
		findings = append(findings, checkEmptyInvariants(desc)...)
	}
	findings = append(findings, checkEmptyAcceptanceCriteria(desc)...)

	// New AST-based checks
	if scannedPorts != nil {
		findings = append(findings, checkPortSignatures(desc, scannedPorts)...)
	}
	findings = append(findings, checkLayerCompliance(desc)...)
	if len(glossaryTerms) > 0 {
		findings = append(findings, checkGlossaryAlignment(desc, glossaryTerms)...)
	}

	return NewDesignTraceResult(ticket.TicketID(), findings)
}

// ValidateImplementabilityPlanWithScanner validates all tickets with AST-based checks.
func ValidateImplementabilityPlanWithScanner(
	tickets []GeneratedTicket,
	scannedPorts map[string]ScannedPort,
	glossaryTerms []string,
) []DesignTraceResult {
	results := make([]DesignTraceResult, len(tickets))
	for i, t := range tickets {
		results[i] = ValidateImplementabilityWithScanner(t, scannedPorts, glossaryTerms)
	}
	return results
}
