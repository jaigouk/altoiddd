package domain

// DetectedTool is an immutable value object representing a detected AI coding tool.
type DetectedTool struct {
	name       string
	configPath string
	version    string
}

// NewDetectedTool creates a DetectedTool value object.
// Use empty string for configPath/version if unknown.
func NewDetectedTool(name, configPath, version string) DetectedTool {
	return DetectedTool{name: name, configPath: configPath, version: version}
}

// Name returns the tool identifier.
func (dt DetectedTool) Name() string { return dt.name }

// ConfigPath returns the config path (empty if unknown).
func (dt DetectedTool) ConfigPath() string { return dt.configPath }

// Version returns the version string (empty if unknown).
func (dt DetectedTool) Version() string { return dt.version }

// Equal returns true if two DetectedTools have the same values.
func (dt DetectedTool) Equal(other DetectedTool) bool {
	return dt.name == other.name && dt.configPath == other.configPath && dt.version == other.version
}
