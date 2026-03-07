package domain_test

import (
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
