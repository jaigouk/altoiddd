package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
)

// ---------------------------------------------------------------------------
// Tests — Start Session
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_StartSession(t *testing.T) {
	t.Parallel()

	t.Run("returns session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, err := handler.StartSession("A project idea in 4-5 sentences.")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCreated, session.Status())
		assert.Equal(t, "A project idea in 4-5 sentences.", session.ReadmeContent())
	})

	t.Run("creates unique ids", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		s1, _ := handler.StartSession("Idea A")
		s2, _ := handler.StartSession("Idea B")
		assert.NotEqual(t, s1.SessionID(), s2.SessionID())
	})
}

// ---------------------------------------------------------------------------
// Tests — Detect Persona
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_DetectPersona(t *testing.T) {
	t.Parallel()

	t.Run("returns updated session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, _ := handler.StartSession("Idea")
		result, err := handler.DetectPersona(session.SessionID(), "1")
		require.NoError(t, err)

		persona, ok := result.Persona()
		assert.True(t, ok)
		assert.Equal(t, domain.PersonaDeveloper, persona)

		register, ok := result.Register()
		assert.True(t, ok)
		assert.Equal(t, domain.RegisterTechnical, register)
		assert.Equal(t, domain.StatusPersonaDetected, result.Status())
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		_, err := handler.DetectPersona("nonexistent-id", "1")
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// Tests — Answer Question
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_AnswerQuestion(t *testing.T) {
	t.Parallel()

	t.Run("returns updated session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		result, err := handler.AnswerQuestion(session.SessionID(), "Q1", "Users and admins")
		require.NoError(t, err)
		assert.Equal(t, 1, len(result.Answers()))
		assert.Equal(t, "Q1", result.Answers()[0].QuestionID())
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		_, err := handler.AnswerQuestion("nonexistent-id", "Q1", "Answer")
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// Tests — Skip Question
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_SkipQuestion(t *testing.T) {
	t.Parallel()

	t.Run("returns updated session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		result, err := handler.SkipQuestion(session.SessionID(), "Q1", "Not relevant")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusPersonaDetected, result.Status())
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		_, err := handler.SkipQuestion("nonexistent-id", "Q1", "Reason")
		require.Error(t, err)
	})

	t.Run("empty reason raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		_, err := handler.SkipQuestion(session.SessionID(), "Q1", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("unknown question raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		_, err := handler.SkipQuestion(session.SessionID(), "Q999", "Reason")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nknown") // "Unknown" or "unknown"
	})
}

// ---------------------------------------------------------------------------
// Tests — Confirm Playback
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_ConfirmPlayback(t *testing.T) {
	t.Parallel()

	t.Run("returns updated session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		handler.AnswerQuestion(session.SessionID(), "Q1", "Users")
		handler.AnswerQuestion(session.SessionID(), "Q2", "Entities")
		handler.AnswerQuestion(session.SessionID(), "Q3", "Use case")

		result, err := handler.ConfirmPlayback(session.SessionID(), true)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusAnswering, result.Status())
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		_, err := handler.ConfirmPlayback("nonexistent-id", true)
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// Tests — Complete
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_Complete(t *testing.T) {
	t.Parallel()

	t.Run("returns completed session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")

		// Answer Q1-Q3, confirm playback
		for _, qid := range []string{"Q1", "Q2", "Q3"} {
			handler.AnswerQuestion(session.SessionID(), qid, "Answer "+qid)
		}
		handler.ConfirmPlayback(session.SessionID(), true)

		// Answer Q4-Q6, confirm playback
		for _, qid := range []string{"Q4", "Q5", "Q6"} {
			handler.AnswerQuestion(session.SessionID(), qid, "Answer "+qid)
		}
		handler.ConfirmPlayback(session.SessionID(), true)

		// Answer Q7-Q9, confirm playback
		for _, qid := range []string{"Q7", "Q8", "Q9"} {
			handler.AnswerQuestion(session.SessionID(), qid, "Answer "+qid)
		}
		handler.ConfirmPlayback(session.SessionID(), true)

		// Answer Q10
		handler.AnswerQuestion(session.SessionID(), "Q10", "Answer Q10")

		result, err := handler.Complete(session.SessionID())
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, result.Status())
		assert.Equal(t, 1, len(result.Events()))
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler()
		_, err := handler.Complete("nonexistent-id")
		require.Error(t, err)
	})
}
