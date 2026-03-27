package domain

import (
	"fmt"
	"strings"
	"time"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// WorkflowType — string enum classifying a researched workflow.
// ---------------------------------------------------------------------------

// WorkflowType classifies a researched workflow: happy_path, failure_case, or secondary.
type WorkflowType string

// WorkflowType constants.
const (
	WorkflowTypeHappyPath   WorkflowType = "happy_path"
	WorkflowTypeFailureCase WorkflowType = "failure_case"
	WorkflowTypeSecondary   WorkflowType = "secondary"
)

var validWorkflowTypes = map[WorkflowType]struct{}{
	WorkflowTypeHappyPath:   {},
	WorkflowTypeFailureCase: {},
	WorkflowTypeSecondary:   {},
}

// NewWorkflowType creates a WorkflowType from a string, returning an error if invalid.
func NewWorkflowType(s string) (WorkflowType, error) {
	wt := WorkflowType(s)
	if err := wt.Validate(); err != nil {
		return "", err
	}

	return wt, nil
}

// AllWorkflowTypes returns all valid WorkflowType values.
func AllWorkflowTypes() []WorkflowType {
	return []WorkflowType{WorkflowTypeHappyPath, WorkflowTypeFailureCase, WorkflowTypeSecondary}
}

// String returns the string representation of a WorkflowType.
func (w WorkflowType) String() string {
	return string(w)
}

// Validate checks whether the WorkflowType holds a valid value.
func (w WorkflowType) Validate() error {
	if _, ok := validWorkflowTypes[w]; !ok {
		return fmt.Errorf("invalid workflow type %q: %w", string(w), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (w WorkflowType) MarshalText() ([]byte, error) {
	if err := w.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling workflow type: %w", err)
	}

	return []byte(w), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (w *WorkflowType) UnmarshalText(data []byte) error {
	parsed, err := NewWorkflowType(string(data))
	if err != nil {
		return err
	}

	*w = parsed

	return nil
}

// ---------------------------------------------------------------------------
// SearchMetadata — query provenance for a domain research run.
// ---------------------------------------------------------------------------

// SearchMetadata captures provenance for a domain research run: queries used,
// source counts, and search duration. Zero-value is valid.
type SearchMetadata struct {
	queriesUsed   []string
	totalSources  int
	usefulSources int
	duration      time.Duration
}

// NewSearchMetadata creates a SearchMetadata. No validation — zero-value is valid.
func NewSearchMetadata(queriesUsed []string, totalSources, usefulSources int, duration time.Duration) SearchMetadata {
	queriesCopy := make([]string, len(queriesUsed))
	copy(queriesCopy, queriesUsed)

	return SearchMetadata{
		queriesUsed:   queriesCopy,
		totalSources:  totalSources,
		usefulSources: usefulSources,
		duration:      duration,
	}
}

// QueriesUsed returns a defensive copy of the queries used during research.
func (m SearchMetadata) QueriesUsed() []string {
	out := make([]string, len(m.queriesUsed))
	copy(out, m.queriesUsed)

	return out
}

// TotalSources returns the total number of sources found.
func (m SearchMetadata) TotalSources() int { return m.totalSources }

// UsefulSources returns the number of useful sources found.
func (m SearchMetadata) UsefulSources() int { return m.usefulSources }

// Duration returns the search duration.
func (m SearchMetadata) Duration() time.Duration { return m.duration }

// ---------------------------------------------------------------------------
// ResearchedActor — an actor identified from AI domain research.
// ---------------------------------------------------------------------------

// ResearchedActor represents an actor identified from AI domain research
// with name, role, and source URLs. Trust level added in Phase 4.4.
type ResearchedActor struct {
	name       string
	role       string
	sourceURLs []string
}

// NewResearchedActor creates a ResearchedActor. Name must not be empty.
func NewResearchedActor(name, role string, sourceURLs []string) (ResearchedActor, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ResearchedActor{}, fmt.Errorf("researched actor name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	urlsCopy := make([]string, len(sourceURLs))
	copy(urlsCopy, sourceURLs)

	return ResearchedActor{
		name:       name,
		role:       role,
		sourceURLs: urlsCopy,
	}, nil
}

// Name returns the actor's name.
func (a ResearchedActor) Name() string { return a.name }

// Role returns the actor's role description.
func (a ResearchedActor) Role() string { return a.role }

// SourceURLs returns a defensive copy of the actor's source URLs.
func (a ResearchedActor) SourceURLs() []string {
	out := make([]string, len(a.sourceURLs))
	copy(out, a.sourceURLs)

	return out
}

// ---------------------------------------------------------------------------
// ResearchedEntity — an entity identified from AI domain research.
// ---------------------------------------------------------------------------

// ResearchedEntity represents an entity identified from AI domain research
// with name, properties, and source URLs. Trust level added in Phase 4.4.
type ResearchedEntity struct {
	name       string
	properties []string
	sourceURLs []string
}

// NewResearchedEntity creates a ResearchedEntity. Name must not be empty.
func NewResearchedEntity(name string, properties, sourceURLs []string) (ResearchedEntity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ResearchedEntity{}, fmt.Errorf("researched entity name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	propsCopy := make([]string, len(properties))
	copy(propsCopy, properties)

	urlsCopy := make([]string, len(sourceURLs))
	copy(urlsCopy, sourceURLs)

	return ResearchedEntity{
		name:       name,
		properties: propsCopy,
		sourceURLs: urlsCopy,
	}, nil
}

// Name returns the entity's name.
func (e ResearchedEntity) Name() string { return e.name }

// Properties returns a defensive copy of the entity's properties.
func (e ResearchedEntity) Properties() []string {
	out := make([]string, len(e.properties))
	copy(out, e.properties)

	return out
}

// SourceURLs returns a defensive copy of the entity's source URLs.
func (e ResearchedEntity) SourceURLs() []string {
	out := make([]string, len(e.sourceURLs))
	copy(out, e.sourceURLs)

	return out
}

// ---------------------------------------------------------------------------
// WorkflowStep — a single coarse-grained step in a researched workflow.
// ---------------------------------------------------------------------------

// WorkflowStep represents a single coarse-grained step in a researched workflow:
// sequence number, actor, activity, and work object.
type WorkflowStep struct {
	sequence   int
	actor      string
	activity   string
	workObject string
}

// NewWorkflowStep creates a WorkflowStep. Actor, activity, and workObject must not
// be empty. Sequence is informational (negative values are allowed).
func NewWorkflowStep(sequence int, actor, activity, workObject string) (WorkflowStep, error) {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return WorkflowStep{}, fmt.Errorf("workflow step actor must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	activity = strings.TrimSpace(activity)
	if activity == "" {
		return WorkflowStep{}, fmt.Errorf("workflow step activity must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	workObject = strings.TrimSpace(workObject)
	if workObject == "" {
		return WorkflowStep{}, fmt.Errorf("workflow step work object must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	return WorkflowStep{
		sequence:   sequence,
		actor:      actor,
		activity:   activity,
		workObject: workObject,
	}, nil
}

// Sequence returns the step's sequence number.
func (s WorkflowStep) Sequence() int { return s.sequence }

// Actor returns the step's actor.
func (s WorkflowStep) Actor() string { return s.actor }

// Activity returns the step's activity.
func (s WorkflowStep) Activity() string { return s.activity }

// WorkObject returns the step's work object.
func (s WorkflowStep) WorkObject() string { return s.workObject }

// ---------------------------------------------------------------------------
// ResearchedWorkflow — a typed workflow discovered by AI research.
// ---------------------------------------------------------------------------

// ResearchedWorkflow represents a typed workflow discovered by AI research,
// containing ordered WorkflowSteps with source URLs.
type ResearchedWorkflow struct {
	name       string
	wfType     WorkflowType
	steps      []WorkflowStep
	sourceURLs []string
}

// NewResearchedWorkflow creates a ResearchedWorkflow. Name must not be empty and
// wfType must be valid. Steps can be empty (zero-length is valid).
func NewResearchedWorkflow(name string, wfType WorkflowType, steps []WorkflowStep, sourceURLs []string) (ResearchedWorkflow, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ResearchedWorkflow{}, fmt.Errorf("researched workflow name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := wfType.Validate(); err != nil {
		return ResearchedWorkflow{}, fmt.Errorf("validating workflow type: %w", err)
	}

	stepsCopy := make([]WorkflowStep, len(steps))
	copy(stepsCopy, steps)

	urlsCopy := make([]string, len(sourceURLs))
	copy(urlsCopy, sourceURLs)

	return ResearchedWorkflow{
		name:       name,
		wfType:     wfType,
		steps:      stepsCopy,
		sourceURLs: urlsCopy,
	}, nil
}

// Name returns the workflow's name.
func (w ResearchedWorkflow) Name() string { return w.name }

// WorkflowType returns the workflow's type classification.
func (w ResearchedWorkflow) WorkflowType() WorkflowType { return w.wfType }

// Steps returns a defensive copy of the workflow's steps.
func (w ResearchedWorkflow) Steps() []WorkflowStep {
	out := make([]WorkflowStep, len(w.steps))
	copy(out, w.steps)

	return out
}

// SourceURLs returns a defensive copy of the workflow's source URLs.
func (w ResearchedWorkflow) SourceURLs() []string {
	out := make([]string, len(w.sourceURLs))
	copy(out, w.sourceURLs)

	return out
}

// ---------------------------------------------------------------------------
// RegulatoryItem — a regulatory requirement from AI domain research.
// ---------------------------------------------------------------------------

// RegulatoryItem represents a regulatory requirement identified from AI domain
// research with name, description, and source URLs.
type RegulatoryItem struct {
	name        string
	description string
	sourceURLs  []string
}

// NewRegulatoryItem creates a RegulatoryItem. Name must not be empty.
func NewRegulatoryItem(name, description string, sourceURLs []string) (RegulatoryItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return RegulatoryItem{}, fmt.Errorf("regulatory item name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	urlsCopy := make([]string, len(sourceURLs))
	copy(urlsCopy, sourceURLs)

	return RegulatoryItem{
		name:        name,
		description: description,
		sourceURLs:  urlsCopy,
	}, nil
}

// Name returns the regulatory item's name.
func (r RegulatoryItem) Name() string { return r.name }

// Description returns the regulatory item's description.
func (r RegulatoryItem) Description() string { return r.description }

// SourceURLs returns a defensive copy of the regulatory item's source URLs.
func (r RegulatoryItem) SourceURLs() []string {
	out := make([]string, len(r.sourceURLs))
	copy(out, r.sourceURLs)

	return out
}

// ---------------------------------------------------------------------------
// ExistingSoftware — known software in the domain from AI research.
// ---------------------------------------------------------------------------

// ExistingSoftware represents known software in the domain identified from AI
// research with name, description, and a single source URL.
type ExistingSoftware struct {
	name        string
	description string
	sourceURL   string
}

// NewExistingSoftware creates an ExistingSoftware. Name must not be empty.
func NewExistingSoftware(name, description, sourceURL string) (ExistingSoftware, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ExistingSoftware{}, fmt.Errorf("existing software name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	return ExistingSoftware{
		name:        name,
		description: description,
		sourceURL:   sourceURL,
	}, nil
}

// Name returns the software's name.
func (s ExistingSoftware) Name() string { return s.name }

// Description returns the software's description.
func (s ExistingSoftware) Description() string { return s.description }

// SourceURL returns the software's source URL.
func (s ExistingSoftware) SourceURL() string { return s.sourceURL }

// ---------------------------------------------------------------------------
// ResearchQuality — computed quality assessment of AI domain research output.
// ---------------------------------------------------------------------------

// Quality floor thresholds from spike Section 9.
const (
	QualityFloorActors        = 3
	QualityFloorEntities      = 3
	QualityFloorWorkflowSteps = 5
	QualityFloorUsefulSources = 5
)

// ResearchQuality represents the computed quality assessment of AI domain research
// output: counts of actors, entities, workflow steps, useful sources, and whether
// it meets the quality floor.
type ResearchQuality struct {
	actorCount      int
	entityCount     int
	workflowStepCnt int
	usefulSourceCnt int
	meetsFloor      bool
}

// ActorCount returns the number of actors found.
func (q ResearchQuality) ActorCount() int { return q.actorCount }

// EntityCount returns the number of entities found.
func (q ResearchQuality) EntityCount() int { return q.entityCount }

// WorkflowStepCount returns the total number of workflow steps found.
func (q ResearchQuality) WorkflowStepCount() int { return q.workflowStepCnt }

// UsefulSourceCount returns the number of useful sources found.
func (q ResearchQuality) UsefulSourceCount() int { return q.usefulSourceCnt }

// MeetsFloor returns true when all quality floor thresholds are met.
func (q ResearchQuality) MeetsFloor() bool { return q.meetsFloor }

// ComputeResearchQuality calculates a ResearchQuality from research output.
// Quality floor: actors>=3, entities>=3, workflowSteps>=5, usefulSources>=5.
func ComputeResearchQuality(
	actors []ResearchedActor,
	entities []ResearchedEntity,
	workflows []ResearchedWorkflow,
	usefulSources int,
) ResearchQuality {
	totalSteps := 0
	for _, wf := range workflows {
		totalSteps += len(wf.steps)
	}

	meetsFloor := len(actors) >= QualityFloorActors &&
		len(entities) >= QualityFloorEntities &&
		totalSteps >= QualityFloorWorkflowSteps &&
		usefulSources >= QualityFloorUsefulSources

	return ResearchQuality{
		actorCount:      len(actors),
		entityCount:     len(entities),
		workflowStepCnt: totalSteps,
		usefulSourceCnt: usefulSources,
		meetsFloor:      meetsFloor,
	}
}

// ---------------------------------------------------------------------------
// DomainResearchResult — top-level container for all AI domain research output.
// ---------------------------------------------------------------------------

// DomainResearchResult is the top-level container for all AI domain research output:
// actors, entities, workflows, failure modes, regulatory items, existing software,
// and auto-computed research quality.
//
//nolint:revive // "DomainResearchResult" is the ubiquitous language term from the domain model.
type DomainResearchResult struct {
	domain       string
	meta         SearchMetadata
	actors       []ResearchedActor
	entities     []ResearchedEntity
	workflows    []ResearchedWorkflow
	failureModes []string
	regulatory   []RegulatoryItem
	software     []ExistingSoftware
}

// NewDomainResearchResult creates a DomainResearchResult. Domain must not be empty.
func NewDomainResearchResult(
	domain string,
	meta SearchMetadata,
	actors []ResearchedActor,
	entities []ResearchedEntity,
	workflows []ResearchedWorkflow,
	failureModes []string,
	regulatory []RegulatoryItem,
	software []ExistingSoftware,
) (DomainResearchResult, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return DomainResearchResult{}, fmt.Errorf("domain research result domain must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	actorsCopy := make([]ResearchedActor, len(actors))
	copy(actorsCopy, actors)

	entitiesCopy := make([]ResearchedEntity, len(entities))
	copy(entitiesCopy, entities)

	workflowsCopy := make([]ResearchedWorkflow, len(workflows))
	copy(workflowsCopy, workflows)

	modesCopy := make([]string, len(failureModes))
	copy(modesCopy, failureModes)

	regCopy := make([]RegulatoryItem, len(regulatory))
	copy(regCopy, regulatory)

	swCopy := make([]ExistingSoftware, len(software))
	copy(swCopy, software)

	return DomainResearchResult{
		domain:       domain,
		meta:         meta,
		actors:       actorsCopy,
		entities:     entitiesCopy,
		workflows:    workflowsCopy,
		failureModes: modesCopy,
		regulatory:   regCopy,
		software:     swCopy,
	}, nil
}

// Domain returns the research domain name.
func (r DomainResearchResult) Domain() string { return r.domain }

// SearchMetadata returns the search metadata.
func (r DomainResearchResult) SearchMetadata() SearchMetadata { return r.meta }

// Actors returns a defensive copy of the researched actors.
func (r DomainResearchResult) Actors() []ResearchedActor {
	out := make([]ResearchedActor, len(r.actors))
	copy(out, r.actors)

	return out
}

// Entities returns a defensive copy of the researched entities.
func (r DomainResearchResult) Entities() []ResearchedEntity {
	out := make([]ResearchedEntity, len(r.entities))
	copy(out, r.entities)

	return out
}

// Workflows returns a defensive copy of the researched workflows.
func (r DomainResearchResult) Workflows() []ResearchedWorkflow {
	out := make([]ResearchedWorkflow, len(r.workflows))
	copy(out, r.workflows)

	return out
}

// FailureModes returns a defensive copy of the failure modes.
func (r DomainResearchResult) FailureModes() []string {
	out := make([]string, len(r.failureModes))
	copy(out, r.failureModes)

	return out
}

// Regulatory returns a defensive copy of the regulatory items.
func (r DomainResearchResult) Regulatory() []RegulatoryItem {
	out := make([]RegulatoryItem, len(r.regulatory))
	copy(out, r.regulatory)

	return out
}

// Software returns a defensive copy of the existing software items.
func (r DomainResearchResult) Software() []ExistingSoftware {
	out := make([]ExistingSoftware, len(r.software))
	copy(out, r.software)

	return out
}

// Quality returns the computed research quality, auto-calculated from the
// result's own actors, entities, workflows, and useful source count.
func (r DomainResearchResult) Quality() ResearchQuality {
	return ComputeResearchQuality(r.actors, r.entities, r.workflows, r.meta.usefulSources)
}
