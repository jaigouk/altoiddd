package infrastructure_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Compile-time interface check.
var _ application.BoundaryPrompter = (*infrastructure.HuhBoundaryPrompter)(nil)

func TestHuhBoundaryPrompter_DisplayBoundaryProposals_EmptySlice(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewHuhBoundaryPrompter()
	accepted, err := prompter.DisplayBoundaryProposals(context.Background(), nil)

	require.NoError(t, err)
	assert.Empty(t, accepted)
}

func TestHuhBoundaryPrompter_DisplayBoundaryProposals_EmptySliceExplicit(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewHuhBoundaryPrompter()
	accepted, err := prompter.DisplayBoundaryProposals(context.Background(), []discoverydomain.BoundedContextSketch{})

	require.NoError(t, err)
	assert.Empty(t, accepted)
}

func TestHuhBoundaryPrompter_DisplayBoundaryProposals_CancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	prompter := infrastructure.NewHuhBoundaryPrompter()

	sketch, err := discoverydomain.NewBoundedContextSketch(
		"TestContext", vo.SubdomainCore, 0.80,
		[]string{"User"}, nil, nil, nil, vo.AIInferred,
	)
	require.NoError(t, err)

	accepted, err := prompter.DisplayBoundaryProposals(ctx, []discoverydomain.BoundedContextSketch{sketch})
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, accepted)
}

func TestHuhBoundaryPrompter_AskMissingContext_CancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	prompter := infrastructure.NewHuhBoundaryPrompter()

	result, err := prompter.AskMissingContext(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, result)
}
