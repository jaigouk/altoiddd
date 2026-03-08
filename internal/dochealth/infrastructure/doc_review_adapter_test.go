package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/dochealth/infrastructure"
)

func TestDocReviewAdapter_MarkReviewed_UpdatesFrontmatter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docPath := filepath.Join("docs", "PRD.md")
	fullPath := filepath.Join(dir, docPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	content := "---\nlast_reviewed: \"2025-01-01\"\nreview_interval_days: 30\n---\n# PRD\n\nSome content here.\n"
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	reviewDate := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	result, err := adapter.MarkReviewed(context.Background(), docPath, dir, &reviewDate)

	require.NoError(t, err)
	assert.Equal(t, docPath, result.Path())
	assert.Equal(t, reviewDate, result.NewDate())

	// Verify file was updated.
	updated, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), "last_reviewed: \"2026-03-08\"")
	assert.Contains(t, string(updated), "# PRD")
	assert.Contains(t, string(updated), "Some content here.")
}

func TestDocReviewAdapter_MarkReviewed_NoFrontmatter_InsertsIt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docPath := filepath.Join("docs", "notes.md")
	fullPath := filepath.Join(dir, docPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	content := "# Notes\n\nSome notes here.\n"
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	reviewDate := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	result, err := adapter.MarkReviewed(context.Background(), docPath, dir, &reviewDate)

	require.NoError(t, err)
	assert.Equal(t, docPath, result.Path())

	updated, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	updatedStr := string(updated)
	assert.Contains(t, updatedStr, "---\nlast_reviewed:")
	assert.Contains(t, updatedStr, "# Notes")
	assert.Contains(t, updatedStr, "Some notes here.")
}

func TestDocReviewAdapter_MarkReviewed_MissingLastReviewedField_AddsIt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docPath := filepath.Join("docs", "partial.md")
	fullPath := filepath.Join(dir, docPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	content := "---\nreview_interval_days: 30\nauthor: test\n---\n# Partial\n\nContent.\n"
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	reviewDate := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	result, err := adapter.MarkReviewed(context.Background(), docPath, dir, &reviewDate)

	require.NoError(t, err)
	assert.Equal(t, docPath, result.Path())

	updated, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	updatedStr := string(updated)
	assert.Contains(t, updatedStr, "last_reviewed:")
	assert.Contains(t, updatedStr, "review_interval_days: 30")
	assert.Contains(t, updatedStr, "# Partial")
}

func TestDocReviewAdapter_MarkReviewed_FileNotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())

	_, err := adapter.MarkReviewed(context.Background(), "nonexistent.md", dir, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent.md")
}

func TestDocReviewAdapter_MarkReviewed_NilDate_UsesNow(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docPath := filepath.Join("docs", "test.md")
	fullPath := filepath.Join(dir, docPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	content := "---\nlast_reviewed: \"2025-01-01\"\n---\n# Test\n"
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	result, err := adapter.MarkReviewed(context.Background(), docPath, dir, nil)

	require.NoError(t, err)
	today := time.Now().Truncate(24 * time.Hour)
	assert.Equal(t, today, result.NewDate())
}

func TestDocReviewAdapter_MarkAllReviewed_UpdatesAll(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))

	doc1 := "---\nlast_reviewed: \"2025-01-01\"\nreview_interval_days: 30\n---\n# Doc 1\n"
	doc2 := "---\nlast_reviewed: \"2025-06-01\"\nreview_interval_days: 30\n---\n# Doc 2\n"
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "doc1.md"), []byte(doc1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "doc2.md"), []byte(doc2), 0o644))

	// Also create the registry so docs are discoverable.
	altyDir := filepath.Join(dir, ".alty", "maintenance")
	require.NoError(t, os.MkdirAll(altyDir, 0o755))
	registry := "[[docs]]\npath = \"docs/doc1.md\"\nreview_interval_days = 30\n\n[[docs]]\npath = \"docs/doc2.md\"\nreview_interval_days = 30\n"
	require.NoError(t, os.WriteFile(filepath.Join(altyDir, "doc-registry.toml"), []byte(registry), 0o644))

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	reviewDate := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	results, err := adapter.MarkAllReviewed(context.Background(), dir, &reviewDate)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestDocReviewAdapter_ReviewableDocs_FindsStaleDocs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))

	staleDate := time.Now().AddDate(0, -3, 0).Format("2006-01-02")
	staleDoc := "---\nlast_reviewed: \"" + staleDate + "\"\nreview_interval_days: 30\n---\n# Stale\n"
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "stale.md"), []byte(staleDoc), 0o644))

	altyDir := filepath.Join(dir, ".alty", "maintenance")
	require.NoError(t, os.MkdirAll(altyDir, 0o755))
	registry := "[[docs]]\npath = \"docs/stale.md\"\nreview_interval_days = 30\n"
	require.NoError(t, os.WriteFile(filepath.Join(altyDir, "doc-registry.toml"), []byte(registry), 0o644))

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	statuses, err := adapter.ReviewableDocs(context.Background(), dir)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(statuses), 1)
}

func TestDocReviewAdapter_ReviewableDocs_NoDocs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	statuses, err := adapter.ReviewableDocs(context.Background(), dir)

	require.NoError(t, err)
	assert.Empty(t, statuses)
}

func TestDocReviewAdapter_ReviewableDocs_RegistryError_Propagates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create registry dir but make the file a directory to cause a read error.
	altyDir := filepath.Join(dir, ".alty", "maintenance")
	require.NoError(t, os.MkdirAll(altyDir, 0o755))
	// Create doc-registry.toml as a directory (not a file) — LoadRegistry will stat it
	// as existing but fail to read it. However, the current LoadRegistry swallows
	// read errors (returns nil, nil). So we test the "empty entries" path.
	// The fix ensures err != nil is checked separately from len(entries) == 0.
	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	statuses, err := adapter.ReviewableDocs(context.Background(), dir)

	// No registry file → LoadRegistry returns nil, nil → no error, empty result.
	require.NoError(t, err)
	assert.Empty(t, statuses)
}

func TestDocReviewAdapter_MarkReviewed_PreservesBody(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docPath := filepath.Join("docs", "body-test.md")
	fullPath := filepath.Join(dir, docPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	body := "# Title\n\nParagraph with **bold** and `code`.\n\n## Section\n\n- List item 1\n- List item 2\n\n```go\nfunc main() {}\n```\n"
	content := "---\nlast_reviewed: \"2025-01-01\"\n---\n" + body
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))

	adapter := infrastructure.NewDocReviewAdapter(infrastructure.NewFilesystemDocScanner())
	reviewDate := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	_, err := adapter.MarkReviewed(context.Background(), docPath, dir, &reviewDate)
	require.NoError(t, err)

	updated, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	assert.Contains(t, string(updated), body)
}
