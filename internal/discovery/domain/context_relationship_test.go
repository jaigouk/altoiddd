package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

func TestNewContextRelationship_Valid(t *testing.T) {
	t.Parallel()
	cr, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"},
		"Payment serves Ordering's needs",
	)
	require.NoError(t, err)
	assert.Equal(t, "Ordering", cr.Upstream())
	assert.Equal(t, "Payment", cr.Downstream())
	assert.Equal(t, domain.RelationshipTypeCustomerSupplier, cr.Type())
	assert.Equal(t, []string{"Order"}, cr.Shared())
	assert.Equal(t, "Payment serves Ordering's needs", cr.Description())
}

func TestNewContextRelationship_EmptyUpstream(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextRelationship(
		"", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"}, "",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewContextRelationship_WhitespaceUpstream(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextRelationship(
		"   ", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"}, "",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewContextRelationship_EmptyDownstream(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextRelationship(
		"Ordering", "",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"}, "",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewContextRelationship_InvalidType(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipType("invalid"),
		[]string{"Order"}, "",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestContextRelationship_SharedDefensiveCopy(t *testing.T) {
	t.Parallel()
	input := []string{"Order", "Payment"}
	cr, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		input, "",
	)
	require.NoError(t, err)

	// Mutate original — should not affect VO.
	input[0] = "MUTATED"
	assert.Equal(t, "Order", cr.Shared()[0])

	// Mutate returned slice — should not affect VO.
	returned := cr.Shared()
	returned[0] = "MUTATED"
	assert.Equal(t, "Order", cr.Shared()[0])
}

func TestNewContextRelationship_EmptyShared(t *testing.T) {
	t.Parallel()
	cr, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		nil, "",
	)
	require.NoError(t, err)
	assert.Empty(t, cr.Shared())
}

func TestNewContextRelationship_EmptyDescription(t *testing.T) {
	t.Parallel()
	cr, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"}, "",
	)
	require.NoError(t, err)
	assert.Empty(t, cr.Description())
}

func TestContextRelationship_String(t *testing.T) {
	t.Parallel()
	cr, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"}, "",
	)
	require.NoError(t, err)
	s := cr.String()
	assert.Contains(t, s, "Ordering")
	assert.Contains(t, s, "Payment")
	assert.Contains(t, s, "customer_supplier")
}
