package infrastructure_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Helper: build a full ecommerce story for PlantUML tests
// ---------------------------------------------------------------------------

func buildPlantUMLEcommerceStory(t *testing.T) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		"E-commerce: Customer Purchases Product",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"Customer visits online store",
	)
	require.NoError(t, err)

	// Actors.
	customer, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(customer))

	paymentGW, err := discoverydomain.NewStoryActor("PaymentGateway", discoverydomain.ActorTypeSystem, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(paymentGW))

	// Work objects.
	productListing, err := discoverydomain.NewWorkObject("ProductListing", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(productListing))

	shoppingCart, err := discoverydomain.NewWorkObject("ShoppingCart", discoverydomain.WorkObjectTypeFolder, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(shoppingCart))

	order, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(order))

	// Sentences.
	s1, err := discoverydomain.NewStorySentence(1, "Customer", "browses", "ProductListing", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(s1))

	s2, err := discoverydomain.NewStorySentence(2, "Customer", "adds", "ProductListing", vo.UserStated, "")
	require.NoError(t, err)
	s2, err = s2.WithPreposition("to", "ShoppingCart")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(s2))

	s3, err := discoverydomain.NewStorySentence(3, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(s3))

	return story
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPlantUMLExporter_Render_HappyPath_EcommerceStory(t *testing.T) {
	t.Parallel()

	story := buildPlantUMLEcommerceStory(t)
	exporter := &infrastructure.PlantUMLExporter{}

	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Must contain PlantUML markers.
	assert.Contains(t, result, "@startuml")
	assert.Contains(t, result, "@enduml")

	// Must contain actors.
	assert.Contains(t, result, "Person(Customer)")
	assert.Contains(t, result, "System(PaymentGateway")

	// Must contain work objects.
	assert.Contains(t, result, "Document(Order)")

	// Must contain activity lines.
	assert.Contains(t, result, "activity(1,")
	assert.Contains(t, result, "activity(2,")
	assert.Contains(t, result, "activity(3,")
}

func TestPlantUMLExporter_Render_GroupActor(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Pharmacy Group Test",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("Pharmacy", discoverydomain.ActorTypeGroup, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Prescription", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := discoverydomain.NewStorySentence(1, "Pharmacy", "processes", "Prescription", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	exporter := &infrastructure.PlantUMLExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	assert.Contains(t, result, "Group(Pharmacy)")
}

func TestPlantUMLExporter_Render_AllWorkObjectTypes(t *testing.T) {
	t.Parallel()

	// Build story with all 7 work object types.
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

	typeMap := map[string]discoverydomain.WorkObjectType{
		"MyDocument":     discoverydomain.WorkObjectTypeDocument,
		"MyFolder":       discoverydomain.WorkObjectTypeFolder,
		"MyCall":         discoverydomain.WorkObjectTypeCall,
		"MyEmail":        discoverydomain.WorkObjectTypeEmail,
		"MyConversation": discoverydomain.WorkObjectTypeConversation,
		"MyInfo":         discoverydomain.WorkObjectTypeInfo,
		"MyData":         discoverydomain.WorkObjectTypeData,
	}

	step := 1
	for name, woType := range typeMap {
		wo, woErr := discoverydomain.NewWorkObject(name, woType, vo.UserStated, "")
		require.NoError(t, woErr)
		require.NoError(t, story.AddWorkObject(wo))

		s, sErr := discoverydomain.NewStorySentence(step, "User", "uses", name, vo.UserStated, "")
		require.NoError(t, sErr)
		require.NoError(t, story.AddSentence(s))

		step++
	}

	exporter := &infrastructure.PlantUMLExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// All 7 PlantUML type names must appear.
	assert.Contains(t, result, "Document(MyDocument)")
	assert.Contains(t, result, "Folder(MyFolder)")
	assert.Contains(t, result, "Call(MyCall)")
	assert.Contains(t, result, "Email(MyEmail)")
	assert.Contains(t, result, "Conversation(MyConversation)")
	assert.Contains(t, result, "Info(MyInfo)")
	// data maps to Document with a comment.
	assert.Contains(t, result, "Document(MyData")
}

func TestPlantUMLExporter_Render_MultiWordActorName(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Multi-Word Actor Name",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("PetOwner", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Pet", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := discoverydomain.NewStorySentence(1, "PetOwner", "registers", "Pet", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	exporter := &infrastructure.PlantUMLExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Multi-word: Person(PetOwner, "Pet Owner")
	assert.Contains(t, result, `Person(PetOwner, "Pet Owner")`)
}

func TestPlantUMLExporter_Render_IndirectObjectIsLiteral(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Indirect Object Literal",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Calendar", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// "available slots" is not an actor/work object — it should be rendered as a quoted literal.
	availSlots, err := discoverydomain.NewWorkObject("available slots", discoverydomain.WorkObjectTypeInfo, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(availSlots))

	sentence, err := discoverydomain.NewStorySentence(1, "Customer", "checks", "Calendar", vo.UserStated, "")
	require.NoError(t, err)
	sentence, err = sentence.WithPreposition("for", "available slots")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	exporter := &infrastructure.PlantUMLExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Literal indirect objects with spaces should be quoted in the activity line.
	assert.Contains(t, result, `"available slots"`)
}

func TestPlantUMLExporter_Render_IndirectObjectIsActor(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Indirect Object Actor",
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

	exporter := &infrastructure.PlantUMLExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Actor indirect objects should be unquoted.
	assert.Contains(t, result, "Seller")
	// The activity line should reference Seller without quotes.
	assert.Contains(t, result, "activity(1,")
	assert.Contains(t, result, ", to, Seller)")
}

func TestPlantUMLExporter_Render_HeaderAndFooter(t *testing.T) {
	t.Parallel()

	story := buildPlantUMLEcommerceStory(t)
	exporter := &infrastructure.PlantUMLExporter{}

	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Must start with @startuml and end with @enduml.
	assert.Contains(t, result, "@startuml")
	assert.Contains(t, result, "@enduml")

	// Must include DomainStory-PlantUML library.
	assert.Contains(t, result, "!include <DomainStory/domainStory>")

	// Must have section comments.
	assert.Contains(t, result, "' --- Actors ---")
	assert.Contains(t, result, "' --- Work Objects ---")
	assert.Contains(t, result, "' --- Sentences ---")
}

func TestPlantUMLExporter_Render_StoryTitleInComment(t *testing.T) {
	t.Parallel()

	story := buildPlantUMLEcommerceStory(t)
	exporter := &infrastructure.PlantUMLExporter{}

	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// Title should appear as a PlantUML comment.
	assert.Contains(t, result, "' E-commerce: Customer Purchases Product")
}

func TestPlantUMLExporter_Render_DataTypeMappedToDocument(t *testing.T) {
	t.Parallel()

	story, err := discoverydomain.NewDomainStory(
		"Data Type Mapping",
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("Analyst", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	// "data" type should map to "Document" in PlantUML with a comment.
	dataWO, err := discoverydomain.NewWorkObject("Metrics", discoverydomain.WorkObjectTypeData, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(dataWO))

	sentence, err := discoverydomain.NewStorySentence(1, "Analyst", "reviews", "Metrics", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	exporter := &infrastructure.PlantUMLExporter{}
	result, err := exporter.Render(context.TODO(), story)
	require.NoError(t, err)

	// data → Document with a comment noting the original type.
	assert.Contains(t, result, "Document(Metrics")
	assert.Contains(t, result, "' mapped from data")
}
