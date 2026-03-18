package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ===========================================================================
// Scenario: alto init creates .alto/ directory files
// ===========================================================================

func TestInitCreatesAltoConfigTOML_GivenEmptyProject_WhenPreviewConfirmExecute_ThenConfigTOMLExists(t *testing.T) {
	t.Parallel()

	// Given: empty project with README
	app := newApp(t)
	dir := makeTempProjectDir(t)

	// When: Preview → Confirm → Execute
	session, err := app.BootstrapHandler.Preview(dir)
	require.NoError(t, err)

	_, err = app.BootstrapHandler.Confirm(session.SessionID())
	require.NoError(t, err)

	_, err = app.BootstrapHandler.Execute(session.SessionID())
	require.NoError(t, err)

	// Then: .alto/config.toml exists with valid TOML
	configPath := filepath.Join(dir, ".alto", "config.toml")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err, ".alto/config.toml should exist after Execute")

	var parsed map[string]any
	_, err = toml.Decode(string(content), &parsed)
	require.NoError(t, err, ".alto/config.toml should contain valid TOML")
	project, ok := parsed["project"].(map[string]any)
	require.True(t, ok, "config.toml should include [project] table")
	assert.NotEmpty(t, project["name"], "config.toml [project] should include name")
}

func TestInitCreatesKnowledgeIndex_GivenEmptyProject_WhenPreviewConfirmExecute_ThenIndexTOMLExists(t *testing.T) {
	t.Parallel()

	// Given: empty project with README
	app := newApp(t)
	dir := makeTempProjectDir(t)

	// When: Preview → Confirm → Execute
	session, err := app.BootstrapHandler.Preview(dir)
	require.NoError(t, err)
	_, err = app.BootstrapHandler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = app.BootstrapHandler.Execute(session.SessionID())
	require.NoError(t, err)

	// Then: .alto/knowledge/_index.toml exists with valid TOML
	indexPath := filepath.Join(dir, ".alto", "knowledge", "_index.toml")
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err, ".alto/knowledge/_index.toml should exist after Execute")

	var parsed map[string]any
	_, err = toml.Decode(string(content), &parsed)
	require.NoError(t, err, "_index.toml should contain valid TOML")
}

func TestInitCreatesDocRegistry_GivenEmptyProject_WhenPreviewConfirmExecute_ThenRegistryTOMLExists(t *testing.T) {
	t.Parallel()

	// Given: empty project with README
	app := newApp(t)
	dir := makeTempProjectDir(t)

	// When: Preview → Confirm → Execute
	session, err := app.BootstrapHandler.Preview(dir)
	require.NoError(t, err)
	_, err = app.BootstrapHandler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = app.BootstrapHandler.Execute(session.SessionID())
	require.NoError(t, err)

	// Then: .alto/maintenance/doc-registry.toml exists with valid TOML
	registryPath := filepath.Join(dir, ".alto", "maintenance", "doc-registry.toml")
	content, err := os.ReadFile(registryPath)
	require.NoError(t, err, ".alto/maintenance/doc-registry.toml should exist after Execute")

	var parsed map[string]any
	_, err = toml.Decode(string(content), &parsed)
	require.NoError(t, err, "doc-registry.toml should contain valid TOML")
}

func TestInitSkipsExistingAltoFiles_GivenProjectWithAltoConfig_WhenPreviewConfirmExecute_ThenFileNotOverwritten(t *testing.T) {
	t.Parallel()

	// Given: project with existing .alto/config.toml
	app := newApp(t)
	dir := makeTempProjectDir(t)
	altoDir := filepath.Join(dir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))
	originalContent := "# original content\nproject_name = \"original\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(altoDir, "config.toml"), []byte(originalContent), 0o644))

	// When: Preview → Confirm → Execute
	session, err := app.BootstrapHandler.Preview(dir)
	require.NoError(t, err)

	// Verify preview marks config.toml as Skip
	preview := session.Preview()
	require.NotNil(t, preview)
	for _, action := range preview.FileActions() {
		if action.Path() == ".alto/config.toml" {
			assert.Equal(t, vo.FileActionSkip, action.ActionType(),
				".alto/config.toml should be marked as Skip when it already exists")
		}
	}

	_, err = app.BootstrapHandler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = app.BootstrapHandler.Execute(session.SessionID())
	require.NoError(t, err)

	// Then: file retains original content (not overwritten)
	content, err := os.ReadFile(filepath.Join(altoDir, "config.toml"))
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(content),
		"existing .alto/config.toml should not be overwritten")
}

func TestInitCreatesAllThreeAltoFiles_GivenEmptyProject_WhenFullFlow_ThenAllExist(t *testing.T) {
	t.Parallel()

	// Given: empty project with README
	app := newApp(t)
	dir := makeTempProjectDir(t)

	// When: Preview → Confirm → Execute
	session, err := app.BootstrapHandler.Preview(dir)
	require.NoError(t, err)
	_, err = app.BootstrapHandler.Confirm(session.SessionID())
	require.NoError(t, err)
	_, err = app.BootstrapHandler.Execute(session.SessionID())
	require.NoError(t, err)

	// Then: all three .alto files exist
	expectedFiles := []string{
		filepath.Join(dir, ".alto", "config.toml"),
		filepath.Join(dir, ".alto", "knowledge", "_index.toml"),
		filepath.Join(dir, ".alto", "maintenance", "doc-registry.toml"),
	}
	for _, path := range expectedFiles {
		_, err := os.Stat(path)
		assert.NoError(t, err, "expected file to exist: %s", path)
	}
}
