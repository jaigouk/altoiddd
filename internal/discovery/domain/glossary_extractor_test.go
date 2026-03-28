package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// buildTestStory creates a valid DomainStory with actors, work objects, and sentences.
func buildTestStory(
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

func TestGlossaryExtractor_Extract_EmptyStories(t *testing.T) {
	t.Parallel()

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract(nil, nil)

	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestGlossaryExtractor_Extract_SingleStory_ActorsAndWorkObjects(t *testing.T) {
	t.Parallel()

	customer, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	system, err := domain.NewStoryActor("System", domain.ActorTypeSystem, vo.UserStated, "")
	require.NoError(t, err)

	order, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	invoice, err := domain.NewWorkObject("Invoice", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTestStory(t, "Order Flow",
		[]domain.StoryActor{customer, system},
		[]domain.WorkObject{order, invoice},
		[]domain.StorySentence{sentence},
	)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story}, nil)

	require.NoError(t, err)
	// 2 actors + 2 work objects + 1 activity from sentence = 5 entries.
	assert.Len(t, entries, 5)

	// Verify actor definitions follow pattern.
	entryMap := make(map[string]vo.UbiquitousLanguageEntry, len(entries))
	for _, e := range entries {
		entryMap[e.Term()] = e
	}

	customerEntry, ok := entryMap["Customer"]
	require.True(t, ok, "expected entry for Customer")
	assert.Contains(t, customerEntry.Definition(), "person")
	assert.Contains(t, customerEntry.Definition(), "Order Flow")

	systemEntry, ok := entryMap["System"]
	require.True(t, ok, "expected entry for System")
	assert.Contains(t, systemEntry.Definition(), "system")
	assert.Contains(t, systemEntry.Definition(), "Order Flow")

	// Work object definitions follow pattern.
	orderEntry, ok := entryMap["Order"]
	require.True(t, ok, "expected entry for Order")
	assert.Contains(t, orderEntry.Definition(), "document")
	assert.Contains(t, orderEntry.Definition(), "Order Flow")

	// All contexts are "General" when contextMap is nil.
	for _, e := range entries {
		assert.Equal(t, "General", e.Context(), "expected General context for %s", e.Term())
	}

	// Each entry has story ref matching the story title.
	for _, e := range entries {
		assert.Contains(t, e.Stories(), "Order Flow", "expected story ref for %s", e.Term())
	}
}

func TestGlossaryExtractor_Extract_DeduplicateAcrossStories(t *testing.T) {
	t.Parallel()

	customer1, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	order1, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence1, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story1 := buildTestStory(t, "Place Order",
		[]domain.StoryActor{customer1},
		[]domain.WorkObject{order1},
		[]domain.StorySentence{sentence1},
	)

	customer2, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	invoice, err := domain.NewWorkObject("Invoice", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence2, err := domain.NewStorySentence(1, "Customer", "receives", "Invoice", vo.UserStated, "")
	require.NoError(t, err)

	story2 := buildTestStory(t, "Pay Invoice",
		[]domain.StoryActor{customer2},
		[]domain.WorkObject{invoice},
		[]domain.StorySentence{sentence2},
	)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story1, story2}, nil)

	require.NoError(t, err)

	// Count entries for "Customer" — should be exactly 1 (deduplicated).
	customerCount := 0
	var customerEntry vo.UbiquitousLanguageEntry

	for _, e := range entries {
		if e.Term() == "Customer" {
			customerCount++
			customerEntry = e
		}
	}

	assert.Equal(t, 1, customerCount, "Customer should appear exactly once")
	// Merged entry has BOTH story titles.
	assert.Contains(t, customerEntry.Stories(), "Place Order")
	assert.Contains(t, customerEntry.Stories(), "Pay Invoice")
}

func TestGlossaryExtractor_Extract_TrustMerge_HigherTrustWins(t *testing.T) {
	t.Parallel()

	customerLow, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIInferred, "")
	require.NoError(t, err)

	order1, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIInferred, "")
	require.NoError(t, err)

	sentence1, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.AIInferred, "")
	require.NoError(t, err)

	story1 := buildTestStory(t, "Story1",
		[]domain.StoryActor{customerLow},
		[]domain.WorkObject{order1},
		[]domain.StorySentence{sentence1},
	)

	customerHigh, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	order2, err := domain.NewWorkObject("Ticket", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence2, err := domain.NewStorySentence(1, "Customer", "views", "Ticket", vo.UserStated, "")
	require.NoError(t, err)

	story2 := buildTestStory(t, "Story2",
		[]domain.StoryActor{customerHigh},
		[]domain.WorkObject{order2},
		[]domain.StorySentence{sentence2},
	)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story1, story2}, nil)

	require.NoError(t, err)

	// Find merged Customer entry.
	for _, e := range entries {
		if e.Term() == "Customer" {
			// Higher trust (UserStated=1) wins over AIInferred(=4).
			assert.Equal(t, vo.UserStated, e.Trust())

			return
		}
	}

	t.Fatal("expected to find Customer entry")
}

func TestGlossaryExtractor_Extract_ContextAssignment_FromContextMap(t *testing.T) {
	t.Parallel()

	customer, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	order, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Customer", "places", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTestStory(t, "Order Flow",
		[]domain.StoryActor{customer},
		[]domain.WorkObject{order},
		[]domain.StorySentence{sentence},
	)

	// Build a BoundedContextSketch with actors and workObjects that include ours.
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering",
		vo.SubdomainCore,
		0.8,
		[]string{"Customer"},
		[]string{"Order"},
		[]string{"Order Flow"},
		nil,
		vo.AIInferred,
	)
	require.NoError(t, err)

	contextMap, err := domain.NewContextMap("TestProject", []domain.BoundedContextSketch{sketch}, nil)
	require.NoError(t, err)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story}, contextMap)

	require.NoError(t, err)

	entryMap := make(map[string]vo.UbiquitousLanguageEntry, len(entries))
	for _, e := range entries {
		entryMap[e.Term()] = e
	}

	customerEntry, ok := entryMap["Customer"]
	require.True(t, ok, "expected Customer entry")
	assert.Equal(t, "Ordering", customerEntry.Context())

	orderEntry, ok := entryMap["Order"]
	require.True(t, ok, "expected Order entry")
	assert.Equal(t, "Ordering", orderEntry.Context())
}

func TestGlossaryExtractor_Extract_ContextAssignment_FallsBackToGeneral(t *testing.T) {
	t.Parallel()

	admin, err := domain.NewStoryActor("Admin", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	report, err := domain.NewWorkObject("Report", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Admin", "reviews", "Report", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTestStory(t, "Admin Flow",
		[]domain.StoryActor{admin},
		[]domain.WorkObject{report},
		[]domain.StorySentence{sentence},
	)

	// Sketch does NOT include "Admin".
	sketch, err := domain.NewBoundedContextSketch(
		"Ordering",
		vo.SubdomainCore,
		0.8,
		[]string{"Customer"},
		[]string{"Order"},
		[]string{"Order Flow"},
		nil,
		vo.AIInferred,
	)
	require.NoError(t, err)

	contextMap, err := domain.NewContextMap("TestProject", []domain.BoundedContextSketch{sketch}, nil)
	require.NoError(t, err)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story}, contextMap)

	require.NoError(t, err)

	for _, e := range entries {
		if e.Term() == "Admin" {
			assert.Equal(t, "General", e.Context())

			return
		}
	}

	t.Fatal("expected to find Admin entry")
}

func TestGlossaryExtractor_Extract_NilContextMap(t *testing.T) {
	t.Parallel()

	customer, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	order, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTestStory(t, "Order Flow",
		[]domain.StoryActor{customer},
		[]domain.WorkObject{order},
		[]domain.StorySentence{sentence},
	)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story}, nil)

	require.NoError(t, err)

	for _, e := range entries {
		assert.Equal(t, "General", e.Context(), "all entries should have General context when contextMap is nil, but %s has %s", e.Term(), e.Context())
	}
}

func TestGlossaryExtractor_Extract_SentenceActivitiesExtracted(t *testing.T) {
	t.Parallel()

	customer, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	order, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTestStory(t, "Order Flow",
		[]domain.StoryActor{customer},
		[]domain.WorkObject{order},
		[]domain.StorySentence{sentence},
	)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story}, nil)

	require.NoError(t, err)

	// Find the activity entry.
	found := false

	for _, e := range entries {
		if e.Term() == "submits" {
			found = true
			assert.Contains(t, e.Definition(), "Activity")
			assert.Contains(t, e.Definition(), "Customer")
			assert.Contains(t, e.Definition(), "Order")

			break
		}
	}

	assert.True(t, found, "expected an entry for the activity 'submits'")
}

func TestGlossaryExtractor_Extract_SortedOutput(t *testing.T) {
	t.Parallel()

	zebra, err := domain.NewStoryActor("Zebra", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	alpha, err := domain.NewStoryActor("Alpha", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	middle, err := domain.NewStoryActor("Middle", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)

	item, err := domain.NewWorkObject("Item", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err := domain.NewStorySentence(1, "Alpha", "processes", "Item", vo.UserStated, "")
	require.NoError(t, err)

	story := buildTestStory(t, "Sort Test",
		[]domain.StoryActor{zebra, alpha, middle},
		[]domain.WorkObject{item},
		[]domain.StorySentence{sentence},
	)

	extractor := domain.GlossaryExtractor{}
	entries, err := extractor.Extract([]*domain.DomainStory{story}, nil)

	require.NoError(t, err)
	require.NotEmpty(t, entries)

	// Verify lexicographic ordering by Term().
	for i := 1; i < len(entries); i++ {
		prev := entries[i-1].Term()
		curr := entries[i].Term()
		assert.LessOrEqual(t, prev, curr,
			"entries not sorted: %q should come before %q", prev, curr)
	}
}
