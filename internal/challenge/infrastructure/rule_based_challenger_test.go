package infrastructure_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	challengeapp "github.com/alty-cli/alty/internal/challenge/application"
	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/challenge/infrastructure"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

func makeModelWithGaps(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("test-model")
	ctx := vo.NewDomainBoundedContext("Sales", "Orders", nil, nil, "")
	require.NoError(t, model.AddBoundedContext(ctx))
	core := vo.SubdomainCore
	require.NoError(t, model.ClassifySubdomain("Sales", core, ""))
	agg := vo.NewAggregateDesign("OrderAggregate", "Sales", "Order", nil, nil, nil, nil)
	require.NoError(t, model.DesignAggregate(agg))
	story := vo.NewDomainStory("Place Order", []string{"Customer"}, "Customer submits",
		[]string{"System creates order"}, nil)
	require.NoError(t, model.AddDomainStory(story))
	require.NoError(t, model.AddTerm("Order", "A purchase", "Sales", nil))
	return model
}

// ---------------------------------------------------------------------------
// Protocol compliance
// ---------------------------------------------------------------------------

func TestRuleBasedChallengerSatisfiesChallengerPort(t *testing.T) {
	t.Parallel()
	var _ challengeapp.Challenger = (*infrastructure.RuleBasedChallengerAdapter)(nil)
}

// ---------------------------------------------------------------------------
// Delegation
// ---------------------------------------------------------------------------

func TestRuleBasedChallengerGeneratesChallenges(t *testing.T) {
	t.Parallel()
	adapter := &infrastructure.RuleBasedChallengerAdapter{}
	model := makeModelWithGaps(t)
	challenges, err := adapter.GenerateChallenges(context.Background(), model, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(challenges), 1)
}

func TestRuleBasedChallengerRespectsMaxPerType(t *testing.T) {
	t.Parallel()
	adapter := &infrastructure.RuleBasedChallengerAdapter{}
	model := makeModelWithGaps(t)
	challenges, err := adapter.GenerateChallenges(context.Background(), model, 1)
	require.NoError(t, err)

	byType := make(map[challengedomain.ChallengeType]int)
	for _, c := range challenges {
		byType[c.ChallengeType()]++
	}
	for ct, count := range byType {
		assert.LessOrEqual(t, count, 1, "type %s has %d challenges (max 1)", ct, count)
	}
}

func TestRuleBasedChallengerEmptyModelReturnsNoChallenges(t *testing.T) {
	t.Parallel()
	adapter := &infrastructure.RuleBasedChallengerAdapter{}
	model := ddd.NewDomainModel("empty")
	challenges, err := adapter.GenerateChallenges(context.Background(), model, 5)
	require.NoError(t, err)
	assert.Empty(t, challenges)
}
