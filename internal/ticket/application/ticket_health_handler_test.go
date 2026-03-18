package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/ticket/application"
	"github.com/alto-cli/alto/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// Mock reader
// ---------------------------------------------------------------------------

type mockTicketReader struct {
	openTickets   []domain.OpenTicketData
	flagsByTicket map[string][]domain.FreshnessFlag
}

func (m *mockTicketReader) ReadOpenTickets(_ context.Context) ([]domain.OpenTicketData, error) {
	return m.openTickets, nil
}

func (m *mockTicketReader) ReadFlags(_ context.Context, ticketID string) ([]domain.FreshnessFlag, error) {
	flags, ok := m.flagsByTicket[ticketID]
	if !ok {
		return nil, nil
	}
	return flags, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeDiff(summary string) domain.ContextDiff {
	diff, _ := domain.NewContextDiff(summary, "k7m.19", "2026-03-01")
	return diff
}

func makeFlag(summary string) domain.FreshnessFlag {
	return domain.NewFreshnessFlag(makeDiff(summary), "2026-03-01T10:00:00")
}

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestTicketHealthHandler(t *testing.T) {
	t.Parallel()

	t.Run("reads open tickets and flags", func(t *testing.T) {
		t.Parallel()
		flag := makeFlag("Change")
		reader := &mockTicketReader{
			openTickets: []domain.OpenTicketData{
				domain.NewOpenTicketData("k7m.25", "Ticket Health", []string{"review_needed"}, strPtr("2026-02-28")),
				domain.NewOpenTicketData("k7m.20", "Ticket Gen", []string{}, strPtr("2026-02-25")),
			},
			flagsByTicket: map[string][]domain.FreshnessFlag{
				"k7m.25": {flag},
			},
		}
		handler := application.NewTicketHealthHandler(reader)

		report, err := handler.Report(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 2, report.TotalOpen())
		assert.Equal(t, 1, report.ReviewNeededCount())
		assert.Equal(t, "k7m.25", report.FlaggedTickets()[0].TicketID())
	})

	t.Run("empty report no open tickets", func(t *testing.T) {
		t.Parallel()
		reader := &mockTicketReader{
			openTickets: nil,
		}
		handler := application.NewTicketHealthHandler(reader)

		report, err := handler.Report(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 0, report.TotalOpen())
		assert.Equal(t, 0, report.ReviewNeededCount())
		assert.False(t, report.HasIssues())
	})

	t.Run("includes context diffs", func(t *testing.T) {
		t.Parallel()
		flag := makeFlag("Added new aggregate")
		reader := &mockTicketReader{
			openTickets: []domain.OpenTicketData{
				domain.NewOpenTicketData("k7m.25", "Ticket Health", []string{"review_needed"}, nil),
			},
			flagsByTicket: map[string][]domain.FreshnessFlag{
				"k7m.25": {flag},
			},
		}
		handler := application.NewTicketHealthHandler(reader)

		report, err := handler.Report(context.Background())

		require.NoError(t, err)
		require.Len(t, report.FlaggedTickets(), 1)
		flagged := report.FlaggedTickets()[0]
		assert.Equal(t, "Added new aggregate", flagged.Flags()[0].ContextDiff().Summary())
	})

	t.Run("excludes non-flagged tickets", func(t *testing.T) {
		t.Parallel()
		reader := &mockTicketReader{
			openTickets: []domain.OpenTicketData{
				domain.NewOpenTicketData("k7m.20", "No flags", []string{}, strPtr("2026-02-28")),
				domain.NewOpenTicketData("k7m.21", "Also no flags", []string{"some_other_label"}, strPtr("2026-02-27")),
			},
		}
		handler := application.NewTicketHealthHandler(reader)

		report, err := handler.Report(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 2, report.TotalOpen())
		assert.Equal(t, 0, report.ReviewNeededCount())
		assert.False(t, report.HasIssues())
	})

	t.Run("finds oldest last reviewed", func(t *testing.T) {
		t.Parallel()
		reader := &mockTicketReader{
			openTickets: []domain.OpenTicketData{
				domain.NewOpenTicketData("k7m.20", "Old", []string{}, strPtr("2026-01-15")),
				domain.NewOpenTicketData("k7m.21", "Newer", []string{}, strPtr("2026-02-28")),
				domain.NewOpenTicketData("k7m.22", "Never reviewed", []string{}, nil),
			},
		}
		handler := application.NewTicketHealthHandler(reader)

		report, err := handler.Report(context.Background())

		require.NoError(t, err)
		require.NotNil(t, report.OldestLastReviewed())
		assert.Equal(t, "2026-01-15", *report.OldestLastReviewed())
	})

	t.Run("oldest last reviewed nil when all nil", func(t *testing.T) {
		t.Parallel()
		reader := &mockTicketReader{
			openTickets: []domain.OpenTicketData{
				domain.NewOpenTicketData("k7m.20", "Never reviewed", []string{}, nil),
			},
		}
		handler := application.NewTicketHealthHandler(reader)

		report, err := handler.Report(context.Background())

		require.NoError(t, err)
		assert.Nil(t, report.OldestLastReviewed())
	})

	t.Run("flagged ticket status is review needed", func(t *testing.T) {
		t.Parallel()
		flag := makeFlag("Change")
		reader := &mockTicketReader{
			openTickets: []domain.OpenTicketData{
				domain.NewOpenTicketData("k7m.25", "Ticket Health", []string{"review_needed"}, nil),
			},
			flagsByTicket: map[string][]domain.FreshnessFlag{
				"k7m.25": {flag},
			},
		}
		handler := application.NewTicketHealthHandler(reader)

		report, err := handler.Report(context.Background())

		require.NoError(t, err)
		assert.Equal(t, domain.FreshnessStatusReviewNeeded, report.FlaggedTickets()[0].Status())
	})
}
