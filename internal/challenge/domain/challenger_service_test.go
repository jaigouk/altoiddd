package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

func makeRichModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	m := ddd.NewDomainModel("test-rich")

	// Two contexts — enables boundary challenges
	err := m.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, ""))
	require.NoError(t, err)
	core := vo.SubdomainCore
	err = m.ClassifySubdomain("Sales", core, "")
	require.NoError(t, err)

	err = m.AddBoundedContext(vo.NewDomainBoundedContext("Shipping", "Deliveries", nil, nil, ""))
	require.NoError(t, err)
	supporting := vo.SubdomainSupporting
	err = m.ClassifySubdomain("Shipping", supporting, "")
	require.NoError(t, err)

	// Aggregate with no invariants — triggers invariant challenge
	err = m.DesignAggregate(vo.NewAggregateDesign(
		"OrderAggregate", "Sales", "Order", nil, nil, nil, nil,
	))
	require.NoError(t, err)

	// Domain story — triggers failure mode challenges
	err = m.AddDomainStory(vo.NewDomainStory(
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
	require.NoError(t, err)

	// Terms — "Order" only in Sales, "Shipment" only in Shipping
	err = m.AddTerm("Order", "A customer purchase", "Sales", nil)
	require.NoError(t, err)
	err = m.AddTerm("Shipment", "A delivery package", "Shipping", nil)
	require.NoError(t, err)

	// Relationship
	err = m.AddContextRelationship(vo.NewContextRelationship("Sales", "Shipping", "Domain Events"))
	require.NoError(t, err)

	return m
}

func makeEmptyModel() *ddd.DomainModel {
	return ddd.NewDomainModel("test-empty")
}

func makeSingleContextModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	m := ddd.NewDomainModel("test-single")

	err := m.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, ""))
	require.NoError(t, err)
	core := vo.SubdomainCore
	err = m.ClassifySubdomain("Sales", core, "")
	require.NoError(t, err)

	err = m.DesignAggregate(vo.NewAggregateDesign(
		"OrderAggregate", "Sales", "Order", nil,
		[]string{"Total must be positive"}, nil, nil,
	))
	require.NoError(t, err)

	err = m.AddDomainStory(vo.NewDomainStory(
		"Place Order",
		[]string{"Customer"},
		"Customer submits order",
		[]string{"System creates order"},
		nil,
	))
	require.NoError(t, err)

	err = m.AddTerm("Order", "A purchase", "Sales", nil)
	require.NoError(t, err)

	return m
}

func makeGenericOnlyModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	m := ddd.NewDomainModel("test-generic")

	err := m.AddBoundedContext(vo.NewDomainBoundedContext("Auth", "Authentication", nil, nil, ""))
	require.NoError(t, err)
	generic := vo.SubdomainGeneric
	err = m.ClassifySubdomain("Auth", generic, "")
	require.NoError(t, err)

	err = m.AddDomainStory(vo.NewDomainStory(
		"Login",
		[]string{"User"},
		"User enters credentials",
		[]string{"System verifies credentials"},
		nil,
	))
	require.NoError(t, err)

	err = m.AddTerm("User", "An authenticated person", "Auth", nil)
	require.NoError(t, err)

	return m
}

func TestChallengerServiceGeneration(t *testing.T) {
	t.Parallel()

	t.Run("generates language challenges for ambiguous terms", func(t *testing.T) {
		t.Parallel()
		m := ddd.NewDomainModel("test-lang")

		err := m.AddBoundedContext(vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, ""))
		require.NoError(t, err)
		core := vo.SubdomainCore
		err = m.ClassifySubdomain("Sales", core, "")
		require.NoError(t, err)

		err = m.AddBoundedContext(vo.NewDomainBoundedContext("Shipping", "Deliveries", nil, nil, ""))
		require.NoError(t, err)
		supporting := vo.SubdomainSupporting
		err = m.ClassifySubdomain("Shipping", supporting, "")
		require.NoError(t, err)

		err = m.DesignAggregate(vo.NewAggregateDesign(
			"OrderAggregate", "Sales", "Order", nil, nil, nil, nil,
		))
		require.NoError(t, err)

		err = m.AddDomainStory(vo.NewDomainStory(
			"Ship Order",
			[]string{"Warehouse"},
			"Order confirmed",
			[]string{"Create shipment for order"},
			nil,
		))
		require.NoError(t, err)

		// Same term in two contexts — ambiguous
		err = m.AddTerm("Order", "A purchase", "Sales", nil)
		require.NoError(t, err)
		err = m.AddTerm("Order", "A shipping request", "Shipping", nil)
		require.NoError(t, err)

		challenges := domain.Generate(m, 5)
		language := filterByType(challenges, domain.ChallengeLanguage)
		assert.GreaterOrEqual(t, len(language), 1)
		assert.True(t, anyContains(language, "order"))
	})

	t.Run("generates invariant challenges for empty aggregates", func(t *testing.T) {
		t.Parallel()
		m := makeRichModel(t)
		challenges := domain.Generate(m, 5)
		invariant := filterByType(challenges, domain.ChallengeInvariant)
		assert.GreaterOrEqual(t, len(invariant), 1)
		assert.True(t, anyContains(invariant, "order"))
	})

	t.Run("generates failure mode challenges for core stories", func(t *testing.T) {
		t.Parallel()
		m := makeRichModel(t)
		challenges := domain.Generate(m, 5)
		failure := filterByType(challenges, domain.ChallengeFailureMode)
		assert.GreaterOrEqual(t, len(failure), 1)
	})

	t.Run("generates boundary challenges for multiple contexts", func(t *testing.T) {
		t.Parallel()
		m := makeRichModel(t)
		challenges := domain.Generate(m, 5)
		boundary := filterByType(challenges, domain.ChallengeBoundary)
		assert.GreaterOrEqual(t, len(boundary), 1)
	})

	t.Run("single context no boundary challenges", func(t *testing.T) {
		t.Parallel()
		m := makeSingleContextModel(t)
		challenges := domain.Generate(m, 5)
		boundary := filterByType(challenges, domain.ChallengeBoundary)
		assert.Empty(t, boundary)
	})

	t.Run("skips generic subdomains for invariant challenges", func(t *testing.T) {
		t.Parallel()
		m := makeGenericOnlyModel(t)
		challenges := domain.Generate(m, 5)
		invariant := filterByType(challenges, domain.ChallengeInvariant)
		assert.Empty(t, invariant)
	})

	t.Run("every challenge has source reference", func(t *testing.T) {
		t.Parallel()
		m := makeRichModel(t)
		challenges := domain.Generate(m, 5)
		require.NotEmpty(t, challenges)
		for _, c := range challenges {
			assert.NotEmpty(t, c.SourceReference(), "Challenge missing source_reference")
		}
	})

	t.Run("max challenges per type respected", func(t *testing.T) {
		t.Parallel()
		m := makeRichModel(t)
		challenges := domain.Generate(m, 2)
		byType := make(map[domain.ChallengeType]int)
		for _, c := range challenges {
			byType[c.ChallengeType()]++
		}
		for ct, count := range byType {
			assert.LessOrEqual(t, count, 2, "%s has %d challenges (max 2)", ct, count)
		}
	})

	t.Run("empty model returns no challenges", func(t *testing.T) {
		t.Parallel()
		m := makeEmptyModel()
		challenges := domain.Generate(m, 5)
		assert.Empty(t, challenges)
	})

	t.Run("challenge context name matches bounded context", func(t *testing.T) {
		t.Parallel()
		m := makeRichModel(t)
		challenges := domain.Generate(m, 5)
		contextNames := make(map[string]struct{})
		for _, ctx := range m.BoundedContexts() {
			contextNames[ctx.Name()] = struct{}{}
		}
		for _, c := range challenges {
			_, ok := contextNames[c.ContextName()]
			assert.True(t, ok, "Challenge references unknown context: %s", c.ContextName())
		}
	})
}

// helpers

func filterByType(challenges []domain.Challenge, ct domain.ChallengeType) []domain.Challenge {
	var result []domain.Challenge
	for _, c := range challenges {
		if c.ChallengeType() == ct {
			result = append(result, c)
		}
	}
	return result
}

func anyContains(challenges []domain.Challenge, substr string) bool {
	for _, c := range challenges {
		lower := strings.ToLower(c.QuestionText())
		if strings.Contains(lower, substr) {
			return true
		}
	}
	return false
}
