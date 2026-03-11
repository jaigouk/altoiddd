package domain

import (
	"fmt"
	"strings"
)

// GlobalConfig is an immutable value object representing a tool's global
// configuration directory path.
type GlobalConfig struct {
	tool string
	path string
}

// NewGlobalConfig creates a GlobalConfig value object with validation.
func NewGlobalConfig(tool, path string) (GlobalConfig, error) {
	if strings.TrimSpace(tool) == "" {
		return GlobalConfig{}, fmt.Errorf("tool name required")
	}
	if strings.TrimSpace(path) == "" {
		return GlobalConfig{}, fmt.Errorf("path required")
	}
	return GlobalConfig{tool: tool, path: path}, nil
}

// Tool returns the tool identifier.
func (gc GlobalConfig) Tool() string { return gc.tool }

// Path returns the global config directory path.
func (gc GlobalConfig) Path() string { return gc.path }

// Equal returns true if two GlobalConfigs have the same values.
func (gc GlobalConfig) Equal(other GlobalConfig) bool {
	return gc.tool == other.tool && gc.path == other.path
}
