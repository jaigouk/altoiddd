package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// -- NarrationPhase enum tests --

func TestNewNarrationPhase_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.NarrationPhase
	}{
		{"opening", "opening", domain.NarrationPhaseOpening},
		{"narration", "narration", domain.NarrationPhaseNarration},
		{"deepening", "deepening", domain.NarrationPhaseDeepening},
		{"closing", "closing", domain.NarrationPhaseClosing},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			np, err := domain.NewNarrationPhase(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, np)
		})
	}
}

func TestNewNarrationPhase_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{"nonsense", "nonsense"},
		{"empty", ""},
		{"uppercase", "Opening"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := domain.NewNarrationPhase(tt.input)
			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestNarrationPhase_Validate_AllValid(t *testing.T) {
	t.Parallel()
	for _, phase := range domain.AllNarrationPhases() {
		t.Run(phase.String(), func(t *testing.T) {
			t.Parallel()
			require.NoError(t, phase.Validate())
		})
	}
}

func TestNarrationPhase_Validate_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.NarrationPhase("bad")
	require.ErrorIs(t, invalid.Validate(), domainerrors.ErrInvariantViolation)
}

func TestNarrationPhase_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		phase    domain.NarrationPhase
		expected string
	}{
		{"opening", domain.NarrationPhaseOpening, "opening"},
		{"narration", domain.NarrationPhaseNarration, "narration"},
		{"deepening", domain.NarrationPhaseDeepening, "deepening"},
		{"closing", domain.NarrationPhaseClosing, "closing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.phase.String())
		})
	}
}

func TestNarrationPhase_MarshalText_Valid(t *testing.T) {
	t.Parallel()
	for _, phase := range domain.AllNarrationPhases() {
		t.Run(phase.String(), func(t *testing.T) {
			t.Parallel()
			data, err := phase.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, phase.String(), string(data))
		})
	}
}

func TestNarrationPhase_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.NarrationPhase("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNarrationPhase_UnmarshalText_Valid(t *testing.T) {
	t.Parallel()
	for _, phase := range domain.AllNarrationPhases() {
		t.Run(phase.String(), func(t *testing.T) {
			t.Parallel()
			var got domain.NarrationPhase
			err := got.UnmarshalText([]byte(phase.String()))
			require.NoError(t, err)
			assert.Equal(t, phase, got)
		})
	}
}

func TestNarrationPhase_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{"nonsense", "nonsense"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var np domain.NarrationPhase
			err := np.UnmarshalText([]byte(tt.input))
			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestAllNarrationPhases_Returns4(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllNarrationPhases(), 4)
}

// -- ModeratorElicits enum tests --

func TestNewModeratorElicits_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.ModeratorElicits
	}{
		{"actor", "actor", domain.ModeratorElicitsActor},
		{"sentence", "sentence", domain.ModeratorElicitsSentence},
		{"annotation", "annotation", domain.ModeratorElicitsAnnotation},
		{"done", "done", domain.ModeratorElicitsDone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			me, err := domain.NewModeratorElicits(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, me)
		})
	}
}

func TestNewModeratorElicits_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{"nonsense", "nonsense"},
		{"empty", ""},
		{"uppercase", "Actor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := domain.NewModeratorElicits(tt.input)
			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestModeratorElicits_Validate_AllValid(t *testing.T) {
	t.Parallel()
	for _, elicits := range domain.AllModeratorElicits() {
		t.Run(elicits.String(), func(t *testing.T) {
			t.Parallel()
			require.NoError(t, elicits.Validate())
		})
	}
}

func TestModeratorElicits_Validate_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.ModeratorElicits("bad")
	require.ErrorIs(t, invalid.Validate(), domainerrors.ErrInvariantViolation)
}

func TestModeratorElicits_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		elicits  domain.ModeratorElicits
		expected string
	}{
		{"actor", domain.ModeratorElicitsActor, "actor"},
		{"sentence", domain.ModeratorElicitsSentence, "sentence"},
		{"annotation", domain.ModeratorElicitsAnnotation, "annotation"},
		{"done", domain.ModeratorElicitsDone, "done"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.elicits.String())
		})
	}
}

func TestModeratorElicits_MarshalText_Valid(t *testing.T) {
	t.Parallel()
	for _, elicits := range domain.AllModeratorElicits() {
		t.Run(elicits.String(), func(t *testing.T) {
			t.Parallel()
			data, err := elicits.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, elicits.String(), string(data))
		})
	}
}

func TestModeratorElicits_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.ModeratorElicits("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestModeratorElicits_UnmarshalText_Valid(t *testing.T) {
	t.Parallel()
	for _, elicits := range domain.AllModeratorElicits() {
		t.Run(elicits.String(), func(t *testing.T) {
			t.Parallel()
			var got domain.ModeratorElicits
			err := got.UnmarshalText([]byte(elicits.String()))
			require.NoError(t, err)
			assert.Equal(t, elicits, got)
		})
	}
}

func TestModeratorElicits_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{"nonsense", "nonsense"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var me domain.ModeratorElicits
			err := me.UnmarshalText([]byte(tt.input))
			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestAllModeratorElicits_Returns4(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllModeratorElicits(), 4)
}

// -- ModeratorQuestion VO tests --

func TestNewModeratorQuestion_Valid(t *testing.T) {
	t.Parallel()
	q, err := domain.NewModeratorQuestion(
		"MQ-O1",
		domain.NarrationPhaseOpening,
		"Who is the primary actor?",
		"Who starts this process?",
		domain.ModeratorElicitsActor,
	)
	require.NoError(t, err)
	assert.Equal(t, "MQ-O1", q.ID())
	assert.Equal(t, domain.NarrationPhaseOpening, q.Phase())
	assert.Equal(t, "Who is the primary actor?", q.TechnicalText())
	assert.Equal(t, "Who starts this process?", q.NonTechnicalText())
	assert.Equal(t, domain.ModeratorElicitsActor, q.Elicits())
}

func TestNewModeratorQuestion_EmptyID_Error(t *testing.T) {
	t.Parallel()
	_, err := domain.NewModeratorQuestion(
		"",
		domain.NarrationPhaseOpening,
		"tech text",
		"plain text",
		domain.ModeratorElicitsActor,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewModeratorQuestion_WhitespaceID_Error(t *testing.T) {
	t.Parallel()
	_, err := domain.NewModeratorQuestion(
		"   ",
		domain.NarrationPhaseOpening,
		"tech text",
		"plain text",
		domain.ModeratorElicitsActor,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewModeratorQuestion_InvalidPhase_Error(t *testing.T) {
	t.Parallel()
	_, err := domain.NewModeratorQuestion(
		"MQ-X1",
		domain.NarrationPhase("invalid"),
		"tech text",
		"plain text",
		domain.ModeratorElicitsActor,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewModeratorQuestion_EmptyTechnicalText_Error(t *testing.T) {
	t.Parallel()
	_, err := domain.NewModeratorQuestion(
		"MQ-O1",
		domain.NarrationPhaseOpening,
		"",
		"plain text",
		domain.ModeratorElicitsActor,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewModeratorQuestion_EmptyNonTechnicalText_Error(t *testing.T) {
	t.Parallel()
	_, err := domain.NewModeratorQuestion(
		"MQ-O1",
		domain.NarrationPhaseOpening,
		"tech text",
		"",
		domain.ModeratorElicitsActor,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewModeratorQuestion_InvalidElicits_Error(t *testing.T) {
	t.Parallel()
	_, err := domain.NewModeratorQuestion(
		"MQ-O1",
		domain.NarrationPhaseOpening,
		"tech text",
		"plain text",
		domain.ModeratorElicits("invalid"),
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestModeratorQuestion_Text_Technical(t *testing.T) {
	t.Parallel()
	q, err := domain.NewModeratorQuestion(
		"MQ-O1",
		domain.NarrationPhaseOpening,
		"technical phrasing",
		"plain phrasing",
		domain.ModeratorElicitsActor,
	)
	require.NoError(t, err)
	assert.Equal(t, "technical phrasing", q.Text(domain.RegisterTechnical))
}

func TestModeratorQuestion_Text_NonTechnical(t *testing.T) {
	t.Parallel()
	q, err := domain.NewModeratorQuestion(
		"MQ-O1",
		domain.NarrationPhaseOpening,
		"technical phrasing",
		"plain phrasing",
		domain.ModeratorElicitsActor,
	)
	require.NoError(t, err)
	assert.Equal(t, "plain phrasing", q.Text(domain.RegisterNonTechnical))
}

func TestModeratorQuestion_Getters(t *testing.T) {
	t.Parallel()
	q, err := domain.NewModeratorQuestion(
		"MQ-D1",
		domain.NarrationPhaseDeepening,
		"Are there invariants?",
		"Are there rules?",
		domain.ModeratorElicitsAnnotation,
	)
	require.NoError(t, err)
	assert.Equal(t, "MQ-D1", q.ID())
	assert.Equal(t, domain.NarrationPhaseDeepening, q.Phase())
	assert.Equal(t, "Are there invariants?", q.TechnicalText())
	assert.Equal(t, "Are there rules?", q.NonTechnicalText())
	assert.Equal(t, domain.ModeratorElicitsAnnotation, q.Elicits())
}

func TestModeratorQuestion_String(t *testing.T) {
	t.Parallel()
	q, err := domain.NewModeratorQuestion(
		"MQ-O1",
		domain.NarrationPhaseOpening,
		"tech",
		"plain",
		domain.ModeratorElicitsActor,
	)
	require.NoError(t, err)
	assert.Equal(t, "MQ-O1", q.String())
}

// -- Question Bank tests --

func TestModeratorQuestionBank_NotEmpty(t *testing.T) {
	t.Parallel()
	bank := domain.ModeratorQuestionBank()
	assert.NotEmpty(t, bank)
}

func TestModeratorQuestionBank_AllPhasesRepresented(t *testing.T) {
	t.Parallel()
	bank := domain.ModeratorQuestionBank()
	phases := make(map[domain.NarrationPhase]bool)
	for _, q := range bank {
		phases[q.Phase()] = true
	}
	for _, phase := range domain.AllNarrationPhases() {
		assert.True(t, phases[phase], "phase %s not represented in bank", phase)
	}
}

func TestModeratorQuestionBank_UniqueIDs(t *testing.T) {
	t.Parallel()
	bank := domain.ModeratorQuestionBank()
	seen := make(map[string]bool)
	for _, q := range bank {
		assert.False(t, seen[q.ID()], "duplicate ID: %s", q.ID())
		seen[q.ID()] = true
	}
}

func TestModeratorQuestionBank_DefensiveCopy(t *testing.T) {
	t.Parallel()
	bank1 := domain.ModeratorQuestionBank()
	bank2 := domain.ModeratorQuestionBank()
	require.NotEmpty(t, bank1)
	// Mutating one slice should not affect the other.
	bank1[0] = domain.ModeratorQuestion{}
	assert.NotEqual(t, bank1[0].ID(), bank2[0].ID())
}

func TestQuestionsByPhase_Opening_AtLeast3(t *testing.T) {
	t.Parallel()
	qs := domain.QuestionsByPhase(domain.NarrationPhaseOpening)
	assert.GreaterOrEqual(t, len(qs), 3)
	for _, q := range qs {
		assert.Equal(t, domain.NarrationPhaseOpening, q.Phase())
	}
}

func TestQuestionsByPhase_Narration_AtLeast3(t *testing.T) {
	t.Parallel()
	qs := domain.QuestionsByPhase(domain.NarrationPhaseNarration)
	assert.GreaterOrEqual(t, len(qs), 3)
	for _, q := range qs {
		assert.Equal(t, domain.NarrationPhaseNarration, q.Phase())
	}
}

func TestQuestionsByPhase_Deepening_AtLeast3(t *testing.T) {
	t.Parallel()
	qs := domain.QuestionsByPhase(domain.NarrationPhaseDeepening)
	assert.GreaterOrEqual(t, len(qs), 3)
	for _, q := range qs {
		assert.Equal(t, domain.NarrationPhaseDeepening, q.Phase())
	}
}

func TestQuestionsByPhase_Closing_AtLeast3(t *testing.T) {
	t.Parallel()
	qs := domain.QuestionsByPhase(domain.NarrationPhaseClosing)
	assert.GreaterOrEqual(t, len(qs), 3)
	for _, q := range qs {
		assert.Equal(t, domain.NarrationPhaseClosing, q.Phase())
	}
}

func TestQuestionsByPhase_Unknown_EmptySlice(t *testing.T) {
	t.Parallel()
	qs := domain.QuestionsByPhase(domain.NarrationPhase("unknown"))
	assert.Empty(t, qs)
}

func TestModeratorQuestionByID_Found(t *testing.T) {
	t.Parallel()
	q, ok := domain.ModeratorQuestionByID("MQ-O1")
	assert.True(t, ok)
	assert.Equal(t, "MQ-O1", q.ID())
}

func TestModeratorQuestionByID_NotFound(t *testing.T) {
	t.Parallel()
	_, ok := domain.ModeratorQuestionByID("NONEXISTENT")
	assert.False(t, ok)
}

func TestElicitsTagsMatchPhase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		phase   domain.NarrationPhase
		elicits domain.ModeratorElicits
	}{
		{"opening elicits actor", domain.NarrationPhaseOpening, domain.ModeratorElicitsActor},
		{"narration elicits sentence", domain.NarrationPhaseNarration, domain.ModeratorElicitsSentence},
		{"deepening elicits annotation", domain.NarrationPhaseDeepening, domain.ModeratorElicitsAnnotation},
		{"closing elicits done", domain.NarrationPhaseClosing, domain.ModeratorElicitsDone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			qs := domain.QuestionsByPhase(tt.phase)
			require.NotEmpty(t, qs)
			for _, q := range qs {
				assert.Equal(t, tt.elicits, q.Elicits(),
					"question %s in phase %s has wrong elicits", q.ID(), tt.phase)
			}
		})
	}
}
