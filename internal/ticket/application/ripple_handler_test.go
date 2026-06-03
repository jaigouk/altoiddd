package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/ticket/application"
	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type fakeGraphReader struct {
	parentByID     map[string]string
	siblings       map[string][]application.BeadsGraphTicket // keyed by parentID
	dependents     map[string][]application.BeadsGraphTicket
	related        map[string][]application.BeadsGraphTicket
	closeContext   map[string]string
	parentErr      error
	siblingsErr    error
	dependentsErr  error
	relatedErr     error
	closeCtxErr    error
	siblingsCallFn func(parentID, selfID string)
}

func (f *fakeGraphReader) ReadParent(_ context.Context, ticketID string) (string, error) {
	if f.parentErr != nil {
		return "", f.parentErr
	}
	return f.parentByID[ticketID], nil
}

func (f *fakeGraphReader) ReadSiblings(_ context.Context, parentID, selfID string) ([]application.BeadsGraphTicket, error) {
	if f.siblingsCallFn != nil {
		f.siblingsCallFn(parentID, selfID)
	}
	if f.siblingsErr != nil {
		return nil, f.siblingsErr
	}
	return f.siblings[parentID], nil
}

func (f *fakeGraphReader) ReadDependents(_ context.Context, ticketID string) ([]application.BeadsGraphTicket, error) {
	if f.dependentsErr != nil {
		return nil, f.dependentsErr
	}
	return f.dependents[ticketID], nil
}

func (f *fakeGraphReader) ReadRelated(_ context.Context, ticketID string) ([]application.BeadsGraphTicket, error) {
	if f.relatedErr != nil {
		return nil, f.relatedErr
	}
	return f.related[ticketID], nil
}

func (f *fakeGraphReader) ReadCloseContext(_ context.Context, ticketID string) (string, error) {
	if f.closeCtxErr != nil {
		return "", f.closeCtxErr
	}
	return f.closeContext[ticketID], nil
}

type fakeLabelWriter struct {
	calls []labelCall
	err   error
}

type labelCall struct {
	op       string
	ticketID string
	label    string
}

func (f *fakeLabelWriter) AddLabel(_ context.Context, ticketID, label string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, labelCall{op: "add", ticketID: ticketID, label: label})
	return nil
}

func (f *fakeLabelWriter) RemoveLabel(_ context.Context, ticketID, label string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, labelCall{op: "remove", ticketID: ticketID, label: label})
	return nil
}

type fakeCommentWriter struct {
	calls []commentCall
	err   error
}

type commentCall struct {
	ticketID string
	comment  string
}

func (f *fakeCommentWriter) AddComment(_ context.Context, ticketID, comment string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, commentCall{ticketID: ticketID, comment: comment})
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func openTicket(id string) application.BeadsGraphTicket {
	return application.BeadsGraphTicket{ID: id, Status: "open"}
}

func inProgressTicket(id string) application.BeadsGraphTicket {
	return application.BeadsGraphTicket{ID: id, Status: "in_progress"}
}

func closedTicket(id string) application.BeadsGraphTicket {
	return application.BeadsGraphTicket{ID: id, Status: "closed"}
}

func newHandlerWithFakes(reader *fakeGraphReader, lw *fakeLabelWriter, cw *fakeCommentWriter) *application.RippleHandler {
	h := application.NewRippleHandler(reader, lw, cw)
	h.WithClock(func() time.Time {
		return time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	})
	h.WithIDGenerator(func() string { return "review-fixed" })
	return h
}

func ripStrPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRippleHandler_FlagsOpenSiblings(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		parentByID: map[string]string{"alty-cli-2f9": "alty-cli-epic"},
		siblings: map[string][]application.BeadsGraphTicket{
			"alty-cli-epic": {openTicket("alty-cli-bf7"), openTicket("alty-cli-zc9")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("Ripple subcommand shipped"))

	require.NoError(t, err)
	assert.Equal(t, 2, result.FlaggedCount)
	require.Len(t, lw.calls, 2)
	assert.Equal(t, "alty-cli-bf7", lw.calls[0].ticketID)
	assert.Equal(t, "review_needed", lw.calls[0].label)
	assert.Equal(t, "alty-cli-zc9", lw.calls[1].ticketID)
}

func TestRippleHandler_FlagsOpenDependents(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 1, result.FlaggedCount)
	require.Len(t, lw.calls, 1)
	assert.Equal(t, "alty-cli-bf7", lw.calls[0].ticketID)
}

func TestRippleHandler_FlagsOpenRelated(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		related: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-aaa")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 1, result.FlaggedCount)
	require.Len(t, lw.calls, 1)
	assert.Equal(t, "alty-cli-aaa", lw.calls[0].ticketID)
}

func TestRippleHandler_DeduplicatesTicketsAcrossSiblingsDependentsRelated(t *testing.T) {
	t.Parallel()

	// alty-cli-bf7 appears in all three sources — must only be flagged once.
	reader := &fakeGraphReader{
		parentByID: map[string]string{"alty-cli-2f9": "epic"},
		siblings: map[string][]application.BeadsGraphTicket{
			"epic": {openTicket("alty-cli-bf7")},
		},
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7"), openTicket("alty-cli-zc9")},
		},
		related: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7"), openTicket("alty-cli-zc9")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 2, result.FlaggedCount, "bf7 + zc9 — bf7 dedup'd from 3 sources")
	require.Len(t, lw.calls, 2)
	assert.Equal(t, "alty-cli-bf7", lw.calls[0].ticketID, "first occurrence wins (siblings ordered first)")
	assert.Equal(t, "alty-cli-zc9", lw.calls[1].ticketID)
}

func TestRippleHandler_SkipsClosedTickets(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		parentByID: map[string]string{"alty-cli-2f9": "epic"},
		siblings: map[string][]application.BeadsGraphTicket{
			"epic": {closedTicket("alty-cli-old"), openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 1, result.FlaggedCount)
	require.Len(t, lw.calls, 1)
	assert.Equal(t, "alty-cli-bf7", lw.calls[0].ticketID)
}

func TestRippleHandler_TreatsInProgressAsOpen(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {inProgressTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 1, result.FlaggedCount)
}

func TestRippleHandler_ExcludesSelf(t *testing.T) {
	t.Parallel()

	// Adapter returns self as a sibling (shouldn't happen, but guard
	// against it at the handler boundary).
	reader := &fakeGraphReader{
		parentByID: map[string]string{"alty-cli-2f9": "epic"},
		siblings: map[string][]application.BeadsGraphTicket{
			"epic": {openTicket("alty-cli-2f9"), openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 1, result.FlaggedCount)
	require.Len(t, lw.calls, 1)
	assert.Equal(t, "alty-cli-bf7", lw.calls[0].ticketID)
}

func TestRippleHandler_BuildsCommentFromAggregate(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("Ripple shipped"))

	require.NoError(t, err)
	require.Len(t, cw.calls, 1)

	// The comment shape MUST come from the domain aggregate, not a
	// handler-local template. Reconstruct the expected shape via the same
	// aggregate to assert structural equivalence.
	diff, derr := ticketdomain.NewContextDiff("Ripple shipped", "alty-cli-2f9", "ignored")
	require.NoError(t, derr)
	expected := ticketdomain.NewRippleReview("rid", "alty-cli-2f9", diff).BuildRippleComment()
	assert.Equal(t, expected, cw.calls[0].comment)
}

func TestRippleHandler_FallsBackToCloseReasonWhenOverrideNil(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		closeContext: map[string]string{
			"alty-cli-2f9": "Subcommand wired in composition",
		},
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", nil)

	require.NoError(t, err)
	assert.Equal(t, 1, result.FlaggedCount)
	require.Len(t, cw.calls, 1)
	assert.Contains(t, cw.calls[0].comment, "Subcommand wired in composition")
}

func TestRippleHandler_RejectsEmptyContextSummary(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		closeContext: map[string]string{
			"alty-cli-2f9": "", // close_reason missing
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-empty context summary")
}

func TestRippleHandler_RejectsClosedStringAsContextSummary(t *testing.T) {
	t.Parallel()

	// bash bd-ripple treats "Closed" (the bd CLI's default reason) as
	// empty. Mirror that semantics so the Go subcommand is a drop-in.
	reader := &fakeGraphReader{
		closeContext: map[string]string{
			"alty-cli-2f9": "Closed",
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-empty context summary")
}

func TestRippleHandler_EmitsTicketFlaggedAndRippleReviewCreatedEvents(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7"), openTicket("alty-cli-zc9")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	// 2 TicketFlagged + 1 RippleReviewCreated = 3
	assert.Equal(t, 3, result.EventCount)
}

func TestRippleHandler_ReturnsEmptyResultWhenNoOpenNeighbours(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		parentByID: map[string]string{"alty-cli-2f9": "epic"},
		siblings: map[string][]application.BeadsGraphTicket{
			"epic": {closedTicket("alty-cli-old")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 0, result.FlaggedCount)
	assert.Empty(t, lw.calls)
	assert.Empty(t, cw.calls)
	// Aggregate still emits RippleReviewCreated even when no flags raised.
	assert.Equal(t, 1, result.EventCount)
}

func TestRippleHandler_SkipsSiblingLookupWhenNoParent(t *testing.T) {
	t.Parallel()

	called := false
	reader := &fakeGraphReader{
		parentByID: map[string]string{}, // no parent
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7")},
		},
		siblingsCallFn: func(_, _ string) { called = true },
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	result, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	assert.Equal(t, 1, result.FlaggedCount)
	assert.False(t, called, "ReadSiblings must NOT be called when parent is empty")
}

func TestRippleHandler_RejectsEmptyClosedTicketID(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "", ripStrPtr("ctx"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed ticket ID")
}

func TestRippleHandler_PropagatesGraphReaderError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("bd transport boom")
	reader := &fakeGraphReader{parentErr: sentinel}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestRippleHandler_PropagatesLabelWriterError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("label boom")
	reader := &fakeGraphReader{
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{err: sentinel}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestRippleHandler_PropagatesCommentWriterError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("comment boom")
	reader := &fakeGraphReader{
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{err: sentinel}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestRippleHandler_OverrideTakesPrecedenceOverCloseReason(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		closeContext: map[string]string{
			"alty-cli-2f9": "stale close reason",
		},
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bf7")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("fresh override"))

	require.NoError(t, err)
	require.Len(t, cw.calls, 1)
	assert.Contains(t, cw.calls[0].comment, "fresh override")
	assert.NotContains(t, cw.calls[0].comment, "stale close reason")
}

func TestRippleHandler_PreservesStableOrderingAcrossSources(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		parentByID: map[string]string{"alty-cli-2f9": "epic"},
		siblings: map[string][]application.BeadsGraphTicket{
			"epic": {openTicket("alty-cli-aaa")},
		},
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-bbb")},
		},
		related: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {openTicket("alty-cli-ccc")},
		},
	}
	lw := &fakeLabelWriter{}
	cw := &fakeCommentWriter{}
	h := newHandlerWithFakes(reader, lw, cw)

	_, err := h.Handle(context.Background(), "alty-cli-2f9", ripStrPtr("ctx"))

	require.NoError(t, err)
	require.Len(t, lw.calls, 3)
	assert.Equal(t, "alty-cli-aaa", lw.calls[0].ticketID, "siblings first")
	assert.Equal(t, "alty-cli-bbb", lw.calls[1].ticketID, "dependents second")
	assert.Equal(t, "alty-cli-ccc", lw.calls[2].ticketID, "related third")
}

func TestRippleHandler_CompilesAgainstPortInterfaces(t *testing.T) {
	t.Parallel()

	// Compile-time guard: the fakes must satisfy the port interfaces the
	// handler depends on. If a port grows a method, this test fails to
	// compile rather than at runtime.
	var (
		_ application.BeadsGraphReader   = (*fakeGraphReader)(nil)
		_ application.BeadsLabelWriter   = (*fakeLabelWriter)(nil)
		_ application.BeadsCommentWriter = (*fakeCommentWriter)(nil)
	)
}
