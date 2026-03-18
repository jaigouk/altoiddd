package domain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type ctxDef struct {
	Name           string
	Classification vo.SubdomainClassification
}

func makeModel(
	contexts []ctxDef,
	aggregates map[string][]string,
) *ddd.DomainModel {
	model := ddd.NewDomainModel("test-model")

	var allNames []string
	for _, c := range contexts {
		allNames = append(allNames, c.Name)
	}

	steps := make([]string, len(allNames))
	for i, n := range allNames {
		steps[i] = "User manages " + n
	}
	story := vo.NewDomainStory("Test flow", []string{"User"}, "User starts", steps, nil)
	_ = model.AddDomainStory(story)

	for _, c := range contexts {
		_ = model.AddTerm(c.Name, c.Name+" domain", c.Name, nil)
		cl := c.Classification
		ctx := vo.NewDomainBoundedContext(c.Name, "Manages "+c.Name, nil, &cl, "")
		_ = model.AddBoundedContext(ctx)
	}

	if aggregates == nil {
		aggregates = make(map[string][]string)
	}
	for _, c := range contexts {
		if aggs, ok := aggregates[c.Name]; ok {
			for _, aggName := range aggs {
				agg := vo.NewAggregateDesign(aggName, c.Name, aggName,
					nil, []string{"must be valid"}, []string{"Create"}, []string{"Created"})
				_ = model.DesignAggregate(agg)
			}
		} else if c.Classification == vo.SubdomainCore {
			aggName := c.Name + "Root"
			agg := vo.NewAggregateDesign(aggName, c.Name, aggName,
				nil, []string{"must be valid"}, []string{"Create"}, []string{"Created"})
			_ = model.DesignAggregate(agg)
		}
	}

	_ = model.Finalize()
	return model
}

// ---------------------------------------------------------------------------
// 1. Empty plan
// ---------------------------------------------------------------------------

func TestEmptyPlan(t *testing.T) {
	t.Parallel()
	plan := domain.NewTicketPlan()
	assert.Empty(t, plan.Epics())
	assert.Empty(t, plan.Tickets())
	assert.Nil(t, plan.DependencyOrder())
	assert.Empty(t, plan.Events())
}

// ---------------------------------------------------------------------------
// 2. One epic per BC
// ---------------------------------------------------------------------------

func TestOneEpicPerBC(t *testing.T) {
	t.Parallel()
	model := makeModel(
		[]ctxDef{
			{"Orders", vo.SubdomainCore},
			{"Shipping", vo.SubdomainSupporting},
		},
		map[string][]string{"Shipping": {"ShipmentRoot"}},
	)
	plan := domain.NewTicketPlan()
	err := plan.GeneratePlan(model, nil)
	require.NoError(t, err)

	assert.Len(t, plan.Epics(), 2)
	names := map[string]bool{}
	for _, e := range plan.Epics() {
		names[e.BoundedContextName()] = true
	}
	assert.True(t, names["Orders"])
	assert.True(t, names["Shipping"])
}

// ---------------------------------------------------------------------------
// 3. Detail level mapping
// ---------------------------------------------------------------------------

func TestCoreTicketsFullDetail(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Orders", vo.SubdomainCore}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	for _, ticket := range plan.Tickets() {
		if ticket.BoundedContextName() == "Orders" {
			assert.Equal(t, vo.TicketDetailFull, ticket.DetailLevel())
		}
	}
}

func TestSupportingStandardDetail(t *testing.T) {
	t.Parallel()
	model := makeModel(
		[]ctxDef{{"Shipping", vo.SubdomainSupporting}},
		map[string][]string{"Shipping": {"ShipmentRoot"}},
	)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	for _, ticket := range plan.Tickets() {
		if ticket.BoundedContextName() == "Shipping" {
			assert.Equal(t, vo.TicketDetailStandard, ticket.DetailLevel())
		}
	}
}

func TestGenericStubDetail(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Logging", vo.SubdomainGeneric}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	for _, ticket := range plan.Tickets() {
		if ticket.BoundedContextName() == "Logging" {
			assert.Equal(t, vo.TicketDetailStub, ticket.DetailLevel())
		}
	}
}

// ---------------------------------------------------------------------------
// 4. Preview
// ---------------------------------------------------------------------------

func TestPreviewShowsSummary(t *testing.T) {
	t.Parallel()
	model := makeModel(
		[]ctxDef{
			{"Orders", vo.SubdomainCore},
			{"Logging", vo.SubdomainGeneric},
		}, nil,
	)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	summary, err := plan.Preview()
	require.NoError(t, err)
	assert.Contains(t, summary, "Epics: 2")
	assert.Contains(t, summary, "FULL=")
	assert.Contains(t, summary, "STUB=")
	assert.Contains(t, summary, "Orders")
	assert.Contains(t, summary, "Logging")
}

func TestPreviewBeforeGenerateRaises(t *testing.T) {
	t.Parallel()
	plan := domain.NewTicketPlan()
	_, err := plan.Preview()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// ---------------------------------------------------------------------------
// 5. Promote stub
// ---------------------------------------------------------------------------

func TestPromoteStubToFull(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Logging", vo.SubdomainGeneric}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	stub := plan.Tickets()[0]
	assert.Equal(t, vo.TicketDetailStub, stub.DetailLevel())

	err := plan.PromoteStub(stub.TicketID(), nil)
	require.NoError(t, err)
	promoted := plan.Tickets()[0]
	assert.Equal(t, vo.TicketDetailFull, promoted.DetailLevel())
	assert.Contains(t, promoted.Description(), "## Goal")
	assert.Contains(t, promoted.Description(), "## SOLID Mapping")
}

func TestPromoteNonStubRaises(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Orders", vo.SubdomainCore}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	full := plan.Tickets()[0]
	assert.Equal(t, vo.TicketDetailFull, full.DetailLevel())

	err := plan.PromoteStub(full.TicketID(), nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestPromoteUnknownTicketRaises(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Logging", vo.SubdomainGeneric}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	err := plan.PromoteStub("nonexistent-id", nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestPromoteStubInheritsStoredProfile(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Logging", vo.SubdomainGeneric}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, vo.GenericProfile{})

	stub := plan.Tickets()[0]
	_ = plan.PromoteStub(stub.TicketID(), nil)
	promoted := plan.Tickets()[0]

	assert.Equal(t, vo.TicketDetailFull, promoted.DetailLevel())
	assert.NotContains(t, promoted.Description(), "## Quality Gates")
}

func TestPromoteStubExplicitProfileOverrides(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Logging", vo.SubdomainGeneric}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, vo.GenericProfile{})

	stub := plan.Tickets()[0]
	pyProfile := vo.PythonUvProfile{}
	_ = plan.PromoteStub(stub.TicketID(), pyProfile)
	promoted := plan.Tickets()[0]

	assert.Equal(t, vo.TicketDetailFull, promoted.DetailLevel())
	assert.Contains(t, promoted.Description(), "## Quality Gates")
}

func TestPromoteStubPreservesDepth(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Logging", vo.SubdomainGeneric}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	stub := plan.Tickets()[0]
	originalDepth := stub.Depth()

	_ = plan.PromoteStub(stub.TicketID(), nil)
	promoted := plan.Tickets()[0]
	assert.Equal(t, originalDepth, promoted.Depth())
}

// ---------------------------------------------------------------------------
// 6. Approve
// ---------------------------------------------------------------------------

func TestApproveAllEmitsEvent(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Orders", vo.SubdomainCore}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)
	err := plan.Approve(nil)
	require.NoError(t, err)

	events := plan.Events()
	require.Len(t, events, 1)
	assert.Equal(t, plan.PlanID(), events[0].PlanID())
	assert.Len(t, events[0].ApprovedTicketIDs(), len(plan.Tickets()))
	assert.Empty(t, events[0].DismissedTicketIDs())
}

func TestApproveSubset(t *testing.T) {
	t.Parallel()
	model := makeModel(
		[]ctxDef{
			{"Orders", vo.SubdomainCore},
			{"Logging", vo.SubdomainGeneric},
		}, nil,
	)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	firstID := plan.Tickets()[0].TicketID()
	err := plan.Approve([]string{firstID})
	require.NoError(t, err)

	event := plan.Events()[0]
	assert.Equal(t, []string{firstID}, event.ApprovedTicketIDs())
	assert.Len(t, event.DismissedTicketIDs(), len(plan.Tickets())-1)
}

func TestApproveTwiceRaises(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Orders", vo.SubdomainCore}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)
	_ = plan.Approve(nil)

	err := plan.Approve(nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestApproveEmptyPlanRaises(t *testing.T) {
	t.Parallel()
	plan := domain.NewTicketPlan()
	err := plan.Approve(nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// ---------------------------------------------------------------------------
// 7. Error cases
// ---------------------------------------------------------------------------

func TestEmptyModelRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	plan := domain.NewTicketPlan()
	err := plan.GeneratePlan(model, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNoClassificationRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	story := vo.NewDomainStory("Test flow", []string{"User"}, "User starts", []string{"User manages Orders"}, nil)
	_ = model.AddDomainStory(story)
	_ = model.AddTerm("Orders", "Orders domain", "Orders", nil)
	ctx := vo.NewDomainBoundedContext("Orders", "Manages Orders", nil, nil, "")
	_ = model.AddBoundedContext(ctx)

	plan := domain.NewTicketPlan()
	err := plan.GeneratePlan(model, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestRegenerateAfterApproveRaises(t *testing.T) {
	t.Parallel()
	model := makeModel([]ctxDef{{"Orders", vo.SubdomainCore}}, nil)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)
	_ = plan.Approve(nil)

	err := plan.GeneratePlan(model, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// ---------------------------------------------------------------------------
// Depth computation (ticket_depths parity)
// ---------------------------------------------------------------------------

func makeChainModel(chainLength int, classification vo.SubdomainClassification) *ddd.DomainModel {
	model := ddd.NewDomainModel("chain-model")

	names := make([]string, chainLength)
	for i := range chainLength {
		names[i] = fmt.Sprintf("BC_%d", i)
	}

	steps := make([]string, chainLength)
	for i, n := range names {
		steps[i] = "User manages " + n
	}
	story := vo.NewDomainStory("Chain flow", []string{"User"}, "User starts", steps, nil)
	_ = model.AddDomainStory(story)

	for _, name := range names {
		_ = model.AddTerm(name, name+" domain", name, nil)
		cl := classification
		ctx := vo.NewDomainBoundedContext(name, "Manages "+name, nil, &cl, "")
		_ = model.AddBoundedContext(ctx)
		aggName := name + "Root"
		agg := vo.NewAggregateDesign(aggName, name, aggName,
			nil, []string{"must be valid"}, []string{"Create"}, []string{"Created"})
		_ = model.DesignAggregate(agg)
	}

	for i := 0; i < chainLength-1; i++ {
		rel := vo.NewContextRelationship(names[i], names[i+1], "Domain Events")
		_ = model.AddContextRelationship(rel)
	}

	_ = model.Finalize()
	return model
}

func TestRootTicketsDepthZero(t *testing.T) {
	t.Parallel()
	model := makeChainModel(1, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	for _, ticket := range plan.Tickets() {
		assert.Equal(t, 0, ticket.Depth())
	}
}

func TestChainDepthIncrements(t *testing.T) {
	t.Parallel()
	model := makeChainModel(4, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	depthByCtx := map[string]int{}
	for _, ticket := range plan.Tickets() {
		depthByCtx[ticket.BoundedContextName()] = ticket.Depth()
	}
	assert.Equal(t, 0, depthByCtx["BC_0"])
	assert.Equal(t, 1, depthByCtx["BC_1"])
	assert.Equal(t, 2, depthByCtx["BC_2"])
	assert.Equal(t, 3, depthByCtx["BC_3"])
}

func TestDiamondDAGUsesMaxDepth(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("diamond")
	names := []string{"A", "B", "C"}
	steps := make([]string, len(names))
	for i, n := range names {
		steps[i] = "User manages " + n
	}
	story := vo.NewDomainStory("Diamond flow", []string{"User"}, "User starts", steps, nil)
	_ = model.AddDomainStory(story)

	for _, name := range names {
		_ = model.AddTerm(name, name+" domain", name, nil)
		cl := vo.SubdomainSupporting
		ctx := vo.NewDomainBoundedContext(name, "Manages "+name, nil, &cl, "")
		_ = model.AddBoundedContext(ctx)
		agg := vo.NewAggregateDesign(name+"Root", name, name+"Root", nil, nil, nil, nil)
		_ = model.DesignAggregate(agg)
	}
	_ = model.AddContextRelationship(vo.NewContextRelationship("A", "C", "ACL"))
	_ = model.AddContextRelationship(vo.NewContextRelationship("B", "C", "ACL"))
	_ = model.Finalize()

	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	depthByCtx := map[string]int{}
	for _, ticket := range plan.Tickets() {
		depthByCtx[ticket.BoundedContextName()] = ticket.Depth()
	}
	assert.Equal(t, 0, depthByCtx["A"])
	assert.Equal(t, 0, depthByCtx["B"])
	assert.Equal(t, 1, depthByCtx["C"])
}

func TestOrphanTicketDepthZero(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("orphan")
	story := vo.NewDomainStory("Orphan flow", []string{"User"}, "User starts", []string{"User manages Orphan"}, nil)
	_ = model.AddDomainStory(story)
	_ = model.AddTerm("Orphan", "Orphan domain", "Orphan", nil)
	cl := vo.SubdomainSupporting
	ctx := vo.NewDomainBoundedContext("Orphan", "Manages Orphan", nil, &cl, "")
	_ = model.AddBoundedContext(ctx)
	agg := vo.NewAggregateDesign("OrphanRoot", "Orphan", "OrphanRoot", nil, nil, nil, nil)
	_ = model.DesignAggregate(agg)
	_ = model.Finalize()

	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)
	assert.Equal(t, 0, plan.Tickets()[0].Depth())
}

// ---------------------------------------------------------------------------
// Depth-based reclassification
// ---------------------------------------------------------------------------

func TestSupportingAtDepth2StaysStandard(t *testing.T) {
	t.Parallel()
	model := makeChainModel(3, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	var found bool
	for _, ticket := range plan.Tickets() {
		if ticket.Depth() == 2 {
			assert.Equal(t, vo.TicketDetailStandard, ticket.DetailLevel())
			found = true
		}
	}
	assert.True(t, found)
}

func TestSupportingAtDepth3BecomesStub(t *testing.T) {
	t.Parallel()
	model := makeChainModel(4, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	var found bool
	for _, ticket := range plan.Tickets() {
		if ticket.Depth() == 3 {
			assert.Equal(t, vo.TicketDetailStub, ticket.DetailLevel())
			found = true
		}
	}
	assert.True(t, found)
}

func TestCoreAtDepth5StaysFull(t *testing.T) {
	t.Parallel()
	model := makeChainModel(6, vo.SubdomainCore)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	var found bool
	for _, ticket := range plan.Tickets() {
		if ticket.Depth() == 5 {
			assert.Equal(t, vo.TicketDetailFull, ticket.DetailLevel())
			found = true
		}
	}
	assert.True(t, found)
}

func TestStubDescriptionEnrichedAtDepth3(t *testing.T) {
	t.Parallel()
	model := makeChainModel(4, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	for _, ticket := range plan.Tickets() {
		if ticket.Depth() == 3 {
			assert.Equal(t, vo.TicketDetailStub, ticket.DetailLevel())
			assert.Contains(t, ticket.Description(), "Stub ticket")
			assert.Contains(t, ticket.Description(), "## DDD Alignment")
		}
	}
}

// ---------------------------------------------------------------------------
// Promotion eligibility
// ---------------------------------------------------------------------------

func TestStubEligibleWhenAllDepsResolved(t *testing.T) {
	t.Parallel()
	model := makeChainModel(4, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	var stub domain.GeneratedTicket
	for _, ticket := range plan.Tickets() {
		if ticket.DetailLevel() == vo.TicketDetailStub {
			stub = ticket
			break
		}
	}

	resolved := map[string]bool{}
	for _, ticket := range plan.Tickets() {
		if ticket.TicketID() != stub.TicketID() {
			resolved[ticket.TicketID()] = true
		}
	}
	eligible := plan.PromotionEligibleIDs(resolved)
	assert.True(t, eligible[stub.TicketID()])
}

func TestStubNotEligibleWhenDepsUnresolved(t *testing.T) {
	t.Parallel()
	model := makeChainModel(4, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	var stub domain.GeneratedTicket
	for _, ticket := range plan.Tickets() {
		if ticket.DetailLevel() == vo.TicketDetailStub {
			stub = ticket
			break
		}
	}

	eligible := plan.PromotionEligibleIDs(map[string]bool{})
	assert.False(t, eligible[stub.TicketID()])
}

func TestNonStubNotEligible(t *testing.T) {
	t.Parallel()
	model := makeChainModel(2, vo.SubdomainSupporting)
	plan := domain.NewTicketPlan()
	_ = plan.GeneratePlan(model, nil)

	allIDs := map[string]bool{}
	for _, ticket := range plan.Tickets() {
		allIDs[ticket.TicketID()] = true
	}
	eligible := plan.PromotionEligibleIDs(allIDs)
	assert.Empty(t, eligible)
}
