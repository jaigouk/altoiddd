package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// DriftSignalType enumerates the kinds of drift that can be detected.
type DriftSignalType string

// Drift signal type constants.
const (
	DriftVersionChange   DriftSignalType = "version_change"
	DriftDocCodeMismatch DriftSignalType = "doc_code_mismatch"
	DriftStale           DriftSignalType = "stale"
)

// DriftSeverity classifies how urgent a drift signal is.
type DriftSeverity string

// Drift severity constants.
const (
	SeverityInfo    DriftSeverity = "info"
	SeverityWarning DriftSeverity = "warning"
	SeverityError   DriftSeverity = "error"
)

// DriftSignal is a single detected drift in a knowledge entry.
type DriftSignal struct {
	entryPath   string
	signalType  DriftSignalType
	description string
	severity    DriftSeverity
}

// NewDriftSignal creates a validated DriftSignal.
func NewDriftSignal(entryPath string, signalType DriftSignalType, description string, severity DriftSeverity) (DriftSignal, error) {
	if strings.TrimSpace(entryPath) == "" {
		return DriftSignal{}, fmt.Errorf("driftSignal entry_path must not be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(description) == "" {
		return DriftSignal{}, fmt.Errorf("driftSignal description must not be empty: %w",
			domainerrors.ErrInvariantViolation)
	}
	return DriftSignal{
		entryPath:   entryPath,
		signalType:  signalType,
		description: description,
		severity:    severity,
	}, nil
}

// EntryPath returns the knowledge entry path that drifted.
func (s DriftSignal) EntryPath() string { return s.entryPath }

// SignalType returns the type of drift detected.
func (s DriftSignal) SignalType() DriftSignalType { return s.signalType }

// Description returns a human-readable description of the drift.
func (s DriftSignal) Description() string { return s.description }

// Severity returns the drift severity level.
func (s DriftSignal) Severity() DriftSeverity { return s.severity }

// DriftReport aggregates drift signals across the knowledge base.
type DriftReport struct {
	signals []DriftSignal
}

// NewDriftReport creates a DriftReport.
func NewDriftReport(signals []DriftSignal) DriftReport {
	s := make([]DriftSignal, len(signals))
	copy(s, signals)
	return DriftReport{signals: s}
}

// Signals returns a defensive copy.
func (r DriftReport) Signals() []DriftSignal {
	out := make([]DriftSignal, len(r.signals))
	copy(out, r.signals)
	return out
}

// TotalCount returns the total number of drift signals.
func (r DriftReport) TotalCount() int { return len(r.signals) }

// HasDrift returns whether any drift was detected.
func (r DriftReport) HasDrift() bool { return r.TotalCount() > 0 }

// CountBySeverity counts signals with a specific severity.
func (r DriftReport) CountBySeverity(severity DriftSeverity) int {
	count := 0
	for _, s := range r.signals {
		if s.severity == severity {
			count++
		}
	}
	return count
}

// CountByType counts signals with a specific type.
func (r DriftReport) CountByType(signalType DriftSignalType) int {
	count := 0
	for _, s := range r.signals {
		if s.signalType == signalType {
			count++
		}
	}
	return count
}
