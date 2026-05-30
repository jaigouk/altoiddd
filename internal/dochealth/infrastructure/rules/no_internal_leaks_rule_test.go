package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func newBodyAsset(t *testing.T, path, body string, isOverlay bool) dochealthdomain.ScaffoldAsset {
	t.Helper()
	a, err := dochealthdomain.NewScaffoldAsset(path, fullFrontmatter(), body, 1, isOverlay)
	require.NoError(t, err)
	return a
}

func TestNoInternalLeaksRule_InternalRefInGeneric_ReturnsError(t *testing.T) {
	t.Parallel()
	asset := newBodyAsset(t, "foo.md", "see internal/ports.go for details", false)
	violations := NewNoInternalLeaksRule().Check(asset, nil)
	assert.Len(t, violations, 1)
}

func TestNoInternalLeaksRule_InternalRefInOverlay_NoViolation(t *testing.T) {
	t.Parallel()
	asset := newBodyAsset(t, "foo.project.md", "see internal/ports.go for details", true)
	violations := NewNoInternalLeaksRule().Check(asset, nil)
	assert.Empty(t, violations, "overlays are exempt by design")
}

func TestNoInternalLeaksRule_GolangciInGeneric_ReturnsError(t *testing.T) {
	t.Parallel()
	asset := newBodyAsset(t, "foo.md", "run golangci-lint", false)
	violations := NewNoInternalLeaksRule().Check(asset, nil)
	assert.Len(t, violations, 1)
}

func TestNoInternalLeaksRule_CmdAltoInGeneric_ReturnsError(t *testing.T) {
	t.Parallel()
	asset := newBodyAsset(t, "foo.md", "build cmd/alto first", false)
	violations := NewNoInternalLeaksRule().Check(asset, nil)
	assert.Len(t, violations, 1)
}

func TestNoInternalLeaksRule_WatermillInGeneric_ReturnsError(t *testing.T) {
	t.Parallel()
	asset := newBodyAsset(t, "foo.md", "publish via Watermill bus", false)
	violations := NewNoInternalLeaksRule().Check(asset, nil)
	assert.Len(t, violations, 1)
}

// Binding RED test from the ticket — proves the regex uses `internal/`
// (with literal slash) not bare `internal`, so prose like "internally"
// does NOT trigger.
func TestNoInternalLeaksRule_ProseInternallyNoSlash_NoFalsePositive(t *testing.T) {
	t.Parallel()
	asset := newBodyAsset(t, "foo.md", "the module is internally consistent and ready", false)
	violations := NewNoInternalLeaksRule().Check(asset, nil)
	assert.Empty(t, violations, "the prose word 'internally' must not match the rule")
}

func TestNoInternalLeaksRule_CleanGeneric_NoViolation(t *testing.T) {
	t.Parallel()
	asset := newBodyAsset(t, "foo.md", "regular markdown body without any alto-internal references", false)
	violations := NewNoInternalLeaksRule().Check(asset, nil)
	// "alto-" pattern catches "alto-internal" in this string — confirming
	// the rule is correctly aggressive. We instead test truly neutral text:
	clean := newBodyAsset(t, "bar.md", "plain markdown body about cats and dogs", false)
	assert.Empty(t, NewNoInternalLeaksRule().Check(clean, nil))
	// Sanity: the original "alto-" reference IS caught.
	assert.Len(t, violations, 1)
}
