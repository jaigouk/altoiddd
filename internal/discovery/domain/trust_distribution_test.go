package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// buildTrustTestStory creates a valid DomainStory for trust distribution tests.
func buildTrustTestStory(
	t *testing.T,
	title string,
	actors []domain.StoryActor,
	workObjects []domain.WorkObject,
	sentences []domain.StorySentence,
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

	for _, a := range actors {
		require.NoError(t, story.AddActor(a))
	}

	for _, wo := range workObjects {
		require.NoError(t, story.AddWorkObject(wo))
	}

	for _, s := range sentences {
		require.NoError(t, story.AddSentence(s))
	}

	return story
}

func TestTrustDistribution_Empty_AllZero(t *testing.T) {
	t.Parallel()

	dist := domain.NewTrustDistribution()

	for _, level := range vo.AllTrustLevels() {
		assert.Equal(t, 0, dist.Count(level), "Count(%s) should be 0 for empty distribution", level)
	}

	assert.Equal(t, 0, dist.Total())
}

func TestTrustDistribution_AddStory_CountsActorsWorkObjectsSentences(t *testing.T) {
	t.Parallel()

	// 2 actors (UserStated), 1 work object (AIInferred), 1 sentence (UserStated).
	actor1, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	actor2, err := domain.NewStoryActor("Admin", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIInferred, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTrustTestStory(t, "Order Flow",
		[]domain.StoryActor{actor1, actor2},
		[]domain.WorkObject{wo},
		[]domain.StorySentence{sentence},
	)

	dist := domain.NewTrustDistribution().AddStory(story)

	// 2 actors (UserStated) + 1 sentence (UserStated) = 3 UserStated
	assert.Equal(t, 3, dist.Count(vo.UserStated))
	// 1 work object (AIInferred)
	assert.Equal(t, 1, dist.Count(vo.AIInferred))
	assert.Equal(t, 4, dist.Total())
}

func TestTrustDistribution_AddStory_TwoStories_Accumulates(t *testing.T) {
	t.Parallel()

	actor1, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	wo1, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIInferred, "")
	require.NoError(t, err)

	sentence1, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story1 := buildTrustTestStory(t, "Story1",
		[]domain.StoryActor{actor1},
		[]domain.WorkObject{wo1},
		[]domain.StorySentence{sentence1},
	)

	actor2, err := domain.NewStoryActor("Admin", domain.ActorTypePerson, vo.AIResearched, "research")
	require.NoError(t, err)

	wo2, err := domain.NewWorkObject("Report", domain.WorkObjectTypeDocument, vo.AIResearched, "research")
	require.NoError(t, err)

	sentence2, err := domain.NewStorySentence(1, "Admin", "reviews", "Report", vo.AIResearched, "research")
	require.NoError(t, err)

	story2 := buildTrustTestStory(t, "Story2",
		[]domain.StoryActor{actor2},
		[]domain.WorkObject{wo2},
		[]domain.StorySentence{sentence2},
	)

	dist := domain.NewTrustDistribution().AddStory(story1).AddStory(story2)

	// Story1: 1 UserStated actor + 1 UserStated sentence + 1 AIInferred wo
	// Story2: 1 AIResearched actor + 1 AIResearched wo + 1 AIResearched sentence
	assert.Equal(t, 2, dist.Count(vo.UserStated))
	assert.Equal(t, 1, dist.Count(vo.AIInferred))
	assert.Equal(t, 3, dist.Count(vo.AIResearched))
	assert.Equal(t, 6, dist.Total())
}

func TestTrustDistribution_AddSketch_CountsSketchTrust(t *testing.T) {
	t.Parallel()

	sketch, err := domain.NewBoundedContextSketch(
		"Ordering",
		vo.SubdomainCore,
		0.8,
		[]string{"Customer"},
		[]string{"Order"},
		[]string{"Story1"},
		nil,
		vo.AIResearched,
	)
	require.NoError(t, err)

	dist := domain.NewTrustDistribution().AddSketch(sketch)

	assert.Equal(t, 1, dist.Count(vo.AIResearched))
	assert.Equal(t, 1, dist.Total())
}

func TestTrustDistribution_Total_SumsAllLevels(t *testing.T) {
	t.Parallel()

	// Build a story with mixed trust levels to populate multiple buckets.
	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIInferred, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.AIResearched, "research")
	require.NoError(t, err)

	story := buildTrustTestStory(t, "Mixed Trust",
		[]domain.StoryActor{actor},
		[]domain.WorkObject{wo},
		[]domain.StorySentence{sentence},
	)

	sketch, err := domain.NewBoundedContextSketch(
		"Ordering",
		vo.SubdomainCore,
		0.8,
		[]string{"Customer"},
		[]string{"Order"},
		[]string{"Mixed Trust"},
		nil,
		vo.UserConfirmed,
	)
	require.NoError(t, err)

	dist := domain.NewTrustDistribution().AddStory(story).AddSketch(sketch)

	// 1 UserStated + 1 UserConfirmed + 1 AIResearched + 1 AIInferred = 4
	total := 0
	for _, level := range vo.AllTrustLevels() {
		total += dist.Count(level)
	}

	assert.Equal(t, total, dist.Total())
	assert.Equal(t, 4, dist.Total())
}

func TestTrustDistribution_CopyOnWrite_OriginalUnchanged(t *testing.T) {
	t.Parallel()

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTrustTestStory(t, "COW Test",
		[]domain.StoryActor{actor},
		[]domain.WorkObject{wo},
		[]domain.StorySentence{sentence},
	)

	d1 := domain.NewTrustDistribution()
	d2 := d1.AddStory(story)

	assert.Equal(t, 0, d1.Total(), "original distribution must remain unchanged")
	assert.Positive(t, d2.Total(), "new distribution must have counts")
}

func TestSummarizeStory_Fields(t *testing.T) {
	t.Parallel()

	actor1, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	actor2, err := domain.NewStoryActor("Admin", domain.ActorTypePerson, vo.AIInferred, "")
	require.NoError(t, err)

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	s1, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	s2, err := domain.NewStorySentence(2, "Admin", "reviews", "Order", vo.AIInferred, "")
	require.NoError(t, err)

	story := buildTrustTestStory(t, "Summary Test",
		[]domain.StoryActor{actor1, actor2},
		[]domain.WorkObject{wo},
		[]domain.StorySentence{s1, s2},
	)

	summary := domain.SummarizeStory(story)

	assert.Equal(t, "Summary Test", summary.Title)
	assert.Equal(t, 2, summary.ActorCount)
	assert.Equal(t, 2, summary.SentenceCount)
	assert.Positive(t, summary.Distribution.Total(), "distribution should be non-empty")
}
