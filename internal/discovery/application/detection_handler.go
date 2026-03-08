package application

import (
	"github.com/alty-cli/alty/internal/discovery/domain"
)

// ToolDetector detects installed AI coding tools and scans for config conflicts.
// Defined where consumed per Go convention.
type ToolDetector interface {
	Detect(projectDir string) ([]string, error)
	ScanConflicts(projectDir string) ([]string, error)
}

// DetectionHandler orchestrates the detect flow: scan tools, classify conflicts.
type DetectionHandler struct {
	toolDetection ToolDetector
	scanner       *domain.ToolScanner
}

// NewDetectionHandler creates a new DetectionHandler.
func NewDetectionHandler(toolDetection ToolDetector) *DetectionHandler {
	return &DetectionHandler{
		toolDetection: toolDetection,
		scanner:       domain.NewToolScanner(),
	}
}

// Detect detects installed AI coding tools and classifies conflicts.
func (h *DetectionHandler) Detect(projectDir string) (domain.DetectionResult, error) {
	toolNames, err := h.toolDetection.Detect(projectDir)
	if err != nil {
		return domain.DetectionResult{}, err
	}
	conflicts, err := h.toolDetection.ScanConflicts(projectDir)
	if err != nil {
		return domain.DetectionResult{}, err
	}
	return h.scanner.BuildResult(toolNames, conflicts), nil
}
