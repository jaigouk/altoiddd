package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	"github.com/alto-cli/alto/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// 1. ContextDiff
// ---------------------------------------------------------------------------

func TestContextDiffValid(t *testing.T) {
	t.Parallel()
	diff, err := domain.NewContextDiff("Added order validation", "k7m.19", "2026-03-01")
	require.NoError(t, err)
	assert.Equal(t, "Added order validation", diff.Summary())
	assert.Equal(t, "k7m.19", diff.TriggeringTicketID())
	assert.Equal(t, "2026-03-01", diff.ProducedAt())
}

func TestContextDiffRejectsEmptySummary(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextDiff("", "k7m.19", "2026-03-01")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestContextDiffRejectsWhitespaceSummary(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextDiff("   \t\n  ", "k7m.19", "2026-03-01")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestContextDiffAcceptsLongSummary(t *testing.T) {
	t.Parallel()
	long := make([]byte, 10_000)
	for i := range long {
		long[i] = 'A'
	}
	diff, err := domain.NewContextDiff(string(long), "k7m.19", "2026-03-01")
	require.NoError(t, err)
	assert.Len(t, diff.Summary(), 10_000)
}

func TestContextDiffAcceptsSpecialChars(t *testing.T) {
	t.Parallel()
	diff, err := domain.NewContextDiff("Added module with\nnewlines & 'quotes' + unicode: \u2713", "k7m.19", "2026-03-01")
	require.NoError(t, err)
	assert.Contains(t, diff.Summary(), "\n")
}

// ---------------------------------------------------------------------------
// 2. FreshnessFlag
// ---------------------------------------------------------------------------

func TestFreshnessFlagStoresContextDiff(t *testing.T) {
	t.Parallel()
	diff, err := domain.NewContextDiff("Implemented fitness tests", "k7m.19", "2026-03-01")
	require.NoError(t, err)
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T10:00:00")
	assert.Equal(t, diff, flag.ContextDiff())
	assert.Equal(t, "2026-03-01T10:00:00", flag.FlaggedAt())
}

// ---------------------------------------------------------------------------
// 3. TicketFreshnessStatus
// ---------------------------------------------------------------------------

func TestTicketFreshnessStatusValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "fresh", string(domain.FreshnessStatusFresh))
	assert.Equal(t, "review_needed", string(domain.FreshnessStatusReviewNeeded))
	assert.Equal(t, "never_reviewed", string(domain.FreshnessStatusNeverReviewed))
}

// ---------------------------------------------------------------------------
// 4. FlaggedTicket
// ---------------------------------------------------------------------------

func TestFlaggedTicketSingleFlag(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("New module added", "k7m.20", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T12:00:00")
	ticket := domain.NewFlaggedTicket("k7m.25", "Ticket Health", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded)

	assert.Equal(t, "k7m.25", ticket.TicketID())
	assert.Equal(t, "Ticket Health", ticket.Title())
	assert.Len(t, ticket.Flags(), 1)
	assert.Equal(t, domain.FreshnessStatusReviewNeeded, ticket.Status())
}

func TestFlaggedTicketMultipleFlags(t *testing.T) {
	t.Parallel()
	diff1, _ := domain.NewContextDiff("Change one", "k7m.19", "2026-02-28")
	diff2, _ := domain.NewContextDiff("Change two", "k7m.20", "2026-03-01")
	flags := []domain.FreshnessFlag{
		domain.NewFreshnessFlag(diff1, "2026-02-28T10:00:00"),
		domain.NewFreshnessFlag(diff2, "2026-03-01T10:00:00"),
	}
	ticket := domain.NewFlaggedTicket("k7m.25", "Ticket Health", flags, domain.FreshnessStatusReviewNeeded)
	assert.Len(t, ticket.Flags(), 2)
}

func TestFlaggedTicketFlagCount(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "k7m.19", "2026-03-01")
	flags := make([]domain.FreshnessFlag, 3)
	for i := range flags {
		flags[i] = domain.NewFreshnessFlag(diff, "2026-03-01T00:00:00")
	}
	ticket := domain.NewFlaggedTicket("k7m.25", "Ticket Health", flags, domain.FreshnessStatusReviewNeeded)
	assert.Equal(t, 3, ticket.FlagCount())
}

func TestFlaggedTicketZeroFlags(t *testing.T) {
	t.Parallel()
	ticket := domain.NewFlaggedTicket("k7m.25", "Test", nil, domain.FreshnessStatusFresh)
	assert.Equal(t, 0, ticket.FlagCount())
}

// ---------------------------------------------------------------------------
// 5. TicketHealthReport
// ---------------------------------------------------------------------------

func TestTicketHealthReportReviewNeededCount(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "k7m.19", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T10:00:00")
	flagged := domain.NewFlaggedTicket("k7m.25", "Ticket Health", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded)

	report := domain.NewTicketHealthReport([]domain.FlaggedTicket{flagged}, 5, nil)
	assert.Equal(t, 1, report.ReviewNeededCount())
}

func TestTicketHealthReportHasIssuesTrue(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "k7m.19", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T10:00:00")
	flagged := domain.NewFlaggedTicket("k7m.25", "Ticket Health", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded)
	report := domain.NewTicketHealthReport([]domain.FlaggedTicket{flagged}, 5, nil)
	assert.True(t, report.HasIssues())
}

func TestTicketHealthReportHasIssuesFalse(t *testing.T) {
	t.Parallel()
	report := domain.NewTicketHealthReport(nil, 5, nil)
	assert.False(t, report.HasIssues())
}

func TestTicketHealthReportOldestLastReviewed(t *testing.T) {
	t.Parallel()
	date := "2026-01-15"
	report := domain.NewTicketHealthReport(nil, 10, &date)
	require.NotNil(t, report.OldestLastReviewed())
	assert.Equal(t, "2026-01-15", *report.OldestLastReviewed())
}

func TestTicketHealthReportOldestLastReviewedDefault(t *testing.T) {
	t.Parallel()
	report := domain.NewTicketHealthReport(nil, 0, nil)
	assert.Nil(t, report.OldestLastReviewed())
}

func TestFreshnessPctAllFresh(t *testing.T) {
	t.Parallel()
	report := domain.NewTicketHealthReport(nil, 10, nil)
	assert.InDelta(t, 100.0, report.FreshnessPct(), 0.001)
}

func TestFreshnessPctSomeStale(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "k7m.19", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T10:00:00")
	var flagged []domain.FlaggedTicket
	for i := range 3 {
		flagged = append(flagged, domain.NewFlaggedTicket(
			"k7m."+string(rune('0'+i)), "T", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded,
		))
	}
	report := domain.NewTicketHealthReport(flagged, 10, nil)
	assert.InDelta(t, 70.0, report.FreshnessPct(), 0.001)
}

func TestFreshnessPctAllStale(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "k7m.19", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T10:00:00")
	var flagged []domain.FlaggedTicket
	for i := range 10 {
		flagged = append(flagged, domain.NewFlaggedTicket(
			"k7m."+string(rune('0'+i)), "T", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded,
		))
	}
	report := domain.NewTicketHealthReport(flagged, 10, nil)
	assert.InDelta(t, 0.0, report.FreshnessPct(), 0.001)
}

func TestFreshnessPctZeroOpen(t *testing.T) {
	t.Parallel()
	report := domain.NewTicketHealthReport(nil, 0, nil)
	assert.InDelta(t, 100.0, report.FreshnessPct(), 0.001)
}

func TestFreshnessPctSingleStale(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "k7m.19", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T10:00:00")
	flagged := []domain.FlaggedTicket{
		domain.NewFlaggedTicket("k7m.25", "T", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded),
	}
	report := domain.NewTicketHealthReport(flagged, 1, nil)
	assert.InDelta(t, 0.0, report.FreshnessPct(), 0.001)
}

func TestFreshnessLabelHealthy(t *testing.T) {
	t.Parallel()
	report := domain.NewTicketHealthReport(nil, 10, nil)
	assert.Equal(t, "healthy", report.FreshnessLabel())
}

func TestFreshnessLabelAcceptable(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "t.1", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01")
	var flagged []domain.FlaggedTicket
	for i := range 2 {
		flagged = append(flagged, domain.NewFlaggedTicket(
			"t."+string(rune('0'+i)), "T", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded,
		))
	}
	report := domain.NewTicketHealthReport(flagged, 10, nil)
	assert.Equal(t, "acceptable", report.FreshnessLabel())
}

func TestFreshnessLabelActionNeeded(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Change", "t.1", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01")
	var flagged []domain.FlaggedTicket
	for i := range 4 {
		flagged = append(flagged, domain.NewFlaggedTicket(
			"t."+string(rune('0'+i)), "T", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded,
		))
	}
	report := domain.NewTicketHealthReport(flagged, 10, nil)
	assert.Equal(t, "action needed", report.FreshnessLabel())
}

func TestFreshnessPctReturnsFloat(t *testing.T) {
	t.Parallel()
	// partial: 2/3 = 66.666...
	diff, _ := domain.NewContextDiff("Change", "k7m.19", "2026-03-01")
	flag := domain.NewFreshnessFlag(diff, "2026-03-01T10:00:00")
	flagged := []domain.FlaggedTicket{
		domain.NewFlaggedTicket("k7m.25", "T", []domain.FreshnessFlag{flag}, domain.FreshnessStatusReviewNeeded),
	}
	report := domain.NewTicketHealthReport(flagged, 3, nil)
	result := report.FreshnessPct()
	assert.InDelta(t, 66.6667, result, 0.001)
}

// ---------------------------------------------------------------------------
// 6. EpicHealthSummary
// ---------------------------------------------------------------------------

func TestEpicHealthSummaryRejectsNegativeFresh(t *testing.T) {
	t.Parallel()
	_, err := domain.NewEpicHealthSummary("k7m", 5, -1, 6)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestEpicHealthSummaryRejectsNegativeStale(t *testing.T) {
	t.Parallel()
	_, err := domain.NewEpicHealthSummary("k7m", 5, 6, -1)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestEpicHealthSummaryRejectsNegativeTotal(t *testing.T) {
	t.Parallel()
	_, err := domain.NewEpicHealthSummary("k7m", -2, -1, -1)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestEpicHealthSummaryRejectsSumMismatch(t *testing.T) {
	t.Parallel()
	_, err := domain.NewEpicHealthSummary("k7m", 5, 2, 2)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestEpicHealthSummaryFreshnessPct(t *testing.T) {
	t.Parallel()
	s, err := domain.NewEpicHealthSummary("k7m", 10, 8, 2)
	require.NoError(t, err)
	assert.InDelta(t, 80.0, s.FreshnessPct(), 0.001)
}

func TestEpicHealthSummaryFreshnessPctZeroTickets(t *testing.T) {
	t.Parallel()
	s, err := domain.NewEpicHealthSummary("k7m", 0, 0, 0)
	require.NoError(t, err)
	assert.InDelta(t, 100.0, s.FreshnessPct(), 0.001)
}

// ---------------------------------------------------------------------------
// 7. OpenTicketData
// ---------------------------------------------------------------------------

func TestOpenTicketDataFields(t *testing.T) {
	t.Parallel()
	date := "2026-02-28"
	data := domain.NewOpenTicketData("k7m.25", "Ticket Health", []string{"review_needed", "core"}, &date)
	assert.Equal(t, "k7m.25", data.TicketID())
	assert.Equal(t, "Ticket Health", data.Title())
	assert.Equal(t, []string{"review_needed", "core"}, data.Labels())
	require.NotNil(t, data.LastReviewed())
	assert.Equal(t, "2026-02-28", *data.LastReviewed())
}

func TestOpenTicketDataDefaultLastReviewed(t *testing.T) {
	t.Parallel()
	data := domain.NewOpenTicketData("k7m.25", "Ticket Health", nil, nil)
	assert.Nil(t, data.LastReviewed())
}
