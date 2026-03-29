package infrastructure

import (
	"context"
	"fmt"
	"sort"
	"strings"

	application "github.com/alto-cli/alto/internal/discovery/application"
	domain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// llmBoundaryEnricher is infrastructure-internal for testability.
// Tests create fakes. Composition root passes *LLMBoundaryDetector directly.
type llmBoundaryEnricher interface {
	DetectBoundarySignals(ctx context.Context, stories []*domain.DomainStory) ([]domain.BoundarySignal, error)
}

// HybridBoundaryDetector composes algorithmic + LLM boundary detection into a
// single BoundaryDetector. It merges LLM-discovered signals into algorithmic
// sketches, re-scores merged clusters, creates orphan clusters from unmatched
// LLM signals, filters by mode, and sorts by confidence.
type HybridBoundaryDetector struct {
	algorithmic application.BoundaryDetector
	llmEnricher llmBoundaryEnricher // nil = offline/algorithmic-only
}

// Compile-time check that HybridBoundaryDetector satisfies BoundaryDetector.
var _ application.BoundaryDetector = (*HybridBoundaryDetector)(nil)

// NewHybridBoundaryDetector creates a HybridBoundaryDetector.
// The llmDetector parameter accepts a concrete *LLMBoundaryDetector; pass nil
// for offline/algorithmic-only mode.
func NewHybridBoundaryDetector(
	algorithmic application.BoundaryDetector,
	llmDetector *LLMBoundaryDetector,
) *HybridBoundaryDetector {
	var enricher llmBoundaryEnricher
	if llmDetector != nil {
		enricher = llmDetector
	}

	return &HybridBoundaryDetector{algorithmic: algorithmic, llmEnricher: enricher}
}

// DetectBoundaries runs the merge pipeline:
// 1. Algorithmic detection
// 2. LLM enrichment (if available)
// 3. Signal matching + merge
// 4. Re-scoring merged clusters
// 5. Orphan cluster creation
// 6. Mode filtering
// 7. Confidence sorting
func (h *HybridBoundaryDetector) DetectBoundaries(
	ctx context.Context,
	stories []*domain.DomainStory,
	mode domain.DiscoveryMode,
) ([]domain.BoundedContextSketch, error) {
	// Step 1: algorithmic detection.
	baseSketches, err := h.algorithmic.DetectBoundaries(ctx, stories, mode)
	if err != nil {
		return nil, fmt.Errorf("algorithmic boundary detection: %w", err)
	}

	// Step 2: if no LLM enricher, return algorithmic directly (NO re-scoring, NO sort).
	if h.llmEnricher == nil {
		return baseSketches, nil
	}

	llmSignals, llmErr := h.llmEnricher.DetectBoundarySignals(ctx, stories)

	// Fallback: LLM error or empty signals — return algorithmic directly.
	if llmErr != nil || len(llmSignals) == 0 {
		return baseSketches, nil
	}

	// Step 3: match LLM signals to algorithmic sketches.
	matched, orphans := matchSignalsToSketches(llmSignals, baseSketches)

	// Step 4+5: build final sketches.
	result, err := buildMergedSketches(baseSketches, matched, orphans)
	if err != nil {
		return nil, fmt.Errorf("building merged sketches: %w", err)
	}

	// Step 6: filter by mode.
	result = filterByMode(result, mode)

	// Step 7: sort by confidence descending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Confidence() > result[j].Confidence()
	})

	return result, nil
}

// matchSignalsToSketches distributes LLM signals to sketches they implicate.
// Returns matched (sketch index -> signals) and orphan (unmatched) signals.
func matchSignalsToSketches(
	signals []domain.BoundarySignal,
	sketches []domain.BoundedContextSketch,
) (map[int][]domain.BoundarySignal, []domain.BoundarySignal) {
	matched := make(map[int][]domain.BoundarySignal)
	var orphans []domain.BoundarySignal

	for _, signal := range signals {
		wasMatched := false

		for i, sketch := range sketches {
			if signalImplicatesSketch(signal, sketch) {
				matched[i] = append(matched[i], signal)
				wasMatched = true

				break
			}
		}

		if !wasMatched {
			orphans = append(orphans, signal)
		}
	}

	return matched, orphans
}

// signalImplicatesSketch checks whether the signal's description mentions any
// actor or work object from the sketch (case-insensitive substring match).
// TODO: plural stripping deferred — spike Section 13.5
func signalImplicatesSketch(signal domain.BoundarySignal, sketch domain.BoundedContextSketch) bool {
	descLower := strings.ToLower(signal.Description())

	for _, actor := range sketch.Actors() {
		if strings.Contains(descLower, strings.ToLower(strings.TrimSpace(actor))) {
			return true
		}
	}

	for _, wo := range sketch.WorkObjects() {
		if strings.Contains(descLower, strings.ToLower(strings.TrimSpace(wo))) {
			return true
		}
	}

	return false
}

// buildMergedSketches combines unchanged, re-scored, and orphan sketches.
func buildMergedSketches(
	baseSketches []domain.BoundedContextSketch,
	matched map[int][]domain.BoundarySignal,
	orphans []domain.BoundarySignal,
) ([]domain.BoundedContextSketch, error) {
	var result []domain.BoundedContextSketch

	for i, sketch := range baseSketches {
		newSignals, wasMerged := matched[i]
		if !wasMerged {
			// Unchanged sketch — keep original.
			result = append(result, sketch)
			continue
		}

		// Re-score merged sketch.
		merged, err := rescoreMergedSketch(sketch, newSignals)
		if err != nil {
			return nil, fmt.Errorf("re-scoring sketch %q: %w", sketch.Name(), err)
		}

		result = append(result, merged)
	}

	// Create orphan clusters.
	for _, signal := range orphans {
		orphanSketch, err := createOrphanSketch(signal)
		if err != nil {
			return nil, fmt.Errorf("creating orphan sketch: %w", err)
		}

		result = append(result, orphanSketch)
	}

	return result, nil
}

// rescoreMergedSketch re-scores an algorithmic sketch that received additional LLM signals.
func rescoreMergedSketch(
	original domain.BoundedContextSketch,
	newSignals []domain.BoundarySignal,
) (domain.BoundedContextSketch, error) {
	// Collect ALL signals (original + new LLM signals).
	allSignals := append(original.Signals(), newSignals...)

	// Compute score using base confidence per signal type. This ignores
	// algorithmic-specific adjustments (notification verb / sequential flow
	// discounts) because BoundarySignal does not carry per-signal confidence.
	// Intentional simplification — adding LLM evidence resets the baseline.
	var confSum float64

	typeSet := make(map[domain.SignalType]struct{})

	for _, sig := range allSignals {
		confSum += baseConfidenceForSignalType(sig.Type())
		typeSet[sig.Type()] = struct{}{}
	}

	signalAvg := confSum / float64(len(allSignals))
	distinctTypeCount := len(typeSet)
	storyCount := len(original.Stories())
	// Org-boundary story ceiling is NOT applied here. Adding LLM signals means
	// the cluster is no longer org-boundary-only, so the ceiling condition from
	// AlgorithmicDetector does not apply.
	score := domain.ComputeBoundaryScore(signalAvg, distinctTypeCount, storyCount)

	// Cap at 1.0.
	if score > 1.0 {
		score = 1.0
	}

	sketch, err := domain.NewBoundedContextSketch(
		original.Name(),
		vo.SubdomainGeneric,
		score,
		original.Actors(),
		original.WorkObjects(),
		original.Stories(),
		allSignals,
		vo.AIInferred,
	)
	if err != nil {
		return domain.BoundedContextSketch{}, fmt.Errorf("creating re-scored sketch: %w", err)
	}

	return sketch, nil
}

// createOrphanSketch creates a new sketch from an unmatched LLM signal.
func createOrphanSketch(signal domain.BoundarySignal) (domain.BoundedContextSketch, error) {
	name := deriveOrphanName(signal.Type())
	baseConf := baseConfidenceForSignalType(signal.Type())
	score := domain.ComputeBoundaryScore(baseConf, 1, 0)

	// Cap at 1.0.
	if score > 1.0 {
		score = 1.0
	}

	sketch, err := domain.NewBoundedContextSketch(
		name,
		vo.SubdomainGeneric,
		score,
		nil,
		nil,
		nil,
		[]domain.BoundarySignal{signal},
		vo.AIInferred,
	)
	if err != nil {
		return domain.BoundedContextSketch{}, fmt.Errorf("creating orphan sketch: %w", err)
	}

	return sketch, nil
}

// deriveOrphanName creates a name from a signal type by capitalizing and replacing
// underscores with dashes: "language_difference" -> "Language-Difference".
func deriveOrphanName(st domain.SignalType) string {
	parts := strings.Split(string(st), "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}

	return strings.Join(parts, "-")
}

// filterByMode removes LOW confidence sketches in rapid mode.
func filterByMode(sketches []domain.BoundedContextSketch, mode domain.DiscoveryMode) []domain.BoundedContextSketch {
	if mode != domain.ModeRapid {
		return sketches
	}

	var filtered []domain.BoundedContextSketch

	for _, sketch := range sketches {
		if sketch.ConfidenceLevel() != domain.ConfidenceLevelLow {
			filtered = append(filtered, sketch)
		}
	}

	return filtered
}

// baseConfidenceForSignalType returns the base confidence for a given signal type.
func baseConfidenceForSignalType(st domain.SignalType) float64 {
	switch st {
	case domain.SignalTypeSameObjectDiffContext:
		return domain.BaseConfidenceSameObjectDiffContext
	case domain.SignalTypeOneWayFlow:
		return domain.BaseConfidenceOneWayFlow
	case domain.SignalTypeOrgBoundary:
		return domain.BaseConfidenceOrgBoundary
	case domain.SignalTypeDifferentTrigger:
		return domain.BaseConfidenceDifferentTrigger
	case domain.SignalTypeLanguageDifference:
		return domain.BaseConfidenceLanguageDifference
	case domain.SignalTypeDifferentLifecycle,
		domain.SignalTypeExternalSystem,
		domain.SignalTypeDifferentActor,
		domain.SignalTypeComplexRules,
		domain.SignalTypeWorkObjectCluster:
		return 0.15
	}

	return 0.15
}
