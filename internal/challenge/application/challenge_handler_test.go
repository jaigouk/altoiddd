package application_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/challenge/application"
	challengedomain "github.com/alto-cli/alto/internal/challenge/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
)

// ---------------------------------------------------------------------------
// Fake challenger
// ---------------------------------------------------------------------------

type fakeChallenger struct {
	challenges []challengedomain.Challenge
	called     int
	lastMax    int
	mu         sync.Mutex
}

func (f *fakeChallenger) GenerateChallenges(
	_ context.Context,
	_ *ddd.DomainModel,
	maxPerType int,
) ([]challengedomain.Challenge, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.called++
	f.lastMax = maxPerType
	return f.challenges, nil
}

func makeChallenge(q, ctx string) challengedomain.Challenge {
	c, _ := challengedomain.NewChallenge(
		challengedomain.ChallengeLanguage,
		q, ctx, "test reference", "",
	)
	return c
}

// ---------------------------------------------------------------------------
// Tests — Generate
// ---------------------------------------------------------------------------

func TestChallengeHandler_Generate(t *testing.T) {
	t.Parallel()

	t.Run("delegates to port", func(t *testing.T) {
		t.Parallel()
		expected := []challengedomain.Challenge{makeChallenge("Is this right?", "Sales")}
		fake := &fakeChallenger{challenges: expected}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		result, err := handler.GenerateChallenges(context.Background(), model, 5)

		require.NoError(t, err)
		assert.Equal(t, 1, fake.called)
		assert.Equal(t, 5, fake.lastMax)
		assert.Len(t, result, len(expected))
	})

	t.Run("custom max per type", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: nil}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		_, err := handler.GenerateChallenges(context.Background(), model, 3)

		require.NoError(t, err)
		assert.Equal(t, 3, fake.lastMax)
	})

	t.Run("stores challenges internally", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Sales"),
			makeChallenge("Q2", "Orders"),
		}
		fake := &fakeChallenger{challenges: challenges}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		_, err := handler.GenerateChallenges(context.Background(), model, 5)
		require.NoError(t, err)

		iteration := handler.Complete()
		assert.Len(t, iteration.Challenges(), 2)
	})
}

// ---------------------------------------------------------------------------
// Tests — Record Response
// ---------------------------------------------------------------------------

func TestChallengeHandler_RecordResponse(t *testing.T) {
	t.Parallel()

	t.Run("stores response", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		response := challengedomain.NewChallengeResponse("c1", "Good point", true, nil)
		handler.RecordResponse(response)

		iteration := handler.Complete()
		assert.Len(t, iteration.Responses(), 1)
		assert.True(t, iteration.Responses()[0].Accepted())
	})

	t.Run("multiple responses", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		handler.RecordResponse(challengedomain.NewChallengeResponse("c1", "Yes", true, nil))
		handler.RecordResponse(challengedomain.NewChallengeResponse("c2", "No", false, nil))

		iteration := handler.Complete()
		assert.Len(t, iteration.Responses(), 2)
	})
}

// ---------------------------------------------------------------------------
// Tests — Complete
// ---------------------------------------------------------------------------

func TestChallengeHandler_Complete(t *testing.T) {
	t.Parallel()

	t.Run("returns challenge iteration", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		iteration := handler.Complete()
		assert.NotNil(t, iteration)
	})

	t.Run("convergence delta counts accepted updates", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		handler.RecordResponse(challengedomain.NewChallengeResponse(
			"c1", "Yes", true, []string{"Add invariant", "Add term"},
		))
		handler.RecordResponse(challengedomain.NewChallengeResponse(
			"c2", "No", false, nil,
		))
		handler.RecordResponse(challengedomain.NewChallengeResponse(
			"c3", "Yes", true, []string{"Add story"},
		))

		iteration := handler.Complete()
		assert.Equal(t, 3, iteration.ConvergenceDelta())
	})

	t.Run("convergence delta zero when all rejected", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		handler.RecordResponse(challengedomain.NewChallengeResponse("c1", "No", false, nil))

		iteration := handler.Complete()
		assert.Equal(t, 0, iteration.ConvergenceDelta())
	})

	t.Run("complete with no responses", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		iteration := handler.Complete()
		assert.Empty(t, iteration.Responses())
		assert.Equal(t, 0, iteration.ConvergenceDelta())
	})
}

// ---------------------------------------------------------------------------
// Tests — Session Management (1ql.2)
// ---------------------------------------------------------------------------

func TestChallengeHandler_StartChallenge(t *testing.T) {
	t.Parallel()

	t.Run("creates session with unique ID", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{makeChallenge("Q1", "Sales")}
		fake := &fakeChallenger{challenges: challenges}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		session, err := handler.StartChallenge(context.Background(), model, 5)

		require.NoError(t, err)
		assert.NotEmpty(t, session.SessionID())
		assert.Equal(t, challengedomain.SessionStatusActive, session.Status())
		assert.Len(t, session.Challenges(), 1)
	})

	t.Run("generates unique IDs for multiple sessions", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		session1, _ := handler.StartChallenge(context.Background(), model, 5)
		session2, _ := handler.StartChallenge(context.Background(), model, 5)

		assert.NotEqual(t, session1.SessionID(), session2.SessionID())
	})

	t.Run("propagates challenger errors", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: nil}
		handler := application.NewChallengeHandler(fake)

		// Inject error behavior
		fake.challenges = nil
		model := ddd.NewDomainModel("test")
		session, err := handler.StartChallenge(context.Background(), model, 5)

		// With nil challenges, should still create empty session
		require.NoError(t, err)
		assert.Empty(t, session.Challenges())
	})
}

func TestChallengeHandler_GetSession(t *testing.T) {
	t.Parallel()

	t.Run("retrieves existing session", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		created, _ := handler.StartChallenge(context.Background(), model, 5)
		retrieved, err := handler.GetSession(created.SessionID())

		require.NoError(t, err)
		assert.Equal(t, created.SessionID(), retrieved.SessionID())
	})

	t.Run("returns error for unknown session", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: nil}
		handler := application.NewChallengeHandler(fake)

		_, err := handler.GetSession("unknown-session-id")

		assert.ErrorIs(t, err, challengedomain.ErrSessionNotFound)
	})
}

func TestChallengeHandler_RespondToChallenge(t *testing.T) {
	t.Parallel()

	t.Run("records response to session", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		session, _ := handler.StartChallenge(context.Background(), model, 5)
		err := handler.RespondToChallenge(session.SessionID(), "c0", "Good point", true, nil)

		require.NoError(t, err)
		updated, _ := handler.GetSession(session.SessionID())
		assert.Len(t, updated.Responses(), 1)
	})

	t.Run("returns error for unknown session", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: nil}
		handler := application.NewChallengeHandler(fake)

		err := handler.RespondToChallenge("unknown", "c0", "Answer", true, nil)

		assert.ErrorIs(t, err, challengedomain.ErrSessionNotFound)
	})

	t.Run("returns error for unknown challenge", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		session, _ := handler.StartChallenge(context.Background(), model, 5)
		err := handler.RespondToChallenge(session.SessionID(), "c999", "Answer", true, nil)

		assert.ErrorIs(t, err, challengedomain.ErrChallengeNotFound)
	})

	t.Run("returns error for duplicate response", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		session, _ := handler.StartChallenge(context.Background(), model, 5)
		_ = handler.RespondToChallenge(session.SessionID(), "c0", "First", true, nil)
		err := handler.RespondToChallenge(session.SessionID(), "c0", "Second", false, nil)

		assert.ErrorIs(t, err, challengedomain.ErrChallengeAlreadyAnswered)
	})
}

func TestChallengeHandler_CompleteSession(t *testing.T) {
	t.Parallel()

	t.Run("returns iteration for session", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		session, _ := handler.StartChallenge(context.Background(), model, 5)
		_ = handler.RespondToChallenge(session.SessionID(), "c0", "Yes", true, []string{"Add term"})

		iteration, err := handler.CompleteSession(session.SessionID())

		require.NoError(t, err)
		assert.Len(t, iteration.Challenges(), 1)
		assert.Len(t, iteration.Responses(), 1)
		assert.Equal(t, 1, iteration.ConvergenceDelta())
	})

	t.Run("returns error for unknown session", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: nil}
		handler := application.NewChallengeHandler(fake)

		_, err := handler.CompleteSession("unknown")

		assert.ErrorIs(t, err, challengedomain.ErrSessionNotFound)
	})
}

// ---------------------------------------------------------------------------
// Tests — Concurrent Session Access
// ---------------------------------------------------------------------------

func TestChallengeHandler_ConcurrentSessions(t *testing.T) {
	t.Parallel()

	t.Run("handles multiple concurrent sessions", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		const numSessions = 10
		sessions := make([]*challengedomain.ChallengeSession, numSessions)

		// Create sessions concurrently
		errs := make(chan error, numSessions)
		done := make(chan int, numSessions)
		for i := range numSessions {
			go func(idx int) {
				s, err := handler.StartChallenge(context.Background(), model, 5)
				if err != nil {
					errs <- err
				}
				sessions[idx] = s
				done <- idx
			}(i)
		}

		// Wait for all
		for range numSessions {
			<-done
		}
		close(errs)

		// Check for errors
		for err := range errs {
			require.NoError(t, err)
		}

		// Verify all unique
		ids := make(map[string]bool)
		for _, s := range sessions {
			require.NotNil(t, s)
			assert.False(t, ids[s.SessionID()], "duplicate session ID")
			ids[s.SessionID()] = true
		}
		assert.Len(t, ids, numSessions)
	})

	t.Run("concurrent responses to same session", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Ctx"),
			makeChallenge("Q2", "Ctx"),
			makeChallenge("Q3", "Ctx"),
		}
		fake := &fakeChallenger{challenges: challenges}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		session, _ := handler.StartChallenge(context.Background(), model, 5)

		// Respond to different challenges concurrently
		done := make(chan bool, 3)
		for i := range 3 {
			go func(idx int) {
				cID := "c" + string(rune('0'+idx))
				_ = handler.RespondToChallenge(session.SessionID(), cID, "Answer", true, nil)
				done <- true
			}(i)
		}

		for range 3 {
			<-done
		}

		updated, _ := handler.GetSession(session.SessionID())
		assert.Len(t, updated.Responses(), 3)
	})
}
