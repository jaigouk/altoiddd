package application

import (
	"context"
	"fmt"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// BoundaryDetectionHandler is the application-layer entry point for boundary detection.
// Thin delegation to the BoundaryDetector port — exists for SRP and future cross-cutting concerns.
type BoundaryDetectionHandler struct {
	detector BoundaryDetector
}

// NewBoundaryDetectionHandler creates a BoundaryDetectionHandler.
func NewBoundaryDetectionHandler(detector BoundaryDetector) *BoundaryDetectionHandler {
	return &BoundaryDetectionHandler{detector: detector}
}

// Detect runs boundary detection on in-memory stories.
// Returns scored BoundedContextSketch proposals.
func (h *BoundaryDetectionHandler) Detect(
	ctx context.Context,
	stories []*discoverydomain.DomainStory,
	mode discoverydomain.DiscoveryMode,
) ([]discoverydomain.BoundedContextSketch, error) {
	sketches, err := h.detector.DetectBoundaries(ctx, stories, mode)
	if err != nil {
		return nil, fmt.Errorf("detecting boundaries: %w", err)
	}
	return sketches, nil
}
