package infrastructure_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Compile-time interface check.
var _ application.ArtifactRenderer = (*infrastructure.MarkdownArtifactRenderer)(nil)

func sampleModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("test-model")

	require.NoError(t, model.AddBoundedContext(
		vo.NewDomainBoundedContext("Ordering", "Manages order lifecycle", nil, nil, "")))
	require.NoError(t, model.AddBoundedContext(
		vo.NewDomainBoundedContext("Shipping", "Handles delivery logistics", nil, nil, "")))

	require.NoError(t, model.AddDomainStory(
		vo.NewDomainStory("Place Order", []string{"Customer"}, "Customer submits cart",
			[]string{"Customer places order", "System validates order"},
			[]string{"Order must have at least one item"})))
	require.NoError(t, model.AddDomainStory(
		vo.NewDomainStory("Ship Order", []string{"Warehouse"}, "Order is paid",
			[]string{"Warehouse picks shipment", "Carrier delivers shipment"}, nil)))

	require.NoError(t, model.AddTerm("Order", "A customer purchase request", "Ordering", []string{"Q2"}))
	require.NoError(t, model.AddTerm("Shipment", "A delivery package", "Shipping", []string{"Q2"}))

	require.NoError(t, model.ClassifySubdomain("Ordering", vo.SubdomainCore, "Competitive advantage"))
	require.NoError(t, model.ClassifySubdomain("Shipping", vo.SubdomainSupporting, "Necessary logistics"))

	require.NoError(t, model.DesignAggregate(
		vo.NewAggregateDesign("OrderRoot", "Ordering", "OrderRoot", nil,
			[]string{"Order must have items"}, nil, []string{"OrderPlaced"})))

	require.NoError(t, model.Finalize())
	return model
}

func minimalModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("minimal")
	require.NoError(t, model.AddBoundedContext(
		vo.NewDomainBoundedContext("Core", "Main domain", nil, nil, "")))
	require.NoError(t, model.AddDomainStory(
		vo.NewDomainStory("Basic Flow", []string{"User"}, "User starts",
			[]string{"User performs action"}, nil)))
	require.NoError(t, model.ClassifySubdomain("Core", vo.SubdomainCore, "Main"))
	require.NoError(t, model.DesignAggregate(
		vo.NewAggregateDesign("CoreRoot", "Core", "CoreRoot", nil, nil, nil, nil)))
	require.NoError(t, model.Finalize())
	return model
}

// -- render_prd tests --

func TestRenderPRD_ProducesMarkdown(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRenderPRD_ContainsHeading(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "# Product Requirements Document")
}

func TestRenderPRD_ContainsContextNames(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "Ordering")
	assert.Contains(t, result, "Shipping")
}

func TestRenderPRD_ContainsStoriesAsScenarios(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "Place Order")
	assert.Contains(t, result, "Ship Order")
}

func TestRenderPRD_ContainsCapabilitiesSection(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "## 5. Capabilities")
}

// -- render_ddd tests --

func TestRenderDDD_ProducesMarkdown(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRenderDDD_ContainsHeading(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "# Domain-Driven Design Artifacts")
}

func TestRenderDDD_ContainsStories(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "Place Order")
	assert.Contains(t, result, "Ship Order")
}

func TestRenderDDD_ContainsStorySteps(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "Customer places order")
	assert.Contains(t, result, "Warehouse picks shipment")
}

func TestRenderDDD_ContainsUbiquitousLanguage(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "Order")
	assert.Contains(t, result, "Shipment")
	assert.Contains(t, result, "A customer purchase request")
}

func TestRenderDDD_ContainsBoundedContexts(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "Ordering")
	assert.Contains(t, result, "Shipping")
	assert.Contains(t, result, "Manages order lifecycle")
}

func TestRenderDDD_ContainsClassifications(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	lower := strings.ToLower(result)
	assert.Contains(t, lower, "core")
	assert.Contains(t, lower, "supporting")
}

func TestRenderDDD_ContainsAggregateDesigns(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderDDD(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "OrderRoot")
	assert.Contains(t, result, "Order must have items")
	assert.Contains(t, result, "OrderPlaced")
}

// -- render_architecture tests --

func TestRenderArchitecture_ProducesMarkdown(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRenderArchitecture_ContainsHeading(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "# Architecture")
}

func TestRenderArchitecture_ContainsBoundedContexts(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)
	assert.Contains(t, result, "Ordering")
	assert.Contains(t, result, "Shipping")
}

func TestRenderArchitecture_ContainsLayerRules(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)
	lower := strings.ToLower(result)
	assert.Contains(t, lower, "domain")
	assert.Contains(t, lower, "application")
	assert.Contains(t, lower, "infrastructure")
}

// -- Edge cases --

func TestMinimalModel(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer()
	model := minimalModel(t)

	prd, err := renderer.RenderPRD(context.Background(), model)
	require.NoError(t, err)
	assert.Contains(t, prd, "Core")

	dddDoc, err := renderer.RenderDDD(context.Background(), model)
	require.NoError(t, err)
	assert.Contains(t, dddDoc, "Basic Flow")

	arch, err := renderer.RenderArchitecture(context.Background(), model)
	require.NoError(t, err)
	assert.Contains(t, arch, "Core")
}

func TestNoAggregatesFallback(t *testing.T) {
	t.Parallel()
	model := ddd.NewDomainModel("no-aggs")
	require.NoError(t, model.AddBoundedContext(
		vo.NewDomainBoundedContext("Logging", "Audit logs", nil, nil, "")))
	require.NoError(t, model.AddDomainStory(
		vo.NewDomainStory("Write Log", []string{"System"}, "Event occurs",
			[]string{"System writes logging entry"}, nil)))
	require.NoError(t, model.ClassifySubdomain("Logging", vo.SubdomainGeneric, "Commodity"))
	require.NoError(t, model.Finalize())

	renderer := infrastructure.NewMarkdownArtifactRenderer()
	arch, err := renderer.RenderArchitecture(context.Background(), model)
	require.NoError(t, err)
	assert.Contains(t, arch, "No aggregate designs yet")

	dddDoc, err := renderer.RenderDDD(context.Background(), model)
	require.NoError(t, err)
	assert.Contains(t, dddDoc, "No aggregate designs yet")
}
