// Package infrastructure provides adapters for the Discovery bounded context.
package infrastructure

import (
	"context"
	"os"
	"path/filepath"

	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
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

// ScanConflicts scans for global settings conflicts by comparing global
// tool config directories (under homeDir) with local project config directories.
func (s *FilesystemToolScanner) ScanConflicts(
	_ context.Context,
	projectDir string,
) ([]domain.SettingsConflict, error) {
	if _, err := os.Stat(s.homeDir); os.IsNotExist(err) {
		return nil, nil
	}

	var conflicts []domain.SettingsConflict

	for toolName, configRelPath := range toolDirs {
		globalPath := filepath.Join(s.homeDir, configRelPath)

		globalInfo, globalErr := os.Stat(globalPath)
		if globalErr != nil {
			if os.IsPermission(globalErr) {
				conflicts = append(conflicts, domain.NewSettingsConflict(
					toolName, globalPath, "", "global_only", domain.SettingsSeverityWarning,
					"config directory not readable (permission denied)",
				))
			}
			continue
		}
		if !globalInfo.IsDir() {
			continue
		}

		// Check if global dir is readable.
		if _, readErr := os.ReadDir(globalPath); readErr != nil {
			if os.IsPermission(readErr) {
				conflicts = append(conflicts, domain.NewSettingsConflict(
					toolName, globalPath, "", "global_only", domain.SettingsSeverityWarning,
					"config directory not readable (permission denied)",
				))
			}
			continue
		}

		localPath := filepath.Join(projectDir, configRelPath)
		localInfo, localErr := os.Stat(localPath)

		if localErr != nil {
			// Global exists, no local — informational.
			conflicts = append(conflicts, domain.NewSettingsConflict(
				toolName, globalPath, "", "global_only", domain.SettingsSeverityInfo,
				"global config exists, no local override",
			))
			continue
		}

		if !localInfo.IsDir() {
			continue
		}

		// Both global and local exist — compare contents.
		conflicts = append(conflicts, s.compareConfigs(toolName, globalPath, localPath)...)
	}

	return conflicts, nil
}

// compareConfigs compares files in global and local config directories.
func (s *FilesystemToolScanner) compareConfigs(toolName, globalPath, localPath string) []domain.SettingsConflict {
	var conflicts []domain.SettingsConflict

	entries, err := os.ReadDir(globalPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		globalFile := filepath.Join(globalPath, entry.Name())
		localFile := filepath.Join(localPath, entry.Name())

		_, localErr := os.Stat(localFile)
		if localErr != nil {
			// Global file exists, no local counterpart — skip (not a conflict).
			continue
		}

		globalContent, gErr := os.ReadFile(globalFile)
		if gErr != nil {
			continue
		}

		localContent, lErr := os.ReadFile(localFile)
		if lErr != nil {
			continue
		}

		if string(globalContent) != string(localContent) {
			conflicts = append(conflicts, domain.NewSettingsConflict(
				toolName, globalFile, localFile, "content_mismatch", domain.SettingsSeverityWarning,
				"global and local '"+entry.Name()+"' have different content",
			))
		}
	}

	return conflicts
}
