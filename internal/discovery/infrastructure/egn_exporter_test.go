package infrastructure_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Helper: build a minimal story for Egon tests
// ---------------------------------------------------------------------------

func buildEgnTestStory(t *testing.T, title string) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		title,
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := discoverydomain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	return story
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestEgnExporter_Render_PersonActor(t *testing.T) {
	t.Parallel()

	story := buildEgnTestStory(t, "Person Actor Test")
	exporter := &infrastructure.EgnExporter{}

	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	assert.Contains(t, result, `"domainStory:actorPerson"`)
	assert.Contains(t, result, `"shape_0001"`)
}

func TestEgnExporter_Render_SystemActor(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"System Actor Test",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("PaymentGateway", discoverydomain.ActorTypeSystem, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Transaction", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := discoverydomain.NewStorySentence(1, "PaymentGateway", "processes", "Transaction", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	exporter := &infrastructure.EgnExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	assert.Contains(t, result, `"domainStory:actorSystem"`)
}

func TestEgnExporter_Render_IndirectObject(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Indirect Object Test",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	customer, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(customer))

	seller, err := discoverydomain.NewStoryActor("Seller", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(seller))

	order, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(order))

	sentence, err := discoverydomain.NewStorySentence(1, "Customer", "sends", "Order", vo.UserStated, "")
	require.NoError(t, err)
	sentence, err = sentence.WithPreposition("to", "Seller")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	exporter := &infrastructure.EgnExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Should have 2 connections: primary activity + bridge with null number.
	assert.Contains(t, result, `"domainStory:activity"`)
	assert.Contains(t, result, `"connection_0001"`)
	// Bridge connection should exist.
	assert.Contains(t, result, `"connection_0001b"`)
}

func TestEgnExporter_Render_DomainBlock(t *testing.T) {
	t.Parallel()

	story := buildEgnTestStory(t, "Domain Block Test")
	exporter := &infrastructure.EgnExporter{}

	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Top-level domain block with name, actors, workObjects.
	assert.Contains(t, result, `"domain"`)
	assert.Contains(t, result, `"name"`)
	assert.Contains(t, result, `"actors"`)
	assert.Contains(t, result, `"workObjects"`)
}

func TestEgnExporter_Render_MetadataBlock(t *testing.T) {
	t.Parallel()

	story := buildEgnTestStory(t, "Metadata Block Test")
	exporter := &infrastructure.EgnExporter{}

	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Must have info and version metadata entries in dst.
	assert.Contains(t, result, `"info"`)
	assert.Contains(t, result, `"version"`)
	assert.Contains(t, result, `"2.0.1"`)
}

func TestEgnExporter_Render_MultipleActors(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Multiple Actors Test",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	customer, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(customer))

	clerk, err := discoverydomain.NewStoryActor("Clerk", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(clerk))

	order, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(order))

	s1, err := discoverydomain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(s1))

	s2, err := discoverydomain.NewStorySentence(2, "Clerk", "reviews", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(s2))

	exporter := &infrastructure.EgnExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Consecutive shape IDs.
	assert.Contains(t, result, `"shape_0001"`)
	assert.Contains(t, result, `"shape_0002"`)

	// Both actors should appear as shapes with different x positions.
	assert.Contains(t, result, `"Customer"`)
	assert.Contains(t, result, `"Clerk"`)
}

func TestEgnExporter_Render_NilStory(t *testing.T) {
	t.Parallel()

	exporter := &infrastructure.EgnExporter{}

	_, err := exporter.Render(context.TODO(), nil)

	// Must return error, not panic.
	require.Error(t, err)
}

func TestEgnExporter_Render_JSONValidity(t *testing.T) {
	t.Parallel()

	story := buildEgnTestStory(t, "JSON Validity Test")
	exporter := &infrastructure.EgnExporter{}

	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Output must be valid JSON.
	var parsed map[string]any
	err = json.Unmarshal([]byte(result), &parsed)
	require.NoError(t, err)
}

func TestEgnExporter_Render_AllWorkObjectTypes(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"All Work Object Types",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("User", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	types := []struct {
		name   string
		woType discoverydomain.WorkObjectType
		egnKey string
	}{
		{"MyDocument", discoverydomain.WorkObjectTypeDocument, "domainStory:workObjectDocument"},
		{"MyFolder", discoverydomain.WorkObjectTypeFolder, "domainStory:workObjectFolder"},
		{"MyCall", discoverydomain.WorkObjectTypeCall, "domainStory:workObjectCall"},
		{"MyEmail", discoverydomain.WorkObjectTypeEmail, "domainStory:workObjectEmail"},
		{"MyConversation", discoverydomain.WorkObjectTypeConversation, "domainStory:workObjectConversation"},
		{"MyInfo", discoverydomain.WorkObjectTypeInfo, "domainStory:workObjectInfo"},
		{"MyData", discoverydomain.WorkObjectTypeData, "domainStory:workObjectData"},
	}

	for i, tt := range types {
		wo, woErr := discoverydomain.NewWorkObject(tt.name, tt.woType, vo.UserStated, "")
		require.NoError(t, woErr)
		require.NoError(t, story.AddWorkObject(wo))

		s, sErr := discoverydomain.NewStorySentence(i+1, "User", "uses", tt.name, vo.UserStated, "")
		require.NoError(t, sErr)
		require.NoError(t, story.AddSentence(s))
	}

	exporter := &infrastructure.EgnExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	for _, tt := range types {
		assert.Contains(t, result, tt.egnKey, "missing Egon type for %s", tt.name)
	}
}

func TestEgnExporter_Render_BridgeConnectionIdSuffix(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Bridge Connection ID",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	customer, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(customer))

	seller, err := discoverydomain.NewStoryActor("Seller", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(seller))

	order, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(order))

	sentence, err := discoverydomain.NewStorySentence(1, "Customer", "sends", "Order", vo.UserStated, "")
	require.NoError(t, err)
	sentence, err = sentence.WithPreposition("to", "Seller")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	exporter := &infrastructure.EgnExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Bridge connection IDs must have "b" suffix: "connection_NNNNb".
	assert.Contains(t, result, `"connection_0001b"`)
}
