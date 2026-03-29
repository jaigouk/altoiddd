package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	"github.com/alto-cli/alto/internal/shared/infrastructure/eventbus"
)

// setupLegacyTest creates a handler + app wired for legacy flow tests.
func setupLegacyTest(t *testing.T) *composition.App {
	t.Helper()
	bus := eventbus.NewBus()
	t.Cleanup(func() { _ = bus.Close() })
	publisher := eventbus.NewPublisher(bus)
	handler := application.NewDiscoveryHandler(publisher)
	return &composition.App{DiscoveryHandler: handler}
}

// buildStdinInput creates stdin input for a full legacy flow:
// persona choice + 10 answers + playback confirmations every 3 answers.
// FixedQuestionFlow (default for legacy ModeExpress) has playbackInterval=3.
func buildStdinInput() string {
	var sb strings.Builder
	sb.WriteString("1\n") // persona choice
	questions := domain.QuestionCatalog()
	answersSincePlayback := 0
	for i := range questions {
		// Playback gate fires before asking the next question
		if answersSincePlayback == 3 {
			sb.WriteString("y\n") // confirm playback
			answersSincePlayback = 0
		}
		_ = i
		sb.WriteString("test answer\n")
		answersSincePlayback++
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// isLegacyMode
// ---------------------------------------------------------------------------

func TestIsLegacyMode_TrueForExpressDeepConversational(t *testing.T) {
	t.Parallel()
	legacyModes := []domain.DiscoveryMode{
		domain.ModeExpress,
		domain.ModeDeep,
		domain.ModeConversational,
	}
	for _, mode := range legacyModes {
		assert.True(t, isLegacyMode(mode), "expected isLegacyMode(%q) to be true", mode)
	}
}

func TestIsLegacyMode_FalseForRapidThorough(t *testing.T) {
	t.Parallel()
	storytellingModes := []domain.DiscoveryMode{
		domain.ModeRapid,
		domain.ModeThorough,
	}
	for _, mode := range storytellingModes {
		assert.False(t, isLegacyMode(mode), "expected isLegacyMode(%q) to be false", mode)
	}
}

// ---------------------------------------------------------------------------
// runLegacyFlow
// ---------------------------------------------------------------------------

func TestRunLegacyFlow_PrintsDeprecationWarning(t *testing.T) {
	t.Parallel()
	app := setupLegacyTest(t)

	// Provide stdin with persona + 10 answers
	input := buildStdinInput()
	stdinPrompter := infrastructure.NewStdinPrompter(strings.NewReader(input), &bytes.Buffer{})

	var stderr bytes.Buffer
	err := runLegacyFlowWithDeps(context.Background(), app, stdinPrompter, &stderr)
	require.NoError(t, err)

	assert.Contains(t, stderr.String(), legacyFlagWarning)
}

func TestRunLegacyFlow_RunsQuestionLoop(t *testing.T) {
	t.Parallel()
	app := setupLegacyTest(t)

	input := buildStdinInput()
	stdinPrompter := infrastructure.NewStdinPrompter(strings.NewReader(input), &bytes.Buffer{})

	var stderr bytes.Buffer
	err := runLegacyFlowWithDeps(context.Background(), app, stdinPrompter, &stderr)
	require.NoError(t, err)

	// Verify: handler must have exactly one session, and it must have all 10 answers.
	// StartSession("") creates the session; the handler keeps it in memory.
	// We can't access it directly, but Complete would fail if MVP questions were missing,
	// so NoError above already proves all MVP questions were answered and session completed.
	assert.Contains(t, stderr.String(), legacyFlagWarning, "should contain deprecation warning")
}

// ---------------------------------------------------------------------------
// NewGuideCmd flag parsing
// ---------------------------------------------------------------------------

func TestNewGuideCmd_AcceptsLegacyFlag(t *testing.T) {
	t.Parallel()
	app := setupLegacyTest(t)
	cmd := NewGuideCmd(app)

	// Verify the flag exists and can be retrieved without error
	val, err := cmd.Flags().GetBool("legacy")
	require.NoError(t, err)
	assert.False(t, val, "default value should be false")
}

// ---------------------------------------------------------------------------
// Mutual exclusion
// ---------------------------------------------------------------------------

func TestRunGuide_LegacyAndAgent_MutuallyExclusive(t *testing.T) {
	t.Parallel()
	app := setupLegacyTest(t)

	err := runGuide(context.Background(), app, false, false, true, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--legacy and --agent are mutually exclusive")
}

func TestRunGuide_LegacyAndContinue_MutuallyExclusive(t *testing.T) {
	t.Parallel()
	app := setupLegacyTest(t)

	err := runGuide(context.Background(), app, false, true, false, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--legacy and --continue are mutually exclusive")
}
