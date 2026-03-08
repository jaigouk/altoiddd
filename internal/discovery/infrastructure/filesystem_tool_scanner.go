// Package infrastructure provides adapters for the Discovery bounded context.
package infrastructure

import (
	"context"
	"os"
	"path/filepath"

	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
)

// toolDirs maps tool names to their config directory relative to home.
var toolDirs = map[string]string{
	"claude-code": ".claude",
	"cursor":      ".cursor",
	"roo-code":    ".roo",
	"opencode":    filepath.Join(".config", "opencode"),
}

// FilesystemToolScanner detects installed AI coding tools by scanning
// the filesystem for known configuration directories.
type FilesystemToolScanner struct {
	homeDir string
}

// Compile-time interface check.
var _ discoveryapp.ToolDetection = (*FilesystemToolScanner)(nil)

// NewFilesystemToolScanner creates a FilesystemToolScanner.
// If homeDir is empty, defaults to the user's home directory.
func NewFilesystemToolScanner(homeDir string) *FilesystemToolScanner {
	if homeDir == "" {
		h, err := os.UserHomeDir()
		if err == nil {
			homeDir = h
		}
	}
	return &FilesystemToolScanner{homeDir: homeDir}
}

// Detect detects installed AI coding tools by checking config directories.
func (s *FilesystemToolScanner) Detect(
	_ context.Context,
	_ string,
) ([]string, error) {
	if _, err := os.Stat(s.homeDir); os.IsNotExist(err) {
		return nil, nil
	}

	var detected []string
	for toolName, configRelPath := range toolDirs {
		configPath := filepath.Join(s.homeDir, configRelPath)
		_, err := os.Stat(configPath)
		if err == nil || os.IsPermission(err) {
			detected = append(detected, toolName)
		}
	}
	return detected, nil
}

// ScanConflicts scans for configuration conflicts between detected tools.
func (s *FilesystemToolScanner) ScanConflicts(
	_ context.Context,
	_ string,
) ([]string, error) {
	if _, err := os.Stat(s.homeDir); os.IsNotExist(err) {
		return nil, nil
	}

	var conflicts []string
	cursorDir := filepath.Join(s.homeDir, toolDirs["cursor"])
	_, err := os.Stat(cursorDir)
	if err == nil {
		conflicts = append(conflicts, "cursor: SQLite-based config detected, cannot read")
	} else if os.IsPermission(err) {
		conflicts = append(conflicts, "cursor: config directory not readable (permission denied)")
	}

	return conflicts, nil
}
