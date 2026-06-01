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
	// Neutral body — no Go-module path references — must not trigger.
	clean := newBodyAsset(t, "bar.md", "plain markdown body about cats and dogs", false)
	assert.Empty(t, NewNoInternalLeaksRule().Check(clean, nil))
	// User-facing references to the alto-scaffold/ directory must NOT trigger:
	// the rule narrowed from \balto- to \balto-cli\b so it only catches the Go
	// module path, not the canonical user-facing folder name.
	scaffold := newBodyAsset(t, "baz.md", "drop alto-scaffold/ into your project root", false)
	assert.Empty(t, NewNoInternalLeaksRule().Check(scaffold, nil))
	// Sanity: a real Go-module-path leak IS caught.
	leak := newBodyAsset(t, "qux.md", "the alto-cli source tree lives elsewhere", false)
	assert.Len(t, NewNoInternalLeaksRule().Check(leak, nil), 1)
}
