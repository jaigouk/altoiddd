package application

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// --- Fakes ---

// fakeStorytellingPrompter implements StorytellingPrompter with scripted responses.
type fakeStorytellingPrompter struct {
	narrationResponses []string
	narrationIdx       int
	narrationErr       error

	confirmSentenceAccepted []bool
	confirmSentenceIdx      int
	confirmSentenceErr      error

	choiceResponses []string
	choiceIdx       int
	choiceErr       error

	displayStoryErr error
	displayCount    int

	synthesisConfirmed []bool
	synthesisIdx       int
	synthesisErr       error

	annotationResponses []struct {
		text        string
		sentenceNum int
	}
	annotationIdx int
	annotationErr error
}

func (f *fakeStorytellingPrompter) SelectMode(_ context.Context) (discoverydomain.DiscoveryMode, error) {
	return discoverydomain.ModeRapid, nil
}

func (f *fakeStorytellingPrompter) ProposeStory(_ context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	return proposed, nil
}

func (f *fakeStorytellingPrompter) AskNarration(_ context.Context, _ string, _ string) (string, error) {
	if f.narrationErr != nil {
		return "", f.narrationErr
	}

	if f.narrationIdx >= len(f.narrationResponses) {
		return "", nil
	}

	resp := f.narrationResponses[f.narrationIdx]
	f.narrationIdx++

	return resp, nil
}

func (f *fakeStorytellingPrompter) ConfirmSentence(_ context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	if f.confirmSentenceErr != nil {
		return discoverydomain.StorySentence{}, false, f.confirmSentenceErr
	}

	accepted := true
	if f.confirmSentenceIdx < len(f.confirmSentenceAccepted) {
		accepted = f.confirmSentenceAccepted[f.confirmSentenceIdx]
		f.confirmSentenceIdx++
	}

	return sentence, accepted, nil
}

func (f *fakeStorytellingPrompter) AskChoice(_ context.Context, _ string, _ []Choice, _ string) (string, error) {
	if f.choiceErr != nil {
		return "", f.choiceErr
	}

	if f.choiceIdx >= len(f.choiceResponses) {
		return "no", nil
	}

	resp := f.choiceResponses[f.choiceIdx]
	f.choiceIdx++

	return resp, nil
}

func (f *fakeStorytellingPrompter) DisplayStory(_ context.Context, _ *discoverydomain.DomainStory) error {
	f.displayCount++

	return f.displayStoryErr
}

func (f *fakeStorytellingPrompter) SynthesisCheckpoint(_ context.Context, _ SynthesisSummary) (bool, error) {
	if f.synthesisErr != nil {
		return false, f.synthesisErr
	}

	confirmed := true
	if f.synthesisIdx < len(f.synthesisConfirmed) {
		confirmed = f.synthesisConfirmed[f.synthesisIdx]
		f.synthesisIdx++
	}

	return confirmed, nil
}

func (f *fakeStorytellingPrompter) AskAnnotation(_ context.Context) (string, int, error) {
	if f.annotationErr != nil {
		return "", 0, f.annotationErr
	}

	if f.annotationIdx >= len(f.annotationResponses) {
		return "", 0, nil
	}

	resp := f.annotationResponses[f.annotationIdx]
	f.annotationIdx++

	return resp.text, resp.sentenceNum, nil
}

// Compile-time check.
var _ StorytellingPrompter = (*fakeStorytellingPrompter)(nil)

// fakeStoryWriter implements StoryWriter, recording calls.
type fakeStoryWriter struct {
	writtenPaths   []string
	writtenStories []*discoverydomain.DomainStory
	writeErr       error
}

func (f *fakeStoryWriter) Write(_ context.Context, path string, story *discoverydomain.DomainStory) error {
	if f.writeErr != nil {
		return f.writeErr
	}

	f.writtenPaths = append(f.writtenPaths, path)
	f.writtenStories = append(f.writtenStories, story)

	return nil
}

// Compile-time check.
var _ StoryWriter = (*fakeStoryWriter)(nil)

// --- Helpers ---

// newTestSession creates a session in PersonaDetected state for testing.
func newTestSession() *discoverydomain.DiscoverySession {
	session := discoverydomain.NewDiscoverySession("test readme")
	_ = session.DetectPersona("1") // Developer persona → RegisterTechnical

	return session
}

// newTestFlow creates a rapid StorytellingFlow for testing.
func newTestFlow() *discoverydomain.StorytellingFlow {
	flow, _ := discoverydomain.NewStorytellingFlow(discoverydomain.ModeRapid)

	return flow
}

// oneSentenceNarrationResponses returns prompter narration responses for a minimal valid story.
// Order: trigger, actor, "what happens next", subject, work object, "" (done narrating).
func oneSentenceNarrationResponses() []string {
	return []string{
		"Customer places order", // trigger (MQ-O2)
		"Customer",              // actor (MQ-O1)
		"creates an order",      // what happens next (MQ-N2)
		"Customer",              // subject (MQ-N3)
		"Order",                 // work object (MQ-N4)
		"",                      // done narrating (MQ-N2 empty = break)
	}
}

// --- Tests ---

func TestStorytellingHandler_RunStory_HappyPath_OneSentence(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	story, narrative, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	require.NotNil(t, story)
	assert.Equal(t, "Customer places order", story.Trigger())
	assert.Len(t, story.Sentences(), 1)
	assert.Len(t, story.Actors(), 1)
	assert.Len(t, story.WorkObjects(), 1)
	assert.Positive(t, narrative.TurnCount())
	assert.Len(t, writer.writtenPaths, 1)
	assert.Contains(t, writer.writtenPaths[0], ".alto/stories/")
}

func TestStorytellingHandler_RunStory_ThreeSentences_MidStoryCheckpoint(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses: []string{
			"Customer places order", // trigger
			"Customer",              // actor
			"creates an order",      // sentence 1 activity
			"Customer",              // sentence 1 subject
			"Order",                 // sentence 1 work object
			"reviews the order",     // sentence 2 activity
			"Customer",              // sentence 2 subject
			"Order",                 // sentence 2 work object
			"submits the order",     // sentence 3 activity
			"Customer",              // sentence 3 subject
			"Order",                 // sentence 3 work object
			"",                      // done narrating
		},
		confirmSentenceAccepted: []bool{true, true, true},
		synthesisConfirmed:      []bool{true, true}, // mid-story + final
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	story, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	assert.Len(t, story.Sentences(), 3)
}

func TestStorytellingHandler_RunStory_BranchingDetected_VariationAdded(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses: []string{
			"Customer places order",         // trigger
			"Customer",                      // actor
			"sometimes creates an order",    // branching activity
			"Customer",                      // subject
			"Order",                         // work object
			"New variation for branch path", // variation description via AskNarration
			"",                              // done narrating
		},
		confirmSentenceAccepted: []bool{true},
		choiceResponses:         []string{"yes"},
		synthesisConfirmed:      []bool{true},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	story, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	assert.Len(t, story.Variations(), 1)
}

func TestStorytellingHandler_RunStory_BranchingRejected_NothingAdded(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses: []string{
			"Customer places order",      // trigger
			"Customer",                   // actor
			"sometimes creates an order", // branching activity
			"Customer",                   // subject
			"Order",                      // work object
			"",                           // done narrating
		},
		confirmSentenceAccepted: []bool{true},
		choiceResponses:         []string{"no"},
		synthesisConfirmed:      []bool{true},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	story, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	assert.Empty(t, story.Variations())
}

func TestStorytellingHandler_RunStory_AnnotationCollected(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
		annotationResponses: []struct {
			text        string
			sentenceNum int
		}{
			{text: "Order total must be positive", sentenceNum: 1},
			{text: "", sentenceNum: 0}, // done
		},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	story, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	assert.Len(t, story.Annotations(), 1)
}

func TestStorytellingHandler_RunStory_MultipleAnnotations(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
		annotationResponses: []struct {
			text        string
			sentenceNum int
		}{
			{text: "Order total must be positive", sentenceNum: 1},
			{text: "Customer must be verified", sentenceNum: 0},
			{text: "", sentenceNum: 0}, // done
		},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	story, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	assert.Len(t, story.Annotations(), 2)
}

func TestStorytellingHandler_RunStory_NoAnnotations(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
		annotationResponses:     nil, // empty = done immediately
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	story, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	assert.Empty(t, story.Annotations())
}

func TestStorytellingHandler_RunStory_StoryValidationFails_ErrorReturned(t *testing.T) {
	t.Parallel()

	// Build a story with no sentences by making narration return done immediately
	// after opening phase. This produces a story with 0 sentences, which Validate() rejects.
	prompter := &fakeStorytellingPrompter{
		narrationResponses: []string{
			"Customer places order", // trigger
			"Customer",              // actor
			"",                      // done narrating immediately (no sentences)
		},
		synthesisConfirmed: []bool{true},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	_, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.Error(t, err)
	assert.Empty(t, writer.writtenPaths) // no write attempted
}

func TestStorytellingHandler_RunStory_StoryWriterFails_ErrorReturned(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
	}
	writer := &fakeStoryWriter{writeErr: errors.New("disk full")}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	_, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
	// Session should NOT have story ref since write failed
	assert.Empty(t, session.StoryRefs())
}

func TestStorytellingHandler_RunStory_ContextCanceled_MidNarration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	prompter := &fakeStorytellingPrompter{
		narrationErr: ctx.Err(), // context.Canceled
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	_, _, err := handler.RunStory(ctx, session, 1, flow)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestStorytellingHandler_RunStory_NarrativeHasCorrectTurnCount(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	_, narrative, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	// Opening: trigger (1) + actor (1) = 2 turns
	// Narration: activity (1) + subject (1) + work object (1) = 3 turns per sentence = 3
	// Total = 5
	assert.Equal(t, 5, narrative.TurnCount())
}

func TestStorytellingHandler_RunStory_StoryRefAddedToSession(t *testing.T) {
	t.Parallel()

	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter)
	session := newTestSession()
	flow := newTestFlow()

	_, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	refs := session.StoryRefs()
	assert.Len(t, refs, 1)
	assert.Contains(t, refs[0], ".alto/stories/")
	assert.Contains(t, refs[0], fmt.Sprintf("%02d-", 1))
}
