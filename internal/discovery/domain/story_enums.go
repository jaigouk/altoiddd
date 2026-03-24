package domain

import (
	"fmt"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// StoryType classifies the granularity of a domain story.
type StoryType string

// StoryType constants.
const (
	StoryTypeCoarseGrained StoryType = "coarse_grained"
	StoryTypeFineGrained   StoryType = "fine_grained"
)

var validStoryTypes = map[StoryType]struct{}{
	StoryTypeCoarseGrained: {},
	StoryTypeFineGrained:   {},
}

// NewStoryType creates a StoryType from a string, returning an error if invalid.
func NewStoryType(s string) (StoryType, error) {
	st := StoryType(s)
	if err := st.Validate(); err != nil {
		return "", err
	}

	return st, nil
}

// AllStoryTypes returns all valid StoryType values.
func AllStoryTypes() []StoryType {
	return []StoryType{StoryTypeCoarseGrained, StoryTypeFineGrained}
}

// String returns the string representation of a StoryType.
func (s StoryType) String() string {
	return string(s)
}

// Validate checks whether the StoryType holds a valid value.
func (s StoryType) Validate() error {
	if _, ok := validStoryTypes[s]; !ok {
		return fmt.Errorf("invalid story type %q: %w", string(s), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (s StoryType) MarshalText() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling story type: %w", err)
	}

	return []byte(s), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *StoryType) UnmarshalText(data []byte) error {
	parsed, err := NewStoryType(string(data))
	if err != nil {
		return err
	}

	*s = parsed

	return nil
}

// TimeType classifies the temporal scope of a domain story.
type TimeType string

// TimeType constants.
const (
	TimeTypeAsIs TimeType = "as_is"
	TimeTypeToBe TimeType = "to_be"
)

var validTimeTypes = map[TimeType]struct{}{
	TimeTypeAsIs: {},
	TimeTypeToBe: {},
}

// NewTimeType creates a TimeType from a string, returning an error if invalid.
func NewTimeType(s string) (TimeType, error) {
	tt := TimeType(s)
	if err := tt.Validate(); err != nil {
		return "", err
	}

	return tt, nil
}

// AllTimeTypes returns all valid TimeType values.
func AllTimeTypes() []TimeType {
	return []TimeType{TimeTypeAsIs, TimeTypeToBe}
}

// String returns the string representation of a TimeType.
func (t TimeType) String() string {
	return string(t)
}

// Validate checks whether the TimeType holds a valid value.
func (t TimeType) Validate() error {
	if _, ok := validTimeTypes[t]; !ok {
		return fmt.Errorf("invalid time type %q: %w", string(t), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (t TimeType) MarshalText() ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling time type: %w", err)
	}

	return []byte(t), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *TimeType) UnmarshalText(data []byte) error {
	parsed, err := NewTimeType(string(data))
	if err != nil {
		return err
	}

	*t = parsed

	return nil
}

// PurityType classifies the domain purity of a domain story.
type PurityType string

// PurityType constants.
const (
	PurityTypePure        PurityType = "pure"
	PurityTypeDigitalized PurityType = "digitalized"
)

var validPurityTypes = map[PurityType]struct{}{
	PurityTypePure:        {},
	PurityTypeDigitalized: {},
}

// NewPurityType creates a PurityType from a string, returning an error if invalid.
func NewPurityType(s string) (PurityType, error) {
	pt := PurityType(s)
	if err := pt.Validate(); err != nil {
		return "", err
	}

	return pt, nil
}

// AllPurityTypes returns all valid PurityType values.
func AllPurityTypes() []PurityType {
	return []PurityType{PurityTypePure, PurityTypeDigitalized}
}

// String returns the string representation of a PurityType.
func (p PurityType) String() string {
	return string(p)
}

// Validate checks whether the PurityType holds a valid value.
func (p PurityType) Validate() error {
	if _, ok := validPurityTypes[p]; !ok {
		return fmt.Errorf("invalid purity type %q: %w", string(p), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (p PurityType) MarshalText() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling purity type: %w", err)
	}

	return []byte(p), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *PurityType) UnmarshalText(data []byte) error {
	parsed, err := NewPurityType(string(data))
	if err != nil {
		return err
	}

	*p = parsed

	return nil
}
