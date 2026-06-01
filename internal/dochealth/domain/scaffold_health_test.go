package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

func TestNewScaffoldParams_Valid_NoError(t *testing.T) {
	t.Parallel()
	p, err := NewScaffoldParams(30, nil)
	require.NoError(t, err)
	assert.Equal(t, 30, p.DefaultStalenessDays())
	assert.Empty(t, p.SecretPatterns())
}

func TestNewScaffoldParams_ZeroDays_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewScaffoldParams(0, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_NegativeDays_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewScaffoldParams(-1, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_WithPatterns_DefensiveCopy(t *testing.T) {
	t.Parallel()
	pat, err := NewSecretPattern("test", `foo`)
	require.NoError(t, err)
	original := []SecretPattern{pat}
	p, err := NewScaffoldParams(7, original)
	require.NoError(t, err)
	// Mutate caller's slice — must not affect the VO.
	original[0] = SecretPattern{}
	got := p.SecretPatterns()
	require.Len(t, got, 1)
	assert.Equal(t, "test", got[0].Name())
}

func TestNewSecretPattern_Valid_NoError(t *testing.T) {
	t.Parallel()
	pat, err := NewSecretPattern("aws", `AKIA[0-9A-Z]{16}`)
	require.NoError(t, err)
	assert.Equal(t, "aws", pat.Name())
	require.NotNil(t, pat.Pattern())
	assert.True(t, pat.Pattern().MatchString("AKIAABCDEFGHIJKLMNOP"))
}

func TestNewSecretPattern_EmptyName_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewSecretPattern("", `foo`)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewSecretPattern_BadRegex_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := NewSecretPattern("bad", `[unclosed`)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldAssetWithModTime_RoundTrips(t *testing.T) {
	t.Parallel()
	want := time.Date(2026, time.May, 30, 12, 0, 0, 0, time.UTC)
	asset, err := NewScaffoldAssetWithModTime("foo.md", nil, "", 0, false, want)
	require.NoError(t, err)
	assert.Equal(t, want, asset.ModTime())
}

func TestNewScaffoldAsset_ZeroModTime_Default(t *testing.T) {
	t.Parallel()
	// The 5-arg constructor is kept for back-compat — modTime defaults to
	// the zero value, which staleness rules interpret as "skip".
	asset, err := NewScaffoldAsset("foo.md", nil, "", 0, false)
	require.NoError(t, err)
	assert.True(t, asset.ModTime().IsZero())
}
