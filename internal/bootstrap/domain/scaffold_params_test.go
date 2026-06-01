package domain_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/bootstrap/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

func TestNewScaffoldParams_HappyPath_AllFieldsSet(t *testing.T) {
	t.Parallel()
	p, err := domain.NewScaffoldParams(
		"demo",
		"demo-",
		"beads",
		[]string{"Orders", "Catalog"},
		"claude",
	)
	require.NoError(t, err)
	assert.Equal(t, "demo", p.ProjectName)
	assert.Equal(t, "demo-", p.TicketPrefix)
	assert.Equal(t, "beads", p.IssueTracker)
	assert.Equal(t, []string{"Orders", "Catalog"}, p.BoundedContexts)
	assert.Equal(t, "claude", p.PrimaryTool)
}

func TestNewScaffoldParams_EmptyProjectName_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("", "demo-", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_ProjectNameWithSlash_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("foo/bar", "demo-", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_ProjectNameWithBackslash_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams(`foo\bar`, "demo-", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_ProjectNameWithDotDot_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("../escape", "demo-", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_ProjectNameWithNUL_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("foo\x00bar", "demo-", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_ProjectNameShellMetachar_TableDriven(t *testing.T) {
	t.Parallel()
	metas := []string{"$", "`", ";", "|", "&", "<", ">", "(", ")", "\n", "\r"}
	for _, m := range metas {
		m := m
		t.Run("meta_"+m, func(t *testing.T) {
			t.Parallel()
			name := "foo" + m + "bar"
			_, err := domain.NewScaffoldParams(name, "demo-", "beads", nil, "claude")
			require.Error(t, err, "metachar %q must be rejected", m)
			assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestNewScaffoldParams_ProjectNameWithBraces_Accepted(t *testing.T) {
	// `{` / `}` deliberately NOT in metachar deny list.
	// text/template data binding handles {{.Evil}} safely at render time.
	t.Parallel()
	p, err := domain.NewScaffoldParams("{{.Evil}}", "demo-", "beads", nil, "claude")
	require.NoError(t, err)
	assert.Equal(t, "{{.Evil}}", p.ProjectName)
}

func TestNewScaffoldParams_TicketPrefixMissingTrailingDash_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "demo", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_TicketPrefixLeadingDigit_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "1demo-", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_TicketPrefixEmbeddedShellMetachar_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "de$mo-", "beads", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_BoundedContextLowercaseFirst_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "demo-", "beads", []string{"orders"}, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_UnknownIssueTracker_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "demo-", "jira", nil, "claude")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_KnownIssueTrackers_Accepted(t *testing.T) {
	t.Parallel()
	for _, tracker := range []string{"beads", "github", "linear"} {
		tracker := tracker
		t.Run(tracker, func(t *testing.T) {
			t.Parallel()
			_, err := domain.NewScaffoldParams("demo", "demo-", tracker, nil, "claude")
			require.NoError(t, err)
		})
	}
}

func TestNewScaffoldParams_UnknownPrimaryTool_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "demo-", "beads", nil, "vscode")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_CursorPrimaryTool_ReturnsErrInvariantViolation(t *testing.T) {
	// §Scope Update: cursor + roo rejected as unknown (not "not yet implemented").
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "demo-", "beads", nil, "cursor")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_RooPrimaryTool_ReturnsErrInvariantViolation(t *testing.T) {
	t.Parallel()
	_, err := domain.NewScaffoldParams("demo", "demo-", "beads", nil, "roo")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewScaffoldParams_OpencodePrimaryTool_Accepted(t *testing.T) {
	t.Parallel()
	p, err := domain.NewScaffoldParams("demo", "demo-", "beads", nil, "opencode")
	require.NoError(t, err)
	assert.Equal(t, "opencode", p.PrimaryTool)
}

func TestNewScaffoldParams_EmptyBoundedContexts_IsValid(t *testing.T) {
	t.Parallel()
	p, err := domain.NewScaffoldParams("demo", "demo-", "beads", nil, "claude")
	require.NoError(t, err)
	assert.Empty(t, p.BoundedContexts)
}

func TestScaffoldParams_FieldsAreExported(t *testing.T) {
	// Required for text/template addressability via {{.ProjectName}} etc.
	t.Parallel()
	typ := reflect.TypeOf(domain.ScaffoldParams{})
	expected := []string{"ProjectName", "TicketPrefix", "IssueTracker", "BoundedContexts", "PrimaryTool"}
	for _, name := range expected {
		f, ok := typ.FieldByName(name)
		require.True(t, ok, "field %s must exist", name)
		assert.True(t, f.IsExported(), "field %s must be exported for text/template", name)
	}
}
