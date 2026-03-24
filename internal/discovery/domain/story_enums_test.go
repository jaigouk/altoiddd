package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// -- StoryType tests --

func TestNewStoryType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.StoryType
	}{
		{"coarse_grained", "coarse_grained", domain.StoryTypeCoarseGrained},
		{"fine_grained", "fine_grained", domain.StoryTypeFineGrained},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			st, err := domain.NewStoryType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, st)
		})
	}
}

func TestNewStoryType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStoryType("nonsense")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStoryType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStoryType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStoryType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		st       domain.StoryType
		expected string
	}{
		{"coarse_grained", domain.StoryTypeCoarseGrained, "coarse_grained"},
		{"fine_grained", domain.StoryTypeFineGrained, "fine_grained"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.st.String())
		})
	}
}

func TestStoryType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		st   domain.StoryType
	}{
		{"coarse_grained", domain.StoryTypeCoarseGrained},
		{"fine_grained", domain.StoryTypeFineGrained},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := tt.st.MarshalText()
			require.NoError(t, err)

			var got domain.StoryType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, tt.st, got)
		})
	}
}

func TestStoryType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.StoryType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStoryType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var st domain.StoryType
	err := st.UnmarshalText([]byte("nonsense"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestStoryType_UnmarshalText_Empty(t *testing.T) {
	t.Parallel()
	var st domain.StoryType
	err := st.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllStoryTypes(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllStoryTypes(), 2)
}

// -- TimeType tests --

func TestNewTimeType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.TimeType
	}{
		{"as_is", "as_is", domain.TimeTypeAsIs},
		{"to_be", "to_be", domain.TimeTypeToBe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt2, err := domain.NewTimeType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt2)
		})
	}
}

func TestNewTimeType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewTimeType("nonsense")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewTimeType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewTimeType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestTimeType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		tt       domain.TimeType
		expected string
	}{
		{"as_is", domain.TimeTypeAsIs, "as_is"},
		{"to_be", domain.TimeTypeToBe, "to_be"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.tt.String())
		})
	}
}

func TestTimeType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tt   domain.TimeType
	}{
		{"as_is", domain.TimeTypeAsIs},
		{"to_be", domain.TimeTypeToBe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := tt.tt.MarshalText()
			require.NoError(t, err)

			var got domain.TimeType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, tt.tt, got)
		})
	}
}

func TestTimeType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.TimeType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestTimeType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var tt domain.TimeType
	err := tt.UnmarshalText([]byte("nonsense"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestTimeType_UnmarshalText_Empty(t *testing.T) {
	t.Parallel()
	var tt domain.TimeType
	err := tt.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllTimeTypes(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllTimeTypes(), 2)
}

// -- PurityType tests --

func TestNewPurityType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.PurityType
	}{
		{"pure", "pure", domain.PurityTypePure},
		{"digitalized", "digitalized", domain.PurityTypeDigitalized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pt, err := domain.NewPurityType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, pt)
		})
	}
}

func TestNewPurityType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewPurityType("nonsense")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewPurityType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewPurityType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestPurityType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pt       domain.PurityType
		expected string
	}{
		{"pure", domain.PurityTypePure, "pure"},
		{"digitalized", domain.PurityTypeDigitalized, "digitalized"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.pt.String())
		})
	}
}

func TestPurityType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		pt   domain.PurityType
	}{
		{"pure", domain.PurityTypePure},
		{"digitalized", domain.PurityTypeDigitalized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := tt.pt.MarshalText()
			require.NoError(t, err)

			var got domain.PurityType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, tt.pt, got)
		})
	}
}

func TestPurityType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.PurityType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestPurityType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var pt domain.PurityType
	err := pt.UnmarshalText([]byte("nonsense"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestPurityType_UnmarshalText_Empty(t *testing.T) {
	t.Parallel()
	var pt domain.PurityType
	err := pt.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllPurityTypes(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllPurityTypes(), 2)
}
