package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- AnnotationType tests --

func TestNewAnnotationType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.AnnotationType
	}{
		{"constraint", "constraint", domain.AnnotationTypeConstraint},
		{"invariant", "invariant", domain.AnnotationTypeInvariant},
		{"assumption", "assumption", domain.AnnotationTypeAssumption},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			at, err := domain.NewAnnotationType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, at)
		})
	}
}

func TestNewAnnotationType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAnnotationType("rule")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotationType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAnnotationType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAnnotationType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		at       domain.AnnotationType
		expected string
	}{
		{"constraint", domain.AnnotationTypeConstraint, "constraint"},
		{"invariant", domain.AnnotationTypeInvariant, "invariant"},
		{"assumption", domain.AnnotationTypeAssumption, "assumption"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.at.String())
		})
	}
}

func TestAnnotationType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		at   domain.AnnotationType
	}{
		{"constraint", domain.AnnotationTypeConstraint},
		{"invariant", domain.AnnotationTypeInvariant},
		{"assumption", domain.AnnotationTypeAssumption},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := tt.at.MarshalText()
			require.NoError(t, err)

			var got domain.AnnotationType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, tt.at, got)
		})
	}
}

func TestAnnotationType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.AnnotationType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAnnotationType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var at domain.AnnotationType
	err := at.UnmarshalText([]byte("bad"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllAnnotationTypes(t *testing.T) {
	t.Parallel()
	got := domain.AllAnnotationTypes()
	assert.Len(t, got, 3)
	assert.Contains(t, got, domain.AnnotationTypeConstraint)
	assert.Contains(t, got, domain.AnnotationTypeInvariant)
	assert.Contains(t, got, domain.AnnotationTypeAssumption)
}

// -- Annotation tests --

func TestNewAnnotation_Constraint(t *testing.T) {
	t.Parallel()
	ann, err := domain.NewAnnotation("Must be authenticated", domain.AnnotationTypeConstraint, nil, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Must be authenticated", ann.Text())
	assert.Equal(t, domain.AnnotationTypeConstraint, ann.Type())
	assert.Nil(t, ann.SentenceRef())
	assert.Equal(t, vo.UserStated, ann.Trust())
	assert.Empty(t, ann.Source())
}

func TestNewAnnotation_Invariant(t *testing.T) {
	t.Parallel()
	ann, err := domain.NewAnnotation("Payment before order", domain.AnnotationTypeInvariant, nil, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Payment before order", ann.Text())
	assert.Equal(t, domain.AnnotationTypeInvariant, ann.Type())
}

func TestNewAnnotation_Assumption(t *testing.T) {
	t.Parallel()
	ann, err := domain.NewAnnotation("Users prefer email", domain.AnnotationTypeAssumption, nil, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Users prefer email", ann.Text())
	assert.Equal(t, domain.AnnotationTypeAssumption, ann.Type())
}

func TestNewAnnotation_EmptyText(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAnnotation("", domain.AnnotationTypeConstraint, nil, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotation_WhitespaceText(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAnnotation("   ", domain.AnnotationTypeConstraint, nil, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotation_InvalidType(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAnnotation("Some text", domain.AnnotationType("invalid"), nil, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotation_InvalidTrustLevel(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAnnotation("Some text", domain.AnnotationTypeConstraint, nil, vo.TrustLevel(99), "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotation_StoryWide(t *testing.T) {
	t.Parallel()
	ann, err := domain.NewAnnotation("Must be authenticated", domain.AnnotationTypeConstraint, nil, vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, ann.IsStoryWide())
	assert.Nil(t, ann.SentenceRef())
}

func TestNewAnnotation_SentenceSpecific(t *testing.T) {
	t.Parallel()
	ref := 3
	ann, err := domain.NewAnnotation("Payment before order", domain.AnnotationTypeInvariant, &ref, vo.UserStated, "")
	require.NoError(t, err)
	assert.False(t, ann.IsStoryWide())
	require.NotNil(t, ann.SentenceRef())
	assert.Equal(t, 3, *ann.SentenceRef())

	// Verify constructor defensive copy: mutating original does not affect stored.
	ref = 99
	assert.Equal(t, 3, *ann.SentenceRef())

	// Verify getter defensive copy: mutating returned pointer does not affect stored.
	ptr := ann.SentenceRef()
	*ptr = 42
	assert.Equal(t, 3, *ann.SentenceRef())
}

func TestNewAnnotation_SentenceRefZero(t *testing.T) {
	t.Parallel()
	ref := 0
	_, err := domain.NewAnnotation("Some text", domain.AnnotationTypeConstraint, &ref, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotation_NegativeSentenceRef(t *testing.T) {
	t.Parallel()
	ref := -1
	_, err := domain.NewAnnotation("Some text", domain.AnnotationTypeConstraint, &ref, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotation_AIResearchedWithoutSource(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAnnotation("Some text", domain.AnnotationTypeConstraint, nil, vo.AIResearched, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewAnnotation_AIResearchedWithSource(t *testing.T) {
	t.Parallel()
	ann, err := domain.NewAnnotation("Some text", domain.AnnotationTypeConstraint, nil, vo.AIResearched, "research-doc")
	require.NoError(t, err)
	assert.Equal(t, "research-doc", ann.Source())
	assert.Equal(t, vo.AIResearched, ann.Trust())
}

func TestNewAnnotation_UserStatedWithoutSource(t *testing.T) {
	t.Parallel()
	ann, err := domain.NewAnnotation("Some text", domain.AnnotationTypeConstraint, nil, vo.UserStated, "")
	require.NoError(t, err)
	assert.Empty(t, ann.Source())
}

func TestAnnotation_String_StoryWide(t *testing.T) {
	t.Parallel()
	ann, err := domain.NewAnnotation("Must be authenticated", domain.AnnotationTypeConstraint, nil, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Annotation: [constraint] Must be authenticated (story-wide, user_stated)", ann.String())
}

func TestAnnotation_String_SentenceSpecific(t *testing.T) {
	t.Parallel()
	ref := 3
	ann, err := domain.NewAnnotation("Payment before order", domain.AnnotationTypeInvariant, &ref, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Annotation: [invariant] Payment before order (sentence 3, user_stated)", ann.String())
}
