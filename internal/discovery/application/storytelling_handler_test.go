package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// --- Fakes ---

// fakeStorytellingPrompter implements StorytellingPrompter with scripted responses.
type fakeStorytellingPrompter struct {
	narrationResponses []string
	narrationQuestions []string
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

func (f *fakeStorytellingPrompter) AskNarration(_ context.Context, question string, _ string) (string, error) {
	f.narrationQuestions = append(f.narrationQuestions, question)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	assert.Contains(t, writer.writtenPaths[0], "alto-scaffold/stories/")
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
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
	handler := NewStorytellingHandler(writer, prompter, nil)
	session := newTestSession()
	flow := newTestFlow()

	_, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)
	refs := session.StoryRefs()
	assert.Len(t, refs, 1)
	assert.Contains(t, refs[0], "alto-scaffold/stories/")
	assert.Contains(t, refs[0], fmt.Sprintf("%02d-", 1))
}

// ===========================================================================
// Integration tests — ProposeResearchStories wiring (alty-cli-1wu.19)
// ===========================================================================

// proposeTrackingPrompter wraps fakeStorytellingPrompter and tracks ProposeStory calls.
type proposeTrackingPrompter struct {
	fakeStorytellingPrompter
	proposeCallCount int
	proposedStories  []*discoverydomain.DomainStory
}

func (p *proposeTrackingPrompter) ProposeStory(ctx context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	p.proposeCallCount++
	p.proposedStories = append(p.proposedStories, proposed)

	return proposed, nil
}

// Compile-time check.
var _ StorytellingPrompter = (*proposeTrackingPrompter)(nil)

// mustResearchResult creates a DomainResearchResult that meets quality floor.
func mustResearchResult(t *testing.T) *discoverydomain.DomainResearchResult {
	t.Helper()

	actors := []discoverydomain.ResearchedActor{
		mustActor(t, "Customer"),
		mustActor(t, "Clerk"),
		mustActor(t, "Manager"),
	}
	entities := []discoverydomain.ResearchedEntity{
		mustEntity(t, "Order"),
		mustEntity(t, "Invoice"),
		mustEntity(t, "Receipt"),
	}
	steps := []discoverydomain.WorkflowStep{
		mustStep(t, 1, "Customer", "places", "Order"),
		mustStep(t, 2, "Clerk", "reviews", "Order"),
		mustStep(t, 3, "Clerk", "generates", "Invoice"),
		mustStep(t, 4, "Customer", "pays", "Invoice"),
		mustStep(t, 5, "Clerk", "issues", "Receipt"),
	}
	wf, err := discoverydomain.NewResearchedWorkflow(
		"Place Order", discoverydomain.WorkflowTypeHappyPath, steps,
		[]string{"https://example.com/wf1"},
	)
	require.NoError(t, err)

	meta := discoverydomain.NewSearchMetadata([]string{"q1", "q2"}, 10, 5, time.Second)

	result, err := discoverydomain.NewDomainResearchResult(
		"retail", meta, actors, entities,
		[]discoverydomain.ResearchedWorkflow{wf},
		nil, nil, nil,
	)
	require.NoError(t, err)

	return &result
}

// mustBelowFloorResearchResult creates a DomainResearchResult below quality floor.
func mustBelowFloorResearchResult(t *testing.T) *discoverydomain.DomainResearchResult {
	t.Helper()

	actors := []discoverydomain.ResearchedActor{mustActor(t, "User")}
	entities := []discoverydomain.ResearchedEntity{mustEntity(t, "Form")}
	steps := []discoverydomain.WorkflowStep{mustStep(t, 1, "User", "fills", "Form")}

	wf, err := discoverydomain.NewResearchedWorkflow(
		"Submit", discoverydomain.WorkflowTypeHappyPath, steps, nil,
	)
	require.NoError(t, err)

	meta := discoverydomain.NewSearchMetadata(nil, 1, 1, time.Second)

	result, err := discoverydomain.NewDomainResearchResult(
		"test", meta, actors, entities,
		[]discoverydomain.ResearchedWorkflow{wf},
		nil, nil, nil,
	)
	require.NoError(t, err)

	return &result
}

func mustActor(t *testing.T, name string) discoverydomain.ResearchedActor {
	t.Helper()

	a, err := discoverydomain.NewResearchedActor(name, "role", []string{"https://example.com"})
	require.NoError(t, err)

	return a
}

func mustEntity(t *testing.T, name string) discoverydomain.ResearchedEntity {
	t.Helper()

	e, err := discoverydomain.NewResearchedEntity(name, nil, []string{"https://example.com"})
	require.NoError(t, err)

	return e
}

func mustStep(t *testing.T, seq int, actor, activity, workObject string) discoverydomain.WorkflowStep {
	t.Helper()

	s, err := discoverydomain.NewWorkflowStep(seq, actor, activity, workObject)
	require.NoError(t, err)

	return s
}

func TestStorytellingHandler_Integration_NilTransformer_WorksAsBeforeNoChanges(t *testing.T) {
	t.Parallel()

	// Given: handler constructed with nil transformer (the no-op path).
	prompter := &proposeTrackingPrompter{
		fakeStorytellingPrompter: fakeStorytellingPrompter{
			narrationResponses:      oneSentenceNarrationResponses(),
			confirmSentenceAccepted: []bool{true},
			synthesisConfirmed:      []bool{true},
		},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter, nil)
	session := newTestSession()
	flow := newTestFlow()

	// When: RunStory is called (same as before the wiring change).
	story, narrative, err := handler.RunStory(context.Background(), session, 1, flow)

	// Then: everything works exactly as before.
	require.NoError(t, err)
	require.NotNil(t, story)
	assert.Equal(t, "Customer places order", story.Trigger())
	assert.Len(t, story.Sentences(), 1)
	assert.Positive(t, narrative.TurnCount())
	assert.Len(t, writer.writtenPaths, 1)

	// ProposeStory was never called via ProposeResearchStories path.
	assert.Equal(t, 0, prompter.proposeCallCount)
}

func TestStorytellingHandler_Integration_ProposeStoryCalledWhenTransformerReturnsStories(t *testing.T) {
	t.Parallel()

	// Given: handler with a real transformer and a quality-floor-passing research result.
	prompter := &proposeTrackingPrompter{}
	writer := &fakeStoryWriter{}
	transformer := NewResearchToStoryTransformer()
	handler := NewStorytellingHandler(writer, prompter, transformer)
	result := mustResearchResult(t)

	// When: ProposeResearchStories is called.
	stories, err := handler.ProposeResearchStories(context.Background(), result)

	// Then: ProposeStory was called for each transformed story.
	require.NoError(t, err)
	require.NotEmpty(t, stories)
	assert.Equal(t, len(stories), prompter.proposeCallCount)
	assert.Len(t, prompter.proposedStories, len(stories))

	// Each proposed story should have sentences (from the workflow steps).
	for _, story := range stories {
		assert.NotEmpty(t, story.Sentences())
	}
}

func TestStorytellingHandler_Integration_ProposeStoryNotCalledWhenTransformerNil(t *testing.T) {
	t.Parallel()

	// Given: handler with nil transformer.
	prompter := &proposeTrackingPrompter{}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter, nil)
	result := mustResearchResult(t)

	// When: ProposeResearchStories is called.
	stories, err := handler.ProposeResearchStories(context.Background(), result)

	// Then: nil, nil returned — ProposeStory never called.
	require.NoError(t, err)
	assert.Nil(t, stories)
	assert.Equal(t, 0, prompter.proposeCallCount)
}

func TestStorytellingHandler_Integration_ProposeStoryNotCalledWhenTransformReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Given: handler with a real transformer but a below-floor research result.
	prompter := &proposeTrackingPrompter{}
	writer := &fakeStoryWriter{}
	transformer := NewResearchToStoryTransformer()
	handler := NewStorytellingHandler(writer, prompter, transformer)
	result := mustBelowFloorResearchResult(t)

	// When: ProposeResearchStories is called.
	stories, err := handler.ProposeResearchStories(context.Background(), result)

	// Then: empty result — Transform returned nil, so ProposeStory never called.
	require.NoError(t, err)
	assert.Empty(t, stories)
	assert.Equal(t, 0, prompter.proposeCallCount)
}

// ===========================================================================
// Bug fix tests — {last_actor} interpolation (alty-cli-9xm)
// ===========================================================================

func TestNarrationPhase_InterpolatesLastActor(t *testing.T) {
	t.Parallel()

	// Given: a storytelling handler with a mock prompter that records questions.
	// The narration responses follow the one-sentence pattern where the actor is "Customer".
	prompter := &fakeStorytellingPrompter{
		narrationResponses:      oneSentenceNarrationResponses(),
		confirmSentenceAccepted: []bool{true},
		synthesisConfirmed:      []bool{true},
	}
	writer := &fakeStoryWriter{}
	handler := NewStorytellingHandler(writer, prompter, nil)
	session := newTestSession()
	flow := newTestFlow()

	// When: RunStory is called (the narration phase asks MQ-N3 "who performs this step?")
	_, _, err := handler.RunStory(context.Background(), session, 1, flow)
	require.NoError(t, err)

	// Then: the MQ-N3 subject question should contain the actual actor name "Customer",
	// not the literal placeholder "{last_actor}".
	var subjectQuestions []string
	for _, q := range prompter.narrationQuestions {
		// MQ-N3 questions contain "Who performs" or "Who does"
		if strings.Contains(q, "Who performs") || strings.Contains(q, "Who does") {
			subjectQuestions = append(subjectQuestions, q)
		}
	}

	require.NotEmpty(t, subjectQuestions, "expected at least one MQ-N3 subject question")
	for _, q := range subjectQuestions {
		assert.NotContains(t, q, "{last_actor}", "subject question should not contain literal {last_actor}")
		assert.Contains(t, q, "Customer", "subject question should contain the interpolated actor name")
	}
}
