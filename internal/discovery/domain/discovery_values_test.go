package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -- Persona enum tests --

func TestPersonaHasFourMembers(t *testing.T) {
	t.Parallel()
	assert.Len(t, AllDiscoveryPersonas(), 4)
}

func TestPersonaValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		persona  DiscoveryPersona
		expected string
	}{
		{"developer", PersonaDeveloper, "developer"},
		{"product_owner", PersonaProductOwner, "product_owner"},
		{"domain_expert", PersonaDomainExpert, "domain_expert"},
		{"mixed", PersonaMixed, "mixed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.persona))
		})
	}
}

// -- Register enum tests --

func TestRegisterHasTwoMembers(t *testing.T) {
	t.Parallel()
	assert.Len(t, AllRegisters(), 2)
}

func TestRegisterValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		register DiscoveryRegister
		expected string
	}{
		{"technical", RegisterTechnical, "technical"},
		{"non_technical", RegisterNonTechnical, "non_technical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.register))
		})
	}
}

// -- QuestionPhase enum tests --

func TestQuestionPhaseHasFivePhases(t *testing.T) {
	t.Parallel()
	assert.Len(t, AllQuestionPhases(), 5)
}

func TestQuestionPhaseValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		phase    QuestionPhase
		expected string
	}{
		{"seed", PhaseSeed, "seed"},
		{"actors", PhaseActors, "actors"},
		{"story", PhaseStory, "story"},
		{"events", PhaseEvents, "events"},
		{"boundaries", PhaseBoundaries, "boundaries"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.phase))
		})
	}
}

func TestQuestionPhaseOrdering(t *testing.T) {
	t.Parallel()
	phases := AllQuestionPhases()
	expected := []QuestionPhase{PhaseSeed, PhaseActors, PhaseStory, PhaseEvents, PhaseBoundaries}
	assert.Equal(t, expected, phases)
}

// -- DiscoveryMode enum tests --

func TestDiscoveryModeValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "express", string(ModeExpress))
	assert.Equal(t, "deep", string(ModeDeep))
}

func TestDiscoveryModeHasTwoMembers(t *testing.T) {
	t.Parallel()
	assert.Len(t, AllDiscoveryModes(), 2)
}

// -- DiscoveryRound enum tests --

func TestDiscoveryRoundValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "discovery", string(RoundDiscovery))
	assert.Equal(t, "challenge", string(RoundChallenge))
	assert.Equal(t, "simulate", string(RoundSimulate))
}

func TestDiscoveryRoundHasThreeMembers(t *testing.T) {
	t.Parallel()
	assert.Len(t, AllDiscoveryRounds(), 3)
}

// -- Answer value object tests --

func TestAnswerCreation(t *testing.T) {
	t.Parallel()
	a := NewAnswer("Q1", "Users and admins")
	assert.Equal(t, "Q1", a.QuestionID())
	assert.Equal(t, "Users and admins", a.ResponseText())
}

func TestAnswerEquality(t *testing.T) {
	t.Parallel()
	a1 := NewAnswer("Q1", "Users and admins")
	a2 := NewAnswer("Q1", "Users and admins")
	assert.True(t, a1.Equal(a2))
}

func TestAnswerInequality(t *testing.T) {
	t.Parallel()
	a1 := NewAnswer("Q1", "Users and admins")
	a2 := NewAnswer("Q1", "Just admins")
	assert.False(t, a1.Equal(a2))
}

// -- Playback value object tests --

func TestPlaybackCreationDefaults(t *testing.T) {
	t.Parallel()
	pb := NewPlayback("Summary here", false, "")
	assert.Equal(t, "Summary here", pb.SummaryText())
	assert.False(t, pb.Confirmed())
	assert.Empty(t, pb.Corrections())
}

func TestPlaybackCreationWithAllFields(t *testing.T) {
	t.Parallel()
	pb := NewPlayback("Sum", true, "Fix actors")
	assert.True(t, pb.Confirmed())
	assert.Equal(t, "Fix actors", pb.Corrections())
}

func TestPlaybackEquality(t *testing.T) {
	t.Parallel()
	p1 := NewPlayback("Sum", true, "")
	p2 := NewPlayback("Sum", true, "")
	assert.True(t, p1.Equal(p2))
}

// -- ParseDiscoveryPersona tests --

func TestParseDiscoveryPersonaValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected DiscoveryPersona
	}{
		{"developer", PersonaDeveloper},
		{"product_owner", PersonaProductOwner},
		{"domain_expert", PersonaDomainExpert},
		{"mixed", PersonaMixed},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			p, err := ParseDiscoveryPersona(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, p)
		})
	}
}

func TestParseDiscoveryPersonaInvalid(t *testing.T) {
	t.Parallel()
	_, err := ParseDiscoveryPersona("alien")
	require.Error(t, err)
}

// -- ParseDiscoveryRegister tests --

func TestParseDiscoveryRegisterValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected DiscoveryRegister
	}{
		{"technical", RegisterTechnical},
		{"non_technical", RegisterNonTechnical},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			r, err := ParseDiscoveryRegister(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, r)
		})
	}
}

func TestParseDiscoveryRegisterInvalid(t *testing.T) {
	t.Parallel()
	_, err := ParseDiscoveryRegister("casual")
	require.Error(t, err)
}

// -- ParseDiscoveryMode tests --

func TestParseDiscoveryModeValid(t *testing.T) {
	t.Parallel()
	m, err := ParseDiscoveryMode("deep")
	require.NoError(t, err)
	assert.Equal(t, ModeDeep, m)
}

func TestParseDiscoveryModeInvalid(t *testing.T) {
	t.Parallel()
	_, err := ParseDiscoveryMode("turbo")
	require.Error(t, err)
}

// -- ParseDiscoveryRound tests --

func TestParseDiscoveryRoundValid(t *testing.T) {
	t.Parallel()
	r, err := ParseDiscoveryRound("challenge")
	require.NoError(t, err)
	assert.Equal(t, RoundChallenge, r)
}

func TestParseDiscoveryRoundInvalid(t *testing.T) {
	t.Parallel()
	_, err := ParseDiscoveryRound("unknown")
	require.Error(t, err)
}
