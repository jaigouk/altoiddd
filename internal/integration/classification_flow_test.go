// Package integration provides cross-context integration tests.
package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fitnessdomain "github.com/alty-cli/alty/internal/fitness/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// Classification Flow Integration Tests (alty-cli-r3i.4)
//
// Verifies that subdomain classification flows through to:
// 1. Fitness function contract strictness
// 2. Ticket detail levels
// ---------------------------------------------------------------------------

func TestClassificationFlowsToFitnessAndTickets(t *testing.T) {
	t.Parallel()

	t.Run("fitness contracts match classification strictness", func(t *testing.T) {
		t.Parallel()
		model := makeThreeClassificationModel(t)
		suite := fitnessdomain.NewFitnessTestSuite("github.com/example/project")

		// Build BoundedContextInput from model
		var inputs []fitnessdomain.BoundedContextInput
		for _, bc := range model.BoundedContexts() {
			inputs = append(inputs, fitnessdomain.BoundedContextInput{
				Name:           bc.Name(),
				Classification: bc.Classification(),
				Responsibility: bc.Responsibility(),
			})
		}

		err := suite.GenerateContracts(inputs)
		require.NoError(t, err)

		contracts := suite.Contracts()
		require.NotEmpty(t, contracts)

		// Count contracts per context
		contractsByContext := make(map[string][]fitnessdomain.Contract)
		for _, c := range contracts {
			contractsByContext[c.ContextName()] = append(contractsByContext[c.ContextName()], c)
		}

		// Core (Orders) should have 4 contracts: Layers, Forbidden, Independence, AcyclicSiblings
		assert.Len(t, contractsByContext["Orders"], 4,
			"Core context should have 4 contracts (strict)")

		// Supporting (Notifications) should have 2 contracts: Layers, Forbidden
		assert.Len(t, contractsByContext["Notifications"], 2,
			"Supporting context should have 2 contracts (moderate)")

		// Generic (Auth) should have 1 contract: Forbidden
		assert.Len(t, contractsByContext["Auth"], 1,
			"Generic context should have 1 contract (minimal)")
	})

	t.Run("ticket detail levels match classification", func(t *testing.T) {
		t.Parallel()
		model := makeThreeClassificationModel(t)
		beadsWriter := newRecordingBeadsWriter()
		handler := makeHandler(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		err = handler.ApproveAndWriteToBeads(context.Background(), preview, nil)
		require.NoError(t, err)

		// Find tickets by context
		ticketsByContext := make(map[string]ticketdomain.GeneratedTicket)
		for _, ticket := range beadsWriter.tickets {
			// Take first ticket from each context
			if _, exists := ticketsByContext[ticket.BoundedContextName()]; !exists {
				ticketsByContext[ticket.BoundedContextName()] = ticket
			}
		}

		// Core (Orders) should have FULL detail
		assert.Equal(t, vo.TicketDetailFull, ticketsByContext["Orders"].DetailLevel(),
			"Core context tickets should have FULL detail")

		// Supporting (Notifications) should have STANDARD detail
		assert.Equal(t, vo.TicketDetailStandard, ticketsByContext["Notifications"].DetailLevel(),
			"Supporting context tickets should have STANDARD detail")

		// Generic (Auth) should have STUB detail
		assert.Equal(t, vo.TicketDetailStub, ticketsByContext["Auth"].DetailLevel(),
			"Generic context tickets should have STUB detail")
	})

	t.Run("end-to-end classification consistency", func(t *testing.T) {
		t.Parallel()
		model := makeThreeClassificationModel(t)

		// Verify model has all three classifications
		bcs := model.BoundedContexts()
		require.Len(t, bcs, 3)

		classifications := make(map[string]vo.SubdomainClassification)
		for _, bc := range bcs {
			require.NotNil(t, bc.Classification(), "all contexts must be classified")
			classifications[bc.Name()] = *bc.Classification()
		}

		assert.Equal(t, vo.SubdomainCore, classifications["Orders"])
		assert.Equal(t, vo.SubdomainSupporting, classifications["Notifications"])
		assert.Equal(t, vo.SubdomainGeneric, classifications["Auth"])

		// Verify fitness strictness mapping
		for name, cls := range classifications {
			strictness := fitnessdomain.StrictnessFromClassification(cls)
			switch cls {
			case vo.SubdomainCore:
				assert.Equal(t, fitnessdomain.ContractStrictnessStrict, strictness,
					"%s (Core) should map to Strict", name)
			case vo.SubdomainSupporting:
				assert.Equal(t, fitnessdomain.ContractStrictnessModerate, strictness,
					"%s (Supporting) should map to Moderate", name)
			case vo.SubdomainGeneric:
				assert.Equal(t, fitnessdomain.ContractStrictnessMinimal, strictness,
					"%s (Generic) should map to Minimal", name)
			}
		}

		// Verify ticket detail level mapping
		for name, cls := range classifications {
			detailLevel := vo.DetailLevelFromClassification(cls)
			switch cls {
			case vo.SubdomainCore:
				assert.Equal(t, vo.TicketDetailFull, detailLevel,
					"%s (Core) should map to Full", name)
			case vo.SubdomainSupporting:
				assert.Equal(t, vo.TicketDetailStandard, detailLevel,
					"%s (Supporting) should map to Standard", name)
			case vo.SubdomainGeneric:
				assert.Equal(t, vo.TicketDetailStub, detailLevel,
					"%s (Generic) should map to Stub", name)
			}
		}
	})
}

// makeThreeClassificationModel creates a model with Core, Supporting, and Generic contexts.
func makeThreeClassificationModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("three-classification-test")

	// Add domain story
	story := vo.NewDomainStory(
		"E-commerce with Auth and Notifications",
		[]string{"Customer", "System"},
		"Customer places order, system sends notification",
		[]string{"Customer places order", "System authenticates user", "System sends notification"},
		nil,
	)
	model.AddDomainStory(story)

	// Add terms
	model.AddTerm("Order", "A customer order", "Orders", nil)
	model.AddTerm("Notification", "System notification", "Notifications", nil)
	model.AddTerm("Auth", "Authentication", "Auth", nil)

	// Add bounded contexts with different classifications
	ordersBC := vo.NewDomainBoundedContext("Orders", "Core order management", nil, nil, "")
	notificationsBC := vo.NewDomainBoundedContext("Notifications", "Notification delivery", nil, nil, "")
	authBC := vo.NewDomainBoundedContext("Auth", "User authentication", nil, nil, "")

	model.AddBoundedContext(ordersBC)
	model.AddBoundedContext(notificationsBC)
	model.AddBoundedContext(authBC)

	// Classify subdomains
	require.NoError(t, model.ClassifySubdomain("Orders", vo.SubdomainCore, "Complex pricing rules, competitive differentiator"))
	require.NoError(t, model.ClassifySubdomain("Notifications", vo.SubdomainSupporting, "Custom templates but standard delivery"))
	require.NoError(t, model.ClassifySubdomain("Auth", vo.SubdomainGeneric, "Off-the-shelf solution available"))

	// Design aggregates for each context
	orderAgg := vo.NewAggregateDesign("OrderRoot", "Orders", "OrderRoot",
		nil, []string{"order must have items"}, nil, nil)
	notificationAgg := vo.NewAggregateDesign("NotificationRoot", "Notifications", "NotificationRoot",
		nil, []string{"notification must have recipient"}, nil, nil)
	authAgg := vo.NewAggregateDesign("AuthRoot", "Auth", "AuthRoot",
		nil, []string{"user must have credentials"}, nil, nil)

	require.NoError(t, model.DesignAggregate(orderAgg))
	require.NoError(t, model.DesignAggregate(notificationAgg))
	require.NoError(t, model.DesignAggregate(authAgg))

	// Finalize
	model.Finalize()

	return model
}
