// Package events defines domain events for cross-context communication.
package events

import (
	"encoding/json"
	"fmt"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
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

// GapAnalysisCompleted is emitted when a rescue gap analysis completes execution.
type GapAnalysisCompleted struct {
	analysisID   string
	projectDir   string
	gapsFound    int
	gapsResolved int
}

// NewGapAnalysisCompleted creates a GapAnalysisCompleted event.
func NewGapAnalysisCompleted(analysisID, projectDir string, gapsFound, gapsResolved int) GapAnalysisCompleted {
	return GapAnalysisCompleted{
		analysisID:   analysisID,
		projectDir:   projectDir,
		gapsFound:    gapsFound,
		gapsResolved: gapsResolved,
	}
}

// AnalysisID returns the analysis identifier.
func (e GapAnalysisCompleted) AnalysisID() string { return e.analysisID }

// ProjectDir returns the project directory.
func (e GapAnalysisCompleted) ProjectDir() string { return e.projectDir }

// GapsFound returns the number of gaps found.
func (e GapAnalysisCompleted) GapsFound() int { return e.gapsFound }

// GapsResolved returns the number of gaps resolved.
func (e GapAnalysisCompleted) GapsResolved() int { return e.gapsResolved }

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

// --- JSON serialization (event bus roundtrip) ---

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e DomainModelGenerated) MarshalJSON() ([]byte, error) {
	type proxy struct {
		ModelID              string                    `json:"model_id"`
		DomainStories        []vo.DomainStory          `json:"domain_stories"`
		UbiquitousLanguage   []vo.TermEntry            `json:"ubiquitous_language"`
		BoundedContexts      []vo.DomainBoundedContext `json:"bounded_contexts"`
		ContextRelationships []vo.ContextRelationship  `json:"context_relationships"`
		AggregateDesigns     []vo.AggregateDesign      `json:"aggregate_designs"`
	}
	data, err := json.Marshal(proxy{
		ModelID:              e.modelID,
		DomainStories:        e.domainStories,
		UbiquitousLanguage:   e.ubiquitousLanguage,
		BoundedContexts:      e.boundedContexts,
		ContextRelationships: e.contextRelationships,
		AggregateDesigns:     e.aggregateDesigns,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling DomainModelGenerated: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *DomainModelGenerated) UnmarshalJSON(data []byte) error {
	type proxy struct {
		ModelID              string                    `json:"model_id"`
		DomainStories        []vo.DomainStory          `json:"domain_stories"`
		UbiquitousLanguage   []vo.TermEntry            `json:"ubiquitous_language"`
		BoundedContexts      []vo.DomainBoundedContext `json:"bounded_contexts"`
		ContextRelationships []vo.ContextRelationship  `json:"context_relationships"`
		AggregateDesigns     []vo.AggregateDesign      `json:"aggregate_designs"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling DomainModelGenerated: %w", err)
	}
	e.modelID = p.ModelID
	e.domainStories = p.DomainStories
	e.ubiquitousLanguage = p.UbiquitousLanguage
	e.boundedContexts = p.BoundedContexts
	e.contextRelationships = p.ContextRelationships
	e.aggregateDesigns = p.AggregateDesigns
	return nil
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e GapAnalysisCompleted) MarshalJSON() ([]byte, error) {
	type proxy struct {
		AnalysisID   string `json:"analysis_id"`
		ProjectDir   string `json:"project_dir"`
		GapsFound    int    `json:"gaps_found"`
		GapsResolved int    `json:"gaps_resolved"`
	}
	data, err := json.Marshal(proxy{
		AnalysisID:   e.analysisID,
		ProjectDir:   e.projectDir,
		GapsFound:    e.gapsFound,
		GapsResolved: e.gapsResolved,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling GapAnalysisCompleted: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *GapAnalysisCompleted) UnmarshalJSON(data []byte) error {
	type proxy struct {
		AnalysisID   string `json:"analysis_id"`
		ProjectDir   string `json:"project_dir"`
		GapsFound    int    `json:"gaps_found"`
		GapsResolved int    `json:"gaps_resolved"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling GapAnalysisCompleted: %w", err)
	}
	e.analysisID = p.AnalysisID
	e.projectDir = p.ProjectDir
	e.gapsFound = p.GapsFound
	e.gapsResolved = p.GapsResolved
	return nil
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e ConfigsGenerated) MarshalJSON() ([]byte, error) {
	type proxy struct {
		ToolNames   []string `json:"tool_names"`
		OutputPaths []string `json:"output_paths"`
	}
	data, err := json.Marshal(proxy{
		ToolNames:   e.toolNames,
		OutputPaths: e.outputPaths,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling ConfigsGenerated: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *ConfigsGenerated) UnmarshalJSON(data []byte) error {
	type proxy struct {
		ToolNames   []string `json:"tool_names"`
		OutputPaths []string `json:"output_paths"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling ConfigsGenerated: %w", err)
	}
	e.toolNames = p.ToolNames
	e.outputPaths = p.OutputPaths
	return nil
}
