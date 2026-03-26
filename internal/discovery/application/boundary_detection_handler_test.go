package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Fake boundary detector
// ---------------------------------------------------------------------------

type fakeBoundaryDetector struct {
	sketches []domain.BoundedContextSketch
	err      error
}

func (f *fakeBoundaryDetector) DetectBoundaries(
	_ context.Context,
	_ []*domain.DomainStory,
	_ domain.DiscoveryMode,
) ([]domain.BoundedContextSketch, error) {
	return f.sketches, f.err
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestBoundaryDetectionHandler_Detect_HappyPath(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.8, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	fake := &fakeBoundaryDetector{sketches: []domain.BoundedContextSketch{sketch}}
	handler := application.NewBoundaryDetectionHandler(fake)

	result, err := handler.Detect(context.TODO(), []*domain.DomainStory{}, domain.ModeRapid)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Ordering", result[0].Name())
}

func TestBoundaryDetectionHandler_Detect_DetectorError(t *testing.T) {
	t.Parallel()
	fake := &fakeBoundaryDetector{err: fmt.Errorf("detector failure")}
	handler := application.NewBoundaryDetectionHandler(fake)

	_, err := handler.Detect(context.TODO(), []*domain.DomainStory{}, domain.ModeRapid)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "detecting boundaries")
	assert.Contains(t, err.Error(), "detector failure")
}

func TestBoundaryDetectionHandler_Detect_EmptyStories(t *testing.T) {
	t.Parallel()
	fake := &fakeBoundaryDetector{sketches: []domain.BoundedContextSketch{}}
	handler := application.NewBoundaryDetectionHandler(fake)

	result, err := handler.Detect(context.TODO(), []*domain.DomainStory{}, domain.ModeRapid)

	require.NoError(t, err)
	assert.Empty(t, result)
}
