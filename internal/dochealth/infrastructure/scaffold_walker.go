package infrastructure

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// scaffoldAssetDirs lists the subdirectories of altoDir that contain
// validated scaffold assets (files with the canonical 8-field
// frontmatter). Other `alto-scaffold/` subtrees (templates/, knowledge/, skills/,
// lifecycle/deprecated/) carry their own minimal schemas and are out of
// scope.
//
// lifecycle/in-progress/ was added by alty-cli-ihk so the canonical
// `--paths=alto-scaffold/` sweep also flags in-progress assets via the 8-field
// rules — unblocks the alty-cli-766.6 meta-skill FIX-1 transition AC.
var scaffoldAssetDirs = []string{"commands", "agents", "lifecycle/in-progress"}

// FilesystemScaffoldWalker enumerates `alto-scaffold/commands/*.md` and
// `alto-scaffold/agents/*.md`, parsing each into a ScaffoldAsset.
//
// Resource defences:
//   - filepath.WalkDir handles symlink cycles deterministically (does not
//     follow symlinks by default).
//   - File-descriptor leak: every os.ReadFile call is self-contained.
//   - Memory: each asset body is read fully into memory; the canonical
//     scaffold size budget (~500 lines per asset per the fast-follow
//     BodySizeRule) keeps total RAM bounded.
type FilesystemScaffoldWalker struct {
	parser *scaffoldFrontmatterParser
}

// NewFilesystemScaffoldWalker constructs a walker with the embedded
// scaffold frontmatter parser.
func NewFilesystemScaffoldWalker() *FilesystemScaffoldWalker {
	return &FilesystemScaffoldWalker{parser: newScaffoldFrontmatterParser()}
}

// Compile-time interface check.
var _ dochealthapp.ScaffoldWalker = (*FilesystemScaffoldWalker)(nil)

// Walk enumerates scaffold-asset subdirs of altoDir, returning every `.md`
// file as a parsed ScaffoldAsset. Non-existent altoDir or scaffold subdir
// is NOT an error — an empty corpus is the valid "no scaffold here" answer.
//
// Two invocation forms are supported:
//  1. Canonical: `altoDir = alto-scaffold/` — walks `commands/`, `agents/`,
//     `lifecycle/in-progress/` subdirs (per scaffoldAssetDirs).
//  2. Flat-dir auto-detect: `altoDir = alto-scaffold/lifecycle/in-progress/` —
//     when none of the known scaffold subdirs exist beneath altoDir, the
//     walker treats altoDir itself as a flat directory of `*.md` files.
//     This supports `alto doc-health --paths=alto-scaffold/lifecycle/in-progress/`.
func (w *FilesystemScaffoldWalker) Walk(_ context.Context, altoDir string) ([]dochealthdomain.ScaffoldAsset, error) {
	cleaned := filepath.Clean(altoDir)
	info, err := os.Stat(cleaned)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat %s: %w", cleaned, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scaffold root %s is not a directory", cleaned)
	}

	var assets []dochealthdomain.ScaffoldAsset
	knownFound := false
	for _, sub := range scaffoldAssetDirs {
		subDir := filepath.Join(cleaned, sub)
		subInfo, serr := os.Stat(subDir)
		if serr != nil || !subInfo.IsDir() {
			continue
		}
		knownFound = true
		walked, werr := w.walkSubtree(subDir)
		if werr != nil {
			return nil, fmt.Errorf("walking %s: %w", subDir, werr)
		}
		assets = append(assets, walked...)
	}

	// Flat-dir auto-detect: if cleaned contains no known scaffold subdir,
	// walk cleaned itself as a flat *.md directory. This makes
	// `--paths=alto-scaffold/lifecycle/in-progress/` (direct invocation) work.
	if !knownFound {
		walked, werr := w.walkSubtree(cleaned)
		if werr != nil {
			return nil, fmt.Errorf("walking %s: %w", cleaned, werr)
		}
		assets = append(assets, walked...)
	}
	return assets, nil
}

// walkSubtree recursively enumerates `.md` files under subDir.
func (w *FilesystemScaffoldWalker) walkSubtree(subDir string) ([]dochealthdomain.ScaffoldAsset, error) {
	var assets []dochealthdomain.ScaffoldAsset
	walkErr := filepath.WalkDir(subDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		// .gitkeep and other non-markdown sentinels are already filtered
		// by the .md suffix check above.
		content, rerr := os.ReadFile(path) //nolint:gosec // path produced by WalkDir under operator-supplied altoDir.
		if rerr != nil {
			return fmt.Errorf("read %s: %w", path, rerr)
		}
		fm, body, lineN, _, perr := w.parser.Parse(string(content))
		if perr != nil {
			return fmt.Errorf("parse %s: %w", path, perr)
		}
		isOverlay := strings.HasSuffix(d.Name(), ".project.md")
		// Capture mtime so LifecycleStalenessRule can compute staleness
		// without doing I/O (rules MUST stay pure).
		fi, ierr := d.Info()
		if ierr != nil {
			return fmt.Errorf("stat %s: %w", path, ierr)
		}
		asset, aerr := dochealthdomain.NewScaffoldAssetWithModTime(
			filepath.ToSlash(path), fm, body, lineN, isOverlay, fi.ModTime(),
		)
		if aerr != nil {
			return fmt.Errorf("construct asset %s: %w", path, aerr)
		}
		assets = append(assets, asset)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", subDir, walkErr)
	}
	return assets, nil
}
