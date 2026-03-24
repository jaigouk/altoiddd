package infrastructure

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ContextMapYAMLParser reads and writes context map YAML files.
type ContextMapYAMLParser struct{}

// Compile-time interface compliance checks.
var (
	_ application.ContextMapReader = (*ContextMapYAMLParser)(nil)
	_ application.ContextMapWriter = (*ContextMapYAMLParser)(nil)
)

// contextMapYAML is the top-level YAML structure for a context map file.
type contextMapYAML struct {
	Version       int                `yaml:"version,omitempty"`
	Project       string             `yaml:"project"`
	Contexts      []contextYAML      `yaml:"contexts"`
	Relationships []relationshipYAML `yaml:"relationships"`
}

// contextYAML is the YAML structure for a single bounded context sketch.
type contextYAML struct {
	Name            string       `yaml:"name"`
	Classification  string       `yaml:"classification"`
	Confidence      float64      `yaml:"confidence"`
	Actors          []string     `yaml:"actors"`
	WorkObjects     []string     `yaml:"work_objects"`
	BoundarySignals []signalYAML `yaml:"boundary_signals"`
	Stories         []string     `yaml:"stories"`
	Trust           string       `yaml:"trust"`
}

// signalYAML is the YAML structure for a single boundary signal.
type signalYAML struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// relationshipYAML is the YAML structure for a single context relationship.
type relationshipYAML struct {
	Upstream    string   `yaml:"upstream"`
	Downstream  string   `yaml:"downstream"`
	Type        string   `yaml:"type"`
	Shared      []string `yaml:"shared"`
	Description string   `yaml:"description,omitempty"`
}

// Read reads a ContextMap from a YAML file at path.
func (p *ContextMapYAMLParser) Read(ctx context.Context, path string) (*domain.ContextMap, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("reading context map: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading context map file %q: %w", path, err)
	}

	return p.Parse(data)
}

// Write writes a ContextMap to a YAML file at path.
func (p *ContextMapYAMLParser) Write(ctx context.Context, path string, cm *domain.ContextMap) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("writing context map: %w", err)
	}

	data, err := p.Serialize(cm)
	if err != nil {
		return fmt.Errorf("serializing context map: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing context map file %q: %w", path, err)
	}

	return nil
}

// Parse parses YAML bytes into a ContextMap.
func (p *ContextMapYAMLParser) Parse(data []byte) (*domain.ContextMap, error) {
	var doc contextMapYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing context map YAML: %w", err)
	}

	// Build bounded context sketches.
	contexts := make([]domain.BoundedContextSketch, 0, len(doc.Contexts))

	for i, c := range doc.Contexts {
		signals := make([]domain.BoundarySignal, 0, len(c.BoundarySignals))

		for j, s := range c.BoundarySignals {
			sigType, err := domain.NewSignalType(s.Type)
			if err != nil {
				return nil, fmt.Errorf("contexts[%d].boundary_signals[%d].type: %w", i, j, err)
			}

			sig, err := domain.NewBoundarySignal(sigType, s.Description)
			if err != nil {
				return nil, fmt.Errorf("contexts[%d].boundary_signals[%d]: %w", i, j, err)
			}

			signals = append(signals, sig)
		}

		trust, err := vo.ParseTrustLevel(c.Trust)
		if err != nil {
			return nil, fmt.Errorf("contexts[%d].trust: %w", i, err)
		}

		classification := vo.SubdomainClassification(c.Classification)

		sketch, err := domain.NewBoundedContextSketch(
			c.Name, classification, c.Confidence,
			c.Actors, c.WorkObjects, c.Stories,
			signals, trust,
		)
		if err != nil {
			return nil, fmt.Errorf("contexts[%d]: %w", i, err)
		}

		contexts = append(contexts, sketch)
	}

	// Build relationships.
	relationships := make([]domain.ContextRelationship, 0, len(doc.Relationships))

	for i, r := range doc.Relationships {
		relType, err := domain.NewRelationshipType(r.Type)
		if err != nil {
			return nil, fmt.Errorf("relationships[%d].type: %w", i, err)
		}

		rel, err := domain.NewContextRelationship(r.Upstream, r.Downstream, relType, r.Shared, r.Description)
		if err != nil {
			return nil, fmt.Errorf("relationships[%d]: %w", i, err)
		}

		relationships = append(relationships, rel)
	}

	cm, err := domain.NewContextMap(doc.Project, contexts, relationships)
	if err != nil {
		return nil, fmt.Errorf("creating context map: %w", err)
	}

	return cm, nil
}

// Serialize converts a ContextMap to YAML bytes.
func (p *ContextMapYAMLParser) Serialize(cm *domain.ContextMap) ([]byte, error) {
	doc := contextMapYAML{
		Version:       1,
		Project:       cm.Project(),
		Contexts:      make([]contextYAML, 0, len(cm.Contexts())),
		Relationships: make([]relationshipYAML, 0, len(cm.Relationships())),
	}

	for _, c := range cm.Contexts() {
		sigs := make([]signalYAML, 0, len(c.Signals()))
		for _, s := range c.Signals() {
			sigs = append(sigs, signalYAML{
				Type:        s.Type().String(),
				Description: s.Description(),
			})
		}

		doc.Contexts = append(doc.Contexts, contextYAML{
			Name:            c.Name(),
			Classification:  string(c.Classification()),
			Confidence:      c.Confidence(),
			Actors:          c.Actors(),
			WorkObjects:     c.WorkObjects(),
			BoundarySignals: sigs,
			Stories:         c.Stories(),
			Trust:           c.Trust().String(),
		})
	}

	for _, r := range cm.Relationships() {
		doc.Relationships = append(doc.Relationships, relationshipYAML{
			Upstream:    r.Upstream(),
			Downstream:  r.Downstream(),
			Type:        r.Type().String(),
			Shared:      r.Shared(),
			Description: r.Description(),
		})
	}

	data, err := yaml.Marshal(&doc)
	if err != nil {
		return nil, fmt.Errorf("serializing context map YAML: %w", err)
	}

	return data, nil
}
