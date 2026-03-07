package domain

import (
	"fmt"
	"strings"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// Gate labels for bullet-point rendering in tickets.
var gateLabels = map[vo.QualityGate]string{
	vo.QualityGateLint:    "zero errors",
	vo.QualityGateTypes:   "zero errors",
	vo.QualityGateTests:   "all pass",
	vo.QualityGateFitness: "all pass",
}

// Gates excluded from ticket quality gates section.
var excludedGates = map[vo.QualityGate]bool{
	vo.QualityGateFitness: true,
}

// RenderTicketDetail renders a ticket description for the given aggregate and detail level.
// Profile may be nil, in which case GenericProfile is used.
func RenderTicketDetail(
	aggregate vo.AggregateDesign,
	detailLevel vo.TicketDetailLevel,
	profile vo.StackProfile,
) string {
	if profile == nil {
		profile = vo.GenericProfile{}
	}
	switch detailLevel {
	case vo.TicketDetailFull:
		return renderFull(aggregate, profile)
	case vo.TicketDetailStandard:
		return renderStandard(aggregate, profile)
	case vo.TicketDetailStub:
		return renderStub(aggregate)
	default:
		return renderStub(aggregate)
	}
}

func renderFull(agg vo.AggregateDesign, profile vo.StackProfile) string {
	sections := []string{
		renderGoalSection(agg),
		renderDDDAlignmentSection(agg),
		renderDesignSection(agg),
		renderSOLIDSection(agg),
		renderTDDSection(profile),
		renderStepsSection(agg),
		renderAcceptanceCriteriaSection(agg),
		renderEdgeCasesSection(),
	}
	gates := renderQualityGatesSection(profile)
	if gates != "" {
		sections = append(sections, gates)
	}
	return strings.Join(sections, "\n")
}

func renderStandard(agg vo.AggregateDesign, profile vo.StackProfile) string {
	lines := []string{
		"## Goal",
		fmt.Sprintf("Implement the `%s` aggregate in the `%s` bounded context.", agg.Name(), agg.ContextName()),
		"",
		"## DDD Alignment",
		fmt.Sprintf("- **Bounded Context:** %s", agg.ContextName()),
		fmt.Sprintf("- **Aggregate Root:** %s", agg.RootEntity()),
		"",
		"## Steps",
		fmt.Sprintf("1. Create `%s` aggregate with core logic", agg.Name()),
		"2. Add repository port interface",
		"3. Write unit tests",
		"",
		"## Acceptance Criteria",
		fmt.Sprintf("- [ ] `%s` aggregate root created", agg.Name()),
		"- [ ] All tests pass",
		"- [ ] Coverage >= 80%",
	}
	gateLines := renderQualityGateLines(profile)
	if len(gateLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, gateLines...)
	}
	return strings.Join(lines, "\n")
}

func renderStub(agg vo.AggregateDesign) string {
	lines := []string{
		"> **Stub ticket.** Full specification will be added when blockers are resolved.",
		"",
		"## Goal / Problem",
		"",
		fmt.Sprintf("Integrate `%s` boundary for `%s`.", agg.ContextName(), agg.Name()),
		"",
		"## DDD Alignment",
		"",
		"| Aspect | Detail |",
		"|--------|--------|",
		fmt.Sprintf("| Bounded Context | %s |", agg.ContextName()),
		"| Layer | domain |",
		"",
		"## Risks / Dependencies",
		"",
		"- Blocked by: (see formal dependencies)",
	}
	return strings.Join(lines, "\n")
}

// -- Private section renderers ------------------------------------------------

func renderGoalSection(agg vo.AggregateDesign) string {
	return fmt.Sprintf("## Goal\nImplement the `%s` aggregate in the `%s` bounded context.\n",
		agg.Name(), agg.ContextName())
}

func renderDDDAlignmentSection(agg vo.AggregateDesign) string {
	lines := []string{
		"## DDD Alignment",
		fmt.Sprintf("- **Bounded Context:** %s", agg.ContextName()),
		fmt.Sprintf("- **Aggregate Root:** %s", agg.RootEntity()),
	}
	if len(agg.ContainedObjects()) > 0 {
		lines = append(lines, fmt.Sprintf("- **Contained Objects:** %s",
			strings.Join(agg.ContainedObjects(), ", ")))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func renderDesignSection(agg vo.AggregateDesign) string {
	lines := []string{"## Design"}
	if len(agg.Invariants()) > 0 {
		lines = append(lines, "### Invariants")
		for _, inv := range agg.Invariants() {
			lines = append(lines, fmt.Sprintf("- %s", inv))
		}
		lines = append(lines, "")
	}
	if len(agg.Commands()) > 0 {
		lines = append(lines, "### Commands")
		for _, cmd := range agg.Commands() {
			lines = append(lines, fmt.Sprintf("- %s", cmd))
		}
		lines = append(lines, "")
	}
	if len(agg.DomainEvents()) > 0 {
		lines = append(lines, "### Domain Events")
		for _, evt := range agg.DomainEvents() {
			lines = append(lines, fmt.Sprintf("- %s", evt))
		}
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func renderSOLIDSection(agg vo.AggregateDesign) string {
	return strings.Join([]string{
		"## SOLID Mapping",
		fmt.Sprintf("- **S:** `%s` owns %s logic only", agg.Name(), agg.ContextName()),
		"- **O:** Extend via new commands/events, not modification",
		"- **L:** Subtypes honor aggregate contract",
		"- **I:** Focused repository interface for this aggregate",
		"- **D:** Depend on ports, not infrastructure",
		"",
	}, "\n")
}

func renderTDDSection(profile vo.StackProfile) string {
	testCmd := "<test-runner>"
	if profile != nil {
		cmds := profile.QualityGateCommands()
		if parts, ok := cmds[vo.QualityGateTests]; ok {
			testCmd = strings.Join(parts, " ")
		}
	}
	return fmt.Sprintf(
		"## TDD Workflow\n"+
			"1. **RED:** Write failing tests for each invariant\n"+
			"   - Run: `%s` → should FAIL\n"+
			"2. **GREEN:** Implement minimal code to pass\n"+
			"   - Run: `%s` → should PASS\n"+
			"3. **REFACTOR:** Clean up while tests stay green\n",
		testCmd, testCmd)
}

func renderStepsSection(agg vo.AggregateDesign) string {
	return strings.Join([]string{
		"## Steps",
		fmt.Sprintf("1. Create `%s` aggregate with invariant enforcement", agg.Name()),
		"2. Implement commands and domain events",
		"3. Add repository port interface",
		"4. Write unit tests for all invariants",
		"",
	}, "\n")
}

func renderAcceptanceCriteriaSection(agg vo.AggregateDesign) string {
	lines := []string{
		"## Acceptance Criteria",
		fmt.Sprintf("- [ ] `%s` aggregate root created", agg.Name()),
	}
	for _, inv := range agg.Invariants() {
		lines = append(lines, fmt.Sprintf("- [ ] Invariant enforced: %s", inv))
	}
	for _, cmd := range agg.Commands() {
		lines = append(lines, fmt.Sprintf("- [ ] Command implemented: %s", cmd))
	}
	lines = append(lines, "- [ ] All tests pass")
	lines = append(lines, "- [ ] Coverage >= 80%")
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func renderEdgeCasesSection() string {
	return "## Edge Cases\n" +
		"- Empty or invalid inputs raise `InvariantViolationError`\n" +
		"- Duplicate operations are idempotent or raise\n"
}

func renderQualityGatesSection(profile vo.StackProfile) string {
	lines := renderQualityGateLines(profile)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func renderQualityGateLines(profile vo.StackProfile) []string {
	cmds := profile.QualityGateCommands()
	if len(cmds) == 0 {
		return nil
	}
	lines := []string{"## Quality Gates"}
	// Render in a stable order.
	for _, gate := range vo.AllQualityGates() {
		if excludedGates[gate] {
			continue
		}
		parts, ok := cmds[gate]
		if !ok {
			continue
		}
		cmdStr := strings.Join(parts, " ")
		label := gateLabels[gate]
		if label == "" {
			label = "passes"
		}
		lines = append(lines, fmt.Sprintf("- `%s` -- %s", cmdStr, label))
	}
	return lines
}
