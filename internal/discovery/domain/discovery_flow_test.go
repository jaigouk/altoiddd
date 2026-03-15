package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// -- QuestionRef Value Object --

func TestNewQuestionRef_WhenValid_ReturnsRef(t *testing.T) {
	t.Parallel()
	ref := NewQuestionRef("Q1", PhaseActors)
	assert.Equal(t, "Q1", ref.ID())
	assert.Equal(t, PhaseActors, ref.Phase())
}

func TestQuestionRef_Equal_WhenSameValues_ReturnsTrue(t *testing.T) {
	t.Parallel()
	a := NewQuestionRef("Q1", PhaseActors)
	b := NewQuestionRef("Q1", PhaseActors)
	assert.True(t, a.Equal(b))
}

func TestQuestionRef_Equal_WhenDifferentValues_ReturnsFalse(t *testing.T) {
	t.Parallel()
	a := NewQuestionRef("Q1", PhaseActors)
	b := NewQuestionRef("Q2", PhaseStory)
	assert.False(t, a.Equal(b))
}

// -- FixedQuestionFlow --

func TestFixedQuestionFlow_ValidateQuestionOrder_WhenSeedPhase_AllowsAlways(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	ref := NewQuestionRef("Q0", PhaseSeed)
	err := flow.ValidateQuestionOrder(ref, nil, nil)
	require.NoError(t, err)
}

func TestFixedQuestionFlow_ValidateQuestionOrder_WhenFirstPhase_Allows(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	ref := NewQuestionRef("Q1", PhaseActors)
	err := flow.ValidateQuestionOrder(ref, nil, nil)
	require.NoError(t, err)
}

func TestFixedQuestionFlow_ValidateQuestionOrder_WhenEarlierPhaseIncomplete_RejectsLaterPhase(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	// Try to answer a Story question without completing Actors phase
	ref := NewQuestionRef("Q3", PhaseStory)
	err := flow.ValidateQuestionOrder(ref, nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestFixedQuestionFlow_ValidateQuestionOrder_WhenEarlierPhaseComplete_AllowsLaterPhase(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	// Complete Actors phase (Q1, Q2)
	answered := []Answer{
		NewAnswer("Q1", "actors answer"),
		NewAnswer("Q2", "entities answer"),
	}
	ref := NewQuestionRef("Q3", PhaseStory)
	err := flow.ValidateQuestionOrder(ref, answered, nil)
	require.NoError(t, err)
}

func TestFixedQuestionFlow_ValidateQuestionOrder_WhenSkippedCountsAsComplete(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	// Q1 answered, Q2 skipped — Actors phase complete
	answered := []Answer{NewAnswer("Q1", "actors")}
	skipped := map[string]bool{"Q2": true}
	ref := NewQuestionRef("Q3", PhaseStory)
	err := flow.ValidateQuestionOrder(ref, answered, skipped)
	require.NoError(t, err)
}

func TestFixedQuestionFlow_IsPlaybackDue_WhenAtInterval_ReturnsTrue(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	assert.True(t, flow.IsPlaybackDue(3))
}

func TestFixedQuestionFlow_IsPlaybackDue_WhenAboveInterval_ReturnsTrue(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	assert.True(t, flow.IsPlaybackDue(4))
}

func TestFixedQuestionFlow_IsPlaybackDue_WhenBelowInterval_ReturnsFalse(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	assert.False(t, flow.IsPlaybackDue(2))
}

func TestFixedQuestionFlow_PlaybackInterval_Returns3(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	assert.Equal(t, 3, flow.PlaybackInterval())
}

func TestFixedQuestionFlow_CheckCompleteness_WhenAllMVPAnswered_ReturnsNoError(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	answers := []Answer{
		NewAnswer("Q1", "a"), NewAnswer("Q3", "a"),
		NewAnswer("Q4", "a"), NewAnswer("Q9", "a"),
		NewAnswer("Q10", "a"),
	}
	err := flow.CheckCompleteness(answers, nil)
	require.NoError(t, err)
}

func TestFixedQuestionFlow_CheckCompleteness_WhenMVPMissing_ReturnsError(t *testing.T) {
	t.Parallel()
	flow := NewFixedQuestionFlow()
	// Missing Q9 and Q10
	answers := []Answer{
		NewAnswer("Q1", "a"), NewAnswer("Q3", "a"),
		NewAnswer("Q4", "a"),
	}
	err := flow.CheckCompleteness(answers, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- DiscoverySession defaults to FixedQuestionFlow --

func TestDiscoverySession_DefaultFlow_IsFixedQuestionFlow(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("test idea")
	// The default flow should behave identically to FixedQuestionFlow:
	// playback at interval 3, phase ordering enforced
	_ = session.DetectPersona("1")
	_ = session.AnswerQuestion("Q1", "actors")
	_ = session.AnswerQuestion("Q2", "entities")

	// Phase ordering should be enforced — can't skip to Events without completing Story
	err := session.AnswerQuestion("Q6", "events")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestDiscoverySession_DefaultFlow_PlaybackAt3(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("test idea")
	_ = session.DetectPersona("1")
	_ = session.AnswerQuestion("Q1", "actors")
	_ = session.AnswerQuestion("Q2", "entities")
	_ = session.AnswerQuestion("Q3", "story")
	// After 3 answers, should be in PlaybackPending
	assert.Equal(t, StatusPlaybackPending, session.Status())
}
