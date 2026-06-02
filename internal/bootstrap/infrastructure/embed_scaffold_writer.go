// Package infrastructure provides adapters for the Bootstrap bounded context.
package infrastructure

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	rootembed "github.com/alto-cli/alto"
	bootstrapapp "github.com/alto-cli/alto/internal/bootstrap/application"
	"github.com/alto-cli/alto/internal/bootstrap/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// Compile-time interface check — matches the convention at git_committer_adapter.go:16.
var _ bootstrapapp.ScaffoldWriter = (*EmbedScaffoldWriter)(nil)

// ExpectedEmbedFileCount is the number of files the scaffold embed
// planner emits — verified at test time. The number reflects the
// GENERIC asset count after the runtime exclusion filter has removed
// *.project.md overlays, lifecycle/, and .gitkeep markers. If the
// alto-scaffold/ scaffold gains or loses GENERIC files, update this constant
// in the same change as the underlying file edit so the test catches
// drift. Tech-lead Phase 1 contract locked at 24; current scaffold
// state (post .gitkeep + overlay filter, + README.md, + scripts/bd-ripple,
// + commands/write-a-workflow-asset.md per alty-cli-766.6, + commands/rca.md
// + templates/beads-bug-template.md + templates/bug-rca-template.md for the
// bug-fix flow) yields 29.
const ExpectedEmbedFileCount = 29

// scriptsPrefix marks files that must be written with the executable bit
// set. Any embed path under alto-scaffold/scripts/ ships as a shell script
// and is unusable without 0o755.
const scriptsPrefix = embedRootPrefix + "/scripts/"

// embedRootPrefix is the path prefix every embed entry begins with — the
// //go:embed directive lists `alto-scaffold/...`, so fs.WalkDir results retain
// that prefix.
const embedRootPrefix = "alto-scaffold"

// EmbedScaffoldWriter writes the embedded alto-scaffold/ scaffold tree into a
// target project directory. The embed.FS handle is provided by the
// build-time resource package at the module root; this adapter owns
// the walk, filter, render, and write logic.
type EmbedScaffoldWriter struct {
	fs embed.FS
}

// NewEmbedScaffoldWriter constructs an adapter bound to the module-root
// scaffold FS. Composition root does not name the FS directly; this
// constructor is the only seam.
func NewEmbedScaffoldWriter() *EmbedScaffoldWriter {
	return &EmbedScaffoldWriter{fs: rootembed.ScaffoldFS}
}

// WriteScaffold extracts the embedded alto-scaffold/ scaffold into targetDir,
// substituting the five template parameters carried by params via
// text/template DATA binding (user values are data, never template
// source). When force is false the call fails with a wrapped
// ErrAlreadyExists if `<targetDir>/alto-scaffold/` already exists. When force is
// true the existing files are announced via [OVERWRITE] lines on stderr,
// then a plan-phase Lstat sweep refuses the entire operation if ANY
// target is a symlink (defence against an attacker pre-planting a symlink
// to trick --force into clobbering a victim file elsewhere), and only
// then are the files truncated and rewritten.
//
// Per-file writes use O_EXCL by default and O_TRUNC | O_NOFOLLOW only
// with force, closing the TOCTOU window between the existence check and
// the write. O_NOFOLLOW is the kernel-level backstop on POSIX; the Lstat
// sweep is the primary, cross-platform user-facing defence.
//
// targetDir is validated against path-traversal escape (leading or
// embedded "..") at entry. Empty string is normalised to "." so the
// CLI's hardcoded "." call site keeps working.
func (w *EmbedScaffoldWriter) WriteScaffold(ctx context.Context, targetDir string, params domain.ScaffoldParams, force bool) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	cleaned, err := sanitizeTargetDir(targetDir)
	if err != nil {
		return err
	}

	altoDir := filepath.Join(cleaned, "alto-scaffold")
	exists, err := pathExistsErr(altoDir)
	if err != nil {
		return fmt.Errorf("checking %s: %w", altoDir, err)
	}
	if exists && !force {
		return fmt.Errorf("target alto-scaffold/ already exists; use --force to overwrite: %w", domainerrors.ErrAlreadyExists)
	}

	plan, err := w.planFiles()
	if err != nil {
		return fmt.Errorf("planning embed files: %w", err)
	}

	if exists && force {
		for _, srcPath := range plan {
			rel := strings.TrimPrefix(srcPath, embedRootPrefix+"/")
			_, _ = fmt.Fprintf(os.Stderr, "[OVERWRITE] alto-scaffold/%s\n", filepath.ToSlash(rel))
		}
		// Plan-phase symlink sweep: abort the ENTIRE operation if any
		// target path resolves to a symlink. Must happen AFTER the
		// preview lines (so the operator sees both the intent and the
		// abort reason) and BEFORE the write loop (so no file is touched
		// when any target is hostile). O_NOFOLLOW on the per-file open
		// is the kernel backstop; this sweep is the primary defence.
		if err := w.refuseSymlinkTargets(cleaned, plan); err != nil {
			return err
		}
	}

	for _, srcPath := range plan {
		if err := w.writeFile(srcPath, cleaned, params, force); err != nil {
			return fmt.Errorf("writing %s: %w", srcPath, err)
		}
	}
	return nil
}

// sanitizeTargetDir applies filepath.Clean, normalises "" to ".", and
// rejects any input that escapes the operator's cwd via leading or
// embedded ".." segments. Returns the cleaned path or a wrapped
// ErrInvariantViolation.
func sanitizeTargetDir(targetDir string) (string, error) {
	cleaned := filepath.Clean(targetDir)
	if cleaned == "" || cleaned == "." {
		return ".", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("targetDir %q escapes operator cwd: %w", targetDir, domainerrors.ErrInvariantViolation)
	}
	for _, p := range strings.Split(cleaned, string(filepath.Separator)) {
		if p == ".." {
			return "", fmt.Errorf("targetDir %q contains traversal segment: %w", targetDir, domainerrors.ErrInvariantViolation)
		}
	}
	return cleaned, nil
}

// refuseSymlinkTargets walks the planned destination paths under
// <targetDir>/alto-scaffold/ and aborts the operation if any target is a
// symlink. Returns the FIRST hit with a wrapped ErrInvariantViolation;
// nil if every target is either absent or a regular file.
func (w *EmbedScaffoldWriter) refuseSymlinkTargets(targetDir string, plan []string) error {
	for _, srcPath := range plan {
		rel := strings.TrimPrefix(srcPath, embedRootPrefix+"/")
		dst := filepath.Join(targetDir, embedRootPrefix, filepath.FromSlash(rel))
		info, err := os.Lstat(dst)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("lstat %s: %w", dst, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to overwrite symlink at %q; resolve manually then re-run: %w", dst, domainerrors.ErrInvariantViolation)
		}
	}
	return nil
}

// planFiles walks the embed FS and returns the list of regular-file paths
// that should be written, skipping defence-in-depth-excluded entries
// (*.project.md overlays, lifecycle/).
func (w *EmbedScaffoldWriter) planFiles() ([]string, error) {
	var paths []string
	err := fs.WalkDir(w.fs, embedRootPrefix, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if isExcludedEmbedPath(path) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking embed: %w", err)
	}
	return paths, nil
}

// writeFile renders one embedded file into the target directory.
func (w *EmbedScaffoldWriter) writeFile(srcPath string, targetDir string, params domain.ScaffoldParams, force bool) error {
	raw, err := w.fs.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading embed %s: %w", srcPath, err)
	}

	rendered, err := renderTemplate(srcPath, raw, params)
	if err != nil {
		return err
	}

	rel := strings.TrimPrefix(srcPath, embedRootPrefix+"/")
	dst := filepath.Join(targetDir, embedRootPrefix, filepath.FromSlash(rel))
	if mkErr := os.MkdirAll(filepath.Dir(dst), 0o755); mkErr != nil {
		return fmt.Errorf("creating dir %s: %w", filepath.Dir(dst), mkErr)
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL | noFollow
	if force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC | noFollow
	}
	mode := os.FileMode(0o644)
	if strings.HasPrefix(srcPath, scriptsPrefix) {
		mode = 0o755
	}
	f, err := os.OpenFile(dst, flags, mode) //nolint:gosec // dst path is derived from sanitized targetDir + embed FS contents; O_NOFOLLOW + plan-phase Lstat sweep guard against symlink overwrites
	if err != nil {
		return fmt.Errorf("opening %s: %w", dst, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(rendered); err != nil {
		return fmt.Errorf("writing %s: %w", dst, err)
	}
	// O_TRUNC reuses the existing inode and keeps its mode bits, so a
	// stale 0o644 from a previous --with-scaffold run would leave the
	// script non-executable. Chmod after write to make the bit
	// idempotent regardless of prior state. Use f.Chmod (fchmod on the
	// already-open, O_NOFOLLOW-validated fd) rather than path-based
	// os.Chmod: a path-based chmod follows symlinks and would re-open the
	// local TOCTOU symlink-swap vector that the O_NOFOLLOW open closes.
	if mode == 0o755 {
		if err := f.Chmod(mode); err != nil {
			return fmt.Errorf("chmod %s: %w", dst, err)
		}
	}
	return nil
}

// renderTemplate parses raw as a text/template and executes it against
// params. User-supplied values flow as struct DATA — a ProjectName of
// "{{.Evil}}" is rendered as literal text, never as a re-evaluated
// template expression. Templates that fail to parse are returned with a
// wrapping error; non-template files (no {{ }} pairs) round-trip
// unchanged through this code path.
func renderTemplate(srcPath string, raw []byte, params domain.ScaffoldParams) ([]byte, error) {
	tmpl, err := template.New(srcPath).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", srcPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, fmt.Errorf("executing template %s: %w", srcPath, err)
	}
	return buf.Bytes(), nil
}

// isExcludedEmbedPath enforces the defence-in-depth runtime filter: even
// if a path slipped through the //go:embed allowlist, *.project.md
// overlays and lifecycle/ entries are skipped.
func isExcludedEmbedPath(path string) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, ".project.md") {
		return true
	}
	if base == ".gitkeep" {
		return true
	}
	if strings.HasPrefix(path, embedRootPrefix+"/lifecycle/") {
		return true
	}
	return false
}

// pathExistsErr reports whether the path exists at all (file or directory),
// distinguishing real errors from "not found". Renamed from `pathExists`
// to avoid colliding with the bool-only helper in filesystem_project_detector.go.
func pathExistsErr(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat %s: %w", p, err)
}

// Test-only helpers (WalkEmbedForTest, EmbedFilesMatchingForTest,
// EmbedReadForTest, PlanFilesMatchingForTest, RenderTemplateForTest) live
// in embed_scaffold_writer_export_test.go so the production translation
// unit does not import "testing".
