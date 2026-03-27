package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- StorySentence.WithTrust tests --

func TestStorySentence_WithTrust_WhenUpgradingFromAIResearched_ExpectSourceCleared(t *testing.T) {
	t.Parallel()

	// Given: sentence with AIResearched trust and a source
	s, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When: upgrading to UserConfirmed
	result := s.WithTrust(vo.UserConfirmed)

	// Then: trust upgraded, source cleared, all other fields preserved
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Empty(t, result.Source())
	assert.Equal(t, 1, result.Step())
	assert.Equal(t, "Customer", result.Subject())
	assert.Equal(t, "submits", result.Activity())
	assert.Equal(t, "Order", result.Object())
	assert.Empty(t, result.Preposition())
	assert.Empty(t, result.IndirectObject())
}

func TestStorySentence_WithTrust_WhenDowngrading_ExpectNoChange(t *testing.T) {
	t.Parallel()

	// Given: sentence with UserConfirmed trust
	s, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.UserConfirmed, "")
	require.NoError(t, err)

	// When: attempting to downgrade to AIResearched
	result := s.WithTrust(vo.AIResearched)

	// Then: returned sentence unchanged
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Empty(t, result.Source())
}

// -- StoryActor.WithTrust tests --

func TestStoryActor_WithTrust_WhenUpgradingFromAIResearched_ExpectSourceCleared(t *testing.T) {
	t.Parallel()

	// Given: actor with AIResearched trust and a source
	a, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When: upgrading to UserConfirmed
	result := a.WithTrust(vo.UserConfirmed)

	// Then: trust upgraded, source cleared, name and actorType preserved
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Empty(t, result.Source())
	assert.Equal(t, "Customer", result.Name())
	assert.Equal(t, domain.ActorTypePerson, result.Type())
}

// -- WorkObject.WithTrust tests --

func TestWorkObject_WithTrust_WhenUpgradingFromAIResearched_ExpectSourceCleared(t *testing.T) {
	t.Parallel()

	// Given: work object with AIResearched trust and a source
	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When: upgrading to UserConfirmed
	result := wo.WithTrust(vo.UserConfirmed)

	// Then: trust upgraded, source cleared, name and objectType preserved
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Empty(t, result.Source())
	assert.Equal(t, "Order", result.Name())
	assert.Equal(t, domain.WorkObjectTypeDocument, result.Type())
}

// -- DomainStory.UpgradeActorTrust tests --

func TestDomainStory_UpgradeActorTrust_WhenActorExists_ExpectTrustUpgraded(t *testing.T) {
	t.Parallel()

	// Given: story with actor "Customer" at AIResearched trust
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	// When: upgrading with case-insensitive name match
	story.UpgradeActorTrust("customer", vo.UserConfirmed)

	// Then: actor trust upgraded
	actors := story.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserConfirmed, actors[0].Trust())
}

func TestDomainStory_UpgradeActorTrust_WhenNotFound_ExpectNoOp(t *testing.T) {
	t.Parallel()

	// Given: story with actor "Customer"
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	// When: upgrading an unknown actor
	story.UpgradeActorTrust("Unknown", vo.UserConfirmed)

	// Then: no change, no error (no-op)
	actors := story.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.AIResearched, actors[0].Trust())
}

// -- DomainStory.UpgradeWorkObjectTrust tests --

func TestDomainStory_UpgradeWorkObjectTrust_WhenObjectExists_ExpectTrustUpgraded(t *testing.T) {
	t.Parallel()

	// Given: story with work object "Order" at AIResearched trust
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// When: upgrading with case-insensitive name match
	story.UpgradeWorkObjectTrust("order", vo.UserConfirmed)

	// Then: work object trust upgraded
	objects := story.WorkObjects()
	require.Len(t, objects, 1)
	assert.Equal(t, vo.UserConfirmed, objects[0].Trust())
}

// -- PropagateConfirmation tests --

func TestPropagateConfirmation_WhenAcceptedUnedited_ExpectUserConfirmed(t *testing.T) {
	t.Parallel()

	// Given: story with actor and work object at AIResearched trust
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// confirmed = same as proposed (unedited)
	confirmed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then: sentence trust is UserConfirmed
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Empty(t, result.Source())

	// And: actor and work object trust also upgraded
	actors := story.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserConfirmed, actors[0].Trust())

	objects := story.WorkObjects()
	require.Len(t, objects, 1)
	assert.Equal(t, vo.UserConfirmed, objects[0].Trust())
}

func TestPropagateConfirmation_WhenAcceptedEdited_SubjectChanged_ExpectUserStated(t *testing.T) {
	t.Parallel()

	// Given
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Client", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// confirmed has subject changed: Customer -> Client
	confirmed, err := domain.NewStorySentence(1, "Client", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then: trust is UserStated (user edited)
	assert.Equal(t, vo.UserStated, result.Trust())
}

func TestPropagateConfirmation_WhenAcceptedEdited_ActivityChanged_ExpectUserStated(t *testing.T) {
	t.Parallel()

	// Given
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// confirmed has activity changed: submits -> sends
	confirmed, err := domain.NewStorySentence(1, "Customer", "sends", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then
	assert.Equal(t, vo.UserStated, result.Trust())
}

func TestPropagateConfirmation_WhenAcceptedEdited_ObjectChanged_ExpectUserStated(t *testing.T) {
	t.Parallel()

	// Given
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Request", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// confirmed has object changed: Order -> Request
	confirmed, err := domain.NewStorySentence(1, "Customer", "submits", "Request", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then
	assert.Equal(t, vo.UserStated, result.Trust())
}

func TestPropagateConfirmation_WhenRejected_ExpectProposedReturned(t *testing.T) {
	t.Parallel()

	// Given
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	confirmed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When: rejected
	result := domain.PropagateConfirmation(proposed, confirmed, false, story)

	// Then: returned sentence is the proposed one, unchanged
	assert.Equal(t, vo.AIResearched, result.Trust())
	assert.Equal(t, "https://example.com", result.Source())

	// And: story actor/work object unchanged
	actors := story.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.AIResearched, actors[0].Trust())

	objects := story.WorkObjects()
	require.Len(t, objects, 1)
	assert.Equal(t, vo.AIResearched, objects[0].Trust())
}

func TestPropagateConfirmation_WhenAlreadyHigherTrust_ExpectNoUpgrade(t *testing.T) {
	t.Parallel()

	// Given: proposed sentence with UserStated trust (already highest)
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	// confirmed = same (unedited), accepted = true
	confirmed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.UserStated, "")
	require.NoError(t, err)

	// When: PropagateConfirmation — would try UserConfirmed, but UserStated is higher
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then: sentence stays UserStated (UserConfirmed is lower trust than UserStated)
	assert.Equal(t, vo.UserStated, result.Trust())

	// And: actor and work object also stay UserStated
	actors := story.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserStated, actors[0].Trust())

	objects := story.WorkObjects()
	require.Len(t, objects, 1)
	assert.Equal(t, vo.UserStated, objects[0].Trust())
}

// -- QA Edge Case Tests --

// TestTrust_QA_WithTrust_PreservesPrepositionAndIndirectObject is the KEY test
// for the grooming correction: WithTrust uses direct field copy, not WithPreposition.
// Preposition and indirect object MUST survive the trust upgrade.
func TestTrust_QA_WithTrust_PreservesPrepositionAndIndirectObject(t *testing.T) {
	t.Parallel()

	// Given: sentence with preposition and indirect object set
	s, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	s, err = s.WithPreposition("for", "CRM System")
	require.NoError(t, err)

	// Precondition: preposition and indirect object are set
	require.Equal(t, "for", s.Preposition())
	require.Equal(t, "CRM System", s.IndirectObject())

	// When: upgrading trust to UserConfirmed
	result := s.WithTrust(vo.UserConfirmed)

	// Then: trust is upgraded
	assert.Equal(t, vo.UserConfirmed, result.Trust())

	// And: preposition and indirect object MUST survive
	assert.Equal(t, "for", result.Preposition())
	assert.Equal(t, "CRM System", result.IndirectObject())

	// And: core fields preserved
	assert.Equal(t, 1, result.Step())
	assert.Equal(t, "Customer", result.Subject())
	assert.Equal(t, "submits", result.Activity())
	assert.Equal(t, "Order", result.Object())
	assert.Empty(t, result.Source())
}

func TestTrust_QA_StoryActor_WithTrust_NoDowngrade(t *testing.T) {
	t.Parallel()

	// Given: actor with UserConfirmed trust (high trust)
	a, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserConfirmed, "")
	require.NoError(t, err)

	// When: attempting downgrade to AIResearched (lower trust)
	result := a.WithTrust(vo.AIResearched)

	// Then: actor unchanged — trust still UserConfirmed
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Equal(t, "Customer", result.Name())
	assert.Equal(t, domain.ActorTypePerson, result.Type())
}

func TestTrust_QA_WorkObject_WithTrust_NoDowngrade(t *testing.T) {
	t.Parallel()

	// Given: work object with UserConfirmed trust (high trust)
	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserConfirmed, "")
	require.NoError(t, err)

	// When: attempting downgrade to AIResearched (lower trust)
	result := wo.WithTrust(vo.AIResearched)

	// Then: work object unchanged — trust still UserConfirmed
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Equal(t, "Order", result.Name())
	assert.Equal(t, domain.WorkObjectTypeDocument, result.Type())
}

func TestTrust_QA_StorySentence_WithTrust_SameTrustIsNoOp(t *testing.T) {
	t.Parallel()

	// Given: sentence at UserConfirmed trust
	s, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.UserConfirmed, "")
	require.NoError(t, err)

	// When: WithTrust called with the SAME trust level
	result := s.WithTrust(vo.UserConfirmed)

	// Then: IsHigherTrust(same) returns false, so sentence is unchanged
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Equal(t, "Customer", result.Subject())
	assert.Equal(t, "submits", result.Activity())
	assert.Equal(t, "Order", result.Object())
}

func TestTrust_QA_WithTrust_SourceClearedOnUpgrade(t *testing.T) {
	t.Parallel()

	// Given: actor at AIResearched trust with a source URL
	a, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.Equal(t, "https://example.com", a.Source())

	// When: upgrading to UserConfirmed
	result := a.WithTrust(vo.UserConfirmed)

	// Then: source is cleared (empty string)
	assert.Equal(t, vo.UserConfirmed, result.Trust())
	assert.Empty(t, result.Source())
}

func TestTrust_QA_UpgradeActorTrust_CaseInsensitiveMatch(t *testing.T) {
	t.Parallel()

	// Given: story with actor "Customer" (mixed case)
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	// When: upgrading with ALL CAPS name
	story.UpgradeActorTrust("CUSTOMER", vo.UserConfirmed)

	// Then: actor trust upgraded despite case mismatch
	actors := story.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserConfirmed, actors[0].Trust())
}

func TestTrust_QA_UpgradeWorkObjectTrust_CaseInsensitiveMatch(t *testing.T) {
	t.Parallel()

	// Given: story with work object "Order" (mixed case)
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// When: upgrading with ALL CAPS name
	story.UpgradeWorkObjectTrust("ORDER", vo.UserConfirmed)

	// Then: work object trust upgraded despite case mismatch
	objects := story.WorkObjects()
	require.Len(t, objects, 1)
	assert.Equal(t, vo.UserConfirmed, objects[0].Trust())
}

func TestTrust_QA_UpgradeWorkObjectTrust_NotFoundIsNoOp(t *testing.T) {
	t.Parallel()

	// Given: story with work object "Order"
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// When: upgrading with an unknown name
	story.UpgradeWorkObjectTrust("Invoice", vo.UserConfirmed)

	// Then: no change, no panic — original work object untouched
	objects := story.WorkObjects()
	require.Len(t, objects, 1)
	assert.Equal(t, vo.AIResearched, objects[0].Trust())
	assert.Equal(t, "Order", objects[0].Name())
}

func TestTrust_QA_PropagateConfirmation_OnlyPrepositionChanged_CountsAsUnedited(t *testing.T) {
	t.Parallel()

	// Given: story with actor and work object
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// proposed with preposition "for" / "System"
	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	proposed, err = proposed.WithPreposition("for", "System")
	require.NoError(t, err)

	// confirmed with DIFFERENT preposition "via" / "API" — but same Subject/Activity/Object
	confirmed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	confirmed, err = confirmed.WithPreposition("via", "API")
	require.NoError(t, err)

	// When: PropagateConfirmation with accepted=true
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then: trust = UserConfirmed (NOT UserStated) because only S/A/O are compared
	assert.Equal(t, vo.UserConfirmed, result.Trust())
}

func TestTrust_QA_PropagateConfirmation_CaseSensitiveEditDetection(t *testing.T) {
	t.Parallel()

	// Given: story with actor and work object
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	// Actor matches the confirmed subject (lowercase "customer")
	actor, err := domain.NewStoryActor("customer", domain.ActorTypePerson, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "https://example.com")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// proposed has "Customer" (capital C)
	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// confirmed has "customer" (lowercase c) — case difference counts as edit
	confirmed, err := domain.NewStorySentence(1, "customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then: trust = UserStated because case-sensitive comparison detects the edit
	assert.Equal(t, vo.UserStated, result.Trust())
}

func TestTrust_QA_PropagateConfirmation_HigherTrustPreservedOnAllThree(t *testing.T) {
	t.Parallel()

	// Given: story with actor and work object already at UserStated trust (highest)
	story, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	// proposed and confirmed are identical (unedited) — at AIResearched trust
	proposed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	confirmed, err := domain.NewStorySentence(1, "Customer", "submits", "Order", vo.AIResearched, "https://example.com")
	require.NoError(t, err)

	// When: PropagateConfirmation — unedited means UserConfirmed, but UserStated > UserConfirmed
	result := domain.PropagateConfirmation(proposed, confirmed, true, story)

	// Then: sentence gets UserConfirmed (upgrade from AIResearched is valid)
	assert.Equal(t, vo.UserConfirmed, result.Trust())

	// But: actor stays at UserStated (UserStated > UserConfirmed, no-downgrade guard)
	actors := story.Actors()
	require.Len(t, actors, 1)
	assert.Equal(t, vo.UserStated, actors[0].Trust())

	// And: work object stays at UserStated (same no-downgrade guard)
	objects := story.WorkObjects()
	require.Len(t, objects, 1)
	assert.Equal(t, vo.UserStated, objects[0].Trust())
}
