package ddd

import (
	"fmt"
	"regexp"
	"strings"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/shared/domain/events"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// DomainModel is the aggregate root for the Domain Model bounded context.
// It manages the complete set of DDD artifacts: domain stories, ubiquitous language,
// bounded contexts, context relationships, and aggregate designs.
//
// Invariants (checked on Finalize):
//  1. Every UL term must appear in at least one DomainStory (word boundary match).
//  2. Every BoundedContext must have a SubdomainClassification.
//  3. Every Core subdomain must have at least one AggregateDesign.
//  4. Ambiguous terms must have per-context definitions.
type DomainModel struct {
	modelID       string
	stories       []vo.DomainStory
	language      *UbiquitousLanguage
	contexts      []vo.DomainBoundedContext
	relationships []vo.ContextRelationship
	aggregates    []vo.AggregateDesign
	domainEvents  []events.DomainModelGenerated
	warnings      []string
}

// NewDomainModel creates a new DomainModel aggregate root with a generated ID.
func NewDomainModel(modelID string) *DomainModel {
	return &DomainModel{
		modelID:  modelID,
		language: NewUbiquitousLanguage(),
	}
}

// ModelID returns the model identifier.
func (dm *DomainModel) ModelID() string { return dm.modelID }

// DomainStories returns a defensive copy of all domain stories.
func (dm *DomainModel) DomainStories() []vo.DomainStory {
	out := make([]vo.DomainStory, len(dm.stories))
	copy(out, dm.stories)
	return out
}

// UbiquitousLanguage returns the ubiquitous language entity.
func (dm *DomainModel) UbiquitousLanguage() *UbiquitousLanguage {
	return dm.language
}

// BoundedContexts returns a defensive copy of all bounded contexts.
func (dm *DomainModel) BoundedContexts() []vo.DomainBoundedContext {
	out := make([]vo.DomainBoundedContext, len(dm.contexts))
	copy(out, dm.contexts)
	return out
}

// ContextRelationships returns a defensive copy of all context relationships.
func (dm *DomainModel) ContextRelationships() []vo.ContextRelationship {
	out := make([]vo.ContextRelationship, len(dm.relationships))
	copy(out, dm.relationships)
	return out
}

// AggregateDesigns returns a defensive copy of all aggregate designs.
func (dm *DomainModel) AggregateDesigns() []vo.AggregateDesign {
	out := make([]vo.AggregateDesign, len(dm.aggregates))
	copy(out, dm.aggregates)
	return out
}

// Events returns a defensive copy of domain events produced by this aggregate.
func (dm *DomainModel) Events() []events.DomainModelGenerated {
	out := make([]events.DomainModelGenerated, len(dm.domainEvents))
	copy(out, dm.domainEvents)
	return out
}

// Warnings returns a defensive copy of warnings produced during Finalize.
func (dm *DomainModel) Warnings() []string {
	out := make([]string, len(dm.warnings))
	copy(out, dm.warnings)
	return out
}

// -- Commands -----------------------------------------------------------------

// AddDomainStory adds a business process narrative.
func (dm *DomainModel) AddDomainStory(story vo.DomainStory) error {
	if strings.TrimSpace(story.Name()) == "" {
		return fmt.Errorf("story name cannot be empty")
	}

	nameLower := strings.ToLower(story.Name())
	for _, s := range dm.stories {
		if strings.ToLower(s.Name()) == nameLower {
			return fmt.Errorf("domain story '%s' already exists: %w",
				story.Name(), domainerrors.ErrAlreadyExists)
		}
	}

	dm.stories = append(dm.stories, story)
	return nil
}

// AddTerm adds a term to the ubiquitous language glossary.
func (dm *DomainModel) AddTerm(term, definition, contextName string, sourceQuestionIDs []string) error {
	return dm.language.AddTerm(term, definition, contextName, sourceQuestionIDs)
}

// ReassignTermsToContext moves all terms from one context to another.
func (dm *DomainModel) ReassignTermsToContext(fromContext, toContext string) error {
	if strings.TrimSpace(fromContext) == "" {
		return fmt.Errorf("source context name cannot be empty")
	}
	if strings.TrimSpace(toContext) == "" {
		return fmt.Errorf("target context name cannot be empty")
	}

	fromLower := strings.ToLower(fromContext)
	oldTerms := dm.language.Terms()
	newTerms := make([]vo.TermEntry, 0, len(oldTerms))
	for _, t := range oldTerms {
		if strings.ToLower(t.ContextName()) == fromLower {
			newTerms = append(newTerms, t.WithContextName(toContext))
		} else {
			newTerms = append(newTerms, t)
		}
	}
	dm.language.SetTerms(newTerms)
	return nil
}

// AddBoundedContext adds a bounded context.
func (dm *DomainModel) AddBoundedContext(ctx vo.DomainBoundedContext) error {
	if strings.TrimSpace(ctx.Name()) == "" {
		return fmt.Errorf("context name cannot be empty")
	}

	nameLower := strings.ToLower(ctx.Name())
	for _, c := range dm.contexts {
		if strings.ToLower(c.Name()) == nameLower {
			return fmt.Errorf("bounded context '%s' already exists", ctx.Name())
		}
	}

	dm.contexts = append(dm.contexts, ctx)
	return nil
}

// ClassifySubdomain classifies a bounded context's subdomain type.
func (dm *DomainModel) ClassifySubdomain(contextName string, classification vo.SubdomainClassification, rationale string) error {
	nameLower := strings.ToLower(contextName)
	for i, ctx := range dm.contexts {
		if strings.ToLower(ctx.Name()) == nameLower {
			dm.contexts[i] = vo.NewDomainBoundedContext(
				ctx.Name(),
				ctx.Responsibility(),
				ctx.KeyDomainObjects(),
				&classification,
				rationale,
			)
			return nil
		}
	}
	return fmt.Errorf("bounded context '%s' not found", contextName)
}

// AddContextRelationship adds a relationship between two bounded contexts.
func (dm *DomainModel) AddContextRelationship(rel vo.ContextRelationship) error {
	if strings.TrimSpace(rel.Upstream()) == "" || strings.TrimSpace(rel.Downstream()) == "" {
		return fmt.Errorf("relationship upstream and downstream cannot be empty")
	}
	dm.relationships = append(dm.relationships, rel)
	return nil
}

// DesignAggregate adds an aggregate design for a bounded context.
func (dm *DomainModel) DesignAggregate(agg vo.AggregateDesign) error {
	if strings.TrimSpace(agg.Name()) == "" {
		return fmt.Errorf("aggregate name cannot be empty")
	}
	if strings.TrimSpace(agg.ContextName()) == "" {
		return fmt.Errorf("aggregate context name cannot be empty")
	}

	aggKey := strings.ToLower(agg.Name()) + "\x00" + strings.ToLower(agg.ContextName())
	for _, a := range dm.aggregates {
		existingKey := strings.ToLower(a.Name()) + "\x00" + strings.ToLower(a.ContextName())
		if existingKey == aggKey {
			return fmt.Errorf("aggregate design '%s' already exists in context '%s'",
				agg.Name(), agg.ContextName())
		}
	}

	dm.aggregates = append(dm.aggregates, agg)
	return nil
}

// Finalize validates all invariants and emits DomainModelGenerated.
func (dm *DomainModel) Finalize() error {
	dm.warnings = nil

	if err := dm.checkTermsInStories(); err != nil {
		return err
	}
	if err := dm.checkContextClassifications(); err != nil {
		return err
	}
	if err := dm.checkCoreAggregates(); err != nil {
		return err
	}
	if err := dm.checkAmbiguousTerms(); err != nil {
		return err
	}

	dm.domainEvents = append(dm.domainEvents, events.NewDomainModelGenerated(
		dm.modelID,
		dm.stories,
		dm.language.Terms(),
		dm.contexts,
		dm.relationships,
		dm.aggregates,
	))
	return nil
}

// -- Invariant checks (private) -----------------------------------------------

func (dm *DomainModel) checkTermsInStories() error {
	var parts []string
	for _, story := range dm.stories {
		parts = append(parts, strings.ToLower(story.Name()))
		for _, a := range story.Actors() {
			parts = append(parts, strings.ToLower(a))
		}
		parts = append(parts, strings.ToLower(story.Trigger()))
		for _, s := range story.Steps() {
			parts = append(parts, strings.ToLower(s))
		}
		for _, o := range story.Observations() {
			parts = append(parts, strings.ToLower(o))
		}
	}
	storyText := strings.Join(parts, " ")

	for _, entry := range dm.language.Terms() {
		pattern := `\b` + regexp.QuoteMeta(strings.ToLower(entry.Term())) + `\b`
		matched, err := regexp.MatchString(pattern, storyText)
		if err != nil || !matched {
			return fmt.Errorf("term '%s' not found in any domain story: %w",
				entry.Term(), domainerrors.ErrInvariantViolation)
		}
	}
	return nil
}

func (dm *DomainModel) checkContextClassifications() error {
	for _, ctx := range dm.contexts {
		if ctx.Classification() == nil {
			return fmt.Errorf("bounded context '%s' has no classification: %w",
				ctx.Name(), domainerrors.ErrInvariantViolation)
		}
	}

	if len(dm.contexts) > 0 {
		allGeneric := true
		for _, ctx := range dm.contexts {
			if *ctx.Classification() != vo.SubdomainGeneric {
				allGeneric = false
				break
			}
		}
		if allGeneric {
			dm.warnings = append(dm.warnings,
				"All bounded contexts are classified as Generic. "+
					"A project with no Core or Supporting subdomain likely "+
					"has misclassified contexts.")
		}
	}
	return nil
}

func (dm *DomainModel) checkCoreAggregates() error {
	coreContexts := make(map[string]struct{})
	for _, ctx := range dm.contexts {
		if ctx.Classification() != nil && *ctx.Classification() == vo.SubdomainCore {
			coreContexts[ctx.Name()] = struct{}{}
		}
	}

	contextsWithAggs := make(map[string]struct{})
	for _, a := range dm.aggregates {
		contextsWithAggs[a.ContextName()] = struct{}{}
	}

	for coreName := range coreContexts {
		if _, ok := contextsWithAggs[coreName]; !ok {
			return fmt.Errorf("core subdomain '%s' has no aggregate design: %w",
				coreName, domainerrors.ErrInvariantViolation)
		}
	}
	return nil
}

func (dm *DomainModel) checkAmbiguousTerms() error {
	ambiguous := dm.language.FindAmbiguousTerms()
	for _, term := range ambiguous {
		if !dm.language.HasPerContextDefinitions(term) {
			return fmt.Errorf("ambiguous term '%s' needs per-context definitions: %w",
				term, domainerrors.ErrInvariantViolation)
		}
	}
	return nil
}
