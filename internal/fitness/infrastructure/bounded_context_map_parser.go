package infrastructure

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/alto-cli/alto/internal/fitness/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// boundedContextMapYAML is the YAML structure for parsing.
type boundedContextMapYAML struct {
	Project struct {
		Name        string `yaml:"name"`
		RootPackage string `yaml:"root_package"`
	} `yaml:"project"`
	BoundedContexts []struct {
		Name           string   `yaml:"name"`
		ModulePath     string   `yaml:"module_path"`
		Classification string   `yaml:"classification"`
		Layers         []string `yaml:"layers"`
		Relationships  []struct {
			Target    string `yaml:"target"`
			Direction string `yaml:"direction"`
			Pattern   string `yaml:"pattern"`
		} `yaml:"relationships"`
	} `yaml:"bounded_contexts"`
}

// BoundedContextMapParser parses bounded_context_map.yaml files.
type BoundedContextMapParser struct{}

// NewBoundedContextMapParser creates a new parser.
func NewBoundedContextMapParser() *BoundedContextMapParser {
	return &BoundedContextMapParser{}
}

// Parse reads and parses a bounded context map YAML file.
func (p *BoundedContextMapParser) Parse(_ context.Context, path string) (*domain.BoundedContextMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading bounded context map: %w", err)
	}

	var yamlData boundedContextMapYAML
	if unmarshalErr := yaml.Unmarshal(data, &yamlData); unmarshalErr != nil {
		return nil, fmt.Errorf("parsing bounded context map: %w", unmarshalErr)
	}

	if validateErr := p.validate(&yamlData); validateErr != nil {
		return nil, validateErr
	}

	contexts, err := p.convertContexts(yamlData.BoundedContexts)
	if err != nil {
		return nil, err
	}

	bcMap := domain.NewBoundedContextMap(
		yamlData.Project.Name,
		yamlData.Project.RootPackage,
		contexts,
	)

	return &bcMap, nil
}

func (p *BoundedContextMapParser) validate(data *boundedContextMapYAML) error {
	if data.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}
	if data.Project.RootPackage == "" {
		return fmt.Errorf("project.root_package is required")
	}
	return nil
}

func (p *BoundedContextMapParser) convertContexts(yamlContexts []struct {
	Name           string   `yaml:"name"`
	ModulePath     string   `yaml:"module_path"`
	Classification string   `yaml:"classification"`
	Layers         []string `yaml:"layers"`
	Relationships  []struct {
		Target    string `yaml:"target"`
		Direction string `yaml:"direction"`
		Pattern   string `yaml:"pattern"`
	} `yaml:"relationships"`
},
) ([]domain.BoundedContextEntry, error) {
	var contexts []domain.BoundedContextEntry

	for _, yc := range yamlContexts {
		classification, err := p.parseClassification(yc.Classification)
		if err != nil {
			return nil, fmt.Errorf("context '%s': %w", yc.Name, err)
		}

		relationships, err := p.convertRelationships(yc.Relationships)
		if err != nil {
			return nil, fmt.Errorf("context '%s': %w", yc.Name, err)
		}

		entry := domain.NewBoundedContextEntry(
			yc.Name,
			yc.ModulePath,
			classification,
			yc.Layers,
			relationships,
		)
		contexts = append(contexts, entry)
	}

	return contexts, nil
}

func (p *BoundedContextMapParser) parseClassification(s string) (vo.SubdomainClassification, error) {
	switch s {
	case "core":
		return vo.SubdomainCore, nil
	case "supporting":
		return vo.SubdomainSupporting, nil
	case "generic":
		return vo.SubdomainGeneric, nil
	default:
		return "", fmt.Errorf("invalid classification '%s': must be core, supporting, or generic", s)
	}
}

func (p *BoundedContextMapParser) convertRelationships(yamlRels []struct {
	Target    string `yaml:"target"`
	Direction string `yaml:"direction"`
	Pattern   string `yaml:"pattern"`
},
) ([]domain.ContextRelationship, error) {
	var relationships []domain.ContextRelationship

	for _, yr := range yamlRels {
		direction, err := p.parseDirection(yr.Direction)
		if err != nil {
			return nil, err
		}

		pattern, err := p.parsePattern(yr.Pattern)
		if err != nil {
			return nil, err
		}

		rel := domain.NewContextRelationship(yr.Target, direction, pattern)
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

func (p *BoundedContextMapParser) parseDirection(s string) (domain.RelationshipDirection, error) {
	switch s {
	case "upstream":
		return domain.RelationshipUpstream, nil
	case "downstream":
		return domain.RelationshipDownstream, nil
	default:
		return "", fmt.Errorf("invalid direction '%s': must be upstream or downstream", s)
	}
}

func (p *BoundedContextMapParser) parsePattern(s string) (domain.RelationshipPattern, error) {
	switch s {
	case "domain_event":
		return domain.PatternDomainEvent, nil
	case "shared_kernel":
		return domain.PatternSharedKernel, nil
	case "acl":
		return domain.PatternACL, nil
	case "open_host":
		return domain.PatternOpenHost, nil
	default:
		return "", fmt.Errorf("invalid pattern '%s': must be domain_event, shared_kernel, acl, or open_host", s)
	}
}
