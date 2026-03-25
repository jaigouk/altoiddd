package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Compile-time interface satisfaction check.
var _ application.StorytellingPrompter = (*infrastructure.HuhStorytellingPrompter)(nil)

func TestHuhStorytellingPrompter_New(t *testing.T) {
	t.Parallel()
	prompter := infrastructure.NewHuhStorytellingPrompter()
	assert.NotNil(t, prompter)
}

func TestHuhStorytellingPrompter_CompileTimeCheck(t *testing.T) {
	t.Parallel()
	// The var _ line above ensures compile-time check. This test confirms
	// the constructor returns a value assignable to the interface.
	var p application.StorytellingPrompter = infrastructure.NewHuhStorytellingPrompter()
	assert.NotNil(t, p)
}

func TestHuhStorytellingPrompter_ProposeStory_ReturnsNotImplemented(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	story := buildTestStory(t)

	result, err := p.ProposeStory(context.Background(), story)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, errors.ErrUnsupported)
}

func TestHuhStorytellingPrompter_ProposeStory_NilStory_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()

	result, err := p.ProposeStory(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestHuhStorytellingPrompter_DisplayStory_NilStory_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()

	err := p.DisplayStory(context.Background(), nil)
	require.Error(t, err)
}

func TestHuhStorytellingPrompter_DisplayStory_CancelledCtx_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.DisplayStory(ctx, buildTestStory(t))
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHuhStorytellingPrompter_SelectMode_CancelledCtx_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.SelectMode(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHuhStorytellingPrompter_AskNarration_CancelledCtx_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.AskNarration(ctx, "question", "some context")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHuhStorytellingPrompter_ConfirmSentence_CancelledCtx_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sentence := buildTestSentence(t)

	_, _, err := p.ConfirmSentence(ctx, sentence)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHuhStorytellingPrompter_AskChoice_CancelledCtx_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	choices := []application.Choice{
		{Key: "a", Label: "Option A", Description: "First option"},
	}

	_, err := p.AskChoice(ctx, "Pick one", choices, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHuhStorytellingPrompter_SynthesisCheckpoint_CancelledCtx_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	summary := application.SynthesisSummary{
		GlossaryTerms: []string{"term1"},
	}

	_, err := p.SynthesisCheckpoint(ctx, summary)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHuhStorytellingPrompter_AskAnnotation_CancelledCtx_ReturnsError(t *testing.T) {
	t.Parallel()
	p := infrastructure.NewHuhStorytellingPrompter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := p.AskAnnotation(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// buildTestStory creates a minimal valid DomainStory for testing.
func buildTestStory(t *testing.T) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		"Test Story",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"User requests something",
	)
	require.NoError(t, err)

	return story
}

// buildTestSentence creates a minimal valid StorySentence for testing.
func buildTestSentence(t *testing.T) discoverydomain.StorySentence {
	t.Helper()

	sentence, err := discoverydomain.NewStorySentence(
		1, "User", "submits", "Order",
		vo.UserStated, "",
	)
	require.NoError(t, err)

	return sentence
}
