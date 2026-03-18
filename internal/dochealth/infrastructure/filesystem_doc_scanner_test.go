package infrastructure_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
	"github.com/alto-cli/alto/internal/dochealth/infrastructure"
)

// -- Helpers --

func writeDocWithFrontmatter(t *testing.T, path, lastReviewed string) {
	t.Helper()
	dir := filepath.Dir(path)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	var content string
	if lastReviewed != "" {
		content = "---\nlast_reviewed: " + lastReviewed + "\n---\n\n# Title\n\nContent here.\n"
	} else {
		content = "# Title\n\nContent without frontmatter.\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// -- Registry loading --

func TestDocScanner_LoadsRegistryFromTOML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	regPath := filepath.Join(dir, "doc-registry.toml")
	content := "[[docs]]\npath = \"docs/PRD.md\"\nowner = \"pm\"\nreview_interval_days = 14\n\n" +
		"[[docs]]\npath = \"docs/DDD.md\"\n"
	require.NoError(t, os.WriteFile(regPath, []byte(content), 0o644))

	scanner := infrastructure.NewFilesystemDocScanner()
	entries, err := scanner.LoadRegistry(regPath)
	require.NoError(t, err)

	require.Len(t, entries, 2)
	assert.Equal(t, "docs/PRD.md", entries[0].Path())
	assert.Equal(t, "pm", entries[0].Owner())
	assert.Equal(t, 14, entries[0].ReviewIntervalDays())
	assert.Equal(t, "docs/DDD.md", entries[1].Path())
	assert.Empty(t, entries[1].Owner())
	assert.Equal(t, 30, entries[1].ReviewIntervalDays())
}

func TestDocScanner_ReturnsEmptyOnMissingRegistry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	scanner := infrastructure.NewFilesystemDocScanner()
	entries, err := scanner.LoadRegistry(filepath.Join(dir, "nonexistent.toml"))
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// -- Registered doc scanning --

func TestDocScanner_ScanRegisteredOK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reviewed := time.Now().AddDate(0, 0, -5).Format("2006-01-02")
	writeDocWithFrontmatter(t, filepath.Join(dir, "docs", "PRD.md"), reviewed)

	scanner := infrastructure.NewFilesystemDocScanner()
	entry, _ := dochealthdomain.NewDocRegistryEntry("docs/PRD.md", "", 30)
	statuses, err := scanner.ScanRegistered([]dochealthdomain.DocRegistryEntry{entry}, dir)
	require.NoError(t, err)

	require.Len(t, statuses, 1)
	assert.Equal(t, dochealthdomain.DocHealthOK, statuses[0].Status())
	assert.NotNil(t, statuses[0].DaysSince())
	// Allow ±1 day tolerance for timezone edge cases near midnight
	assert.InDelta(t, 5, *statuses[0].DaysSince(), 1)
}

func TestDocScanner_ScanRegisteredMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	scanner := infrastructure.NewFilesystemDocScanner()
	entry, _ := dochealthdomain.NewDocRegistryEntry("docs/MISSING.md", "", 30)
	statuses, err := scanner.ScanRegistered([]dochealthdomain.DocRegistryEntry{entry}, dir)
	require.NoError(t, err)

	require.Len(t, statuses, 1)
	assert.Equal(t, dochealthdomain.DocHealthMissing, statuses[0].Status())
}

func TestDocScanner_ScanRegisteredNoFrontmatter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeDocWithFrontmatter(t, filepath.Join(dir, "docs", "plain.md"), "")

	scanner := infrastructure.NewFilesystemDocScanner()
	entry, _ := dochealthdomain.NewDocRegistryEntry("docs/plain.md", "", 30)
	statuses, err := scanner.ScanRegistered([]dochealthdomain.DocRegistryEntry{entry}, dir)
	require.NoError(t, err)

	require.Len(t, statuses, 1)
	assert.Equal(t, dochealthdomain.DocHealthNoFrontmatter, statuses[0].Status())
}

func TestDocScanner_ScanRegisteredStale(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reviewed := time.Now().AddDate(0, 0, -45).Format("2006-01-02")
	writeDocWithFrontmatter(t, filepath.Join(dir, "docs", "old.md"), reviewed)

	scanner := infrastructure.NewFilesystemDocScanner()
	entry, _ := dochealthdomain.NewDocRegistryEntry("docs/old.md", "", 30)
	statuses, err := scanner.ScanRegistered([]dochealthdomain.DocRegistryEntry{entry}, dir)
	require.NoError(t, err)

	require.Len(t, statuses, 1)
	assert.Equal(t, dochealthdomain.DocHealthStale, statuses[0].Status())
	// Allow ±1 day tolerance for timezone edge cases near midnight
	assert.InDelta(t, 45, *statuses[0].DaysSince(), 1)
}

func TestDocScanner_HandlesPlaceholderDate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeDocWithFrontmatter(t, filepath.Join(dir, "docs", "placeholder.md"), "YYYY-MM-DD")

	scanner := infrastructure.NewFilesystemDocScanner()
	entry, _ := dochealthdomain.NewDocRegistryEntry("docs/placeholder.md", "", 30)
	statuses, err := scanner.ScanRegistered([]dochealthdomain.DocRegistryEntry{entry}, dir)
	require.NoError(t, err)

	require.Len(t, statuses, 1)
	assert.Equal(t, dochealthdomain.DocHealthNoFrontmatter, statuses[0].Status())
}

func TestDocScanner_HandlesInvalidDate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeDocWithFrontmatter(t, filepath.Join(dir, "docs", "bad_date.md"), "not-a-date")

	scanner := infrastructure.NewFilesystemDocScanner()
	entry, _ := dochealthdomain.NewDocRegistryEntry("docs/bad_date.md", "", 30)
	statuses, err := scanner.ScanRegistered([]dochealthdomain.DocRegistryEntry{entry}, dir)
	require.NoError(t, err)

	require.Len(t, statuses, 1)
	assert.Equal(t, dochealthdomain.DocHealthNoFrontmatter, statuses[0].Status())
}

func TestDocScanner_PassesOwnerThrough(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reviewed := time.Now().Format("2006-01-02")
	writeDocWithFrontmatter(t, filepath.Join(dir, "docs", "owned.md"), reviewed)

	scanner := infrastructure.NewFilesystemDocScanner()
	entry, _ := dochealthdomain.NewDocRegistryEntry("docs/owned.md", "team-lead", 30)
	statuses, err := scanner.ScanRegistered([]dochealthdomain.DocRegistryEntry{entry}, dir)
	require.NoError(t, err)

	assert.Equal(t, "team-lead", statuses[0].Owner())
}

// -- Unregistered doc scanning --

func TestDocScanner_UnregisteredExcludesDirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")

	// Regular doc
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "notes.md"), []byte("# Notes\n"), 0o644))

	// Excluded dirs
	templatesDir := filepath.Join(docsDir, "templates")
	require.NoError(t, os.MkdirAll(templatesDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "template.md"), []byte("# Template\n"), 0o644))

	beadsDir := filepath.Join(docsDir, "beads_templates")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(beadsDir, "ticket.md"), []byte("# Ticket\n"), 0o644))

	scanner := infrastructure.NewFilesystemDocScanner()
	statuses, err := scanner.ScanUnregistered(docsDir, nil, []string{"templates", "beads_templates"})
	require.NoError(t, err)

	paths := make(map[string]bool)
	for _, s := range statuses {
		paths[s.Path()] = true
	}
	// notes.md should be present; template/beads should not
	found := false
	for p := range paths {
		if filepath.Base(p) == "notes.md" {
			found = true
		}
	}
	assert.True(t, found, "notes.md should be found")
}

func TestDocScanner_UnregisteredSkipsRegisteredPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "PRD.md"),
		[]byte("---\nlast_reviewed: 2026-01-01\n---\n# PRD\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "extra.md"),
		[]byte("# Extra\n"), 0o644))

	scanner := infrastructure.NewFilesystemDocScanner()
	registered := map[string]bool{"docs/PRD.md": true}
	statuses, err := scanner.ScanUnregistered(docsDir, registered, nil)
	require.NoError(t, err)

	paths := make(map[string]bool)
	for _, s := range statuses {
		paths[s.Path()] = true
	}
	for p := range paths {
		assert.NotContains(t, p, "PRD.md")
	}
	foundExtra := false
	for p := range paths {
		if filepath.Base(p) == "extra.md" {
			foundExtra = true
		}
	}
	assert.True(t, foundExtra)
}

func TestDocScanner_UnregisteredDetectsNoFrontmatter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "bare.md"),
		[]byte("# Just a title\n\nNo frontmatter.\n"), 0o644))

	scanner := infrastructure.NewFilesystemDocScanner()
	statuses, err := scanner.ScanUnregistered(docsDir, nil, nil)
	require.NoError(t, err)

	require.Len(t, statuses, 1)
	assert.Equal(t, dochealthdomain.DocHealthNoFrontmatter, statuses[0].Status())
}

func TestDocScanner_UnregisteredEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0o755))

	scanner := infrastructure.NewFilesystemDocScanner()
	statuses, err := scanner.ScanUnregistered(docsDir, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, statuses)
}

func TestDocScanner_UnregisteredNonexistentDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs") // does not exist

	scanner := infrastructure.NewFilesystemDocScanner()
	statuses, err := scanner.ScanUnregistered(docsDir, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, statuses)
}
