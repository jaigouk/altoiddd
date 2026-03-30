package infrastructure_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// PRD — Go profile
// ---------------------------------------------------------------------------

func TestRenderPRD_GoProfile_ContainsGoLanguage(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.GoModProfile{})
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.Contains(t, result, "Go")
	assert.Contains(t, result, "go mod")
	assert.NotContains(t, result, "Python")
	assert.NotContains(t, result, "uv")
}

// ---------------------------------------------------------------------------
// Architecture — Go profile
// ---------------------------------------------------------------------------

func TestRenderArchitecture_GoProfile_ContainsPureGo(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.GoModProfile{})
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.Contains(t, result, "pure Go")
	assert.NotContains(t, result, "pure Python")
}

func TestRenderArchitecture_GoProfile_ContainsGoSourceLayout(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.GoModProfile{})
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.Contains(t, result, "internal/")
	assert.NotContains(t, result, "src/domain/")
	assert.NotContains(t, result, "src/application/")
	assert.NotContains(t, result, "src/infrastructure/")
}

// ---------------------------------------------------------------------------
// Nil profile — fallback to Python defaults
// ---------------------------------------------------------------------------

func TestRenderPRD_NilProfile_FallsBackToDefaults(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(nil)
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Python 3.12+")
	assert.Contains(t, result, "uv")
}

func TestRenderArchitecture_NilProfile_FallsBackToDefaults(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(nil)
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "pure Python")
	assert.Contains(t, result, "src/domain/")
}

// ---------------------------------------------------------------------------
// Python profile — explicit (preserves current behavior)
// ---------------------------------------------------------------------------

func TestRenderPRD_PythonProfile_ContainsPythonLanguage(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.PythonUvProfile{})
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.Contains(t, result, "Python 3.12+")
	assert.Contains(t, result, "uv")
}

func TestRenderArchitecture_PythonProfile_ContainsPurePython(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.PythonUvProfile{})
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.Contains(t, result, "pure Python")
	assert.Contains(t, result, "src/domain/")
}

// ---------------------------------------------------------------------------
// Generic profile — uses generic fallback language
// ---------------------------------------------------------------------------

func TestRenderPRD_GenericProfile_NoSpecificLanguage(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.GenericProfile{})
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)

	assert.NotEmpty(t, result)
	// Generic should NOT claim Python or Go
	assert.NotContains(t, result, "Python")
	assert.NotContains(t, result, "go mod")
}

// ---------------------------------------------------------------------------
// Go profile — negative: no Python anywhere
// ---------------------------------------------------------------------------

func TestRenderPRD_GoProfile_NoPythonAnywhere(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.GoModProfile{})
	result, err := renderer.RenderPRD(context.Background(), sampleModel(t))
	require.NoError(t, err)

	lower := strings.ToLower(result)
	assert.NotContains(t, lower, "python")
	assert.NotContains(t, lower, " uv ")
}

func TestRenderArchitecture_GoProfile_NoPythonAnywhere(t *testing.T) {
	t.Parallel()
	renderer := infrastructure.NewMarkdownArtifactRenderer(vo.GoModProfile{})
	result, err := renderer.RenderArchitecture(context.Background(), sampleModel(t))
	require.NoError(t, err)

	lower := strings.ToLower(result)
	assert.NotContains(t, lower, "python")
}
