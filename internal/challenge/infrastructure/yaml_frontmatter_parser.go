package infrastructure

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	challengeapp "github.com/alty-cli/alty/internal/challenge/application"
	"github.com/alty-cli/alty/internal/challenge/domain"
)

// frontmatterData is used for YAML parsing/serialization.
type frontmatterData struct {
	Version          int    `yaml:"version,omitempty"`
	Round            string `yaml:"round,omitempty"`
	Updated          string `yaml:"updated,omitempty"`
	ConvergenceDelta int    `yaml:"convergence_delta,omitempty"`
}

// YAMLFrontmatterParser implements DDDVersionParser using YAML frontmatter.
type YAMLFrontmatterParser struct{}

// Compile-time interface check.
var _ challengeapp.DDDVersionParser = (*YAMLFrontmatterParser)(nil)

// NewYAMLFrontmatterParser creates a new YAMLFrontmatterParser.
func NewYAMLFrontmatterParser() *YAMLFrontmatterParser {
	return &YAMLFrontmatterParser{}
}

// ParseVersion extracts version metadata from DDD.md content.
// If no frontmatter is present or it's invalid, returns a zero-version.
func (p *YAMLFrontmatterParser) ParseVersion(content string) (domain.DDDVersion, error) {
	return ParseDDDVersionFromContent(content)
}

// ApplyVersion updates or adds version frontmatter to DDD.md content.
func (p *YAMLFrontmatterParser) ApplyVersion(content string, version domain.DDDVersion) string {
	return ApplyVersionToContent(content, version)
}

// ParseDDDVersionFromContent extracts version metadata from DDD.md content.
// If no frontmatter is present or it's invalid, returns a zero-version.
func ParseDDDVersionFromContent(content string) (domain.DDDVersion, error) {
	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		return domain.DDDVersion{}, fmt.Errorf("extracting frontmatter: %w", err)
	}

	if frontmatter == "" {
		return domain.DDDVersion{}, nil
	}

	var data frontmatterData
	if err := yaml.Unmarshal([]byte(frontmatter), &data); err != nil {
		return domain.DDDVersion{}, fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	return domain.NewDDDVersion(
		data.Version,
		data.Round,
		data.Updated,
		data.ConvergenceDelta,
	), nil
}

// ApplyVersionToContent updates or adds version frontmatter to DDD.md content.
// If frontmatter exists, it is replaced. If not, it is prepended.
func ApplyVersionToContent(content string, version domain.DDDVersion) string {
	data := frontmatterData{
		Version:          version.Version(),
		Round:            version.Round(),
		Updated:          version.Updated(),
		ConvergenceDelta: version.ConvergenceDelta(),
	}

	yamlBytes, _ := yaml.Marshal(&data)
	newFrontmatter := "---\n" + string(yamlBytes) + "---\n"

	// Check if content has frontmatter to replace
	body := extractBody(content)

	if body == "" {
		return newFrontmatter
	}

	return newFrontmatter + "\n" + body
}

// extractFrontmatter extracts YAML frontmatter from content.
// Returns empty string if no valid frontmatter is found.
func extractFrontmatter(content string) (string, error) {
	if !strings.HasPrefix(content, "---") {
		return "", nil
	}

	// Find the closing ---
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		// Unclosed frontmatter
		return "", nil
	}

	// Extract frontmatter content (after first ---, before closing ---)
	frontmatter := strings.TrimSpace(rest[:idx])
	return frontmatter, nil
}

// extractBody returns the content after frontmatter (or all content if no frontmatter).
func extractBody(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}

	// Find the closing ---
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		// No closing delimiter, treat as no frontmatter
		return content
	}

	// Return everything after the closing ---
	afterClosing := rest[idx+4:] // +4 for "\n---"
	return strings.TrimPrefix(afterClosing, "\n")
}
