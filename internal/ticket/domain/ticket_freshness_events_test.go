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
