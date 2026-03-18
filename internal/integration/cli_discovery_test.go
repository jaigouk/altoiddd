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
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
)

// ---------------------------------------------------------------------------
// Fake Prompter for CLI Discovery Tests
// ---------------------------------------------------------------------------

// fakePrompter implements application.Prompter for testing without TUI.
type fakePrompter struct {
	personaChoice     string
	personaErr        error
	answers           []string // Answers for each question (empty = skip)
	skipReasons       []string // Reasons for skipped questions
	answerIdx         int      // Current answer index
	skipIdx           int      // Current skip reason index
	questionErr       error    // Error to return from AskQuestion
	skipReasonErr     error    // Error to return from AskSkipReason
	questionsAsked    []string // Records questions asked (for verification)
	playbackConfirmed bool     // What to return from ConfirmPlayback
	playbackErr       error    // Error to return from ConfirmPlayback
	playbackSummaries []string // Records summaries shown (for verification)
}

func (f *fakePrompter) SelectPersona(_ context.Context) (string, error) {
	return f.personaChoice, f.personaErr
}

func (f *fakePrompter) AskQuestion(_ context.Context, question string) (string, error) {
	f.questionsAsked = append(f.questionsAsked, question)
	if f.questionErr != nil {
		return "", f.questionErr
	}
	if f.answerIdx >= len(f.answers) {
		return "", nil // No more answers, return empty (skip)
	}
	answer := f.answers[f.answerIdx]
	f.answerIdx++
	return answer, nil
}

func (f *fakePrompter) AskSkipReason(_ context.Context) (string, error) {
	if f.skipReasonErr != nil {
		return "", f.skipReasonErr
	}
	if f.skipIdx >= len(f.skipReasons) {
		return "no reason given", nil
	}
	reason := f.skipReasons[f.skipIdx]
	f.skipIdx++
	return reason, nil
}

func (f *fakePrompter) ConfirmPlayback(_ context.Context, summary string) (bool, error) {
	f.playbackSummaries = append(f.playbackSummaries, summary)
	if f.playbackErr != nil {
		return false, f.playbackErr
	}
	return f.playbackConfirmed, nil
}

// Compile-time interface check.
var _ application.Prompter = (*fakePrompter)(nil)

// ---------------------------------------------------------------------------
// Fake Event Publisher
// ---------------------------------------------------------------------------

type fakePublisher struct{}

func (f *fakePublisher) Publish(_ context.Context, _ any) error { return nil }

var _ sharedapp.EventPublisher = (*fakePublisher)(nil)

// ---------------------------------------------------------------------------
// Integration Tests
// ---------------------------------------------------------------------------

func TestCLIDiscovery_HappyPath_AllQuestionsAnswered(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea for testing"), 0o644))

	// And: a prompter configured to answer all 10 questions
	answers := make([]string, 10)
	for i := range answers {
		answers[i] = "Answer for question " + string(rune('A'+i))
	}
	prompter := &fakePrompter{
		personaChoice:     "1", // Developer (technical register)
		answers:           answers,
		playbackConfirmed: true,
	}

	// And: a fully wired discovery adapter
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error
	require.NoError(t, err)

	// And: all 10 questions were asked
	assert.Len(t, prompter.questionsAsked, 10)

	// And: playback was triggered 3 times (after Q3, Q6, Q9)
	assert.Len(t, prompter.playbackSummaries, 3)

	// And: each playback summary contains answers
	for _, summary := range prompter.playbackSummaries {
		assert.NotEmpty(t, summary)
		assert.Contains(t, summary, "Q:")
		assert.Contains(t, summary, "A:")
	}
}

func TestCLIDiscovery_SkipQuestion_TriggersSkipReason(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// And: answers with Q3 skipped (empty string)
	answers := []string{"A1", "A2", "", "A4", "A5", "A6", "A7", "A8", "A9", "A10"}
	prompter := &fakePrompter{
		personaChoice:     "1",
		answers:           answers,
		skipReasons:       []string{"Not relevant to my project"},
		playbackConfirmed: true,
	}

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error (skip is valid)
	require.NoError(t, err)

	// And: AskSkipReason was called (skipIdx advanced)
	assert.Equal(t, 1, prompter.skipIdx, "AskSkipReason should have been called once")
}

func TestCLIDiscovery_PlaybackRejection_ContinuesFlow(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// And: prompter configured to reject playback (user wants to review)
	// Note: Current implementation continues to next questions even on rejection
	// because handler.ConfirmPlayback(id, false) transitions back to StatusAnswering
	answers := make([]string, 10)
	for i := range answers {
		answers[i] = "Answer"
	}

	prompter := &fakePrompter{
		personaChoice:     "1",
		answers:           answers,
		playbackConfirmed: false, // User rejects playback (wants to review)
	}

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error (rejection transitions back to answering, flow continues)
	require.NoError(t, err)

	// And: all 10 questions were answered
	assert.Len(t, prompter.questionsAsked, 10)

	// And: playback was triggered 3 times (after Q3, Q6, Q9)
	assert.Len(t, prompter.playbackSummaries, 3)
}

func TestCLIDiscovery_Cancellation_PropagatesError(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// And: prompter configured to cancel immediately on first question
	// Note: questionErr is checked before consuming answers, so this cancels on Q1
	prompter := &fakePrompter{
		personaChoice: "1",
		questionErr:   context.Canceled,
	}

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: context.Canceled is returned
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCLIDiscovery_PersonaCancellation_PropagatesError(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// And: prompter configured to cancel during persona selection
	prompter := &fakePrompter{
		personaErr: context.Canceled,
	}

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: context.Canceled is returned
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCLIDiscovery_PlaybackCancellation_PropagatesError(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// And: answers for first 3 questions to trigger playback
	answers := []string{"A1", "A2", "A3"}
	prompter := &fakePrompter{
		personaChoice: "1",
		answers:       answers,
		playbackErr:   context.Canceled, // Cancel during playback confirmation
	}

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: context.Canceled is returned
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCLIDiscovery_NonTechnicalRegister_UsesPlainLanguage(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// And: prompter configured as Product Owner (non-technical)
	answers := make([]string, 10)
	for i := range answers {
		answers[i] = "Answer"
	}
	prompter := &fakePrompter{
		personaChoice:     "2", // Product Owner (non-technical register)
		answers:           answers,
		playbackConfirmed: true,
	}

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error
	require.NoError(t, err)

	// And: questions use non-technical language (no "actors", "entities", "domain")
	// First question for non-technical: "Who will use this product..."
	require.NotEmpty(t, prompter.questionsAsked)
	firstQuestion := prompter.questionsAsked[0]
	assert.Contains(t, firstQuestion, "Who will use this product")
}

func TestCLIDiscovery_TechnicalRegister_UsesDDDLanguage(t *testing.T) {
	t.Parallel()

	// Given: a project directory with README
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("My project idea"), 0o644))

	// And: prompter configured as Developer (technical)
	answers := make([]string, 10)
	for i := range answers {
		answers[i] = "Answer"
	}
	prompter := &fakePrompter{
		personaChoice:     "1", // Developer (technical register)
		answers:           answers,
		playbackConfirmed: true,
	}

	handler := application.NewDiscoveryHandler(&fakePublisher{})
	adapter := infrastructure.NewCLIDiscoveryAdapter(handler, prompter, tmpDir)

	// When: running the discovery flow
	err := adapter.Run(context.Background())

	// Then: no error
	require.NoError(t, err)

	// And: questions use technical/DDD language ("actors", "entities")
	require.NotEmpty(t, prompter.questionsAsked)
	firstQuestion := prompter.questionsAsked[0]
	assert.Contains(t, firstQuestion, "actors")
}
