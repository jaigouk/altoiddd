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

// Compile-time interface satisfaction check.
var _ application.StorytellingPrompter = (*infrastructure.AgentStorytellingPrompter)(nil)

func TestAgentStorytellingPrompter_SelectMode_ReturnsPreconfiguredMode(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, []string{"line1"})

	mode, err := prompter.SelectMode(context.Background())
	require.NoError(t, err)
	assert.Equal(t, discoverydomain.ModeRapid, mode)
}

func TestAgentStorytellingPrompter_AskNarration_ReturnsLinesSequentially(t *testing.T) {
	t.Parallel()

	lines := []string{"first line", "second line", "third line"}
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, lines)

	r1, err := prompter.AskNarration(context.Background(), "q1", "ctx1")
	require.NoError(t, err)
	assert.Equal(t, "first line", r1)

	r2, err := prompter.AskNarration(context.Background(), "q2", "ctx2")
	require.NoError(t, err)
	assert.Equal(t, "second line", r2)

	r3, err := prompter.AskNarration(context.Background(), "q3", "ctx3")
	require.NoError(t, err)
	assert.Equal(t, "third line", r3)
}

func TestAgentStorytellingPrompter_AskNarration_ReturnsEmptyWhenExhausted(t *testing.T) {
	t.Parallel()

	lines := []string{"only line"}
	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, lines)

	r1, err := prompter.AskNarration(context.Background(), "q1", "ctx1")
	require.NoError(t, err)
	assert.Equal(t, "only line", r1)

	r2, err := prompter.AskNarration(context.Background(), "q2", "ctx2")
	require.NoError(t, err)
	assert.Empty(t, r2)
}

func TestAgentStorytellingPrompter_ConfirmSentence_AutoAccepts(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)

	sentence, err := discoverydomain.NewStorySentence(1, "User", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	result, accepted, confirmErr := prompter.ConfirmSentence(context.Background(), sentence)
	require.NoError(t, confirmErr)
	assert.True(t, accepted)
	assert.Equal(t, sentence.Step(), result.Step())
	assert.Equal(t, sentence.Subject(), result.Subject())
	assert.Equal(t, sentence.Activity(), result.Activity())
	assert.Equal(t, sentence.Object(), result.Object())
}

func TestAgentStorytellingPrompter_AskChoice_ReturnsFirstChoice(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)

	options := []application.Choice{
		{Key: "1", Label: "Option A", Description: "First"},
		{Key: "2", Label: "Option B", Description: "Second"},
	}

	choice, err := prompter.AskChoice(context.Background(), "Pick one", options, "")
	require.NoError(t, err)
	assert.Equal(t, "1", choice)
}

func TestAgentStorytellingPrompter_AskAnnotation_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)

	text, sentenceNum, err := prompter.AskAnnotation(context.Background())
	require.NoError(t, err)
	assert.Empty(t, text)
	assert.Equal(t, 0, sentenceNum)
}

func TestAgentStorytellingPrompter_SynthesisCheckpoint_ReturnsTrue(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)

	result, err := prompter.SynthesisCheckpoint(context.Background(), application.SynthesisSummary{})
	require.NoError(t, err)
	assert.True(t, result)
}

func TestAgentStorytellingPrompter_DisplayStory_IsNoop(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)

	story, err := discoverydomain.NewDomainStory(
		"Test Story",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"trigger",
	)
	require.NoError(t, err)

	err = prompter.DisplayStory(context.Background(), story)
	assert.NoError(t, err)
}

func TestAgentStorytellingPrompter_ProposeStory_ReturnsUnchanged(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)

	story, err := discoverydomain.NewDomainStory(
		"Proposed Story",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"trigger",
	)
	require.NoError(t, err)

	result, proposeErr := prompter.ProposeStory(context.Background(), story)
	require.NoError(t, proposeErr)
	assert.Equal(t, story, result)
}

func TestAgentStorytellingPrompter_ProposeStory_NilReturnsError(t *testing.T) {
	t.Parallel()

	prompter := infrastructure.NewAgentStorytellingPrompter(discoverydomain.ModeRapid, nil)

	result, err := prompter.ProposeStory(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, result)
}
