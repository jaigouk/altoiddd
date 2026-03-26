package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// SignalType classifies the kind of boundary signal observed between bounded contexts.
type SignalType string

// SignalType constants.
//
// Spike-validated signals (20260323_5_boundary_detection_heuristics_validation.md Section 9):
//   - same_object_diff_context: spike-validated, P=0.85
//   - one_way_flow: spike-validated, P=0.70
//   - org_boundary: spike-validated, P=0.67
//   - different_trigger: LLM-enhanced (spike Section 9)
//   - language_difference: LLM-only (spike Section 9)
//
// Methodology-derived signals (Hofer & Schwentner):
//   - different_lifecycle, external_system, different_actor, complex_rules
const (
	SignalTypeDifferentTrigger      SignalType = "different_trigger"
	SignalTypeOneWayFlow            SignalType = "one_way_flow"
	SignalTypeLanguageDifference    SignalType = "language_difference"
	SignalTypeDifferentLifecycle    SignalType = "different_lifecycle"
	SignalTypeExternalSystem        SignalType = "external_system"
	SignalTypeDifferentActor        SignalType = "different_actor"
	SignalTypeComplexRules          SignalType = "complex_rules"
	SignalTypeSameObjectDiffContext SignalType = "same_object_diff_context"
	SignalTypeOrgBoundary           SignalType = "org_boundary"
)

var validSignalTypes = map[SignalType]struct{}{
	SignalTypeDifferentTrigger:      {},
	SignalTypeOneWayFlow:            {},
	SignalTypeLanguageDifference:    {},
	SignalTypeDifferentLifecycle:    {},
	SignalTypeExternalSystem:        {},
	SignalTypeDifferentActor:        {},
	SignalTypeComplexRules:          {},
	SignalTypeSameObjectDiffContext: {},
	SignalTypeOrgBoundary:           {},
}

// NewSignalType creates a SignalType from a string, returning an error if invalid.
func NewSignalType(s string) (SignalType, error) {
	st := SignalType(s)
	if err := st.Validate(); err != nil {
		return "", err
	}

	return st, nil
}

// AllSignalTypes returns all valid SignalType values.
func AllSignalTypes() []SignalType {
	return []SignalType{
		SignalTypeDifferentTrigger,
		SignalTypeOneWayFlow,
		SignalTypeLanguageDifference,
		SignalTypeDifferentLifecycle,
		SignalTypeExternalSystem,
		SignalTypeDifferentActor,
		SignalTypeComplexRules,
		SignalTypeSameObjectDiffContext,
		SignalTypeOrgBoundary,
	}
}

// String returns the string representation of a SignalType.
func (s SignalType) String() string {
	return string(s)
}

// Validate checks whether the SignalType holds a valid value.
func (s SignalType) Validate() error {
	if _, ok := validSignalTypes[s]; !ok {
		return fmt.Errorf("invalid signal type %q: %w", string(s), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (s SignalType) MarshalText() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling signal type: %w", err)
	}

	return []byte(s), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *SignalType) UnmarshalText(data []byte) error {
	parsed, err := NewSignalType(string(data))
	if err != nil {
		return err
	}

	*s = parsed

	return nil
}

// BoundarySignal represents an observed indicator that a bounded context boundary exists.
type BoundarySignal struct {
	signalType  SignalType
	description string
}

// NewBoundarySignal creates a BoundarySignal, enforcing all domain invariants.
func NewBoundarySignal(signalType SignalType, description string) (BoundarySignal, error) {
	if err := signalType.Validate(); err != nil {
		return BoundarySignal{}, fmt.Errorf("validating signal type: %w", err)
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return BoundarySignal{}, fmt.Errorf("boundary signal description must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	return BoundarySignal{
		signalType:  signalType,
		description: description,
	}, nil
}

// Type returns the signal's type.
func (b BoundarySignal) Type() SignalType { return b.signalType }

// Description returns the signal's description.
func (b BoundarySignal) Description() string { return b.description }

// String returns a human-readable representation of the boundary signal.
func (b BoundarySignal) String() string {
	return fmt.Sprintf("BoundarySignal: [%s] %s", b.signalType, b.description)
}
