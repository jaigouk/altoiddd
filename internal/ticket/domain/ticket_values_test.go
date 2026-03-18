package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/ticket/domain"
)

func TestGeneratedEpic(t *testing.T) {
	t.Parallel()

	core := vo.SubdomainCore
	epic := domain.NewGeneratedEpic("e1", "Orders Epic", "Implement Orders", "Orders", core)

	assert.Equal(t, "e1", epic.EpicID())
	assert.Equal(t, "Orders Epic", epic.Title())
	assert.Equal(t, "Implement Orders", epic.Description())
	assert.Equal(t, "Orders", epic.BoundedContextName())
	assert.Equal(t, core, epic.Classification())
}

func TestGeneratedTicket(t *testing.T) {
	t.Parallel()

	ticket := domain.NewGeneratedTicket(
		"t1", "Implement Order", "desc body",
		vo.TicketDetailFull, "e1", "Orders", "Order",
		[]string{"t0"}, 2,
	)

	assert.Equal(t, "t1", ticket.TicketID())
	assert.Equal(t, "Implement Order", ticket.Title())
	assert.Equal(t, "desc body", ticket.Description())
	assert.Equal(t, vo.TicketDetailFull, ticket.DetailLevel())
	assert.Equal(t, "e1", ticket.EpicID())
	assert.Equal(t, "Orders", ticket.BoundedContextName())
	assert.Equal(t, "Order", ticket.AggregateName())
	assert.Equal(t, []string{"t0"}, ticket.Dependencies())
	assert.Equal(t, 2, ticket.Depth())
}

func TestGeneratedTicketDefensiveCopy(t *testing.T) {
	t.Parallel()

	deps := []string{"t0", "t1"}
	ticket := domain.NewGeneratedTicket(
		"t2", "Title", "desc",
		vo.TicketDetailStub, "e1", "Logging", "Log",
		deps, 0,
	)

	// Mutate original slice — should not affect ticket.
	deps[0] = "mutated"
	assert.Equal(t, "t0", ticket.Dependencies()[0])

	// Mutate returned slice — should not affect ticket.
	returned := ticket.Dependencies()
	returned[0] = "mutated"
	assert.Equal(t, "t0", ticket.Dependencies()[0])
}

func TestGeneratedTicketDefaultDepthZero(t *testing.T) {
	t.Parallel()

	ticket := domain.NewGeneratedTicket(
		"t1", "Title", "desc",
		vo.TicketDetailStub, "e1", "Ctx", "Agg",
		nil, 0,
	)
	assert.Equal(t, 0, ticket.Depth())
}

func TestDependencyOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ids  []string
	}{
		{"single", []string{"t1"}},
		{"multiple", []string{"t1", "t2", "t3"}},
		{"empty", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			order := domain.NewDependencyOrder(tt.ids)
			require.Equal(t, tt.ids, order.OrderedIDs())
		})
	}
}

func TestDependencyOrderDefensiveCopy(t *testing.T) {
	t.Parallel()

	ids := []string{"t1", "t2"}
	order := domain.NewDependencyOrder(ids)

	// Mutate original — should not affect order.
	ids[0] = "mutated"
	assert.Equal(t, "t1", order.OrderedIDs()[0])

	// Mutate returned — should not affect order.
	returned := order.OrderedIDs()
	returned[0] = "mutated"
	assert.Equal(t, "t1", order.OrderedIDs()[0])
}
