package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/docimport/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

func TestNewParseWarning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		section string
		reason  string
		wantErr error
	}{
		{"valid warning", "## Bounded Contexts", "could not parse heading", nil},
		{"empty section", "", "some reason", domainerrors.ErrInvariantViolation},
		{"whitespace section", "   ", "some reason", domainerrors.ErrInvariantViolation},
		{"empty reason", "## Bounded Contexts", "", domainerrors.ErrInvariantViolation},
		{"whitespace reason", "## Bounded Contexts", "   ", domainerrors.ErrInvariantViolation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w, err := domain.NewParseWarning(tt.section, tt.reason)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.section, w.Section())
			assert.Equal(t, tt.reason, w.Reason())
		})
	}
}

func TestNewImportResult(t *testing.T) {
	t.Parallel()

	t.Run("valid with no warnings", func(t *testing.T) {
		t.Parallel()
		model := ddd.NewDomainModel("test-id")
		result, err := domain.NewImportResult(model, nil)
		require.NoError(t, err)
		assert.Equal(t, model, result.Model())
		assert.Empty(t, result.Warnings())
	})

	t.Run("valid with warnings", func(t *testing.T) {
		t.Parallel()
		model := ddd.NewDomainModel("test-id")
		w, err := domain.NewParseWarning("section", "reason")
		require.NoError(t, err)
		result, err := domain.NewImportResult(model, []domain.ParseWarning{w})
		require.NoError(t, err)
		assert.Equal(t, model, result.Model())
		assert.Len(t, result.Warnings(), 1)
	})

	t.Run("nil model returns error", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewImportResult(nil, nil)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("warnings are defensive copy", func(t *testing.T) {
		t.Parallel()
		model := ddd.NewDomainModel("test-id")
		w, err := domain.NewParseWarning("section", "reason")
		require.NoError(t, err)
		warnings := []domain.ParseWarning{w}
		result, err := domain.NewImportResult(model, warnings)
		require.NoError(t, err)
		// Mutating original slice should not affect result
		warnings[0] = domain.ParseWarning{}
		assert.Equal(t, "section", result.Warnings()[0].Section())
	})
}
