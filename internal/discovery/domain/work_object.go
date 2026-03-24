package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// WorkObjectType classifies the kind of work object in a domain story.
type WorkObjectType string

// WorkObjectType constants.
const (
	WorkObjectTypeDocument     WorkObjectType = "document"
	WorkObjectTypeFolder       WorkObjectType = "folder"
	WorkObjectTypeCall         WorkObjectType = "call"
	WorkObjectTypeEmail        WorkObjectType = "email"
	WorkObjectTypeConversation WorkObjectType = "conversation"
	WorkObjectTypeInfo         WorkObjectType = "info"
	WorkObjectTypeData         WorkObjectType = "data"
)

var validWorkObjectTypes = map[WorkObjectType]struct{}{
	WorkObjectTypeDocument:     {},
	WorkObjectTypeFolder:       {},
	WorkObjectTypeCall:         {},
	WorkObjectTypeEmail:        {},
	WorkObjectTypeConversation: {},
	WorkObjectTypeInfo:         {},
	WorkObjectTypeData:         {},
}

// NewWorkObjectType creates a WorkObjectType from a string, returning an error if invalid.
func NewWorkObjectType(s string) (WorkObjectType, error) {
	wot := WorkObjectType(s)
	if err := wot.Validate(); err != nil {
		return "", err
	}

	return wot, nil
}

// AllWorkObjectTypes returns all valid WorkObjectType values.
func AllWorkObjectTypes() []WorkObjectType {
	return []WorkObjectType{
		WorkObjectTypeDocument,
		WorkObjectTypeFolder,
		WorkObjectTypeCall,
		WorkObjectTypeEmail,
		WorkObjectTypeConversation,
		WorkObjectTypeInfo,
		WorkObjectTypeData,
	}
}

// String returns the string representation of a WorkObjectType.
func (w WorkObjectType) String() string {
	return string(w)
}

// Validate checks whether the WorkObjectType holds a valid value.
func (w WorkObjectType) Validate() error {
	if _, ok := validWorkObjectTypes[w]; !ok {
		return fmt.Errorf("invalid work object type %q: %w", string(w), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (w WorkObjectType) MarshalText() ([]byte, error) {
	if err := w.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling work object type: %w", err)
	}

	return []byte(w), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (w *WorkObjectType) UnmarshalText(data []byte) error {
	parsed, err := NewWorkObjectType(string(data))
	if err != nil {
		return err
	}

	*w = parsed

	return nil
}

// WorkObject represents a work item or artifact in a domain story.
type WorkObject struct {
	name       string
	objectType WorkObjectType
	trust      vo.TrustLevel
	source     string
}

// NewWorkObject creates a WorkObject, enforcing all domain invariants.
func NewWorkObject(name string, objectType WorkObjectType, trust vo.TrustLevel, source string) (WorkObject, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return WorkObject{}, fmt.Errorf("work object name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := objectType.Validate(); err != nil {
		return WorkObject{}, fmt.Errorf("validating work object type: %w", err)
	}

	if err := trust.Validate(); err != nil {
		return WorkObject{}, fmt.Errorf("validating trust level: %w", err)
	}

	source = strings.TrimSpace(source)
	if trust == vo.AIResearched && source == "" {
		return WorkObject{}, fmt.Errorf("source required when trust is ai_researched: %w", domainerrors.ErrInvariantViolation)
	}

	return WorkObject{
		name:       name,
		objectType: objectType,
		trust:      trust,
		source:     source,
	}, nil
}

// Name returns the work object's name.
func (w WorkObject) Name() string { return w.name }

// Type returns the work object's type.
func (w WorkObject) Type() WorkObjectType { return w.objectType }

// Trust returns the work object's trust level.
func (w WorkObject) Trust() vo.TrustLevel { return w.trust }

// Source returns the work object's source reference.
func (w WorkObject) Source() string { return w.source }

// Equals compares two work objects by name (case-insensitive).
func (w WorkObject) Equals(other WorkObject) bool {
	return strings.EqualFold(w.name, other.name)
}

// String returns a human-readable representation of the work object.
func (w WorkObject) String() string {
	return fmt.Sprintf("WorkObject: %s (%s, %s)", w.name, w.objectType, w.trust)
}
