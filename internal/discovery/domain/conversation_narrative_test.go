package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
)

func TestNewConversationNarrative_Empty(t *testing.T) {
	t.Parallel()

	n := domain.NewConversationNarrative()

	assert.Equal(t, 0, n.TurnCount())
	assert.Nil(t, n.LastTurn())
	assert.Empty(t, n.Turns())
}

func TestConversationNarrative_AddTurn(t *testing.T) {
	t.Parallel()

	turn, err := domain.NewConversationTurn("What domain?", "E-commerce")
	require.NoError(t, err)

	original := domain.NewConversationNarrative()
	updated := original.AddTurn(turn)

	// New instance has the turn.
	assert.Equal(t, 1, updated.TurnCount())

	// Original is unchanged.
	assert.Equal(t, 0, original.TurnCount())
}

func TestConversationNarrative_TurnCount(t *testing.T) {
	t.Parallel()

	turn1, err := domain.NewConversationTurn("Q1", "A1")
	require.NoError(t, err)

	turn2, err := domain.NewConversationTurn("Q2", "A2")
	require.NoError(t, err)

	n := domain.NewConversationNarrative().AddTurn(turn1).AddTurn(turn2)

	assert.Equal(t, 2, n.TurnCount())
}

func TestConversationNarrative_LastTurn_Empty(t *testing.T) {
	t.Parallel()

	n := domain.NewConversationNarrative()

	assert.Nil(t, n.LastTurn())
}

func TestConversationNarrative_LastTurn_NonEmpty(t *testing.T) {
	t.Parallel()

	turn1, err := domain.NewConversationTurn("Q1", "A1")
	require.NoError(t, err)

	turn2, err := domain.NewConversationTurn("Q2", "A2")
	require.NoError(t, err)

	n := domain.NewConversationNarrative().AddTurn(turn1).AddTurn(turn2)
	last := n.LastTurn()

	require.NotNil(t, last)
	assert.Equal(t, "Q2", last.ConsultantAction())
	assert.Equal(t, "A2", last.UserResponse())
}

func TestConversationNarrative_Turns_DefensiveCopy(t *testing.T) {
	t.Parallel()

	turn, err := domain.NewConversationTurn("Q1", "A1")
	require.NoError(t, err)

	n := domain.NewConversationNarrative().AddTurn(turn)

	// Get turns and verify contents.
	turns := n.Turns()
	assert.Len(t, turns, 1)
	assert.Equal(t, "Q1", turns[0].ConsultantAction())

	// Mutating the returned slice should not affect the narrative.
	turn2, err := domain.NewConversationTurn("Q2", "A2")
	require.NoError(t, err)

	turns[0] = turn2

	// Original narrative is unchanged.
	originalTurns := n.Turns()
	assert.Equal(t, "Q1", originalTurns[0].ConsultantAction())
}

func TestConversationNarrative_SynthesisCheckpoints(t *testing.T) {
	t.Parallel()

	turn1, err := domain.NewConversationTurn("Q1", "A1")
	require.NoError(t, err)

	turn2, err := domain.NewConversationTurn("Q2", "A2")
	require.NoError(t, err)
	turn2 = turn2.WithSynthesis("Synthesis for Q2")

	turn3, err := domain.NewConversationTurn("Q3", "A3")
	require.NoError(t, err)

	turn4, err := domain.NewConversationTurn("Q4", "A4")
	require.NoError(t, err)
	turn4 = turn4.WithSynthesis("Synthesis for Q4")

	n := domain.NewConversationNarrative().
		AddTurn(turn1).
		AddTurn(turn2).
		AddTurn(turn3).
		AddTurn(turn4)

	checkpoints := n.SynthesisCheckpoints()

	assert.Len(t, checkpoints, 2)
	assert.Equal(t, "Synthesis for Q2", checkpoints[0].Synthesis())
	assert.Equal(t, "Synthesis for Q4", checkpoints[1].Synthesis())
}

func TestConversationNarrative_SynthesisCheckpoints_NoneWithSynthesis(t *testing.T) {
	t.Parallel()

	turn1, err := domain.NewConversationTurn("Q1", "A1")
	require.NoError(t, err)

	turn2, err := domain.NewConversationTurn("Q2", "A2")
	require.NoError(t, err)

	n := domain.NewConversationNarrative().AddTurn(turn1).AddTurn(turn2)

	checkpoints := n.SynthesisCheckpoints()

	assert.Empty(t, checkpoints)
}
