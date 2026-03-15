package infrastructure

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

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

	// Extract module path from manifest if found.
	modulePath := extractModulePath(projectDir, manifestPath)

	return domain.NewProjectDetectionResult(
		hasSourceCode,
		language,
		hasDocsFolder,
		hasAltyConfig,
		hasAIToolConfig,
		manifestPath,
		modulePath,
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

// extractModulePath reads the manifest file and extracts the module/package path.
// Supports go.mod, pyproject.toml, and package.json. Returns empty string on
// failure or for unsupported manifest types.
func extractModulePath(projectDir, manifestPath string) string {
	if manifestPath == "" {
		return ""
	}

	fullPath := filepath.Join(projectDir, manifestPath)

	switch manifestPath {
	case "go.mod":
		return extractGoModModule(fullPath)
	case "pyproject.toml":
		return extractPyprojectName(fullPath)
	case "package.json":
		return extractPackageJSONName(fullPath)
	default:
		return ""
	}
}

// extractGoModModule scans go.mod for the module directive and extracts the module path.
func extractGoModModule(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	return ""
}

// extractPyprojectName reads pyproject.toml and finds name under [project].
func extractPyprojectName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	inProject := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[project]" {
			inProject = true
			continue
		}
		if strings.HasPrefix(trimmed, "[") && trimmed != "[project]" {
			inProject = false
			continue
		}
		if inProject && strings.HasPrefix(trimmed, "name") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, "\"'")
				return val
			}
		}
	}

	return ""
}

// extractPackageJSONName reads package.json and extracts the "name" field.
func extractPackageJSONName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var pkg struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}

	return pkg.Name
}
