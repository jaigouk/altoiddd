package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolScannerBuildResultEmpty(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	result := scanner.BuildResult(nil, nil)
	assert.Empty(t, result.DetectedTools())
	assert.Empty(t, result.Conflicts())
	assert.Empty(t, result.SeverityMap())
}

func TestToolScannerBuildResultWithKnownTools(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	result := scanner.BuildResult([]string{"claude-code", "cursor"}, nil)
	assert.Len(t, result.DetectedTools(), 2)
	toolsByName := make(map[string]DetectedTool)
	for _, dt := range result.DetectedTools() {
		toolsByName[dt.Name()] = dt
	}
	assert.Contains(t, toolsByName, "claude-code")
	assert.Contains(t, toolsByName, "cursor")
}

func TestToolScannerMapsKnownConfigPaths(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	result := scanner.BuildResult([]string{"claude-code"}, nil)
	require.Len(t, result.DetectedTools(), 1)
	tool := result.DetectedTools()[0]
	assert.Equal(t, "claude-code", tool.Name())
	assert.NotEmpty(t, tool.ConfigPath())
	assert.Contains(t, tool.ConfigPath(), ".claude")
}

func TestToolScannerUnknownTool(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	result := scanner.BuildResult([]string{"unknown-tool"}, nil)
	require.Len(t, result.DetectedTools(), 1)
	assert.Equal(t, "unknown-tool", result.DetectedTools()[0].Name())
	assert.Empty(t, result.DetectedTools()[0].ConfigPath())
}

func TestToolScannerClassifyNoConflicts(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	result := scanner.BuildResult([]string{"claude-code"}, nil)
	assert.Empty(t, result.SeverityMap())
}

func TestToolScannerClassifySQLiteAsWarning(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	conflict := "cursor: SQLite-based config detected, cannot read"
	result := scanner.BuildResult([]string{"cursor"}, []string{conflict})
	assert.Equal(t, SeverityWarning, result.SeverityMap()[conflict])
}

func TestToolScannerClassifyContradictionAsConflict(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	conflict := "claude-code: global setting 'model' contradicts local value"
	result := scanner.BuildResult([]string{"claude-code"}, []string{conflict})
	assert.Equal(t, SeverityConflict, result.SeverityMap()[conflict])
}

func TestToolScannerClassifyCompatibleSetting(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	conflict := "claude-code: global setting 'theme' is compatible with local"
	result := scanner.BuildResult([]string{"claude-code"}, []string{conflict})
	assert.Equal(t, SeverityCompatible, result.SeverityMap()[conflict])
}

func TestToolScannerClassifyRestrictionAsWarning(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	conflict := "opencode: global restriction on model access"
	result := scanner.BuildResult([]string{"opencode"}, []string{conflict})
	assert.Equal(t, SeverityWarning, result.SeverityMap()[conflict])
}

func TestToolScannerClassifyUnknownDefaultsToWarning(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	conflict := "some weird message that does not match keywords"
	result := scanner.BuildResult([]string{"claude-code"}, []string{conflict})
	assert.Equal(t, SeverityWarning, result.SeverityMap()[conflict])
}

func TestToolScannerKnownToolsIncludesAll(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	kt := scanner.KnownTools()
	assert.Contains(t, kt, "claude-code")
	assert.Contains(t, kt, "cursor")
	assert.Contains(t, kt, "roo-code")
	assert.Contains(t, kt, "opencode")
}

func TestToolScannerKnownToolConfigPaths(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	for name, configPath := range scanner.KnownTools() {
		assert.NotEmpty(t, configPath, "%s should have a config path", name)
	}
}

func TestToolScannerMultipleToolsWithMixedConflicts(t *testing.T) {
	t.Parallel()
	scanner := NewToolScanner()
	conflicts := []string{
		"cursor: SQLite-based config detected, cannot read",
		"claude-code: global setting 'model' contradicts local value",
		"opencode: global setting 'theme' is compatible with local",
	}
	result := scanner.BuildResult([]string{"claude-code", "cursor", "opencode"}, conflicts)
	assert.Len(t, result.DetectedTools(), 3)
	assert.Len(t, result.Conflicts(), 3)
	assert.Equal(t, SeverityWarning, result.SeverityMap()[conflicts[0]])
	assert.Equal(t, SeverityConflict, result.SeverityMap()[conflicts[1]])
	assert.Equal(t, SeverityCompatible, result.SeverityMap()[conflicts[2]])
}
