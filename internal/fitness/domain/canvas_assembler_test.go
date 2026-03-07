package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/fitness/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeFinalizedModel(t *testing.T, opts modelOpts) *ddd.DomainModel {
	t.Helper()
	m := ddd.NewDomainModel("test-canvas")

	for _, ctx := range opts.contexts {
		err := m.AddBoundedContext(ctx.bc)
		require.NoError(t, err)
	}

	for _, rel := range opts.relationships {
		err := m.AddContextRelationship(rel)
		require.NoError(t, err)
	}

	// Build default story if not provided
	stories := opts.stories
	if stories == nil {
		var termWords []string
		for _, te := range opts.terms {
			termWords = append(termWords, te.term)
		}
		if len(termWords) == 0 {
			termWords = []string{"placeholder"}
		}
		stepsText := strings.Join(termWords, ", ")
		stories = []vo.DomainStory{
			vo.NewDomainStory("Default Story", []string{"Actor"}, "System processes "+stepsText,
				[]string{"Actor manages " + stepsText}, nil),
		}
	}
	for _, story := range stories {
		err := m.AddDomainStory(story)
		require.NoError(t, err)
	}

	for _, te := range opts.terms {
		err := m.AddTerm(te.term, te.definition, te.contextName, nil)
		require.NoError(t, err)
	}

	for _, ctx := range opts.contexts {
		if ctx.classification != nil {
			err := m.ClassifySubdomain(ctx.bc.Name(), *ctx.classification, "")
			require.NoError(t, err)
		}
	}

	for _, agg := range opts.aggregates {
		err := m.DesignAggregate(agg)
		require.NoError(t, err)
	}

	err := m.Finalize()
	require.NoError(t, err)
	return m
}

type contextEntry struct {
	classification *vo.SubdomainClassification
	bc             vo.DomainBoundedContext
}

type termEntry struct {
	term        string
	definition  string
	contextName string
}

type modelOpts struct {
	contexts      []contextEntry
	relationships []vo.ContextRelationship
	aggregates    []vo.AggregateDesign
	stories       []vo.DomainStory
	terms         []termEntry
}

func ctxEntry(name string, responsibility string, cl vo.SubdomainClassification) contextEntry {
	return contextEntry{
		classification: &cl,
		bc:             vo.NewDomainBoundedContext(name, responsibility, nil, nil, ""),
	}
}

// ---------------------------------------------------------------------------
// Assemble empty model
// ---------------------------------------------------------------------------

func TestAssembleEmptyModel(t *testing.T) {
	t.Parallel()
	m := makeFinalizedModel(t, modelOpts{})
	result := domain.Assemble(m)
	assert.Empty(t, result)
}

// ---------------------------------------------------------------------------
// Single context
// ---------------------------------------------------------------------------

func TestAssembleSingleContext(t *testing.T) {
	t.Parallel()

	t.Run("single core context", func(t *testing.T) {
		t.Parallel()
		m := makeFinalizedModel(t, modelOpts{
			contexts:   []contextEntry{ctxEntry("Sales", "Manages order lifecycle", vo.SubdomainCore)},
			aggregates: []vo.AggregateDesign{vo.NewAggregateDesign("SalesRoot", "Sales", "SalesRoot", nil, nil, nil, nil)},
			terms:      []termEntry{{"Order", "A purchase request", "Sales"}},
		})
		canvases := domain.Assemble(m)
		require.Len(t, canvases, 1)
		assert.Equal(t, "Sales", canvases[0].ContextName())
		assert.Equal(t, "Manages order lifecycle", canvases[0].Purpose())
		assert.Contains(t, canvases[0].Roles(), domain.RoleExecution)
	})

	t.Run("single supporting context", func(t *testing.T) {
		t.Parallel()
		m := makeFinalizedModel(t, modelOpts{
			contexts: []contextEntry{ctxEntry("Notifications", "Sends alerts", vo.SubdomainSupporting)},
		})
		canvases := domain.Assemble(m)
		require.Len(t, canvases, 1)
		assert.Contains(t, canvases[0].Roles(), domain.RoleSpecification)
	})

	t.Run("single generic context", func(t *testing.T) {
		t.Parallel()
		m := makeFinalizedModel(t, modelOpts{
			contexts: []contextEntry{ctxEntry("Logging", "Records events", vo.SubdomainGeneric)},
		})
		canvases := domain.Assemble(m)
		require.Len(t, canvases, 1)
		assert.Contains(t, canvases[0].Roles(), domain.RoleGateway)
	})
}

// ---------------------------------------------------------------------------
// Missing classification (fallback to generic)
// ---------------------------------------------------------------------------

func TestAssembleMissingClassification(t *testing.T) {
	t.Parallel()
	// Build unfinalized model with nil classification
	m := ddd.NewDomainModel("test-nil-class")
	_ = m.AddBoundedContext(vo.NewDomainBoundedContext("Orphan", "Unclassified context", nil, nil, ""))
	_ = m.AddDomainStory(vo.NewDomainStory("Orphan Story", []string{"Actor"}, "System starts",
		[]string{"Actor uses Orphan"}, nil))
	// Don't finalize — test assembler directly
	canvases := domain.Assemble(m)
	require.Len(t, canvases, 1)
	assert.Equal(t, vo.SubdomainGeneric, canvases[0].Classification().Domain())
	assert.Equal(t, "unclassified", canvases[0].Classification().BusinessModel())
}

// ---------------------------------------------------------------------------
// Multiple contexts
// ---------------------------------------------------------------------------

func TestAssembleMultipleContexts(t *testing.T) {
	t.Parallel()
	m := makeFinalizedModel(t, modelOpts{
		contexts: []contextEntry{
			ctxEntry("Sales", "Order management", vo.SubdomainCore),
			ctxEntry("Inventory", "Stock tracking", vo.SubdomainSupporting),
		},
		aggregates: []vo.AggregateDesign{vo.NewAggregateDesign("SalesRoot", "Sales", "SalesRoot", nil, nil, nil, nil)},
		terms: []termEntry{
			{"Order", "A purchase request", "Sales"},
			{"Stock", "Available inventory", "Inventory"},
		},
	})
	canvases := domain.Assemble(m)
	assert.Len(t, canvases, 2)
	names := make(map[string]bool)
	for _, c := range canvases {
		names[c.ContextName()] = true
	}
	assert.True(t, names["Sales"])
	assert.True(t, names["Inventory"])
}

// ---------------------------------------------------------------------------
// Communication mapping
// ---------------------------------------------------------------------------

func TestAssembleCommunication(t *testing.T) {
	t.Parallel()
	m := makeFinalizedModel(t, modelOpts{
		contexts: []contextEntry{
			ctxEntry("Sales", "Orders", vo.SubdomainCore),
			ctxEntry("Shipping", "Delivery", vo.SubdomainSupporting),
		},
		relationships: []vo.ContextRelationship{
			vo.NewContextRelationship("Sales", "Shipping", "Domain Events"),
		},
		aggregates: []vo.AggregateDesign{vo.NewAggregateDesign("SalesRoot", "Sales", "SalesRoot", nil, nil, nil, nil)},
		terms:      []termEntry{{"Order", "A purchase", "Sales"}},
	})
	canvases := domain.Assemble(m)
	var salesCanvas, shippingCanvas domain.BoundedContextCanvas
	for _, c := range canvases {
		if c.ContextName() == "Sales" {
			salesCanvas = c
		} else {
			shippingCanvas = c
		}
	}

	// Sales is upstream → outbound
	assert.NotEmpty(t, salesCanvas.OutboundCommunication())
	found := false
	for _, msg := range salesCanvas.OutboundCommunication() {
		if msg.Counterpart() == "Shipping" {
			found = true
		}
	}
	assert.True(t, found)

	// Shipping is downstream → inbound
	assert.NotEmpty(t, shippingCanvas.InboundCommunication())
	found = false
	for _, msg := range shippingCanvas.InboundCommunication() {
		if msg.Counterpart() == "Sales" {
			found = true
		}
	}
	assert.True(t, found)
}

// ---------------------------------------------------------------------------
// UL filtering
// ---------------------------------------------------------------------------

func TestAssembleUbiquitousLanguage(t *testing.T) {
	t.Parallel()

	t.Run("terms filtered by context", func(t *testing.T) {
		t.Parallel()
		m := makeFinalizedModel(t, modelOpts{
			contexts: []contextEntry{
				ctxEntry("Sales", "Orders", vo.SubdomainCore),
				ctxEntry("Inventory", "Stock", vo.SubdomainSupporting),
			},
			aggregates: []vo.AggregateDesign{vo.NewAggregateDesign("SalesRoot", "Sales", "SalesRoot", nil, nil, nil, nil)},
			terms: []termEntry{
				{"Order", "A purchase request", "Sales"},
				{"Stock", "Available items", "Inventory"},
			},
		})
		canvases := domain.Assemble(m)
		var salesCanvas, invCanvas domain.BoundedContextCanvas
		for _, c := range canvases {
			if c.ContextName() == "Sales" {
				salesCanvas = c
			} else {
				invCanvas = c
			}
		}

		salesTerms := make(map[string]string)
		for _, pair := range salesCanvas.UbiquitousLanguage() {
			salesTerms[pair[0]] = pair[1]
		}
		invTerms := make(map[string]string)
		for _, pair := range invCanvas.UbiquitousLanguage() {
			invTerms[pair[0]] = pair[1]
		}

		assert.Contains(t, salesTerms, "Order")
		assert.NotContains(t, salesTerms, "Stock")
		assert.Contains(t, invTerms, "Stock")
		assert.NotContains(t, invTerms, "Order")
	})

	t.Run("empty ul for context", func(t *testing.T) {
		t.Parallel()
		m := makeFinalizedModel(t, modelOpts{
			contexts: []contextEntry{ctxEntry("Logging", "Records events", vo.SubdomainGeneric)},
		})
		canvases := domain.Assemble(m)
		assert.Empty(t, canvases[0].UbiquitousLanguage())
	})
}

// ---------------------------------------------------------------------------
// Business decisions
// ---------------------------------------------------------------------------

func TestAssembleBusinessDecisions(t *testing.T) {
	t.Parallel()
	m := makeFinalizedModel(t, modelOpts{
		contexts: []contextEntry{ctxEntry("Sales", "Orders", vo.SubdomainCore)},
		aggregates: []vo.AggregateDesign{
			vo.NewAggregateDesign("SalesRoot", "Sales", "SalesRoot", nil,
				[]string{"Order must have items", "Payment must be positive"}, nil, nil),
		},
		terms: []termEntry{{"Order", "A purchase", "Sales"}},
	})
	canvases := domain.Assemble(m)
	assert.Contains(t, canvases[0].BusinessDecisions(), "Order must have items")
	assert.Contains(t, canvases[0].BusinessDecisions(), "Payment must be positive")
}

// ---------------------------------------------------------------------------
// Assumptions and questions empty
// ---------------------------------------------------------------------------

func TestAssembleAssumptionsAndQuestions(t *testing.T) {
	t.Parallel()
	m := makeFinalizedModel(t, modelOpts{
		contexts:   []contextEntry{ctxEntry("Sales", "Orders", vo.SubdomainCore)},
		aggregates: []vo.AggregateDesign{vo.NewAggregateDesign("SalesRoot", "Sales", "SalesRoot", nil, nil, nil, nil)},
	})
	canvases := domain.Assemble(m)
	assert.Empty(t, canvases[0].Assumptions())
	assert.Empty(t, canvases[0].OpenQuestions())
}

// ---------------------------------------------------------------------------
// Render markdown
// ---------------------------------------------------------------------------

func TestRenderMarkdownEmpty(t *testing.T) {
	t.Parallel()
	result := domain.RenderMarkdown(nil)
	assert.Empty(t, result)
}

func TestRenderMarkdownSingleCanvas(t *testing.T) {
	t.Parallel()

	makeCanvas := func(opts ...func(*domain.BoundedContextCanvas)) domain.BoundedContextCanvas {
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
		c := domain.NewBoundedContextCanvas("Sales", "Manages orders", sc,
			[]domain.Role{domain.RoleExecution}, nil, nil, nil, nil, nil, nil)
		for _, opt := range opts {
			opt(&c)
		}
		return c
	}

	t.Run("contains context name heading", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
		canvas := domain.NewBoundedContextCanvas("Sales", "Manages orders", sc,
			[]domain.Role{domain.RoleExecution},
			[]domain.CommunicationMessage{domain.NewCommunicationMessage("PlaceOrder", "Command", "API Gateway")},
			[]domain.CommunicationMessage{domain.NewCommunicationMessage("OrderPlaced", "Event", "Fulfillment")},
			[][2]string{{"Order", "A purchase request"}},
			[]string{"Order must have items"},
			nil, nil,
		)
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{canvas})
		assert.Contains(t, md, "# Bounded Context Canvas: Sales")
	})

	t.Run("contains purpose section", func(t *testing.T) {
		t.Parallel()
		c := makeCanvas()
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{c})
		assert.Contains(t, md, "## Purpose")
		assert.Contains(t, md, "Manages orders")
	})

	t.Run("contains strategic classification table", func(t *testing.T) {
		t.Parallel()
		c := makeCanvas()
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{c})
		assert.Contains(t, md, "## Strategic Classification")
		assert.Contains(t, strings.ToLower(md), "core")
		assert.Contains(t, md, "Revenue")
		assert.Contains(t, md, "Custom")
	})

	t.Run("contains domain roles checklist", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
		canvas := domain.NewBoundedContextCanvas("Sales", "Manages orders", sc,
			[]domain.Role{domain.RoleExecution, domain.RoleAnalysis},
			nil, nil, nil, nil, nil, nil)
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{canvas})
		assert.Contains(t, md, "## Domain Roles")
		assert.Contains(t, strings.ToLower(md), "execution")
		assert.Contains(t, strings.ToLower(md), "analysis")
	})

	t.Run("contains communication tables", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
		canvas := domain.NewBoundedContextCanvas("Sales", "Manages orders", sc,
			[]domain.Role{domain.RoleExecution},
			[]domain.CommunicationMessage{domain.NewCommunicationMessage("PlaceOrder", "Command", "API Gateway")},
			[]domain.CommunicationMessage{domain.NewCommunicationMessage("OrderPlaced", "Event", "Fulfillment")},
			nil, nil, nil, nil)
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{canvas})
		assert.Contains(t, md, "## Inbound Communication")
		assert.Contains(t, md, "PlaceOrder")
		assert.Contains(t, md, "## Outbound Communication")
		assert.Contains(t, md, "OrderPlaced")
	})

	t.Run("contains UL table", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
		canvas := domain.NewBoundedContextCanvas("Sales", "Manages orders", sc,
			[]domain.Role{domain.RoleExecution}, nil, nil,
			[][2]string{{"Order", "A purchase request"}}, nil, nil, nil)
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{canvas})
		assert.Contains(t, md, "## Ubiquitous Language")
		assert.Contains(t, md, "Order")
		assert.Contains(t, md, "A purchase request")
	})

	t.Run("contains business decisions", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
		canvas := domain.NewBoundedContextCanvas("Sales", "Manages orders", sc,
			[]domain.Role{domain.RoleExecution}, nil, nil, nil,
			[]string{"Order must have items"}, nil, nil)
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{canvas})
		assert.Contains(t, md, "Business Decisions")
		assert.Contains(t, md, "Order must have items")
	})

	t.Run("special chars in markdown", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainSupporting, "Compliance", "Product")
		canvas := domain.NewBoundedContextCanvas(`Auth & Identity "Service"`, "Handles auth", sc,
			[]domain.Role{domain.RoleSpecification}, nil, nil, nil, nil, nil, nil)
		md := domain.RenderMarkdown([]domain.BoundedContextCanvas{canvas})
		assert.Contains(t, md, `Auth & Identity "Service"`)
	})
}
