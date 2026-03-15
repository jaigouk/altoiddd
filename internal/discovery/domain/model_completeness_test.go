package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -- ModelGap Value Object --

func TestNewModelGap_WhenValid_ReturnsGap(t *testing.T) {
	t.Parallel()
	gap, err := NewModelGap(PhaseActors, "no actors defined")
	require.NoError(t, err)
	assert.Equal(t, PhaseActors, gap.Phase())
	assert.Equal(t, "no actors defined", gap.Description())
}

func TestNewModelGap_WhenEmptyDescription_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewModelGap(PhaseActors, "")
	require.Error(t, err)
}

func TestNewModelGap_WhenWhitespaceDescription_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewModelGap(PhaseActors, "   ")
	require.Error(t, err)
}

func TestModelGap_Equal_WhenSameValues_ReturnsTrue(t *testing.T) {
	t.Parallel()
	a, _ := NewModelGap(PhaseActors, "no actors")
	b, _ := NewModelGap(PhaseActors, "no actors")
	assert.True(t, a.Equal(b))
}

func TestModelGap_Equal_WhenDifferentValues_ReturnsFalse(t *testing.T) {
	t.Parallel()
	a, _ := NewModelGap(PhaseActors, "no actors")
	b, _ := NewModelGap(PhaseStory, "no stories")
	assert.False(t, a.Equal(b))
}

// -- ModelCompleteness Value Object --

func TestNewModelCompleteness_WhenNoGaps_IsComplete(t *testing.T) {
	t.Parallel()
	mc := NewModelCompleteness(nil)
	assert.True(t, mc.IsComplete())
	assert.Empty(t, mc.Gaps())
}

func TestNewModelCompleteness_WhenHasGaps_IsNotComplete(t *testing.T) {
	t.Parallel()
	gap, _ := NewModelGap(PhaseActors, "no actors defined")
	mc := NewModelCompleteness([]ModelGap{gap})
	assert.False(t, mc.IsComplete())
	assert.Len(t, mc.Gaps(), 1)
	assert.Equal(t, PhaseActors, mc.Gaps()[0].Phase())
}

func TestModelCompleteness_Gaps_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	gap, _ := NewModelGap(PhaseActors, "no actors")
	mc := NewModelCompleteness([]ModelGap{gap})
	gaps := mc.Gaps()
	assert.Len(t, gaps, 1)
	// Modifying returned slice should not affect the original
	gaps[0] = ModelGap{}
	assert.Equal(t, PhaseActors, mc.Gaps()[0].Phase())
}

func TestModelCompleteness_CoveredPhases_ReturnsPhasesCovered(t *testing.T) {
	t.Parallel()
	// Only actors and story gaps — events and boundaries are covered
	gapActors, _ := NewModelGap(PhaseActors, "no actors")
	gapStory, _ := NewModelGap(PhaseStory, "no stories")
	mc := NewModelCompleteness([]ModelGap{gapActors, gapStory})

	covered := mc.CoveredPhases()
	// Required phases: actors, story, events, boundaries
	// Gaps in: actors, story → covered: events, boundaries
	assert.Contains(t, covered, PhaseEvents)
	assert.Contains(t, covered, PhaseBoundaries)
	assert.NotContains(t, covered, PhaseActors)
	assert.NotContains(t, covered, PhaseStory)
}
