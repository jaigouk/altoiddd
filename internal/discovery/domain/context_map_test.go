package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestNewContextMap_Valid(t *testing.T) {
	t.Parallel()
	ctx, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.9,
		[]string{"Customer"}, []string{"Order"}, []string{"story1"},
		nil, vo.UserStated,
	)
	require.NoError(t, err)

	rel, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"}, "desc",
	)
	require.NoError(t, err)

	cm, err := domain.NewContextMap("E-commerce", []domain.BoundedContextSketch{ctx}, []domain.ContextRelationship{rel})
	require.NoError(t, err)
	assert.Equal(t, "E-commerce", cm.Project())
	assert.Len(t, cm.Contexts(), 1)
	assert.Len(t, cm.Relationships(), 1)
}

func TestNewContextMap_EmptyProject(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextMap("", nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewContextMap_WhitespaceProject(t *testing.T) {
	t.Parallel()
	_, err := domain.NewContextMap("   ", nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewContextMap_EmptyContextsAndRelationships(t *testing.T) {
	t.Parallel()
	cm, err := domain.NewContextMap("MyProject", nil, nil)
	require.NoError(t, err)
	assert.Empty(t, cm.Contexts())
	assert.Empty(t, cm.Relationships())
}

func TestContextMap_ContextsDefensiveCopy(t *testing.T) {
	t.Parallel()
	ctx1, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.9,
		[]string{"Customer"}, []string{"Order"}, []string{"story1"},
		nil, vo.UserStated,
	)
	require.NoError(t, err)

	input := []domain.BoundedContextSketch{ctx1}
	cm, err := domain.NewContextMap("MyProject", input, nil)
	require.NoError(t, err)

	// Mutate original — should not affect the map.
	input[0] = domain.BoundedContextSketch{}
	assert.Equal(t, "Ordering", cm.Contexts()[0].Name())

	// Mutate returned slice — should not affect the map.
	returned := cm.Contexts()
	returned[0] = domain.BoundedContextSketch{}
	assert.Equal(t, "Ordering", cm.Contexts()[0].Name())
}

func TestContextMap_RelationshipsDefensiveCopy(t *testing.T) {
	t.Parallel()
	rel, err := domain.NewContextRelationship(
		"Ordering", "Payment",
		domain.RelationshipTypeCustomerSupplier,
		[]string{"Order"}, "",
	)
	require.NoError(t, err)

	input := []domain.ContextRelationship{rel}
	cm, err := domain.NewContextMap("MyProject", nil, input)
	require.NoError(t, err)

	// Mutate original — should not affect the map.
	input[0] = domain.ContextRelationship{}
	assert.Equal(t, "Ordering", cm.Relationships()[0].Upstream())

	// Mutate returned slice — should not affect the map.
	returned := cm.Relationships()
	returned[0] = domain.ContextRelationship{}
	assert.Equal(t, "Ordering", cm.Relationships()[0].Upstream())
}
