package infrastructure_test

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/bootstrap/infrastructure"
)

func TestAltyConfigContent_IsValidTOML(t *testing.T) {
	t.Parallel()

	content := infrastructure.AltyConfigContent("my-project")

	var parsed map[string]any
	_, err := toml.Decode(content, &parsed)
	require.NoError(t, err, "AltyConfigContent must produce valid TOML")
	assert.Equal(t, "my-project", parsed["project_name"])
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
