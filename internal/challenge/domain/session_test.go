package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
)

func makeChallenge(q, ctx string) challengedomain.Challenge {
	c, _ := challengedomain.NewChallenge(
		challengedomain.ChallengeLanguage,
		q, ctx, "test reference", "",
	)
	return c
}

func TestNewChallengeSession(t *testing.T) {
	t.Parallel()

	t.Run("creates session with ID and challenges", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Sales"),
			makeChallenge("Q2", "Orders"),
		}

		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		assert.Equal(t, "sess-123", session.SessionID())
		assert.Equal(t, "model-1", session.DomainModelID())
		assert.Equal(t, challengedomain.SessionStatusActive, session.Status())
		assert.Len(t, session.Challenges(), 2)
		assert.Empty(t, session.Responses())
	})

	t.Run("defensive copy of challenges", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{makeChallenge("Q1", "Sales")}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		// Mutate original
		challenges[0] = makeChallenge("Modified", "Other")

		// Session should be unaffected
		assert.Equal(t, "Q1", session.Challenges()[0].QuestionText())
	})
}

func TestChallengeSession_ChallengeByID(t *testing.T) {
	t.Parallel()

	t.Run("returns challenge when found", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Sales"),
			makeChallenge("Q2", "Orders"),
		}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		c, found := session.ChallengeByID("c0")
		require.True(t, found)
		assert.Equal(t, "Q1", c.QuestionText())

		c, found = session.ChallengeByID("c1")
		require.True(t, found)
		assert.Equal(t, "Q2", c.QuestionText())
	})

	t.Run("returns false when not found", func(t *testing.T) {
		t.Parallel()
		session := challengedomain.NewChallengeSession("sess-123", "model-1", nil)

		_, found := session.ChallengeByID("c999")
		assert.False(t, found)
	})
}

func TestChallengeSession_RecordResponse(t *testing.T) {
	t.Parallel()

	t.Run("records response successfully", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{makeChallenge("Q1", "Sales")}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		response := challengedomain.NewChallengeResponse("c0", "Good point", true, nil)
		err := session.RecordResponse(response)

		require.NoError(t, err)
		assert.Len(t, session.Responses(), 1)
		assert.True(t, session.HasResponse("c0"))
	})

	t.Run("returns error for unknown challenge", func(t *testing.T) {
		t.Parallel()
		session := challengedomain.NewChallengeSession("sess-123", "model-1", nil)

		response := challengedomain.NewChallengeResponse("c999", "Answer", true, nil)
		err := session.RecordResponse(response)

		assert.ErrorIs(t, err, challengedomain.ErrChallengeNotFound)
	})

	t.Run("returns error for duplicate response", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{makeChallenge("Q1", "Sales")}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		response1 := challengedomain.NewChallengeResponse("c0", "First", true, nil)
		err := session.RecordResponse(response1)
		require.NoError(t, err)

		response2 := challengedomain.NewChallengeResponse("c0", "Second", false, nil)
		err = session.RecordResponse(response2)

		assert.ErrorIs(t, err, challengedomain.ErrChallengeAlreadyAnswered)
	})

	t.Run("auto-completes when all challenges answered", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Sales"),
			makeChallenge("Q2", "Orders"),
		}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		_ = session.RecordResponse(challengedomain.NewChallengeResponse("c0", "A1", true, nil))
		assert.Equal(t, challengedomain.SessionStatusActive, session.Status())

		_ = session.RecordResponse(challengedomain.NewChallengeResponse("c1", "A2", true, nil))
		assert.Equal(t, challengedomain.SessionStatusCompleted, session.Status())
	})
}

func TestChallengeSession_ChallengeIDs(t *testing.T) {
	t.Parallel()

	t.Run("returns IDs for all challenges", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Sales"),
			makeChallenge("Q2", "Orders"),
			makeChallenge("Q3", "Shipping"),
		}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		ids := session.ChallengeIDs()

		assert.Equal(t, []string{"c0", "c1", "c2"}, ids)
	})

	t.Run("handles 10+ challenges correctly", func(t *testing.T) {
		t.Parallel()
		// Create 12 challenges to test double-digit IDs
		challenges := make([]challengedomain.Challenge, 12)
		for i := range 12 {
			challenges[i] = makeChallenge("Q", "Ctx")
		}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		ids := session.ChallengeIDs()

		// Verify double-digit IDs are formatted correctly
		assert.Equal(t, "c10", ids[10])
		assert.Equal(t, "c11", ids[11])

		// Verify we can look up challenge 10 by ID
		c, found := session.ChallengeByID("c10")
		require.True(t, found)
		assert.Equal(t, "Q", c.QuestionText())
	})
}

func TestChallengeSession_ToIteration(t *testing.T) {
	t.Parallel()

	t.Run("converts to iteration with convergence delta", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Sales"),
			makeChallenge("Q2", "Orders"),
		}
		session := challengedomain.NewChallengeSession("sess-123", "model-1", challenges)

		_ = session.RecordResponse(challengedomain.NewChallengeResponse(
			"c0", "Yes", true, []string{"Add invariant", "Add term"},
		))
		_ = session.RecordResponse(challengedomain.NewChallengeResponse(
			"c1", "No", false, nil,
		))

		iteration := session.ToIteration()

		assert.Len(t, iteration.Challenges(), 2)
		assert.Len(t, iteration.Responses(), 2)
		assert.Equal(t, 2, iteration.ConvergenceDelta())
	})
}
