package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/ticket/domain"
)

func TestTicketFlagged(t *testing.T) {
	t.Parallel()

	diff, err := domain.NewContextDiff("New fitness tests", "k7m.19", "2026-03-01")
	require.NoError(t, err)

	event := domain.NewTicketFlagged("rr-001", "k7m.25", diff, "2026-03-01T10:00:00")

	assert.Equal(t, "rr-001", event.ReviewID())
	assert.Equal(t, "k7m.25", event.TicketID())
	assert.Equal(t, diff, event.ContextDiff())
	assert.Equal(t, "2026-03-01T10:00:00", event.FlaggedAt())
}

func TestFlagCleared(t *testing.T) {
	t.Parallel()

	event := domain.NewFlagCleared("rr-001", "k7m.25", "2026-03-01T11:00:00")

	assert.Equal(t, "rr-001", event.ReviewID())
	assert.Equal(t, "k7m.25", event.TicketID())
	assert.Equal(t, "2026-03-01T11:00:00", event.ClearedAt())
}

func TestTicketFlagged_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	diff, err := domain.NewContextDiff("New fitness tests", "k7m.19", "2026-03-01")
	require.NoError(t, err)

	original := domain.NewTicketFlagged("rr-rt", "k7m.25", diff, "2026-03-01T10:00:00")

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"review_id"`)
	assert.Contains(t, string(data), `"flagged_at"`)

	var restored domain.TicketFlagged
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "rr-rt", restored.ReviewID())
	assert.Equal(t, "k7m.25", restored.TicketID())
	assert.Equal(t, "2026-03-01T10:00:00", restored.FlaggedAt())
	assert.Equal(t, "New fitness tests", restored.ContextDiff().Summary())
	assert.Equal(t, "k7m.19", restored.ContextDiff().TriggeringTicketID())
	assert.Equal(t, "2026-03-01", restored.ContextDiff().ProducedAt())
}

func TestFlagCleared_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := domain.NewFlagCleared("rr-rt", "k7m.25", "2026-03-01T11:00:00")

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"review_id"`)
	assert.Contains(t, string(data), `"cleared_at"`)

	var restored domain.FlagCleared
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "rr-rt", restored.ReviewID())
	assert.Equal(t, "k7m.25", restored.TicketID())
	assert.Equal(t, "2026-03-01T11:00:00", restored.ClearedAt())
}

// RippleReviewCreated event tests

func TestNewRippleReviewCreated_ValidatesReviewID(t *testing.T) {
	t.Parallel()

	_, err := domain.NewRippleReviewCreated("", "t-123", 3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "review ID")
}

func TestNewRippleReviewCreated_ValidatesClosedTicketID(t *testing.T) {
	t.Parallel()

	_, err := domain.NewRippleReviewCreated("rr-001", "", 3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed ticket ID")
}

func TestRippleReviewCreated_AllowsZeroFlaggedCount(t *testing.T) {
	t.Parallel()

	event, err := domain.NewRippleReviewCreated("rr-001", "t-123", 0)
	require.NoError(t, err)
	assert.Equal(t, 0, event.FlaggedCount())
}

func TestRippleReviewCreated_Accessors(t *testing.T) {
	t.Parallel()

	event, err := domain.NewRippleReviewCreated("rr-001", "t-123", 5)
	require.NoError(t, err)

	assert.Equal(t, "rr-001", event.ReviewID())
	assert.Equal(t, "t-123", event.ClosedTicketID())
	assert.Equal(t, 5, event.FlaggedCount())
}

func TestRippleReviewCreated_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original, err := domain.NewRippleReviewCreated("rr-rt", "t-123", 7)
	require.NoError(t, err)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"review_id"`)
	assert.Contains(t, string(data), `"closed_ticket_id"`)
	assert.Contains(t, string(data), `"flagged_count"`)

	var restored domain.RippleReviewCreated
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "rr-rt", restored.ReviewID())
	assert.Equal(t, "t-123", restored.ClosedTicketID())
	assert.Equal(t, 7, restored.FlaggedCount())
}
