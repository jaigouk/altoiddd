package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectedToolCreationWithAllFields(t *testing.T) {
	t.Parallel()
	tool := NewDetectedTool("claude-code", "/home/user/.claude", "1.2.3")
	assert.Equal(t, "claude-code", tool.Name())
	assert.Equal(t, "/home/user/.claude", tool.ConfigPath())
	assert.Equal(t, "1.2.3", tool.Version())
}

func TestDetectedToolCreationWithDefaults(t *testing.T) {
	t.Parallel()
	tool := NewDetectedTool("cursor", "", "")
	assert.Equal(t, "cursor", tool.Name())
	assert.Empty(t, tool.ConfigPath())
	assert.Empty(t, tool.Version())
}

func TestDetectedToolCreationWithConfigPathOnly(t *testing.T) {
	t.Parallel()
	tool := NewDetectedTool("roo-code", "/home/user/.roo", "")
	assert.Equal(t, "roo-code", tool.Name())
	assert.Equal(t, "/home/user/.roo", tool.ConfigPath())
	assert.Empty(t, tool.Version())
}

func TestDetectedToolEquality(t *testing.T) {
	t.Parallel()
	t1 := NewDetectedTool("claude-code", "/a", "1.0")
	t2 := NewDetectedTool("claude-code", "/a", "1.0")
	assert.True(t, t1.Equal(t2))
}

func TestDetectedToolDifferentNamesNotEqual(t *testing.T) {
	t.Parallel()
	t1 := NewDetectedTool("claude-code", "", "")
	t2 := NewDetectedTool("cursor", "", "")
	assert.False(t, t1.Equal(t2))
}

func TestDetectedToolDifferentVersionsNotEqual(t *testing.T) {
	t.Parallel()
	t1 := NewDetectedTool("claude-code", "", "1.0")
	t2 := NewDetectedTool("claude-code", "", "2.0")
	assert.False(t, t1.Equal(t2))
}
