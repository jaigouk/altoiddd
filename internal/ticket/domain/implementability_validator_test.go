package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/ticket/domain"
)

func makeFullDescription(invariants, acItems, designExtra string) string {
	if invariants == "" {
		invariants = ""
	}
	if acItems == "" {
		acItems = ""
	}
	return "## Goal\nImplement the `Order` aggregate in the `Sales` bounded context.\n\n" +
		"## DDD Alignment\n- **Bounded Context:** Sales\n- **Aggregate Root:** Order\n\n" +
		"## Design\n" + invariants + designExtra + "\n" +
		"## SOLID Mapping\n- **S:** Order owns Sales logic only\n\n" +
		"## TDD Workflow\n1. RED: Write failing tests\n\n" +
		"## Steps\n1. Create Order aggregate\n\n" +
		"## Acceptance Criteria\n" + acItems + "\n" +
		"## Edge Cases\n- Empty inputs raise InvariantViolationError\n\n" +
		"## Quality Gates\n- `uv run pytest` -- all pass\n"
}

func makeValidTicket(detailLevel vo.TicketDetailLevel, desc string, ticketID string) domain.GeneratedTicket {
	if desc == "" {
		desc = makeFullDescription(
			"### Invariants\n- Order total must be positive\n",
			"- [ ] Aggregate created\n- [ ] All tests pass\n",
			"",
		)
	}
	return domain.NewGeneratedTicket(
		ticketID, "Implement Order aggregate", desc,
		detailLevel, "e-001", "Sales", "Order", nil, 0,
	)
}

func TestPassesWellFormedTicket(t *testing.T) {
	t.Parallel()
	ticket := makeValidTicket(vo.TicketDetailFull, "", "t-001")
	result := domain.ValidateImplementability(ticket)
	assert.True(t, result.IsValid())
}

func TestDetectsUnspecifiedDependency(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The adapter performs iterative web search to gather results.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-001")
	result := domain.ValidateImplementability(ticket)
	assert.False(t, result.IsValid())
	var hasCritical bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityCritical {
			hasCritical = true
		}
	}
	assert.True(t, hasCritical)
}

func TestDetectsEmptyInvariantsOnFull(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-001")
	result := domain.ValidateImplementability(ticket)
	var hasMajor bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityMajor {
			hasMajor = true
		}
	}
	assert.True(t, hasMajor)
}

func TestDetectsEmptyAcceptanceCriteria(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"TBD\n",
		"",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-001")
	result := domain.ValidateImplementability(ticket)
	var hasMajor bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityMajor {
			hasMajor = true
		}
	}
	assert.True(t, hasMajor)
}

func TestPassesStubTicket(t *testing.T) {
	t.Parallel()
	stubDesc := "> **Stub ticket.**\n\n## Goal / Problem\nIntegrate boundary.\n\n## DDD Alignment\n| Aspect | Detail |\n"
	ticket := makeValidTicket(vo.TicketDetailStub, stubDesc, "t-001")
	result := domain.ValidateImplementability(ticket)
	assert.True(t, result.IsValid())
}

func TestReturnsStructuredResult(t *testing.T) {
	t.Parallel()
	ticket := makeValidTicket(vo.TicketDetailFull, "", "t-001")
	result := domain.ValidateImplementability(ticket)
	assert.Equal(t, "t-001", result.TicketID())
}

func TestMultipleFindingsAccumulated(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription("", "TBD\n", "")
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-001")
	result := domain.ValidateImplementability(ticket)
	assert.GreaterOrEqual(t, len(result.Findings()), 2)
}

func TestValidatePlanReturnsPerTicketResults(t *testing.T) {
	t.Parallel()
	tickets := []domain.GeneratedTicket{
		makeValidTicket(vo.TicketDetailFull, "", "t-001"),
		makeValidTicket(vo.TicketDetailFull, "", "t-002"),
	}
	results := domain.ValidateImplementabilityPlan(tickets)
	require.Len(t, results, 2)
	assert.Equal(t, "t-001", results[0].TicketID())
	assert.Equal(t, "t-002", results[1].TicketID())
}

func TestStandardSectionsSubsetOfFull(t *testing.T) {
	t.Parallel()
	fullSet := make(map[string]bool)
	for _, s := range domain.FullSections {
		fullSet[s] = true
	}
	for _, s := range domain.StandardSections {
		assert.True(t, fullSet[s], "standard section %q not in full sections", s)
	}
}

// ---------------------------------------------------------------------------
// Regression: 20c.5
// ---------------------------------------------------------------------------

func TestRegression20c5(t *testing.T) {
	t.Parallel()
	description := "## Goal\n" +
		"Implement domain research with RLM pattern.\n\n" +
		"## DDD Alignment\n" +
		"Bounded Context: Knowledge Base\n\n" +
		"## Design\n" +
		"### Invariants\n" +
		"- Research findings must have sources\n\n" +
		"The RLM research adapter performs iterative web search to " +
		"gather domain intelligence. Results are synthesized via LLM.\n\n" +
		"## SOLID Mapping\n" +
		"- SRP: RlmResearchAdapter handles research only\n\n" +
		"## TDD Workflow\n" +
		"RED: test_research_returns_findings\n\n" +
		"## Steps\n" +
		"1. Create RlmResearchAdapter\n\n" +
		"## Acceptance Criteria\n" +
		"- [ ] Adapter returns ResearchBriefing\n" +
		"- [ ] Findings have source attribution\n\n" +
		"## Edge Cases\n" +
		"- LLM unavailable -> graceful degradation\n"

	ticket := domain.NewGeneratedTicket(
		"test-20c5", "Domain Research Port and RLM Adapter", description,
		vo.TicketDetailFull, "test-epic", "Knowledge Base", "Research", nil, 0,
	)

	result := domain.ValidateImplementability(ticket)
	assert.False(t, result.IsValid())
	var hasCritical bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityCritical {
			hasCritical = true
		}
	}
	assert.True(t, hasCritical)
}
