package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func TestBodySizeRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "body_size", NewBodySizeRule().Name())
}

func TestBodySizeRule_UnderThreshold_NoViolation(t *testing.T) {
	t.Parallel()
	body := strings.Repeat("x\n", 100)
	asset, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, body, 100, false)
	require.NoError(t, err)
	assert.Empty(t, NewBodySizeRule().Check(asset, nil))
}

func TestBodySizeRule_OverThreshold_ReturnsWarning(t *testing.T) {
	t.Parallel()
	body := strings.Repeat("x\n", 501)
	asset, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, body, 501, false)
	require.NoError(t, err)
	v := NewBodySizeRule().Check(asset, nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
	assert.Equal(t, "body_size", v[0].Rule())
}

func TestBodySizeRule_ExactlyThreshold_NoViolation(t *testing.T) {
	t.Parallel()
	body := strings.Repeat("x\n", 500)
	asset, err := dochealthdomain.NewScaffoldAsset("foo.md", nil, body, 500, false)
	require.NoError(t, err)
	assert.Empty(t, NewBodySizeRule().Check(asset, nil))
}

func TestBodySizeRule_OverlayNotExempt(t *testing.T) {
	t.Parallel()
	// Overlays count toward merged body size; the rule operates on the
	// asset's own body and assumes the walker has not pre-merged. The
	// per-asset check still produces a violation when an overlay alone
	// blows the budget (defense in depth).
	body := strings.Repeat("x\n", 600)
	overlay, err := dochealthdomain.NewScaffoldAsset("foo.project.md", nil, body, 600, true)
	require.NoError(t, err)
	v := NewBodySizeRule().Check(overlay, nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
}
