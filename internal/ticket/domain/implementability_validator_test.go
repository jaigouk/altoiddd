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

// ---------------------------------------------------------------------------
// Tests — ValidateImplementabilityWithScanner (1wu.2)
// ---------------------------------------------------------------------------

func TestValidateWithScanner_MatchingPortSignature_NoExtraFinding(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The handler calls OrderRepository.Save( to persist the order.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-scan-1")

	scannedPorts := map[string]domain.ScannedPort{
		"OrderRepository": domain.NewScannedPort("OrderRepository", "ports.go", []domain.ScannedMethod{
			domain.NewScannedMethod("Save", map[string]string{"ctx": "context.Context", "order": "Order"}),
		}),
	}

	result := domain.ValidateImplementabilityWithScanner(ticket, scannedPorts, nil)
	// Should have no findings beyond what the base validator would produce.
	// The base validator passes for a well-formed ticket, and the port match is correct.
	assert.True(t, result.IsValid(), "well-formed ticket with matching port should be valid")
}

func TestValidateWithScanner_MismatchedMethod_CriticalFinding(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The handler calls OrderRepository.Delete( to remove the order.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-scan-2")

	// Port exists but does NOT have Delete method.
	scannedPorts := map[string]domain.ScannedPort{
		"OrderRepository": domain.NewScannedPort("OrderRepository", "ports.go", []domain.ScannedMethod{
			domain.NewScannedMethod("Save", map[string]string{"ctx": "context.Context"}),
		}),
	}

	result := domain.ValidateImplementabilityWithScanner(ticket, scannedPorts, nil)
	assert.False(t, result.IsValid())

	var hasCritical bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityCritical {
			hasCritical = true
		}
	}
	assert.True(t, hasCritical, "mismatched method should produce CRITICAL finding")
}

func TestValidateWithScanner_UnknownPort_MajorFinding(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The handler calls UnknownService.Process( to handle the request.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-scan-3")

	// scannedPorts does NOT contain UnknownService.
	scannedPorts := map[string]domain.ScannedPort{
		"OrderRepository": domain.NewScannedPort("OrderRepository", "ports.go", []domain.ScannedMethod{
			domain.NewScannedMethod("Save", map[string]string{}),
		}),
	}

	result := domain.ValidateImplementabilityWithScanner(ticket, scannedPorts, nil)
	assert.False(t, result.IsValid())

	var hasMajor bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityMajor && f.Location() == "Design" {
			hasMajor = true
		}
	}
	assert.True(t, hasMajor, "unknown port should produce MAJOR finding")
}

func TestValidateWithScanner_LayerViolation_CriticalFinding(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The domain layer imports infrastructure HTTP client directly.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-layer-1")

	result := domain.ValidateImplementabilityWithScanner(ticket, nil, nil)
	assert.False(t, result.IsValid())

	var hasCritical bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityCritical && f.Location() == "DDD Alignment" {
			hasCritical = true
		}
	}
	assert.True(t, hasCritical, "layer violation should produce CRITICAL finding")
}

func TestValidateWithScanner_CleanLayers_NoLayerFinding(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The handler depends on the OrderRepository port interface.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-layer-2")

	result := domain.ValidateImplementabilityWithScanner(ticket, nil, nil)
	assert.True(t, result.IsValid(), "clean layers should produce no findings")
}

func TestValidateWithScanner_GlossaryKnownTerms_NoFinding(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The OrderRepository port handles DomainStory persistence.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-gloss-1")

	// Must include all PascalCase multi-word types from the full description,
	// including "InvariantViolationError" from the base template edge cases section.
	glossary := []string{"OrderRepository", "DomainStory", "InvariantViolationError"}
	result := domain.ValidateImplementabilityWithScanner(ticket, nil, glossary)
	assert.True(t, result.IsValid(), "known glossary terms should produce no findings")
}

func TestValidateWithScanner_GlossaryUnknownTerm_MajorFinding(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The FooBarBaz handler processes requests.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-gloss-2")

	glossary := []string{"OrderRepository", "DomainStory"}
	result := domain.ValidateImplementabilityWithScanner(ticket, nil, glossary)
	assert.False(t, result.IsValid())

	var hasMajor bool
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityMajor && f.Location() == "DDD Alignment" {
			hasMajor = true
		}
	}
	assert.True(t, hasMajor, "unknown glossary term should produce MAJOR finding")
}

func TestValidateWithScanner_GlossaryCaseInsensitive(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The OrderRepository port is used.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-gloss-ci")

	// Glossary uses different casing -- should still match via case-insensitive comparison.
	glossary := []string{"orderrepository", "invariantviolationerror"}
	result := domain.ValidateImplementabilityWithScanner(ticket, nil, glossary)
	assert.True(t, result.IsValid(), "glossary should match case-insensitively")
}

func TestValidateWithScanner_EmptyGlossary_SkipsGlossaryCheck(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The FooBarBaz handler processes requests.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-gloss-empty")

	// Empty glossary should skip the glossary check entirely.
	result := domain.ValidateImplementabilityWithScanner(ticket, nil, []string{})
	assert.True(t, result.IsValid(), "empty glossary should skip check — no findings")
}

func TestValidateWithScanner_EmptyScannedPorts_SkipsPortCheck(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The handler calls OrderRepository.Save( to persist.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-empty-ports")

	// Non-nil but empty map: checkPortSignatures returns early when len == 0.
	scannedPorts := map[string]domain.ScannedPort{}
	result := domain.ValidateImplementabilityWithScanner(ticket, scannedPorts, nil)
	assert.True(t, result.IsValid(), "empty scannedPorts should produce no port findings")
}

func TestValidateWithScanner_CombinesAllChecks(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The domain imports http. UnknownService.Process( is called. FooBarBaz is used.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-combo-1")

	scannedPorts := map[string]domain.ScannedPort{
		"OrderRepository": domain.NewScannedPort("OrderRepository", "ports.go", nil),
	}
	glossary := []string{"OrderRepository"}

	result := domain.ValidateImplementabilityWithScanner(ticket, scannedPorts, glossary)
	assert.False(t, result.IsValid())
	// Should have multiple findings: layer violation + unknown port + unknown glossary terms.
	assert.GreaterOrEqual(t, len(result.Findings()), 3,
		"combined check should produce findings from multiple checkers")
}

func TestValidateWithScanner_NilPorts_SkipsPortCheck(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The handler calls OrderRepository.Save( to persist.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-nil-1")

	// nil scannedPorts should skip port signature checking entirely.
	result := domain.ValidateImplementabilityWithScanner(ticket, nil, nil)
	assert.True(t, result.IsValid(), "nil ports should skip port check — no findings")
}

func TestValidateWithScanner_StubTicket_SkipsAllChecks(t *testing.T) {
	t.Parallel()
	stubDesc := "> **Stub ticket.**\n\n## Goal / Problem\nIntegrate boundary.\n"
	ticket := makeValidTicket(vo.TicketDetailStub, stubDesc, "t-stub-1")

	scannedPorts := map[string]domain.ScannedPort{
		"Foo": domain.NewScannedPort("Foo", "foo.go", nil),
	}

	result := domain.ValidateImplementabilityWithScanner(ticket, scannedPorts, []string{"Bar"})
	assert.True(t, result.IsValid(), "stub tickets should skip all checks")
}

func TestValidateWithScanner_DuplicatePortMethodRef_DeduplicatesFindings(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"Call OrderRepository.Delete( first, then OrderRepository.Delete( again.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-dedup-1")

	scannedPorts := map[string]domain.ScannedPort{
		"OrderRepository": domain.NewScannedPort("OrderRepository", "ports.go", []domain.ScannedMethod{
			domain.NewScannedMethod("Save", nil),
		}),
	}

	result := domain.ValidateImplementabilityWithScanner(ticket, scannedPorts, nil)
	assert.False(t, result.IsValid())

	// Count CRITICAL findings at "Design" location -- should be exactly 1 (deduplicated).
	criticalCount := 0
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityCritical && f.Location() == "Design" {
			criticalCount++
		}
	}
	assert.Equal(t, 1, criticalCount, "duplicate port.method references should be deduplicated")
}

func TestValidateWithScanner_MultipleLayerViolations(t *testing.T) {
	t.Parallel()
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"The domain imports http and the domain depends on database.\n",
	)
	ticket := makeValidTicket(vo.TicketDetailFull, desc, "t-multi-layer")

	result := domain.ValidateImplementabilityWithScanner(ticket, nil, nil)
	assert.False(t, result.IsValid())

	criticalCount := 0
	for _, f := range result.Findings() {
		if f.Severity() == domain.FindingSeverityCritical && f.Location() == "DDD Alignment" {
			criticalCount++
		}
	}
	assert.Equal(t, 2, criticalCount, "each layer violation pattern should produce its own CRITICAL finding")
}

func TestValidatePlanWithScanner_ReturnsPerTicketResults(t *testing.T) {
	t.Parallel()
	tickets := []domain.GeneratedTicket{
		makeValidTicket(vo.TicketDetailFull, "", "t-plan-1"),
		makeValidTicket(vo.TicketDetailFull, "", "t-plan-2"),
	}
	results := domain.ValidateImplementabilityPlanWithScanner(tickets, nil, nil)
	require.Len(t, results, 2)
	assert.Equal(t, "t-plan-1", results[0].TicketID())
	assert.Equal(t, "t-plan-2", results[1].TicketID())
}

func TestValidatePlanWithScanner_PropagatesPortsAndGlossary(t *testing.T) {
	t.Parallel()
	// Both tickets reference a port method that does not exist.
	desc := makeFullDescription(
		"### Invariants\n- Order total must be positive\n",
		"- [ ] Aggregate created\n- [ ] All tests pass\n",
		"Call OrderRepository.Delete( here.\n",
	)
	tickets := []domain.GeneratedTicket{
		makeValidTicket(vo.TicketDetailFull, desc, "t-plan-p1"),
		makeValidTicket(vo.TicketDetailFull, desc, "t-plan-p2"),
	}
	scannedPorts := map[string]domain.ScannedPort{
		"OrderRepository": domain.NewScannedPort("OrderRepository", "ports.go", []domain.ScannedMethod{
			domain.NewScannedMethod("Save", nil),
		}),
	}
	results := domain.ValidateImplementabilityPlanWithScanner(tickets, scannedPorts, nil)
	require.Len(t, results, 2)
	// Both tickets should fail because Delete is not on OrderRepository.
	assert.False(t, results[0].IsValid(), "first ticket should fail with mismatched method")
	assert.False(t, results[1].IsValid(), "second ticket should fail with mismatched method")
}

func TestValidatePlanWithScanner_EmptyPlan(t *testing.T) {
	t.Parallel()
	results := domain.ValidateImplementabilityPlanWithScanner(nil, nil, nil)
	assert.Empty(t, results, "empty ticket list should produce empty results")
}
