package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
)

// fakePublisher is a spy that records published events.
type fakePublisher struct {
	published []any
}

func (f *fakePublisher) Publish(_ context.Context, event any) error {
	f.published = append(f.published, event)
	return nil
}

// fakeSessionRepo is a spy that records Save calls and supports Load/Exists.
type fakeSessionRepo struct {
	saved      []*domain.DiscoverySession
	saveErr    error
	loadResult *domain.DiscoverySession
	loadErr    error
	existsVal  bool
	existsErr  error
}

func (f *fakeSessionRepo) Save(_ context.Context, session *domain.DiscoverySession) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, session)
	return nil
}

func (f *fakeSessionRepo) Load(_ context.Context, _ string) (*domain.DiscoverySession, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return f.loadResult, nil
}

func (f *fakeSessionRepo) Exists(_ context.Context, _ string) (bool, error) {
	return f.existsVal, f.existsErr
}

// ---------------------------------------------------------------------------
// Tests — Start Session
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_StartSession(t *testing.T) {
	t.Parallel()

	t.Run("returns session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
		session, err := handler.StartSession("A project idea in 4-5 sentences.")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCreated, session.Status())
		assert.Equal(t, "A project idea in 4-5 sentences.", session.ReadmeContent())
	})

	t.Run("creates unique ids", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
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
		handler := application.NewDiscoveryHandler(&fakePublisher{})
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
		handler := application.NewDiscoveryHandler(&fakePublisher{})
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
		handler := application.NewDiscoveryHandler(&fakePublisher{})
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		result, err := handler.AnswerQuestion(session.SessionID(), "Q1", "Users and admins")
		require.NoError(t, err)
		assert.Len(t, result.Answers(), 1)
		assert.Equal(t, "Q1", result.Answers()[0].QuestionID())
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
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
		handler := application.NewDiscoveryHandler(&fakePublisher{})
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		result, err := handler.SkipQuestion(session.SessionID(), "Q1", "Not relevant")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusPersonaDetected, result.Status())
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
		_, err := handler.SkipQuestion("nonexistent-id", "Q1", "Reason")
		require.Error(t, err)
	})

	t.Run("empty reason raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
		session, _ := handler.StartSession("Idea")
		handler.DetectPersona(session.SessionID(), "1")
		_, err := handler.SkipQuestion(session.SessionID(), "Q1", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("unknown question raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
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
		handler := application.NewDiscoveryHandler(&fakePublisher{})
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
		handler := application.NewDiscoveryHandler(&fakePublisher{})
		_, err := handler.ConfirmPlayback("nonexistent-id", true)
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// Tests — GetSession
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_GetSession_Found(t *testing.T) {
	t.Parallel()
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	session, err := handler.StartSession("An idea")
	require.NoError(t, err)

	found, err := handler.GetSession(session.SessionID())

	require.NoError(t, err)
	assert.Equal(t, session.SessionID(), found.SessionID())
	assert.Equal(t, "An idea", found.ReadmeContent())
}

func TestDiscoveryHandler_GetSession_NotFound(t *testing.T) {
	t.Parallel()
	handler := application.NewDiscoveryHandler(&fakePublisher{})

	_, err := handler.GetSession("nonexistent-id")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active discovery session")
}

// ---------------------------------------------------------------------------
// Tests — Complete
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_Complete(t *testing.T) {
	t.Parallel()

	t.Run("returns completed session", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
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
		assert.Len(t, result.Events(), 1)
	})

	t.Run("not found raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewDiscoveryHandler(&fakePublisher{})
		_, err := handler.Complete("nonexistent-id")
		require.Error(t, err)
	})
}

func TestDiscoveryHandler_Complete_PublishesEvent(t *testing.T) {
	t.Parallel()

	pub := &fakePublisher{}
	handler := application.NewDiscoveryHandler(pub)
	session, err := handler.StartSession("Idea")
	require.NoError(t, err)
	handler.DetectPersona(session.SessionID(), "1")

	for _, qid := range []string{"Q1", "Q2", "Q3"} {
		handler.AnswerQuestion(session.SessionID(), qid, "Answer "+qid)
	}
	handler.ConfirmPlayback(session.SessionID(), true)

	for _, qid := range []string{"Q4", "Q5", "Q6"} {
		handler.AnswerQuestion(session.SessionID(), qid, "Answer "+qid)
	}
	handler.ConfirmPlayback(session.SessionID(), true)

	for _, qid := range []string{"Q7", "Q8", "Q9"} {
		handler.AnswerQuestion(session.SessionID(), qid, "Answer "+qid)
	}
	handler.ConfirmPlayback(session.SessionID(), true)

	handler.AnswerQuestion(session.SessionID(), "Q10", "Answer Q10")

	_, err = handler.Complete(session.SessionID())
	require.NoError(t, err)

	require.Len(t, pub.published, 1)
	_, ok := pub.published[0].(domain.DiscoveryCompletedEvent)
	assert.True(t, ok, "expected DiscoveryCompletedEvent, got %T", pub.published[0])
}

// ---------------------------------------------------------------------------
// Tests — Session Persistence (optional SessionRepository)
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_AnswerQuestion_WhenSessionRepoExists_ExpectSaved(t *testing.T) {
	t.Parallel()
	repo := &fakeSessionRepo{}
	handler := application.NewDiscoveryHandler(&fakePublisher{}, application.WithSessionRepository(repo))
	session, _ := handler.StartSession("Idea")
	handler.DetectPersona(session.SessionID(), "1")

	_, err := handler.AnswerQuestion(session.SessionID(), "Q1", "Users and admins")
	require.NoError(t, err)

	require.Len(t, repo.saved, 1)
	assert.Equal(t, session.SessionID(), repo.saved[0].SessionID())
}

func TestDiscoveryHandler_SkipQuestion_WhenSessionRepoExists_ExpectSaved(t *testing.T) {
	t.Parallel()
	repo := &fakeSessionRepo{}
	handler := application.NewDiscoveryHandler(&fakePublisher{}, application.WithSessionRepository(repo))
	session, _ := handler.StartSession("Idea")
	handler.DetectPersona(session.SessionID(), "1")

	_, err := handler.SkipQuestion(session.SessionID(), "Q1", "Not relevant")
	require.NoError(t, err)

	require.Len(t, repo.saved, 1)
	assert.Equal(t, session.SessionID(), repo.saved[0].SessionID())
}

func TestDiscoveryHandler_WhenSessionRepoNil_ExpectNoSaveAttempt(t *testing.T) {
	t.Parallel()
	// No WithSessionRepository option — sessionRepo is nil
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	session, _ := handler.StartSession("Idea")
	handler.DetectPersona(session.SessionID(), "1")

	// Should not panic when sessionRepo is nil
	_, err := handler.AnswerQuestion(session.SessionID(), "Q1", "Users and admins")
	require.NoError(t, err)

	_, err = handler.SkipQuestion(session.SessionID(), "Q2", "Not relevant")
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Tests — LoadOrGetSession
// ---------------------------------------------------------------------------

func TestDiscoveryHandler_LoadOrGetSession_WhenInMemory_ExpectReturned(t *testing.T) {
	t.Parallel()
	handler := application.NewDiscoveryHandler(&fakePublisher{})
	session, err := handler.StartSession("Idea")
	require.NoError(t, err)

	found, err := handler.LoadOrGetSession(session.SessionID())
	require.NoError(t, err)
	assert.Equal(t, session.SessionID(), found.SessionID())
}

func TestDiscoveryHandler_LoadOrGetSession_WhenNotInMemory_ExpectLoadedFromRepo(t *testing.T) {
	t.Parallel()

	// Create a session via snapshot to simulate a persisted session
	snapshot := map[string]interface{}{
		"session_id":                  "persisted-session-123",
		"readme_content":              "Persisted idea",
		"status":                      "persona_detected",
		"persona":                     "developer",
		"register":                    "technical",
		"answers":                     []interface{}{},
		"skipped":                     []interface{}{},
		"playback_confirmations":      []interface{}{},
		"answers_since_last_playback": float64(0),
	}
	persistedSession, err := domain.FromSnapshot(snapshot)
	require.NoError(t, err)

	repo := &fakeSessionRepo{loadResult: persistedSession}
	handler := application.NewDiscoveryHandler(&fakePublisher{}, application.WithSessionRepository(repo))

	found, err := handler.LoadOrGetSession("persisted-session-123")
	require.NoError(t, err)
	assert.Equal(t, "persisted-session-123", found.SessionID())
	assert.Equal(t, "Persisted idea", found.ReadmeContent())
}

func TestDiscoveryHandler_LoadOrGetSession_WhenLoadedFromRepo_ExpectCachedInMemory(t *testing.T) {
	t.Parallel()

	snapshot := map[string]interface{}{
		"session_id":                  "cached-session-456",
		"readme_content":              "Cached idea",
		"status":                      "persona_detected",
		"persona":                     "developer",
		"register":                    "technical",
		"answers":                     []interface{}{},
		"skipped":                     []interface{}{},
		"playback_confirmations":      []interface{}{},
		"answers_since_last_playback": float64(0),
	}
	persistedSession, err := domain.FromSnapshot(snapshot)
	require.NoError(t, err)

	repo := &fakeSessionRepo{loadResult: persistedSession}
	handler := application.NewDiscoveryHandler(&fakePublisher{}, application.WithSessionRepository(repo))

	// First call loads from repo
	_, err = handler.LoadOrGetSession("cached-session-456")
	require.NoError(t, err)

	// Second call should find it in-memory via GetSession
	found, err := handler.GetSession("cached-session-456")
	require.NoError(t, err)
	assert.Equal(t, "cached-session-456", found.SessionID())
}

func TestDiscoveryHandler_LoadOrGetSession_WhenNoRepoAndNotInMemory_ExpectError(t *testing.T) {
	t.Parallel()
	handler := application.NewDiscoveryHandler(&fakePublisher{})

	_, err := handler.LoadOrGetSession("nonexistent-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active discovery session")
}

func TestDiscoveryHandler_LoadOrGetSession_WhenRepoLoadFails_ExpectError(t *testing.T) {
	t.Parallel()
	repo := &fakeSessionRepo{loadErr: fmt.Errorf("disk read error")}
	handler := application.NewDiscoveryHandler(&fakePublisher{}, application.WithSessionRepository(repo))

	_, err := handler.LoadOrGetSession("some-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading session")
}
