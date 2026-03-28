package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// buildTestSketch creates a BoundedContextSketch for context map builder tests.
func buildTestSketch(t *testing.T, name string, workObjects []string) domain.BoundedContextSketch {
	t.Helper()

	sketch, err := domain.NewBoundedContextSketch(
		name,
		vo.SubdomainCore,
		0.8,
		[]string{"Actor1"},
		workObjects,
		[]string{"Story1"},
		nil,
		vo.AIInferred,
	)
	require.NoError(t, err)

	return sketch
}

func TestBuildContextMap_SharedWorkObjects_InfersSharedKernel(t *testing.T) {
	t.Parallel()

	sketchA := buildTestSketch(t, "ContextA", []string{"Order", "Invoice"})
	sketchB := buildTestSketch(t, "ContextB", []string{"Invoice", "Receipt"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketchA, sketchB})

	require.NoError(t, err)
	rels := cm.Relationships()
	require.Len(t, rels, 1)
	assert.Equal(t, domain.RelationshipTypeSharedKernel, rels[0].Type())
	assert.Equal(t, []string{"Invoice"}, rels[0].Shared())
}

func TestBuildContextMap_NoSharedObjects_NoRelationship(t *testing.T) {
	t.Parallel()

	sketchA := buildTestSketch(t, "ContextA", []string{"Order"})
	sketchB := buildTestSketch(t, "ContextB", []string{"Invoice"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketchA, sketchB})

	require.NoError(t, err)
	assert.Empty(t, cm.Relationships())
}

func TestBuildContextMap_CaseInsensitiveMatch(t *testing.T) {
	t.Parallel()

	sketchA := buildTestSketch(t, "ContextA", []string{"order"})
	sketchB := buildTestSketch(t, "ContextB", []string{"Order"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketchA, sketchB})

	require.NoError(t, err)
	require.Len(t, cm.Relationships(), 1)
}

func TestBuildContextMap_LexicographicOrdering(t *testing.T) {
	t.Parallel()

	sketchZebra := buildTestSketch(t, "Zebra", []string{"Shared"})
	sketchAlpha := buildTestSketch(t, "Alpha", []string{"Shared"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketchZebra, sketchAlpha})

	require.NoError(t, err)
	rels := cm.Relationships()
	require.Len(t, rels, 1)
	// Lexicographic: Alpha < Zebra, so Alpha is upstream.
	assert.Equal(t, "Alpha", rels[0].Upstream())
	assert.Equal(t, "Zebra", rels[0].Downstream())
}

func TestBuildContextMap_MultipleSketchPairs(t *testing.T) {
	t.Parallel()

	sketchA := buildTestSketch(t, "A", []string{"X", "Y"})
	sketchB := buildTestSketch(t, "B", []string{"X"})
	sketchC := buildTestSketch(t, "C", []string{"Y"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketchA, sketchB, sketchC})

	require.NoError(t, err)
	// A-B share X, A-C share Y, B-C share nothing.
	assert.Len(t, cm.Relationships(), 2)
}

func TestBuildContextMap_EmptyProjectName_Error(t *testing.T) {
	t.Parallel()

	sketch := buildTestSketch(t, "ContextA", []string{"Order"})

	_, err := domain.BuildContextMap("", []domain.BoundedContextSketch{sketch})

	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestBuildContextMap_ZeroSketches(t *testing.T) {
	t.Parallel()

	cm, err := domain.BuildContextMap("project", nil)

	require.NoError(t, err)
	assert.Empty(t, cm.Contexts())
	assert.Empty(t, cm.Relationships())
}

func TestBuildContextMap_SingleSketch(t *testing.T) {
	t.Parallel()

	sketch := buildTestSketch(t, "OnlyContext", []string{"Order"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketch})

	require.NoError(t, err)
	assert.Len(t, cm.Contexts(), 1)
	assert.Empty(t, cm.Relationships())
}

func TestBuildContextMap_SelfRelationshipGuard(t *testing.T) {
	t.Parallel()

	sketch1 := buildTestSketch(t, "Ordering", []string{"Order"})
	sketch2 := buildTestSketch(t, "Ordering", []string{"Order"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketch1, sketch2})

	require.NoError(t, err)
	// Self-relationship must be suppressed.
	assert.Empty(t, cm.Relationships())
}

func TestBuildContextMap_BidirectionalDedup(t *testing.T) {
	t.Parallel()

	sketchA := buildTestSketch(t, "ContextA", []string{"Order"})
	sketchB := buildTestSketch(t, "ContextB", []string{"Order"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketchA, sketchB})

	require.NoError(t, err)
	// Exactly 1 relationship (not 2 — i < j loop ensures dedup).
	assert.Len(t, cm.Relationships(), 1)
}

func TestBuildContextMap_DescriptionIncludesAlphabeticalNote(t *testing.T) {
	t.Parallel()

	sketchA := buildTestSketch(t, "Alpha", []string{"SharedItem"})
	sketchB := buildTestSketch(t, "Beta", []string{"SharedItem"})

	cm, err := domain.BuildContextMap("project", []domain.BoundedContextSketch{sketchA, sketchB})

	require.NoError(t, err)
	rels := cm.Relationships()
	require.Len(t, rels, 1)
	assert.Contains(t, rels[0].Description(), "alphabetical")
}
