package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/ticket/domain"
)

func makeTestAggregate() vo.AggregateDesign {
	return vo.NewAggregateDesign(
		"OrderAggregate", "Orders", "Order",
		[]string{"OrderLine", "OrderStatus"},
		[]string{"total must be positive", "at least one line item"},
		[]string{"PlaceOrder", "CancelOrder"},
		[]string{"OrderPlaced", "OrderCancelled"},
	)
}

// ---------------------------------------------------------------------------
// FULL detail
// ---------------------------------------------------------------------------

func TestFullHasAllSections(t *testing.T) {
	t.Parallel()
	agg := makeTestAggregate()
	result := domain.RenderTicketDetail(agg, vo.TicketDetailFull, vo.PythonUvProfile{})

	for _, section := range []string{
		"## Goal", "## DDD Alignment", "## Design",
		"### Invariants", "### Commands", "### Domain Events",
		"## SOLID Mapping", "## TDD Workflow", "## Steps",
		"## Acceptance Criteria", "## Edge Cases", "## Quality Gates",
	} {
		assert.Contains(t, result, section)
	}
}

func TestFullIncludesAggregateName(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.PythonUvProfile{})
	assert.Contains(t, result, "OrderAggregate")
	assert.Contains(t, result, "Orders")
}

func TestFullIncludesInvariants(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.PythonUvProfile{})
	assert.Contains(t, result, "total must be positive")
	assert.Contains(t, result, "at least one line item")
}

func TestFullIncludesCommandsAndEvents(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.PythonUvProfile{})
	assert.Contains(t, result, "PlaceOrder")
	assert.Contains(t, result, "CancelOrder")
	assert.Contains(t, result, "OrderPlaced")
	assert.Contains(t, result, "OrderCancelled")
}

// ---------------------------------------------------------------------------
// STANDARD detail
// ---------------------------------------------------------------------------

func TestStandardHasCoreSections(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailStandard, vo.PythonUvProfile{})
	for _, section := range []string{
		"## Goal", "## DDD Alignment", "## Steps",
		"## Acceptance Criteria", "## Quality Gates",
	} {
		assert.Contains(t, result, section)
	}
}

func TestStandardOmitsFullSections(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailStandard, vo.PythonUvProfile{})
	assert.NotContains(t, result, "## Design")
	assert.NotContains(t, result, "## SOLID Mapping")
	assert.NotContains(t, result, "## TDD Workflow")
	assert.NotContains(t, result, "## Edge Cases")
}

// ---------------------------------------------------------------------------
// STUB detail
// ---------------------------------------------------------------------------

func TestStubMatchesTemplate(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailStub, vo.PythonUvProfile{})
	assert.Contains(t, result, "Stub ticket")
	assert.Contains(t, result, "## Goal / Problem")
	assert.Contains(t, result, "Integrate")
	assert.Contains(t, result, "## DDD Alignment")
	assert.Contains(t, result, "## Risks / Dependencies")
}

func TestStubNoImplementationSections(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailStub, vo.PythonUvProfile{})
	assert.NotContains(t, result, "## Design")
	assert.NotContains(t, result, "## Steps")
	assert.NotContains(t, result, "## SOLID")
	assert.NotContains(t, result, "## TDD")
}

// ---------------------------------------------------------------------------
// Profile integration
// ---------------------------------------------------------------------------

func TestPythonProfileIncludesQualityGates(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.PythonUvProfile{})
	assert.Contains(t, result, "## Quality Gates")
}

func TestPythonProfileBulletFormat(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.PythonUvProfile{})
	assert.Contains(t, result, "- `uv run ruff check .` -- zero errors")
	assert.Contains(t, result, "- `uv run mypy .` -- zero errors")
	assert.Contains(t, result, "- `uv run pytest` -- all pass")
	assert.NotContains(t, result, "```bash")
}

func TestGenericProfileOmitsQualityGates(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.GenericProfile{})
	assert.NotContains(t, result, "## Quality Gates")
	assert.NotContains(t, result, "uv run")
}

func TestPythonProfileExcludesFitnessGate(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.PythonUvProfile{})
	qgIdx := strings.Index(result, "## Quality Gates")
	if qgIdx == -1 {
		t.Fatal("No Quality Gates section found")
	}
	qgSection := result[qgIdx:]
	nextSection := strings.Index(qgSection[1:], "\n## ")
	if nextSection != -1 {
		qgSection = qgSection[:nextSection+1]
	}
	assert.NotContains(t, strings.ToLower(qgSection), "fitness")
}

func TestStandardPythonProfileIncludesQualityGates(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailStandard, vo.PythonUvProfile{})
	assert.Contains(t, result, "## Quality Gates")
	assert.Contains(t, result, "uv run ruff check .")
}

func TestStandardGenericProfileOmitsQualityGates(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailStandard, vo.GenericProfile{})
	assert.NotContains(t, result, "## Quality Gates")
	assert.NotContains(t, result, "uv run")
}

func TestStubNoQualityGatesRegardless(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailStub, vo.PythonUvProfile{})
	assert.NotContains(t, result, "## Quality Gates")
}

// ---------------------------------------------------------------------------
// TDD section
// ---------------------------------------------------------------------------

func TestNoneProfileTDDPlaceholder(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.GenericProfile{})
	tdd := extractTDDSection(result)
	assert.Contains(t, tdd, "<test-runner>")
	assert.NotContains(t, tdd, "pytest")
	assert.NotContains(t, tdd, "uv run")
}

func TestPythonTDDIncludesTestRunner(t *testing.T) {
	t.Parallel()
	result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, vo.PythonUvProfile{})
	tdd := extractTDDSection(result)
	assert.Contains(t, tdd, "uv run pytest")
}

func TestTDDHasUniversalPrinciples(t *testing.T) {
	t.Parallel()
	for _, profile := range []vo.StackProfile{vo.PythonUvProfile{}, vo.GenericProfile{}} {
		result := domain.RenderTicketDetail(makeTestAggregate(), vo.TicketDetailFull, profile)
		tdd := extractTDDSection(result)
		assert.Contains(t, tdd, "RED")
		assert.Contains(t, tdd, "GREEN")
		assert.Contains(t, tdd, "REFACTOR")
	}
}

// ---------------------------------------------------------------------------
// Stub tier tests
// ---------------------------------------------------------------------------

func TestStubHasNotice(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("ShipmentRoot", "Shipping", "ShipmentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.Contains(t, result, "Stub ticket")
}

func TestStubHasGoalSection(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("ShipmentRoot", "Shipping", "ShipmentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.Contains(t, result, "## Goal / Problem")
}

func TestStubHasContextInGoal(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("PaymentRoot", "Payments", "PaymentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.Contains(t, result, "Payments")
	assert.Contains(t, result, "PaymentRoot")
}

func TestStubHasDDDAlignmentTable(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("ShipmentRoot", "Shipping", "ShipmentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.Contains(t, result, "## DDD Alignment")
	assert.Contains(t, result, "| Bounded Context | Shipping |")
}

func TestStubHasRisksSection(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("ShipmentRoot", "Shipping", "ShipmentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.Contains(t, result, "## Risks / Dependencies")
}

func TestStubNoSOLIDSection(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("ShipmentRoot", "Shipping", "ShipmentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.NotContains(t, result, "## SOLID")
}

func TestStubNoTDDSection(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("ShipmentRoot", "Shipping", "ShipmentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.NotContains(t, result, "## TDD")
}

func TestStubNoQualityGatesSection(t *testing.T) {
	t.Parallel()
	agg := vo.NewAggregateDesign("ShipmentRoot", "Shipping", "ShipmentRoot", nil, nil, nil, nil)
	result := domain.RenderTicketDetail(agg, vo.TicketDetailStub, nil)
	assert.NotContains(t, result, "## Quality Gates")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func extractTDDSection(rendered string) string {
	start := strings.Index(rendered, "## TDD Workflow")
	if start == -1 {
		return ""
	}
	next := strings.Index(rendered[start+1:], "\n## ")
	if next == -1 {
		return rendered[start:]
	}
	return rendered[start : start+1+next]
}
