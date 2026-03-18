package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/knowledge/domain"
	"github.com/alto-cli/alto/internal/knowledge/infrastructure"
)

// ---------------------------------------------------------------------------
// Tests — DriftDetectionAdapter
// ---------------------------------------------------------------------------

func TestDriftDetectionAdapter_Detect(t *testing.T) {
	t.Parallel()

	t.Run("returns empty report when no knowledge dir", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 0, report.TotalCount())
	})

	t.Run("returns empty report when no meta files", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", "test-tool")
		require.NoError(t, os.MkdirAll(kbDir, 0o755))

		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 0, report.TotalCount())
	})

	t.Run("detects stale entry from last_verified", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", "claude-code")
		require.NoError(t, os.MkdirAll(kbDir, 0o755))

		// Create _meta.toml with old last_verified date
		metaContent := `[tool]
name = "claude-code"

[versions.v2_1]
last_verified = "2025-01-01"
`
		require.NoError(t, os.WriteFile(filepath.Join(kbDir, "_meta.toml"), []byte(metaContent), 0o644))

		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err)
		assert.GreaterOrEqual(t, report.TotalCount(), 1)
		assert.Equal(t, 1, report.CountByType(domain.DriftStale))
	})

	t.Run("no stale for recently verified entry", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", "claude-code")
		require.NoError(t, os.MkdirAll(kbDir, 0o755))

		// Create _meta.toml with recent last_verified date (within 14-day threshold)
		recentDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		metaContent := `[tool]
name = "claude-code"

[versions.v2_1]
last_verified = "` + recentDate + `"
`
		require.NoError(t, os.WriteFile(filepath.Join(kbDir, "_meta.toml"), []byte(metaContent), 0o644))

		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 0, report.CountByType(domain.DriftStale))
	})

	t.Run("treats missing last_verified as stale", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", "test-tool")
		require.NoError(t, os.MkdirAll(kbDir, 0o755))

		// Create _meta.toml without last_verified
		metaContent := `[tool]
name = "test-tool"

[versions.v1_0]
version_range = ">=1.0.0"
`
		require.NoError(t, os.WriteFile(filepath.Join(kbDir, "_meta.toml"), []byte(metaContent), 0o644))

		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 1, report.CountByType(domain.DriftStale))
	})

	t.Run("skips malformed toml gracefully", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", "bad-tool")
		require.NoError(t, os.MkdirAll(kbDir, 0o755))

		// Create malformed _meta.toml
		require.NoError(t, os.WriteFile(filepath.Join(kbDir, "_meta.toml"), []byte("not valid toml [[["), 0o644))

		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err) // Should not error, just skip
		assert.Equal(t, 0, report.TotalCount())
	})

	t.Run("scans multiple tools", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create two tool directories with stale entries
		for _, tool := range []string{"tool-a", "tool-b"} {
			kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", tool)
			require.NoError(t, os.MkdirAll(kbDir, 0o755))

			metaContent := `[tool]
name = "` + tool + `"

[versions.v1_0]
last_verified = "2024-01-01"
`
			require.NoError(t, os.WriteFile(filepath.Join(kbDir, "_meta.toml"), []byte(metaContent), 0o644))
		}

		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 2, report.CountByType(domain.DriftStale))
	})

	t.Run("entry path contains tool name", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", "cursor")
		require.NoError(t, os.MkdirAll(kbDir, 0o755))

		metaContent := `[tool]
name = "cursor"

[versions.v1_0]
last_verified = "2024-01-01"
`
		require.NoError(t, os.WriteFile(filepath.Join(kbDir, "_meta.toml"), []byte(metaContent), 0o644))

		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		report, err := adapter.Detect(context.Background())

		require.NoError(t, err)
		require.Equal(t, 1, report.TotalCount())
		assert.Contains(t, report.Signals()[0].EntryPath(), "tools/cursor")
	})

	t.Run("context cancellation returns error", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := adapter.Detect(ctx)

		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("custom threshold changes staleness detection", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, ".alto", "knowledge", "tools", "test-tool")
		require.NoError(t, os.MkdirAll(kbDir, 0o755))

		// Create _meta.toml with 20-day-old date
		oldDate := time.Now().AddDate(0, 0, -20).Format("2006-01-02")
		metaContent := `[tool]
name = "test-tool"

[versions.v1_0]
last_verified = "` + oldDate + `"
`
		require.NoError(t, os.WriteFile(filepath.Join(kbDir, "_meta.toml"), []byte(metaContent), 0o644))

		// Default threshold (14 days) should mark as stale
		adapter := infrastructure.NewDriftDetectionAdapter(tmpDir)
		report, err := adapter.Detect(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, report.CountByType(domain.DriftStale), "default 14-day threshold should detect stale")

		// Custom threshold (30 days) should NOT mark as stale
		adapterLenient := infrastructure.NewDriftDetectionAdapter(tmpDir).WithStaleThreshold(30)
		reportLenient, err := adapterLenient.Detect(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, reportLenient.CountByType(domain.DriftStale), "30-day threshold should not detect stale")
	})
}
