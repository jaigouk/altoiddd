package domain

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DDDVersion represents the version metadata of a DDD.md document.
// It is a value object - immutable after creation.
type DDDVersion struct {
	version          int
	round            string
	updated          string
	convergenceDelta int
}

// frontmatterData is used for YAML parsing/serialization.
type frontmatterData struct {
	Version          int    `yaml:"version,omitempty"`
	Round            string `yaml:"round,omitempty"`
	Updated          string `yaml:"updated,omitempty"`
	ConvergenceDelta int    `yaml:"convergence_delta,omitempty"`
}

// NewDDDVersion creates a new DDDVersion with the given values.
func NewDDDVersion(version int, round, updated string, convergenceDelta int) DDDVersion {
	return DDDVersion{
		version:          version,
		round:            round,
		updated:          updated,
		convergenceDelta: convergenceDelta,
	}
}

// Version returns the version number.
func (v DDDVersion) Version() int { return v.version }

// Round returns the round name (e.g., "express", "challenge", "simulate").
func (v DDDVersion) Round() string { return v.round }

// Updated returns the update date string.
func (v DDDVersion) Updated() string { return v.updated }

// ConvergenceDelta returns the convergence delta from the last round.
func (v DDDVersion) ConvergenceDelta() int { return v.convergenceDelta }

// Increment creates a new DDDVersion with incremented version number
// and updated metadata. The original DDDVersion is unchanged (immutability).
func (v DDDVersion) Increment(round string, convergenceDelta int, updatedAt time.Time) DDDVersion {
	return DDDVersion{
		version:          v.version + 1,
		round:            round,
		updated:          updatedAt.Format("2006-01-02"),
		convergenceDelta: convergenceDelta,
	}
}

// ParseDDDVersion extracts version metadata from DDD.md content.
// If no frontmatter is present or it's invalid, returns a zero-version.
func ParseDDDVersion(content string) (DDDVersion, error) {
	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		return DDDVersion{}, fmt.Errorf("extracting frontmatter: %w", err)
	}

	if frontmatter == "" {
		return DDDVersion{}, nil
	}

	var data frontmatterData
	if err := yaml.Unmarshal([]byte(frontmatter), &data); err != nil {
		return DDDVersion{}, fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	return DDDVersion{
		version:          data.Version,
		round:            data.Round,
		updated:          data.Updated,
		convergenceDelta: data.ConvergenceDelta,
	}, nil
}

// ApplyVersion updates or adds version frontmatter to DDD.md content.
// If frontmatter exists, it is replaced. If not, it is prepended.
func ApplyVersion(content string, version DDDVersion) string {
	data := frontmatterData{
		Version:          version.version,
		Round:            version.round,
		Updated:          version.updated,
		ConvergenceDelta: version.convergenceDelta,
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
