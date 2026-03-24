package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// -- RelationshipType tests --

func TestNewRelationshipType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.RelationshipType
	}{
		{"shared_kernel", "shared_kernel", domain.RelationshipTypeSharedKernel},
		{"customer_supplier", "customer_supplier", domain.RelationshipTypeCustomerSupplier},
		{"conformist", "conformist", domain.RelationshipTypeConformist},
		{"anticorruption_layer", "anticorruption_layer", domain.RelationshipTypeAnticorruptionLayer},
		{"open_host_service", "open_host_service", domain.RelationshipTypeOpenHostService},
		{"published_language", "published_language", domain.RelationshipTypePublishedLanguage},
		{"partnership", "partnership", domain.RelationshipTypePartnership},
		{"separate_ways", "separate_ways", domain.RelationshipTypeSeparateWays},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rt, err := domain.NewRelationshipType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, rt)
		})
	}
}

func TestNewRelationshipType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewRelationshipType("nonsense")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewRelationshipType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewRelationshipType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestRelationshipType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		rt       domain.RelationshipType
		expected string
	}{
		{"shared_kernel", domain.RelationshipTypeSharedKernel, "shared_kernel"},
		{"customer_supplier", domain.RelationshipTypeCustomerSupplier, "customer_supplier"},
		{"conformist", domain.RelationshipTypeConformist, "conformist"},
		{"anticorruption_layer", domain.RelationshipTypeAnticorruptionLayer, "anticorruption_layer"},
		{"open_host_service", domain.RelationshipTypeOpenHostService, "open_host_service"},
		{"published_language", domain.RelationshipTypePublishedLanguage, "published_language"},
		{"partnership", domain.RelationshipTypePartnership, "partnership"},
		{"separate_ways", domain.RelationshipTypeSeparateWays, "separate_ways"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.rt.String())
		})
	}
}

func TestRelationshipType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	for _, rt := range domain.AllRelationshipTypes() {
		t.Run(rt.String(), func(t *testing.T) {
			t.Parallel()
			data, err := rt.MarshalText()
			require.NoError(t, err)

			var got domain.RelationshipType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, rt, got)
		})
	}
}

func TestRelationshipType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.RelationshipType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestRelationshipType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var rt domain.RelationshipType
	err := rt.UnmarshalText([]byte("nonsense"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestRelationshipType_UnmarshalText_Empty(t *testing.T) {
	t.Parallel()
	var rt domain.RelationshipType
	err := rt.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllRelationshipTypes(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllRelationshipTypes(), 8)
}
