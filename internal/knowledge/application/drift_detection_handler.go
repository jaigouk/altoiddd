package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/alty-cli/alty/internal/knowledge/domain"
)

// DriftDetector is a handler-local interface for drift detection.
// Defined where consumed per Go convention.
type DriftDetector interface {
	// Detect detects drift across all knowledge entries.
	Detect(ctx context.Context) (domain.DriftReport, error)
}

// DriftDetectionHandler orchestrates drift detection operations.
// It delegates to a DriftDetector for actual detection and filters results by tool.
type DriftDetectionHandler struct {
	detector DriftDetector
}

// NewDriftDetectionHandler creates a new DriftDetectionHandler.
func NewDriftDetectionHandler(detector DriftDetector) *DriftDetectionHandler {
	return &DriftDetectionHandler{detector: detector}
}

// DetectDrift detects drift across knowledge entries, optionally filtering by tool.
// If toolFilter is nil or empty, returns all drift signals.
// If toolFilter is set, returns only signals from that tool's knowledge entries.
func (h *DriftDetectionHandler) DetectDrift(ctx context.Context, toolFilter *string) (domain.DriftReport, error) {
	report, err := h.detector.Detect(ctx)
	if err != nil {
		return domain.DriftReport{}, fmt.Errorf("detect drift: %w", err)
	}

	// No filter or empty filter → return full report
	if toolFilter == nil || *toolFilter == "" {
		return report, nil
	}

	// Filter signals by tool
	var filtered []domain.DriftSignal
	for _, sig := range report.Signals() {
		if h.matchesTool(sig.EntryPath(), *toolFilter) {
			filtered = append(filtered, sig)
		}
	}
	return domain.NewDriftReport(filtered), nil
}

// matchesTool checks if an entry path belongs to a specific tool.
// Entry paths are expected to be in format "tools/<tool-name>/<topic>".
// Non-tool entries (e.g., "ddd/aggregate") return false.
func (h *DriftDetectionHandler) matchesTool(entryPath, tool string) bool {
	segments := strings.Split(entryPath, "/")
	if len(segments) < 2 || segments[0] != "tools" {
		return false // Non-tool entry
	}
	return strings.EqualFold(segments[1], tool)
}
