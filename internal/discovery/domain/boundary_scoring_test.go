package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
)

func TestComputeBoundaryScore_ZeroInputs(t *testing.T) {
	t.Parallel()
	got := domain.ComputeBoundaryScore(0, 0, 0)
	assert.InDelta(t, 0.0, got, 0.001)
}

func TestComputeBoundaryScore_SingleSignal(t *testing.T) {
	t.Parallel()
	// 0.40 + 0.15*1 + 0.10*(2/3.0) = 0.40 + 0.15 + 0.0667 = 0.6167
	got := domain.ComputeBoundaryScore(0.40, 1, 2)
	assert.InDelta(t, 0.6167, got, 0.001)
}

func TestComputeBoundaryScore_MultipleTypes(t *testing.T) {
	t.Parallel()
	// 0.30 + 0.15*3 + 0.10*(3/3.0) = 0.30 + 0.45 + 0.10 = 0.85
	got := domain.ComputeBoundaryScore(0.30, 3, 3)
	assert.InDelta(t, 0.85, got, 0.001)
}

func TestComputeBoundaryScore_HighConfidence(t *testing.T) {
	t.Parallel()
	// 0.40 + 0.15*3 + 0.10*(3/3.0) = 0.40 + 0.45 + 0.10 = 0.95
	got := domain.ComputeBoundaryScore(0.40, 3, 3)
	assert.InDelta(t, 0.95, got, 0.001)
}

func TestComputeBoundaryScore_LowConfidence(t *testing.T) {
	t.Parallel()
	// 0.20 + 0.15*1 + 0.10*(2/3.0) = 0.20 + 0.15 + 0.0667 = 0.4167
	got := domain.ComputeBoundaryScore(0.20, 1, 2)
	assert.InDelta(t, 0.4167, got, 0.001)
}

func TestComputeBoundaryScore_StoryBonusScaling(t *testing.T) {
	t.Parallel()
	// 0.30 + 0.15*2 + 0.10*(6/3.0) = 0.30 + 0.30 + 0.20 = 0.80
	got := domain.ComputeBoundaryScore(0.30, 2, 6)
	assert.InDelta(t, 0.80, got, 0.001)
}

func TestComputeBoundaryScore_ZeroSignalAvgWithTypes(t *testing.T) {
	t.Parallel()
	// 0.0 + 0.15*2 + 0.10*(3/3.0) = 0.0 + 0.30 + 0.10 = 0.40
	got := domain.ComputeBoundaryScore(0.0, 2, 3)
	assert.InDelta(t, 0.40, got, 0.001)
}

func TestComputeBoundaryScore_AlwaysNonNegative(t *testing.T) {
	t.Parallel()
	// 0.0 + 0.15*0 + 0.10*(1/3.0) = 0.0 + 0.0 + 0.0333 = 0.0333
	got := domain.ComputeBoundaryScore(0.0, 0, 1)
	assert.InDelta(t, 0.0333, got, 0.001)
	assert.GreaterOrEqual(t, got, 0.0)
}
