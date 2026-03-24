package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ActorType classifies the kind of actor in a domain story.
type ActorType string

// ActorType constants.
const (
	ActorTypePerson ActorType = "person"
	ActorTypeSystem ActorType = "system"
	ActorTypeGroup  ActorType = "group"
)

var validActorTypes = map[ActorType]struct{}{
	ActorTypePerson: {},
	ActorTypeSystem: {},
	ActorTypeGroup:  {},
}

// NewActorType creates an ActorType from a string, returning an error if invalid.
func NewActorType(s string) (ActorType, error) {
	at := ActorType(s)
	if err := at.Validate(); err != nil {
		return "", err
	}

	return at, nil
}

// AllActorTypes returns all valid ActorType values.
func AllActorTypes() []ActorType {
	return []ActorType{ActorTypePerson, ActorTypeSystem, ActorTypeGroup}
}

// String returns the string representation of an ActorType.
func (a ActorType) String() string {
	return string(a)
}

// Validate checks whether the ActorType holds a valid value.
func (a ActorType) Validate() error {
	if _, ok := validActorTypes[a]; !ok {
		return fmt.Errorf("invalid actor type %q: %w", string(a), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (a ActorType) MarshalText() ([]byte, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling actor type: %w", err)
	}

	return []byte(a), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (a *ActorType) UnmarshalText(data []byte) error {
	parsed, err := NewActorType(string(data))
	if err != nil {
		return err
	}

	*a = parsed

	return nil
}

// StoryActor represents a participant in a domain story.
type StoryActor struct {
	name      string
	actorType ActorType
	trust     vo.TrustLevel
	source    string
}

// NewStoryActor creates a StoryActor, enforcing all domain invariants.
func NewStoryActor(name string, actorType ActorType, trust vo.TrustLevel, source string) (StoryActor, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return StoryActor{}, fmt.Errorf("actor name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := actorType.Validate(); err != nil {
		return StoryActor{}, fmt.Errorf("validating actor type: %w", err)
	}

	if err := trust.Validate(); err != nil {
		return StoryActor{}, fmt.Errorf("validating trust level: %w", err)
	}

	source = strings.TrimSpace(source)
	if trust == vo.AIResearched && source == "" {
		return StoryActor{}, fmt.Errorf("source required when trust is ai_researched: %w", domainerrors.ErrInvariantViolation)
	}

	return StoryActor{
		name:      name,
		actorType: actorType,
		trust:     trust,
		source:    source,
	}, nil
}

// Name returns the actor's name.
func (a StoryActor) Name() string { return a.name }

// Type returns the actor's type.
func (a StoryActor) Type() ActorType { return a.actorType }

// Trust returns the actor's trust level.
func (a StoryActor) Trust() vo.TrustLevel { return a.trust }

// Source returns the actor's source reference.
func (a StoryActor) Source() string { return a.source }

// Equals compares two actors by name (case-insensitive).
func (a StoryActor) Equals(other StoryActor) bool {
	return strings.EqualFold(a.name, other.name)
}

// String returns a human-readable representation of the actor.
func (a StoryActor) String() string {
	return fmt.Sprintf("Actor: %s (%s, %s)", a.name, a.actorType, a.trust)
}
