package mcp_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/mcp"
)

// ---------------------------------------------------------------------------
// SafeComponent tests
// ---------------------------------------------------------------------------

func TestSafeComponent_ValidName(t *testing.T) {
	t.Parallel()
	assert.NoError(t, mcp.SafeComponent("my-project"))
	assert.NoError(t, mcp.SafeComponent("project_v2"))
	assert.NoError(t, mcp.SafeComponent("CamelCase"))
	assert.NoError(t, mcp.SafeComponent("project.name"))
}

func TestSafeComponent_PathTraversal(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeComponent("../../../etc/passwd"))
	require.Error(t, mcp.SafeComponent(".."))
	require.Error(t, mcp.SafeComponent("foo/../bar"))
}

func TestSafeComponent_EmptyString(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeComponent(""))
}

func TestSafeComponent_WithSlash(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeComponent("foo/bar"))
	require.Error(t, mcp.SafeComponent("/absolute"))
}

func TestSafeComponent_WithBackslash(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeComponent(`foo\bar`))
}

func TestSafeComponent_NullByte(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeComponent("foo\x00bar"))
}

// ---------------------------------------------------------------------------
// SafeTicketID tests
// ---------------------------------------------------------------------------

func TestSafeTicketID_ValidIDs(t *testing.T) {
	t.Parallel()
	assert.NoError(t, mcp.SafeTicketID("alto-0m9.5"))
	assert.NoError(t, mcp.SafeTicketID("k7m.12"))
	assert.NoError(t, mcp.SafeTicketID("abc-123"))
	assert.NoError(t, mcp.SafeTicketID("a"))
}

func TestSafeTicketID_Empty(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeTicketID(""))
}

func TestSafeTicketID_ShellInjection(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeTicketID("'; cat /etc/passwd"))
	require.Error(t, mcp.SafeTicketID("id; rm -rf /"))
	require.Error(t, mcp.SafeTicketID("$(whoami)"))
	require.Error(t, mcp.SafeTicketID("id | cat"))
}

func TestSafeTicketID_PathTraversal(t *testing.T) {
	t.Parallel()
	require.Error(t, mcp.SafeTicketID("../../../etc"))
	require.Error(t, mcp.SafeTicketID("/absolute"))
}

func TestSafeTicketID_TooLong(t *testing.T) {
	t.Parallel()
	long := "a" + strings.Repeat("b", 64) // 65 chars total
	require.Error(t, mcp.SafeTicketID(long))
}

// ---------------------------------------------------------------------------
// SafeProjectPath tests
// ---------------------------------------------------------------------------

func TestSafeProjectPath_ValidPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	projectDir := filepath.Join(root, "my-project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	resolved, err := mcp.SafeProjectPath("my-project", []string{root})

	require.NoError(t, err)
	// On macOS, /var is a symlink to /private/var. Resolve both for comparison.
	expectedResolved, err := filepath.EvalSymlinks(projectDir)
	require.NoError(t, err)
	assert.Equal(t, expectedResolved, resolved)
}

func TestSafeProjectPath_TraversalAttempt(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	_, err := mcp.SafeProjectPath("project/../../secret", []string{root})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside allowed")
}

func TestSafeProjectPath_AbsolutePath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	_, err := mcp.SafeProjectPath("/etc/passwd", []string{root})

	require.Error(t, err)
}

func TestSafeProjectPath_EmptyPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	_, err := mcp.SafeProjectPath("", []string{root})

	require.Error(t, err)
}

func TestSafeProjectPath_NoAllowedRoots(t *testing.T) {
	t.Parallel()

	_, err := mcp.SafeProjectPath("project", nil)

	require.Error(t, err)
}

func TestSafeProjectPath_Symlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outsideDir := t.TempDir()

	// Create a symlink inside root pointing outside.
	symlinkPath := filepath.Join(root, "escape")
	err := os.Symlink(outsideDir, symlinkPath)
	if err != nil {
		t.Skip("cannot create symlinks on this OS")
	}

	_, err = mcp.SafeProjectPath("escape", []string{root})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside allowed")
}
