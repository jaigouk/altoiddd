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

// --- Fake StorytellingPrompter for Testing ---

type fakeStorytellingPrompter struct {
	// SelectMode
	modeChoice discoverydomain.DiscoveryMode
	modeErr    error

	// AskChoice
	choiceResponses []string // sequential responses for AskChoice calls
	choiceIdx       int
	choiceErr       error

	// AskNarration
	narrationResponses []string // sequential narration answers
	narrationIdx       int
	narrationErr       error

	// ConfirmSentence
	confirmSentenceAccepted bool
	confirmSentenceErr      error

	// DisplayStory
	displayStoryErr    error
	displayStoryCalled int

	// SynthesisCheckpoint
	synthesisResult bool
	synthesisErr    error

	// ProposeStory
	proposeErr error

	// AskAnnotation — returns empty to signal "done" immediately
	annotationErr error
}

func (f *fakeStorytellingPrompter) SelectMode(_ context.Context) (discoverydomain.DiscoveryMode, error) {
	return f.modeChoice, f.modeErr
}

func (f *fakeStorytellingPrompter) ProposeStory(_ context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error) {
	if f.proposeErr != nil {
		return nil, f.proposeErr
	}
	return proposed, nil
}

func (f *fakeStorytellingPrompter) AskNarration(_ context.Context, _ string, _ string) (string, error) {
	if f.narrationErr != nil {
		return "", f.narrationErr
	}
	if f.narrationIdx >= len(f.narrationResponses) {
		return "", nil // empty = end narration
	}
	resp := f.narrationResponses[f.narrationIdx]
	f.narrationIdx++
	return resp, nil
}

func (f *fakeStorytellingPrompter) ConfirmSentence(_ context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error) {
	if f.confirmSentenceErr != nil {
		return discoverydomain.StorySentence{}, false, f.confirmSentenceErr
	}
	return sentence, f.confirmSentenceAccepted, nil
}

func (f *fakeStorytellingPrompter) AskChoice(_ context.Context, _ string, _ []application.Choice, _ string) (string, error) {
	if f.choiceErr != nil {
		return "", f.choiceErr
	}
	if f.choiceIdx >= len(f.choiceResponses) {
		return "", nil
	}
	resp := f.choiceResponses[f.choiceIdx]
	f.choiceIdx++
	return resp, nil
}

func (f *fakeStorytellingPrompter) DisplayStory(_ context.Context, _ *discoverydomain.DomainStory) error {
	f.displayStoryCalled++
	return f.displayStoryErr
}

func (f *fakeStorytellingPrompter) SynthesisCheckpoint(_ context.Context, _ application.SynthesisSummary) (bool, error) {
	return f.synthesisResult, f.synthesisErr
}

func (f *fakeStorytellingPrompter) AskAnnotation(_ context.Context) (string, int, error) {
	return "", 0, f.annotationErr // empty text = done immediately
}

// Compile-time check.
var _ application.StorytellingPrompter = (*fakeStorytellingPrompter)(nil)

// --- Fake Story Writer ---

type fakeStoryWriter struct {
	written []*discoverydomain.DomainStory
	err     error
}

func (f *fakeStoryWriter) Write(_ context.Context, _ string, story *discoverydomain.DomainStory) error {
	if f.err != nil {
		return f.err
	}
	f.written = append(f.written, story)
	return nil
}

var _ application.StoryWriter = (*fakeStoryWriter)(nil)

// --- Fake Event Publisher ---

type fakePublisher struct{}

func (f *fakePublisher) Publish(_ context.Context, _ any) error { return nil }

var _ sharedapp.EventPublisher = (*fakePublisher)(nil)

// --- Helper: create a minimal fakeStorytellingPrompter for a single rapid-mode story ---

// newMinimalStoryPrompter creates a prompter that completes one story:
// - Narration: trigger response, actor response, then activity/subject/object for one sentence, then empty to end
// - Sentences are accepted
// - Annotations: done immediately
// - Synthesis: confirmed
// - AskChoice responses: persona choice "1", then "no" (stop after first story... but we need 3 for rapid)
func newRapidStoryPrompter(storyCount int) *fakeStorytellingPrompter {
	// Each RunStory call needs narration responses:
	// 1. trigger response (opening: MQ-O2)
	// 2. actor response (opening: MQ-O1)
	// 3. activity response (narration: MQ-N2) — one sentence
	// 4. subject response (narration: MQ-N3) — for that sentence
	// 5. object response (narration: MQ-N4) — for that sentence
	// 6. empty (narration: MQ-N2 = end)
	narrations := make([]string, 0, storyCount*6)
	for i := range storyCount {
		narrations = append(narrations,
			"User places order", // trigger
			"Customer",          // actor
			"submits the form",  // activity
			"Customer",          // subject (defaults to lastActor if empty)
			"Order Form",        // work object
			"",                  // end narration
		)
		_ = i
	}

	// AskChoice responses:
	// 1st call = persona selection ("1" = Developer)
	// Then after each story (except last): "yes" to continue
	// After last story: doesn't matter, loop breaks on CheckStoryCompleteness
	choices := []string{"1"} // persona
	for range storyCount - 1 {
		choices = append(choices, "yes") // continue after each non-final story
	}

	return &fakeStorytellingPrompter{
		modeChoice:              discoverydomain.ModeRapid,
		narrationResponses:      narrations,
		confirmSentenceAccepted: true,
		synthesisResult:         true,
		choiceResponses:         choices,
	}
}

// --- Tests ---

func TestCLIDiscoveryAdapter_Run_HappyPath_RapidMode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	prompter := newRapidStoryPrompter(3) // RAPID = 3 stories
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)
	detector, bPrompter, cmWriter := newBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	// Verify 3 stories were written
	assert.Len(t, writer.written, 3)
}

func TestCLIDiscoveryAdapter_Run_ModeCanceled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))

	prompter := &fakeStorytellingPrompter{
		modeErr: context.Canceled,
	}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)
	detector, bPrompter, cmWriter := newBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCLIDiscoveryAdapter_Run_PersonaCanceled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))

	prompter := &fakeStorytellingPrompter{
		modeChoice:      discoverydomain.ModeRapid,
		choiceErr:       context.Canceled, // Cancel on persona AskChoice
		choiceResponses: []string{},
	}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)
	detector, bPrompter, cmWriter := newBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCLIDiscoveryAdapter_Run_MissingREADME(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir() // No README.md

	prompter := &fakeStorytellingPrompter{}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)
	detector, bPrompter, cmWriter := newBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "README")
}

func TestCLIDiscoveryAdapter_Run_CompleteCalled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	prompter := newRapidStoryPrompter(3)
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)
	detector, bPrompter, cmWriter := newBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	// After Run completes, session should be in completed state.
	// We verify indirectly: no error from Run means Complete succeeded.
	assert.Len(t, writer.written, 3)
}

func TestCLIDiscoveryAdapter_Run_MultipleStories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	// Let's test THOROUGH with exactly 5 stories.
	prompter := newRapidStoryPrompter(5)
	prompter.modeChoice = discoverydomain.ModeThorough
	// Fix choices: persona + 4 "yes" to continue
	prompter.choiceResponses = []string{"1", "yes", "yes", "yes", "yes"}

	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)
	detector, bPrompter, cmWriter := newBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	assert.Len(t, writer.written, 5)
}

// --- Boundary Detection Tests ---

func TestCLIDiscoveryAdapter_Run_BoundaryDetection_WithContextMapWritten(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	prompter := newRapidStoryPrompter(3)
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)

	// Detector returns 2 sketches, prompter accepts both, no missing context
	detector := &fakeBoundaryDetector{
		sketches: makeTestSketches(t, "Orders", "Shipping"),
	}
	bPrompter := &fakeBoundaryPrompter{
		acceptedNames: []string{"Orders", "Shipping"},
	}
	cmWriter := &fakeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	// Context map written
	require.NotNil(t, cmWriter.writtenMap)
	assert.Len(t, cmWriter.writtenMap.Contexts(), 2)
	assert.Contains(t, cmWriter.writtenPath, "context-map.yaml")
}

func TestCLIDiscoveryAdapter_Run_BoundaryDetection_WithMissingContext(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	prompter := newRapidStoryPrompter(3)
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)

	detector := &fakeBoundaryDetector{
		sketches: makeTestSketches(t, "Orders"),
	}
	bPrompter := &fakeBoundaryPrompter{
		acceptedNames: []string{"Orders"},
		missingName:   "Payments", // user adds a missing context
	}
	cmWriter := &fakeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	// Context map should have 2 contexts: 1 detected + 1 user-stated
	require.NotNil(t, cmWriter.writtenMap)
	assert.Len(t, cmWriter.writtenMap.Contexts(), 2)
}

func TestCLIDiscoveryAdapter_Run_BoundaryDetection_UserRejectsAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))

	prompter := newRapidStoryPrompter(3)
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)

	detector := &fakeBoundaryDetector{
		sketches: makeTestSketches(t, "Orders"),
	}
	bPrompter := &fakeBoundaryPrompter{
		acceptedNames: []string{}, // reject all
	}
	cmWriter := &fakeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	require.NoError(t, err)

	// Empty context map is valid (single-context domain)
	require.NotNil(t, cmWriter.writtenMap)
	assert.Empty(t, cmWriter.writtenMap.Contexts())
}

func TestCLIDiscoveryAdapter_Run_BoundaryDetection_CancelDuringDisplay(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))

	prompter := newRapidStoryPrompter(3)
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter)

	detector := &fakeBoundaryDetector{
		sketches: makeTestSketches(t, "Orders"),
	}
	bPrompter := &fakeBoundaryPrompter{
		displayErr: context.Canceled,
	}
	cmWriter := &fakeContextMapWriter{}
	bdHandler := application.NewBoundaryDetectionHandler(detector)

	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	err := adapter.Run(context.Background())
	assert.ErrorIs(t, err, context.Canceled)
}

// --- Test Helpers ---

func makeTestSketches(t *testing.T, names ...string) []discoverydomain.BoundedContextSketch {
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

// --- Fake BoundaryDetector ---

type fakeBoundaryDetector struct {
	sketches []discoverydomain.BoundedContextSketch
	err      error
}

func (f *fakeBoundaryDetector) DetectBoundaries(_ context.Context, _ []*discoverydomain.DomainStory, _ discoverydomain.DiscoveryMode) ([]discoverydomain.BoundedContextSketch, error) {
	return f.sketches, f.err
}

var _ application.BoundaryDetector = (*fakeBoundaryDetector)(nil)

// --- Fake BoundaryPrompter ---

type fakeBoundaryPrompter struct {
	// DisplayBoundaryProposals
	acceptedNames []string
	displayErr    error

	// AskMissingContext
	missingName string
	missingErr  error
}

func (f *fakeBoundaryPrompter) DisplayBoundaryProposals(_ context.Context, proposals []discoverydomain.BoundedContextSketch) ([]string, error) {
	if f.displayErr != nil {
		return nil, f.displayErr
	}
	if f.acceptedNames != nil {
		return f.acceptedNames, nil
	}
	// Default: accept all proposals
	names := make([]string, len(proposals))
	for i, p := range proposals {
		names[i] = p.Name()
	}
	return names, nil
}

func (f *fakeBoundaryPrompter) AskMissingContext(_ context.Context) (string, error) {
	return f.missingName, f.missingErr
}

var _ application.BoundaryPrompter = (*fakeBoundaryPrompter)(nil)

// --- Fake ContextMapWriter ---

type fakeContextMapWriter struct {
	writtenPath string
	writtenMap  *discoverydomain.ContextMap
	err         error
}

func (f *fakeContextMapWriter) Write(_ context.Context, path string, cm *discoverydomain.ContextMap) error {
	if f.err != nil {
		return f.err
	}
	f.writtenPath = path
	f.writtenMap = cm
	return nil
}

var _ application.ContextMapWriter = (*fakeContextMapWriter)(nil)

// --- Helper: create standard boundary detection fakes ---

func newBoundaryFakes() (*fakeBoundaryDetector, *fakeBoundaryPrompter, *fakeContextMapWriter) {
	return &fakeBoundaryDetector{}, &fakeBoundaryPrompter{}, &fakeContextMapWriter{}
}

// --- Interface compliance checks ---

func TestFakeStoryWriter_InterfaceCompliance(t *testing.T) {
	t.Parallel()
	var p application.StorytellingPrompter = &fakeStorytellingPrompter{}
	assert.NotNil(t, p)
	var w application.StoryWriter = &fakeStoryWriter{}
	assert.NotNil(t, w)
	var bd application.BoundaryDetector = &fakeBoundaryDetector{}
	assert.NotNil(t, bd)
	var bp application.BoundaryPrompter = &fakeBoundaryPrompter{}
	assert.NotNil(t, bp)
	var cmw application.ContextMapWriter = &fakeContextMapWriter{}
	assert.NotNil(t, cmw)
}
