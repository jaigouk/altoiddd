package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestNewTrustLevel_ValidValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input int
		want  vo.TrustLevel
	}{
		{"UserStated", 1, vo.UserStated},
		{"UserConfirmed", 2, vo.UserConfirmed},
		{"AIResearched", 3, vo.AIResearched},
		{"AIInferred", 4, vo.AIInferred},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := vo.NewTrustLevel(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewTrustLevel_InvalidValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too high", 5},
		{"way too high", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := vo.NewTrustLevel(tt.input)
			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestTrustLevel_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		level vo.TrustLevel
		want  string
	}{
		{"UserStated", vo.UserStated, "user_stated"},
		{"UserConfirmed", vo.UserConfirmed, "user_confirmed"},
		{"AIResearched", vo.AIResearched, "ai_researched"},
		{"AIInferred", vo.AIInferred, "ai_inferred"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.level.String())
		})
	}
}

func TestTrustLevel_IsHigherTrust(t *testing.T) {
	t.Parallel()
	// Lower numeric value = higher trust.
	// UserStated(1) > UserConfirmed(2) > AIResearched(3) > AIInferred(4)
	tests := []struct {
		name   string
		a      vo.TrustLevel
		b      vo.TrustLevel
		expect bool
	}{
		{"UserStated > UserConfirmed", vo.UserStated, vo.UserConfirmed, true},
		{"UserStated > AIResearched", vo.UserStated, vo.AIResearched, true},
		{"UserStated > AIInferred", vo.UserStated, vo.AIInferred, true},
		{"UserConfirmed > AIResearched", vo.UserConfirmed, vo.AIResearched, true},
		{"UserConfirmed > AIInferred", vo.UserConfirmed, vo.AIInferred, true},
		{"AIResearched > AIInferred", vo.AIResearched, vo.AIInferred, true},
		{"AIInferred not > AIResearched", vo.AIInferred, vo.AIResearched, false},
		{"AIInferred not > UserStated", vo.AIInferred, vo.UserStated, false},
		{"AIResearched not > UserConfirmed", vo.AIResearched, vo.UserConfirmed, false},
		{"UserConfirmed not > UserStated", vo.UserConfirmed, vo.UserStated, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, tt.a.IsHigherTrust(tt.b))
		})
	}
}

func TestTrustLevel_IsHigherTrust_Equal(t *testing.T) {
	t.Parallel()
	for _, level := range vo.AllTrustLevels() {
		t.Run(level.String(), func(t *testing.T) {
			t.Parallel()
			assert.False(t, level.IsHigherTrust(level))
		})
	}
}

func TestTrustLevel_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid levels pass", func(t *testing.T) {
		t.Parallel()
		for _, level := range vo.AllTrustLevels() {
			require.NoError(t, level.Validate())
		}
	})

	t.Run("invalid level fails", func(t *testing.T) {
		t.Parallel()
		invalid := vo.TrustLevel(99)
		err := invalid.Validate()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

func TestTrustLevel_TextRoundTrip(t *testing.T) {
	t.Parallel()
	for _, level := range vo.AllTrustLevels() {
		t.Run(level.String(), func(t *testing.T) {
			t.Parallel()
			data, err := level.MarshalText()
			require.NoError(t, err)

			var got vo.TrustLevel
			require.NoError(t, got.UnmarshalText(data))
			assert.Equal(t, level, got)
		})
	}
}

func TestTrustLevel_UnmarshalText_InvalidString(t *testing.T) {
	t.Parallel()
	var tl vo.TrustLevel
	err := tl.UnmarshalText([]byte("bogus"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestTrustLevel_UnmarshalText_EmptyString(t *testing.T) {
	t.Parallel()
	var tl vo.TrustLevel
	err := tl.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestTrustLevel_String_InvalidValue(t *testing.T) {
	t.Parallel()
	invalid := vo.TrustLevel(99)
	assert.Equal(t, "TrustLevel(99)", invalid.String())
}

func TestTrustLevel_MarshalText_InvalidValue(t *testing.T) {
	t.Parallel()
	invalid := vo.TrustLevel(99)
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}
