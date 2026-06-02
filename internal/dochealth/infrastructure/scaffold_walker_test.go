package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAltoTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, sub := range []string{"commands", "agents"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, sub), 0o755))
	}
	return root
}

const validScaffoldFrontmatter = `---
name: foo
description: x
kind: command
phase: groom
when_to_use: test
tools: Read
bash_substitution_policy: none
license: Apache-2.0
---
body
`

func TestFilesystemScaffoldWalker_EmptyAltoDir_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestFilesystemScaffoldWalker_NonexistentAltoDir_NoError(t *testing.T) {
	t.Parallel()
	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), "/nonexistent/path/foo/bar")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestFilesystemScaffoldWalker_WalksCommandsAndAgents(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "commands", "foo.md"), []byte(validScaffoldFrontmatter), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "agents", "bar.md"), []byte(validScaffoldFrontmatter), 0o600))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestFilesystemScaffoldWalker_IgnoresNonMd(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "commands", "ignored.txt"), []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "commands", "ok.md"), []byte(validScaffoldFrontmatter), 0o600))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestFilesystemScaffoldWalker_IgnoresGitkeep(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "commands", ".gitkeep"), []byte(""), 0o600))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestFilesystemScaffoldWalker_DetectsOverlay(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "commands", "foo.md"), []byte(validScaffoldFrontmatter), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "commands", "foo.project.md"), []byte("# Pure overlay"), 0o600))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	require.Len(t, got, 2)

	overlays := 0
	for _, a := range got {
		if a.IsOverlay() {
			overlays++
		}
	}
	assert.Equal(t, 1, overlays)
}

func TestFilesystemScaffoldWalker_SkipsUnknownSubdirs(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	// templates/ is not a scaffold-asset dir per scaffoldAssetDirs.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "templates"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "templates", "epic.md"), []byte("# Template"), 0o600))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	assert.Empty(t, got, "templates/ must not be walked")
}

func TestFilesystemScaffoldWalker_LifecycleInProgress_Walks(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	require.NoError(t, os.MkdirAll(filepath.Join(root, "lifecycle", "in-progress"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "lifecycle", "in-progress", "draft.md"),
		[]byte(validScaffoldFrontmatter), 0o600,
	))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Contains(t, got[0].Path(), "lifecycle/in-progress/draft.md")
}

func TestFilesystemScaffoldWalker_LifecycleInProgress_SweepsCanonical(t *testing.T) {
	t.Parallel()
	// Canonical invocation: --paths=alto-scaffold/ must include lifecycle/in-progress/ files.
	root := setupAltoTree(t)
	require.NoError(t, os.MkdirAll(filepath.Join(root, "lifecycle", "in-progress"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "commands", "foo.md"), []byte(validScaffoldFrontmatter), 0o600))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "lifecycle", "in-progress", "draft.md"),
		[]byte(validScaffoldFrontmatter), 0o600,
	))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	assert.Len(t, got, 2, "must walk commands/ + lifecycle/in-progress/")
}

func TestFilesystemScaffoldWalker_FlatDirAutoDetect_Walks(t *testing.T) {
	t.Parallel()
	// Direct invocation: --paths=alto-scaffold/lifecycle/in-progress/ must walk *.md
	// in that dir directly when no known subdir exists at that level.
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.md"), []byte(validScaffoldFrontmatter), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "b.md"), []byte(validScaffoldFrontmatter), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "ignore.txt"), []byte("x"), 0o600))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestFilesystemScaffoldWalker_CapturesModTime(t *testing.T) {
	t.Parallel()
	root := setupAltoTree(t)
	p := filepath.Join(root, "commands", "foo.md")
	require.NoError(t, os.WriteFile(p, []byte(validScaffoldFrontmatter), 0o600))

	w := NewFilesystemScaffoldWalker()
	got, err := w.Walk(context.TODO(), root)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.False(t, got[0].ModTime().IsZero(), "walker must capture mtime")
}
