package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

func TestParseQualityGates_HappyPath(t *testing.T) {
	t.Parallel()
	gates, err := ParseQualityGates([]string{"lint", "types", "tests", "fitness"})
	require.NoError(t, err)
	assert.Equal(t, []vo.QualityGate{
		vo.QualityGateLint, vo.QualityGateTypes,
		vo.QualityGateTests, vo.QualityGateFitness,
	}, gates)
}

func TestParseQualityGates_UnknownGate(t *testing.T) {
	t.Parallel()
	_, err := ParseQualityGates([]string{"lint", "nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown quality gate")
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestParseQualityGates_EmptyList(t *testing.T) {
	t.Parallel()
	gates, err := ParseQualityGates(nil)
	require.NoError(t, err)
	assert.Empty(t, gates)
}

func TestParseSupportedTools_HappyPath(t *testing.T) {
	t.Parallel()
	tools, err := ParseSupportedTools([]string{"claude-code", "cursor", "roo-code", "opencode"})
	require.NoError(t, err)
	assert.Equal(t, []ttdomain.SupportedTool{
		ttdomain.ToolClaudeCode, ttdomain.ToolCursor,
		ttdomain.ToolRooCode, ttdomain.ToolOpenCode,
	}, tools)
}

func TestParseSupportedTools_UnknownTool(t *testing.T) {
	t.Parallel()
	_, err := ParseSupportedTools([]string{"invalid"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
	assert.Contains(t, err.Error(), "invalid")
}

func TestParseSupportedTools_EmptyList(t *testing.T) {
	t.Parallel()
	tools, err := ParseSupportedTools(nil)
	require.NoError(t, err)
	assert.Empty(t, tools)
}
