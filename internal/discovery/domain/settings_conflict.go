package domain

// SettingsConflictSeverity classifies the severity of a global settings conflict.
type SettingsConflictSeverity string

// SettingsConflictSeverity constants.
const (
	// SettingsSeverityInfo indicates a global config exists with no local counterpart.
	SettingsSeverityInfo SettingsConflictSeverity = "info"
	// SettingsSeverityWarning indicates global and local configs both exist but differ.
	SettingsSeverityWarning SettingsConflictSeverity = "warning"
)

// AllSettingsConflictSeverities returns all valid severity values.
func AllSettingsConflictSeverities() []SettingsConflictSeverity {
	return []SettingsConflictSeverity{SettingsSeverityInfo, SettingsSeverityWarning}
}

// SettingsConflict is an immutable value object representing a conflict between
// a global tool configuration and a local project configuration.
type SettingsConflict struct {
	tool     string
	global   string // global config path
	local    string // local config path (empty for global_only)
	typ      string // "global_only", "content_mismatch"
	severity SettingsConflictSeverity
	message  string
}

// NewSettingsConflict creates a SettingsConflict value object.
func NewSettingsConflict(tool, global, local, typ string, severity SettingsConflictSeverity, message string) SettingsConflict {
	return SettingsConflict{
		tool:     tool,
		global:   global,
		local:    local,
		typ:      typ,
		severity: severity,
		message:  message,
	}
}

// Tool returns the tool identifier.
func (sc SettingsConflict) Tool() string { return sc.tool }

// Global returns the global config path.
func (sc SettingsConflict) Global() string { return sc.global }

// Local returns the local config path.
func (sc SettingsConflict) Local() string { return sc.local }

// Type returns the conflict type.
func (sc SettingsConflict) Type() string { return sc.typ }

// Severity returns the conflict severity.
func (sc SettingsConflict) Severity() SettingsConflictSeverity { return sc.severity }

// Message returns the human-readable conflict description.
func (sc SettingsConflict) Message() string { return sc.message }

// Description returns a formatted description suitable for display.
func (sc SettingsConflict) Description() string { return sc.tool + ": " + sc.message }
