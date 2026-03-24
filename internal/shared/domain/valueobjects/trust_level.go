// Package valueobjects provides shared value objects for the alto domain.
package valueobjects

import (
	"fmt"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// TrustLevel classifies how much a piece of knowledge can be trusted.
// Lower numeric value = higher trust.
type TrustLevel int

// Trust level constants ordered from most to least trusted.
const (
	UserStated    TrustLevel = 1
	UserConfirmed TrustLevel = 2
	AIResearched  TrustLevel = 3
	AIInferred    TrustLevel = 4
)

// trustLevelStrings maps TrustLevel values to their string representations.
var trustLevelStrings = map[TrustLevel]string{
	UserStated:    "user_stated",
	UserConfirmed: "user_confirmed",
	AIResearched:  "ai_researched",
	AIInferred:    "ai_inferred",
}

// trustLevelFromString maps string representations to TrustLevel values.
var trustLevelFromString = map[string]TrustLevel{
	"user_stated":    UserStated,
	"user_confirmed": UserConfirmed,
	"ai_researched":  AIResearched,
	"ai_inferred":    AIInferred,
}

// NewTrustLevel creates a TrustLevel from an integer value.
func NewTrustLevel(value int) (TrustLevel, error) {
	tl := TrustLevel(value)
	if err := tl.Validate(); err != nil {
		return 0, err
	}

	return tl, nil
}

// AllTrustLevels returns all valid TrustLevel values.
func AllTrustLevels() []TrustLevel {
	return []TrustLevel{UserStated, UserConfirmed, AIResearched, AIInferred}
}

// ParseTrustLevel parses a string into a TrustLevel.
func ParseTrustLevel(s string) (TrustLevel, error) {
	tl, ok := trustLevelFromString[s]
	if !ok {
		return 0, fmt.Errorf("invalid trust level string %q: %w", s, domainerrors.ErrInvariantViolation)
	}

	return tl, nil
}

// String returns the string representation of a TrustLevel.
func (t TrustLevel) String() string {
	if s, ok := trustLevelStrings[t]; ok {
		return s
	}

	return fmt.Sprintf("TrustLevel(%d)", int(t))
}

// Validate checks whether the TrustLevel holds a valid value.
func (t TrustLevel) Validate() error {
	if _, ok := trustLevelStrings[t]; !ok {
		return fmt.Errorf("invalid trust level value %d: %w", int(t), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// IsHigherTrust returns true if t represents higher trust than other.
// Lower numeric value = higher trust.
func (t TrustLevel) IsHigherTrust(other TrustLevel) bool {
	return t < other
}

// MarshalText implements encoding.TextMarshaler.
func (t TrustLevel) MarshalText() ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling trust level: %w", err)
	}

	return []byte(t.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *TrustLevel) UnmarshalText(data []byte) error {
	parsed, err := ParseTrustLevel(string(data))
	if err != nil {
		return err
	}

	*t = parsed

	return nil
}
