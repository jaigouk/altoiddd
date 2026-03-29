package infrastructure_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Compile-time interface satisfaction check.
var _ application.StorytellingPrompter = (*infrastructure.StdinStorytellingPrompter)(nil)

// validTestStory creates a valid DomainStory with an actor, work object, and sentence.
func validTestStory(t *testing.T) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		"Test Story",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"User clicks button",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("User", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	return story
}

// --- SelectMode Tests ---

func TestSelectModeRapid(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("1\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	mode, err := p.SelectMode(context.Background())

	require.NoError(t, err)
	assert.Equal(t, discoverydomain.ModeRapid, mode)
}

func TestSelectModeThorough(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("2\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	mode, err := p.SelectMode(context.Background())

	require.NoError(t, err)
	assert.Equal(t, discoverydomain.ModeThorough, mode)
}

func TestSelectModeInvalid(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("3\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.SelectMode(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestSelectModeEOF(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.SelectMode(context.Background())

	assert.ErrorIs(t, err, context.Canceled)
}

func TestSelectModeDisplaysOptions(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("1\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.SelectMode(context.Background())

	require.NoError(t, err)
	output := strings.ToLower(writer.String())
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "rapid")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "thorough")
}

// --- AskNarration Tests ---

func TestAskNarrationReturnsInput(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("my answer\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	answer, err := p.AskNarration(context.Background(), "Tell me about actors", "context info")

	require.NoError(t, err)
	assert.Equal(t, "my answer", answer)
}

func TestAskNarrationEmptyInput(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	answer, err := p.AskNarration(context.Background(), "question", "context")

	require.NoError(t, err)
	assert.Empty(t, answer)
}

func TestAskNarrationDisplaysContext(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("answer\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.AskNarration(context.Background(), "Who are the actors?", "Think about roles")

	require.NoError(t, err)
	assert.Contains(t, writer.String(), "Who are the actors?")
	assert.Contains(t, writer.String(), "Think about roles")
}

func TestAskNarrationEOF(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.AskNarration(context.Background(), "question", "context")

	assert.ErrorIs(t, err, context.Canceled)
}

// --- ConfirmSentence Tests ---

func TestConfirmSentenceAccept(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	reader := bytes.NewBufferString("a\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, accepted, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	require.NoError(t, confirmErr)
	assert.True(t, accepted)
	assert.Equal(t, sentence.Subject(), result.Subject())
	assert.Equal(t, sentence.Activity(), result.Activity())
	assert.Equal(t, sentence.Object(), result.Object())
}

func TestConfirmSentenceReject(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	reader := bytes.NewBufferString("r\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, accepted, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	require.NoError(t, confirmErr)
	assert.False(t, accepted)
}

func TestConfirmSentenceEdit(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	reader := bytes.NewBufferString("e\nAdmin\nreviews\nReport\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, accepted, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	require.NoError(t, confirmErr)
	assert.True(t, accepted)
	assert.Equal(t, "Admin", result.Subject())
	assert.Equal(t, "reviews", result.Activity())
	assert.Equal(t, "Report", result.Object())
	assert.Equal(t, vo.UserStated, result.Trust())
}

func TestConfirmSentenceEditKeepFields(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	reader := bytes.NewBufferString("e\n\n\n\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, accepted, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	require.NoError(t, confirmErr)
	assert.True(t, accepted)
	assert.Equal(t, "User", result.Subject())
	assert.Equal(t, "creates", result.Activity())
	assert.Equal(t, "Order", result.Object())
}

func TestConfirmSentenceEditWithPreposition(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	sentence, err = sentence.WithPreposition("for", "Manager")
	require.NoError(t, err)

	reader := bytes.NewBufferString("e\nAdmin\nreviews\nReport\nto\nDirector\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, accepted, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	require.NoError(t, confirmErr)
	assert.True(t, accepted)
	assert.Equal(t, "Admin", result.Subject())
	assert.Equal(t, "reviews", result.Activity())
	assert.Equal(t, "Report", result.Object())
	assert.Equal(t, "to", result.Preposition())
	assert.Equal(t, "Director", result.IndirectObject())
}

func TestConfirmSentenceInvalid(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	reader := bytes.NewBufferString("x\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, _, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	require.Error(t, confirmErr)
	assert.Contains(t, confirmErr.Error(), "invalid")
}

func TestConfirmSentenceEOF(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, _, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	assert.ErrorIs(t, confirmErr, context.Canceled)
}

func TestConfirmSentenceDisplaysSentence(t *testing.T) {
	t.Parallel()

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	reader := bytes.NewBufferString("a\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, _, confirmErr := p.ConfirmSentence(context.Background(), sentence)

	require.NoError(t, confirmErr)
	assert.Contains(t, writer.String(), sentence.FormatText())
}

// --- AskChoice Tests ---

func TestAskChoiceReturnsKey(t *testing.T) {
	t.Parallel()

	options := []application.Choice{
		{Key: "a", Label: "Option A", Description: "First"},
		{Key: "b", Label: "Option B", Description: "Second"},
	}

	reader := bytes.NewBufferString("a\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	choice, err := p.AskChoice(context.Background(), "Pick one", options, "")

	require.NoError(t, err)
	assert.Equal(t, "a", choice)
}

func TestAskChoiceInvalid(t *testing.T) {
	t.Parallel()

	options := []application.Choice{
		{Key: "a", Label: "Option A", Description: "First"},
		{Key: "b", Label: "Option B", Description: "Second"},
	}

	reader := bytes.NewBufferString("c\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.AskChoice(context.Background(), "Pick one", options, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAskChoiceRecommended(t *testing.T) {
	t.Parallel()

	options := []application.Choice{
		{Key: "a", Label: "Option A", Description: "First"},
		{Key: "b", Label: "Option B", Description: "Second"},
	}

	reader := bytes.NewBufferString("b\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	choice, err := p.AskChoice(context.Background(), "Pick one", options, "b")

	require.NoError(t, err)
	assert.Equal(t, "b", choice)
	assert.Contains(t, strings.ToLower(writer.String()), "recommended")
}

func TestAskChoiceEOF(t *testing.T) {
	t.Parallel()

	options := []application.Choice{
		{Key: "a", Label: "Option A", Description: "First"},
	}

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.AskChoice(context.Background(), "Pick one", options, "")

	assert.ErrorIs(t, err, context.Canceled)
}

// --- DisplayStory Tests ---

func TestDisplayStoryPrintsText(t *testing.T) {
	t.Parallel()

	story := validTestStory(t)

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	err := p.DisplayStory(context.Background(), story)

	require.NoError(t, err)
	assert.Contains(t, writer.String(), story.FormatText())
}

func TestDisplayStoryNilError(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	err := p.DisplayStory(context.Background(), nil)

	require.Error(t, err)
}

// --- SynthesisCheckpoint Tests ---

func TestSynthesisCheckpointYes(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("y\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, err := p.SynthesisCheckpoint(context.Background(), application.SynthesisSummary{})

	require.NoError(t, err)
	assert.True(t, result)
}

func TestSynthesisCheckpointNo(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("n\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, err := p.SynthesisCheckpoint(context.Background(), application.SynthesisSummary{})

	require.NoError(t, err)
	assert.False(t, result)
}

func TestSynthesisCheckpointYesVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"Yes\n", true},
		{"n\n", false},
		{"no\n", false},
		{"anything\n", false},
	}

	for _, tc := range tests {
		reader := bytes.NewBufferString(tc.input)
		writer := &bytes.Buffer{}
		p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

		result, err := p.SynthesisCheckpoint(context.Background(), application.SynthesisSummary{})

		require.NoError(t, err)
		assert.Equal(t, tc.expected, result, "input: %q", tc.input)
	}
}

func TestSynthesisCheckpointDisplaysSummary(t *testing.T) {
	t.Parallel()

	story := validTestStory(t)

	summary := application.SynthesisSummary{
		StoriesSoFar:    []*discoverydomain.DomainStory{story},
		ActorInventory:  []discoverydomain.StoryActor{},
		ObjectInventory: []discoverydomain.WorkObject{},
		BoundarySignals: []discoverydomain.BoundarySignal{},
		GlossaryTerms:   []string{"Order", "User"},
	}

	reader := bytes.NewBufferString("y\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.SynthesisCheckpoint(context.Background(), summary)

	require.NoError(t, err)
	output := writer.String()
	assert.Contains(t, output, "1") // 1 story
	assert.Contains(t, output, "2") // 2 glossary terms
}

func TestSynthesisCheckpointEOF(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.SynthesisCheckpoint(context.Background(), application.SynthesisSummary{})

	assert.ErrorIs(t, err, context.Canceled)
}

// --- AskAnnotation Tests ---

func TestAskAnnotationNo(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("n\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	text, num, err := p.AskAnnotation(context.Background())

	require.NoError(t, err)
	assert.Empty(t, text)
	assert.Equal(t, 0, num)
}

func TestAskAnnotationYesTextAndNum(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("y\nrule line1\nrule line2\n\n3\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	text, num, err := p.AskAnnotation(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "rule line1\nrule line2", text)
	assert.Equal(t, 3, num)
}

func TestAskAnnotationDefaultNum(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("y\nsome rule\n\n\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	text, num, err := p.AskAnnotation(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, text)
	assert.Equal(t, 0, num)
}

func TestAskAnnotationInvalidNum(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("y\nsome rule\n\nabc\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, _, err := p.AskAnnotation(context.Background())

	require.Error(t, err)
}

func TestAskAnnotationEOF(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, _, err := p.AskAnnotation(context.Background())

	assert.ErrorIs(t, err, context.Canceled)
}

// --- ProposeStory Tests ---

func TestProposeStoryAccept(t *testing.T) {
	t.Parallel()

	story := validTestStory(t)

	reader := bytes.NewBufferString("y\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, err := p.ProposeStory(context.Background(), story)

	require.NoError(t, err)
	assert.Equal(t, story, result)
}

func TestProposeStoryReject(t *testing.T) {
	t.Parallel()

	story := validTestStory(t)

	reader := bytes.NewBufferString("n\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, err := p.ProposeStory(context.Background(), story)

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestProposeStoryNilError(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, err := p.ProposeStory(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestProposeStoryEOF(t *testing.T) {
	t.Parallel()

	story := validTestStory(t)

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.ProposeStory(context.Background(), story)

	assert.ErrorIs(t, err, context.Canceled)
}

func TestProposeStoryDisplaysText(t *testing.T) {
	t.Parallel()

	story := validTestStory(t)

	reader := bytes.NewBufferString("y\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	_, err := p.ProposeStory(context.Background(), story)

	require.NoError(t, err)
	assert.Contains(t, writer.String(), story.FormatText())
}

func TestProposeStoryEdit(t *testing.T) {
	t.Parallel()

	story := validTestStory(t)

	reader := bytes.NewBufferString("edit\nNew Title\nNew Trigger\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(reader, writer)

	result, err := p.ProposeStory(context.Background(), story)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "New Title", result.Title())
	assert.Equal(t, "New Trigger", result.Trigger())
}

// --- ScannerError Tests ---

func TestScannerErrorWrapped(t *testing.T) {
	t.Parallel()

	readErr := errors.New("broken pipe")
	input := &errorReader{err: readErr}
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinStorytellingPrompter(input, writer)

	_, err := p.SelectMode(context.Background())

	require.Error(t, err)
	require.NotErrorIs(t, err, context.Canceled, "should not be context.Canceled for I/O error")
	require.ErrorIs(t, err, readErr, "should wrap the original error")
}
