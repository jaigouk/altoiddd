package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// SubdomainClassification
// ---------------------------------------------------------------------------

func TestSubdomainClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		class vo.SubdomainClassification
		want  string
	}{
		{"core value", vo.SubdomainCore, "core"},
		{"supporting value", vo.SubdomainSupporting, "supporting"},
		{"generic value", vo.SubdomainGeneric, "generic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, string(tt.class))
		})
	}
}

func TestSubdomainClassificationAllMembers(t *testing.T) {
	t.Parallel()

	all := vo.AllSubdomainClassifications()
	values := make(map[string]struct{}, len(all))
	for _, c := range all {
		values[string(c)] = struct{}{}
	}
	expected := map[string]struct{}{
		"core":       {},
		"supporting": {},
		"generic":    {},
	}
	assert.Equal(t, expected, values)
}

// ---------------------------------------------------------------------------
// DomainStory
// ---------------------------------------------------------------------------

func TestDomainStoryCreateMinimal(t *testing.T) {
	t.Parallel()

	story := vo.NewDomainStory(
		"Checkout",
		[]string{"Customer"},
		"Customer clicks checkout",
		[]string{"Customer reviews cart"},
		nil,
	)
	assert.Equal(t, "Checkout", story.Name())
	assert.Equal(t, []string{"Customer"}, story.Actors())
	assert.Equal(t, []string{}, story.Observations())
}

func TestDomainStoryWithObservations(t *testing.T) {
	t.Parallel()

	story := vo.NewDomainStory(
		"Test",
		[]string{"A"},
		"T",
		[]string{"S"},
		[]string{"Surprising finding"},
	)
	assert.Equal(t, []string{"Surprising finding"}, story.Observations())
}

func TestDomainStoryDefensiveCopy(t *testing.T) {
	t.Parallel()

	story := vo.NewDomainStory("Test", []string{"A"}, "T", []string{"S"}, nil)
	actors := story.Actors()
	actors[0] = "MODIFIED"
	assert.Equal(t, "A", story.Actors()[0], "Actors() must return a defensive copy")
}

// ---------------------------------------------------------------------------
// DomainBoundedContext
// ---------------------------------------------------------------------------

func TestDomainBoundedContextWithoutClassification(t *testing.T) {
	t.Parallel()

	ctx := vo.NewDomainBoundedContext("Orders", "Manages orders", nil, nil, "")
	assert.Equal(t, "Orders", ctx.Name())
	assert.Equal(t, "Manages orders", ctx.Responsibility())
	assert.Nil(t, ctx.Classification())
	assert.Empty(t, ctx.ClassificationRationale())
}

func TestDomainBoundedContextWithClassification(t *testing.T) {
	t.Parallel()

	core := vo.SubdomainCore
	ctx := vo.NewDomainBoundedContext(
		"Orders",
		"Manages orders",
		nil,
		&core,
		"Competitive advantage",
	)
	assert.NotNil(t, ctx.Classification())
	assert.Equal(t, vo.SubdomainCore, *ctx.Classification())
}

func TestDomainBoundedContextKeyDomainObjectsDefensiveCopy(t *testing.T) {
	t.Parallel()

	ctx := vo.NewDomainBoundedContext("Orders", "test", []string{"Order", "Line"}, nil, "")
	objs := ctx.KeyDomainObjects()
	objs[0] = "MODIFIED"
	assert.Equal(t, "Order", ctx.KeyDomainObjects()[0])
}

// ---------------------------------------------------------------------------
// ContextRelationship
// ---------------------------------------------------------------------------

func TestContextRelationshipCreate(t *testing.T) {
	t.Parallel()

	rel := vo.NewContextRelationship("Orders", "Shipping", "Domain Events")
	assert.Equal(t, "Orders", rel.Upstream())
	assert.Equal(t, "Shipping", rel.Downstream())
	assert.Equal(t, "Domain Events", rel.IntegrationPattern())
}

// ---------------------------------------------------------------------------
// AggregateDesign
// ---------------------------------------------------------------------------

func TestAggregateDesignCreateMinimal(t *testing.T) {
	t.Parallel()

	agg := vo.NewAggregateDesign("OrderAggregate", "Orders", "Order", nil, nil, nil, nil)
	assert.Equal(t, "OrderAggregate", agg.Name())
	assert.Equal(t, "Orders", agg.ContextName())
	assert.Equal(t, "Order", agg.RootEntity())
	assert.Equal(t, []string{}, agg.ContainedObjects())
	assert.Equal(t, []string{}, agg.Invariants())
}

func TestAggregateDesignCreateFull(t *testing.T) {
	t.Parallel()

	agg := vo.NewAggregateDesign(
		"OrderAggregate",
		"Orders",
		"Order",
		[]string{"OrderLine", "ShippingAddress"},
		[]string{"Total cannot be negative"},
		[]string{"place_order", "cancel_order"},
		[]string{"OrderPlaced", "OrderCancelled"},
	)
	assert.Len(t, agg.ContainedObjects(), 2)
	assert.Len(t, agg.Commands(), 2)
	assert.Len(t, agg.DomainEvents(), 2)
}

func TestAggregateDesignDefensiveCopy(t *testing.T) {
	t.Parallel()

	agg := vo.NewAggregateDesign(
		"OrderAggregate", "Orders", "Order",
		[]string{"OrderLine"}, nil, nil, nil,
	)
	objs := agg.ContainedObjects()
	objs[0] = "MODIFIED"
	assert.Equal(t, "OrderLine", agg.ContainedObjects()[0])
}
