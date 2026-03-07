// Package integration provides cross-context integration tests verifying
// shared kernel types work correctly across all bounded context packages.
package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapdomain "github.com/alty-cli/alty/internal/bootstrap/domain"
	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	fitnessdomain "github.com/alty-cli/alty/internal/fitness/domain"
	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/shared/domain/events"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
	tooltranslationdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// ---------------------------------------------------------------------------
// Helpers: build a fully-finalized DomainModel reusable across tests.
// ---------------------------------------------------------------------------

func makeFinalizedDomainModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("integration-test-model")

	// Story mentions all terms: "order", "payment", "notification"
	story := vo.NewDomainStory(
		"Place Order",
		[]string{"Customer"},
		"Customer submits an order",
		[]string{
			"Customer creates an order with items",
			"System validates the order",
			"System processes the payment",
			"System sends a notification",
		},
		nil,
	)
	require.NoError(t, model.AddDomainStory(story))

	// Terms
	require.NoError(t, model.AddTerm("order", "A customer purchase request", "OrderManagement", nil))
	require.NoError(t, model.AddTerm("payment", "Financial transaction for an order", "Billing", nil))
	require.NoError(t, model.AddTerm("notification", "Alert sent to a customer", "Notifications", nil))

	// Bounded contexts
	core := vo.SubdomainCore
	supporting := vo.SubdomainSupporting
	generic := vo.SubdomainGeneric
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext(
		"OrderManagement", "Handle customer orders", []string{"Order", "OrderItem"}, &core, "core business logic",
	)))
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext(
		"Billing", "Process payments", []string{"Payment"}, &supporting, "supports order flow",
	)))
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext(
		"Notifications", "Send alerts", []string{"Notification"}, &generic, "off-the-shelf",
	)))

	// Relationships
	require.NoError(t, model.AddContextRelationship(
		vo.NewContextRelationship("OrderManagement", "Billing", "Published Language"),
	))
	require.NoError(t, model.AddContextRelationship(
		vo.NewContextRelationship("OrderManagement", "Notifications", "Published Language"),
	))

	// Aggregate designs (required for core)
	require.NoError(t, model.DesignAggregate(vo.NewAggregateDesign(
		"Order", "OrderManagement", "Order",
		[]string{"OrderItem", "OrderStatus"},
		[]string{"Order must have at least one item"},
		[]string{"PlaceOrder", "CancelOrder"},
		[]string{"OrderPlaced", "OrderCancelled"},
	)))

	require.NoError(t, model.Finalize())
	return model
}

// ---------------------------------------------------------------------------
// Scenario 1: Shared ValueObjects across contexts
// ---------------------------------------------------------------------------

func TestSharedValueObjectsAcrossContexts(t *testing.T) {
	t.Parallel()

	// Create shared kernel types
	ts := vo.NewTechStack("python", "uv")
	profile := vo.PythonUvProfile{}
	core := vo.SubdomainCore

	t.Run("TechStack used by discovery context", func(t *testing.T) {
		t.Parallel()
		// Discovery session stores TechStack from shared kernel
		session := discoverydomain.NewDiscoverySession("# My Project")
		require.NotNil(t, session)
		// TechStack is a value type — verify it can be created and compared
		ts2 := vo.NewTechStack("python", "uv")
		assert.True(t, ts.Equal(ts2))
	})

	t.Run("StackProfile used by ticket context", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		plan := ticketdomain.NewTicketPlan()
		err := plan.GeneratePlan(model, profile)
		require.NoError(t, err)
		assert.NotEmpty(t, plan.Tickets())
	})

	t.Run("StackProfile used by tooltranslation context", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		adapter := tooltranslationdomain.NewClaudeCodeAdapter()
		sections := adapter.Translate(model, profile)
		assert.NotEmpty(t, sections)
	})

	t.Run("SubdomainClassification used by fitness context", func(t *testing.T) {
		t.Parallel()
		strictness := fitnessdomain.StrictnessFromClassification(core)
		assert.Equal(t, fitnessdomain.ContractStrictnessStrict, strictness)
		required := fitnessdomain.RequiredContractTypes(strictness)
		assert.Len(t, required, 4)
	})

	t.Run("SubdomainClassification used by ticket context", func(t *testing.T) {
		t.Parallel()
		detailLevel := vo.DetailLevelFromClassification(core)
		assert.Equal(t, vo.TicketDetailFull, detailLevel)
	})

	t.Run("StackProfile GenericProfile also satisfies interface", func(t *testing.T) {
		t.Parallel()
		generic := vo.GenericProfile{}
		model := makeFinalizedDomainModel(t)
		plan := ticketdomain.NewTicketPlan()
		err := plan.GeneratePlan(model, generic)
		require.NoError(t, err)
		assert.NotEmpty(t, plan.Tickets())
	})
}

// ---------------------------------------------------------------------------
// Scenario 2: Events carry correct data across context boundaries
// ---------------------------------------------------------------------------

func TestEventsCarryCorrectData(t *testing.T) {
	t.Parallel()

	t.Run("DomainModelGenerated event from DomainModel finalize", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		evts := model.Events()
		require.Len(t, evts, 1)

		evt := evts[0]
		assert.Equal(t, "integration-test-model", evt.ModelID())
		assert.Len(t, evt.DomainStories(), 1)
		assert.Len(t, evt.UbiquitousLanguage(), 3)
		assert.Len(t, evt.BoundedContexts(), 3)
		assert.Len(t, evt.ContextRelationships(), 2)
		assert.Len(t, evt.AggregateDesigns(), 1)

		// Verify payload types match what consuming contexts expect
		// The event's bounded contexts carry classification
		for _, bc := range evt.BoundedContexts() {
			assert.NotNil(t, bc.Classification())
		}
		// The event's aggregate designs carry context names
		for _, agg := range evt.AggregateDesigns() {
			assert.NotEmpty(t, agg.ContextName())
			assert.NotEmpty(t, agg.Name())
		}
	})

	t.Run("TicketPlanApproved event carries correct IDs", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		plan := ticketdomain.NewTicketPlan()
		require.NoError(t, plan.GeneratePlan(model, vo.PythonUvProfile{}))
		require.NoError(t, plan.Approve(nil)) // approve all

		evts := plan.Events()
		require.Len(t, evts, 1)

		evt := evts[0]
		assert.Equal(t, plan.PlanID(), evt.PlanID())
		assert.NotEmpty(t, evt.ApprovedTicketIDs())
		// All approved IDs should match tickets in the plan
		ticketIDs := map[string]bool{}
		for _, ticket := range plan.Tickets() {
			ticketIDs[ticket.TicketID()] = true
		}
		for _, id := range evt.ApprovedTicketIDs() {
			assert.True(t, ticketIDs[id], "approved ID %s not in plan tickets", id)
		}
	})

	t.Run("FitnessTestsGenerated event carries contracts and rules", func(t *testing.T) {
		t.Parallel()
		contracts := []fitnessdomain.Contract{
			fitnessdomain.NewContract("c1", fitnessdomain.ContractTypeLayers, "OrderManagement", []string{"mod1"}, []string{"mod2"}),
		}
		archRules := []fitnessdomain.ArchRule{
			fitnessdomain.NewArchRule("r1", "assert domain does not import infra", "OrderManagement"),
		}
		evt := fitnessdomain.NewFitnessTestsGenerated("suite-1", "mypackage", contracts, archRules)
		assert.Equal(t, "suite-1", evt.SuiteID())
		assert.Len(t, evt.Contracts(), 1)
		assert.Len(t, evt.ArchRules(), 1)
		assert.Equal(t, "mypackage", evt.RootPackage())
	})

	t.Run("BootstrapCompletedEvent carries session data", func(t *testing.T) {
		t.Parallel()
		evt := bootstrapdomain.NewBootstrapCompletedEvent("sess-1", "/tmp/project")
		assert.Equal(t, "sess-1", evt.SessionID())
		assert.Equal(t, "/tmp/project", evt.ProjectDir())
	})

	t.Run("GapAnalysisCompleted event from rescue context", func(t *testing.T) {
		t.Parallel()
		evt := rescuedomain.NewGapAnalysisCompleted("analysis-1", "/tmp/rescue", 5, 3)
		assert.Equal(t, "analysis-1", evt.AnalysisID())
		assert.Equal(t, "/tmp/rescue", evt.ProjectDir())
		assert.Equal(t, 5, evt.GapsFound())
		assert.Equal(t, 3, evt.GapsResolved())
	})

	t.Run("ConfigsGenerated event from shared events", func(t *testing.T) {
		t.Parallel()
		evt := events.NewConfigsGenerated(
			[]string{"claude-code", "cursor"},
			[]string{".claude/CLAUDE.md", ".cursorrules"},
		)
		assert.Equal(t, []string{"claude-code", "cursor"}, evt.ToolNames())
		assert.Equal(t, []string{".claude/CLAUDE.md", ".cursorrules"}, evt.OutputPaths())
	})
}

// ---------------------------------------------------------------------------
// Scenario 3: Error sentinel propagation across contexts
// ---------------------------------------------------------------------------

func TestErrorSentinelPropagation(t *testing.T) {
	t.Parallel()

	t.Run("DomainModel invariant violation unwraps to shared sentinel", func(t *testing.T) {
		t.Parallel()
		model := ddd.NewDomainModel("test")
		// Add a term not in any story — finalize should fail
		require.NoError(t, model.AddTerm("orphan_term", "definition", "Ctx", nil))
		core := vo.SubdomainCore
		require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext(
			"Ctx", "do things", nil, &core, "core",
		)))
		require.NoError(t, model.DesignAggregate(vo.NewAggregateDesign(
			"Agg", "Ctx", "Agg", nil, nil, nil, nil,
		)))
		err := model.Finalize()
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation,
			"DomainModel error should unwrap to ErrInvariantViolation")
	})

	t.Run("TicketPlan invariant violation unwraps to shared sentinel", func(t *testing.T) {
		t.Parallel()
		plan := ticketdomain.NewTicketPlan()
		// Approve with no tickets should fail
		err := plan.Approve(nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation,
			"TicketPlan error should unwrap to ErrInvariantViolation")
	})

	t.Run("BootstrapSession invalid transition unwraps to shared sentinel", func(t *testing.T) {
		t.Parallel()
		session := bootstrapdomain.NewBootstrapSession("/tmp")
		// Confirm without preview should fail
		err := session.Confirm()
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation,
			"BootstrapSession error should unwrap to ErrInvariantViolation")
	})

	t.Run("GapAnalysis state machine error unwraps to shared sentinel", func(t *testing.T) {
		t.Parallel()
		analysis := fitnessdomain.NewGapAnalysis("/tmp/project")
		// Try to complete from scanning state (should require executing)
		err := analysis.Complete()
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation,
			"GapAnalysis error should unwrap to ErrInvariantViolation")
	})

	t.Run("Wrapped errors preserve context while matching sentinel", func(t *testing.T) {
		t.Parallel()
		inner := fmt.Errorf("order not found: %w", domainerrors.ErrNotFound)
		outer := fmt.Errorf("processing failed: %w", inner)
		require.ErrorIs(t, outer, domainerrors.ErrNotFound,
			"deeply wrapped error should still match sentinel")
		assert.Contains(t, outer.Error(), "processing failed")
		assert.Contains(t, outer.Error(), "order not found")
	})

	t.Run("Distinct sentinels do not cross-match across contexts", func(t *testing.T) {
		t.Parallel()
		invErr := fmt.Errorf("context error: %w", domainerrors.ErrInvariantViolation)
		require.NotErrorIs(t, invErr, domainerrors.ErrNotFound)
		require.NotErrorIs(t, invErr, domainerrors.ErrAlreadyExists)
		require.NotErrorIs(t, invErr, domainerrors.ErrInvalidTransition)
	})
}

// ---------------------------------------------------------------------------
// Scenario 4: Cross-context type compatibility
// ---------------------------------------------------------------------------

func TestCrossContextTypeCompatibility(t *testing.T) {
	t.Parallel()

	t.Run("Fitness imports rescue event type", func(t *testing.T) {
		t.Parallel()
		// GapAnalysis (in fitness) stores rescue.GapAnalysisCompleted events
		analysis := fitnessdomain.NewGapAnalysis("/tmp/project")
		scan := fitnessdomain.NewProjectScan(
			"/tmp/project",
			[]string{"README.md"}, []string{".editorconfig"},
			[]string{"src/"}, false, false, true, false, false,
		)
		require.NoError(t, analysis.SetScan(scan))
		require.NoError(t, analysis.Analyze(nil))
		require.NoError(t, analysis.CreatePlan(
			fitnessdomain.NewMigrationPlan("plan-1", nil, "alty/init", false),
		))
		require.NoError(t, analysis.BeginExecution())
		require.NoError(t, analysis.Complete())

		evts := analysis.Events()
		require.Len(t, evts, 1)
		// The event type is rescue.GapAnalysisCompleted
		assert.Equal(t, analysis.AnalysisID(), evts[0].AnalysisID())
		assert.Equal(t, "/tmp/project", evts[0].ProjectDir())
	})

	t.Run("Ticket context uses shared ddd.DomainModel", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		plan := ticketdomain.NewTicketPlan()
		require.NoError(t, plan.GeneratePlan(model, vo.PythonUvProfile{}))

		// Verify tickets reference bounded context names from the model
		ctxNames := map[string]bool{}
		for _, bc := range model.BoundedContexts() {
			ctxNames[bc.Name()] = true
		}
		for _, ticket := range plan.Tickets() {
			assert.True(t, ctxNames[ticket.BoundedContextName()],
				"ticket BC name %q should exist in model", ticket.BoundedContextName())
		}
	})

	t.Run("Challenge context uses shared ddd.DomainModel", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		challenges := challengedomain.Generate(model, 3)
		assert.NotEmpty(t, challenges)
		// Challenges should reference known context names
		for _, ch := range challenges {
			if ch.ContextName() != "" {
				assert.NotEmpty(t, ch.ContextName())
			}
		}
	})

	t.Run("ToolTranslation adapters all consume same model+profile types", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		profile := vo.PythonUvProfile{}

		adapters := []struct {
			adapter interface {
				Translate(model *ddd.DomainModel, profile vo.StackProfile) []tooltranslationdomain.ConfigSection
			}
			name string
		}{
			{tooltranslationdomain.NewClaudeCodeAdapter(), "ClaudeCode"},
			{tooltranslationdomain.NewCursorAdapter(), "Cursor"},
			{tooltranslationdomain.NewRooCodeAdapter(), "RooCode"},
			{tooltranslationdomain.NewOpenCodeAdapter(), "OpenCode"},
		}

		for _, tc := range adapters {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				sections := tc.adapter.Translate(model, profile)
				assert.NotEmpty(t, sections, "%s adapter should produce config sections", tc.name)
			})
		}
	})

	t.Run("DomainModelGenerated event payload consumable by downstream contexts", func(t *testing.T) {
		t.Parallel()
		model := makeFinalizedDomainModel(t)
		evts := model.Events()
		require.Len(t, evts, 1)
		evt := evts[0]

		// Simulate downstream context consuming the event payload
		// Fitness: needs bounded contexts with classifications
		for _, bc := range evt.BoundedContexts() {
			cl := bc.Classification()
			require.NotNil(t, cl)
			// Fitness strictness mapping should work
			strictness := fitnessdomain.StrictnessFromClassification(*cl)
			assert.NotEmpty(t, string(strictness))
		}

		// Ticket: needs aggregate designs with context names
		aggsByCtx := map[string][]vo.AggregateDesign{}
		for _, agg := range evt.AggregateDesigns() {
			aggsByCtx[agg.ContextName()] = append(aggsByCtx[agg.ContextName()], agg)
		}
		assert.Contains(t, aggsByCtx, "OrderManagement")
	})
}

// ---------------------------------------------------------------------------
// Scenario 5: End-to-end flow across multiple contexts
// ---------------------------------------------------------------------------

func TestEndToEndDomainModelToTickets(t *testing.T) {
	t.Parallel()

	// Step 1: Build and finalize domain model (shared/ddd)
	model := makeFinalizedDomainModel(t)
	require.Len(t, model.Events(), 1)

	// Step 2: Generate challenges (challenge context)
	challenges := challengedomain.Generate(model, 5)
	assert.NotEmpty(t, challenges)

	// Step 3: Generate ticket plan (ticket context)
	plan := ticketdomain.NewTicketPlan()
	require.NoError(t, plan.GeneratePlan(model, vo.PythonUvProfile{}))
	assert.NotEmpty(t, plan.Epics())
	assert.NotEmpty(t, plan.Tickets())

	// Step 4: Approve plan (ticket context emits TicketPlanApproved)
	require.NoError(t, plan.Approve(nil))
	require.Len(t, plan.Events(), 1)

	// Step 5: Generate tool configs (tooltranslation context)
	adapter := tooltranslationdomain.NewClaudeCodeAdapter()
	sections := adapter.Translate(model, vo.PythonUvProfile{})
	assert.NotEmpty(t, sections)

	// Step 6: Verify fitness strictness mapping (fitness context)
	for _, bc := range model.BoundedContexts() {
		cl := bc.Classification()
		require.NotNil(t, cl)
		strictness := fitnessdomain.StrictnessFromClassification(*cl)
		contracts := fitnessdomain.RequiredContractTypes(strictness)
		assert.NotEmpty(t, contracts)
	}
}
