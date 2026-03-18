package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/fitness/domain"
	"github.com/alto-cli/alto/internal/fitness/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestBoundedContextMapParser_Parse_ValidYAML(t *testing.T) {
	t.Parallel()

	yaml := `
project:
  name: "alto"
  root_package: "github.com/alto-cli/alto"

bounded_contexts:
  - name: "Bootstrap"
    module_path: "bootstrap"
    classification: "supporting"
    layers:
      - domain
      - application
      - infrastructure
    relationships:
      - target: "Guided Discovery"
        direction: "downstream"
        pattern: "domain_event"

  - name: "Guided Discovery"
    module_path: "discovery"
    classification: "core"
    layers:
      - domain
      - application
      - infrastructure
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bounded_context_map.yaml")
	err := os.WriteFile(yamlPath, []byte(yaml), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	bcMap, err := parser.Parse(context.Background(), yamlPath)

	require.NoError(t, err)
	assert.Equal(t, "alto", bcMap.ProjectName())
	assert.Equal(t, "github.com/alto-cli/alto", bcMap.RootPackage())
	assert.Len(t, bcMap.Contexts(), 2)

	// Check Bootstrap context
	bootstrap, found := bcMap.FindContext("Bootstrap")
	require.True(t, found)
	assert.Equal(t, "bootstrap", bootstrap.ModulePath())
	assert.Equal(t, vo.SubdomainSupporting, bootstrap.Classification())
	assert.Len(t, bootstrap.Layers(), 3)
	assert.Len(t, bootstrap.Relationships(), 1)

	rel := bootstrap.Relationships()[0]
	assert.Equal(t, "Guided Discovery", rel.Target())
	assert.Equal(t, domain.RelationshipDownstream, rel.Direction())
	assert.Equal(t, domain.PatternDomainEvent, rel.Pattern())

	// Check Discovery context
	discovery, found := bcMap.FindContext("Guided Discovery")
	require.True(t, found)
	assert.Equal(t, vo.SubdomainCore, discovery.Classification())
}

func TestBoundedContextMapParser_Parse_AllClassifications(t *testing.T) {
	t.Parallel()

	yaml := `
project:
  name: "test"
  root_package: "github.com/org/test"

bounded_contexts:
  - name: "Core"
    module_path: "core"
    classification: "core"
    layers: [domain]

  - name: "Supporting"
    module_path: "supporting"
    classification: "supporting"
    layers: [domain]

  - name: "Generic"
    module_path: "generic"
    classification: "generic"
    layers: [domain]
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bc_map.yaml")
	err := os.WriteFile(yamlPath, []byte(yaml), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	bcMap, err := parser.Parse(context.Background(), yamlPath)

	require.NoError(t, err)

	core, found := bcMap.FindContext("Core")
	require.True(t, found)
	assert.Equal(t, vo.SubdomainCore, core.Classification())

	supporting, found := bcMap.FindContext("Supporting")
	require.True(t, found)
	assert.Equal(t, vo.SubdomainSupporting, supporting.Classification())

	generic, found := bcMap.FindContext("Generic")
	require.True(t, found)
	assert.Equal(t, vo.SubdomainGeneric, generic.Classification())
}

func TestBoundedContextMapParser_Parse_AllRelationshipPatterns(t *testing.T) {
	t.Parallel()

	yaml := `
project:
  name: "test"
  root_package: "github.com/org/test"

bounded_contexts:
  - name: "Context"
    module_path: "ctx"
    classification: "core"
    layers: [domain]
    relationships:
      - target: "A"
        direction: "upstream"
        pattern: "domain_event"
      - target: "B"
        direction: "downstream"
        pattern: "shared_kernel"
      - target: "C"
        direction: "upstream"
        pattern: "acl"
      - target: "D"
        direction: "downstream"
        pattern: "open_host"
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bc_map.yaml")
	err := os.WriteFile(yamlPath, []byte(yaml), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	bcMap, err := parser.Parse(context.Background(), yamlPath)

	require.NoError(t, err)

	ctx, found := bcMap.FindContext("Context")
	require.True(t, found)

	rels := ctx.Relationships()
	require.Len(t, rels, 4)

	assert.Equal(t, domain.PatternDomainEvent, rels[0].Pattern())
	assert.Equal(t, domain.PatternSharedKernel, rels[1].Pattern())
	assert.Equal(t, domain.PatternACL, rels[2].Pattern())
	assert.Equal(t, domain.PatternOpenHost, rels[3].Pattern())
}

func TestBoundedContextMapParser_Parse_NoRelationships(t *testing.T) {
	t.Parallel()

	yaml := `
project:
  name: "simple"
  root_package: "github.com/org/simple"

bounded_contexts:
  - name: "Standalone"
    module_path: "standalone"
    classification: "generic"
    layers: [domain, infrastructure]
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bc_map.yaml")
	err := os.WriteFile(yamlPath, []byte(yaml), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	bcMap, err := parser.Parse(context.Background(), yamlPath)

	require.NoError(t, err)

	ctx, found := bcMap.FindContext("Standalone")
	require.True(t, found)
	assert.Empty(t, ctx.Relationships())
}

func TestBoundedContextMapParser_Parse_FileNotFound(t *testing.T) {
	t.Parallel()

	parser := infrastructure.NewBoundedContextMapParser()
	_, err := parser.Parse(context.Background(), "/nonexistent/path.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading bounded context map")
}

func TestBoundedContextMapParser_Parse_InvalidYAML(t *testing.T) {
	t.Parallel()

	invalidYAML := `
project:
  name: "test"
  root_package: "github.com/org/test"

bounded_contexts:
  - name: "Bad"
    this is not valid yaml: [
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bad.yaml")
	err := os.WriteFile(yamlPath, []byte(invalidYAML), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	_, err = parser.Parse(context.Background(), yamlPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing bounded context map")
}

func TestBoundedContextMapParser_Parse_MissingProjectName(t *testing.T) {
	t.Parallel()

	yaml := `
project:
  root_package: "github.com/org/test"

bounded_contexts:
  - name: "Context"
    module_path: "ctx"
    classification: "core"
    layers: [domain]
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bc_map.yaml")
	err := os.WriteFile(yamlPath, []byte(yaml), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	_, err = parser.Parse(context.Background(), yamlPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project.name is required")
}

func TestBoundedContextMapParser_Parse_MissingRootPackage(t *testing.T) {
	t.Parallel()

	yaml := `
project:
  name: "test"

bounded_contexts:
  - name: "Context"
    module_path: "ctx"
    classification: "core"
    layers: [domain]
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bc_map.yaml")
	err := os.WriteFile(yamlPath, []byte(yaml), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	_, err = parser.Parse(context.Background(), yamlPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project.root_package is required")
}

func TestBoundedContextMapParser_Parse_InvalidClassification(t *testing.T) {
	t.Parallel()

	yaml := `
project:
  name: "test"
  root_package: "github.com/org/test"

bounded_contexts:
  - name: "Context"
    module_path: "ctx"
    classification: "invalid_classification"
    layers: [domain]
`

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bc_map.yaml")
	err := os.WriteFile(yamlPath, []byte(yaml), 0o644)
	require.NoError(t, err)

	parser := infrastructure.NewBoundedContextMapParser()
	_, err = parser.Parse(context.Background(), yamlPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid classification")
}
