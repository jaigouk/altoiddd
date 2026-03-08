package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigsGeneratedEvent_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := newConfigsGeneratedEvent(
		[]string{"claude-code", "cursor"},
		[]string{".claude/CLAUDE.md", "AGENTS.md"},
	)

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"tool_names"`)
	assert.Contains(t, string(data), `"output_paths"`)

	var restored ConfigsGeneratedEvent
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.ToolNames(), restored.ToolNames())
	assert.Equal(t, original.OutputPaths(), restored.OutputPaths())
}
