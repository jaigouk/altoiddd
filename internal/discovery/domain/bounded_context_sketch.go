package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ConfidenceLevel classifies how confident a bounded context sketch is.
type ConfidenceLevel string

// ConfidenceLevel constants.
const (
	ConfidenceLevelHigh   ConfidenceLevel = "HIGH"
	ConfidenceLevelMedium ConfidenceLevel = "MEDIUM"
	ConfidenceLevelLow    ConfidenceLevel = "LOW"
)

// BoundedContextSketch represents an early-stage hypothesis about a bounded context
// discovered through domain storytelling analysis.
type BoundedContextSketch struct {
	name           string
	classification vo.SubdomainClassification
	confidence     float64
	actors         []string
	workObjects    []string
	stories        []string
	signals        []BoundarySignal
	trust          vo.TrustLevel
}

// NewBoundedContextSketch creates a BoundedContextSketch, enforcing all domain invariants.
func NewBoundedContextSketch(
	name string,
	classification vo.SubdomainClassification,
	confidence float64,
	actors []string,
	workObjects []string,
	stories []string,
	signals []BoundarySignal,
	trust vo.TrustLevel,
) (BoundedContextSketch, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return BoundedContextSketch{}, fmt.Errorf("bounded context sketch name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if !isValidClassification(classification) {
		return BoundedContextSketch{}, fmt.Errorf("invalid subdomain classification %q: %w", string(classification), domainerrors.ErrInvariantViolation)
	}

	if confidence < 0 || confidence > 1 {
		return BoundedContextSketch{}, fmt.Errorf("confidence must be between 0 and 1, got %f: %w", confidence, domainerrors.ErrInvariantViolation)
	}

	if err := trust.Validate(); err != nil {
		return BoundedContextSketch{}, fmt.Errorf("validating trust level: %w", err)
	}

	// Defensive copies of all slices.
	actorsCopy := make([]string, len(actors))
	copy(actorsCopy, actors)

	workObjectsCopy := make([]string, len(workObjects))
	copy(workObjectsCopy, workObjects)

	storiesCopy := make([]string, len(stories))
	copy(storiesCopy, stories)

	signalsCopy := make([]BoundarySignal, len(signals))
	copy(signalsCopy, signals)

	return BoundedContextSketch{
		name:           name,
		classification: classification,
		confidence:     confidence,
		actors:         actorsCopy,
		workObjects:    workObjectsCopy,
		stories:        storiesCopy,
		signals:        signalsCopy,
		trust:          trust,
	}, nil
}

// isValidClassification checks if a SubdomainClassification is one of the known values.
func isValidClassification(c vo.SubdomainClassification) bool {
	for _, valid := range vo.AllSubdomainClassifications() {
		if c == valid {
			return true
		}
	}

	return false
}

// Name returns the sketch's name.
func (s BoundedContextSketch) Name() string { return s.name }

// Classification returns the sketch's subdomain classification.
func (s BoundedContextSketch) Classification() vo.SubdomainClassification { return s.classification }

// Confidence returns the sketch's confidence score (0.0-1.0).
func (s BoundedContextSketch) Confidence() float64 { return s.confidence }

// Actors returns a defensive copy of the sketch's actors.
func (s BoundedContextSketch) Actors() []string {
	out := make([]string, len(s.actors))
	copy(out, s.actors)

	return out
}

// WorkObjects returns a defensive copy of the sketch's work objects.
func (s BoundedContextSketch) WorkObjects() []string {
	out := make([]string, len(s.workObjects))
	copy(out, s.workObjects)

	return out
}

// Stories returns a defensive copy of the sketch's story references.
func (s BoundedContextSketch) Stories() []string {
	out := make([]string, len(s.stories))
	copy(out, s.stories)

	return out
}

// Signals returns a defensive copy of the sketch's boundary signals.
func (s BoundedContextSketch) Signals() []BoundarySignal {
	out := make([]BoundarySignal, len(s.signals))
	copy(out, s.signals)

	return out
}

// Trust returns the sketch's trust level.
func (s BoundedContextSketch) Trust() vo.TrustLevel { return s.trust }

// ConfidenceLevel returns the computed confidence level based on confidence score
// and number of distinct signal types.
//
// Thresholds (from spike Section 10):
//
//	HIGH:   confidence >= 0.65 OR 2+ distinct signal types
//	MEDIUM: confidence >= 0.45 AND < 0.65 with < 2 distinct signal types
//	LOW:    confidence < 0.45 with < 2 distinct signal types
func (s BoundedContextSketch) ConfidenceLevel() ConfidenceLevel {
	distinctTypes := s.distinctSignalTypeCount()

	if s.confidence >= 0.65 || distinctTypes >= 2 {
		return ConfidenceLevelHigh
	}

	if s.confidence >= 0.45 {
		return ConfidenceLevelMedium
	}

	return ConfidenceLevelLow
}

// distinctSignalTypeCount returns the number of unique signal types in the sketch.
func (s BoundedContextSketch) distinctSignalTypeCount() int {
	seen := make(map[SignalType]struct{}, len(s.signals))
	for _, sig := range s.signals {
		seen[sig.Type()] = struct{}{}
	}

	return len(seen)
}

// String returns a human-readable representation of the bounded context sketch.
func (s BoundedContextSketch) String() string {
	return fmt.Sprintf("BoundedContextSketch: %s (%s, confidence=%.2f, level=%s, %s)",
		s.name, s.classification, s.confidence, s.ConfidenceLevel(), s.trust)
}
