package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- StorySentence construction tests --

func TestNewStorySentence_Valid(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, 1, s.Step())
	assert.Equal(t, "Customer", s.Subject())
	assert.Equal(t, "browses", s.Activity())
	assert.Equal(t, "Product Listing", s.Object())
	assert.Equal(t, vo.UserStated, s.Trust())
	assert.Empty(t, s.Source())
}

func TestNewStorySentence_StepZero(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(0, "Customer", "browses", "Product Listing", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorySentence_NegativeStep(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(-1, "Customer", "browses", "Product Listing", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorySentence_EmptySubject(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(1, "", "browses", "Product Listing", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorySentence_WhitespaceSubject(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(1, "   ", "browses", "Product Listing", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorySentence_EmptyActivity(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(1, "Customer", "", "Product Listing", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorySentence_EmptyObject(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(1, "Customer", "browses", "", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorySentence_AIResearchedWithoutSource(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.AIResearched, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStorySentence_AIResearchedWithSource(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.AIResearched, "domain expert interview")
	require.NoError(t, err)
	assert.Equal(t, "domain expert interview", s.Source())
}

func TestNewStorySentence_InvalidTrustLevel(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.TrustLevel(99), "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

// -- WithPreposition tests --

func TestStorySentence_WithPreposition_Valid(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Receptionist", "checks", "Vet Schedule", vo.UserStated, "")
	require.NoError(t, err)

	s2, err := s.WithPreposition("for", "available slots")
	require.NoError(t, err)
	assert.Equal(t, "for", s2.Preposition())
	assert.Equal(t, "available slots", s2.IndirectObject())
	// Original is unchanged (immutable VO).
	assert.Empty(t, s.Preposition())
	assert.Empty(t, s.IndirectObject())
}

func TestStorySentence_WithPreposition_InvalidPreposition(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.UserStated, "")
	require.NoError(t, err)

	_, err = s.WithPreposition("beside", "the shelf")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStorySentence_WithPreposition_EmptyIndirectObject(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.UserStated, "")
	require.NoError(t, err)

	_, err = s.WithPreposition("for", "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStorySentence_WithPreposition_AllTen(t *testing.T) {
	t.Parallel()
	prepositions := []string{"for", "to", "via", "using", "from", "with", "in", "about", "based on", "on"}
	for _, prep := range prepositions {
		t.Run(prep, func(t *testing.T) {
			t.Parallel()
			s, err := domain.NewStorySentence(1, "Actor", "does", "Thing", vo.UserStated, "")
			require.NoError(t, err)

			s2, err := s.WithPreposition(prep, "target")
			require.NoError(t, err)
			assert.Equal(t, prep, s2.Preposition())
		})
	}
}

func TestStorySentence_WithPreposition_CaseInsensitive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{"capitalized", "For"},
		{"upper", "VIA"},
		{"mixed case multi-word", "Based On"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := domain.NewStorySentence(1, "Actor", "does", "Thing", vo.UserStated, "")
			require.NoError(t, err)

			_, err = s.WithPreposition(tt.input, "target")
			require.NoError(t, err)
		})
	}
}

func TestStorySentence_WithPreposition_StoredLowercase(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Actor", "does", "Thing", vo.UserStated, "")
	require.NoError(t, err)

	s2, err := s.WithPreposition("For", "target")
	require.NoError(t, err)
	assert.Equal(t, "for", s2.Preposition())
}

// -- HasIndirectObject tests --

func TestStorySentence_HasIndirectObject_True(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Actor", "does", "Thing", vo.UserStated, "")
	require.NoError(t, err)

	s2, err := s.WithPreposition("for", "target")
	require.NoError(t, err)
	assert.True(t, s2.HasIndirectObject())
}

func TestStorySentence_HasIndirectObject_False(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Actor", "does", "Thing", vo.UserStated, "")
	require.NoError(t, err)
	assert.False(t, s.HasIndirectObject())
}

// -- FormatText tests --

func TestStorySentence_FormatText_Simple(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "1. Customer browses Product Listing", s.FormatText())
}

func TestStorySentence_FormatText_WithPreposition(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(3, "Receptionist", "checks", "Vet Schedule", vo.UserStated, "")
	require.NoError(t, err)

	s2, err := s.WithPreposition("for", "available slots")
	require.NoError(t, err)
	assert.Equal(t, "3. Receptionist checks Vet Schedule for available slots", s2.FormatText())
}

func TestStorySentence_String(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "browses", "Product Listing", vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, s.FormatText(), s.String())
}

// -- ContainsBranching tests --

func TestStorySentence_ContainsBranching_Sometimes(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "sometimes rejects", "Order", vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, s.ContainsBranching())
}

func TestStorySentence_ContainsBranching_If(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "System", "if available sends", "Notification", vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, s.ContainsBranching())
}

func TestStorySentence_ContainsBranching_Optionally(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "User", "optionally adds", "Coupon", vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, s.ContainsBranching())
}

func TestStorySentence_ContainsBranching_Alternatively(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "User", "alternatively selects", "Option", vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, s.ContainsBranching())
}

func TestStorySentence_ContainsBranching_When(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "System", "when triggered sends", "Alert", vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, s.ContainsBranching())
}

func TestStorySentence_ContainsBranching_Unless(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Admin", "unless blocked approves", "Request", vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, s.ContainsBranching())
}

func TestStorySentence_ContainsBranching_Normal(t *testing.T) {
	t.Parallel()
	s, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	assert.False(t, s.ContainsBranching())
}

func TestStorySentence_ContainsBranching_Embedded(t *testing.T) {
	t.Parallel()
	// "sometimes-rejects" should NOT match: hyphenated word is not whitespace-bounded.
	s, err := domain.NewStorySentence(1, "Customer", "sometimes-rejects", "Order", vo.UserStated, "")
	require.NoError(t, err)
	assert.False(t, s.ContainsBranching())
}
