package infrastructure_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/research/application"
	"github.com/alto-cli/alto/internal/research/infrastructure"
)

// Compile-time interface check.
var _ application.SpikeFollowUp = (*infrastructure.SpikeFollowUpAdapter)(nil)

// -- Helpers --

func writeSpikeReport(t *testing.T, projectDir, filename, content string) string {
	t.Helper()
	researchDir := filepath.Join(projectDir, "docs", "research")
	require.NoError(t, os.MkdirAll(researchDir, 0o755))
	path := filepath.Join(researchDir, filename)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func writeBeadsIssues(t *testing.T, projectDir string, titles []string) {
	t.Helper()
	beadsDir := filepath.Join(projectDir, ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0o755))
	issuesPath := filepath.Join(beadsDir, "issues.jsonl")

	var lines []byte
	for i, title := range titles {
		issue := map[string]string{
			"id":     fmt.Sprintf("alto-t%d", i),
			"title":  title,
			"status": "open",
		}
		data, err := json.Marshal(issue)
		require.NoError(t, err)
		lines = append(lines, data...)
		lines = append(lines, '\n')
	}
	require.NoError(t, os.WriteFile(issuesPath, lines, 0o644))
}

// -- Happy path: audit --

func TestFollowUp_DetectsOrphanedFollowUps(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_test_spike.md",
		"# Test Spike\n\n## Follow-Up Tickets\n\n"+
			"### Ticket 1: Build the widget\n\nDetails.\n\n"+
			"### Ticket 2: Test the widget\n\nDetails.\n")
	writeBeadsIssues(t, dir, []string{}) // No tickets

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "test-spike", dir)
	require.NoError(t, err)
	assert.Equal(t, 2, result.DefinedCount())
	assert.Equal(t, 2, result.OrphanedCount())
	assert.True(t, result.HasOrphans())
}

func TestFollowUp_AllMatchedNoOrphans(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_test_spike.md",
		"# Spike\n\n## Follow-Up Tickets\n\n"+
			"### Ticket 1: Build the widget\n\nDetails.\n\n"+
			"### Ticket 2: Test the widget\n\nDetails.\n")
	writeBeadsIssues(t, dir, []string{"Build the widget", "Test the widget"})

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "test-spike", dir)
	require.NoError(t, err)
	assert.Equal(t, 2, result.DefinedCount())
	assert.Equal(t, 0, result.OrphanedCount())
	assert.False(t, result.HasOrphans())
	assert.Len(t, result.MatchedTicketIDs(), 2)
}

func TestFollowUp_PartialMatchReportsOrphans(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_test_spike.md",
		"# Spike\n\n## Follow-Up Tickets\n\n"+
			"### Ticket 1: Build parser\n\nD.\n\n"+
			"### Ticket 2: Build adapter\n\nD.\n\n"+
			"### Ticket 3: Wire composition\n\nD.\n")
	writeBeadsIssues(t, dir, []string{"Build parser"})

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "test-spike", dir)
	require.NoError(t, err)
	assert.Equal(t, 3, result.DefinedCount())
	assert.Equal(t, 2, result.OrphanedCount())
	assert.Len(t, result.MatchedTicketIDs(), 1)
}

func TestFollowUp_FuzzyMatchCaseInsensitive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_test_spike.md",
		"# Spike\n\n## Follow-Up Tickets\n\n### Ticket 1: Build The Widget\n\nD.\n")
	writeBeadsIssues(t, dir, []string{"build the widget"})

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "test-spike", dir)
	require.NoError(t, err)
	assert.Equal(t, 0, result.OrphanedCount())
}

func TestFollowUp_FuzzyMatchSubstring(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_test_spike.md",
		"# Spike\n\n## Follow-Up Tickets\n\n### Ticket 1: Implement SessionStore\n\nD.\n")
	writeBeadsIssues(t, dir, []string{"Implement SessionStore and MCP discovery tools"})

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "test-spike", dir)
	require.NoError(t, err)
	assert.Equal(t, 0, result.OrphanedCount())
}

func TestFollowUp_ReportPathInResult(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_gap_analysis.md",
		"# Spike\n\n## Follow-Up Tickets\n\n### Ticket 1: Task\n\nD.\n")
	writeBeadsIssues(t, dir, []string{})

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "gap-spike", dir)
	require.NoError(t, err)
	assert.Contains(t, result.ReportPath(), "20260223_gap_analysis.md")
}

// -- Edge cases --

func TestFollowUp_NoResearchDirReturnsEmptyResult(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "missing-spike", dir)
	require.NoError(t, err)
	assert.Equal(t, 0, result.DefinedCount())
	assert.Equal(t, 0, result.OrphanedCount())
}

func TestFollowUp_NoBeadsDirTreatsAllAsOrphaned(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_test.md",
		"# S\n\n## Follow-Up Tickets\n\n### Ticket 1: Task A\n\nD.\n")
	// No .beads/ directory

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "test", dir)
	require.NoError(t, err)
	assert.Equal(t, 1, result.OrphanedCount())
}

func TestFollowUp_SpikeWithNoFollowUpsCleanResult(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_no_followups.md",
		"# Research\n\n## Findings\n\nJust research, no tickets.\n")
	writeBeadsIssues(t, dir, []string{})

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "no-followups", dir)
	require.NoError(t, err)
	assert.Equal(t, 0, result.DefinedCount())
	assert.False(t, result.HasOrphans())
}

func TestFollowUp_MultipleReportsScansAll(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSpikeReport(t, dir, "20260223_report_a.md",
		"# A\n\n## Follow-Up Tickets\n\n### Ticket 1: Task from A\n\nD.\n")
	writeSpikeReport(t, dir, "20260223_report_b.md",
		"# B\n\n## Follow-Up Tickets\n\n### Ticket 1: Task from B\n\nD.\n")
	writeBeadsIssues(t, dir, []string{})

	adapter := infrastructure.NewSpikeFollowUpAdapter()
	result, err := adapter.Audit(context.Background(), "multi", dir)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.DefinedCount(), 1)
}

// -- Fuzzy match: prefix stripping --

func TestFuzzyMatch_TaskPrefixStripped(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch("Task: Build the parser", map[string]string{"t1": "Build the parser"})
	assert.Equal(t, "t1", result)
}

func TestFuzzyMatch_SpikePrefixStripped(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch("Spike: Research caching strategies",
		map[string]string{"t1": "Research caching strategies"})
	assert.Equal(t, "t1", result)
}

func TestFuzzyMatch_OptionalPrefixStripped(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch("(Optional) Add retry logic",
		map[string]string{"t1": "Add retry logic"})
	assert.Equal(t, "t1", result)
}

func TestFuzzyMatch_ImplementAltoPrefixStripped(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch(
		"Implement fitness function generation",
		map[string]string{"t1": "Implement alto generate fitness"})
	assert.Equal(t, "t1", result)
}

func TestFuzzyMatch_EmptyAfterStripNoMatch(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch("Task:", map[string]string{"t1": "Implement alto detect"})
	assert.Empty(t, result)
}

// -- Fuzzy match: keyword overlap --

func TestFuzzyMatch_KeywordOverlapAboveThreshold(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch(
		"Implement fitness function generation (import-linter + pytestarch)",
		map[string]string{"t1": "Implement alto generate fitness (import-linter + pytestarch)"})
	assert.Equal(t, "t1", result)
}

func TestFuzzyMatch_KeywordOverlapBelowThresholdNoMatch(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch(
		"Build the parser for YAML",
		map[string]string{"t1": "Deploy the server to production"})
	assert.Empty(t, result)
}

func TestFuzzyMatch_KeywordReorderMatches(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch(
		"fitness function generation",
		map[string]string{"t1": "generate fitness functions"})
	assert.Equal(t, "t1", result)
}

func TestFuzzyMatch_ShortTitlesSharedCommonWordNoFalsePositive(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch(
		"Implement parser",
		map[string]string{"t1": "Implement deploy pipeline"})
	assert.Empty(t, result)
}

// -- Fuzzy match: parenthetical --

func TestFuzzyMatch_ParentheticalToolsMatch(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch(
		"Task: Implement fitness generation (import-linter + pytestarch)",
		map[string]string{"t1": "Implement alto generate fitness (import-linter + pytestarch)"})
	assert.Equal(t, "t1", result)
}

func TestFuzzyMatch_NoParentheticalStillUsesKeywords(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch(
		"Implement drift detection for knowledge base",
		map[string]string{"t1": "Implement knowledge base drift detection"})
	assert.Equal(t, "t1", result)
}

// -- Fuzzy match: empty existing --

func TestFuzzyMatch_EmptyExistingReturnsEmpty(t *testing.T) {
	t.Parallel()
	result := infrastructure.FuzzyMatch("Build something", map[string]string{})
	assert.Empty(t, result)
}
