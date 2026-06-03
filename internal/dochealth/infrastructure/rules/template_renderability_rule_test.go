package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func TestTemplateRenderabilityRule_RuleName_IsTemplateRenderability(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "template_renderability", NewTemplateRenderabilityRule().Name())
}

func TestTemplateRenderabilityRule_WhenAssetReferencesUnknownField_ReturnsERRORViolation(t *testing.T) {
	t.Parallel()
	body := "Hello {{.UnknownField}} world"
	asset, err := dochealthdomain.NewScaffoldAsset("commands/sample.md", nil, body, 1, false)
	require.NoError(t, err)

	violations := NewTemplateRenderabilityRule().Check(asset, nil)
	require.Len(t, violations, 1)
	assert.Equal(t, dochealthdomain.SeverityError, violations[0].Severity())
	assert.Equal(t, "template_renderability", violations[0].Rule())
	assert.Contains(t, violations[0].Message(), "UnknownField",
		"violation message must surface the offending field name")
}

func TestTemplateRenderabilityRule_WhenAssetIsValidTemplate_ReturnsNoViolations(t *testing.T) {
	t.Parallel()
	body := "Project: {{.ProjectName}} Prefix: {{.TicketPrefix}} Tracker: {{.IssueTracker}}"
	asset, err := dochealthdomain.NewScaffoldAsset("commands/sample.md", nil, body, 1, false)
	require.NoError(t, err)

	assert.Empty(t, NewTemplateRenderabilityRule().Check(asset, nil))
}

func TestTemplateRenderabilityRule_WhenAssetHasNoTemplateSyntax_ReturnsNoViolations(t *testing.T) {
	t.Parallel()
	body := "Plain markdown with no template directives at all.\n\nJust prose."
	asset, err := dochealthdomain.NewScaffoldAsset("commands/sample.md", nil, body, 3, false)
	require.NoError(t, err)

	assert.Empty(t, NewTemplateRenderabilityRule().Check(asset, nil))
}

func TestTemplateRenderabilityRule_WhenAssetHasParseError_ReturnsERRORViolation(t *testing.T) {
	t.Parallel()
	// `{{` with no closing — text/template parse error.
	body := "Open brace {{ with no close"
	asset, err := dochealthdomain.NewScaffoldAsset("commands/sample.md", nil, body, 1, false)
	require.NoError(t, err)

	violations := NewTemplateRenderabilityRule().Check(asset, nil)
	require.Len(t, violations, 1)
	assert.Equal(t, dochealthdomain.SeverityError, violations[0].Severity())
	assert.Equal(t, "template_renderability", violations[0].Rule())
	assert.Contains(t, violations[0].Message(), "parse",
		"parse-time errors should be marked as such in the message")
}

func TestTemplateRenderabilityRule_EmptyBody_NoViolation(t *testing.T) {
	t.Parallel()
	asset, err := dochealthdomain.NewScaffoldAsset("commands/empty.md", nil, "", 0, false)
	require.NoError(t, err)

	assert.Empty(t, NewTemplateRenderabilityRule().Check(asset, nil))
}

func TestTemplateRenderabilityRule_NestedUnknownField_ReturnsERRORViolation(t *testing.T) {
	t.Parallel()
	// `{{.BoundedContexts.Missing}}` — BoundedContexts is real but `.Missing`
	// is a map/struct lookup on a []string, which fails at execute time
	// under Option("missingkey=error") (or fails as a type error during
	// evaluation — either way, an ERROR violation).
	body := "List: {{.BoundedContexts.Missing}}"
	asset, err := dochealthdomain.NewScaffoldAsset("commands/sample.md", nil, body, 1, false)
	require.NoError(t, err)

	violations := NewTemplateRenderabilityRule().Check(asset, nil)
	require.Len(t, violations, 1)
	assert.Equal(t, dochealthdomain.SeverityError, violations[0].Severity())
}

func TestTemplateRenderabilityRule_AssetPathInFile_Field(t *testing.T) {
	t.Parallel()
	body := "{{.Bogus}}"
	asset, err := dochealthdomain.NewScaffoldAsset("agents/researcher.md", nil, body, 1, false)
	require.NoError(t, err)

	violations := NewTemplateRenderabilityRule().Check(asset, nil)
	require.Len(t, violations, 1)
	assert.Equal(t, "agents/researcher.md", violations[0].File())
}
