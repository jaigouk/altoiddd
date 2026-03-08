package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
)

// ---------------------------------------------------------------------------
// Fake tool detection
// ---------------------------------------------------------------------------

type fakeToolDetect struct {
	tools        []string
	conflicts    []string
	receivedDirs []string
}

func (f *fakeToolDetect) Detect(projectDir string) ([]string, error) {
	f.receivedDirs = append(f.receivedDirs, projectDir)
	return f.tools, nil
}

func (f *fakeToolDetect) ScanConflicts(projectDir string) ([]string, error) {
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
		fake := &fakeToolDetect{tools: []string{}, conflicts: []string{}}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.Equal(t, 0, len(result.DetectedTools()))
		assert.Equal(t, 0, len(result.Conflicts()))
	})

	t.Run("with tools", func(t *testing.T) {
		t.Parallel()
		fake := &fakeToolDetect{tools: []string{"claude-code", "cursor"}}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.Equal(t, 2, len(result.DetectedTools()))
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
			tools:     []string{"cursor"},
			conflicts: []string{"cursor: SQLite-based config detected, cannot read"},
		}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		assert.Equal(t, 1, len(result.Conflicts()))
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
			conflicts: []string{
				"cursor: SQLite-based config detected, cannot read",
				"claude-code: global setting 'model' contradicts local value",
			},
		}
		handler := application.NewDetectionHandler(fake)

		result, err := handler.Detect("/tmp/proj")

		require.NoError(t, err)
		severityMap := result.SeverityMap()
		severities := make(map[domain.ConflictSeverity]bool)
		for _, sev := range severityMap {
			severities[sev] = true
		}
		assert.True(t, severities[domain.SeverityWarning])
		assert.True(t, severities[domain.SeverityConflict])
	})
}
