package infrastructure_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Helper: build a rich DomainStory with actor, work object, sentence,
// annotation, and variation for DTO mapping tests.
// ---------------------------------------------------------------------------

func buildRichTestStory(t *testing.T) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		"Customer Places Order",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"Customer visits online store",
	)
	require.NoError(t, err)

	// Actor.
	customer, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(customer))

	// Work object.
	order, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(order))

	// Sentence.
	sentence, err := discoverydomain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	// Annotation.
	sentenceRef := 1
	annotation, err := discoverydomain.NewAnnotation(
		"Orders must have at least one item",
		discoverydomain.AnnotationTypeConstraint,
		&sentenceRef,
		vo.UserStated,
		"",
	)
	require.NoError(t, err)
	require.NoError(t, story.AddAnnotation(annotation))

	// Variation.
	require.NoError(t, story.AddVariation("Customer cancels order before payment"))

	return story
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewStoryOutput_MapsAllFields(t *testing.T) {
	t.Parallel()

	story := buildRichTestStory(t)
	output := infrastructure.NewStoryOutput("session-123", 1, story)

	assert.Equal(t, "session-123", output.SessionID)
	assert.Equal(t, 1, output.StoryIndex)
	assert.Equal(t, "Customer Places Order", output.Title)
	assert.Equal(t, string(discoverydomain.StoryTypeCoarseGrained), output.StoryType)
	assert.Equal(t, string(discoverydomain.TimeTypeAsIs), output.TimeType)
	assert.Equal(t, string(discoverydomain.PurityTypePure), output.PurityType)
	assert.Equal(t, "Customer visits online store", output.Trigger)

	// Actor.
	require.Len(t, output.Actors, 1)
	assert.Equal(t, "Customer", output.Actors[0].Name)
	assert.Equal(t, string(discoverydomain.ActorTypePerson), output.Actors[0].Type)
	assert.Equal(t, vo.UserStated.String(), output.Actors[0].Trust)

	// Work object.
	require.Len(t, output.WorkObjects, 1)
	assert.Equal(t, "Order", output.WorkObjects[0].Name)
	assert.Equal(t, string(discoverydomain.WorkObjectTypeDocument), output.WorkObjects[0].Type)
	assert.Equal(t, vo.UserStated.String(), output.WorkObjects[0].Trust)

	// Sentence.
	require.Len(t, output.Sentences, 1)
	assert.Equal(t, 1, output.Sentences[0].Step)
	assert.Equal(t, "Customer", output.Sentences[0].Subject)
	assert.Equal(t, "creates", output.Sentences[0].Activity)
	assert.Equal(t, "Order", output.Sentences[0].Object)
	assert.Equal(t, vo.UserStated.String(), output.Sentences[0].Trust)

	// Annotation.
	require.Len(t, output.Annotations, 1)
	assert.Equal(t, "Orders must have at least one item", output.Annotations[0].Text)
	assert.Equal(t, "constraint", output.Annotations[0].Type)
	require.NotNil(t, output.Annotations[0].SentenceRef)
	assert.Equal(t, 1, *output.Annotations[0].SentenceRef)
	assert.Equal(t, vo.UserStated.String(), output.Annotations[0].Trust)

	// Variation.
	require.Len(t, output.Variations, 1)
	assert.Equal(t, "Customer cancels order before payment", output.Variations[0])
}

func TestNewStoryOutput_EmptyAnnotations_OmittedFromJSON(t *testing.T) {
	t.Parallel()

	// Build story without annotations or variations.
	story, err := discoverydomain.NewDomainStory(
		"Simple Story",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("User", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Task", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := discoverydomain.NewStorySentence(1, "User", "creates", "Task", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	output := infrastructure.NewStoryOutput("session-456", 1, story)

	data, err := json.Marshal(output)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.NotContains(t, jsonStr, `"annotations"`)
	assert.NotContains(t, jsonStr, `"variations"`)
}

func TestStoryOutput_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	sentenceRef := 2
	original := infrastructure.StoryOutput{
		SessionID:  "sess-rt",
		StoryIndex: 3,
		Title:      "Round Trip Story",
		StoryType:  "coarse_grained",
		TimeType:   "as_is",
		PurityType: "pure",
		Trigger:    "User opens app",
		Actors: []infrastructure.ActorOutput{
			{Name: "Admin", Type: "person", Trust: "user_stated"},
		},
		WorkObjects: []infrastructure.WorkObjectOutput{
			{Name: "Report", Type: "document", Trust: "user_stated"},
		},
		Sentences: []infrastructure.SentenceOutput{
			{Step: 1, Subject: "Admin", Activity: "generates", Object: "Report", Trust: "user_stated"},
		},
		Annotations: []infrastructure.AnnotationOutput{
			{Text: "Must be PDF", Type: "constraint", SentenceRef: &sentenceRef, Trust: "user_stated"},
		},
		Variations: []string{"Admin exports CSV instead"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored infrastructure.StoryOutput
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.SessionID, restored.SessionID)
	assert.Equal(t, original.StoryIndex, restored.StoryIndex)
	assert.Equal(t, original.Title, restored.Title)
	assert.Equal(t, original.StoryType, restored.StoryType)
	assert.Equal(t, original.TimeType, restored.TimeType)
	assert.Equal(t, original.PurityType, restored.PurityType)
	assert.Equal(t, original.Trigger, restored.Trigger)
	assert.Len(t, restored.Actors, 1)
	assert.Len(t, restored.WorkObjects, 1)
	assert.Len(t, restored.Sentences, 1)
	assert.Len(t, restored.Annotations, 1)
	assert.Len(t, restored.Variations, 1)
}

func TestDiscoveryCompleteOutput_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := infrastructure.DiscoveryCompleteOutput{
		SessionID:   "sess-dc",
		StoryCount:  5,
		SketchCount: 3,
		Mode:        "rapid",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored infrastructure.DiscoveryCompleteOutput
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.SessionID, restored.SessionID)
	assert.Equal(t, original.StoryCount, restored.StoryCount)
	assert.Equal(t, original.SketchCount, restored.SketchCount)
	assert.Equal(t, original.Mode, restored.Mode)
}

func TestBoundaryProposalOutput_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := infrastructure.BoundaryProposalOutput{
		Name:           "OrderManagement",
		Classification: "core",
		Confidence:     0.85,
		Actors:         []string{"Customer", "Clerk"},
		WorkObjects:    []string{"Order", "Invoice"},
		Stories:        []string{"Customer Places Order"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored infrastructure.BoundaryProposalOutput
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Classification, restored.Classification)
	assert.InDelta(t, original.Confidence, restored.Confidence, 0.001)
	assert.Equal(t, original.Actors, restored.Actors)
	assert.Equal(t, original.WorkObjects, restored.WorkObjects)
	assert.Equal(t, original.Stories, restored.Stories)
}
