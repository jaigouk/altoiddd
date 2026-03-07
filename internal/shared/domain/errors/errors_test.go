package errors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

func TestErrorHierarchy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      error
		sentinel error
		name     string
	}{
		{domainerrors.ErrInvariantViolation, domainerrors.ErrInvariantViolation, "ErrInvariantViolation is an error"},
		{domainerrors.ErrNotFound, domainerrors.ErrNotFound, "ErrNotFound is an error"},
		{domainerrors.ErrAlreadyExists, domainerrors.ErrAlreadyExists, "ErrAlreadyExists is an error"},
		{domainerrors.ErrInvalidTransition, domainerrors.ErrInvalidTransition, "ErrInvalidTransition is an error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, tt.err)
			require.ErrorIs(t, tt.err, tt.sentinel)
		})
	}
}

func TestMessagePreserved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{"invariant violation message", domainerrors.ErrInvariantViolation, "invariant violation"},
		{"not found message", domainerrors.ErrNotFound, "not found"},
		{"already exists message", domainerrors.ErrAlreadyExists, "already exists"},
		{"invalid transition message", domainerrors.ErrInvalidTransition, "invalid state transition"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantMsg, tt.err.Error())
		})
	}
}

func TestDistinctSentinels(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		domainerrors.ErrInvariantViolation,
		domainerrors.ErrNotFound,
		domainerrors.ErrAlreadyExists,
		domainerrors.ErrInvalidTransition,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j {
				assert.NotErrorIs(t, a, b, "sentinel %d should not match sentinel %d", i, j)
			}
		}
	}
}

func TestWrappedErrorMatchesSentinel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sentinel error
		name     string
	}{
		{domainerrors.ErrInvariantViolation, "wrapped invariant violation"},
		{domainerrors.ErrNotFound, "wrapped not found"},
		{domainerrors.ErrAlreadyExists, "wrapped already exists"},
		{domainerrors.ErrInvalidTransition, "wrapped invalid transition"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wrapped := fmt.Errorf("context: %w", tt.sentinel)
			require.ErrorIs(t, wrapped, tt.sentinel)
			assert.NotEqual(t, tt.sentinel.Error(), wrapped.Error())
		})
	}
}

func TestWrappedErrorUnwraps(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("creating order: %w", domainerrors.ErrInvariantViolation)
	require.ErrorIs(t, wrapped, domainerrors.ErrInvariantViolation)
	assert.Contains(t, wrapped.Error(), "creating order")
}
