package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/ticket/application"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

type stubContentReader struct {
	content map[string]string
}

func (s *stubContentReader) ReadTicketContent(_ context.Context, ticketID string) (string, error) {
	if c, ok := s.content[ticketID]; ok {
		return c, nil
	}
	return "", nil
}

type stubCommandRunner struct {
	outputs map[string]string
}

func (s *stubCommandRunner) Run(_ context.Context, command string) (string, error) {
	if o, ok := s.outputs[command]; ok {
		return o, nil
	}
	return "", nil
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func TestTicketVerifyHandler_VerifiesClaimsInTicket(t *testing.T) {
	t.Parallel()

	reader := &stubContentReader{
		content: map[string]string{
			"t-123": "## Analysis\n\n```bash\ndeadcode ./...\n```\n\nFound **14 issues** to fix.\n",
		},
	}
	runner := &stubCommandRunner{
		outputs: map[string]string{
			"deadcode ./...": "14\n",
		},
	}

	handler := application.NewTicketVerifyHandler(reader, runner)
	results, err := handler.Verify(context.Background(), "t-123")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].Match())
}

func TestTicketVerifyHandler_DetectsMismatch(t *testing.T) {
	t.Parallel()

	reader := &stubContentReader{
		content: map[string]string{
			"t-123": "## Analysis\n\n```bash\ndeadcode ./...\n```\n\nFound **14 issues** to fix.\n",
		},
	}
	runner := &stubCommandRunner{
		outputs: map[string]string{
			"deadcode ./...": "288\n",
		},
	}

	handler := application.NewTicketVerifyHandler(reader, runner)
	results, err := handler.Verify(context.Background(), "t-123")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.False(t, results[0].Match())
	assert.Contains(t, results[0].Discrepancy(), "claimed 14")
	assert.Contains(t, results[0].Discrepancy(), "actual 288")
}

func TestTicketVerifyHandler_NoClaimsReturnsEmpty(t *testing.T) {
	t.Parallel()

	reader := &stubContentReader{
		content: map[string]string{
			"t-123": "## Design\n\nThis ticket has no quantitative claims.\n",
		},
	}
	runner := &stubCommandRunner{}

	handler := application.NewTicketVerifyHandler(reader, runner)
	results, err := handler.Verify(context.Background(), "t-123")

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestTicketVerifyHandler_EmptyTicketIDReturnsError(t *testing.T) {
	t.Parallel()

	handler := application.NewTicketVerifyHandler(&stubContentReader{}, &stubCommandRunner{})
	_, err := handler.Verify(context.Background(), "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ticket ID")
}
