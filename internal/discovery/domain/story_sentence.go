package domain

import (
	"fmt"
	"regexp"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// validPrepositions lists the allowed preposition values for a StorySentence.
var validPrepositions = map[string]struct{}{
	"for":      {},
	"to":       {},
	"via":      {},
	"using":    {},
	"from":     {},
	"with":     {},
	"in":       {},
	"about":    {},
	"based on": {},
	"on":       {},
}

// branchingPattern matches branching keywords bounded by whitespace (or string edges),
// not by regex \b (which treats hyphens as boundaries).
var branchingPattern = regexp.MustCompile(`(?i)(^|\s)(sometimes|if|optionally|alternatively|when|unless)(\s|$)`)

// StorySentence represents a single step in a domain story.
type StorySentence struct {
	step           int
	subject        string
	activity       string
	object         string
	preposition    string
	indirectObject string
	trust          vo.TrustLevel
	source         string
}

// NewStorySentence creates a StorySentence, enforcing all domain invariants.
func NewStorySentence(
	step int,
	subject, activity, object string,
	trust vo.TrustLevel,
	source string,
) (StorySentence, error) {
	if step < 1 {
		return StorySentence{}, fmt.Errorf("step must be positive, got %d: %w", step, domainerrors.ErrInvariantViolation)
	}

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return StorySentence{}, fmt.Errorf("subject must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	activity = strings.TrimSpace(activity)
	if activity == "" {
		return StorySentence{}, fmt.Errorf("activity must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	object = strings.TrimSpace(object)
	if object == "" {
		return StorySentence{}, fmt.Errorf("object must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := trust.Validate(); err != nil {
		return StorySentence{}, fmt.Errorf("validating trust level: %w", err)
	}

	source = strings.TrimSpace(source)
	if trust == vo.AIResearched && source == "" {
		return StorySentence{}, fmt.Errorf("source required when trust is ai_researched: %w", domainerrors.ErrInvariantViolation)
	}

	return StorySentence{
		step:     step,
		subject:  subject,
		activity: activity,
		object:   object,
		trust:    trust,
		source:   source,
	}, nil
}

// WithPreposition returns a new StorySentence with the preposition and indirect object set.
func (s StorySentence) WithPreposition(preposition, indirectObject string) (StorySentence, error) {
	preposition = strings.ToLower(strings.TrimSpace(preposition))
	if _, ok := validPrepositions[preposition]; !ok {
		return StorySentence{}, fmt.Errorf("invalid preposition %q: %w", preposition, domainerrors.ErrInvariantViolation)
	}

	indirectObject = strings.TrimSpace(indirectObject)
	if indirectObject == "" {
		return StorySentence{}, fmt.Errorf("indirect object must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	return StorySentence{
		step:           s.step,
		subject:        s.subject,
		activity:       s.activity,
		object:         s.object,
		preposition:    preposition,
		indirectObject: indirectObject,
		trust:          s.trust,
		source:         s.source,
	}, nil
}

// Step returns the sentence's step number.
func (s StorySentence) Step() int { return s.step }

// Subject returns the actor name performing the activity.
func (s StorySentence) Subject() string { return s.subject }

// Activity returns the verb phrase describing what the subject does.
func (s StorySentence) Activity() string { return s.activity }

// Object returns the work object name being acted upon.
func (s StorySentence) Object() string { return s.object }

// Preposition returns the preposition linking the object to the indirect object.
func (s StorySentence) Preposition() string { return s.preposition }

// IndirectObject returns the target of the preposition.
func (s StorySentence) IndirectObject() string { return s.indirectObject }

// Trust returns the sentence's trust level.
func (s StorySentence) Trust() vo.TrustLevel { return s.trust }

// Source returns the sentence's source reference.
func (s StorySentence) Source() string { return s.source }

// HasIndirectObject returns true when both preposition and indirect object are set.
func (s StorySentence) HasIndirectObject() bool {
	return s.preposition != "" && s.indirectObject != ""
}

// FormatText returns the human-readable sentence format.
func (s StorySentence) FormatText() string {
	base := fmt.Sprintf("%d. %s %s %s", s.step, s.subject, s.activity, s.object)
	if s.HasIndirectObject() {
		return fmt.Sprintf("%s %s %s", base, s.preposition, s.indirectObject)
	}

	return base
}

// ContainsBranching returns true if the activity contains branching keywords
// bounded by whitespace (not substrings within hyphenated words).
func (s StorySentence) ContainsBranching() bool {
	return branchingPattern.MatchString(s.activity)
}

// String returns the formatted text representation of the sentence.
func (s StorySentence) String() string {
	return s.FormatText()
}
