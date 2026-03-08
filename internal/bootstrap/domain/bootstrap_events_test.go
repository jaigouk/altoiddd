package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapCompletedEvent_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := NewBootstrapCompletedEvent("sess-rt", "/tmp/project")

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"session_id"`)
	assert.Contains(t, string(data), `"project_dir"`)

	var restored BootstrapCompletedEvent
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "sess-rt", restored.SessionID())
	assert.Equal(t, "/tmp/project", restored.ProjectDir())
}
