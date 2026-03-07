package ddd_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// makeValidModel creates a DomainModel that passes all 4 invariants.
func makeValidModel() *ddd.DomainModel {
	model := ddd.NewDomainModel("test-id")

	_ = model.AddDomainStory(vo.NewDomainStory(
		"Checkout Flow",
		[]string{"Customer"},
		"Customer clicks checkout",
		[]string{
			"Customer reviews order",
			"System validates payment",
			"System creates shipment",
		},
		nil,
	))

	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Manages orders", nil, nil, ""))
	_ = model.ClassifySubdomain("Sales", vo.SubdomainCore, "Competitive advantage")

	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Shipping", "Manages shipments", nil, nil, ""))
	_ = model.ClassifySubdomain("Shipping", vo.SubdomainSupporting, "Necessary plumbing")

	_ = model.AddTerm("Order", "A customer purchase", "Sales", []string{"Q2"})
	_ = model.AddTerm("Shipment", "A delivery package", "Shipping", []string{"Q2"})

	_ = model.DesignAggregate(vo.NewAggregateDesign(
		"OrderAggregate", "Sales", "Order",
		nil, []string{"Total must be positive"}, nil, nil,
	))

	_ = model.AddContextRelationship(vo.NewContextRelationship("Sales", "Shipping", "Domain Events"))
	_ = model.AddContextRelationship(vo.NewContextRelationship("Shipping", "Sales", "Query"))

	return model
}

func makeValidModelWithoutRelationships() *ddd.DomainModel {
	model := ddd.NewDomainModel("test-id")

	_ = model.AddDomainStory(vo.NewDomainStory(
		"Checkout Flow",
		[]string{"Customer"},
		"Customer clicks checkout",
		[]string{
			"Customer reviews order",
			"System validates payment",
			"System creates shipment",
		},
		nil,
	))

	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Manages orders", nil, nil, ""))
	_ = model.ClassifySubdomain("Sales", vo.SubdomainCore, "Competitive advantage")

	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Shipping", "Manages shipments", nil, nil, ""))
	_ = model.ClassifySubdomain("Shipping", vo.SubdomainSupporting, "Necessary plumbing")

	_ = model.AddTerm("Order", "A customer purchase", "Sales", []string{"Q2"})
	_ = model.AddTerm("Shipment", "A delivery package", "Shipping", []string{"Q2"})

	_ = model.DesignAggregate(vo.NewAggregateDesign(
		"OrderAggregate", "Sales", "Order",
		nil, []string{"Total must be positive"}, nil, nil,
	))

	return model
}

// ---------------------------------------------------------------------------
// CreateDomainModel
// ---------------------------------------------------------------------------

func TestEmptyModel(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test-id")
	assert.Empty(t, model.DomainStories())
	assert.Empty(t, model.BoundedContexts())
	assert.Empty(t, model.AggregateDesigns())
	assert.Empty(t, model.Events())
}

func TestModelIDGenerated(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test-id")
	assert.NotEmpty(t, model.ModelID())
}

// ---------------------------------------------------------------------------
// AddDomainStory
// ---------------------------------------------------------------------------

func TestAddStory(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.AddDomainStory(vo.NewDomainStory("Test", []string{"A"}, "T", []string{"S"}, nil))
	require.NoError(t, err)
	assert.Len(t, model.DomainStories(), 1)
	assert.Equal(t, "Test", model.DomainStories()[0].Name())
}

func TestDuplicateStoryRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	story := vo.NewDomainStory("Test", []string{"A"}, "T", []string{"S"}, nil)
	require.NoError(t, model.AddDomainStory(story))
	err := model.AddDomainStory(story)
	require.ErrorIs(t, err, domainerrors.ErrAlreadyExists)
	assert.Contains(t, err.Error(), "'Test' already exists")
}

func TestDuplicateStoryCaseInsensitive(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	require.NoError(t, model.AddDomainStory(vo.NewDomainStory("Test", []string{"A"}, "T", []string{"S"}, nil)))
	err := model.AddDomainStory(vo.NewDomainStory("test", []string{"B"}, "T2", []string{"S2"}, nil))
	require.ErrorIs(t, err, domainerrors.ErrAlreadyExists)
}

func TestEmptyStoryNameRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.AddDomainStory(vo.NewDomainStory("", []string{"A"}, "T", []string{"S"}, nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "story name cannot be empty")
}

// ---------------------------------------------------------------------------
// AddBoundedContext
// ---------------------------------------------------------------------------

func TestAddContext(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, ""))
	require.NoError(t, err)
	assert.Len(t, model.BoundedContexts(), 1)
}

func TestDuplicateContextRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, "")))
	err := model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Other", nil, nil, ""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'Sales' already exists")
}

func TestEmptyContextNameRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.AddBoundedContext(vo.NewDomainBoundedContext("", "X", nil, nil, ""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context name cannot be empty")
}

// ---------------------------------------------------------------------------
// ClassifySubdomain
// ---------------------------------------------------------------------------

func TestClassify(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, "")))
	require.NoError(t, model.ClassifySubdomain("Sales", vo.SubdomainCore, "Key value"))
	assert.Equal(t, vo.SubdomainCore, *model.BoundedContexts()[0].Classification())
}

func TestUnknownContextRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.ClassifySubdomain("Missing", vo.SubdomainCore, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// DesignAggregate
// ---------------------------------------------------------------------------

func TestAddAggregate(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.DesignAggregate(vo.NewAggregateDesign("OrderAgg", "Sales", "Order", nil, nil, nil, nil))
	require.NoError(t, err)
	assert.Len(t, model.AggregateDesigns(), 1)
}

func TestEmptyAggregateNameRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.DesignAggregate(vo.NewAggregateDesign("", "Sales", "X", nil, nil, nil, nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "aggregate name cannot be empty")
}

func TestEmptyAggregateContextRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.DesignAggregate(vo.NewAggregateDesign("Agg", "", "X", nil, nil, nil, nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context name cannot be empty")
}

func TestDuplicateAggregateNameSameContextRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	require.NoError(t, model.DesignAggregate(vo.NewAggregateDesign("OrderAgg", "Sales", "Order", nil, nil, nil, nil)))
	err := model.DesignAggregate(vo.NewAggregateDesign("OrderAgg", "Sales", "Other", nil, nil, nil, nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'OrderAgg'")
	assert.Contains(t, err.Error(), "already exists")
}

func TestSameNameDifferentContextAllowed(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	require.NoError(t, model.DesignAggregate(vo.NewAggregateDesign("RootAgg", "Sales", "Order", nil, nil, nil, nil)))
	require.NoError(t, model.DesignAggregate(vo.NewAggregateDesign("RootAgg", "Shipping", "Shipment", nil, nil, nil, nil)))
	assert.Len(t, model.AggregateDesigns(), 2)
}

func TestDuplicateAggregateCaseInsensitive(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	require.NoError(t, model.DesignAggregate(vo.NewAggregateDesign("OrderAgg", "Sales", "Order", nil, nil, nil, nil)))
	err := model.DesignAggregate(vo.NewAggregateDesign("orderagg", "sales", "Other", nil, nil, nil, nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// ---------------------------------------------------------------------------
// AddContextRelationship
// ---------------------------------------------------------------------------

func TestAddRelationship(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.AddContextRelationship(vo.NewContextRelationship("A", "B", "Events"))
	require.NoError(t, err)
	assert.Len(t, model.ContextRelationships(), 1)
}

func TestEmptyUpstreamRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.AddContextRelationship(vo.NewContextRelationship("", "B", "Events"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// ---------------------------------------------------------------------------
// Finalize: Invariant 1 — Terms in Stories
// ---------------------------------------------------------------------------

func TestTermInStoryPasses(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
	assert.Len(t, model.Events(), 1)
}

func TestTermNotInStoryRaises(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	_ = model.AddTerm("Widget", "A widget thing", "Sales", []string{"Q2"})
	err := model.Finalize()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "term 'Widget' not found in any domain story")
}

// ---------------------------------------------------------------------------
// Finalize: Invariant 2 — Context Classification
// ---------------------------------------------------------------------------

func TestAllClassifiedPasses(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
}

func TestUnclassifiedContextRaises(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Unclassified", "Test", nil, nil, ""))
	err := model.Finalize()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "'Unclassified' has no classification")
}

// ---------------------------------------------------------------------------
// Finalize: Invariant 3 — Core Aggregates
// ---------------------------------------------------------------------------

func TestCoreWithAggregatePasses(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
}

func TestCoreWithoutAggregateRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Test Story", []string{"Actor"}, "Start",
		[]string{"Actor does discovery"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Discovery", "Guides", nil, nil, ""))
	_ = model.ClassifySubdomain("Discovery", vo.SubdomainCore, "")
	_ = model.AddTerm("Discovery", "The process", "Discovery", nil)
	err := model.Finalize()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "core subdomain 'Discovery' has no aggregate")
}

// ---------------------------------------------------------------------------
// Finalize: Invariant 4 — Ambiguous Terms
// ---------------------------------------------------------------------------

func TestNoAmbiguityPasses(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
}

func TestAmbiguousWithDefinitionsPasses(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	_ = model.AddTerm("Order", "Work order", "Shipping", nil)
	require.NoError(t, model.Finalize())
}

func TestAmbiguousWithoutDefinitionsRaises(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	// Bypass validation to add term with empty definition.
	model.UbiquitousLanguage().AddTermEntry(
		vo.NewTermEntry("Order", "", "Shipping", nil),
	)
	err := model.Finalize()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "ambiguous term 'order' needs per-context")
}

// ---------------------------------------------------------------------------
// Finalize: Emits Event
// ---------------------------------------------------------------------------

func TestEmitsDomainModelGenerated(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
	assert.Len(t, model.Events(), 1)
	event := model.Events()[0]
	assert.Equal(t, model.ModelID(), event.ModelID())
	assert.Len(t, event.DomainStories(), 1)
	assert.Len(t, event.BoundedContexts(), 2)
	assert.Len(t, event.AggregateDesigns(), 1)
}

// ---------------------------------------------------------------------------
// Defensive Copies
// ---------------------------------------------------------------------------

func TestStoriesDefensiveCopy(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory("S", []string{"A"}, "T", []string{"S"}, nil))
	s1 := model.DomainStories()
	s2 := model.DomainStories()
	assert.NotSame(t, &s1[0], &s2[0])
}

func TestContextsDefensiveCopy(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Ctx", "Test", nil, nil, ""))
	c1 := model.BoundedContexts()
	c2 := model.BoundedContexts()
	assert.NotSame(t, &c1[0], &c2[0])
}

func TestEventsDefensiveCopy(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
	e1 := model.Events()
	e2 := model.Events()
	assert.NotSame(t, &e1[0], &e2[0])
}

// ---------------------------------------------------------------------------
// ReassignTermsToContext
// ---------------------------------------------------------------------------

func TestReassignMovesAllMatchingTerms(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddTerm("Order", "A purchase", "Default", nil)
	_ = model.AddTerm("Product", "An item", "Default", nil)
	require.NoError(t, model.ReassignTermsToContext("Default", "Sales"))
	sales := model.UbiquitousLanguage().GetTermsForContext("Sales")
	assert.Len(t, sales, 2)
}

func TestReassignNoMatchIsNoop(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddTerm("Order", "A purchase", "Sales", nil)
	require.NoError(t, model.ReassignTermsToContext("NonExistent", "Other"))
	assert.Len(t, model.UbiquitousLanguage().GetTermsForContext("Sales"), 1)
	assert.Empty(t, model.UbiquitousLanguage().GetTermsForContext("Other"))
}

func TestReassignPreservesOtherContextTerms(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddTerm("Order", "A purchase", "Default", nil)
	_ = model.AddTerm("Product", "An item", "Catalog", nil)
	require.NoError(t, model.ReassignTermsToContext("Default", "Sales"))
	assert.Len(t, model.UbiquitousLanguage().GetTermsForContext("Sales"), 1)
	assert.Len(t, model.UbiquitousLanguage().GetTermsForContext("Catalog"), 1)
}

func TestReassignPreservesDefinitionAndSource(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddTerm("Order", "A customer purchase", "Default", []string{"Q2"})
	require.NoError(t, model.ReassignTermsToContext("Default", "Sales"))
	term := model.UbiquitousLanguage().GetTermsForContext("Sales")[0]
	assert.Equal(t, "A customer purchase", term.Definition())
	assert.Equal(t, []string{"Q2"}, term.SourceQuestionIDs())
}

func TestReassignEmptyFromRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.ReassignTermsToContext("", "Sales")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestReassignEmptyToRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.ReassignTermsToContext("Default", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestReassignWhitespaceFromRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.ReassignTermsToContext("   ", "Sales")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestReassignWhitespaceToRaises(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	err := model.ReassignTermsToContext("Default", "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestReassignCaseInsensitiveMatch(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddTerm("Order", "A purchase", "default", nil)
	require.NoError(t, model.ReassignTermsToContext("Default", "Sales"))
	assert.Len(t, model.UbiquitousLanguage().GetTermsForContext("Sales"), 1)
	assert.Empty(t, model.UbiquitousLanguage().GetTermsForContext("default"))
}

// ---------------------------------------------------------------------------
// Relaxed Relationships
// ---------------------------------------------------------------------------

func TestUnidirectionalPassesFinalize(t *testing.T) {
	t.Parallel()
	model := makeValidModelWithoutRelationships()
	_ = model.AddContextRelationship(vo.NewContextRelationship("Sales", "Shipping", "Domain Events"))
	require.NoError(t, model.Finalize())
	assert.NotEmpty(t, model.Events())
}

func TestNoRelationshipsPassesFinalize(t *testing.T) {
	t.Parallel()
	model := makeValidModelWithoutRelationships()
	require.NoError(t, model.Finalize())
	assert.NotEmpty(t, model.Events())
}

func TestBidirectionalStillAccepted(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
}

// ---------------------------------------------------------------------------
// Events Returns Slice
// ---------------------------------------------------------------------------

func TestEventsIsSliceAfterFinalize(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
	assert.NotNil(t, model.Events())
}

func TestEmptyEventsIsSlice(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	assert.NotNil(t, model.Events())
}

// ---------------------------------------------------------------------------
// Word Boundary Term Matching
// ---------------------------------------------------------------------------

func TestShortTermNotSubstringOfLongerWord(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Flow", []string{"Operator"}, "Start",
		[]string{"Operator processes work"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Ops", "Operations", nil, nil, ""))
	_ = model.ClassifySubdomain("Ops", vo.SubdomainSupporting, "")
	_ = model.AddTerm("Or", "Logical operator concept", "Ops", nil)
	err := model.Finalize()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "term 'Or'")
}

func TestExactWordMatchPasses(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Checkout", []string{"Customer"}, "Customer clicks buy",
		[]string{"Customer creates Order"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Handles orders", nil, nil, ""))
	_ = model.ClassifySubdomain("Sales", vo.SubdomainSupporting, "")
	_ = model.AddTerm("Order", "A purchase", "Sales", nil)
	require.NoError(t, model.Finalize())
}

func TestTermAtStartOfStoryName(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Order Flow", []string{"User"}, "Start",
		[]string{"User submits"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, ""))
	_ = model.ClassifySubdomain("Sales", vo.SubdomainSupporting, "")
	_ = model.AddTerm("Order", "A purchase", "Sales", nil)
	require.NoError(t, model.Finalize())
}

func TestMultiWordTermMatching(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Payment", []string{"User"}, "User pays",
		[]string{"System processes sales order"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, ""))
	_ = model.ClassifySubdomain("Sales", vo.SubdomainSupporting, "")
	_ = model.AddTerm("Sales Order", "A customer order", "Sales", nil)
	require.NoError(t, model.Finalize())
}

func TestTermInObservationsMatches(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Flow", []string{"User"}, "Start",
		[]string{"User acts"},
		[]string{"System creates Invoice"},
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Billing", "Invoicing", nil, nil, ""))
	_ = model.ClassifySubdomain("Billing", vo.SubdomainSupporting, "")
	_ = model.AddTerm("Invoice", "A bill", "Billing", nil)
	require.NoError(t, model.Finalize())
}

func TestTermInActorListMatches(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Flow", []string{"Admin"}, "Start",
		[]string{"Admin reviews"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Mgmt", "Management", nil, nil, ""))
	_ = model.ClassifySubdomain("Mgmt", vo.SubdomainSupporting, "")
	_ = model.AddTerm("Admin", "An administrator", "Mgmt", nil)
	require.NoError(t, model.Finalize())
}

// ---------------------------------------------------------------------------
// All-Generic Warning
// ---------------------------------------------------------------------------

func TestAllGenericProducesWarning(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Simple Flow", []string{"User"}, "User starts",
		[]string{"User uses Auth", "User uses Logging"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Auth", "Auth", nil, nil, ""))
	_ = model.ClassifySubdomain("Auth", vo.SubdomainGeneric, "Off-the-shelf")
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Logging", "Logs", nil, nil, ""))
	_ = model.ClassifySubdomain("Logging", vo.SubdomainGeneric, "Off-the-shelf")
	_ = model.AddTerm("Auth", "Authentication", "Auth", nil)
	_ = model.AddTerm("Logging", "Application logs", "Logging", nil)

	require.NoError(t, model.Finalize())
	assert.NotEmpty(t, model.Warnings())
	hasGenericWarning := false
	for _, w := range model.Warnings() {
		if strings.Contains(strings.ToLower(w), "generic") {
			hasGenericWarning = true
			break
		}
	}
	assert.True(t, hasGenericWarning)
}

func TestAtLeastOneCoreNoWarning(t *testing.T) {
	t.Parallel()
	model := makeValidModel()
	require.NoError(t, model.Finalize())
	for _, w := range model.Warnings() {
		assert.NotContains(t, strings.ToLower(w), "generic")
	}
}

func TestMixedSupportingGenericNoWarning(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	_ = model.AddDomainStory(vo.NewDomainStory(
		"Flow", []string{"User"}, "User starts",
		[]string{"User uses Billing", "User uses Auth"}, nil,
	))
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Billing", "Bills", nil, nil, ""))
	_ = model.ClassifySubdomain("Billing", vo.SubdomainSupporting, "Needed")
	_ = model.AddBoundedContext(vo.NewDomainBoundedContext("Auth", "Auth", nil, nil, ""))
	_ = model.ClassifySubdomain("Auth", vo.SubdomainGeneric, "Commodity")
	_ = model.AddTerm("Billing", "Payment processing", "Billing", nil)
	_ = model.AddTerm("Auth", "Authentication", "Auth", nil)

	require.NoError(t, model.Finalize())
	for _, w := range model.Warnings() {
		assert.NotContains(t, strings.ToLower(w), "generic")
	}
}

func TestWarningsPropertyExistsBeforeFinalize(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("test")
	assert.Empty(t, model.Warnings())
}
