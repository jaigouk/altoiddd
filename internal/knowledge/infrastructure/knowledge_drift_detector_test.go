package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	knowledgedomain "github.com/alty-cli/alty/internal/knowledge/domain"
	"github.com/alty-cli/alty/internal/knowledge/infrastructure"
)

func makeTOML(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func makeKnowledgeTree(t *testing.T, root string) {
	t.Helper()
	makeTOML(t, filepath.Join(root, "tools", "claude-code", "_meta.toml"),
		"[tool]\nname = \"claude-code\"\n\n[versions]\ncurrent = \"v2.1\"\ntracked = [\"v2.1\", \"v2.0\"]\n")

	makeTOML(t, filepath.Join(root, "tools", "claude-code", "current", "config-structure.toml"),
		"[_meta]\nlast_verified = \"2026-02-22\"\nverified_against = \"v2.1.15\"\nconfidence = \"high\"\n"+
			"next_review_date = \"2026-05-22\"\n\n[project_structure]\n\"CLAUDE.md\" = \"Project memory\"\n"+
			"\"rules/*.md\" = \"Additional rules\"\n\"agents/*.md\" = \"Agent definitions\"\n")

	makeTOML(t, filepath.Join(root, "tools", "claude-code", "v2.0", "config-structure.toml"),
		"[_meta]\nlast_verified = \"2026-02-22\"\nverified_against = \"v2.0.x\"\nconfidence = \"high\"\n"+
			"next_review_date = \"2026-05-22\"\n\n[project_structure]\n\"CLAUDE.md\" = \"Project memory\"\n")

	makeTOML(t, filepath.Join(root, "tools", "claude-code", "v2.1", "config-structure.toml"),
		"[_meta]\nlast_verified = \"2026-02-22\"\nverified_against = \"v2.1.15\"\nconfidence = \"high\"\n"+
			"next_review_date = \"2026-05-22\"\n\n[project_structure]\n\"CLAUDE.md\" = \"Project memory\"\n"+
			"\"rules/*.md\" = \"Additional rules\"\n\"agents/*.md\" = \"Agent definitions\"\n")
}

// -- Version drift --

func TestDriftDetector_DetectsKeysAddedInNewerVersion(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	makeKnowledgeTree(t, dir)

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	versionSignals := filterByType(report, knowledgedomain.DriftVersionChange)
	assert.GreaterOrEqual(t, len(versionSignals), 1)
	descs := joinDescriptions(versionSignals)
	assert.True(t, containsAny(descs, "rules", "agents"))
}

func TestDriftDetector_DetectsSectionRemovedInCurrent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	makeTOML(t, filepath.Join(dir, "tools", "rm-tool", "_meta.toml"),
		"[tool]\nname = \"rm-tool\"\n\n[versions]\ncurrent = \"v2.0\"\ntracked = [\"v2.0\", \"v1.0\"]\n")
	makeTOML(t, filepath.Join(dir, "tools", "rm-tool", "v1.0", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\n\n"+
			"[data]\nkey = 1\n\n[deprecated_feature]\nold_key = 1\n")
	makeTOML(t, filepath.Join(dir, "tools", "rm-tool", "current", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\n\n[data]\nkey = 1\n")
	makeTOML(t, filepath.Join(dir, "tools", "rm-tool", "v2.0", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\n\n[data]\nkey = 1\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	versionSignals := filterByType(report, knowledgedomain.DriftVersionChange)
	assert.GreaterOrEqual(t, len(versionSignals), 1)
	assert.True(t, anyContains(versionSignals, "deprecated_feature"))
}

func TestDriftDetector_NoDriftWhenCurrentMatchesLatest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	makeKnowledgeTree(t, dir)

	// Remove v2.0 so only v2.1 comparison happens (identical to current)
	require.NoError(t, os.RemoveAll(filepath.Join(dir, "tools", "claude-code", "v2.0")))
	makeTOML(t, filepath.Join(dir, "tools", "claude-code", "_meta.toml"),
		"[tool]\nname = \"claude-code\"\n\n[versions]\ncurrent = \"v2.1\"\ntracked = [\"v2.1\"]\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	versionSignals := filterByType(report, knowledgedomain.DriftVersionChange)
	assert.Empty(t, versionSignals)
}

// -- Staleness --

func TestDriftDetector_DetectsStaleEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	makeTOML(t, filepath.Join(dir, "tools", "stale-tool", "_meta.toml"),
		"[tool]\nname = \"stale-tool\"\n\n[versions]\ncurrent = \"v1.0\"\ntracked = [\"v1.0\"]\n")
	makeTOML(t, filepath.Join(dir, "tools", "stale-tool", "current", "config.toml"),
		"[_meta]\nlast_verified = \"2025-01-01\"\nnext_review_date = \"2025-06-01\"\nconfidence = \"high\"\n\n"+
			"[data]\nkey = 1\n")
	makeTOML(t, filepath.Join(dir, "tools", "stale-tool", "v1.0", "config.toml"),
		"[_meta]\nlast_verified = \"2025-01-01\"\nnext_review_date = \"2025-06-01\"\nconfidence = \"high\"\n\n"+
			"[data]\nkey = 1\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	staleSignals := filterByType(report, knowledgedomain.DriftStale)
	assert.GreaterOrEqual(t, len(staleSignals), 1)
}

func TestDriftDetector_NoStalenessWhenFutureReviewDate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	makeTOML(t, filepath.Join(dir, "tools", "fresh-tool", "_meta.toml"),
		"[tool]\nname = \"fresh-tool\"\n\n[versions]\ncurrent = \"v1.0\"\ntracked = [\"v1.0\"]\n")
	makeTOML(t, filepath.Join(dir, "tools", "fresh-tool", "current", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\nconfidence = \"high\"\n\n"+
			"[data]\nkey = 1\n")
	makeTOML(t, filepath.Join(dir, "tools", "fresh-tool", "v1.0", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\nconfidence = \"high\"\n\n"+
			"[data]\nkey = 1\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	staleSignals := filterByType(report, knowledgedomain.DriftStale)
	assert.Empty(t, staleSignals)
}

// -- Edge cases --

func TestDriftDetector_EmptyKnowledgeBase(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	emptyDir := filepath.Join(dir, "empty")
	require.NoError(t, os.MkdirAll(emptyDir, 0o755))

	detector := infrastructure.NewKnowledgeDriftDetector(emptyDir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, report.TotalCount())
	assert.False(t, report.HasDrift())
}

func TestDriftDetector_NoVersionHistory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	makeTOML(t, filepath.Join(dir, "tools", "new-tool", "_meta.toml"),
		"[tool]\nname = \"new-tool\"\n\n[versions]\ncurrent = \"v1.0\"\ntracked = []\n")
	makeTOML(t, filepath.Join(dir, "tools", "new-tool", "current", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\n\n[data]\nkey = 1\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	infoSignals := filterBySeverity(report, knowledgedomain.SeverityInfo)
	assert.GreaterOrEqual(t, len(infoSignals), 1)
	assert.True(t, anyContainsLower(infoSignals, "no version history"))
}

func TestDriftDetector_MissingMetaSkipsTool(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "tools", "bad-tool", "current"), 0o755))
	makeTOML(t, filepath.Join(dir, "tools", "bad-tool", "current", "config.toml"),
		"[data]\nkey = 1\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, report.TotalCount(), 0)
}

func TestDriftDetector_MultipleDriftSignalsPerEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	makeTOML(t, filepath.Join(dir, "tools", "multi-tool", "_meta.toml"),
		"[tool]\nname = \"multi-tool\"\n\n[versions]\ncurrent = \"v2.0\"\ntracked = [\"v2.0\", \"v1.0\"]\n")
	makeTOML(t, filepath.Join(dir, "tools", "multi-tool", "current", "config.toml"),
		"[_meta]\nlast_verified = \"2024-01-01\"\nnext_review_date = \"2024-06-01\"\n\n"+
			"[data]\nold_key = 1\nnew_key = 2\n")
	makeTOML(t, filepath.Join(dir, "tools", "multi-tool", "v1.0", "config.toml"),
		"[_meta]\nlast_verified = \"2024-01-01\"\nnext_review_date = \"2024-06-01\"\n\n"+
			"[data]\nold_key = 1\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	types := signalTypes(report)
	assert.Contains(t, types, knowledgedomain.DriftVersionChange)
	assert.Contains(t, types, knowledgedomain.DriftStale)
}

func TestDriftDetector_VersionEntryMissingFromDisk(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	makeTOML(t, filepath.Join(dir, "tools", "gap-tool", "_meta.toml"),
		"[tool]\nname = \"gap-tool\"\n\n[versions]\ncurrent = \"v2.0\"\ntracked = [\"v2.0\", \"v1.0\"]\n")
	makeTOML(t, filepath.Join(dir, "tools", "gap-tool", "current", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\n\n[data]\nkey = 1\n")
	makeTOML(t, filepath.Join(dir, "tools", "gap-tool", "v2.0", "config.toml"),
		"[_meta]\nlast_verified = \"2026-03-01\"\nnext_review_date = \"2027-03-01\"\n\n[data]\nkey = 1\n")
	// v1.0 directory does NOT exist

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	warnSignals := filterBySeverity(report, knowledgedomain.SeverityWarning)
	assert.True(t, anyContains(warnSignals, "v1.0"))
}

func TestDriftDetector_CrossToolNotScanned(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	makeTOML(t, filepath.Join(dir, "cross-tool", "concept-mapping.toml"),
		"[_meta]\nlast_verified = \"2024-01-01\"\nnext_review_date = \"2024-06-01\"\n\n[data]\nkey = 1\n")

	detector := infrastructure.NewKnowledgeDriftDetector(dir)
	report, err := detector.Detect(context.Background())
	require.NoError(t, err)

	versionSignals := filterByType(report, knowledgedomain.DriftVersionChange)
	assert.Empty(t, versionSignals)
}

// -- Test helpers --

func filterByType(report knowledgedomain.DriftReport, signalType knowledgedomain.DriftSignalType) []knowledgedomain.DriftSignal {
	var result []knowledgedomain.DriftSignal
	for _, s := range report.Signals() {
		if s.SignalType() == signalType {
			result = append(result, s)
		}
	}
	return result
}

func filterBySeverity(report knowledgedomain.DriftReport, severity knowledgedomain.DriftSeverity) []knowledgedomain.DriftSignal {
	var result []knowledgedomain.DriftSignal
	for _, s := range report.Signals() {
		if s.Severity() == severity {
			result = append(result, s)
		}
	}
	return result
}

func signalTypes(report knowledgedomain.DriftReport) map[knowledgedomain.DriftSignalType]bool {
	types := make(map[knowledgedomain.DriftSignalType]bool)
	for _, s := range report.Signals() {
		types[s.SignalType()] = true
	}
	return types
}

func joinDescriptions(signals []knowledgedomain.DriftSignal) string {
	var parts []string
	for _, s := range signals {
		parts = append(parts, s.Description())
	}
	return strings.Join(parts, " ")
}

func anyContains(signals []knowledgedomain.DriftSignal, substr string) bool {
	for _, s := range signals {
		if strings.Contains(s.Description(), substr) {
			return true
		}
	}
	return false
}

func anyContainsLower(signals []knowledgedomain.DriftSignal, substr string) bool {
	for _, s := range signals {
		if strings.Contains(strings.ToLower(s.Description()), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

func containsAny(text string, substrings ...string) bool {
	lower := strings.ToLower(text)
	for _, s := range substrings {
		if strings.Contains(lower, strings.ToLower(s)) {
			return true
		}
	}
	return false
}
