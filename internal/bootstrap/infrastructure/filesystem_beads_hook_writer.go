// Package infrastructure provides adapters for the Bootstrap bounded context.
package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	bootstrapapp "github.com/alto-cli/alto/internal/bootstrap/application"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// Compile-time interface check.
var _ bootstrapapp.BeadsHookWriter = (*FilesystemBeadsHookWriter)(nil)

// posixPostCloseHook is the content of `.beads/hooks/post-close` on
// POSIX systems. Kept as a top-of-file constant per the design
// decision: the adapter MUST NOT inline templates in the function body.
const posixPostCloseHook = `#!/usr/bin/env bash
exec alto ticket-ripple "$@"
`

// windowsPostCloseHookBat is the content of
// `.beads/hooks/post-close.bat` on Windows hosts. CRLF line endings are
// used so cmd.exe accepts the script verbatim.
const windowsPostCloseHookBat = "@echo off\r\nalto ticket-ripple %*\r\n"

// posixHookMode is the on-disk mode for the executable POSIX hook.
const posixHookMode os.FileMode = 0o755

// windowsHookMode is the on-disk mode for the Windows shim (non-exec).
const windowsHookMode os.FileMode = 0o644

// FilesystemBeadsHookWriter implements BeadsHookWriter by writing the
// post-close hook + Windows shim to <targetDir>/.beads/hooks/.
type FilesystemBeadsHookWriter struct {
	previewWriter func(format string, a ...any) (int, error)
}

// NewFilesystemBeadsHookWriter constructs the default adapter. Overwrite
// preview lines go to os.Stderr — matching the EmbedScaffoldWriter
// pattern at embed_scaffold_writer.go:109.
func NewFilesystemBeadsHookWriter() *FilesystemBeadsHookWriter {
	return &FilesystemBeadsHookWriter{
		previewWriter: func(format string, a ...any) (int, error) {
			return fmt.Fprintf(os.Stderr, format, a...)
		},
	}
}

// NewFilesystemBeadsHookWriterWithPreview constructs an adapter with a
// custom preview sink — used by tests to capture [OVERWRITE] lines
// without polluting CI stderr.
func NewFilesystemBeadsHookWriterWithPreview(preview func(format string, a ...any) (int, error)) *FilesystemBeadsHookWriter {
	return &FilesystemBeadsHookWriter{previewWriter: preview}
}

// WriteBeadsPostCloseHook writes the POSIX hook + Windows shim to
// <targetDir>/.beads/hooks/. primaryTool is currently unused — reserved
// for future per-tool template variants.
func (w *FilesystemBeadsHookWriter) WriteBeadsPostCloseHook(
	ctx context.Context,
	targetDir string,
	_ string,
	force bool,
) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}
	cleaned, err := sanitizeTargetDir(targetDir)
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(cleaned, ".beads", "hooks")
	if mkErr := os.MkdirAll(hooksDir, 0o755); mkErr != nil {
		return fmt.Errorf("creating %s: %w", hooksDir, mkErr)
	}

	files := []hookFile{
		{
			path:    filepath.Join(hooksDir, "post-close"),
			content: posixPostCloseHook,
			mode:    posixHookMode,
		},
		{
			path:    filepath.Join(hooksDir, "post-close.bat"),
			content: windowsPostCloseHookBat,
			mode:    windowsHookMode,
		},
	}

	for _, f := range files {
		if err := w.writeOne(f, force); err != nil {
			return err
		}
	}
	return nil
}

type hookFile struct {
	path    string
	content string
	mode    os.FileMode
}

// writeOne handles the three conflict cases per the design decision:
//   - file does not exist: create with O_EXCL + O_NOFOLLOW.
//   - file exists and content matches: no-op (idempotent re-run).
//   - file exists and content differs:
//   - force=false: return wrapped ErrAlreadyExists.
//   - force=true: emit [OVERWRITE] preview, then truncate + write.
//
// The chmod-after-write pattern is lifted verbatim from
// embed_scaffold_writer.go:237 to defeat the path-based TOCTOU symlink
// swap.
func (w *FilesystemBeadsHookWriter) writeOne(f hookFile, force bool) error {
	existing, exists, err := readIfExists(f.path)
	if err != nil {
		return err
	}
	if exists {
		if existing == f.content {
			return nil // idempotent no-op
		}
		if !force {
			return fmt.Errorf("hook at %s exists with different content; rerun with --force-hooks to overwrite: %w",
				f.path, domainerrors.ErrAlreadyExists)
		}
		_, _ = w.previewWriter("[OVERWRITE] %s\n", f.path)
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL | noFollow
	if exists && force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC | noFollow
	}

	fd, err := os.OpenFile(f.path, flags, f.mode) //nolint:gosec // path derived from sanitized targetDir + fixed constant suffix; O_NOFOLLOW closes the TOCTOU window
	if err != nil {
		return fmt.Errorf("opening %s: %w", f.path, err)
	}
	defer func() { _ = fd.Close() }()
	if _, err := fd.Write([]byte(f.content)); err != nil {
		return fmt.Errorf("writing %s: %w", f.path, err)
	}
	// O_TRUNC reuses the existing inode and its mode bits, so a stale
	// 0o644 from a prior --no-hooks state would leave the POSIX hook
	// non-executable. Chmod the open O_NOFOLLOW fd (fchmod) to make
	// the bit idempotent. Same TOCTOU rationale as
	// embed_scaffold_writer.go:237.
	if err := fd.Chmod(f.mode); err != nil {
		return fmt.Errorf("chmod %s: %w", f.path, err)
	}
	return nil
}

// readIfExists returns (content, true, nil) when the path exists as a
// regular file, ("", false, nil) when absent, and ("", false, err) on
// any other Lstat / read failure. A symlink at the target is rejected
// with a wrapped ErrInvariantViolation — the operator must resolve it
// manually before re-running.
func readIfExists(path string) (string, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("lstat %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", false, fmt.Errorf("refusing to overwrite symlink at %q; resolve manually then re-run: %w",
			path, domainerrors.ErrInvariantViolation)
	}
	raw, err := os.ReadFile(path) //nolint:gosec // path validated via Lstat to be a regular file
	if err != nil {
		return "", false, fmt.Errorf("reading %s: %w", path, err)
	}
	return string(raw), true, nil
}
