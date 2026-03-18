package domain

// ProjectDetectionResult is a value object that captures what was found
// when scanning a project directory. Used by the init command to decide
// whether to run the new-project or existing-project (rescue) path.
type ProjectDetectionResult struct {
	hasSourceCode   bool
	language        string // "go", "python", "typescript", ""
	hasDocsFolder   bool
	hasAltoConfig   bool
	hasAIToolConfig bool
	manifestPath    string // "go.mod", "pyproject.toml", etc.
	modulePath      string // extracted from manifest (e.g. "github.com/user/project")
}

// NewProjectDetectionResult creates a ProjectDetectionResult with all fields.
func NewProjectDetectionResult(
	hasSourceCode bool,
	language string,
	hasDocsFolder bool,
	hasAltoConfig bool,
	hasAIToolConfig bool,
	manifestPath string,
	modulePath string,
) ProjectDetectionResult {
	return ProjectDetectionResult{
		hasSourceCode:   hasSourceCode,
		language:        language,
		hasDocsFolder:   hasDocsFolder,
		hasAltoConfig:   hasAltoConfig,
		hasAIToolConfig: hasAIToolConfig,
		manifestPath:    manifestPath,
		modulePath:      modulePath,
	}
}

// HasSourceCode returns true if source code was detected.
func (r ProjectDetectionResult) HasSourceCode() bool { return r.hasSourceCode }

// Language returns the detected language (e.g. "go", "python", "typescript").
func (r ProjectDetectionResult) Language() string { return r.language }

// HasDocsFolder returns true if a docs/ directory was found.
func (r ProjectDetectionResult) HasDocsFolder() bool { return r.hasDocsFolder }

// HasAltoConfig returns true if .alto/config.toml was found.
func (r ProjectDetectionResult) HasAltoConfig() bool { return r.hasAltoConfig }

// HasAIToolConfig returns true if AI tool configs (.claude/, .cursor/, CLAUDE.md) were found.
func (r ProjectDetectionResult) HasAIToolConfig() bool { return r.hasAIToolConfig }

// ManifestPath returns the path to the detected manifest file (e.g. "go.mod").
func (r ProjectDetectionResult) ManifestPath() string { return r.manifestPath }

// ModulePath returns the module/package path extracted from the manifest.
func (r ProjectDetectionResult) ModulePath() string { return r.modulePath }

// IsExistingProject returns true if the directory appears to contain an existing project.
func (r ProjectDetectionResult) IsExistingProject() bool {
	return r.hasSourceCode || r.hasDocsFolder
}

// IsAmbiguous returns true if there are docs but no source code,
// indicating an unclear project state that may need user confirmation.
func (r ProjectDetectionResult) IsAmbiguous() bool {
	return r.hasDocsFolder && !r.hasSourceCode
}
