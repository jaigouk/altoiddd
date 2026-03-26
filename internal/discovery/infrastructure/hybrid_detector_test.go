package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	application "github.com/alto-cli/alto/internal/discovery/application"
	domain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- Test fakes ---

type fakeAlgorithmicDetector struct {
	sketches []domain.BoundedContextSketch
	err      error
}

func (f *fakeAlgorithmicDetector) DetectBoundaries(
	_ context.Context,
	_ []*domain.DomainStory,
	_ domain.DiscoveryMode,
) ([]domain.BoundedContextSketch, error) {
	return f.sketches, f.err
}

type fakeLLMEnricher struct {
	signals []domain.BoundarySignal
	err     error
	called  bool
}

func (f *fakeLLMEnricher) DetectBoundarySignals(
	_ context.Context,
	_ []*domain.DomainStory,
) ([]domain.BoundarySignal, error) {
	f.called = true
	return f.signals, f.err
}

// --- Helper to build a sketch for testing ---

func makeTestSketch(
	t *testing.T,
	name string,
	confidence float64,
	actors []string,
	workObjects []string,
	stories []string,
	signals []domain.BoundarySignal,
) domain.BoundedContextSketch {
	t.Helper()

	sketch, err := domain.NewBoundedContextSketch(
		name, vo.SubdomainGeneric, confidence, actors, workObjects, stories, signals, vo.AIInferred,
	)
	require.NoError(t, err)

	return sketch
}

func makeTestSignal(t *testing.T, st domain.SignalType, desc string) domain.BoundarySignal {
	t.Helper()

	signal, err := domain.NewBoundarySignal(st, desc)
	require.NoError(t, err)

	return signal
}

// --- Tests ---

func TestHybridBoundaryDetector_CompileTimeCheck(t *testing.T) {
	t.Parallel()

	// Compile-time verification that HybridBoundaryDetector satisfies BoundaryDetector.
	var _ application.BoundaryDetector = (*HybridBoundaryDetector)(nil)

	detector := NewHybridBoundaryDetector(&fakeAlgorithmicDetector{}, nil)
	assert.NotNil(t, detector)
}

func TestHybridBoundaryDetector_MergesAlgorithmicAndLLMSignals(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns 1 sketch with actor "Clerk", workObject "Invoice", 1 signal
	algoSignal := makeTestSignal(t, domain.SignalTypeSameObjectDiffContext, "invoice used in diff contexts")
	algoSketch := makeTestSketch(t, "Invoice-Context", 0.40,
		[]string{"Clerk"}, []string{"Invoice"}, []string{"Story1", "Story2"},
		[]domain.BoundarySignal{algoSignal},
	)

	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{algoSketch}}

	// And: LLM returns 1 signal mentioning "Invoice" (matches the sketch)
	llmSignal := makeTestSignal(t, domain.SignalTypeLanguageDifference, "Invoice means different things in billing vs shipping")
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{llmSignal}}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeThorough)

	// Then
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Len(t, result[0].Signals(), 2)
	assert.True(t, enricher.called)
	// Re-scored: should use AIInferred trust and SubdomainGeneric classification
	assert.Equal(t, vo.SubdomainGeneric, result[0].Classification())
	assert.Equal(t, vo.AIInferred, result[0].Trust())
}

func TestHybridBoundaryDetector_FallbackWhenLLMNil(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns 2 sketches with specific confidences
	sketch1 := makeTestSketch(t, "A-Context", 0.50,
		[]string{"Alice"}, []string{"Report"}, []string{"S1"},
		[]domain.BoundarySignal{makeTestSignal(t, domain.SignalTypeOrgBoundary, "org boundary signal")},
	)
	sketch2 := makeTestSketch(t, "B-Context", 0.70,
		[]string{"Bob"}, []string{"Invoice"}, []string{"S2"},
		[]domain.BoundarySignal{makeTestSignal(t, domain.SignalTypeSameObjectDiffContext, "same obj diff ctx")},
	)

	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{sketch1, sketch2}}

	// When: nil LLM detector
	detector := NewHybridBoundaryDetector(algo, nil)
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeRapid)

	// Then: algorithmic results returned unchanged (same order, same confidence)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "A-Context", result[0].Name())
	assert.InDelta(t, 0.50, result[0].Confidence(), 0)
	assert.Equal(t, "B-Context", result[1].Name())
	assert.InDelta(t, 0.70, result[1].Confidence(), 0)
}

func TestHybridBoundaryDetector_FallbackOnLLMError(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns 1 sketch
	algoSketch := makeTestSketch(t, "Ctx", 0.55,
		[]string{"Alice"}, []string{"Order"}, []string{"S1"},
		[]domain.BoundarySignal{makeTestSignal(t, domain.SignalTypeOneWayFlow, "one way flow")},
	)
	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{algoSketch}}

	// And: LLM returns error
	enricher := &fakeLLMEnricher{err: errors.New("boom")}
	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeRapid)

	// Then: error NOT propagated, algorithmic results returned unchanged
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Ctx", result[0].Name())
	assert.InDelta(t, 0.55, result[0].Confidence(), 0)
}

func TestHybridBoundaryDetector_FiltersLowInRapidMode(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns 2 sketches - one HIGH confidence, one LOW confidence
	highSignal := makeTestSignal(t, domain.SignalTypeSameObjectDiffContext, "high confidence signal mentioning HighActor")
	highSketch := makeTestSketch(t, "High-Context", 0.70,
		[]string{"HighActor"}, []string{"HighObj"}, []string{"S1", "S2"},
		[]domain.BoundarySignal{highSignal},
	)

	// LOW: confidence < 0.45 with < 2 distinct signal types
	lowSignal := makeTestSignal(t, domain.SignalTypeOrgBoundary, "low confidence signal mentioning LowActor")
	lowSketch := makeTestSketch(t, "Low-Context", 0.20,
		[]string{"LowActor"}, []string{"LowObj"}, []string{"S3"},
		[]domain.BoundarySignal{lowSignal},
	)

	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{highSketch, lowSketch}}

	// LLM returns signals that match both sketches (to trigger re-scoring)
	llmHighSignal := makeTestSignal(t, domain.SignalTypeLanguageDifference, "language diff involving HighObj")
	llmLowSignal := makeTestSignal(t, domain.SignalTypeLanguageDifference, "language diff involving LowObj")
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{llmHighSignal, llmLowSignal}}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When: RAPID mode
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeRapid)

	// Then: LOW sketch filtered out
	require.NoError(t, err)

	for _, sketch := range result {
		assert.NotEqual(t, domain.ConfidenceLevelLow, sketch.ConfidenceLevel(),
			"rapid mode should filter out LOW confidence sketches, found: %s", sketch.Name())
	}
}

func TestHybridBoundaryDetector_KeepsLowInThoroughMode(t *testing.T) {
	t.Parallel()

	// Given: same setup as rapid mode test
	highSignal := makeTestSignal(t, domain.SignalTypeSameObjectDiffContext, "high confidence signal mentioning HighActor")
	highSketch := makeTestSketch(t, "High-Context", 0.70,
		[]string{"HighActor"}, []string{"HighObj"}, []string{"S1", "S2"},
		[]domain.BoundarySignal{highSignal},
	)

	lowSignal := makeTestSignal(t, domain.SignalTypeOrgBoundary, "low confidence signal mentioning LowActor")
	lowSketch := makeTestSketch(t, "Low-Context", 0.20,
		[]string{"LowActor"}, []string{"LowObj"}, []string{"S3"},
		[]domain.BoundarySignal{lowSignal},
	)

	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{highSketch, lowSketch}}

	llmHighSignal := makeTestSignal(t, domain.SignalTypeLanguageDifference, "language diff involving HighObj")
	llmLowSignal := makeTestSignal(t, domain.SignalTypeLanguageDifference, "language diff involving LowObj")
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{llmHighSignal, llmLowSignal}}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When: THOROUGH mode
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeThorough)

	// Then: both sketches present (low NOT filtered)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(result), 2)
}

func TestHybridBoundaryDetector_SortsByScoreDescending(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns 3 sketches with different confidences
	sketch1 := makeTestSketch(t, "Low-Ctx", 0.30,
		[]string{"Alice"}, []string{"Alpha"}, []string{"S1"},
		[]domain.BoundarySignal{makeTestSignal(t, domain.SignalTypeOrgBoundary, "signal about Alpha")},
	)
	sketch2 := makeTestSketch(t, "High-Ctx", 0.80,
		[]string{"Bob"}, []string{"Beta"}, []string{"S2"},
		[]domain.BoundarySignal{makeTestSignal(t, domain.SignalTypeSameObjectDiffContext, "signal about Beta")},
	)
	sketch3 := makeTestSketch(t, "Mid-Ctx", 0.55,
		[]string{"Charlie"}, []string{"Gamma"}, []string{"S3"},
		[]domain.BoundarySignal{makeTestSignal(t, domain.SignalTypeOneWayFlow, "signal about Gamma")},
	)

	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{sketch1, sketch2, sketch3}}

	// LLM returns signals matching each sketch
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{
		makeTestSignal(t, domain.SignalTypeLanguageDifference, "language diff about Alpha"),
		makeTestSignal(t, domain.SignalTypeLanguageDifference, "language diff about Beta"),
		makeTestSignal(t, domain.SignalTypeLanguageDifference, "language diff about Gamma"),
	}}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeThorough)

	// Then: sorted by confidence descending
	require.NoError(t, err)
	require.NotEmpty(t, result)

	for i := 1; i < len(result); i++ {
		assert.GreaterOrEqual(t, result[i-1].Confidence(), result[i].Confidence(),
			"result[%d] (%.2f) should be >= result[%d] (%.2f)",
			i-1, result[i-1].Confidence(), i, result[i].Confidence())
	}
}

func TestHybridBoundaryDetector_EmptyStoriesReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns (nil, nil) for empty input
	algo := &fakeAlgorithmicDetector{sketches: nil, err: nil}
	enricher := &fakeLLMEnricher{signals: nil}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeRapid)

	// Then
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestHybridBoundaryDetector_AlgorithmicErrorPropagates(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns error
	algoErr := errors.New("algorithmic failure")
	algo := &fakeAlgorithmicDetector{err: algoErr}
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{
		makeTestSignal(t, domain.SignalTypeLanguageDifference, "should not be called"),
	}}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	_, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeRapid)

	// Then: error propagated, LLM NOT called
	require.Error(t, err)
	assert.False(t, enricher.called, "LLM enricher should not be called when algorithmic fails")
}

func TestHybridBoundaryDetector_OrphanLLMSignalsCreateNewClusters(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns 1 sketch with actor "Clerk"
	algoSignal := makeTestSignal(t, domain.SignalTypeSameObjectDiffContext, "invoice diff context")
	algoSketch := makeTestSketch(t, "Clerk-Context", 0.50,
		[]string{"Clerk"}, []string{"Invoice"}, []string{"S1"},
		[]domain.BoundarySignal{algoSignal},
	)

	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{algoSketch}}

	// And: LLM returns 1 signal mentioning "Warehouse" (not in any sketch)
	orphanSignal := makeTestSignal(t, domain.SignalTypeLanguageDifference, "Warehouse uses different terminology")
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{orphanSignal}}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeThorough)

	// Then: 2 sketches (original + orphan)
	require.NoError(t, err)
	require.Len(t, result, 2)

	// Verify orphan sketch properties
	var orphanFound bool

	for _, sketch := range result {
		if sketch.Name() != "Clerk-Context" {
			orphanFound = true
			assert.Equal(t, vo.SubdomainGeneric, sketch.Classification())
			assert.Equal(t, vo.AIInferred, sketch.Trust())
			assert.Len(t, sketch.Signals(), 1)
		}
	}

	assert.True(t, orphanFound, "expected an orphan cluster sketch")
}

func TestHybridBoundaryDetector_FallbackOnLLMEmptySignals(t *testing.T) {
	t.Parallel()

	// Given: algorithmic returns 1 sketch
	algoSketch := makeTestSketch(t, "Ctx", 0.55,
		[]string{"Alice"}, []string{"Order"}, []string{"S1"},
		[]domain.BoundarySignal{makeTestSignal(t, domain.SignalTypeOneWayFlow, "one way flow")},
	)
	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{algoSketch}}

	// And: LLM returns empty (non-nil) slice
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{}}
	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeRapid)

	// Then: algorithmic results returned unchanged
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Ctx", result[0].Name())
	assert.InDelta(t, 0.55, result[0].Confidence(), 0)
}

func TestHybridBoundaryDetector_NameMatchingCaseInsensitive(t *testing.T) {
	t.Parallel()

	// Given: sketch has actor "Invoice" (capitalized)
	algoSignal := makeTestSignal(t, domain.SignalTypeSameObjectDiffContext, "invoice used differently")
	algoSketch := makeTestSketch(t, "Invoice-Context", 0.40,
		[]string{"Invoice"}, nil, []string{"S1"},
		[]domain.BoundarySignal{algoSignal},
	)

	algo := &fakeAlgorithmicDetector{sketches: []domain.BoundedContextSketch{algoSketch}}

	// And: LLM signal description contains "invoice" (lowercase)
	llmSignal := makeTestSignal(t, domain.SignalTypeLanguageDifference, "the term invoice differs between contexts")
	enricher := &fakeLLMEnricher{signals: []domain.BoundarySignal{llmSignal}}

	detector := &HybridBoundaryDetector{algorithmic: algo, llmEnricher: enricher}

	// When
	result, err := detector.DetectBoundaries(context.Background(), nil, domain.ModeThorough)

	// Then: signal matched to sketch (case-insensitive), not orphan
	require.NoError(t, err)
	require.Len(t, result, 1)
	// The merged sketch should have 2 signals (original + matched LLM signal)
	assert.Len(t, result[0].Signals(), 2)
}
