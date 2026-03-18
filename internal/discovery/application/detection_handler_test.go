package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
)

// ---------------------------------------------------------------------------
// Fake tool detection
// ---------------------------------------------------------------------------

type fakeToolDetect struct {
	tools        []string
	conflicts    []domain.SettingsConflict
	receivedDirs []string
}

func (f *fakeToolDetect) Detect(projectDir string) ([]string, error) {
	f.receivedDirs = append(f.receivedDirs, projectDir)
	return f.tools, nil
}

func (f *fakeToolDetect) ScanConflicts(projectDir string) ([]domain.SettingsConflict, error) {
	f.receivedDirs = append(f.receivedDirs, projectDir)
	return f.conflicts, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestDetectionHandler(t *testing.T) {
	t.Parallel()

	t.Run("returns detection result", func(t *testing.T) {
		t.Parallel()
		fake := &fakeToolDetect{tools: []string{"claude-code"}}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("no tools", func(t *testing.T) {
		t.Parallel()
		fake := &fakeToolDetect{tools: []string{}, conflicts: []domain.SettingsConflict{}}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.Empty(t, result.DetectedTools())
		assert.Empty(t, result.Conflicts())
	})

	t.Run("with tools", func(t *testing.T) {
		t.Parallel()
		fake := &fakeToolDetect{tools: []string{"claude-code", "cursor"}}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.Len(t, result.DetectedTools(), 2)
		names := make([]string, 0)
		for _, dt := range result.DetectedTools() {
			names = append(names, dt.Name())
		}
		assert.Contains(t, names, "claude-code")
		assert.Contains(t, names, "cursor")
	})

	t.Run("with conflicts", func(t *testing.T) {
		t.Parallel()
		fake := &fakeToolDetect{
			tools: []string{"cursor"},
			conflicts: []domain.SettingsConflict{
				domain.NewSettingsConflict("cursor", "/home/.cursor", "", "global_only", domain.SettingsSeverityWarning, "SQLite-based config detected, cannot read"),
			},
		}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.Len(t, result.Conflicts(), 1)
		severityMap := result.SeverityMap()
		hasWarning := false
		for _, sev := range severityMap {
			if sev == domain.SeverityWarning {
				hasWarning = true
			}
		}
		assert.True(t, hasWarning)
	})

	t.Run("passes project dir to port", func(t *testing.T) {
		t.Parallel()
		fake := &fakeToolDetect{tools: []string{}}
		handler := application.NewDetectionHandler(fake)

		_, err := handler.Detect("/tmp/myproject")

		require.NoError(t, err)
		assert.Contains(t, fake.receivedDirs, "/tmp/myproject")
	})

	t.Run("with multiple severity levels", func(t *testing.T) {
		t.Parallel()
		fake := &fakeToolDetect{
			tools: []string{"claude-code", "cursor"},
			conflicts: []domain.SettingsConflict{
				domain.NewSettingsConflict("cursor", "/home/.cursor", "", "global_only", domain.SettingsSeverityWarning, "SQLite-based config detected, cannot read"),
				domain.NewSettingsConflict("claude-code", "/home/.claude/settings.json", "/proj/.claude/settings.json", "content_mismatch", domain.SettingsSeverityWarning, "global setting 'model' contradicts local value"),
			},
		}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.Len(t, result.Conflicts(), 2)
		severityMap := result.SeverityMap()
		assert.NotEmpty(t, severityMap)
	})
}
