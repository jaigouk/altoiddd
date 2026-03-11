package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettingsConflictSeverity_AllValues(t *testing.T) {
	t.Parallel()
	all := AllSettingsConflictSeverities()
	assert.Len(t, all, 2)
	assert.Contains(t, all, SettingsSeverityInfo)
	assert.Contains(t, all, SettingsSeverityWarning)
}

func TestNewSettingsConflict_GlobalOnly(t *testing.T) {
	t.Parallel()
	sc := NewSettingsConflict("claude-code", "/home/.claude/settings.json", "", "global_only", SettingsSeverityInfo, "global config exists, no local override")

	assert.Equal(t, "claude-code", sc.Tool())
	assert.Equal(t, "/home/.claude/settings.json", sc.Global())
	assert.Empty(t, sc.Local())
	assert.Equal(t, "global_only", sc.Type())
	assert.Equal(t, SettingsSeverityInfo, sc.Severity())
	assert.Equal(t, "global config exists, no local override", sc.Message())
}

func TestNewSettingsConflict_ContentMismatch(t *testing.T) {
	t.Parallel()
	sc := NewSettingsConflict("cursor", "/home/.cursor/rules", "/proj/.cursor/rules", "content_mismatch", SettingsSeverityWarning, "global and local configs differ")

	assert.Equal(t, "cursor", sc.Tool())
	assert.Equal(t, "/home/.cursor/rules", sc.Global())
	assert.Equal(t, "/proj/.cursor/rules", sc.Local())
	assert.Equal(t, "content_mismatch", sc.Type())
	assert.Equal(t, SettingsSeverityWarning, sc.Severity())
	assert.Equal(t, "global and local configs differ", sc.Message())
}

func TestSettingsConflict_Description(t *testing.T) {
	t.Parallel()
	sc := NewSettingsConflict("claude-code", "/home/.claude/settings.json", "", "global_only", SettingsSeverityInfo, "global config exists")
	assert.Equal(t, "claude-code: global config exists", sc.Description())
}
