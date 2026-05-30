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
// frontmatter). Other `.alto/` subtrees (templates/, knowledge/,
// lifecycle/, skills/ — when populated) carry their own minimal schemas
// and are out of scope for this ticket. The fast-follow ticket extends
// coverage to template/skill assets.
var scaffoldAssetDirs = []string{"commands", "agents"}

// FilesystemScaffoldWalker enumerates `.alto/commands/*.md` and
// `.alto/agents/*.md`, parsing each into a ScaffoldAsset.
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
	for _, sub := range scaffoldAssetDirs {
		subDir := filepath.Join(cleaned, sub)
		subInfo, serr := os.Stat(subDir)
		if serr != nil || !subInfo.IsDir() {
			continue
		}
		walked, werr := w.walkSubtree(subDir)
		if werr != nil {
			return nil, fmt.Errorf("walking %s: %w", subDir, werr)
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
		asset, aerr := dochealthdomain.NewScaffoldAsset(
			filepath.ToSlash(path), fm, body, lineN, isOverlay,
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
