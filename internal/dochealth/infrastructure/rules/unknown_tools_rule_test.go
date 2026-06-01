package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func toolsAsset(t *testing.T, tools any, isOverlay bool) dochealthdomain.ScaffoldAsset {
	t.Helper()
	fm := fullFrontmatter()
	fm["tools_required"] = tools
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", fm, "", 0, isOverlay)
	require.NoError(t, err)
	return a
}

func TestUnknownToolsRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "unknown_tools", NewUnknownToolsRule().Name())
}

func TestUnknownToolsRule_ValidNames_NoViolation(t *testing.T) {
	t.Parallel()
	cases := []any{
		"Read",
		"Read, Write, Bash",
		"mcp__qmd__query",
		"Read, mcp__qmd__query, Write",
		[]any{"Read", "Write"},
		[]any{"mcp__qmd__query", "mcp__pencil__open_document"},
	}
	rule := NewUnknownToolsRule()
	for _, tc := range cases {
		asset := toolsAsset(t, tc, false)
		assert.Empty(t, rule.Check(asset, nil), "input: %v", tc)
	}
}

func TestUnknownToolsRule_InvalidName_ReturnsWarning(t *testing.T) {
	t.Parallel()
	asset := toolsAsset(t, "Read, lower-case-tool", false)
	v := NewUnknownToolsRule().Check(asset, nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
}

func TestUnknownToolsRule_MalformedMcp_ReturnsWarning(t *testing.T) {
	t.Parallel()
	asset := toolsAsset(t, "mcp__BAD__name", false)
	v := NewUnknownToolsRule().Check(asset, nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
}

func TestUnknownToolsRule_OverlayExempt(t *testing.T) {
	t.Parallel()
	// Overlays have no frontmatter at all — rule no-ops.
	a, err := dochealthdomain.NewScaffoldAsset("foo.project.md", nil, "", 0, true)
	require.NoError(t, err)
	assert.Empty(t, NewUnknownToolsRule().Check(a, nil))
}

func TestUnknownToolsRule_MissingTools_NoViolation(t *testing.T) {
	t.Parallel()
	// FrontmatterSchemaRule reports missing fields; this rule only validates
	// the values that are present.
	fm := fullFrontmatter()
	delete(fm, "tools_required")
	a, err := dochealthdomain.NewScaffoldAsset("foo.md", fm, "", 0, false)
	require.NoError(t, err)
	assert.Empty(t, NewUnknownToolsRule().Check(a, nil))
}
