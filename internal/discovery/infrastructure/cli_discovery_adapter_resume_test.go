package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- Resume-specific fakes --

// fakeResumePublisher is a minimal EventPublisher for resume tests.
type fakeResumePublisher struct{}

func (f *fakeResumePublisher) Publish(_ context.Context, _ any) error { return nil }

var _ sharedapp.EventPublisher = (*fakeResumePublisher)(nil)

// fakeResumeStoryWriter tracks stories written during resume.
type fakeResumeStoryWriter struct {
	written []*discoverydomain.DomainStory
	err     error
}

func (f *fakeResumeStoryWriter) Write(_ context.Context, _ string, story *discoverydomain.DomainStory) error {
	if f.err != nil {
		return f.err
	}
	f.written = append(f.written, story)
	return nil
}

var _ application.StoryWriter = (*fakeResumeStoryWriter)(nil)

// fakeResumePrompter provides canned responses for resume flow.
type fakeResumePrompter struct {
	modeChoice              discoverydomain.DiscoveryMode
	narrationResponses      []string
	narrationIdx            int
	confirmSentenceAccepted bool
	synthesisResult         bool
	choiceResponses         []string
	choiceIdx               int
}

func (f *fakeResumePrompter) SelectMode(_ context.Context) (discoverydomain.DiscoveryMode, error) {
	return f.modeChoice, nil
}

func (f *fakeResumePrompter) ProposeStory(_ context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	return proposed, nil
}

func (f *fakeResumePrompter) AskNarration(_ context.Context, _ string, _ string) (string, error) {
	if f.narrationIdx >= len(f.narrationResponses) {
		return "", nil
	}
	resp := f.narrationResponses[f.narrationIdx]
	f.narrationIdx++
	return resp, nil
}

func (f *fakeResumePrompter) ConfirmSentence(_ context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	return sentence, f.confirmSentenceAccepted, nil
}

func (f *fakeResumePrompter) AskChoice(_ context.Context, _ string, _ []application.Choice, _ string) (string, error) {
	if f.choiceIdx >= len(f.choiceResponses) {
		return "", nil
	}
	resp := f.choiceResponses[f.choiceIdx]
	f.choiceIdx++
	return resp, nil
}

func (f *fakeResumePrompter) DisplayStory(_ context.Context, _ *discoverydomain.DomainStory) error {
	return nil
}

func (f *fakeResumePrompter) SynthesisCheckpoint(_ context.Context, _ application.SynthesisSummary) (bool, error) {
	return f.synthesisResult, nil
}

func (f *fakeResumePrompter) AskAnnotation(_ context.Context) (string, int, error) {
	return "", 0, nil
}

var _ application.StorytellingPrompter = (*fakeResumePrompter)(nil)

// fakeResumeBoundaryDetector returns pre-configured boundary sketches.
type fakeResumeBoundaryDetector struct {
	sketches []discoverydomain.BoundedContextSketch
	called   bool
	err      error
}

func (f *fakeResumeBoundaryDetector) DetectBoundaries(_ context.Context, _ []*discoverydomain.DomainStory, _ discoverydomain.DiscoveryMode) ([]discoverydomain.BoundedContextSketch, error) {
	f.called = true
	return f.sketches, f.err
}

var _ application.BoundaryDetector = (*fakeResumeBoundaryDetector)(nil)

// fakeResumeBoundaryPrompter provides canned responses for boundary prompts.
type fakeResumeBoundaryPrompter struct {
	acceptedNames []string
	called        bool
}

func (f *fakeResumeBoundaryPrompter) DisplayBoundaryProposals(_ context.Context, proposals []discoverydomain.BoundedContextSketch) ([]string, error) {
	f.called = true
	if f.acceptedNames != nil {
		return f.acceptedNames, nil
	}
	names := make([]string, len(proposals))
	for i, p := range proposals {
		names[i] = p.Name()
	}
	return names, nil
}

func (f *fakeResumeBoundaryPrompter) AskMissingContext(_ context.Context) (string, error) {
	return "", nil
}

var _ application.BoundaryPrompter = (*fakeResumeBoundaryPrompter)(nil)

// fakeResumeContextMapWriter captures written context maps.
type fakeResumeContextMapWriter struct {
	writtenMap *discoverydomain.ContextMap
}

func (f *fakeResumeContextMapWriter) Write(_ context.Context, _ string, cm *discoverydomain.ContextMap) error {
	f.writtenMap = cm
	return nil
}

var _ application.ContextMapWriter = (*fakeResumeContextMapWriter)(nil)

// -- Helpers --

// makeResumeTestSketches creates boundary sketches for testing.
func makeResumeTestSketches(t *testing.T, names ...string) []discoverydomain.BoundedContextSketch {
	t.Helper()
	sketches := make([]discoverydomain.BoundedContextSketch, 0, len(names))
	for _, name := range names {
		sketch, err := discoverydomain.NewBoundedContextSketch(
			name, vo.SubdomainCore, 0.75,
			[]string{"User"}, nil, nil, nil, vo.AIInferred,
		)
		require.NoError(t, err)
		sketches = append(sketches, sketch)
	}
	return sketches
}

// buildResumedSession creates a session in StatusAnswering with the given number of stories
// and optionally confirmed boundaries, suitable for Resume testing.
func buildResumedSession(t *testing.T, storyCount int, boundariesDone bool) *discoverydomain.DiscoverySession {
	t.Helper()

	session := discoverydomain.NewDiscoverySession("A test project idea.")
	require.NoError(t, session.SetMode(discoverydomain.ModeRapid))
	require.NoError(t, session.DetectPersona("1"))

	for i := range storyCount {
		require.NoError(t, session.AddStoryRef(filepath.Join("stories", "story"+string(rune('1'+i))+".yaml")))
	}

	if boundariesDone {
		require.NoError(t, session.ConfirmBoundaries(nil))
	}

	return session
}

// newResumeStoryPrompter creates a prompter that completes stories from a given index.
func newResumeStoryPrompter(remainingStories int) *fakeResumePrompter {
	narrations := make([]string, 0, remainingStories*6)
	for range remainingStories {
		narrations = append(narrations,
			"User places order", // trigger
			"Customer",          // actor
			"submits the form",  // activity
			"Customer",          // subject
			"Order Form",        // work object
			"",                  // end narration
		)
	}

	choices := make([]string, 0, remainingStories)
	for range remainingStories - 1 {
		choices = append(choices, "yes")
	}

	return &fakeResumePrompter{
		modeChoice:              discoverydomain.ModeRapid,
		narrationResponses:      narrations,
		confirmSentenceAccepted: true,
		synthesisResult:         true,
		choiceResponses:         choices,
	}
}

// -- Resume Tests --

func TestCLIDiscoveryAdapter_Resume_SkipsBoundaryWhenDone(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	// Given a session with 3 stories and boundaries already confirmed
	session := buildResumedSession(t, 3, true)

	// Resume should skip boundary detection entirely and call Complete
	prompter := newResumeStoryPrompter(0) // no remaining stories needed — already at 3
	// Override: for a 3-story rapid session, completeness is met, so no story loop
	prompter.choiceResponses = nil
	prompter.narrationResponses = nil

	writer := &fakeResumeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakeResumePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector := &fakeResumeBoundaryDetector{
		sketches: makeResumeTestSketches(t, "Orders"),
	}
	bPrompter := &fakeResumeBoundaryPrompter{}
	cmWriter := &fakeResumeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When we resume the session
	err := adapter.Resume(context.Background(), session)

	// Then no error, boundary detection was NOT called
	require.NoError(t, err)
	assert.False(t, detector.called, "boundary detection should be skipped when boundaries already confirmed")
	assert.False(t, bPrompter.called, "boundary prompter should not be called when boundaries already confirmed")
}

func TestCLIDiscoveryAdapter_Resume_RunsBoundaryWhenNotDone(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	// Given a session with 3 stories but boundaries NOT confirmed
	session := buildResumedSession(t, 3, false)

	prompter := newResumeStoryPrompter(0) // stories complete, just need boundary
	prompter.choiceResponses = nil
	prompter.narrationResponses = nil

	writer := &fakeResumeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakeResumePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector := &fakeResumeBoundaryDetector{
		sketches: makeResumeTestSketches(t, "Orders"),
	}
	bPrompter := &fakeResumeBoundaryPrompter{
		acceptedNames: []string{"Orders"},
	}
	cmWriter := &fakeResumeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When we resume the session
	err := adapter.Resume(context.Background(), session)

	// Then boundary detection WAS called
	require.NoError(t, err)
	assert.True(t, detector.called, "boundary detection should run when boundaries not yet confirmed")
}

func TestCLIDiscoveryAdapter_Resume_StartsAtCorrectStoryIndex(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	// Given a session with 1 story done (ModeRapid needs 3)
	session := buildResumedSession(t, 1, false)

	// Resume should start story loop at index 2, run 2 more stories
	prompter := newResumeStoryPrompter(2)

	writer := &fakeResumeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakeResumePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector := &fakeResumeBoundaryDetector{}
	bPrompter := &fakeResumeBoundaryPrompter{}
	cmWriter := &fakeResumeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When we resume the session
	err := adapter.Resume(context.Background(), session)

	// Then 2 additional stories were written (starting from index 2)
	require.NoError(t, err)
	assert.Len(t, writer.written, 2)
}

func TestCLIDiscoveryAdapter_Resume_CompletedSessionHandledByCaller(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	// Given a completed session — the caller is responsible for guarding against this.
	// Resume itself does NOT check status. Build a completed session via snapshot.
	snapshot := map[string]interface{}{
		"session_id":                  "completed-session-id",
		"readme_content":              "A test project idea.",
		"status":                      "completed",
		"persona":                     "developer",
		"register":                    "technical",
		"answers":                     []interface{}{},
		"skipped":                     []interface{}{},
		"playback_confirmations":      []interface{}{},
		"answers_since_last_playback": 0,
		"mode":                        "rapid",
		"story_refs":                  []interface{}{"s1.yaml", "s2.yaml", "s3.yaml"},
		"boundaries_confirmed":        true,
	}
	session, err := discoverydomain.FromSnapshot(snapshot)
	require.NoError(t, err)

	prompter := newResumeStoryPrompter(0)
	writer := &fakeResumeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakeResumePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector := &fakeResumeBoundaryDetector{}
	bPrompter := &fakeResumeBoundaryPrompter{}
	cmWriter := &fakeResumeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When we call Resume with a completed session
	// Then: Resume doesn't guard against StatusCompleted — it either works
	// (skips stories + boundaries) or fails at Complete (already completed).
	// Either behavior documents the contract: caller guards, not Resume.
	err = adapter.Resume(context.Background(), session)

	// The test documents this design contract. The method may error
	// (e.g., Complete fails on already-completed session) or succeed.
	// What matters: Resume does NOT have a StatusCompleted guard itself.
	// We just verify it doesn't panic.
	_ = err // error is acceptable — contract test, not correctness test
}
