package domain

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- Helpers --

func sessionWithPersona(choice string) *DiscoverySession {
	session := NewDiscoverySession("A test project idea.")
	err := session.DetectPersona(choice)
	if err != nil {
		panic(fmt.Sprintf("failed to detect persona: %v", err))
	}
	return session
}

func answerQuestions(session *DiscoverySession, questionIDs []string) {
	for _, qid := range questionIDs {
		if session.Status() == StatusPlaybackPending {
			_ = session.ConfirmPlayback(true, "")
		}
		_ = session.AnswerQuestion(qid, fmt.Sprintf("Answer for %s", qid))
		if session.Status() == StatusPlaybackPending {
			_ = session.ConfirmPlayback(true, "")
		}
	}
}

func answerAllQuestions(session *DiscoverySession) {
	catalog := QuestionCatalog()
	for _, q := range catalog {
		if session.Status() == StatusPlaybackPending {
			_ = session.ConfirmPlayback(true, "")
		}
		_ = session.AnswerQuestion(q.ID(), fmt.Sprintf("Answer for %s", q.ID()))
	}
	if session.Status() == StatusPlaybackPending {
		_ = session.ConfirmPlayback(true, "")
	}
}

// -- Creation --

func TestNewSessionStartsInCreatedState(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("An idea.")
	assert.Equal(t, StatusCreated, session.Status())
}

func TestNewSessionHasUniqueID(t *testing.T) {
	t.Parallel()
	s1 := NewDiscoverySession("Idea A")
	s2 := NewDiscoverySession("Idea B")
	assert.NotEqual(t, s1.SessionID(), s2.SessionID())
}

func TestNewSessionStoresReadme(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("My project idea.")
	assert.Equal(t, "My project idea.", session.ReadmeContent())
}

func TestNewSessionHasNoPersona(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	p, ok := session.Persona()
	assert.False(t, ok)
	assert.Equal(t, DiscoveryPersona(""), p)
}

func TestNewSessionHasNoRegister(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	r, ok := session.Register()
	assert.False(t, ok)
	assert.Equal(t, DiscoveryRegister(""), r)
}

func TestNewSessionHasNoAnswers(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	assert.Empty(t, session.Answers())
}

func TestNewSessionHasNoEvents(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	assert.Empty(t, session.Events())
}

func TestNewSessionCurrentPhaseIsSeed(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	assert.Equal(t, PhaseSeed, session.CurrentPhase())
}

// -- Persona Detection --

func TestPersonaDetection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		choice          string
		expectedPersona DiscoveryPersona
		expectedReg     DiscoveryRegister
	}{
		{"choice_1_developer_technical", "1", PersonaDeveloper, RegisterTechnical},
		{"choice_2_product_owner_non_technical", "2", PersonaProductOwner, RegisterNonTechnical},
		{"choice_3_domain_expert_non_technical", "3", PersonaDomainExpert, RegisterNonTechnical},
		{"choice_4_mixed_non_technical", "4", PersonaMixed, RegisterNonTechnical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			session := NewDiscoverySession("Idea")
			err := session.DetectPersona(tt.choice)
			require.NoError(t, err)
			p, ok := session.Persona()
			assert.True(t, ok)
			assert.Equal(t, tt.expectedPersona, p)
			r, ok := session.Register()
			assert.True(t, ok)
			assert.Equal(t, tt.expectedReg, r)
		})
	}
}

func TestDetectPersonaTransitionsToPersonaDetected(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.DetectPersona("1"))
	assert.Equal(t, StatusPersonaDetected, session.Status())
}

func TestInvalidPersonaChoiceReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	err := session.DetectPersona("5")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid persona choice")
}

func TestDetectPersonaNotFromCreatedReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.DetectPersona("2")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- Answer Question --

func TestAnswerRecordsResponse(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users and admins"))
	assert.Len(t, session.Answers(), 1)
	assert.Equal(t, "Q1", session.Answers()[0].QuestionID())
	assert.Equal(t, "Users and admins", session.Answers()[0].ResponseText())
}

func TestAnswerTransitionsToAnswering(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	assert.Equal(t, StatusAnswering, session.Status())
}

func TestAnswerAdvancesPhase(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	assert.Equal(t, PhaseStory, session.CurrentPhase())
}

// -- Invariant 1: Cannot answer before persona detection --

func TestAnswerBeforePersonaReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	err := session.AnswerQuestion("Q1", "Answer")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- Invariant 2: Phase order enforced --

func TestAnswerOutOfPhaseReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.AnswerQuestion("Q6", "Some events")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "phase")
}

func TestAnswerBoundaryQuestionInActorsPhaseReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.AnswerQuestion("Q9", "Bounded contexts")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- Invariant 3: Playback after 3 answers --

func TestPlaybackTriggeredAfterThreeAnswers(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	require.NoError(t, session.AnswerQuestion("Q3", "Use case"))
	assert.Equal(t, StatusPlaybackPending, session.Status())
}

func TestCannotAnswerDuringPlayback(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	require.NoError(t, session.AnswerQuestion("Q3", "Use case"))
	err := session.AnswerQuestion("Q4", "Failure")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestConfirmPlaybackResumesAnswering(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	require.NoError(t, session.AnswerQuestion("Q3", "Use case"))
	require.NoError(t, session.ConfirmPlayback(true, ""))
	assert.Equal(t, StatusAnswering, session.Status())
}

func TestRejectPlaybackWithCorrections(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	require.NoError(t, session.AnswerQuestion("Q3", "Use case"))
	require.NoError(t, session.ConfirmPlayback(false, "Fix actors list"))
	assert.Equal(t, StatusAnswering, session.Status())
}

func TestConfirmPlaybackNotInPlaybackStateReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.ConfirmPlayback(true, "")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSecondPlaybackAfterSixAnswers(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	// Answer Q1-Q3, playback auto-confirmed by helper
	_ = session.AnswerQuestion("Q1", "Users")
	_ = session.AnswerQuestion("Q2", "Entities")
	_ = session.AnswerQuestion("Q3", "Use case")
	_ = session.ConfirmPlayback(true, "")
	// After first playback confirm, continue
	require.NoError(t, session.AnswerQuestion("Q4", "Failure"))
	require.NoError(t, session.AnswerQuestion("Q5", "Workflows"))
	require.NoError(t, session.AnswerQuestion("Q6", "Events"))
	assert.Equal(t, StatusPlaybackPending, session.Status())
}

func TestPlaybackStoresConfirmation(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	require.NoError(t, session.AnswerQuestion("Q3", "Use case"))
	require.NoError(t, session.ConfirmPlayback(true, ""))
	assert.Len(t, session.PlaybackConfirmations(), 1)
	assert.True(t, session.PlaybackConfirmations()[0].Confirmed())
}

// -- Invariant 4: Skip requires reason --

func TestSkipWithReasonSucceeds(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.SkipQuestion("Q1", "Not relevant"))
	for _, a := range session.Answers() {
		assert.NotEqual(t, "Q1", a.QuestionID())
	}
}

func TestSkipWithoutReasonReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.SkipQuestion("Q1", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reason")
}

func TestSkipAdvancesPastQuestion(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.SkipQuestion("Q1", "Not relevant"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	assert.Len(t, session.Answers(), 1)
	assert.Equal(t, "Q2", session.Answers()[0].QuestionID())
}

func TestSkipBeforePersonaReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	err := session.SkipQuestion("Q1", "Not relevant")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSkipDuringPlaybackReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	require.NoError(t, session.AnswerQuestion("Q2", "Entities"))
	require.NoError(t, session.AnswerQuestion("Q3", "Use case"))
	err := session.SkipQuestion("Q4", "Not needed")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSkipUnknownQuestionReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.SkipQuestion("Q99", "Invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown question")
}

func TestDiscoverySession_WhenQuestionSkipped_StoresReason(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.SkipQuestion("Q1", "Not relevant to this project"))
	reason := session.SkipReason("Q1")
	assert.Equal(t, "Not relevant to this project", reason)
}

func TestDiscoverySession_WhenQuestionSkipped_ReasonRetrievable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		questionID string
		reason     string
	}{
		{"short reason", "Q1", "Not relevant"},
		{"detailed reason", "Q1", "This question does not apply because we have no external actors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			session := sessionWithPersona("1")
			require.NoError(t, session.SkipQuestion(tt.questionID, tt.reason))
			assert.Equal(t, tt.reason, session.SkipReason(tt.questionID))
		})
	}
	// Unskipped question returns empty reason
	t.Run("unskipped question returns empty", func(t *testing.T) {
		t.Parallel()
		session := sessionWithPersona("1")
		assert.Empty(t, session.SkipReason("Q1"))
	})
}

func TestDiscoverySession_WhenQuestionUnskipped_CanBeAnswered(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.SkipQuestion("Q1", "Skipping for now"))
	require.NoError(t, session.UnskipQuestion("Q1"))
	// After unskipping, the question can be answered
	require.NoError(t, session.AnswerQuestion("Q1", "Users and admins"))
	assert.Len(t, session.Answers(), 1)
	assert.Equal(t, "Q1", session.Answers()[0].QuestionID())
	// Reason should be gone
	assert.Empty(t, session.SkipReason("Q1"))
}

func TestDiscoverySession_WhenUnskippingUnskippedQuestion_ReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.UnskipQuestion("Q1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not skipped")
}

func TestDiscoverySession_ToSnapshot_IncludesSkipReasons(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.SkipQuestion("Q1", "Not relevant"))
	require.NoError(t, session.SkipQuestion("Q2", "Will revisit later"))
	snapshot := session.ToSnapshot()
	skipped, ok := snapshot["skipped"].([]map[string]string)
	require.True(t, ok, "skipped should be []map[string]string, got %T", snapshot["skipped"])
	assert.Len(t, skipped, 2)
	// Build a map for order-independent checking
	reasons := make(map[string]string)
	for _, entry := range skipped {
		reasons[entry["question_id"]] = entry["reason"]
	}
	assert.Equal(t, "Not relevant", reasons["Q1"])
	assert.Equal(t, "Will revisit later", reasons["Q2"])
}

func TestDiscoverySession_FromSnapshot_ParsesSkipReasons(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.SkipQuestion("Q1", "Not relevant"))
	require.NoError(t, session.SkipQuestion("Q2", "Will revisit later"))
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, "Not relevant", restored.SkipReason("Q1"))
	assert.Equal(t, "Will revisit later", restored.SkipReason("Q2"))
}

func TestDiscoverySession_FromSnapshot_HandlesLegacyBoolFormat(t *testing.T) {
	t.Parallel()
	// Simulate the old format: skipped was []string of question IDs
	session := sessionWithPersona("1")
	snapshot := session.ToSnapshot()
	// Overwrite skipped with legacy format
	snapshot["skipped"] = []string{"Q1", "Q2"}
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	// Legacy format should parse with empty reason strings
	assert.Empty(t, restored.SkipReason("Q1"))
	assert.Empty(t, restored.SkipReason("Q2"))
	// But the questions should still be considered skipped (phase advancement works)
	// Verify by checking that answering Q3 works (Q1, Q2 are handled)
	require.NoError(t, restored.AnswerQuestion("Q3", "Use cases"))
}

// -- Unknown question ID --

func TestAnswerUnknownQuestionReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.AnswerQuestion("Q99", "Answer")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown question")
}

// -- Invariant 5: Complete requires MVP questions --

func TestCompleteWithAllMVPQuestionsSucceeds(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	assert.Equal(t, StatusCompleted, session.Status())
}

func TestCompleteWithFewerThan5AnswersReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerQuestions(session, []string{"Q1", "Q2", "Q3"})
	// answerQuestions auto-confirms playback so session is in ANSWERING state
	err := session.Complete()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "MVP")
}

// -- Domain Events --

func TestCompleteEmitsDiscoveryCompleted(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	assert.Len(t, session.Events(), 1)
	event := session.Events()[0]
	assert.Equal(t, session.SessionID(), event.SessionID())
	assert.Equal(t, PersonaDeveloper, event.Persona())
	assert.Equal(t, RegisterTechnical, event.Register())
	assert.Len(t, event.Answers(), 10)
}

func TestEventsReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	events := session.Events()
	events = events[:0]
	_ = events
	assert.Len(t, session.Events(), 1)
}

// -- Edge Cases --

func TestDuplicateAnswerReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	err := session.AnswerQuestion("Q1", "Different answer")
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "already answered")
}

func TestEmptyAnswerReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.AnswerQuestion("Q1", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestWhitespaceOnlyAnswerReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.AnswerQuestion("Q1", "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestDoubleCompleteReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	err := session.Complete()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestCompleteFromCreatedReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	err := session.Complete()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- TechStack wiring --

func TestDefaultTechStackIsNil(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("# Test")
	assert.Nil(t, session.TechStack())
}

func TestSetTechStackInCreatedState(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("# Test")
	ts := vo.NewTechStack("python", "uv")
	require.NoError(t, session.SetTechStack(&ts))
	require.NotNil(t, session.TechStack())
	assert.Equal(t, "python", session.TechStack().Language())
}

func TestSetTechStackInPersonaDetectedState(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	ts := vo.NewTechStack("python", "uv")
	require.NoError(t, session.SetTechStack(&ts))
	require.NotNil(t, session.TechStack())
}

func TestSetTechStackTwiceOverwrites(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("# Test")
	ts1 := vo.NewTechStack("python", "uv")
	ts2 := vo.NewTechStack("rust", "cargo")
	require.NoError(t, session.SetTechStack(&ts1))
	require.NoError(t, session.SetTechStack(&ts2))
	assert.Equal(t, "rust", session.TechStack().Language())
}

func TestSetTechStackInAnsweringStateReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Actors"))
	ts := vo.NewTechStack("python", "uv")
	err := session.SetTechStack(&ts)
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- Discovery Mode tests --

func TestDefaultModeIsExpress(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	assert.Equal(t, ModeExpress, session.Mode())
}

func TestSetModeDeepInCreatedState(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	assert.Equal(t, ModeDeep, session.Mode())
}

func TestSetModeNotInCreatedStateReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	err := session.SetMode(ModeDeep)
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSetModeTwiceReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	err := session.SetMode(ModeExpress)
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- Deep mode state transitions --

func TestDeepCompleteToRound1Complete(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	assert.Equal(t, StatusRound1Complete, session.Status())
	assert.Empty(t, session.Events())
}

func TestStartChallengeFromRound1Complete(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.NoError(t, session.StartChallenge())
	assert.Equal(t, StatusChallenging, session.Status())
}

func TestStartChallengeInExpressModeReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	err := session.StartChallenge()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestCompleteChallengeFromChallenging(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.NoError(t, session.StartChallenge())
	require.NoError(t, session.CompleteChallenge())
	assert.Equal(t, StatusRound2Complete, session.Status())
}

func TestStartSimulateFromRound2Complete(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.NoError(t, session.StartChallenge())
	require.NoError(t, session.CompleteChallenge())
	require.NoError(t, session.StartSimulate())
	assert.Equal(t, StatusSimulating, session.Status())
}

func TestCompleteSimulationFromSimulating(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.NoError(t, session.StartChallenge())
	require.NoError(t, session.CompleteChallenge())
	require.NoError(t, session.StartSimulate())
	require.NoError(t, session.CompleteSimulation())
	assert.Equal(t, StatusCompleted, session.Status())
	assert.Len(t, session.Events(), 1)
}

func TestFullDeepFlowEmitsEventOnlyAtEnd(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	assert.Empty(t, session.Events())
	require.NoError(t, session.StartChallenge())
	assert.Empty(t, session.Events())
	require.NoError(t, session.CompleteChallenge())
	assert.Empty(t, session.Events())
	require.NoError(t, session.StartSimulate())
	assert.Empty(t, session.Events())
	require.NoError(t, session.CompleteSimulation())
	assert.Len(t, session.Events(), 1)
}

// -- Snapshot round-trip tests --

func TestSnapshotRoundTripCreatedState(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("# My Project\nA cool tool.")
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, session.SessionID(), restored.SessionID())
	assert.Equal(t, session.ReadmeContent(), restored.ReadmeContent())
	assert.Equal(t, StatusCreated, restored.Status())
	_, ok := restored.Persona()
	assert.False(t, ok)
}

func TestSnapshotRoundTripPersonaDetectedState(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, StatusPersonaDetected, restored.Status())
	p, ok := restored.Persona()
	assert.True(t, ok)
	assert.Equal(t, PersonaDeveloper, p)
}

func TestSnapshotRoundTripAnsweringState(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Answer for Q1"))
	require.NoError(t, session.AnswerQuestion("Q2", "Answer for Q2"))
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, StatusAnswering, restored.Status())
	assert.Len(t, restored.Answers(), 2)
}

func TestSnapshotRoundTripPlaybackPendingState(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Answer for Q1"))
	require.NoError(t, session.AnswerQuestion("Q2", "Answer for Q2"))
	require.NoError(t, session.AnswerQuestion("Q3", "Answer for Q3"))
	assert.Equal(t, StatusPlaybackPending, session.Status())
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, StatusPlaybackPending, restored.Status())
}

func TestSnapshotRoundTripCompletedState(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, restored.Status())
	assert.Len(t, restored.Answers(), 10)
}

func TestSnapshotJSONSerializable(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	snapshot := session.ToSnapshot()
	jsonBytes, err := json.Marshal(snapshot)
	require.NoError(t, err)
	var loaded map[string]interface{}
	require.NoError(t, json.Unmarshal(jsonBytes, &loaded))
	restored, err := FromSnapshot(loaded)
	require.NoError(t, err)
	assert.Equal(t, session.SessionID(), restored.SessionID())
}

func TestSnapshotPreservesPlaybackConfirmations(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "A"))
	require.NoError(t, session.AnswerQuestion("Q2", "B"))
	require.NoError(t, session.AnswerQuestion("Q3", "C"))
	require.NoError(t, session.ConfirmPlayback(true, "Minor fix"))
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Len(t, restored.PlaybackConfirmations(), 1)
	assert.True(t, restored.PlaybackConfirmations()[0].Confirmed())
	assert.Equal(t, "Minor fix", restored.PlaybackConfirmations()[0].Corrections())
}

func TestSnapshotPreservesPlaybackCounter(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "A"))
	require.NoError(t, session.AnswerQuestion("Q2", "B"))
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	// One more answer should trigger playback
	require.NoError(t, restored.AnswerQuestion("Q3", "C"))
	assert.Equal(t, StatusPlaybackPending, restored.Status())
}

func TestSnapshotMissingFieldReturnsError(t *testing.T) {
	t.Parallel()
	_, err := FromSnapshot(map[string]interface{}{})
	require.Error(t, err)
}

func TestSnapshotInvalidStatusReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	snapshot := session.ToSnapshot()
	snapshot["status"] = "bogus_state"
	_, err := FromSnapshot(snapshot)
	require.Error(t, err)
}

func TestSnapshotInvalidPersonaReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	snapshot := session.ToSnapshot()
	snapshot["persona"] = "alien"
	_, err := FromSnapshot(snapshot)
	require.Error(t, err)
}

func TestSnapshotCreatedWithPersonaSetReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	snapshot := session.ToSnapshot()
	snapshot["persona"] = "developer"
	snapshot["register"] = "technical"
	_, err := FromSnapshot(snapshot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CREATED state must have persona=nil")
}

func TestSnapshotAnsweringWithNilPersonaReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "A"))
	snapshot := session.ToSnapshot()
	snapshot["persona"] = nil
	snapshot["register"] = nil
	_, err := FromSnapshot(snapshot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a persona")
}

func TestSnapshotPlaybackPendingCounterMismatchReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "A"))
	require.NoError(t, session.AnswerQuestion("Q2", "B"))
	require.NoError(t, session.AnswerQuestion("Q3", "C"))
	snapshot := session.ToSnapshot()
	snapshot["answers_since_last_playback"] = float64(1) // JSON number
	_, err := FromSnapshot(snapshot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PLAYBACK_PENDING state requires counter=3")
}

func TestSnapshotNegativePlaybackCounterReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "A"))
	snapshot := session.ToSnapshot()
	snapshot["answers_since_last_playback"] = float64(-1)
	_, err := FromSnapshot(snapshot)
	require.Error(t, err)
}

func TestSnapshotRestoredSessionEnforcesInvariants(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "A"))
	restored, err := FromSnapshot(session.ToSnapshot())
	require.NoError(t, err)
	// Duplicate answer rejected
	err = restored.AnswerQuestion("Q1", "Dup")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already answered")
}

func TestSnapshotWithTechStack(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("# Test")
	ts := vo.NewTechStack("python", "uv")
	require.NoError(t, session.SetTechStack(&ts))
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	require.NotNil(t, restored.TechStack())
	assert.Equal(t, "python", restored.TechStack().Language())
}

func TestSnapshotWithNilTechStack(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("# Test")
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Nil(t, restored.TechStack())
}

func TestSnapshotModeRoundTrip(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, ModeDeep, restored.Mode())
}

func TestSnapshotOldWithoutModeDefaultsExpress(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	snapshot := session.ToSnapshot()
	delete(snapshot, "mode")
	delete(snapshot, "round")
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, ModeExpress, restored.Mode())
}

// -- Missing parity tests from test_discovery_mode.py --

func TestSetModeExpressInCreatedState(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeExpress))
	assert.Equal(t, ModeExpress, session.Mode())
}

func TestSetModeInAnsweringStateReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	err := session.SetMode(ModeDeep)
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSetModeTwiceSameValueReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	err := session.SetMode(ModeDeep)
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestDefaultModeAfterPersonaDetection(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	assert.Equal(t, ModeExpress, session.Mode())
}

func TestExplicitExpressModeCompleteSetCompleted(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeExpress))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	assert.Equal(t, StatusCompleted, session.Status())
	assert.Len(t, session.Events(), 1)
}

func TestStartChallengeFromWrongStateReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	// Still in ANSWERING, not ROUND_1_COMPLETE
	err := session.StartChallenge()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestCompleteChallengeFromWrongStateReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete()) // ROUND_1_COMPLETE
	err := session.CompleteChallenge()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStartSimulateBeforeRound2CompleteReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete()) // ROUND_1_COMPLETE
	err := session.StartSimulate()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStartSimulateInExpressModeReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	err := session.StartSimulate()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestCompleteSimulationFromWrongStateReturnsError(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.NoError(t, session.StartChallenge())
	require.NoError(t, session.CompleteChallenge())
	// ROUND_2_COMPLETE, not SIMULATING
	err := session.CompleteSimulation()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSnapshotIncludesModeAndRound(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	snapshot := session.ToSnapshot()
	assert.Equal(t, "deep", snapshot["mode"])
}

func TestSnapshotNoneModeWhenNotSet(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	snapshot := session.ToSnapshot()
	assert.Nil(t, snapshot["mode"])
	assert.Nil(t, snapshot["round"])
}

func TestSnapshotRoundTripExpressDefaultPreservesExpress(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, ModeExpress, restored.Mode())
}

func TestSnapshotRoundTripDeepRound1Complete(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, ModeDeep, restored.Mode())
	assert.Equal(t, StatusRound1Complete, restored.Status())
}

func TestSnapshotRoundTripThroughFullDeepFlow(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.NoError(t, session.StartChallenge())
	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)
	assert.Equal(t, StatusChallenging, restored.Status())
	assert.Equal(t, ModeDeep, restored.Mode())
}

func TestDeepCompleteSimulationEventHasAllFields(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.NoError(t, session.StartChallenge())
	require.NoError(t, session.CompleteChallenge())
	require.NoError(t, session.StartSimulate())
	require.NoError(t, session.CompleteSimulation())
	require.Len(t, session.Events(), 1)
	event := session.Events()[0]
	assert.Equal(t, session.SessionID(), event.SessionID())
	assert.Equal(t, PersonaDeveloper, event.Persona())
	assert.Equal(t, RegisterTechnical, event.Register())
	assert.NotEmpty(t, event.Answers())
}

func TestModeSurvivesPersonaDetection(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	assert.Equal(t, ModeDeep, session.Mode())
}

func TestModeSurvivesAnswering(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	require.NoError(t, session.AnswerQuestion("Q1", "Users"))
	assert.Equal(t, ModeDeep, session.Mode())
}

func TestExpressCannotCompleteTwice(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	err := session.Complete()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestDeepRound1CannotCompleteTwice(t *testing.T) {
	t.Parallel()
	session := NewDiscoverySession("Idea")
	require.NoError(t, session.SetMode(ModeDeep))
	require.NoError(t, session.DetectPersona("1"))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	err := session.Complete()
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestDiscoveryCompletedEventCarriesTechStack(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	ts := vo.NewTechStack("python", "uv")
	require.NoError(t, session.SetTechStack(&ts))
	answerAllQuestions(session)
	require.NoError(t, session.Complete())
	require.Len(t, session.Events(), 1)
	assert.NotNil(t, session.Events()[0].TechStack())
	assert.Equal(t, "python", session.Events()[0].TechStack().Language())
}
