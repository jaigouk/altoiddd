package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

func TestLifecycleStalenessRule_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "lifecycle_staleness", NewLifecycleStalenessRule(30).Name())
}

func TestLifecycleStalenessRule_NonLifecyclePath_NoViolation(t *testing.T) {
	t.Parallel()
	old := time.Now().AddDate(0, 0, -365)
	a, err := dochealthdomain.NewScaffoldAssetWithModTime("commands/foo.md", nil, "", 0, false, old)
	require.NoError(t, err)
	// Rule only scopes to lifecycle/in-progress/ — everything else is exempt.
	assert.Empty(t, NewLifecycleStalenessRule(30).Check(a, nil))
}

func TestLifecycleStalenessRule_LifecycleStale_ReturnsWarning(t *testing.T) {
	t.Parallel()
	old := time.Now().AddDate(0, 0, -45)
	a, err := dochealthdomain.NewScaffoldAssetWithModTime(
		"lifecycle/in-progress/foo.md", nil, "", 0, false, old,
	)
	require.NoError(t, err)
	v := NewLifecycleStalenessRule(30).Check(a, nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
	// QA-MIN-3 — tightened: stale message must include both terms.
	assert.Contains(t, v[0].Message(), "stale")
	assert.Contains(t, v[0].Message(), "days")
}

// --- Round 1 fix-cycle additions ---

// WH-MED-3 — future mtime must produce a distinct WARNING, not silently skip.

func TestLifecycleStalenessRule_FutureMtime_ReturnsWarning(t *testing.T) {
	t.Parallel()
	future := time.Now().AddDate(1, 0, 0)
	a, err := dochealthdomain.NewScaffoldAssetWithModTime(
		"lifecycle/in-progress/foo.md", nil, "", 0, false, future,
	)
	require.NoError(t, err)
	v := NewLifecycleStalenessRule(30).Check(a, nil)
	require.Len(t, v, 1)
	assert.Equal(t, dochealthdomain.SeverityWarning, v[0].Severity())
	assert.Contains(t, v[0].Message(), "future",
		"future-mtime message must explicitly say 'future'")
}

func TestLifecycleStalenessRule_NearFutureMtime_ReturnsWarning(t *testing.T) {
	t.Parallel()
	// No grace period — even 1 second in the future fires.
	future := time.Now().Add(2 * time.Second)
	a, err := dochealthdomain.NewScaffoldAssetWithModTime(
		"lifecycle/in-progress/foo.md", nil, "", 0, false, future,
	)
	require.NoError(t, err)
	v := NewLifecycleStalenessRule(30).Check(a, nil)
	require.Len(t, v, 1)
	assert.Contains(t, v[0].Message(), "future")
}

func TestLifecycleStalenessRule_NowMtime_NoViolation(t *testing.T) {
	t.Parallel()
	// Boundary: mtime exactly now (or imperceptibly before) is fresh.
	now := time.Now()
	a, err := dochealthdomain.NewScaffoldAssetWithModTime(
		"lifecycle/in-progress/foo.md", nil, "", 0, false, now,
	)
	require.NoError(t, err)
	assert.Empty(t, NewLifecycleStalenessRule(30).Check(a, nil))
}

func TestLifecycleStalenessRule_LifecycleFresh_NoViolation(t *testing.T) {
	t.Parallel()
	recent := time.Now().AddDate(0, 0, -7)
	a, err := dochealthdomain.NewScaffoldAssetWithModTime(
		"lifecycle/in-progress/foo.md", nil, "", 0, false, recent,
	)
	require.NoError(t, err)
	assert.Empty(t, NewLifecycleStalenessRule(30).Check(a, nil))
}

func TestLifecycleStalenessRule_ZeroModTime_NoViolation(t *testing.T) {
	t.Parallel()
	// Walker may have failed to capture mtime; rule must not panic — and
	// should treat zero-time as "skip" rather than spuriously flag.
	a, err := dochealthdomain.NewScaffoldAsset("lifecycle/in-progress/foo.md", nil, "", 0, false)
	require.NoError(t, err)
	assert.Empty(t, NewLifecycleStalenessRule(30).Check(a, nil))
}

func TestLifecycleStalenessRule_PathNormalisation(t *testing.T) {
	t.Parallel()
	// Windows-style backslash path must still be recognised by the rule
	// (walker normalises via filepath.ToSlash, but defense in depth).
	old := time.Now().AddDate(0, 0, -60)
	a, err := dochealthdomain.NewScaffoldAssetWithModTime(
		"some/prefix/alto-scaffold/lifecycle/in-progress/foo.md", nil, "", 0, false, old,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, NewLifecycleStalenessRule(30).Check(a, nil))
}
