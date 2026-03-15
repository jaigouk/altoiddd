package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// -- ConversationalFlow --

func TestConversationalFlow_ValidateQuestionOrder_WhenAnyOrder_AllowsAlways(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	// Try answering Events phase without completing earlier phases — should be allowed
	ref := NewQuestionRef("conv_events_1", PhaseEvents)
	err := flow.ValidateQuestionOrder(ref, nil, nil)
	require.NoError(t, err)
}

func TestConversationalFlow_ValidateQuestionOrder_WhenOutOfOrder_AllowsAlways(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	ref := NewQuestionRef("conv_boundaries_1", PhaseBoundaries)
	// No prior answers at all — conversational flow doesn't enforce order
	err := flow.ValidateQuestionOrder(ref, nil, nil)
	require.NoError(t, err)
}

func TestConversationalFlow_IsPlaybackDue_WhenAtInterval_ReturnsTrue(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	assert.True(t, flow.IsPlaybackDue(5))
}

func TestConversationalFlow_IsPlaybackDue_WhenAboveInterval_ReturnsTrue(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	assert.True(t, flow.IsPlaybackDue(6))
}

func TestConversationalFlow_IsPlaybackDue_WhenBelowInterval_ReturnsFalse(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	assert.False(t, flow.IsPlaybackDue(4))
}

func TestConversationalFlow_PlaybackInterval_ReturnsConfigured(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(7)
	assert.Equal(t, 7, flow.PlaybackInterval())
}

func TestConversationalFlow_PlaybackInterval_DefaultsTo5_WhenZero(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(0)
	assert.Equal(t, defaultConversationalPlaybackInterval, flow.PlaybackInterval())
}

func TestConversationalFlow_PlaybackInterval_DefaultsTo5_WhenNegative(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(-1)
	assert.Equal(t, defaultConversationalPlaybackInterval, flow.PlaybackInterval())
}

func TestConversationalFlow_CheckCompleteness_WhenAllPhasesHaveAnswers_ReturnsNoError(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	answers := []Answer{
		NewAnswer("conv_actors_1", "users and admins"),
		NewAnswer("conv_story_1", "user logs in"),
		NewAnswer("conv_events_1", "UserLoggedIn event"),
		NewAnswer("conv_boundaries_1", "auth context"),
	}
	// Need a way to map answer question IDs to phases for semantic check.
	// ConversationalFlow needs phase metadata — use QuestionRef registry.
	flow.RegisterQuestion(NewQuestionRef("conv_actors_1", PhaseActors))
	flow.RegisterQuestion(NewQuestionRef("conv_story_1", PhaseStory))
	flow.RegisterQuestion(NewQuestionRef("conv_events_1", PhaseEvents))
	flow.RegisterQuestion(NewQuestionRef("conv_boundaries_1", PhaseBoundaries))

	err := flow.CheckCompleteness(answers, nil)
	require.NoError(t, err)
}

func TestConversationalFlow_CheckCompleteness_WhenMissingPhase_ReturnsError(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	answers := []Answer{
		NewAnswer("conv_actors_1", "users"),
		NewAnswer("conv_story_1", "user logs in"),
		// Missing Events and Boundaries
	}
	flow.RegisterQuestion(NewQuestionRef("conv_actors_1", PhaseActors))
	flow.RegisterQuestion(NewQuestionRef("conv_story_1", PhaseStory))

	err := flow.CheckCompleteness(answers, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestConversationalFlow_CheckCompleteness_WhenNoAnswers_ReturnsError(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	err := flow.CheckCompleteness(nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestConversationalFlow_RegisterQuestion_TracksPhase(t *testing.T) {
	t.Parallel()
	flow := NewConversationalFlow(5)
	flow.RegisterQuestion(NewQuestionRef("conv_actors_1", PhaseActors))
	flow.RegisterQuestion(NewQuestionRef("conv_actors_2", PhaseActors))

	completeness := flow.Completeness([]Answer{
		NewAnswer("conv_actors_1", "a"),
		NewAnswer("conv_actors_2", "b"),
	})
	// Only actors phase covered
	assert.False(t, completeness.IsComplete())
	assert.Len(t, completeness.Gaps(), 3) // story, events, boundaries missing
}

// -- DiscoverySession with ConversationalFlow --

func TestDiscoverySession_ConversationalMode_AcceptsOutOfOrderAnswers(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("test idea")
	err := session.SetMode(ModeConversational)
	require.NoError(t, err)
	err = session.DetectPersona("1")
	require.NoError(t, err)

	// Answer Q9 (Boundaries) before Q1 (Actors) — allowed in conversational mode
	err = session.AnswerQuestion("Q9", "auth and billing contexts")
	require.NoError(t, err)
	assert.Equal(t, StatusAnswering, session.Status())
}

func TestDiscoverySession_ConversationalMode_PlaybackAt5(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("test idea")
	err := session.SetMode(ModeConversational)
	require.NoError(t, err)
	err = session.DetectPersona("1")
	require.NoError(t, err)

	// Answer 4 questions — should NOT trigger playback (interval=5)
	for _, qid := range []string{"Q1", "Q2", "Q3", "Q4"} {
		err = session.AnswerQuestion(qid, "answer for "+qid)
		require.NoError(t, err)
	}
	assert.Equal(t, StatusAnswering, session.Status())

	// 5th answer should trigger playback
	err = session.AnswerQuestion("Q5", "answer for Q5")
	require.NoError(t, err)
	assert.Equal(t, StatusPlaybackPending, session.Status())
}

func TestDiscoverySession_ConversationalMode_PreservesUniversalInvariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		action  func(*DiscoverySession) error
		wantErr bool
	}{
		{
			name: "cannot answer before persona detection",
			action: func(s *DiscoverySession) error {
				return s.AnswerQuestion("Q1", "answer")
			},
			wantErr: true,
		},
		{
			name: "empty answer rejected",
			action: func(s *DiscoverySession) error {
				_ = s.DetectPersona("1")
				return s.AnswerQuestion("Q1", "")
			},
			wantErr: true,
		},
		{
			name: "duplicate answer rejected",
			action: func(s *DiscoverySession) error {
				_ = s.DetectPersona("1")
				_ = s.AnswerQuestion("Q1", "first")
				return s.AnswerQuestion("Q1", "second")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			session := NewDiscoverySession("test")
			_ = session.SetMode(ModeConversational)
			err := tt.action(session)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
