// Package integration provides BDD-style integration tests.
package integration

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
)

// ---------------------------------------------------------------------------
// Fake StorytellingPrompter for Integration Tests
// ---------------------------------------------------------------------------

type fakeStorytellingPrompter struct {
	// SelectMode
	modeChoice discoverydomain.DiscoveryMode
	modeErr    error

	// AskChoice
	choiceResponses []string
	choiceIdx       int
	choiceErr       error

	// AskNarration
	narrationResponses []string
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

	// AskAnnotation
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
	return "", 0, f.annotationErr
}

var _ application.StorytellingPrompter = (*fakeStorytellingPrompter)(nil)

// ---------------------------------------------------------------------------
// Fake Story Writer
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Fake Event Publisher
// ---------------------------------------------------------------------------

type fakePublisher struct{}

func (f *fakePublisher) Publish(_ context.Context, _ any) error { return nil }

var _ sharedapp.EventPublisher = (*fakePublisher)(nil)

// ---------------------------------------------------------------------------
// Fake Boundary Detection Collaborators
// ---------------------------------------------------------------------------

type fakeBoundaryDetector struct {
	sketches []discoverydomain.BoundedContextSketch
	err      error
}

func (f *fakeBoundaryDetector) DetectBoundaries(_ context.Context, _ []*discoverydomain.DomainStory, _ discoverydomain.DiscoveryMode) ([]discoverydomain.BoundedContextSketch, error) {
	return f.sketches, f.err
}

var _ application.BoundaryDetector = (*fakeBoundaryDetector)(nil)

type fakeBoundaryPrompter struct {
	acceptedNames []string
	displayErr    error
	missingName   string
	missingErr    error
}

func (f *fakeBoundaryPrompter) DisplayBoundaryProposals(_ context.Context, proposals []discoverydomain.BoundedContextSketch) ([]string, error) {
	if f.displayErr != nil {
		return nil, f.displayErr
	}
	if f.acceptedNames != nil {
		return f.acceptedNames, nil
	}
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

func newIntegrationBoundaryFakes() (*fakeBoundaryDetector, *fakeBoundaryPrompter, *fakeContextMapWriter) {
	return &fakeBoundaryDetector{}, &fakeBoundaryPrompter{}, &fakeContextMapWriter{}
}

// ---------------------------------------------------------------------------
// Helper: build a prompter for N rapid-mode stories
// ---------------------------------------------------------------------------

func newIntegrationStoryPrompter(storyCount int) *fakeStorytellingPrompter {
	// Each RunStory call consumes:
	// 1. trigger (opening MQ-O2)
	// 2. actor (opening MQ-O1)
	// 3. activity (narration MQ-N2) — one sentence
	// 4. subject (narration MQ-N3)
	// 5. object (narration MQ-N4)
	// 6. "" (narration MQ-N2 = end narration loop)
	narrations := make([]string, 0, storyCount*6)
	for range storyCount {
		narrations = append(narrations,
			"User places order",
			"Customer",
			"submits the form",
			"Customer",
			"Order Form",
			"",
		)
	}

	// AskChoice: first call is persona ("1"), then "yes" to continue for non-final stories
	choices := []string{"1"}
	for range storyCount - 1 {
		choices = append(choices, "yes")
	}

	return &fakeStorytellingPrompter{
		modeChoice:              discoverydomain.ModeRapid,
		narrationResponses:      narrations,
		confirmSentenceAccepted: true,
		synthesisResult:         true,
		choiceResponses:         choices,
	}
}

// ---------------------------------------------------------------------------
// Integration Tests — Storytelling Flow
// ---------------------------------------------------------------------------

// TestCLIDiscovery_HappyPath_StorytellingFlow verifies the full cross-layer wiring:
// handler ↔ adapter ↔ prompter ↔ storytellingHandler all connected correctly.
func TestCLIDiscovery_HappyPath_StorytellingFlow(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea for testing"), 0o644))

	// And: a prompter configured for 3 rapid-mode stories
	prompter := newIntegrationStoryPrompter(3)
	writer := &fakeStoryWriter{}

	// And: fully wired handler + storytelling handler + adapter
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector, bPrompter, cmWriter := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error — full flow completed
	require.NoError(t, err)

	// And: 3 stories were persisted via the writer
	assert.Len(t, writer.written, 3)

	// And: display story was called once per story (closing phase)
	assert.Equal(t, 3, prompter.displayStoryCalled)
}

// TestCLIDiscovery_ModeCancellation_PropagatesError verifies that
// context.Canceled from SelectMode propagates through the adapter.
func TestCLIDiscovery_ModeCancellation_PropagatesError(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))

	// And: a prompter that cancels during mode selection
	prompter := &fakeStorytellingPrompter{
		modeErr: context.Canceled,
	}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector, bPrompter, cmWriter := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: context.Canceled propagates
	assert.ErrorIs(t, err, context.Canceled)
}

// TestCLIDiscovery_PersonaCancellation_PropagatesError verifies that
// context.Canceled from persona AskChoice propagates through the adapter.
func TestCLIDiscovery_PersonaCancellation_PropagatesError(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))

	// And: mode succeeds but persona AskChoice cancels
	prompter := &fakeStorytellingPrompter{
		modeChoice: discoverydomain.ModeRapid,
		choiceErr:  context.Canceled,
	}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector, bPrompter, cmWriter := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: context.Canceled propagates
	assert.ErrorIs(t, err, context.Canceled)
}

// TestCLIDiscovery_StoryCancellation_PropagatesError verifies that
// context.Canceled from a narration prompt mid-story propagates through the adapter.
func TestCLIDiscovery_StoryCancellation_PropagatesError(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))

	// And: persona succeeds but narration cancels mid-story
	prompter := &fakeStorytellingPrompter{
		modeChoice:      discoverydomain.ModeRapid,
		choiceResponses: []string{"1"}, // persona succeeds
		narrationErr:    context.Canceled,
	}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector, bPrompter, cmWriter := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: context.Canceled propagates
	assert.ErrorIs(t, err, context.Canceled)
}

// TestCLIDiscovery_MultipleStories verifies the multi-story loop:
// user says "yes" to continue → additional stories collected.
func TestCLIDiscovery_MultipleStories(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("My project idea"), 0o644))

	// And: thorough mode requiring 5 stories
	prompter := newIntegrationStoryPrompter(5)
	prompter.modeChoice = discoverydomain.ModeThorough
	prompter.choiceResponses = []string{"1", "yes", "yes", "yes", "yes"}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector, bPrompter, cmWriter := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".alto"), 0o755))
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error
	require.NoError(t, err)

	// And: 5 stories written
	assert.Len(t, writer.written, 5)
}

// TestCLIDiscovery_MissingREADME verifies that a missing README.md
// returns an error before any prompts are shown.
func TestCLIDiscovery_MissingREADME(t *testing.T) {
	t.Parallel()

	// Given: an empty project directory (no README.md)
	tmpDir := t.TempDir()

	prompter := &fakeStorytellingPrompter{}
	writer := &fakeStoryWriter{}
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	storytellingHandler := application.NewStorytellingHandler(writer, prompter, nil)
	detector, bPrompter, cmWriter := newIntegrationBoundaryFakes()
	bdHandler := application.NewBoundaryDetectionHandler(detector)
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, storytellingHandler, bdHandler, bPrompter, cmWriter, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: error mentions README
	require.Error(t, err)
	assert.Contains(t, err.Error(), "README")

	// And: no stories were written (failed before prompts)
	assert.Empty(t, writer.written)
}
