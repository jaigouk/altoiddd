package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rescuedomain "github.com/alto-cli/alto/internal/rescue/domain"
)

// ===========================================================================
// Scenario: alto gap reports structural gaps
// ===========================================================================

func TestGapEmptyProject_GivenEmptyDirectory_WhenAnalyzeGaps_ThenReportsAllRequiredGaps(t *testing.T) {
	t.Parallel()

	// Given: an empty directory (no docs, no configs, no .alto)
	app := newApp(t)
	dir := t.TempDir()

	// When: AnalyzeGaps
	report, err := app.GapQueryHandler.AnalyzeGaps(context.Background(), dir)
	require.NoError(t, err)

	// Then: reports required gaps for docs and configs
	assert.NotEmpty(t, report.Entries, "empty directory should have gaps")

	requiredPaths := map[string]bool{
		"docs/PRD.md":          false,
		"docs/DDD.md":          false,
		"docs/ARCHITECTURE.md": false,
		".claude/CLAUDE.md":    false,
	}

	for _, e := range report.Entries {
		if e.Severity == string(rescuedomain.GapSeverityRequired) {
			if _, ok := requiredPaths[e.Path]; ok {
				requiredPaths[e.Path] = true
			}
		}
	}

	for path, found := range requiredPaths {
		assert.True(t, found, "expected required gap for %s", path)
	}
}

func TestGapCompleteProject_GivenFullyStructuredProject_WhenAnalyzeGaps_ThenReportsNoGaps(t *testing.T) {
	t.Parallel()

	// Given: a fully structured project with all required and recommended files
	app := newApp(t)
	dir := t.TempDir()

	// Create required docs
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("# PRD\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "DDD.md"), []byte("# DDD\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "ARCHITECTURE.md"), []byte("# Arch\n"), 0o644))

	// Create required config
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "CLAUDE.md"), []byte("# Claude\n"), 0o644))

	// Create recommended .alto structure
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alto", "knowledge"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alto", "maintenance"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".alto", "config.toml"), []byte("project_name = \"test\"\n"), 0o644))

	// Create AGENTS.md
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Agents\n"), 0o644))

	// When: AnalyzeGaps
	report, err := app.GapQueryHandler.AnalyzeGaps(context.Background(), dir)
	require.NoError(t, err)

	// Then: no gaps
	assert.Empty(t, report.Entries, "fully structured project should have no gaps")
}

func TestGapPartialProject_GivenProjectMissingAlto_WhenAnalyzeGaps_ThenReportsAltoGapsAsRecommended(t *testing.T) {
	t.Parallel()

	// Given: a project with docs and configs but no .alto/ directory
	app := newApp(t)
	dir := t.TempDir()

	// Create required docs
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("# PRD\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "DDD.md"), []byte("# DDD\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "ARCHITECTURE.md"), []byte("# Arch\n"), 0o644))

	// Create required config
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "CLAUDE.md"), []byte("# Claude\n"), 0o644))

	// Create AGENTS.md
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Agents\n"), 0o644))

	// When: AnalyzeGaps
	report, err := app.GapQueryHandler.AnalyzeGaps(context.Background(), dir)
	require.NoError(t, err)

	// Then: .alto gaps should be recommended, not required
	altoGaps := map[string]string{}
	for _, e := range report.Entries {
		if e.Path == ".alto/config.toml" ||
			e.Path == ".alto/knowledge/" ||
			e.Path == ".alto/maintenance/" {
			altoGaps[e.Path] = e.Severity
		}
	}

	assert.NotEmpty(t, altoGaps, "should detect missing .alto/ structure")
	for path, sev := range altoGaps {
		assert.Equal(t, string(rescuedomain.GapSeverityRecommended), sev,
			".alto gap %s should be recommended severity", path)
	}

	// Should have no required gaps (docs and configs are present)
	assert.False(t, report.HasRequired, "should have no required gaps when docs/configs present")
}

func TestGapEmptyProject_GivenEmptyDir_WhenAnalyzeGaps_ThenReportsRecommendedAltoGaps(t *testing.T) {
	t.Parallel()

	// Given: an empty directory
	app := newApp(t)
	dir := t.TempDir()

	// When: AnalyzeGaps
	report, err := app.GapQueryHandler.AnalyzeGaps(context.Background(), dir)
	require.NoError(t, err)

	// Then: should include recommended .alto gaps
	recommendedPaths := map[string]bool{
		".alto/config.toml":  false,
		".alto/knowledge/":   false,
		".alto/maintenance/": false,
	}

	for _, e := range report.Entries {
		if e.Severity == string(rescuedomain.GapSeverityRecommended) {
			if _, ok := recommendedPaths[e.Path]; ok {
				recommendedPaths[e.Path] = true
			}
		}
	}

	for path, found := range recommendedPaths {
		assert.True(t, found, "expected recommended gap for %s", path)
	}
}
