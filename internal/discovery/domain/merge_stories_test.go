package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- merge-specific helpers (trust-aware) ---

func mergeActor(t *testing.T, name string, actorType domain.ActorType, trust vo.TrustLevel) domain.StoryActor {
	t.Helper()

	source := ""
	if trust == vo.AIResearched {
		source = "https://example.com"
	}

	a, err := domain.NewStoryActor(name, actorType, trust, source)
	require.NoError(t, err)

	return a
}

func mergeWorkObject(t *testing.T, name string, objectType domain.WorkObjectType, trust vo.TrustLevel) domain.WorkObject {
	t.Helper()

	source := ""
	if trust == vo.AIResearched {
		source = "https://example.com"
	}

	wo, err := domain.NewWorkObject(name, objectType, trust, source)
	require.NoError(t, err)

	return wo
}

func mergeSentence(t *testing.T, step int, subject, activity, object string, trust vo.TrustLevel) domain.StorySentence {
	t.Helper()

	source := ""
	if trust == vo.AIResearched {
		source = "https://example.com"
	}

	s, err := domain.NewStorySentence(step, subject, activity, object, trust, source)
	require.NoError(t, err)

	return s
}

func mergeAnnotation(t *testing.T, text string, annType domain.AnnotationType) domain.Annotation {
	t.Helper()

	a, err := domain.NewAnnotation(text, annType, nil, vo.UserStated, "")
	require.NoError(t, err)

	return a
}

// buildMergeStory creates a valid story with one actor, one work object, one sentence.
func buildMergeStory(t *testing.T, title, trigger, actorName, woName string, trust vo.TrustLevel) *domain.DomainStory {
	t.Helper()

	s, err := domain.NewDomainStory(title, domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, trigger)
	require.NoError(t, err)
	require.NoError(t, s.AddActor(mergeActor(t, actorName, domain.ActorTypePerson, trust)))
	require.NoError(t, s.AddWorkObject(mergeWorkObject(t, woName, domain.WorkObjectTypeDocument, trust)))
	require.NoError(t, s.AddSentence(mergeSentence(t, 1, actorName, "creates", woName, trust)))

	return s
}

// --- tests ---

func TestMergeStories_NilNarratedReturnsError(t *testing.T) {
	t.Parallel()

	proposed := buildMergeStory(t, "Proposed", "trigger", "Alice", "Order", vo.AIResearched)

	_, err := domain.MergeStories(nil, proposed)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestMergeStories_NilProposedReturnsError(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Narrated", "trigger", "Alice", "Order", vo.UserStated)

	_, err := domain.MergeStories(narrated, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestMergeStories_MetadataFromNarrated(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Narrated Title", "narrated trigger", "Alice", "Order", vo.UserStated)

	proposed, err := domain.NewDomainStory("Proposed Title", domain.StoryTypeFineGrained, domain.TimeTypeToBe, domain.PurityTypeDigitalized, "proposed trigger")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "Bob", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "Bob", "reviews", "Invoice", vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	assert.Equal(t, "Narrated Title", merged.Title())
	assert.Equal(t, domain.StoryTypeCoarseGrained, merged.Type())
	assert.Equal(t, domain.TimeTypeAsIs, merged.Time())
	assert.Equal(t, domain.PurityTypePure, merged.Purity())
	assert.Equal(t, "narrated trigger", merged.Trigger())
}

func TestMergeStories_AllNarratedPreserved(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Other", "other trigger", "Bob", "Invoice", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	actors := merged.Actors()
	require.Len(t, actors, 2)
	assert.Equal(t, "Alice", actors[0].Name())
	assert.Equal(t, vo.UserStated, actors[0].Trust())

	wos := merged.WorkObjects()
	require.Len(t, wos, 2)
	assert.Equal(t, "Order", wos[0].Name())
	assert.Equal(t, vo.UserStated, wos[0].Trust())
}

func TestMergeStories_UserActorWinsOnCollision(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)

	// Proposed has same actor "alice" (case-insensitive match) with lower trust.
	proposed := buildMergeStory(t, "Other", "other trigger", "alice", "Invoice", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	actors := merged.Actors()

	var found bool
	for _, a := range actors {
		if a.Name() == "Alice" {
			found = true
			assert.Equal(t, vo.UserStated, a.Trust())
		}
	}

	assert.True(t, found, "narrated Alice should be in merged story")
}

func TestMergeStories_UserWorkObjectWinsOnCollision(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)

	// Proposed has same work object "order" (case-insensitive) with lower trust.
	proposed := buildMergeStory(t, "Other", "other trigger", "Bob", "order", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	wos := merged.WorkObjects()

	var found bool
	for _, wo := range wos {
		if wo.Name() == "Order" {
			found = true
			assert.Equal(t, vo.UserStated, wo.Trust())
		}
	}

	assert.True(t, found, "narrated Order should be in merged story")
}

func TestMergeStories_ProposedActorNotInNarratedIsAdded(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Other", "other trigger", "Bob", "Invoice", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	actors := merged.Actors()
	require.Len(t, actors, 2)

	names := []string{actors[0].Name(), actors[1].Name()}
	assert.Contains(t, names, "Alice")
	assert.Contains(t, names, "Bob")
}

func TestMergeStories_ProposedWorkObjectNotInNarratedIsAdded(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Other", "other trigger", "Bob", "Invoice", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	wos := merged.WorkObjects()
	require.Len(t, wos, 2)

	names := []string{wos[0].Name(), wos[1].Name()}
	assert.Contains(t, names, "Order")
	assert.Contains(t, names, "Invoice")
}

func TestMergeStories_SentencesFromNarratedOnly(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)

	proposed := buildMergeStory(t, "Other", "other trigger", "Alice", "Order", vo.AIResearched)
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 2, "Alice", "deletes", "Order", vo.AIResearched)))

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	sentences := merged.Sentences()
	require.Len(t, sentences, 1)
	assert.Equal(t, "creates", sentences[0].Activity())
}

func TestMergeStories_AnnotationsFromNarratedOnly(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	require.NoError(t, narrated.AddAnnotation(mergeAnnotation(t, "narrated note", domain.AnnotationTypeConstraint)))

	proposed := buildMergeStory(t, "Other", "other trigger", "Alice", "Order", vo.AIResearched)
	require.NoError(t, proposed.AddAnnotation(mergeAnnotation(t, "proposed note", domain.AnnotationTypeAssumption)))

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	annotations := merged.Annotations()
	require.Len(t, annotations, 1)
	assert.Equal(t, "narrated note", annotations[0].Text())
}

func TestMergeStories_VariationsFromNarratedOnly(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	require.NoError(t, narrated.AddVariation("narrated variation"))

	proposed := buildMergeStory(t, "Other", "other trigger", "Alice", "Order", vo.AIResearched)
	require.NoError(t, proposed.AddVariation("proposed variation"))

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	variations := merged.Variations()
	require.Len(t, variations, 1)
	assert.Equal(t, "narrated variation", variations[0])
}

func TestMergeStories_EmptyProposed(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Empty", "empty trigger", "Bot", "Log", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	assert.Equal(t, "Story", merged.Title())
	assert.Len(t, merged.Actors(), 2)
	assert.Len(t, merged.WorkObjects(), 2)
	assert.Len(t, merged.Sentences(), 1)
}

func TestMergeStories_HigherTrustProposedWins(t *testing.T) {
	t.Parallel()

	// Narrated actor has AIResearched trust (lower trust).
	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.AIResearched)

	// Proposed actor "alice" has UserConfirmed trust (higher trust).
	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other trigger")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "alice", domain.ActorTypePerson, vo.UserConfirmed)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "order", domain.WorkObjectTypeDocument, vo.UserConfirmed)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "alice", "reviews", "order", vo.UserConfirmed)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	// Proposed wins because UserConfirmed (2) has higher trust than AIResearched (3).
	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserConfirmed, actors[0].Trust())

	wos := merged.WorkObjects()
	require.Len(t, wos, 1)
	assert.Equal(t, vo.UserConfirmed, wos[0].Trust())
}

// --- QA Trust Comparison Edge Cases ---

// TestMergeStories_QA_Trust_ActorUserStatedVsAIResearched verifies that a narrated
// UserStated actor beats a proposed AIResearched actor on name collision.
func TestMergeStories_QA_Trust_ActorUserStatedVsAIResearched(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Other", "other", "alice", "Invoice", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, "Alice", actors[0].Name(), "narrated name casing preserved")
	assert.Equal(t, vo.UserStated, actors[0].Trust(), "UserStated(1) beats AIResearched(3)")
	assert.Empty(t, actors[0].Source(), "UserStated actors have no source")
}

// TestMergeStories_QA_Trust_ActorUserConfirmedVsAIResearched verifies UserConfirmed wins.
func TestMergeStories_QA_Trust_ActorUserConfirmedVsAIResearched(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserConfirmed)
	proposed := buildMergeStory(t, "Other", "other", "alice", "Invoice", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserConfirmed, actors[0].Trust(), "UserConfirmed(2) beats AIResearched(3)")
}

// TestMergeStories_QA_Trust_ActorAIResearchedVsUserConfirmed verifies proposed
// UserConfirmed replaces narrated AIResearched (trust inversion — proposed wins).
func TestMergeStories_QA_Trust_ActorAIResearchedVsUserConfirmed(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.AIResearched)

	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "alice", domain.ActorTypePerson, vo.UserConfirmed)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "alice", "reviews", "Invoice", vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserConfirmed, actors[0].Trust(), "proposed UserConfirmed(2) beats narrated AIResearched(3)")
	assert.Equal(t, "alice", actors[0].Name(), "proposed version replaces narrated entirely")
}

// TestMergeStories_QA_Trust_ActorAIResearchedVsUserStated verifies proposed
// UserStated replaces narrated AIResearched.
func TestMergeStories_QA_Trust_ActorAIResearchedVsUserStated(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.AIResearched)

	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "alice", domain.ActorTypePerson, vo.UserStated)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "alice", "reviews", "Invoice", vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserStated, actors[0].Trust(), "proposed UserStated(1) beats narrated AIResearched(3)")
}

// TestMergeStories_QA_Trust_ActorSameTrustNarratedWins verifies narrated precedence
// when both sides have UserStated (same trust level).
func TestMergeStories_QA_Trust_ActorSameTrustNarratedWins(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Other", "other", "alice", "Invoice", vo.UserStated)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, "Alice", actors[0].Name(), "narrated version kept on equal trust (name casing proves identity)")
	assert.Equal(t, vo.UserStated, actors[0].Trust())
}

// TestMergeStories_QA_Trust_ActorSameTrustUserConfirmedNarratedWins verifies
// narrated wins when both have UserConfirmed.
func TestMergeStories_QA_Trust_ActorSameTrustUserConfirmedNarratedWins(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserConfirmed)
	proposed := buildMergeStory(t, "Other", "other", "alice", "Invoice", vo.UserConfirmed)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, "Alice", actors[0].Name(), "narrated version kept on equal UserConfirmed trust")
}

// TestMergeStories_QA_Trust_WorkObjectUserStatedVsAIResearched mirrors actor test for WOs.
func TestMergeStories_QA_Trust_WorkObjectUserStatedVsAIResearched(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Other", "other", "Bob", "order", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	wos := merged.WorkObjects()
	var order domain.WorkObject
	for _, wo := range wos {
		if wo.Name() == "Order" || wo.Name() == "order" {
			order = wo

			break
		}
	}

	assert.Equal(t, "Order", order.Name(), "narrated name casing preserved")
	assert.Equal(t, vo.UserStated, order.Trust(), "UserStated(1) beats AIResearched(3)")
}

// TestMergeStories_QA_Trust_WorkObjectUserConfirmedVsAIResearched verifies WO trust.
func TestMergeStories_QA_Trust_WorkObjectUserConfirmedVsAIResearched(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserConfirmed)
	proposed := buildMergeStory(t, "Other", "other", "Bob", "order", vo.AIResearched)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	wos := merged.WorkObjects()
	var order domain.WorkObject
	for _, wo := range wos {
		if wo.Name() == "Order" || wo.Name() == "order" {
			order = wo

			break
		}
	}

	assert.Equal(t, vo.UserConfirmed, order.Trust(), "UserConfirmed(2) beats AIResearched(3)")
}

// TestMergeStories_QA_Trust_WorkObjectAIResearchedVsUserConfirmed verifies proposed
// UserConfirmed WO replaces narrated AIResearched WO.
func TestMergeStories_QA_Trust_WorkObjectAIResearchedVsUserConfirmed(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.AIResearched)

	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "Bob", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "order", domain.WorkObjectTypeDocument, vo.UserConfirmed)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "Bob", "reviews", "order", vo.UserConfirmed)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	wos := merged.WorkObjects()
	var order domain.WorkObject
	for _, wo := range wos {
		if wo.Name() == "Order" || wo.Name() == "order" {
			order = wo

			break
		}
	}

	assert.Equal(t, vo.UserConfirmed, order.Trust(), "proposed UserConfirmed(2) beats narrated AIResearched(3)")
	assert.Equal(t, "order", order.Name(), "proposed version replaces narrated entirely")
}

// TestMergeStories_QA_Trust_WorkObjectSameTrustNarratedWins verifies narrated WO
// wins on same trust level.
func TestMergeStories_QA_Trust_WorkObjectSameTrustNarratedWins(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	proposed := buildMergeStory(t, "Other", "other", "Bob", "order", vo.UserStated)

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	wos := merged.WorkObjects()
	var order domain.WorkObject
	for _, wo := range wos {
		if wo.Name() == "Order" || wo.Name() == "order" {
			order = wo

			break
		}
	}

	assert.Equal(t, "Order", order.Name(), "narrated version kept on equal trust")
}

// TestMergeStories_QA_Trust_MixedActorsOverlapAndNovel tests 3 narrated + 3 proposed,
// 1 overlap (narrated wins), 2 novel → merged has 5 actors.
func TestMergeStories_QA_Trust_MixedActorsOverlapAndNovel(t *testing.T) {
	t.Parallel()

	narrated, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)
	require.NoError(t, narrated.AddActor(mergeActor(t, "Alice", domain.ActorTypePerson, vo.UserStated)))
	require.NoError(t, narrated.AddActor(mergeActor(t, "Bob", domain.ActorTypeSystem, vo.UserConfirmed)))
	require.NoError(t, narrated.AddActor(mergeActor(t, "Carol", domain.ActorTypeGroup, vo.UserStated)))
	require.NoError(t, narrated.AddWorkObject(mergeWorkObject(t, "Order", domain.WorkObjectTypeDocument, vo.UserStated)))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 1, "Alice", "creates", "Order", vo.UserStated)))

	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	// Overlap: alice (case-insensitive match) with lower trust
	require.NoError(t, proposed.AddActor(mergeActor(t, "alice", domain.ActorTypePerson, vo.AIResearched)))
	// Novel actors
	require.NoError(t, proposed.AddActor(mergeActor(t, "Dave", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, proposed.AddActor(mergeActor(t, "Eve", domain.ActorTypeSystem, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "alice", "reviews", "Invoice", vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	actors := merged.Actors()
	require.Len(t, actors, 5, "3 narrated + 2 novel proposed = 5 (1 overlap)")

	// Narrated actors preserve position and trust.
	assert.Equal(t, "Alice", actors[0].Name())
	assert.Equal(t, vo.UserStated, actors[0].Trust(), "overlap: narrated UserStated wins over proposed AIResearched")
	assert.Equal(t, "Bob", actors[1].Name())
	assert.Equal(t, "Carol", actors[2].Name())

	// Novel proposed actors appended.
	assert.Equal(t, "Dave", actors[3].Name())
	assert.Equal(t, vo.AIResearched, actors[3].Trust())
	assert.Equal(t, "Eve", actors[4].Name())
	assert.Equal(t, vo.AIResearched, actors[4].Trust())
}

// TestMergeStories_QA_Trust_MixedTrustInversionDifferentNames verifies that when
// narrated has AIResearched and proposed has UserConfirmed with DIFFERENT names,
// both are present — proposed keeps its higher trust.
func TestMergeStories_QA_Trust_MixedTrustInversionDifferentNames(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.AIResearched)

	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "Bob", domain.ActorTypePerson, vo.UserConfirmed)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.UserConfirmed)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "Bob", "reviews", "Invoice", vo.UserConfirmed)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	actors := merged.Actors()
	require.Len(t, actors, 2, "different names: both present")

	assert.Equal(t, "Alice", actors[0].Name())
	assert.Equal(t, vo.AIResearched, actors[0].Trust(), "narrated keeps AIResearched")
	assert.NotEmpty(t, actors[0].Source(), "AIResearched actors retain source")

	assert.Equal(t, "Bob", actors[1].Name())
	assert.Equal(t, vo.UserConfirmed, actors[1].Trust(), "proposed keeps UserConfirmed")
}

// TestMergeStories_QA_Trust_NoWithTrustCalled verifies that merge selects the entire
// VO rather than calling WithTrust to upgrade in-place. We prove this by checking
// that when proposed wins, the full proposed VO identity is used (name casing, type),
// not just the trust field upgraded on the narrated VO.
func TestMergeStories_QA_Trust_NoWithTrustCalled(t *testing.T) {
	t.Parallel()

	// Narrated: actor "ALICE" as person, AIResearched.
	narrated, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)
	require.NoError(t, narrated.AddActor(mergeActor(t, "ALICE", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, narrated.AddWorkObject(mergeWorkObject(t, "ORDER", domain.WorkObjectTypeDocument, vo.AIResearched)))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 1, "ALICE", "creates", "ORDER", vo.AIResearched)))

	// Proposed: actor "alice" as system, UserStated — different type AND casing.
	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "alice", domain.ActorTypeSystem, vo.UserStated)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "order", domain.WorkObjectTypeFolder, vo.UserStated)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "alice", "reviews", "order", vo.UserStated)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	// If WithTrust were used, trust would change but name casing and type would stay "ALICE"/"person".
	// The entire proposed VO replaces narrated, so we get "alice"/"system".
	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, "alice", actors[0].Name(), "proposed VO replaces entirely — name casing from proposed")
	assert.Equal(t, domain.ActorTypeSystem, actors[0].Type(), "proposed VO replaces entirely — type from proposed")
	assert.Equal(t, vo.UserStated, actors[0].Trust())

	wos := merged.WorkObjects()
	require.Len(t, wos, 1)
	assert.Equal(t, "order", wos[0].Name(), "proposed VO replaces entirely — name casing from proposed")
	assert.Equal(t, domain.WorkObjectTypeFolder, wos[0].Type(), "proposed VO replaces entirely — type from proposed")
	assert.Equal(t, vo.UserStated, wos[0].Trust())
}

// --- QA Merge Rules Edge Cases (TestMergeStories_QA_Rules_*) ---

func TestMergeStories_QA_Rules_BothNil(t *testing.T) {
	t.Parallel()

	_, err := domain.MergeStories(nil, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestMergeStories_QA_Rules_NilNarratedErrorMessage(t *testing.T) {
	t.Parallel()

	proposed := buildMergeStory(t, "P", "trigger", "A", "O", vo.AIResearched)

	_, err := domain.MergeStories(nil, proposed)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.ErrorContains(t, err, "narrated")
}

func TestMergeStories_QA_Rules_NilProposedErrorMessage(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "N", "trigger", "A", "O", vo.UserStated)

	_, err := domain.MergeStories(narrated, nil)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.ErrorContains(t, err, "proposed")
}

func TestMergeStories_QA_Rules_EmptyProposedMergedEqualsNarrated(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	require.NoError(t, narrated.AddAnnotation(mergeAnnotation(t, "constraint note", domain.AnnotationTypeConstraint)))
	require.NoError(t, narrated.AddVariation("alt flow"))

	// Proposed with no actors, WOs, sentences, annotations, or variations.
	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	assert.Len(t, merged.Actors(), 1)
	assert.Equal(t, "Alice", merged.Actors()[0].Name())
	assert.Len(t, merged.WorkObjects(), 1)
	assert.Equal(t, "Order", merged.WorkObjects()[0].Name())
	assert.Len(t, merged.Sentences(), 1)
	assert.Len(t, merged.Annotations(), 1)
	assert.Equal(t, "constraint note", merged.Annotations()[0].Text())
	assert.Len(t, merged.Variations(), 1)
	assert.Equal(t, "alt flow", merged.Variations()[0])
}

func TestMergeStories_QA_Rules_FullOverlapSameCount(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	require.NoError(t, narrated.AddActor(mergeActor(t, "Bob", domain.ActorTypePerson, vo.UserStated)))
	require.NoError(t, narrated.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.UserStated)))

	// Proposed has exact same names with lower trust.
	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "Alice", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, proposed.AddActor(mergeActor(t, "Bob", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Order", domain.WorkObjectTypeDocument, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	// Same count as narrated — no novel elements added.
	assert.Len(t, merged.Actors(), 2)
	assert.Len(t, merged.WorkObjects(), 2)

	// All narrated versions preserved (higher trust).
	for _, a := range merged.Actors() {
		assert.Equal(t, vo.UserStated, a.Trust())
	}

	for _, wo := range merged.WorkObjects() {
		assert.Equal(t, vo.UserStated, wo.Trust())
	}
}

func TestMergeStories_QA_Rules_NoOverlapAllAdded(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	require.NoError(t, narrated.AddActor(mergeActor(t, "Bob", domain.ActorTypePerson, vo.UserStated)))

	// Proposed has entirely different names.
	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "Charlie", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, proposed.AddActor(mergeActor(t, "Diana", domain.ActorTypeSystem, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Receipt", domain.WorkObjectTypeData, vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	// narrated(2 actors) + proposed(2 actors) = 4.
	assert.Len(t, merged.Actors(), 4)
	// narrated(1 WO) + proposed(2 WOs) = 3.
	assert.Len(t, merged.WorkObjects(), 3)

	actorNames := make([]string, len(merged.Actors()))
	for i, a := range merged.Actors() {
		actorNames[i] = a.Name()
	}

	assert.Contains(t, actorNames, "Alice")
	assert.Contains(t, actorNames, "Bob")
	assert.Contains(t, actorNames, "Charlie")
	assert.Contains(t, actorNames, "Diana")
}

func TestMergeStories_QA_Rules_CaseMismatchNarratedNamePreserved(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Pet Owner", "Order Form", vo.UserStated)

	// Proposed uses different casing.
	proposed, err := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, err)
	require.NoError(t, proposed.AddActor(mergeActor(t, "pet owner", domain.ActorTypePerson, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "order form", domain.WorkObjectTypeDocument, vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	// Exactly 1 actor, narrated name casing preserved.
	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, "Pet Owner", actors[0].Name())
	assert.Equal(t, vo.UserStated, actors[0].Trust())

	// Exactly 1 WO, narrated name casing preserved.
	wos := merged.WorkObjects()
	require.Len(t, wos, 1)
	assert.Equal(t, "Order Form", wos[0].Name())
}

func TestMergeStories_QA_Rules_AllMetadataFromNarrated(t *testing.T) {
	t.Parallel()

	// Narrated uses non-default enum values to distinguish from proposed.
	narrated, err := domain.NewDomainStory(
		"Narrated Title",
		domain.StoryTypeFineGrained,
		domain.TimeTypeToBe,
		domain.PurityTypeDigitalized,
		"narrated trigger",
	)
	require.NoError(t, err)
	require.NoError(t, narrated.AddActor(mergeActor(t, "Alice", domain.ActorTypePerson, vo.UserStated)))
	require.NoError(t, narrated.AddWorkObject(mergeWorkObject(t, "Order", domain.WorkObjectTypeDocument, vo.UserStated)))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 1, "Alice", "creates", "Order", vo.UserStated)))

	proposed, propErr := domain.NewDomainStory(
		"Proposed Title",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"proposed trigger",
	)
	require.NoError(t, propErr)

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	assert.Equal(t, "Narrated Title", merged.Title())
	assert.Equal(t, domain.StoryTypeFineGrained, merged.Type())
	assert.Equal(t, domain.TimeTypeToBe, merged.Time())
	assert.Equal(t, domain.PurityTypeDigitalized, merged.Purity())
	assert.Equal(t, "narrated trigger", merged.Trigger())
}

func TestMergeStories_QA_Rules_SentencesNarratedOnlyExactCount(t *testing.T) {
	t.Parallel()

	// Narrated with 2 sentences.
	narrated, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)
	require.NoError(t, narrated.AddActor(mergeActor(t, "Alice", domain.ActorTypePerson, vo.UserStated)))
	require.NoError(t, narrated.AddWorkObject(mergeWorkObject(t, "Order", domain.WorkObjectTypeDocument, vo.UserStated)))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 1, "Alice", "creates", "Order", vo.UserStated)))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 2, "Alice", "reviews", "Order", vo.UserStated)))

	// Proposed with 5 sentences.
	proposed, propErr := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, propErr)
	require.NoError(t, proposed.AddActor(mergeActor(t, "Bot", domain.ActorTypeSystem, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Log", domain.WorkObjectTypeData, vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 1, "Bot", "reads", "Log", vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 2, "Bot", "parses", "Log", vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 3, "Bot", "validates", "Log", vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 4, "Bot", "archives", "Log", vo.AIResearched)))
	require.NoError(t, proposed.AddSentence(mergeSentence(t, 5, "Bot", "deletes", "Log", vo.AIResearched)))

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	// Exactly 2 sentences from narrated, none from proposed.
	require.Len(t, merged.Sentences(), 2)
	assert.Equal(t, "creates", merged.Sentences()[0].Activity())
	assert.Equal(t, "reviews", merged.Sentences()[1].Activity())
}

func TestMergeStories_QA_Rules_AnnotationsNarratedOnlyExactCount(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	require.NoError(t, narrated.AddAnnotation(mergeAnnotation(t, "narrated ann", domain.AnnotationTypeConstraint)))

	proposed := buildMergeStory(t, "Other", "other", "Alice", "Order", vo.AIResearched)
	require.NoError(t, proposed.AddAnnotation(mergeAnnotation(t, "proposed ann 1", domain.AnnotationTypeAssumption)))
	require.NoError(t, proposed.AddAnnotation(mergeAnnotation(t, "proposed ann 2", domain.AnnotationTypeInvariant)))
	require.NoError(t, proposed.AddAnnotation(mergeAnnotation(t, "proposed ann 3", domain.AnnotationTypeConstraint)))

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	// Exactly 1 annotation from narrated, none from proposed.
	require.Len(t, merged.Annotations(), 1)
	assert.Equal(t, "narrated ann", merged.Annotations()[0].Text())
}

func TestMergeStories_QA_Rules_VariationsNarratedOnlyExactCount(t *testing.T) {
	t.Parallel()

	narrated := buildMergeStory(t, "Story", "trigger", "Alice", "Order", vo.UserStated)
	require.NoError(t, narrated.AddVariation("narrated var"))

	proposed := buildMergeStory(t, "Other", "other", "Alice", "Order", vo.AIResearched)
	require.NoError(t, proposed.AddVariation("proposed var 1"))
	require.NoError(t, proposed.AddVariation("proposed var 2"))

	merged, err := domain.MergeStories(narrated, proposed)
	require.NoError(t, err)

	// Exactly 1 variation from narrated, none from proposed.
	require.Len(t, merged.Variations(), 1)
	assert.Equal(t, "narrated var", merged.Variations()[0])
}

func TestMergeStories_QA_Rules_ValidatePassesOnMerged(t *testing.T) {
	t.Parallel()

	// Narrated with 2 actors, 2 WOs, 2 sentences, and a sentence-scoped annotation.
	narrated, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)
	require.NoError(t, narrated.AddActor(mergeActor(t, "Alice", domain.ActorTypePerson, vo.UserStated)))
	require.NoError(t, narrated.AddActor(mergeActor(t, "Bob", domain.ActorTypePerson, vo.UserStated)))
	require.NoError(t, narrated.AddWorkObject(mergeWorkObject(t, "Order", domain.WorkObjectTypeDocument, vo.UserStated)))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 1, "Alice", "creates", "Order", vo.UserStated)))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 2, "Bob", "reviews", "Order", vo.UserStated)))

	ref := 1
	ann, annErr := domain.NewAnnotation("constraint note", domain.AnnotationTypeConstraint, &ref, vo.UserStated, "")
	require.NoError(t, annErr)
	require.NoError(t, narrated.AddAnnotation(ann))

	// Proposed adds novel actor+WO.
	proposed, propErr := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, propErr)
	require.NoError(t, proposed.AddActor(mergeActor(t, "Charlie", domain.ActorTypeSystem, vo.AIResearched)))
	require.NoError(t, proposed.AddWorkObject(mergeWorkObject(t, "Invoice", domain.WorkObjectTypeData, vo.AIResearched)))

	// MergeStories calls Validate() internally — no error means it passes.
	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	assert.Len(t, merged.Actors(), 3)
	assert.Len(t, merged.Sentences(), 2)
	assert.Len(t, merged.Annotations(), 1)
}

func TestMergeStories_QA_Rules_SourceStringPreserved(t *testing.T) {
	t.Parallel()

	sourceURL := "https://example.com/research/pet-care"

	actor, err := domain.NewStoryActor("AI Bot", domain.ActorTypeSystem, vo.AIResearched, sourceURL)
	require.NoError(t, err)

	wo, woErr := domain.NewWorkObject("Research Report", domain.WorkObjectTypeDocument, vo.AIResearched, sourceURL)
	require.NoError(t, woErr)

	narrated, storyErr := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, storyErr)
	require.NoError(t, narrated.AddActor(actor))
	require.NoError(t, narrated.AddWorkObject(wo))
	require.NoError(t, narrated.AddSentence(mergeSentence(t, 1, "AI Bot", "generates", "Research Report", vo.UserStated)))

	// Empty proposed.
	proposed, propErr := domain.NewDomainStory("Other", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "other")
	require.NoError(t, propErr)

	merged, mergeErr := domain.MergeStories(narrated, proposed)
	require.NoError(t, mergeErr)

	actors := merged.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, "AI Bot", actors[0].Name())
	assert.Equal(t, vo.AIResearched, actors[0].Trust())
	assert.Equal(t, sourceURL, actors[0].Source())

	wos := merged.WorkObjects()
	require.Len(t, wos, 1)
	assert.Equal(t, "Research Report", wos[0].Name())
	assert.Equal(t, sourceURL, wos[0].Source())
}
