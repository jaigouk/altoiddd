package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/challenge/domain"
)

func TestChallengeTypeEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ct       domain.ChallengeType
		expected string
	}{
		{"language", domain.ChallengeLanguage, "language"},
		{"invariant", domain.ChallengeInvariant, "invariant"},
		{"failure_mode", domain.ChallengeFailureMode, "failure_mode"},
		{"boundary", domain.ChallengeBoundary, "boundary"},
		{"aggregate", domain.ChallengeAggregate, "aggregate"},
		{"communication", domain.ChallengeCommunication, "communication"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.ct))
		})
	}

	t.Run("has six members", func(t *testing.T) {
		t.Parallel()
		all := domain.AllChallengeTypes()
		assert.Len(t, all, 6)
	})
}

func TestChallengeVO(t *testing.T) {
	t.Parallel()

	t.Run("valid challenge creation", func(t *testing.T) {
		t.Parallel()
		c, err := domain.NewChallenge(
			domain.ChallengeBoundary,
			"Should Shipping own delivery tracking?",
			"Shipping",
			"Context map: Sales→Shipping",
			"",
		)
		require.NoError(t, err)
		assert.Equal(t, domain.ChallengeBoundary, c.ChallengeType())
		assert.Equal(t, "Should Shipping own delivery tracking?", c.QuestionText())
		assert.Equal(t, "Shipping", c.ContextName())
		assert.Equal(t, "Context map: Sales→Shipping", c.SourceReference())
		assert.Empty(t, c.Evidence())
	})

	t.Run("requires question_text", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewChallenge(
			domain.ChallengeLanguage,
			"",
			"Sales",
			"UL term",
			"",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "question_text")
	})

	t.Run("requires context_name", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewChallenge(
			domain.ChallengeLanguage,
			"Is this ambiguous?",
			"",
			"UL term",
			"",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context_name")
	})

	t.Run("requires source_reference", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewChallenge(
			domain.ChallengeLanguage,
			"Is this ambiguous?",
			"Sales",
			"",
			"",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source_reference")
	})

	t.Run("whitespace-only fields rejected", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewChallenge(
			domain.ChallengeLanguage,
			"   ",
			"Sales",
			"UL term",
			"",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "question_text")
	})

	t.Run("default evidence empty", func(t *testing.T) {
		t.Parallel()
		c, err := domain.NewChallenge(
			domain.ChallengeInvariant,
			"What prevents negative totals?",
			"Sales",
			"OrderAggregate",
			"",
		)
		require.NoError(t, err)
		assert.Empty(t, c.Evidence())
	})

	t.Run("with evidence", func(t *testing.T) {
		t.Parallel()
		c, err := domain.NewChallenge(
			domain.ChallengeFailureMode,
			"What happens if payment fails mid-checkout?",
			"Sales",
			"Checkout Flow story",
			"Step 2 says 'System validates payment'",
		)
		require.NoError(t, err)
		assert.Equal(t, "Step 2 says 'System validates payment'", c.Evidence())
	})
}

func TestChallengeResponseVO(t *testing.T) {
	t.Parallel()

	t.Run("captures acceptance", func(t *testing.T) {
		t.Parallel()
		r := domain.NewChallengeResponse("abc-123", "Good point", true, nil)
		assert.Equal(t, "abc-123", r.ChallengeID())
		assert.Equal(t, "Good point", r.UserResponse())
		assert.True(t, r.Accepted())
	})

	t.Run("captures rejection", func(t *testing.T) {
		t.Parallel()
		r := domain.NewChallengeResponse("abc-123", "Not relevant", false, nil)
		assert.False(t, r.Accepted())
	})

	t.Run("default artifact_updates empty", func(t *testing.T) {
		t.Parallel()
		r := domain.NewChallengeResponse("abc-123", "Yes", true, nil)
		assert.Empty(t, r.ArtifactUpdates())
	})

	t.Run("with artifact updates", func(t *testing.T) {
		t.Parallel()
		r := domain.NewChallengeResponse("abc-123", "Yes, add invariant", true,
			[]string{"Add invariant: Total must be positive"})
		assert.Len(t, r.ArtifactUpdates(), 1)
	})

	t.Run("defensive copy of artifact updates", func(t *testing.T) {
		t.Parallel()
		updates := []string{"update1", "update2"}
		r := domain.NewChallengeResponse("abc-123", "Yes", true, updates)
		updates[0] = "mutated"
		assert.Equal(t, "update1", r.ArtifactUpdates()[0])
	})
}

func TestChallengeIterationVO(t *testing.T) {
	t.Parallel()

	t.Run("convergence delta", func(t *testing.T) {
		t.Parallel()
		it := domain.NewChallengeIteration(nil, nil, 3)
		assert.Equal(t, 3, it.ConvergenceDelta())
	})

	t.Run("holds challenges and responses", func(t *testing.T) {
		t.Parallel()
		c, _ := domain.NewChallenge(
			domain.ChallengeLanguage,
			"Is 'Order' ambiguous?",
			"Sales",
			"UL glossary",
			"",
		)
		r := domain.NewChallengeResponse("r1", "Yes", true, nil)
		it := domain.NewChallengeIteration([]domain.Challenge{c}, []domain.ChallengeResponse{r}, 1)
		assert.Len(t, it.Challenges(), 1)
		assert.Len(t, it.Responses(), 1)
	})

	t.Run("defensive copy of challenges", func(t *testing.T) {
		t.Parallel()
		c, _ := domain.NewChallenge(
			domain.ChallengeLanguage,
			"Is 'Order' ambiguous?",
			"Sales",
			"UL glossary",
			"",
		)
		challenges := []domain.Challenge{c}
		it := domain.NewChallengeIteration(challenges, nil, 0)
		challenges[0] = domain.Challenge{} // mutate original
		assert.Equal(t, "Is 'Order' ambiguous?", it.Challenges()[0].QuestionText())
	})
}
