package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

func TestNewConversationTurn_Valid(t *testing.T) {
	t.Parallel()

	turn, err := domain.NewConversationTurn("What problem are you solving?", "We need better onboarding")

	require.NoError(t, err)
	assert.Equal(t, "What problem are you solving?", turn.ConsultantAction())
	assert.Equal(t, "We need better onboarding", turn.UserResponse())
	assert.Empty(t, turn.Synthesis())
	assert.False(t, turn.IsConfirmed())
}

func TestNewConversationTurn_EmptyAction(t *testing.T) {
	t.Parallel()

	_, err := domain.NewConversationTurn("", "some response")

	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewConversationTurn_EmptyResponse(t *testing.T) {
	t.Parallel()

	_, err := domain.NewConversationTurn("some action", "")

	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewConversationTurn_WhitespaceOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action string
		resp   string
	}{
		{"whitespace action", "   \t\n  ", "valid response"},
		{"whitespace response", "valid action", "   \t\n  "},
		{"both whitespace", "   ", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewConversationTurn(tt.action, tt.resp)

			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestConversationTurn_WithSynthesis(t *testing.T) {
	t.Parallel()

	original, err := domain.NewConversationTurn("What is the domain?", "E-commerce platform")
	require.NoError(t, err)

	updated := original.WithSynthesis("The domain is e-commerce")

	// New instance has synthesis.
	assert.Equal(t, "The domain is e-commerce", updated.Synthesis())

	// Original is unchanged.
	assert.Empty(t, original.Synthesis())
}

func TestConversationTurn_Confirm(t *testing.T) {
	t.Parallel()

	original, err := domain.NewConversationTurn("Is this correct?", "Yes")
	require.NoError(t, err)

	confirmed := original.Confirm()

	// New instance is confirmed.
	assert.True(t, confirmed.IsConfirmed())

	// Original is unchanged.
	assert.False(t, original.IsConfirmed())
}

func TestConversationTurn_String(t *testing.T) {
	t.Parallel()

	turn, err := domain.NewConversationTurn("Ask something", "Answer something")
	require.NoError(t, err)

	s := turn.String()

	assert.Contains(t, s, "Ask something")
	assert.Contains(t, s, "Answer something")
}

func TestConversationTurn_Getters(t *testing.T) {
	t.Parallel()

	turn, err := domain.NewConversationTurn("action text", "response text")
	require.NoError(t, err)

	turn = turn.WithSynthesis("synthesis text").Confirm()

	assert.Equal(t, "action text", turn.ConsultantAction())
	assert.Equal(t, "response text", turn.UserResponse())
	assert.Equal(t, "synthesis text", turn.Synthesis())
	assert.True(t, turn.IsConfirmed())
}
