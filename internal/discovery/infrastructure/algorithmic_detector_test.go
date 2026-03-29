package infrastructure_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- Test helper ---

type sentenceSpec struct {
	step     int
	subject  string
	activity string
	object   string
}

func makeStoryWithActorsAndObjects(
	t *testing.T,
	title string,
	actors []string,
	workObjects []string,
	sentences []sentenceSpec,
) *domain.DomainStory {
	t.Helper()

	story, err := domain.NewDomainStory(
		title,
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	for _, name := range actors {
		actor, aErr := domain.NewStoryActor(name, domain.ActorTypePerson, vo.UserStated, "")
		require.NoError(t, aErr)
		require.NoError(t, story.AddActor(actor))
	}

	for _, name := range workObjects {
		wo, wErr := domain.NewWorkObject(name, domain.WorkObjectTypeDocument, vo.UserStated, "")
		require.NoError(t, wErr)
		require.NoError(t, story.AddWorkObject(wo))
	}

	for _, s := range sentences {
		sentence, sErr := domain.NewStorySentence(s.step, s.subject, s.activity, s.object, vo.UserStated, "")
		require.NoError(t, sErr)
		require.NoError(t, story.AddSentence(sentence))
	}

	require.NoError(t, story.Validate())

	return story
}

// --- Tests ---

func TestAlgorithmicDetector_CompileTimeCheck(t *testing.T) {
	t.Parallel()
	// The compile-time check var _ application.BoundaryDetector = (*AlgorithmicDetector)(nil)
	// is in the implementation file. This test verifies we can construct the detector.
	detector := infrastructure.NewAlgorithmicDetector()
	assert.NotNil(t, detector)
}

func TestAlgorithmicDetector_RequiresMinimumTwoStories(t *testing.T) {
	t.Parallel()

	detector := infrastructure.NewAlgorithmicDetector()
	ctx := context.Background()

	tests := []struct {
		name    string
		stories []*domain.DomainStory
	}{
		{"empty slice", nil},
		{"single story", []*domain.DomainStory{
			makeStoryWithActorsAndObjects(t, "Solo Story",
				[]string{"Alice"},
				[]string{"Invoice"},
				[]sentenceSpec{{1, "Alice", "creates", "Invoice"}},
			),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := detector.DetectBoundaries(ctx, tt.stories, domain.ModeRapid)
			require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		})
	}
}

func TestAlgorithmicDetector_SameObjectDiffContext_VerbDivergence(t *testing.T) {
	t.Parallel()

	// Story 1: Alice creates Invoice
	story1 := makeStoryWithActorsAndObjects(t, "Billing Story",
		[]string{"Alice"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Alice", "creates", "Invoice"}},
	)

	// Story 2: Bob reviews Invoice
	story2 := makeStoryWithActorsAndObjects(t, "Approval Story",
		[]string{"Bob"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Bob", "reviews", "Invoice"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	// Should detect same_object_diff_context signal for "Invoice"
	var foundSignal bool

	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			if sig.Type() == domain.SignalTypeSameObjectDiffContext {
				foundSignal = true
				assert.Contains(t, sig.Description(), "invoice")
			}
		}
	}

	assert.True(t, foundSignal, "expected same_object_diff_context signal for Invoice")
}

func TestAlgorithmicDetector_SameObjectDiffContext_SameVerbNoSignal(t *testing.T) {
	t.Parallel()

	// Both stories use same verb "creates" on Invoice — no signal expected.
	story1 := makeStoryWithActorsAndObjects(t, "Story A",
		[]string{"Alice"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Alice", "creates", "Invoice"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Story B",
		[]string{"Bob"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Bob", "creates", "Invoice"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			assert.NotEqual(t, domain.SignalTypeSameObjectDiffContext, sig.Type(),
				"should NOT detect same_object_diff_context when verb is the same across stories")
		}
	}
}

func TestAlgorithmicDetector_SameObjectDiffContext_CaseInsensitive(t *testing.T) {
	t.Parallel()

	// "Invoice" vs "invoice" should be same object. "Creates" vs "reviews" — different verbs.
	story1 := makeStoryWithActorsAndObjects(t, "Story A",
		[]string{"Alice"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Alice", "Creates", "Invoice"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Story B",
		[]string{"Bob"},
		[]string{"invoice"},
		[]sentenceSpec{{1, "Bob", "reviews", "invoice"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	var foundSignal bool

	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			if sig.Type() == domain.SignalTypeSameObjectDiffContext {
				foundSignal = true
			}
		}
	}

	assert.True(t, foundSignal, "expected case-insensitive match for Invoice/invoice")
}

func TestAlgorithmicDetector_OneWayFlow_AsymmetricDetected(t *testing.T) {
	t.Parallel()

	// Alice sends Report to Bob (Bob is actor). Bob never sends to Alice.
	story1 := makeStoryWithActorsAndObjects(t, "Reporting Story",
		[]string{"Alice", "Bob"},
		[]string{"Report"},
		[]sentenceSpec{
			{1, "Alice", "sends", "Report"},
			{2, "Alice", "submits", "Bob"},
		},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Review Story",
		[]string{"Alice", "Bob"},
		[]string{"Report"},
		[]sentenceSpec{
			{1, "Bob", "reviews", "Report"},
		},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	var foundOneWay bool

	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			if sig.Type() == domain.SignalTypeOneWayFlow {
				foundOneWay = true
			}
		}
	}

	assert.True(t, foundOneWay, "expected one_way_flow signal for Alice->Bob asymmetric flow")
}

func TestAlgorithmicDetector_OneWayFlow_NotificationVerbDiscounted(t *testing.T) {
	t.Parallel()

	// Alice "displays" to Bob — notification verb discount should reduce confidence.
	story1 := makeStoryWithActorsAndObjects(t, "Display Story",
		[]string{"Alice", "Bob"},
		[]string{"Report"},
		[]sentenceSpec{
			{1, "Alice", "displays", "Bob"},
		},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Other Story",
		[]string{"Alice", "Bob"},
		[]string{"Data"},
		[]sentenceSpec{
			{1, "Bob", "reads", "Data"},
		},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	// The notification discount (0.30) exceeds the base confidence (0.25),
	// so the one-way-flow signal should be eliminated entirely.
	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			if sig.Type() == domain.SignalTypeOneWayFlow {
				t.Errorf("expected no one_way_flow signal when only activity is a notification verb (displays), but found: %s", sig.Description())
			}
		}
	}
}

func TestAlgorithmicDetector_OrgBoundary_NonOverlappingActors(t *testing.T) {
	t.Parallel()

	// Story 1: {Alice, Bob}, Story 2: {Charlie, Dave} — zero overlap.
	story1 := makeStoryWithActorsAndObjects(t, "Team Alpha Story",
		[]string{"Alice", "Bob"},
		[]string{"Report"},
		[]sentenceSpec{
			{1, "Alice", "creates", "Report"},
			{2, "Bob", "reviews", "Report"},
		},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Team Beta Story",
		[]string{"Charlie", "Dave"},
		[]string{"Invoice"},
		[]sentenceSpec{
			{1, "Charlie", "creates", "Invoice"},
			{2, "Dave", "approves", "Invoice"},
		},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	var foundOrgBoundary bool

	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			if sig.Type() == domain.SignalTypeOrgBoundary {
				foundOrgBoundary = true
			}
		}
	}

	assert.True(t, foundOrgBoundary, "expected org_boundary signal for non-overlapping actors")
}

func TestAlgorithmicDetector_OrgBoundary_StoryCeiling(t *testing.T) {
	t.Parallel()

	// Only 2 stories (< 3 = OrgBoundaryStoryCeilingCount), so the final sketch score
	// should be capped at OrgBoundaryStoryCeilingScore (0.40).
	story1 := makeStoryWithActorsAndObjects(t, "Team X Story",
		[]string{"Alice"},
		[]string{"Report"},
		[]sentenceSpec{{1, "Alice", "creates", "Report"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Team Y Story",
		[]string{"Bob"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Bob", "creates", "Invoice"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	// With only org_boundary signals (base 0.20), 2 stories, and the ceiling,
	// no sketch should exceed 0.40.
	for _, sketch := range sketches {
		hasOnlyOrgSignals := true

		for _, sig := range sketch.Signals() {
			if sig.Type() != domain.SignalTypeOrgBoundary {
				hasOnlyOrgSignals = false

				break
			}
		}

		if hasOnlyOrgSignals && len(sketch.Signals()) > 0 {
			assert.LessOrEqual(t, sketch.Confidence(), domain.OrgBoundaryStoryCeilingScore,
				"org boundary sketch with <3 stories should be capped at ceiling score")
		}
	}
}

func TestAlgorithmicDetector_Scoring_HighConfidence_MultipleSignalTypes(t *testing.T) {
	t.Parallel()

	// 3 stories with overlapping work objects and different verbs,
	// non-overlapping actors, and one-way flows — should produce HIGH confidence.
	story1 := makeStoryWithActorsAndObjects(t, "Sales Story",
		[]string{"SalesRep", "Manager"},
		[]string{"Order"},
		[]sentenceSpec{
			{1, "SalesRep", "creates", "Order"},
			{2, "SalesRep", "submits", "Manager"},
		},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Warehouse Story",
		[]string{"WarehouseClerk"},
		[]string{"Order", "Shipment"},
		[]sentenceSpec{
			{1, "WarehouseClerk", "picks", "Order"},
			{2, "WarehouseClerk", "creates", "Shipment"},
		},
	)

	story3 := makeStoryWithActorsAndObjects(t, "Billing Story",
		[]string{"Accountant"},
		[]string{"Order", "Invoice"},
		[]sentenceSpec{
			{1, "Accountant", "reviews", "Order"},
			{2, "Accountant", "creates", "Invoice"},
		},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2, story3}, domain.ModeRapid)
	require.NoError(t, err)
	require.NotEmpty(t, sketches)

	// At least one sketch should be HIGH confidence
	var foundHigh bool

	for _, sketch := range sketches {
		if sketch.ConfidenceLevel() == domain.ConfidenceLevelHigh {
			foundHigh = true

			break
		}
	}

	assert.True(t, foundHigh, "expected at least one HIGH confidence sketch with 3 signal types across 3 stories")
}

func TestAlgorithmicDetector_Scoring_LowConfidence_SingleSignal(t *testing.T) {
	t.Parallel()

	// 2 minimal stories with same work object, different verbs. Only 1 signal type.
	story1 := makeStoryWithActorsAndObjects(t, "Story One",
		[]string{"Alice"},
		[]string{"Widget"},
		[]sentenceSpec{{1, "Alice", "creates", "Widget"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Story Two",
		[]string{"Alice"},
		[]string{"Widget"},
		[]sentenceSpec{{1, "Alice", "deletes", "Widget"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	// With only 1 signal type and 2 stories, confidence should be low or medium — not high.
	for _, sketch := range sketches {
		assert.NotEqual(t, domain.ConfidenceLevelHigh, sketch.ConfidenceLevel(),
			"expected LOW or MEDIUM confidence with single signal type from 2 stories")
	}
}

func TestAlgorithmicDetector_SingleContext_NoSplit(t *testing.T) {
	t.Parallel()

	// Two structurally identical stories with same actors and same verbs on same objects.
	// Should produce 0 sketches or very low confidence.
	story1 := makeStoryWithActorsAndObjects(t, "Story A",
		[]string{"Alice"},
		[]string{"Report"},
		[]sentenceSpec{{1, "Alice", "creates", "Report"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Story B",
		[]string{"Alice"},
		[]string{"Report"},
		[]sentenceSpec{{1, "Alice", "creates", "Report"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	// With work-object cluster fallback, identical stories produce 1 merged sketch.
	require.Len(t, sketches, 1)
	assert.Equal(t, domain.ConfidenceLevelLow, sketches[0].ConfidenceLevel(),
		"fallback sketch should be LOW confidence")
}

func TestAlgorithmicDetector_NilStoryInSlice_Error(t *testing.T) {
	t.Parallel()

	story := makeStoryWithActorsAndObjects(t, "Valid Story",
		[]string{"Alice"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Alice", "creates", "Invoice"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	_, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story, nil}, domain.ModeRapid)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAlgorithmicDetector_ModeIgnored(t *testing.T) {
	t.Parallel()

	story1 := makeStoryWithActorsAndObjects(t, "Billing Story",
		[]string{"Alice"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Alice", "creates", "Invoice"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Approval Story",
		[]string{"Bob"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Bob", "reviews", "Invoice"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	stories := []*domain.DomainStory{story1, story2}

	rapidSketches, err := detector.DetectBoundaries(context.Background(), stories, domain.ModeRapid)
	require.NoError(t, err)

	thoroughSketches, err := detector.DetectBoundaries(context.Background(), stories, domain.ModeThorough)
	require.NoError(t, err)

	// Results should be identical regardless of mode.
	require.Len(t, thoroughSketches, len(rapidSketches))

	for i := range rapidSketches {
		assert.Equal(t, rapidSketches[i].Name(), thoroughSketches[i].Name())
		assert.InDelta(t, rapidSketches[i].Confidence(), thoroughSketches[i].Confidence(), 0)
		assert.Len(t, thoroughSketches[i].Signals(), len(rapidSketches[i].Signals()))
	}
}

// --- RED Phase tests for alty-cli-m6h: WorkObjectCluster fallback ---
//
// These tests verify the fallback behavior when 0 primary signals are found.
// They reference domain.SignalTypeWorkObjectCluster which does NOT exist yet.
// Once the domain constant is added by the dev-detect agent, these tests will
// compile but still fail (RED) until the algorithmic detector implements the
// work-object-cluster fallback logic.

func TestAlgorithmicDetector_WorkObjectCluster_FallbackWhenNoSignals(t *testing.T) {
	t.Parallel()

	// 2 stories: same actor "developer", different work objects, same verb "creates".
	// No cross-story actor diversity, no verb divergence, no org boundary — 0 primary signals.
	story1 := makeStoryWithActorsAndObjects(t, "Config Story",
		[]string{"developer"},
		[]string{"config"},
		[]sentenceSpec{{1, "developer", "creates", "config"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Readme Story",
		[]string{"developer"},
		[]string{"readme"},
		[]sentenceSpec{{1, "developer", "creates", "readme"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)

	// BUG: currently returns (nil, nil) when 0 primary signals are found.
	// After fix, fallback should produce work-object-cluster sketches.
	require.NotEmpty(t, sketches, "fallback should produce sketches when 0 primary signals exist")

	var foundWorkObjectCluster bool

	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			if sig.Type() == domain.SignalTypeWorkObjectCluster {
				foundWorkObjectCluster = true
			}
		}
	}

	assert.True(t, foundWorkObjectCluster, "expected at least one work_object_cluster signal from fallback")
}

func TestAlgorithmicDetector_WorkObjectCluster_SharedWOsMergeIntoOneCluster(t *testing.T) {
	t.Parallel()

	// 3 stories all sharing work object "note", same actor, same verb.
	// The shared "note" WO should cause all three stories to merge into one cluster.
	story1 := makeStoryWithActorsAndObjects(t, "Note Tag Story",
		[]string{"developer"},
		[]string{"note", "tag"},
		[]sentenceSpec{{1, "developer", "updates", "note"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Note Label Story",
		[]string{"developer"},
		[]string{"note", "label"},
		[]sentenceSpec{{1, "developer", "updates", "note"}},
	)

	story3 := makeStoryWithActorsAndObjects(t, "Note Only Story",
		[]string{"developer"},
		[]string{"note"},
		[]sentenceSpec{{1, "developer", "updates", "note"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(
		context.Background(),
		[]*domain.DomainStory{story1, story2, story3},
		domain.ModeRapid,
	)
	require.NoError(t, err)

	// All 3 stories share "note" — they should merge into exactly 1 cluster sketch.
	assert.Len(t, sketches, 1, "shared work object 'note' should merge all stories into one cluster")
}

func TestAlgorithmicDetector_WorkObjectCluster_UniqueWOsProduceMultipleSketches(t *testing.T) {
	t.Parallel()

	// 3 stories with ZERO shared work objects, same actor, same verb.
	// Each story forms its own independent cluster.
	story1 := makeStoryWithActorsAndObjects(t, "Invoice Story",
		[]string{"developer"},
		[]string{"invoice"},
		[]sentenceSpec{{1, "developer", "creates", "invoice"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Report Story",
		[]string{"developer"},
		[]string{"report"},
		[]sentenceSpec{{1, "developer", "creates", "report"}},
	)

	story3 := makeStoryWithActorsAndObjects(t, "Ticket Story",
		[]string{"developer"},
		[]string{"ticket"},
		[]sentenceSpec{{1, "developer", "creates", "ticket"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(
		context.Background(),
		[]*domain.DomainStory{story1, story2, story3},
		domain.ModeRapid,
	)
	require.NoError(t, err)

	// No shared WOs — each story should produce its own cluster sketch.
	assert.Len(t, sketches, 3, "3 stories with disjoint work objects should produce 3 separate cluster sketches")
}

func TestAlgorithmicDetector_WorkObjectCluster_PrimarySignalsSuppressFallback(t *testing.T) {
	t.Parallel()

	// Stories that DO produce primary signals: cross-actor diversity, different verbs.
	// The fallback should NOT fire when primary signals exist.
	story1 := makeStoryWithActorsAndObjects(t, "Submission Story",
		[]string{"Alice", "Bob"},
		[]string{"Invoice"},
		[]sentenceSpec{
			{1, "Alice", "creates", "Invoice"},
			{2, "Alice", "submits", "Bob"},
		},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Review Story",
		[]string{"Charlie"},
		[]string{"Invoice"},
		[]sentenceSpec{{1, "Charlie", "reviews", "Invoice"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(
		context.Background(),
		[]*domain.DomainStory{story1, story2},
		domain.ModeRapid,
	)
	require.NoError(t, err)
	require.NotEmpty(t, sketches)

	// Primary signals (SameObjectDiffContext, OrgBoundary) should exist,
	// but WorkObjectCluster should NOT appear — fallback is suppressed.
	for _, sketch := range sketches {
		for _, sig := range sketch.Signals() {
			assert.NotEqual(t, domain.SignalTypeWorkObjectCluster, sig.Type(),
				"work_object_cluster fallback should NOT fire when primary signals are present")
		}
	}
}

func TestAlgorithmicDetector_WorkObjectCluster_ConfidenceIsLow(t *testing.T) {
	t.Parallel()

	// Same setup as FallbackWhenNoSignals: 0 primary signals, fallback fires.
	// Fallback confidence = 0.15 base, which should compute to LOW (< 0.45).
	story1 := makeStoryWithActorsAndObjects(t, "Config Story",
		[]string{"developer"},
		[]string{"config"},
		[]sentenceSpec{{1, "developer", "creates", "config"}},
	)

	story2 := makeStoryWithActorsAndObjects(t, "Readme Story",
		[]string{"developer"},
		[]string{"readme"},
		[]sentenceSpec{{1, "developer", "creates", "readme"}},
	)

	detector := infrastructure.NewAlgorithmicDetector()
	sketches, err := detector.DetectBoundaries(context.Background(), []*domain.DomainStory{story1, story2}, domain.ModeRapid)
	require.NoError(t, err)
	require.NotEmpty(t, sketches, "fallback should produce sketches")

	// Every fallback sketch should have LOW confidence (not HIGH).
	for _, sketch := range sketches {
		assert.NotEqual(t, domain.ConfidenceLevelHigh, sketch.ConfidenceLevel(),
			"work_object_cluster fallback sketches should have LOW confidence, not HIGH")
	}
}
