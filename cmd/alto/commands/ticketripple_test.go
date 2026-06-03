package commands_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/cmd/alto/commands"
	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/ticket/application"
)

// ---------------------------------------------------------------------------
// Fakes for the App.RippleHandler dependency
// ---------------------------------------------------------------------------

type fakeGraphReader struct {
	closeReason map[string]string
	dependents  map[string][]application.BeadsGraphTicket
}

func (f *fakeGraphReader) ReadParent(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (f *fakeGraphReader) ReadSiblings(_ context.Context, _, _ string) ([]application.BeadsGraphTicket, error) {
	return nil, nil
}

func (f *fakeGraphReader) ReadDependents(_ context.Context, ticketID string) ([]application.BeadsGraphTicket, error) {
	return f.dependents[ticketID], nil
}

func (f *fakeGraphReader) ReadRelated(_ context.Context, _ string) ([]application.BeadsGraphTicket, error) {
	return nil, nil
}

func (f *fakeGraphReader) ReadCloseContext(_ context.Context, ticketID string) (string, error) {
	return f.closeReason[ticketID], nil
}

type recordingLabelWriter struct{ calls []string }

func (r *recordingLabelWriter) AddLabel(_ context.Context, ticketID, _ string) error {
	r.calls = append(r.calls, ticketID)
	return nil
}

func (r *recordingLabelWriter) RemoveLabel(_ context.Context, _, _ string) error { return nil }

type recordingCommentWriter struct{ bodies []string }

func (r *recordingCommentWriter) AddComment(_ context.Context, _, body string) error {
	r.bodies = append(r.bodies, body)
	return nil
}

// appWithRippleHandler builds a minimal composition.App carrying only the
// RippleHandler — the Cobra command depends on App.RippleHandler and
// nothing else, so this is the smallest viable wiring for unit tests.
func appWithRippleHandler(handler *application.RippleHandler) *composition.App {
	return &composition.App{RippleHandler: handler}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestTicketRippleCmd_RequiresExactlyOneTicketIDArg(t *testing.T) {
	t.Parallel()

	app := appWithRippleHandler(application.NewRippleHandler(
		&fakeGraphReader{}, &recordingLabelWriter{}, &recordingCommentWriter{}))
	cmd := commands.NewTicketRippleCmd(app)
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg")
}

func TestTicketRippleCmd_FlagsOpenDependentsAndPrintsSummary(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {{ID: "alty-cli-bf7", Status: "open"}},
		},
	}
	lw := &recordingLabelWriter{}
	cw := &recordingCommentWriter{}
	app := appWithRippleHandler(application.NewRippleHandler(reader, lw, cw))

	out := &bytes.Buffer{}
	cmd := commands.NewTicketRippleCmd(app)
	cmd.SetArgs([]string{"alty-cli-2f9", "--context", "subcommand shipped"})
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, []string{"alty-cli-bf7"}, lw.calls)
	require.Len(t, cw.bodies, 1)
	assert.Contains(t, cw.bodies[0], "subcommand shipped")
	assert.Contains(t, out.String(), "Flagged 1 ticket")
}

func TestTicketRippleCmd_ContextFlagPassesThroughToHandler(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		closeReason: map[string]string{"alty-cli-2f9": "stale close reason"},
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {{ID: "alty-cli-bf7", Status: "open"}},
		},
	}
	lw := &recordingLabelWriter{}
	cw := &recordingCommentWriter{}
	app := appWithRippleHandler(application.NewRippleHandler(reader, lw, cw))

	cmd := commands.NewTicketRippleCmd(app)
	cmd.SetArgs([]string{"alty-cli-2f9", "--context", "fresh override"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.Len(t, cw.bodies, 1)
	assert.Contains(t, cw.bodies[0], "fresh override")
	assert.NotContains(t, cw.bodies[0], "stale close reason")
}

func TestTicketRippleCmd_FallsBackToCloseReasonWhenNoContext(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{
		closeReason: map[string]string{"alty-cli-2f9": "shipped via close_reason"},
		dependents: map[string][]application.BeadsGraphTicket{
			"alty-cli-2f9": {{ID: "alty-cli-bf7", Status: "open"}},
		},
	}
	lw := &recordingLabelWriter{}
	cw := &recordingCommentWriter{}
	app := appWithRippleHandler(application.NewRippleHandler(reader, lw, cw))

	cmd := commands.NewTicketRippleCmd(app)
	cmd.SetArgs([]string{"alty-cli-2f9"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	require.Len(t, cw.bodies, 1)
	assert.Contains(t, cw.bodies[0], "shipped via close_reason")
}

func TestTicketRippleCmd_ReturnsErrorWhenContextEmpty(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{closeReason: map[string]string{}} // no close_reason, no override
	lw := &recordingLabelWriter{}
	cw := &recordingCommentWriter{}
	app := appWithRippleHandler(application.NewRippleHandler(reader, lw, cw))

	cmd := commands.NewTicketRippleCmd(app)
	cmd.SetArgs([]string{"alty-cli-2f9"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-empty context summary")
}

func TestTicketRippleCmd_EmptyGraphPrintsNoOpMessage(t *testing.T) {
	t.Parallel()

	reader := &fakeGraphReader{} // no dependents, no parent
	lw := &recordingLabelWriter{}
	cw := &recordingCommentWriter{}
	app := appWithRippleHandler(application.NewRippleHandler(reader, lw, cw))

	out := &bytes.Buffer{}
	cmd := commands.NewTicketRippleCmd(app)
	cmd.SetArgs([]string{"alty-cli-2f9", "--context", "x"})
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})

	require.NoError(t, cmd.Execute())
	assert.Empty(t, lw.calls)
	assert.Contains(t, out.String(), "No open siblings or dependents")
}

func TestTicketRippleCmd_PropagatesHandlerError(t *testing.T) {
	t.Parallel()

	// Reader that errors — the handler should bubble up; the cobra command
	// wraps the error with "ripple review: ...".
	reader := &erroringGraphReader{err: errors.New("transport boom")}
	lw := &recordingLabelWriter{}
	cw := &recordingCommentWriter{}
	app := appWithRippleHandler(application.NewRippleHandler(reader, lw, cw))

	cmd := commands.NewTicketRippleCmd(app)
	cmd.SetArgs([]string{"alty-cli-2f9", "--context", "x"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ripple review")
	assert.Contains(t, err.Error(), "transport boom")
}

type erroringGraphReader struct{ err error }

func (e *erroringGraphReader) ReadParent(_ context.Context, _ string) (string, error) {
	return "", e.err
}

func (e *erroringGraphReader) ReadSiblings(_ context.Context, _, _ string) ([]application.BeadsGraphTicket, error) {
	return nil, e.err
}

func (e *erroringGraphReader) ReadDependents(_ context.Context, _ string) ([]application.BeadsGraphTicket, error) {
	return nil, e.err
}

func (e *erroringGraphReader) ReadRelated(_ context.Context, _ string) ([]application.BeadsGraphTicket, error) {
	return nil, e.err
}

func (e *erroringGraphReader) ReadCloseContext(_ context.Context, _ string) (string, error) {
	return "", e.err
}
