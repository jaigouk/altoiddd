package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapapp "github.com/alto-cli/alto/internal/bootstrap/application"
	"github.com/alto-cli/alto/internal/bootstrap/infrastructure"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// Compile-time interface check.
var _ bootstrapapp.BeadsHookWriter = (*infrastructure.FilesystemBeadsHookWriter)(nil)

func newWriterWithSink(t *testing.T) (*infrastructure.FilesystemBeadsHookWriter, *strings.Builder) {
	t.Helper()
	var sink strings.Builder
	w := infrastructure.NewFilesystemBeadsHookWriterWithPreview(func(format string, a ...any) (int, error) {
		s := strings.NewReplacer().Replace(format)
		_ = s
		return sink.WriteString(format)
	})
	return w, &sink
}

func TestFilesystemBeadsHookWriter_NewDefaultsAreSane(t *testing.T) {
	t.Parallel()

	w := infrastructure.NewFilesystemBeadsHookWriter()

	require.NotNil(t, w)
}

func TestFilesystemBeadsHookWriter_Write_CreatesExecutablePostCloseHook(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, _ := newWriterWithSink(t)

	err := w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false)

	require.NoError(t, err)
	hookPath := filepath.Join(dir, ".beads", "hooks", "post-close")
	info, statErr := os.Stat(hookPath)
	require.NoError(t, statErr)
	assert.False(t, info.IsDir())
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o755), info.Mode().Perm(), "POSIX hook must be 0o755")
	}
}

func TestFilesystemBeadsHookWriter_Write_PostCloseHookExecsAltoTicketRipple(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, _ := newWriterWithSink(t)

	err := w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false)
	require.NoError(t, err)

	body, readErr := os.ReadFile(filepath.Join(dir, ".beads", "hooks", "post-close"))
	require.NoError(t, readErr)
	got := string(body)
	assert.True(t, strings.HasPrefix(got, "#!/usr/bin/env bash"))
	assert.Contains(t, got, `exec alto ticket-ripple "$@"`)
}

func TestFilesystemBeadsHookWriter_Write_AlwaysWritesWindowsBatShim(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, _ := newWriterWithSink(t)

	err := w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false)
	require.NoError(t, err)

	body, readErr := os.ReadFile(filepath.Join(dir, ".beads", "hooks", "post-close.bat"))
	require.NoError(t, readErr)
	got := string(body)
	assert.Contains(t, got, "@echo off")
	assert.Contains(t, got, "alto ticket-ripple %*")
	assert.Contains(t, got, "\r\n", "bat shim must use CRLF line endings")
}

func TestFilesystemBeadsHookWriter_Write_WhenHookExistsAndSame_NoOp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, sink := newWriterWithSink(t)

	require.NoError(t, w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false))
	// Second call must succeed without OVERWRITE preview.
	require.NoError(t, w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false))

	assert.NotContains(t, sink.String(), "[OVERWRITE]")
}

func TestFilesystemBeadsHookWriter_Write_WhenHookExistsAndDiffers_ReturnsErrAlreadyExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	hooksDir := filepath.Join(dir, ".beads", "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(hooksDir, "post-close"), []byte("# user hook\n"), 0o755))

	w, _ := newWriterWithSink(t)
	err := w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false)

	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrAlreadyExists)
	assert.Contains(t, err.Error(), "--force-hooks")
}

func TestFilesystemBeadsHookWriter_Write_WhenForce_OverwritesExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	hooksDir := filepath.Join(dir, ".beads", "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0o755))
	hookPath := filepath.Join(hooksDir, "post-close")
	require.NoError(t, os.WriteFile(hookPath, []byte("# stale hook\n"), 0o644))

	w, sink := newWriterWithSink(t)
	err := w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", true)

	require.NoError(t, err)
	body, _ := os.ReadFile(hookPath)
	assert.Contains(t, string(body), "alto ticket-ripple")
	assert.NotContains(t, string(body), "stale hook")
	assert.Contains(t, sink.String(), "[OVERWRITE]")
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(hookPath)
		assert.Equal(t, os.FileMode(0o755), info.Mode().Perm(),
			"overwrite must restore 0o755 even when prior inode was 0o644")
	}
}

func TestFilesystemBeadsHookWriter_Write_CreatesParentDirectoriesIfMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// .beads/ does not exist at all.
	w, _ := newWriterWithSink(t)

	err := w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false)

	require.NoError(t, err)
	info, statErr := os.Stat(filepath.Join(dir, ".beads", "hooks"))
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
}

func TestFilesystemBeadsHookWriter_Write_RejectsSymlinkAtTargetPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink test requires POSIX")
	}
	t.Parallel()

	dir := t.TempDir()
	hooksDir := filepath.Join(dir, ".beads", "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0o755))
	victim := filepath.Join(dir, "victim.txt")
	require.NoError(t, os.WriteFile(victim, []byte("private"), 0o600))
	require.NoError(t, os.Symlink(victim, filepath.Join(hooksDir, "post-close")))

	w, _ := newWriterWithSink(t)
	err := w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", true)

	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)

	body, _ := os.ReadFile(victim)
	assert.Equal(t, "private", string(body), "victim file must remain untouched")
}

func TestFilesystemBeadsHookWriter_Write_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w, _ := newWriterWithSink(t)
	err := w.WriteBeadsPostCloseHook(ctx, dir, "claude", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestFilesystemBeadsHookWriter_Write_TargetDirCannotEscapeCwd(t *testing.T) {
	t.Parallel()

	w, _ := newWriterWithSink(t)
	err := w.WriteBeadsPostCloseHook(context.Background(), "../escape", "claude", false)

	require.Error(t, err)
}

func TestFilesystemBeadsHookWriter_Write_IdempotentAcrossManyCalls(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, _ := newWriterWithSink(t)

	for i := 0; i < 5; i++ {
		require.NoError(t, w.WriteBeadsPostCloseHook(context.Background(), dir, "claude", false))
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".beads", "hooks", "post-close"))
	assert.Contains(t, string(body), "alto ticket-ripple")
}
