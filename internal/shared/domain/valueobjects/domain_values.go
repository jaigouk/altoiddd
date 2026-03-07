// Package valueobjects provides shared value objects for the alty domain.
package valueobjects

// SubdomainClassification classifies a subdomain per Khononov's decision tree.
type SubdomainClassification string

// Subdomain classification constants.
const (
	SubdomainCore       SubdomainClassification = "core"
	SubdomainSupporting SubdomainClassification = "supporting"
	SubdomainGeneric    SubdomainClassification = "generic"
)

// AllSubdomainClassifications returns all valid classification values.
func AllSubdomainClassifications() []SubdomainClassification {
	return []SubdomainClassification{SubdomainCore, SubdomainSupporting, SubdomainGeneric}
}

// DomainStory is a business process narrative using domain language.
type DomainStory struct {
	name         string
	actors       []string
	trigger      string
	steps        []string
	observations []string
}

// NewDomainStory creates a DomainStory with required fields and optional observations.
func NewDomainStory(name string, actors []string, trigger string, steps []string, observations []string) DomainStory {
	a := make([]string, len(actors))
	copy(a, actors)
	s := make([]string, len(steps))
	copy(s, steps)
	o := make([]string, len(observations))
	copy(o, observations)
	return DomainStory{
		name:         name,
		actors:       a,
		trigger:      trigger,
		steps:        s,
		observations: o,
	}
}

// Name returns the story name.
func (d DomainStory) Name() string { return d.name }

// Trigger returns what starts this process.
func (d DomainStory) Trigger() string { return d.trigger }

// Actors returns a defensive copy of the actors.
func (d DomainStory) Actors() []string {
	out := make([]string, len(d.actors))
	copy(out, d.actors)
	return out
}

// Steps returns a defensive copy of the steps.
func (d DomainStory) Steps() []string {
	out := make([]string, len(d.steps))
	copy(out, d.steps)
	return out
}

// Observations returns a defensive copy of the observations.
func (d DomainStory) Observations() []string {
	out := make([]string, len(d.observations))
	copy(out, d.observations)
	return out
}

// DomainBoundedContext is a DDD bounded context value object for the domain model.
// Named DomainBoundedContext to avoid collision with ddd.BoundedContext.
type DomainBoundedContext struct {
	classification          *SubdomainClassification
	name                    string
	responsibility          string
	classificationRationale string
	keyDomainObjects        []string
}

// NewDomainBoundedContext creates a DomainBoundedContext value object.
func NewDomainBoundedContext(
	name, responsibility string,
	keyDomainObjects []string,
	classification *SubdomainClassification,
	classificationRationale string,
) DomainBoundedContext {
	objs := make([]string, len(keyDomainObjects))
	copy(objs, keyDomainObjects)
	return DomainBoundedContext{
		name:                    name,
		responsibility:          responsibility,
		keyDomainObjects:        objs,
		classification:          classification,
		classificationRationale: classificationRationale,
	}
}

// Name returns the context name.
func (bc DomainBoundedContext) Name() string { return bc.name }

// Responsibility returns the context responsibility.
func (bc DomainBoundedContext) Responsibility() string { return bc.responsibility }

// KeyDomainObjects returns a defensive copy of the key domain objects.
func (bc DomainBoundedContext) KeyDomainObjects() []string {
	out := make([]string, len(bc.keyDomainObjects))
	copy(out, bc.keyDomainObjects)
	return out
}

// Classification returns the subdomain classification pointer (nil if unclassified).
func (bc DomainBoundedContext) Classification() *SubdomainClassification { return bc.classification }

// ClassificationRationale returns the rationale for the classification.
func (bc DomainBoundedContext) ClassificationRationale() string { return bc.classificationRationale }

// ContextRelationship is a relationship between two bounded contexts.
type ContextRelationship struct {
	upstream           string
	downstream         string
	integrationPattern string
}

// NewContextRelationship creates a ContextRelationship value object.
func NewContextRelationship(upstream, downstream, integrationPattern string) ContextRelationship {
	return ContextRelationship{
		upstream:           upstream,
		downstream:         downstream,
		integrationPattern: integrationPattern,
	}
}

// Upstream returns the upstream context name.
func (cr ContextRelationship) Upstream() string { return cr.upstream }

// Downstream returns the downstream context name.
func (cr ContextRelationship) Downstream() string { return cr.downstream }

// IntegrationPattern returns the integration pattern.
func (cr ContextRelationship) IntegrationPattern() string { return cr.integrationPattern }

// AggregateDesign is a design for an aggregate within a Core subdomain bounded context.
type AggregateDesign struct {
	name             string
	contextName      string
	rootEntity       string
	containedObjects []string
	invariants       []string
	commands         []string
	domainEvents     []string
}

// NewAggregateDesign creates an AggregateDesign value object.
func NewAggregateDesign(
	name, contextName, rootEntity string,
	containedObjects, invariants, commands, domainEvents []string,
) AggregateDesign {
	co := make([]string, len(containedObjects))
	copy(co, containedObjects)
	inv := make([]string, len(invariants))
	copy(inv, invariants)
	cmd := make([]string, len(commands))
	copy(cmd, commands)
	de := make([]string, len(domainEvents))
	copy(de, domainEvents)
	return AggregateDesign{
		name:             name,
		contextName:      contextName,
		rootEntity:       rootEntity,
		containedObjects: co,
		invariants:       inv,
		commands:         cmd,
		domainEvents:     de,
	}
}

// Name returns the aggregate name.
func (a AggregateDesign) Name() string { return a.name }

// ContextName returns the bounded context name.
func (a AggregateDesign) ContextName() string { return a.contextName }

// RootEntity returns the aggregate root entity name.
func (a AggregateDesign) RootEntity() string { return a.rootEntity }

// ContainedObjects returns a defensive copy of contained objects.
func (a AggregateDesign) ContainedObjects() []string {
	out := make([]string, len(a.containedObjects))
	copy(out, a.containedObjects)
	return out
}

// Invariants returns a defensive copy of invariants.
func (a AggregateDesign) Invariants() []string {
	out := make([]string, len(a.invariants))
	copy(out, a.invariants)
	return out
}

// Commands returns a defensive copy of commands.
func (a AggregateDesign) Commands() []string {
	out := make([]string, len(a.commands))
	copy(out, a.commands)
	return out
}

// DomainEvents returns a defensive copy of domain events.
func (a AggregateDesign) DomainEvents() []string {
	out := make([]string, len(a.domainEvents))
	copy(out, a.domainEvents)
	return out
}
