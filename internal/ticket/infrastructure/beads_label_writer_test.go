package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/ticket/application"
	"github.com/alto-cli/alto/internal/ticket/infrastructure"
)

// ---------------------------------------------------------------------------
// Compile-time interface check
// ---------------------------------------------------------------------------

var _ application.BeadsLabelWriter = (*infrastructure.BeadsLabelWriter)(nil)

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewBeadsLabelWriter_ReturnsWriter(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()

	require.NotNil(t, writer)
}

// ---------------------------------------------------------------------------
// AddLabel
// ---------------------------------------------------------------------------

func TestBeadsLabelWriter_AddLabel_RejectsEmptyTicketID(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()
	ctx := context.Background()

	err := writer.AddLabel(ctx, "", "review_needed")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ticket ID")
}

func TestBeadsLabelWriter_AddLabel_RejectsEmptyLabel(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()
	ctx := context.Background()

	err := writer.AddLabel(ctx, "alto-123", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
}

func TestBeadsLabelWriter_AddLabel_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := writer.AddLabel(ctx, "alto-123", "review_needed")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// ---------------------------------------------------------------------------
// RemoveLabel
// ---------------------------------------------------------------------------

func TestBeadsLabelWriter_RemoveLabel_RejectsEmptyTicketID(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()
	ctx := context.Background()

	err := writer.RemoveLabel(ctx, "", "review_needed")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ticket ID")
}

func TestBeadsLabelWriter_RemoveLabel_RejectsEmptyLabel(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()
	ctx := context.Background()

	err := writer.RemoveLabel(ctx, "alto-123", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
}

func TestBeadsLabelWriter_RemoveLabel_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := writer.RemoveLabel(ctx, "alto-123", "review_needed")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// ---------------------------------------------------------------------------
// Timeout
// ---------------------------------------------------------------------------

func TestBeadsLabelWriter_DefaultTimeout(t *testing.T) {
	t.Parallel()

	writer := infrastructure.NewBeadsLabelWriter()

	assert.Equal(t, 5*time.Second, writer.Timeout())
}
