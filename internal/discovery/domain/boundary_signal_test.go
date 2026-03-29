package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// -- SignalType tests --

func TestNewSignalType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.SignalType
	}{
		{"different_trigger", "different_trigger", domain.SignalTypeDifferentTrigger},
		{"one_way_flow", "one_way_flow", domain.SignalTypeOneWayFlow},
		{"language_difference", "language_difference", domain.SignalTypeLanguageDifference},
		{"different_lifecycle", "different_lifecycle", domain.SignalTypeDifferentLifecycle},
		{"external_system", "external_system", domain.SignalTypeExternalSystem},
		{"different_actor", "different_actor", domain.SignalTypeDifferentActor},
		{"complex_rules", "complex_rules", domain.SignalTypeComplexRules},
		{"same_object_diff_context", "same_object_diff_context", domain.SignalTypeSameObjectDiffContext},
		{"org_boundary", "org_boundary", domain.SignalTypeOrgBoundary},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			st, err := domain.NewSignalType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, st)
		})
	}
}

func TestNewSignalType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewSignalType("nonsense")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewSignalType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewSignalType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSignalType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		st       domain.SignalType
		expected string
	}{
		{"different_trigger", domain.SignalTypeDifferentTrigger, "different_trigger"},
		{"one_way_flow", domain.SignalTypeOneWayFlow, "one_way_flow"},
		{"language_difference", domain.SignalTypeLanguageDifference, "language_difference"},
		{"different_lifecycle", domain.SignalTypeDifferentLifecycle, "different_lifecycle"},
		{"external_system", domain.SignalTypeExternalSystem, "external_system"},
		{"different_actor", domain.SignalTypeDifferentActor, "different_actor"},
		{"complex_rules", domain.SignalTypeComplexRules, "complex_rules"},
		{"same_object_diff_context", domain.SignalTypeSameObjectDiffContext, "same_object_diff_context"},
		{"org_boundary", domain.SignalTypeOrgBoundary, "org_boundary"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.st.String())
		})
	}
}

func TestSignalType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	for _, st := range domain.AllSignalTypes() {
		t.Run(st.String(), func(t *testing.T) {
			t.Parallel()
			data, err := st.MarshalText()
			require.NoError(t, err)

			var got domain.SignalType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, st, got)
		})
	}
}

func TestSignalType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.SignalType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSignalType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var st domain.SignalType
	err := st.UnmarshalText([]byte("bad"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestSignalType_UnmarshalText_Empty(t *testing.T) {
	t.Parallel()
	var st domain.SignalType
	err := st.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllSignalTypes(t *testing.T) {
	t.Parallel()
	got := domain.AllSignalTypes()
	assert.Len(t, got, 10)
	assert.Contains(t, got, domain.SignalTypeDifferentTrigger)
	assert.Contains(t, got, domain.SignalTypeOneWayFlow)
	assert.Contains(t, got, domain.SignalTypeLanguageDifference)
	assert.Contains(t, got, domain.SignalTypeDifferentLifecycle)
	assert.Contains(t, got, domain.SignalTypeExternalSystem)
	assert.Contains(t, got, domain.SignalTypeDifferentActor)
	assert.Contains(t, got, domain.SignalTypeComplexRules)
	assert.Contains(t, got, domain.SignalTypeSameObjectDiffContext)
	assert.Contains(t, got, domain.SignalTypeOrgBoundary)
	assert.Contains(t, got, domain.SignalTypeWorkObjectCluster)
}

// -- BoundarySignal tests --

func TestNewBoundarySignal_Valid(t *testing.T) {
	t.Parallel()
	bs, err := domain.NewBoundarySignal(domain.SignalTypeDifferentTrigger, "Order vs Payment triggers differ")
	require.NoError(t, err)
	assert.Equal(t, domain.SignalTypeDifferentTrigger, bs.Type())
	assert.Equal(t, "Order vs Payment triggers differ", bs.Description())
}

func TestNewBoundarySignal_AllTypes(t *testing.T) {
	t.Parallel()
	for _, st := range domain.AllSignalTypes() {
		t.Run(st.String(), func(t *testing.T) {
			t.Parallel()
			bs, err := domain.NewBoundarySignal(st, "signal description")
			require.NoError(t, err)
			assert.Equal(t, st, bs.Type())
		})
	}
}

func TestNewBoundarySignal_InvalidType(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundarySignal(domain.SignalType("invalid"), "some description")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundarySignal_EmptyDescription(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundarySignal(domain.SignalTypeDifferentTrigger, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundarySignal_WhitespaceDescription(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundarySignal(domain.SignalTypeDifferentTrigger, "   ")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundarySignal_TrimsDescription(t *testing.T) {
	t.Parallel()
	bs, err := domain.NewBoundarySignal(domain.SignalTypeDifferentTrigger, "  trimmed  ")
	require.NoError(t, err)
	assert.Equal(t, "trimmed", bs.Description())
}

func TestBoundarySignal_String(t *testing.T) {
	t.Parallel()
	bs, err := domain.NewBoundarySignal(domain.SignalTypeOneWayFlow, "Data flows one direction")
	require.NoError(t, err)
	assert.Equal(t, "BoundarySignal: [one_way_flow] Data flows one direction", bs.String())
}
