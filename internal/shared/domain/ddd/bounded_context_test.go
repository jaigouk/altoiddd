package ddd_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

func TestNewBoundedContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctxName    string
		aggregates []string
		wantErr    error
		wantName   string
		wantAggs   []string
	}{
		{
			name:       "valid with aggregates",
			ctxName:    "Orders",
			aggregates: []string{"Order", "LineItem"},
			wantErr:    nil,
			wantName:   "Orders",
			wantAggs:   []string{"Order", "LineItem"},
		},
		{
			name:       "valid with nil aggregates",
			ctxName:    "Shipping",
			aggregates: nil,
			wantErr:    nil,
			wantName:   "Shipping",
			wantAggs:   []string{},
		},
		{
			name:       "valid with empty aggregates slice",
			ctxName:    "Billing",
			aggregates: []string{},
			wantErr:    nil,
			wantName:   "Billing",
			wantAggs:   []string{},
		},
		{
			name:       "empty name",
			ctxName:    "",
			aggregates: nil,
			wantErr:    domainerrors.ErrInvariantViolation,
		},
		{
			name:       "whitespace-only name",
			ctxName:    "   ",
			aggregates: nil,
			wantErr:    domainerrors.ErrInvariantViolation,
		},
		{
			name:       "duplicate aggregates",
			ctxName:    "Orders",
			aggregates: []string{"Order", "Order"},
			wantErr:    domainerrors.ErrInvariantViolation,
		},
		{
			name:       "very long name is valid",
			ctxName:    strings.Repeat("A", 1000),
			aggregates: nil,
			wantErr:    nil,
			wantName:   strings.Repeat("A", 1000),
			wantAggs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bc, err := ddd.NewBoundedContext(tt.ctxName, tt.aggregates)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, bc)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, bc)
			assert.Equal(t, tt.wantName, bc.Name())
			assert.Equal(t, tt.wantAggs, bc.Aggregates())
		})
	}
}

func TestBoundedContext_AggregatesDefensiveCopy(t *testing.T) {
	t.Parallel()

	bc, err := ddd.NewBoundedContext("Orders", []string{"Order", "LineItem"})
	require.NoError(t, err)

	// Modifying the returned slice must not affect the original.
	aggs := bc.Aggregates()
	aggs[0] = "MODIFIED"

	assert.Equal(t, "Order", bc.Aggregates()[0], "Aggregates() must return a defensive copy")
}
