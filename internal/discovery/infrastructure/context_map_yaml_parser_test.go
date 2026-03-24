package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestContextMapYAMLParser_Parse_SampleFile(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/research/samples/context-map.yaml")
	require.NoError(t, err)

	parser := &infrastructure.ContextMapYAMLParser{}
	cm, err := parser.Parse(data)
	require.NoError(t, err)

	assert.Equal(t, "E-commerce Marketplace", cm.Project())
	assert.Len(t, cm.Contexts(), 4)
	assert.Len(t, cm.Relationships(), 4)

	// Spot-check first context.
	ctxs := cm.Contexts()
	assert.Equal(t, "Catalog", ctxs[0].Name())
	assert.Equal(t, vo.SubdomainSupporting, ctxs[0].Classification())
	assert.InDelta(t, 0.85, ctxs[0].Confidence(), 0.001)
	assert.Equal(t, []string{"Seller", "Customer"}, ctxs[0].Actors())
	assert.Equal(t, []string{"Product Listing", "Inventory"}, ctxs[0].WorkObjects())
	assert.Len(t, ctxs[0].Signals(), 2)
	assert.Equal(t, vo.UserStated, ctxs[0].Trust())

	// Spot-check first relationship.
	rels := cm.Relationships()
	assert.Equal(t, "Catalog", rels[0].Upstream())
	assert.Equal(t, "Ordering", rels[0].Downstream())
	assert.Equal(t, domain.RelationshipTypeConformist, rels[0].Type())
	assert.Equal(t, []string{"Product Listing"}, rels[0].Shared())
	assert.NotEmpty(t, rels[0].Description())
}

func TestContextMapYAMLParser_RoundTrip(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/research/samples/context-map.yaml")
	require.NoError(t, err)

	parser := &infrastructure.ContextMapYAMLParser{}

	// Parse original.
	cm1, err := parser.Parse(data)
	require.NoError(t, err)

	// Serialize.
	out, err := parser.Serialize(cm1)
	require.NoError(t, err)

	// Parse again.
	cm2, err := parser.Parse(out)
	require.NoError(t, err)

	// Compare field by field.
	assert.Equal(t, cm1.Project(), cm2.Project())

	ctxs1 := cm1.Contexts()
	ctxs2 := cm2.Contexts()
	require.Len(t, ctxs2, len(ctxs1))

	for i := range ctxs1 {
		assert.Equal(t, ctxs1[i].Name(), ctxs2[i].Name(), "context name mismatch at %d", i)
		assert.Equal(t, ctxs1[i].Classification(), ctxs2[i].Classification(), "classification mismatch at %d", i)
		assert.InDelta(t, ctxs1[i].Confidence(), ctxs2[i].Confidence(), 0.001, "confidence mismatch at %d", i)
		assert.Equal(t, ctxs1[i].Actors(), ctxs2[i].Actors(), "actors mismatch at %d", i)
		assert.Equal(t, ctxs1[i].WorkObjects(), ctxs2[i].WorkObjects(), "work objects mismatch at %d", i)
		assert.Equal(t, ctxs1[i].Trust(), ctxs2[i].Trust(), "trust mismatch at %d", i)
		assert.Equal(t, ctxs1[i].Stories(), ctxs2[i].Stories(), "stories mismatch at %d", i)

		sigs1 := ctxs1[i].Signals()
		sigs2 := ctxs2[i].Signals()
		require.Len(t, sigs2, len(sigs1), "signals count mismatch at context %d", i)

		for j := range sigs1 {
			assert.Equal(t, sigs1[j].Type(), sigs2[j].Type(), "signal type mismatch at context %d signal %d", i, j)
			assert.Equal(t, sigs1[j].Description(), sigs2[j].Description(), "signal desc mismatch at context %d signal %d", i, j)
		}
	}

	rels1 := cm1.Relationships()
	rels2 := cm2.Relationships()
	require.Len(t, rels2, len(rels1))

	for i := range rels1 {
		assert.Equal(t, rels1[i].Upstream(), rels2[i].Upstream(), "upstream mismatch at %d", i)
		assert.Equal(t, rels1[i].Downstream(), rels2[i].Downstream(), "downstream mismatch at %d", i)
		assert.Equal(t, rels1[i].Type(), rels2[i].Type(), "rel type mismatch at %d", i)
		assert.Equal(t, rels1[i].Shared(), rels2[i].Shared(), "shared mismatch at %d", i)
		assert.Equal(t, rels1[i].Description(), rels2[i].Description(), "description mismatch at %d", i)
	}
}

func TestContextMapYAMLParser_Serialize_IncludesVersion(t *testing.T) {
	t.Parallel()
	cm, err := domain.NewContextMap("TestProject", nil, nil)
	require.NoError(t, err)

	parser := &infrastructure.ContextMapYAMLParser{}
	out, err := parser.Serialize(cm)
	require.NoError(t, err)
	assert.Contains(t, string(out), "version: 1")
}

func TestContextMapYAMLParser_Parse_MissingVersion_DefaultsToOne(t *testing.T) {
	t.Parallel()
	yaml := `project: TestProject
contexts: []
relationships: []
`
	parser := &infrastructure.ContextMapYAMLParser{}
	cm, err := parser.Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Equal(t, "TestProject", cm.Project())
}

func TestContextMapYAMLParser_Parse_MissingProject(t *testing.T) {
	t.Parallel()
	yaml := `contexts: []
relationships: []
`
	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestContextMapYAMLParser_Parse_InvalidClassification(t *testing.T) {
	t.Parallel()
	yaml := `project: Test
contexts:
  - name: Foo
    classification: invalid
    confidence: 0.5
    actors: [A]
    work_objects: [W]
    boundary_signals: []
    stories: [s1]
    trust: user_stated
relationships: []
`
	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contexts[0]")
}

func TestContextMapYAMLParser_Parse_InvalidRelationshipType(t *testing.T) {
	t.Parallel()
	yaml := `project: Test
contexts: []
relationships:
  - upstream: A
    downstream: B
    type: invalid
    shared: []
`
	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "relationships[0]")
}

func TestContextMapYAMLParser_Parse_MissingContextName(t *testing.T) {
	t.Parallel()
	yaml := `project: Test
contexts:
  - classification: core
    confidence: 0.5
    actors: [A]
    work_objects: [W]
    boundary_signals: []
    stories: [s1]
    trust: user_stated
relationships: []
`
	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contexts[0]")
}

func TestContextMapYAMLParser_Parse_EmptyYAML(t *testing.T) {
	t.Parallel()
	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Parse([]byte(""))
	require.Error(t, err)
}

func TestContextMapYAMLParser_ReadWrite_RoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "context-map.yaml")

	ctx1, err := domain.NewBoundedContextSketch(
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

	cm, err := domain.NewContextMap("TestProject", []domain.BoundedContextSketch{ctx1}, []domain.ContextRelationship{rel})
	require.NoError(t, err)

	parser := &infrastructure.ContextMapYAMLParser{}
	bgCtx := context.Background()

	err = parser.Write(bgCtx, path, cm)
	require.NoError(t, err)

	cm2, err := parser.Read(bgCtx, path)
	require.NoError(t, err)
	assert.Equal(t, "TestProject", cm2.Project())
	assert.Len(t, cm2.Contexts(), 1)
	assert.Len(t, cm2.Relationships(), 1)
}

func TestContextMapYAMLParser_Read_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Read(ctx, "/nonexistent")
	require.Error(t, err)
}

func TestContextMapYAMLParser_Write_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cm, err := domain.NewContextMap("Test", nil, nil)
	require.NoError(t, err)

	parser := &infrastructure.ContextMapYAMLParser{}
	err = parser.Write(ctx, "/nonexistent", cm)
	require.Error(t, err)
}

func TestContextMapYAMLParser_Read_NonExistentFile(t *testing.T) {
	t.Parallel()

	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Read(context.Background(), "/nonexistent/context-map.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading context map file")
}

func TestContextMapYAMLParser_Parse_MalformedYAML(t *testing.T) {
	t.Parallel()

	parser := &infrastructure.ContextMapYAMLParser{}
	_, err := parser.Parse([]byte("{{invalid yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing context map YAML")
}
