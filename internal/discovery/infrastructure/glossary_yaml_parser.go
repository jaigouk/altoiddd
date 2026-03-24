package infrastructure

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/alto-cli/alto/internal/discovery/application"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// GlossaryYAMLParser reads and writes ubiquitous language glossary YAML files.
type GlossaryYAMLParser struct{}

// Compile-time interface compliance checks.
var (
	_ application.GlossaryReader = (*GlossaryYAMLParser)(nil)
	_ application.GlossaryWriter = (*GlossaryYAMLParser)(nil)
)

// glossaryYAML is the top-level YAML structure for a glossary file.
type glossaryYAML struct {
	Version int        `yaml:"version,omitempty"`
	Terms   []termYAML `yaml:"terms"`
}

// termYAML is the YAML structure for a single glossary term.
type termYAML struct {
	Term       string   `yaml:"term"`
	Definition string   `yaml:"definition"`
	Context    string   `yaml:"context"`
	Trust      string   `yaml:"trust"`
	Source     string   `yaml:"source,omitempty"`
	Stories    []string `yaml:"stories"`
	Aliases    []string `yaml:"aliases,omitempty"`
	Note       string   `yaml:"note,omitempty"`
}

// Read reads ubiquitous language entries from a YAML file at path.
func (p *GlossaryYAMLParser) Read(ctx context.Context, path string) ([]vo.UbiquitousLanguageEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("reading glossary: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading glossary file %q: %w", path, err)
	}

	return p.Parse(data)
}

// Write writes ubiquitous language entries to a YAML file at path.
func (p *GlossaryYAMLParser) Write(ctx context.Context, path string, entries []vo.UbiquitousLanguageEntry) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("writing glossary: %w", err)
	}

	data, err := p.Serialize(entries)
	if err != nil {
		return fmt.Errorf("serializing glossary: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing glossary file %q: %w", path, err)
	}

	return nil
}

// Parse parses YAML bytes into ubiquitous language entries.
func (p *GlossaryYAMLParser) Parse(data []byte) ([]vo.UbiquitousLanguageEntry, error) {
	var doc glossaryYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing glossary YAML: %w", err)
	}

	if len(doc.Terms) == 0 {
		return nil, fmt.Errorf("glossary YAML has no terms")
	}

	entries := make([]vo.UbiquitousLanguageEntry, 0, len(doc.Terms))

	for i, t := range doc.Terms {
		if t.Term == "" {
			return nil, fmt.Errorf("terms[%d].term: must not be empty", i)
		}

		if t.Definition == "" {
			return nil, fmt.Errorf("terms[%d].definition: must not be empty", i)
		}

		if t.Context == "" {
			return nil, fmt.Errorf("terms[%d].context: must not be empty", i)
		}

		if t.Trust == "" {
			return nil, fmt.Errorf("terms[%d].trust: must not be empty", i)
		}

		if t.Stories == nil {
			return nil, fmt.Errorf("terms[%d].stories: must not be empty", i)
		}

		trust, err := vo.ParseTrustLevel(t.Trust)
		if err != nil {
			return nil, fmt.Errorf("terms[%d].trust: %w", i, err)
		}

		entry, err := vo.NewUbiquitousLanguageEntry(t.Term, t.Definition, t.Context, trust, t.Source)
		if err != nil {
			return nil, fmt.Errorf("terms[%d]: %w", i, err)
		}

		if len(t.Aliases) > 0 {
			entry = entry.WithAliases(t.Aliases)
		}

		if t.Note != "" {
			entry = entry.WithNote(t.Note)
		}

		if len(t.Stories) > 0 {
			entry = entry.WithStories(t.Stories)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Serialize converts ubiquitous language entries to YAML bytes.
func (p *GlossaryYAMLParser) Serialize(entries []vo.UbiquitousLanguageEntry) ([]byte, error) {
	doc := glossaryYAML{
		Version: 1,
		Terms:   make([]termYAML, 0, len(entries)),
	}

	for _, e := range entries {
		t := termYAML{
			Term:       e.Term(),
			Definition: e.Definition(),
			Context:    e.Context(),
			Trust:      e.Trust().String(),
			Source:     e.Source(),
			Stories:    e.Stories(),
			Aliases:    e.Aliases(),
			Note:       e.Note(),
		}

		// Normalize nil slices to empty for clean YAML output.
		if t.Stories == nil {
			t.Stories = []string{}
		}

		doc.Terms = append(doc.Terms, t)
	}

	data, err := yaml.Marshal(&doc)
	if err != nil {
		return nil, fmt.Errorf("serializing glossary YAML: %w", err)
	}

	return data, nil
}
