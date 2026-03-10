package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/ticket/domain"
)

func TestTicketPlanApproved(t *testing.T) {
	t.Parallel()

	event := domain.NewTicketPlanApproved("plan-1", []string{"t-1", "t-2"}, []string{"t-3"})

	assert.Equal(t, "plan-1", event.PlanID())
	assert.Equal(t, []string{"t-1", "t-2"}, event.ApprovedTicketIDs())
	assert.Equal(t, []string{"t-3"}, event.DismissedTicketIDs())
}

func TestTicketPlanApprovedEmptySlices(t *testing.T) {
	t.Parallel()

	event := domain.NewTicketPlanApproved("plan-1", []string{}, []string{})

	assert.Equal(t, []string{}, event.ApprovedTicketIDs())
	assert.Equal(t, []string{}, event.DismissedTicketIDs())
}

func TestTicketPlanApprovedDefensiveCopy(t *testing.T) {
	t.Parallel()

	approved := []string{"t-1", "t-2"}
	dismissed := []string{"t-3"}
	event := domain.NewTicketPlanApproved("plan-1", approved, dismissed)

	// Mutate originals.
	approved[0] = "mutated"
	dismissed[0] = "mutated"

	assert.Equal(t, "t-1", event.ApprovedTicketIDs()[0])
	assert.Equal(t, "t-3", event.DismissedTicketIDs()[0])
}

func TestTicketPlanApproved_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := domain.NewTicketPlanApproved("plan-rt", []string{"t-1", "t-2"}, []string{"t-3"})

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"plan_id"`)
	assert.Contains(t, string(data), `"approved_ticket_ids"`)

	var restored domain.TicketPlanApproved
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "plan-rt", restored.PlanID())
	assert.Equal(t, original.ApprovedTicketIDs(), restored.ApprovedTicketIDs())
	assert.Equal(t, original.DismissedTicketIDs(), restored.DismissedTicketIDs())
}

// TicketClosed event tests

func TestNewTicketClosed_ValidatesTicketID(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTicketClosed("", "2026-03-09T10:00:00", "Added fitness tests")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ticket ID")
}

func TestNewTicketClosed_AllowsEmptyContextDiff(t *testing.T) {
	t.Parallel()

	event, err := domain.NewTicketClosed("t-123", "2026-03-09T10:00:00", "")
	require.NoError(t, err)
	assert.Empty(t, event.ContextDiff())
}

func TestTicketClosed_Accessors(t *testing.T) {
	t.Parallel()

	event, err := domain.NewTicketClosed("t-123", "2026-03-09T10:00:00", "Added fitness tests")
	require.NoError(t, err)

	assert.Equal(t, "t-123", event.TicketID())
	assert.Equal(t, "2026-03-09T10:00:00", event.ClosedAt())
	assert.Equal(t, "Added fitness tests", event.ContextDiff())
}

func TestTicketClosed_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original, err := domain.NewTicketClosed("t-rt", "2026-03-09T10:00:00", "Added fitness tests")
	require.NoError(t, err)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"ticket_id"`)
	assert.Contains(t, string(data), `"closed_at"`)
	assert.Contains(t, string(data), `"context_diff"`)

	var restored domain.TicketClosed
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "t-rt", restored.TicketID())
	assert.Equal(t, "2026-03-09T10:00:00", restored.ClosedAt())
	assert.Equal(t, "Added fitness tests", restored.ContextDiff())
}
