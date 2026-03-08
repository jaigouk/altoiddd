package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagContent_ToolOutput(t *testing.T) {
	t.Parallel()
	result := TagContent("Hello world", "tools/call")
	assert.Equal(t, "[TOOL OUTPUT START]\nHello world\n[TOOL OUTPUT END]", result)
}

func TestTagContent_ResourceOutput_UsesDefaultLabel(t *testing.T) {
	t.Parallel()
	result := TagContent("DDD knowledge", "resources/read")
	assert.Equal(t, "[MCP OUTPUT START]\nDDD knowledge\n[MCP OUTPUT END]", result)
}

func TestTagContent_OtherMethod(t *testing.T) {
	t.Parallel()
	result := TagContent("data", "prompts/list")
	assert.Equal(t, "[MCP OUTPUT START]\ndata\n[MCP OUTPUT END]", result)
}

func TestMethodLabel_KnownMethods(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "TOOL OUTPUT", methodLabel("tools/call"))
	assert.Equal(t, "MCP OUTPUT", methodLabel("resources/read"))
	assert.Equal(t, "MCP OUTPUT", methodLabel("unknown/method"))
}
