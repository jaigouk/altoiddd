package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// AnnotationType classifies the kind of annotation in a domain story.
type AnnotationType string

// AnnotationType constants.
const (
	AnnotationTypeConstraint AnnotationType = "constraint"
	AnnotationTypeInvariant  AnnotationType = "invariant"
	AnnotationTypeAssumption AnnotationType = "assumption"
)

var validAnnotationTypes = map[AnnotationType]struct{}{
	AnnotationTypeConstraint: {},
	AnnotationTypeInvariant:  {},
	AnnotationTypeAssumption: {},
}

// NewAnnotationType creates an AnnotationType from a string, returning an error if invalid.
func NewAnnotationType(s string) (AnnotationType, error) {
	at := AnnotationType(s)
	if err := at.Validate(); err != nil {
		return "", err
	}

	return at, nil
}

// AllAnnotationTypes returns all valid AnnotationType values.
func AllAnnotationTypes() []AnnotationType {
	return []AnnotationType{AnnotationTypeConstraint, AnnotationTypeInvariant, AnnotationTypeAssumption}
}

// String returns the string representation of an AnnotationType.
func (a AnnotationType) String() string {
	return string(a)
}

// Validate checks whether the AnnotationType holds a valid value.
func (a AnnotationType) Validate() error {
	if _, ok := validAnnotationTypes[a]; !ok {
		return fmt.Errorf("invalid annotation type %q: %w", string(a), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (a AnnotationType) MarshalText() ([]byte, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling annotation type: %w", err)
	}

	return []byte(a), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (a *AnnotationType) UnmarshalText(data []byte) error {
	parsed, err := NewAnnotationType(string(data))
	if err != nil {
		return err
	}

	*a = parsed

	return nil
}

// Annotation represents a constraint, invariant, or assumption attached to a domain story.
type Annotation struct {
	text           string
	annotationType AnnotationType
	sentenceRef    *int
	trust          vo.TrustLevel
	source         string
}

// NewAnnotation creates an Annotation, enforcing all domain invariants.
func NewAnnotation(text string, annotationType AnnotationType, sentenceRef *int, trust vo.TrustLevel, source string) (Annotation, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Annotation{}, fmt.Errorf("annotation text must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := annotationType.Validate(); err != nil {
		return Annotation{}, fmt.Errorf("validating annotation type: %w", err)
	}

	if sentenceRef != nil && *sentenceRef <= 0 {
		return Annotation{}, fmt.Errorf("sentence ref must be >= 1, got %d: %w", *sentenceRef, domainerrors.ErrInvariantViolation)
	}

	if err := trust.Validate(); err != nil {
		return Annotation{}, fmt.Errorf("validating trust level: %w", err)
	}

	source = strings.TrimSpace(source)
	if trust == vo.AIResearched && source == "" {
		return Annotation{}, fmt.Errorf("source required when trust is ai_researched: %w", domainerrors.ErrInvariantViolation)
	}

	// Defensive copy of sentenceRef to prevent external mutation.
	var ref *int
	if sentenceRef != nil {
		v := *sentenceRef
		ref = &v
	}

	return Annotation{
		text:           text,
		annotationType: annotationType,
		sentenceRef:    ref,
		trust:          trust,
		source:         source,
	}, nil
}

// Text returns the annotation's text.
func (a Annotation) Text() string { return a.text }

// Type returns the annotation's type.
func (a Annotation) Type() AnnotationType { return a.annotationType }

// SentenceRef returns the sentence reference, or nil if story-wide.
func (a Annotation) SentenceRef() *int {
	if a.sentenceRef == nil {
		return nil
	}

	v := *a.sentenceRef

	return &v
}

// Trust returns the annotation's trust level.
func (a Annotation) Trust() vo.TrustLevel { return a.trust }

// Source returns the annotation's source reference.
func (a Annotation) Source() string { return a.source }

// IsStoryWide returns true when the annotation applies to the entire story.
func (a Annotation) IsStoryWide() bool { return a.sentenceRef == nil }

// String returns a human-readable representation of the annotation.
func (a Annotation) String() string {
	scope := "story-wide"
	if a.sentenceRef != nil {
		scope = fmt.Sprintf("sentence %d", *a.sentenceRef)
	}

	return fmt.Sprintf("Annotation: [%s] %s (%s, %s)", a.annotationType, a.text, scope, a.trust)
}
