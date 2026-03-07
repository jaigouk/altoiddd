// Package events defines domain events for cross-context communication.
package events

import (
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// DomainModelGenerated is emitted when a DomainModel passes all invariant checks.
type DomainModelGenerated struct {
	modelID              string
	domainStories        []vo.DomainStory
	ubiquitousLanguage   []vo.TermEntry
	boundedContexts      []vo.DomainBoundedContext
	contextRelationships []vo.ContextRelationship
	aggregateDesigns     []vo.AggregateDesign
}

// NewDomainModelGenerated creates a DomainModelGenerated event.
func NewDomainModelGenerated(
	modelID string,
	domainStories []vo.DomainStory,
	ubiquitousLanguage []vo.TermEntry,
	boundedContexts []vo.DomainBoundedContext,
	contextRelationships []vo.ContextRelationship,
	aggregateDesigns []vo.AggregateDesign,
) DomainModelGenerated {
	ds := make([]vo.DomainStory, len(domainStories))
	copy(ds, domainStories)
	ul := make([]vo.TermEntry, len(ubiquitousLanguage))
	copy(ul, ubiquitousLanguage)
	bc := make([]vo.DomainBoundedContext, len(boundedContexts))
	copy(bc, boundedContexts)
	cr := make([]vo.ContextRelationship, len(contextRelationships))
	copy(cr, contextRelationships)
	ad := make([]vo.AggregateDesign, len(aggregateDesigns))
	copy(ad, aggregateDesigns)
	return DomainModelGenerated{
		modelID:              modelID,
		domainStories:        ds,
		ubiquitousLanguage:   ul,
		boundedContexts:      bc,
		contextRelationships: cr,
		aggregateDesigns:     ad,
	}
}

// ModelID returns the domain model identifier.
func (e DomainModelGenerated) ModelID() string { return e.modelID }

// DomainStories returns a defensive copy of domain stories.
func (e DomainModelGenerated) DomainStories() []vo.DomainStory {
	out := make([]vo.DomainStory, len(e.domainStories))
	copy(out, e.domainStories)
	return out
}

// UbiquitousLanguage returns a defensive copy of term entries.
func (e DomainModelGenerated) UbiquitousLanguage() []vo.TermEntry {
	out := make([]vo.TermEntry, len(e.ubiquitousLanguage))
	copy(out, e.ubiquitousLanguage)
	return out
}

// BoundedContexts returns a defensive copy of bounded contexts.
func (e DomainModelGenerated) BoundedContexts() []vo.DomainBoundedContext {
	out := make([]vo.DomainBoundedContext, len(e.boundedContexts))
	copy(out, e.boundedContexts)
	return out
}

// ContextRelationships returns a defensive copy of context relationships.
func (e DomainModelGenerated) ContextRelationships() []vo.ContextRelationship {
	out := make([]vo.ContextRelationship, len(e.contextRelationships))
	copy(out, e.contextRelationships)
	return out
}

// AggregateDesigns returns a defensive copy of aggregate designs.
func (e DomainModelGenerated) AggregateDesigns() []vo.AggregateDesign {
	out := make([]vo.AggregateDesign, len(e.aggregateDesigns))
	copy(out, e.aggregateDesigns)
	return out
}

// ConfigsGenerated is emitted when tool configs are approved and ready for output.
type ConfigsGenerated struct {
	toolNames   []string
	outputPaths []string
}

// NewConfigsGenerated creates a ConfigsGenerated event.
func NewConfigsGenerated(toolNames, outputPaths []string) ConfigsGenerated {
	tn := make([]string, len(toolNames))
	copy(tn, toolNames)
	op := make([]string, len(outputPaths))
	copy(op, outputPaths)
	return ConfigsGenerated{toolNames: tn, outputPaths: op}
}

// ToolNames returns a defensive copy of tool names.
func (e ConfigsGenerated) ToolNames() []string {
	out := make([]string, len(e.toolNames))
	copy(out, e.toolNames)
	return out
}

// OutputPaths returns a defensive copy of output paths.
func (e ConfigsGenerated) OutputPaths() []string {
	out := make([]string, len(e.outputPaths))
	copy(out, e.outputPaths)
	return out
}
