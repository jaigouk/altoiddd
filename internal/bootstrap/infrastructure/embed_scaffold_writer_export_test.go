package infrastructure

// This file holds test-only helpers that introspect the embedded scaffold
// FS. It uses the same package name as the production file (no `_test`
// suffix on the package declaration) so the helpers retain access to
// unexported fields like EmbedScaffoldWriter.fs. The `_test.go` filename
// suffix scopes the file to test builds only — production binaries link
// against neither this file nor the "testing" package.
//
// Moved here from embed_scaffold_writer.go per Round 1 Fix 3 (QA-MAJ-MIN-1)
// so the production translation unit no longer imports "testing".

import (
	"io/fs"
	"testing"

	"github.com/alto-cli/alto/internal/bootstrap/domain"
)

// WalkEmbedForTest counts files in the embed (post-exclusion), for use by
// the file-count assertion test. Calling code MUST pass *testing.T to
// signal test-only intent; the parameter is otherwise unused.
func WalkEmbedForTest(t *testing.T) int {
	t.Helper()
	w := NewEmbedScaffoldWriter()
	plan, err := w.planFiles()
	if err != nil {
		t.Fatalf("walk embed: %v", err)
	}
	return len(plan)
}

// EmbedFilesMatchingForTest returns embed-relative paths whose basenames
// satisfy match. Test-only; *testing.T pin marks intent.
func EmbedFilesMatchingForTest(t *testing.T, match func(name string) bool) []string {
	t.Helper()
	w := NewEmbedScaffoldWriter()
	var hits []string
	err := fs.WalkDir(w.fs, embedRootPrefix, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if match(path) {
			hits = append(hits, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embed: %v", err)
	}
	return hits
}

// EmbedReadForTest reads one embedded file. Test-only.
func EmbedReadForTest(t *testing.T, path string) []byte {
	t.Helper()
	w := NewEmbedScaffoldWriter()
	content, err := w.fs.ReadFile(path)
	if err != nil {
		t.Fatalf("read embed %s: %v", path, err)
	}
	return content
}

// PlanFilesMatchingForTest returns paths from the post-filter writeset
// (i.e. after runtime exclusion of overlays and lifecycle/) whose names
// satisfy match. Test-only.
func PlanFilesMatchingForTest(t *testing.T, match func(name string) bool) []string {
	t.Helper()
	w := NewEmbedScaffoldWriter()
	plan, err := w.planFiles()
	if err != nil {
		t.Fatalf("plan files: %v", err)
	}
	var hits []string
	for _, p := range plan {
		if match(p) {
			hits = append(hits, p)
		}
	}
	return hits
}

// RenderTemplateForTest exposes the renderTemplate helper for direct
// unit testing of the text/template data-binding security property.
// Test-only.
func RenderTemplateForTest(t *testing.T, srcPath string, raw []byte, params domain.ScaffoldParams) ([]byte, error) {
	t.Helper()
	return renderTemplate(srcPath, raw, params)
}
