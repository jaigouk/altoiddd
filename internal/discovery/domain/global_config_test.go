package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGlobalConfig_Valid(t *testing.T) {
	t.Parallel()
	gc, err := NewGlobalConfig("claude-code", "/home/.claude")
	require.NoError(t, err)
	assert.Equal(t, "claude-code", gc.Tool())
	assert.Equal(t, "/home/.claude", gc.Path())
}

func TestNewGlobalConfig_EmptyTool(t *testing.T) {
	t.Parallel()
	_, err := NewGlobalConfig("", "/home/.claude")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool name required")
}

func TestNewGlobalConfig_EmptyPath(t *testing.T) {
	t.Parallel()
	_, err := NewGlobalConfig("claude-code", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path required")
}

func TestGlobalConfig_Equal(t *testing.T) {
	t.Parallel()
	gc1, _ := NewGlobalConfig("claude-code", "/home/.claude")
	gc2, _ := NewGlobalConfig("claude-code", "/home/.claude")
	gc3, _ := NewGlobalConfig("cursor", "/home/.cursor")
	assert.True(t, gc1.Equal(gc2))
	assert.False(t, gc1.Equal(gc3))
}
