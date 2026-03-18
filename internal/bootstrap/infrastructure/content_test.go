package infrastructure_test

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/bootstrap/domain"
	"github.com/alto-cli/alto/internal/bootstrap/infrastructure"
)

func TestAltoConfigContent_IsValidTOML(t *testing.T) {
	t.Parallel()

	config := domain.NewProjectConfig("my-project", "go", "github.com/user/my-project", []string{"claude", "cursor"})
	content := infrastructure.AltoConfigContent(config)

	var parsed map[string]any
	_, err := toml.Decode(content, &parsed)
	require.NoError(t, err, "AltoConfigContent must produce valid TOML")

	project, ok := parsed["project"].(map[string]any)
	require.True(t, ok, "expected [project] table")
	assert.Equal(t, "my-project", project["name"])
	assert.Equal(t, "go", project["language"])
	assert.Equal(t, "github.com/user/my-project", project["module_path"])

	tools, ok := parsed["tools"].(map[string]any)
	require.True(t, ok, "expected [tools] table")
	detected, ok := tools["detected"].([]any)
	require.True(t, ok, "expected detected array")
	assert.Len(t, detected, 2)

	discovery, ok := parsed["discovery"].(map[string]any)
	require.True(t, ok, "expected [discovery] table")
	assert.Equal(t, false, discovery["completed"])
}

func TestAltoConfigContent_WhenEmptyLanguage_OmitsLanguageLine(t *testing.T) {
	t.Parallel()

	config := domain.NewProjectConfig("svc", "", "", nil)
	content := infrastructure.AltoConfigContent(config)

	var parsed map[string]any
	_, err := toml.Decode(content, &parsed)
	require.NoError(t, err)

	project := parsed["project"].(map[string]any)
	_, hasLang := project["language"]
	assert.False(t, hasLang, "empty language should be omitted")
	_, hasModule := project["module_path"]
	assert.False(t, hasModule, "empty module_path should be omitted")
}

func TestAltoConfigContent_WhenNoTools_EmitsEmptyArray(t *testing.T) {
	t.Parallel()

	config := domain.NewProjectConfig("svc", "go", "", nil)
	content := infrastructure.AltoConfigContent(config)

	var parsed map[string]any
	_, err := toml.Decode(content, &parsed)
	require.NoError(t, err)

	tools := parsed["tools"].(map[string]any)
	detected := tools["detected"].([]any)
	assert.Empty(t, detected)
}

func TestKnowledgeIndexContent_IsValidTOML(t *testing.T) {
	t.Parallel()

	content := infrastructure.KnowledgeIndexContent()

	var parsed map[string]any
	_, err := toml.Decode(content, &parsed)
	require.NoError(t, err, "KnowledgeIndexContent must produce valid TOML")
}

func TestDocRegistryContent_IsValidTOML(t *testing.T) {
	t.Parallel()

	content := infrastructure.DocRegistryContent()

	var parsed map[string]any
	_, err := toml.Decode(content, &parsed)
	require.NoError(t, err, "DocRegistryContent must produce valid TOML")
}
