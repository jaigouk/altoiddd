package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	"github.com/alto-cli/alto/internal/ticket/domain"
)

func makeContextDiff() domain.ContextDiff {
	d, _ := domain.NewContextDiff("Implemented fitness test generation", "k7m.19", "2026-03-01")
	return d
}

// ---------------------------------------------------------------------------
// 1. Creation
// ---------------------------------------------------------------------------

func TestRippleReviewCreation(t *testing.T) {
	t.Parallel()
	diff := makeContextDiff()
	review := domain.NewRippleReview("rr-001", "k7m.19", diff)

	assert.Equal(t, "rr-001", review.ReviewID())
	assert.Equal(t, "k7m.19", review.ClosedTicketID())
	assert.Equal(t, diff, review.ContextDiff())
	assert.Empty(t, review.FlaggedTickets())
	assert.Empty(t, review.Events())
}

// ---------------------------------------------------------------------------
// 2. Flagging tickets
// ---------------------------------------------------------------------------

func TestRippleReviewFlagOpenTicket(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	err := review.FlagTicket("k7m.25", true, "")
	require.NoError(t, err)
	assert.Contains(t, review.FlaggedTickets(), "k7m.25")
}

func TestRippleReviewRejectsClosedTicket(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	err := review.FlagTicket("k7m.18", false, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestRippleReviewFlagMultipleTickets(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.FlagTicket("k7m.20", true, "")

	flagged := review.FlaggedTickets()
	assert.Len(t, flagged, 2)
	assert.Contains(t, flagged, "k7m.25")
	assert.Contains(t, flagged, "k7m.20")
}

// ---------------------------------------------------------------------------
// 3. Clearing flags
// ---------------------------------------------------------------------------

func TestRippleReviewClearFlag(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	err := review.ClearFlag("k7m.25", "")
	require.NoError(t, err)
	assert.NotContains(t, review.FlaggedTickets(), "k7m.25")
}

func TestRippleReviewClearUnflaggedRaises(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	err := review.ClearFlag("k7m.25", "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// ---------------------------------------------------------------------------
// 4. Domain events
// ---------------------------------------------------------------------------

func TestRippleReviewEmitsTicketFlaggedEvent(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")

	events := review.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(domain.TicketFlagged)
	require.True(t, ok)
	assert.Equal(t, "rr-001", evt.ReviewID())
	assert.Equal(t, "k7m.25", evt.TicketID())
}

func TestRippleReviewEmitsFlagClearedEvent(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.ClearFlag("k7m.25", "")

	events := review.Events()
	require.Len(t, events, 2)
	_, ok := events[1].(domain.FlagCleared)
	require.True(t, ok)
}

func TestRippleReviewCompleteEmitsRippleReviewCreated(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.FlagTicket("k7m.20", true, "")

	err := review.Complete()
	require.NoError(t, err)

	events := review.Events()
	require.Len(t, events, 3) // 2 TicketFlagged + 1 RippleReviewCreated
	evt, ok := events[2].(domain.RippleReviewCreated)
	require.True(t, ok)
	assert.Equal(t, "rr-001", evt.ReviewID())
	assert.Equal(t, "k7m.19", evt.ClosedTicketID())
	assert.Equal(t, 2, evt.FlaggedCount())
}

func TestRippleReviewCompleteWithZeroFlagged(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())

	err := review.Complete()
	require.NoError(t, err)

	events := review.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(domain.RippleReviewCreated)
	require.True(t, ok)
	assert.Equal(t, 0, evt.FlaggedCount())
}

func TestRippleReviewCompleteIsIdempotent(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")

	_ = review.Complete()
	err := review.Complete()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// ---------------------------------------------------------------------------
// 5. Defensive copies
// ---------------------------------------------------------------------------

func TestRippleReviewFlaggedTicketsDefensiveCopy(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")

	flagged := review.FlaggedTickets()
	assert.Equal(t, review.FlaggedTickets(), flagged)
}

func TestRippleReviewEventsDefensiveCopy(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")

	events := review.Events()
	assert.Equal(t, review.Events(), events)
}

// ---------------------------------------------------------------------------
// 6. Review checklist
// ---------------------------------------------------------------------------

func TestReviewChecklistTemplateExists(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, domain.ReviewChecklistTemplate)
}

func TestReviewChecklistTemplateHasKeyItems(t *testing.T) {
	t.Parallel()
	lower := domain.ReviewChecklistTemplate
	assert.Contains(t, lower, "description")
	assert.Contains(t, lower, "acceptance criteria")
}

func TestBuildRippleCommentIncludesChecklist(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	comment := review.BuildRippleComment()
	assert.Contains(t, comment, "k7m.19")
	assert.Contains(t, comment, "fitness")
	assert.Contains(t, comment, domain.ReviewChecklistTemplate)
}

// ---------------------------------------------------------------------------
// Edge cases: duplicate flagging, lifecycle
// ---------------------------------------------------------------------------

func TestFlagSameTicketTwice(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.FlagTicket("k7m.25", true, "")

	flagged := review.FlaggedTickets()
	count := 0
	for _, id := range flagged {
		if id == "k7m.25" {
			count++
		}
	}
	assert.Equal(t, 2, count)
}

func TestClearOneInstanceOfDuplicateFlag(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.ClearFlag("k7m.25", "")

	count := 0
	for _, id := range review.FlaggedTickets() {
		if id == "k7m.25" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestFlagThreeThenClearAll(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.FlagTicket("k7m.20", true, "")
	_ = review.FlagTicket("k7m.21", true, "")
	assert.Len(t, review.FlaggedTickets(), 3)

	_ = review.ClearFlag("k7m.25", "")
	_ = review.ClearFlag("k7m.20", "")
	_ = review.ClearFlag("k7m.21", "")
	assert.Empty(t, review.FlaggedTickets())
}

func TestClearAlreadyClearedRaises(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.ClearFlag("k7m.25", "")

	err := review.ClearFlag("k7m.25", "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestEventsAccumulateThroughLifecycle(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.19", makeContextDiff())
	_ = review.FlagTicket("k7m.25", true, "")
	_ = review.FlagTicket("k7m.20", true, "")
	_ = review.ClearFlag("k7m.25", "")

	events := review.Events()
	assert.Len(t, events, 3)
	_, ok0 := events[0].(domain.TicketFlagged)
	_, ok1 := events[1].(domain.TicketFlagged)
	_, ok2 := events[2].(domain.FlagCleared)
	assert.True(t, ok0)
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func TestBuildRippleCommentIncludesClosedTicketID(t *testing.T) {
	t.Parallel()
	review := domain.NewRippleReview("rr-001", "k7m.42", makeContextDiff())
	comment := review.BuildRippleComment()
	assert.Contains(t, comment, "k7m.42")
}

func TestBuildRippleCommentIncludesSummary(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Added new StackProfile protocol", "k7m.19", "2026-03-01")
	review := domain.NewRippleReview("rr-001", "k7m.19", diff)
	comment := review.BuildRippleComment()
	assert.Contains(t, comment, "Added new StackProfile protocol")
}

func TestBuildRippleCommentSpecialChars(t *testing.T) {
	t.Parallel()
	diff, _ := domain.NewContextDiff("Added `code` & 'quotes' + <tags>", "k7m.19", "2026-03-01")
	review := domain.NewRippleReview("rr-001", "k7m.19", diff)
	comment := review.BuildRippleComment()
	assert.Contains(t, comment, "`code`")
	assert.Contains(t, comment, "<tags>")
}
