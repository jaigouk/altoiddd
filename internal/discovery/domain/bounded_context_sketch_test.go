package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- ConfidenceLevel tests --

func TestConfidenceLevel_Constants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, domain.ConfidenceLevelHigh, domain.ConfidenceLevel("HIGH"))
	assert.Equal(t, domain.ConfidenceLevelMedium, domain.ConfidenceLevel("MEDIUM"))
	assert.Equal(t, domain.ConfidenceLevelLow, domain.ConfidenceLevel("LOW"))
}

// -- BoundedContextSketch tests --

func helperSignal(t *testing.T, st domain.SignalType) domain.BoundarySignal {
	t.Helper()

	bs, err := domain.NewBoundarySignal(st, "test signal")
	require.NoError(t, err)

	return bs
}

func TestNewBoundedContextSketch_Valid(t *testing.T) {
	t.Parallel()

	signals := []domain.BoundarySignal{
		helperSignal(t, domain.SignalTypeDifferentTrigger),
	}

	sketch, err := domain.NewBoundedContextSketch(
		"Ordering",
		vo.SubdomainCore,
		0.75,
		[]string{"Customer"},
		[]string{"Order"},
		[]string{"Place Order"},
		signals,
		vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, "Ordering", sketch.Name())
	assert.Equal(t, vo.SubdomainCore, sketch.Classification())
	assert.InDelta(t, 0.75, sketch.Confidence(), 0)
	assert.Equal(t, []string{"Customer"}, sketch.Actors())
	assert.Equal(t, []string{"Order"}, sketch.WorkObjects())
	assert.Equal(t, []string{"Place Order"}, sketch.Stories())
	assert.Len(t, sketch.Signals(), 1)
	assert.Equal(t, vo.UserStated, sketch.Trust())
}

func TestNewBoundedContextSketch_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundedContextSketch(
		"", vo.SubdomainCore, 0.5, nil, nil, nil, nil, vo.UserStated,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundedContextSketch_WhitespaceName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundedContextSketch(
		"   ", vo.SubdomainCore, 0.5, nil, nil, nil, nil, vo.UserStated,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundedContextSketch_TrimsName(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"  Ordering  ", vo.SubdomainCore, 0.5, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, "Ordering", sketch.Name())
}

func TestNewBoundedContextSketch_InvalidClassification(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainClassification("invalid"), 0.5, nil, nil, nil, nil, vo.UserStated,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundedContextSketch_AllClassifications(t *testing.T) {
	t.Parallel()
	for _, c := range vo.AllSubdomainClassifications() {
		t.Run(string(c), func(t *testing.T) {
			t.Parallel()
			sketch, err := domain.NewBoundedContextSketch(
				"Test", c, 0.5, nil, nil, nil, nil, vo.UserStated,
			)
			require.NoError(t, err)
			assert.Equal(t, c, sketch.Classification())
		})
	}
}

func TestNewBoundedContextSketch_ConfidenceBelowZero(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, -0.01, nil, nil, nil, nil, vo.UserStated,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundedContextSketch_ConfidenceAboveOne(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 1.01, nil, nil, nil, nil, vo.UserStated,
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundedContextSketch_ConfidenceBoundaryZero(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.0, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, sketch.Confidence(), 0)
}

func TestNewBoundedContextSketch_ConfidenceBoundaryOne(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 1.0, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, sketch.Confidence(), 0)
}

func TestNewBoundedContextSketch_InvalidTrust(t *testing.T) {
	t.Parallel()
	_, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.5, nil, nil, nil, nil, vo.TrustLevel(99),
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewBoundedContextSketch_EmptySlices(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.5, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Empty(t, sketch.Actors())
	assert.Empty(t, sketch.WorkObjects())
	assert.Empty(t, sketch.Stories())
	assert.Empty(t, sketch.Signals())
}

// -- Defensive copy tests --

func TestBoundedContextSketch_ActorsDefensiveCopy(t *testing.T) {
	t.Parallel()
	actors := []string{"Customer", "Seller"}
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.5, actors, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)

	// Mutate original slice.
	actors[0] = "Hacker"
	assert.Equal(t, "Customer", sketch.Actors()[0])

	// Mutate returned slice.
	got := sketch.Actors()
	got[0] = "Hacker"
	assert.Equal(t, "Customer", sketch.Actors()[0])
}

func TestBoundedContextSketch_WorkObjectsDefensiveCopy(t *testing.T) {
	t.Parallel()
	objs := []string{"Order", "Invoice"}
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.5, nil, objs, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)

	// Mutate original slice.
	objs[0] = "Hacked"
	assert.Equal(t, "Order", sketch.WorkObjects()[0])

	// Mutate returned slice.
	got := sketch.WorkObjects()
	got[0] = "Hacked"
	assert.Equal(t, "Order", sketch.WorkObjects()[0])
}

func TestBoundedContextSketch_StoriesDefensiveCopy(t *testing.T) {
	t.Parallel()
	stories := []string{"Place Order", "Cancel Order"}
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.5, nil, nil, stories, nil, vo.UserStated,
	)
	require.NoError(t, err)

	// Mutate original slice.
	stories[0] = "Hacked"
	assert.Equal(t, "Place Order", sketch.Stories()[0])

	// Mutate returned slice.
	got := sketch.Stories()
	got[0] = "Hacked"
	assert.Equal(t, "Place Order", sketch.Stories()[0])
}

func TestBoundedContextSketch_SignalsDefensiveCopy(t *testing.T) {
	t.Parallel()
	sig1 := helperSignal(t, domain.SignalTypeDifferentTrigger)
	sig2 := helperSignal(t, domain.SignalTypeOneWayFlow)
	signals := []domain.BoundarySignal{sig1, sig2}

	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.5, nil, nil, nil, signals, vo.UserStated,
	)
	require.NoError(t, err)

	// Mutate original slice.
	signals[0] = helperSignal(t, domain.SignalTypeComplexRules)
	assert.Equal(t, domain.SignalTypeDifferentTrigger, sketch.Signals()[0].Type())

	// Mutate returned slice.
	got := sketch.Signals()
	got[0] = helperSignal(t, domain.SignalTypeComplexRules)
	assert.Equal(t, domain.SignalTypeDifferentTrigger, sketch.Signals()[0].Type())
}

// -- ConfidenceLevel threshold tests --

func TestBoundedContextSketch_ConfidenceLevel_HighByScore(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.65, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelHigh, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_HighByScoreAbove(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.90, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelHigh, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_HighByDistinctSignals(t *testing.T) {
	t.Parallel()
	// Low confidence but 2+ distinct signal types => HIGH.
	signals := []domain.BoundarySignal{
		helperSignal(t, domain.SignalTypeDifferentTrigger),
		helperSignal(t, domain.SignalTypeOneWayFlow),
	}
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.30, nil, nil, nil, signals, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelHigh, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_HighByDistinctSignals_DuplicatesNotCounted(t *testing.T) {
	t.Parallel()
	// 3 signals but only 1 distinct type => NOT enough for HIGH via signals alone.
	signals := []domain.BoundarySignal{
		helperSignal(t, domain.SignalTypeDifferentTrigger),
		helperSignal(t, domain.SignalTypeDifferentTrigger),
		helperSignal(t, domain.SignalTypeDifferentTrigger),
	}
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.50, nil, nil, nil, signals, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelMedium, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_Medium(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.50, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelMedium, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_MediumLowerBound(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.45, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelMedium, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_MediumUpperBound(t *testing.T) {
	t.Parallel()
	// 0.64 with < 2 distinct signal types => MEDIUM.
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.64, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelMedium, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_Low(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.30, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelLow, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_LowAtZero(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.0, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelLow, sketch.ConfidenceLevel())
}

func TestBoundedContextSketch_ConfidenceLevel_LowJustBelow045(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.44, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.ConfidenceLevelLow, sketch.ConfidenceLevel())
}

// -- String tests --

func TestBoundedContextSketch_String(t *testing.T) {
	t.Parallel()
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering", vo.SubdomainCore, 0.75, nil, nil, nil, nil, vo.UserStated,
	)
	require.NoError(t, err)
	expected := "BoundedContextSketch: Ordering (core, confidence=0.75, level=HIGH, user_stated)"
	assert.Equal(t, expected, sketch.String())
}
