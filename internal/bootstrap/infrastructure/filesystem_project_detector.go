package infrastructure

import (
	"os"
	"path/filepath"

	"github.com/alty-cli/alty/internal/bootstrap/domain"
)

// manifestEntry maps a manifest file to its language.
type manifestEntry struct {
	file     string
	language string
}

// knownManifests lists language manifest files in priority order.
var knownManifests = []manifestEntry{
	{"go.mod", "go"},
	{"pyproject.toml", "python"},
	{"requirements.txt", "python"},
	{"package.json", "typescript"},
}

// aiToolPaths lists paths that indicate AI coding tool configuration.
var aiToolPaths = []string{
	".claude",
	".cursor",
	"CLAUDE.md",
}

// FileSystemProjectDetector detects existing project state by scanning
// the file system for known manifest files, docs, and tool configs.
type FileSystemProjectDetector struct{}

// Detect scans projectDir and returns a ProjectDetectionResult describing
// what was found.
func (d *FileSystemProjectDetector) Detect(projectDir string) (domain.ProjectDetectionResult, error) {
	var (
		hasSourceCode   bool
		language        string
		manifestPath    string
		hasDocsFolder   bool
		hasAltyConfig   bool
		hasAIToolConfig bool
	)

	// Check for language manifests.
	for _, m := range knownManifests {
		if fileExists(filepath.Join(projectDir, m.file)) {
			hasSourceCode = true
			language = m.language
			manifestPath = m.file

			break
		}
	}

	// Check for docs/ folder.
	if dirExists(filepath.Join(projectDir, "docs")) {
		hasDocsFolder = true
	}

	// Check for .alty/config.toml.
	if fileExists(filepath.Join(projectDir, ".alty", "config.toml")) {
		hasAltyConfig = true
	}

	// Check for AI tool configs.
	for _, p := range aiToolPaths {
		if pathExists(filepath.Join(projectDir, p)) {
			hasAIToolConfig = true

			break
		}
	}

	return domain.NewProjectDetectionResult(
		hasSourceCode,
		language,
		hasDocsFolder,
		hasAltyConfig,
		hasAIToolConfig,
		manifestPath,
	), nil
}

// fileExists returns true if path is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists returns true if path is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// pathExists returns true if path exists (file or directory).
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
