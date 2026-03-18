package domain

import (
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// RelationshipDirection indicates the direction of a context relationship.
type RelationshipDirection string

// Relationship direction constants.
const (
	RelationshipUpstream   RelationshipDirection = "upstream"
	RelationshipDownstream RelationshipDirection = "downstream"
)

// AllRelationshipDirections returns all valid direction values.
func AllRelationshipDirections() []RelationshipDirection {
	return []RelationshipDirection{RelationshipUpstream, RelationshipDownstream}
}

// RelationshipPattern indicates the integration pattern used.
type RelationshipPattern string

// Relationship pattern constants.
const (
	PatternDomainEvent  RelationshipPattern = "domain_event"
	PatternSharedKernel RelationshipPattern = "shared_kernel"
	PatternACL          RelationshipPattern = "acl"
	PatternOpenHost     RelationshipPattern = "open_host"
)

// AllRelationshipPatterns returns all valid pattern values.
func AllRelationshipPatterns() []RelationshipPattern {
	return []RelationshipPattern{PatternDomainEvent, PatternSharedKernel, PatternACL, PatternOpenHost}
}

// ContextRelationship describes a relationship to another bounded context.
type ContextRelationship struct {
	target    string
	direction RelationshipDirection
	pattern   RelationshipPattern
}

// NewContextRelationship creates a ContextRelationship value object.
func NewContextRelationship(target string, direction RelationshipDirection, pattern RelationshipPattern) ContextRelationship {
	return ContextRelationship{
		target:    target,
		direction: direction,
		pattern:   pattern,
	}
}

// Target returns the target context name.
func (r ContextRelationship) Target() string { return r.target }

// Direction returns the relationship direction.
func (r ContextRelationship) Direction() RelationshipDirection { return r.direction }

// Pattern returns the integration pattern.
func (r ContextRelationship) Pattern() RelationshipPattern { return r.pattern }

// BoundedContextEntry represents a single bounded context in the map.
type BoundedContextEntry struct {
	name           string
	modulePath     string
	classification vo.SubdomainClassification
	layers         []string
	relationships  []ContextRelationship
}

// NewBoundedContextEntry creates a BoundedContextEntry value object.
func NewBoundedContextEntry(
	name, modulePath string,
	classification vo.SubdomainClassification,
	layers []string,
	relationships []ContextRelationship,
) BoundedContextEntry {
	// Defensive copies
	layersCopy := make([]string, len(layers))
	copy(layersCopy, layers)

	relsCopy := make([]ContextRelationship, len(relationships))
	copy(relsCopy, relationships)

	return BoundedContextEntry{
		name:           name,
		modulePath:     modulePath,
		classification: classification,
		layers:         layersCopy,
		relationships:  relsCopy,
	}
}

// Name returns the context name.
func (e BoundedContextEntry) Name() string { return e.name }

// ModulePath returns the module path (e.g., "bootstrap").
func (e BoundedContextEntry) ModulePath() string { return e.modulePath }

// Classification returns the subdomain classification.
func (e BoundedContextEntry) Classification() vo.SubdomainClassification { return e.classification }

// Layers returns a defensive copy of layer names.
func (e BoundedContextEntry) Layers() []string {
	out := make([]string, len(e.layers))
	copy(out, e.layers)
	return out
}

// Relationships returns a defensive copy of relationships.
func (e BoundedContextEntry) Relationships() []ContextRelationship {
	out := make([]ContextRelationship, len(e.relationships))
	copy(out, e.relationships)
	return out
}

// BoundedContextMap is a collection of bounded contexts with their relationships.
type BoundedContextMap struct {
	projectName string
	rootPackage string
	contexts    []BoundedContextEntry
}

// NewBoundedContextMap creates a BoundedContextMap value object.
func NewBoundedContextMap(projectName, rootPackage string, contexts []BoundedContextEntry) BoundedContextMap {
	// Defensive copy
	ctxCopy := make([]BoundedContextEntry, len(contexts))
	copy(ctxCopy, contexts)

	return BoundedContextMap{
		projectName: projectName,
		rootPackage: rootPackage,
		contexts:    ctxCopy,
	}
}

// ProjectName returns the project name.
func (m BoundedContextMap) ProjectName() string { return m.projectName }

// RootPackage returns the root package path.
func (m BoundedContextMap) RootPackage() string { return m.rootPackage }

// Contexts returns a defensive copy of all contexts.
func (m BoundedContextMap) Contexts() []BoundedContextEntry {
	out := make([]BoundedContextEntry, len(m.contexts))
	copy(out, m.contexts)
	return out
}

// FindContext finds a context by name.
func (m BoundedContextMap) FindContext(name string) (BoundedContextEntry, bool) {
	for _, ctx := range m.contexts {
		if ctx.name == name {
			return ctx, true
		}
	}
	return BoundedContextEntry{}, false
}

// ContextNames returns all context names.
func (m BoundedContextMap) ContextNames() []string {
	names := make([]string, len(m.contexts))
	for i, ctx := range m.contexts {
		names[i] = ctx.name
	}
	return names
}

// ContextsWithClassification returns contexts matching the given classification.
func (m BoundedContextMap) ContextsWithClassification(classification vo.SubdomainClassification) []BoundedContextEntry {
	var result []BoundedContextEntry
	for _, ctx := range m.contexts {
		if ctx.classification == classification {
			result = append(result, ctx)
		}
	}
	return result
}
