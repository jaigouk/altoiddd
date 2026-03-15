package domain

// ProjectConfig is a value object that captures detected project settings
// for use during bootstrap file generation.
type ProjectConfig struct {
	name          string
	language      string
	modulePath    string
	detectedTools []string
}

// NewProjectConfig creates a ProjectConfig with all fields.
func NewProjectConfig(name, language, modulePath string, tools []string) ProjectConfig {
	t := make([]string, len(tools))
	copy(t, tools)

	return ProjectConfig{
		name:          name,
		language:      language,
		modulePath:    modulePath,
		detectedTools: t,
	}
}

// Name returns the project name.
func (c ProjectConfig) Name() string { return c.name }

// Language returns the detected language (e.g. "go", "python", "typescript").
func (c ProjectConfig) Language() string { return c.language }

// ModulePath returns the module path extracted from the manifest file.
func (c ProjectConfig) ModulePath() string { return c.modulePath }

// DetectedTools returns a defensive copy of detected tool names.
func (c ProjectConfig) DetectedTools() []string {
	out := make([]string, len(c.detectedTools))
	copy(out, c.detectedTools)
	return out
}
