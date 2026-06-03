package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/ticket/application"
	"github.com/alto-cli/alto/internal/ticket/infrastructure"
)

// Compile-time interface check.
var _ application.BeadsCommentWriter = (*infrastructure.BeadsCommentWriter)(nil)

type captureCommentRunner struct {
	ticketID string
	body     string
	err      error
	called   bool
	deadline bool
}

func (c *captureCommentRunner) run(ctx context.Context, ticketID, body string) error {
	c.called = true
	c.ticketID = ticketID
	c.body = body
	_, c.deadline = ctx.Deadline()
	return c.err
}

func TestBeadsCommentWriter_NewBeadsCommentWriter_Defaults(t *testing.T) {
	t.Parallel()

	w := infrastructure.NewBeadsCommentWriter()

	require.NotNil(t, w)
	assert.Equal(t, 5*time.Second, w.Timeout())
}

func TestBeadsCommentWriter_AddComment_InvokesRunnerWithBody(t *testing.T) {
	t.Parallel()

	c := &captureCommentRunner{}
	w := infrastructure.NewBeadsCommentWriterWithRunner(c.run)

	err := w.AddComment(context.Background(), "alty-cli-bf7", "**ripple review** body")

	require.NoError(t, err)
	assert.True(t, c.called)
	assert.Equal(t, "alty-cli-bf7", c.ticketID)
	assert.Equal(t, "**ripple review** body", c.body)
	assert.True(t, c.deadline, "adapter must apply its own timeout deadline")
}

func TestBeadsCommentWriter_AddComment_RejectsEmptyTicketID(t *testing.T) {
	t.Parallel()

	c := &captureCommentRunner{}
	w := infrastructure.NewBeadsCommentWriterWithRunner(c.run)

	err := w.AddComment(context.Background(), "", "body")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ticket ID")
	assert.False(t, c.called)
}

func TestBeadsCommentWriter_AddComment_RejectsEmptyComment(t *testing.T) {
	t.Parallel()

	c := &captureCommentRunner{}
	w := infrastructure.NewBeadsCommentWriterWithRunner(c.run)

	err := w.AddComment(context.Background(), "alty-cli-bf7", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment")
	assert.False(t, c.called)
}

func TestBeadsCommentWriter_AddComment_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	c := &captureCommentRunner{}
	w := infrastructure.NewBeadsCommentWriterWithRunner(c.run)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := w.AddComment(ctx, "alty-cli-bf7", "body")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestBeadsCommentWriter_AddComment_WrapsRunnerError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("bd boom")
	c := &captureCommentRunner{err: sentinel}
	w := infrastructure.NewBeadsCommentWriterWithRunner(c.run)

	err := w.AddComment(context.Background(), "alty-cli-bf7", "body")

	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	assert.Contains(t, err.Error(), "running bd comment add alty-cli-bf7")
}
