package infrastructure

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	challengeapp "github.com/alto-cli/alto/internal/challenge/application"
	"github.com/alto-cli/alto/internal/challenge/domain"
	"github.com/alto-cli/alto/internal/shared/infrastructure/markdown"
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
// If no frontmatter is present or it's unclosed, returns a zero-version
// without error — matches pre-Wave-2 lenient behaviour used by the
// challenge round-trip flow.
func ParseDDDVersionFromContent(content string) (domain.DDDVersion, error) {
	raw, _, hasFrontmatter, err := markdown.ExtractFrontmatter(content)
	if err != nil {
		// Unclosed frontmatter is treated as "no frontmatter" here — author
		// typos surface elsewhere via the schema rule, not via panic on
		// version parsing.
		if errors.Is(err, markdown.ErrMissingClosingDelimiter) {
			return domain.DDDVersion{}, nil
		}
		return domain.DDDVersion{}, fmt.Errorf("extracting frontmatter: %w", err)
	}

	if !hasFrontmatter || raw == "" {
		return domain.DDDVersion{}, nil
	}

	data, err := markdown.ParseTyped[frontmatterData](raw)
	if err != nil {
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

// extractBody returns the content after frontmatter (or all content if no
// frontmatter / unclosed). Used by ApplyVersionToContent to preserve the
// document body when rewriting the frontmatter block.
func extractBody(content string) string {
	_, body, hasFrontmatter, err := markdown.ExtractFrontmatter(content)
	if err != nil || !hasFrontmatter {
		return content
	}
	return body
}
