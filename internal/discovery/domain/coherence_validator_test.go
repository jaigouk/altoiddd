package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// buildCoherenceStory creates a DomainStory with the given actors, work objects,
// and sentences for coherence validation tests.
func buildCoherenceStory(
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

func makeActor(t *testing.T, name string, actorType domain.ActorType) domain.StoryActor {
	t.Helper()

	actor, err := domain.NewStoryActor(name, actorType, vo.UserStated, "")
	require.NoError(t, err)

	return actor
}

func makeWorkObject(t *testing.T, name string, woType domain.WorkObjectType) domain.WorkObject {
	t.Helper()

	wo, err := domain.NewWorkObject(name, woType, vo.UserStated, "")
	require.NoError(t, err)

	return wo
}

func makeSentence(t *testing.T, step int, subject, activity, object string) domain.StorySentence {
	t.Helper()

	sentence, err := domain.NewStorySentence(step, subject, activity, object, vo.UserStated, "")
	require.NoError(t, err)

	return sentence
}

func makeSentenceWithIndirect(
	t *testing.T,
	step int,
	subject, activity, object, preposition, indirectObject string,
) domain.StorySentence {
	t.Helper()

	sentence, err := domain.NewStorySentence(step, subject, activity, object, vo.UserStated, "")
	require.NoError(t, err)

	sentence, err = sentence.WithPreposition(preposition, indirectObject)
	require.NoError(t, err)

	return sentence
}

func TestCoherenceValidator_Validate_EmptyStories(t *testing.T) {
	t.Parallel()

	validator := domain.CoherenceValidator{}
	report := validator.Validate(nil)

	assert.True(t, report.IsCoherent())
	assert.False(t, report.HasFindings())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_SingleStory_NoConflicts(t *testing.T) {
	t.Parallel()

	story := buildCoherenceStory(t, "Order Flow",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "submits", "Order"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{story})

	assert.True(t, report.IsCoherent())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_ActorTypeConflict_AcrossStories(t *testing.T) {
	t.Parallel()

	// "Customer" is a person in story A but a system in story B.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "submits", "Order"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypeSystem),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "generates", "Invoice"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.False(t, report.IsCoherent())
	assert.True(t, report.HasFindings())
	require.Len(t, report.Findings(), 1)

	finding := report.Findings()[0]
	assert.Equal(t, domain.CoherenceSeverityWarning, finding.Severity())
	assert.Contains(t, finding.Description(), "Customer")
	assert.Contains(t, finding.Location(), "Story A")
	assert.Contains(t, finding.Location(), "Story B")
}

func TestCoherenceValidator_Validate_WorkObjectTypeConflict_AcrossStories(t *testing.T) {
	t.Parallel()

	// "Report" is a document in story A but a folder in story B.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Manager", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Report", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Manager", "reviews", "Report"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "Analyst", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Report", domain.WorkObjectTypeFolder),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Analyst", "archives", "Report"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.False(t, report.IsCoherent())
	require.Len(t, report.Findings(), 1)

	finding := report.Findings()[0]
	assert.Equal(t, domain.CoherenceSeverityWarning, finding.Severity())
	assert.Contains(t, finding.Description(), "Report")
}

func TestCoherenceValidator_Validate_CrossDimensionConflict_ActorVsWorkObject(t *testing.T) {
	t.Parallel()

	// "Gateway" is an actor in story A but a work object in story B.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Gateway", domain.ActorTypeSystem),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Request", domain.WorkObjectTypeData),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Gateway", "routes", "Request"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "Admin", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Gateway", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Admin", "configures", "Gateway"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.False(t, report.IsCoherent())
	require.Len(t, report.Findings(), 1)

	finding := report.Findings()[0]
	assert.Equal(t, domain.CoherenceSeverityWarning, finding.Severity())
	assert.Contains(t, finding.Description(), "Gateway")
	assert.Contains(t, finding.Description(), "actor")
	assert.Contains(t, finding.Description(), "work object")
}

func TestCoherenceValidator_Validate_CaseInsensitiveNormalization(t *testing.T) {
	t.Parallel()

	// "customer" (lower) vs "CUSTOMER" (upper) with different types -> conflict.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "customer", "submits", "Order"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "CUSTOMER", domain.ActorTypeSystem),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "CUSTOMER", "generates", "Invoice"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.False(t, report.IsCoherent())
	require.Len(t, report.Findings(), 1)
}

func TestCoherenceValidator_Validate_SameTermSameType_NoConflict(t *testing.T) {
	t.Parallel()

	// "Customer" is a person in both stories -> no conflict.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "submits", "Order"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "pays", "Invoice"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.True(t, report.IsCoherent())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_UndeclaredCrossStoryRef_Subject(t *testing.T) {
	t.Parallel()

	// Story A has a sentence with subject "Admin" but "Admin" is NOT declared
	// as an actor in ANY story.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "submits", "Order"),
		},
	)

	// Manually add a sentence referencing an undeclared actor.
	// AddSentence doesn't validate refs (that's deferred to Validate).
	undeclaredSentence := makeSentence(t, 2, "Admin", "approves", "Order")
	require.NoError(t, storyA.AddSentence(undeclaredSentence))

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA})

	assert.False(t, report.IsCoherent())
	require.Len(t, report.Findings(), 1)

	finding := report.Findings()[0]
	assert.Contains(t, finding.Description(), "Admin")
	assert.Contains(t, finding.Location(), "step: 2")
}

func TestCoherenceValidator_Validate_CrossStoryValidRef_NotFlagged(t *testing.T) {
	t.Parallel()

	// Story A defines "Warehouse" as an actor.
	// Story B's sentence references "Warehouse" as subject — valid cross-story ref.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Warehouse", domain.ActorTypeSystem),
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "submits", "Order"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "Clerk", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Shipment", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Clerk", "prepares", "Shipment"),
		},
	)

	// Add sentence in story B referencing "Warehouse" (defined in story A).
	crossRef := makeSentence(t, 2, "Warehouse", "ships", "Shipment")
	require.NoError(t, storyB.AddSentence(crossRef))

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.True(t, report.IsCoherent())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_TrulyUndeclaredRef_NotInAnyStory(t *testing.T) {
	t.Parallel()

	// "Ghost" is not declared as an actor or work object in any story.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "submits", "Order"),
		},
	)

	// Add sentence where object "Ghost" is not in any global set.
	ghostSentence := makeSentence(t, 2, "Customer", "sends", "Ghost")
	require.NoError(t, storyA.AddSentence(ghostSentence))

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA})

	assert.False(t, report.IsCoherent())

	// Should have one finding for the undeclared object "Ghost".
	findings := report.Findings()
	found := false

	for _, f := range findings {
		if f.Severity() == domain.CoherenceSeverityWarning && strings.Contains(f.Description(), "Ghost") {
			found = true
		}
	}

	assert.True(t, found, "expected finding for undeclared ref 'Ghost'")
}

func TestCoherenceValidator_Validate_IndirectObject_TrulyUndeclared(t *testing.T) {
	t.Parallel()

	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{},
	)

	// Sentence: "Customer submits Order to Phantom" where Phantom is undeclared.
	sentence := makeSentenceWithIndirect(t, 1, "Customer", "submits", "Order", "to", "Phantom")
	require.NoError(t, storyA.AddSentence(sentence))

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA})

	assert.False(t, report.IsCoherent())

	findings := report.Findings()
	found := false

	for _, f := range findings {
		if strings.Contains(f.Description(), "Phantom") {
			found = true
			assert.Contains(t, f.Location(), "step: 1")
		}
	}

	assert.True(t, found, "expected finding for undeclared indirect object 'Phantom'")
}

func TestCoherenceValidator_Validate_MultipleConflicts_AllReturned(t *testing.T) {
	t.Parallel()

	// Two type conflicts: "Customer" (person vs system) and "Report" (document vs folder).
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Report", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "reads", "Report"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypeSystem),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Report", domain.WorkObjectTypeFolder),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "archives", "Report"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.False(t, report.IsCoherent())
	assert.GreaterOrEqual(t, len(report.Findings()), 2)
}

func TestCoherenceFinding_Constructor_EmptyDescription_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := domain.NewCoherenceFinding(domain.CoherenceSeverityWarning, "some location", "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestCoherenceReport_Findings_DefensiveCopy(t *testing.T) {
	t.Parallel()

	finding, err := domain.NewCoherenceFinding(domain.CoherenceSeverityWarning, "loc", "desc")
	require.NoError(t, err)

	report := domain.NewCoherenceReport([]domain.CoherenceFinding{finding})

	// Mutate the returned slice — the original report should be unaffected.
	findings := report.Findings()
	require.Len(t, findings, 1)

	findings[0] = domain.CoherenceFinding{} // zero value

	// Re-read from report — should still have original finding.
	assert.Len(t, report.Findings(), 1)
	assert.Equal(t, "desc", report.Findings()[0].Description())
}

func TestCoherenceValidator_Validate_EmptySlice_NotNil(t *testing.T) {
	t.Parallel()

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{})

	assert.True(t, report.IsCoherent())
	assert.False(t, report.HasFindings())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_SameTermSameType_ThreeStories_NoFalsePositive(t *testing.T) {
	t.Parallel()

	// "Customer" is a person in all three stories -> no conflict.
	stories := make([]*domain.DomainStory, 0, 3)
	for i, title := range []string{"Story A", "Story B", "Story C"} {
		stories = append(stories, buildCoherenceStory(t, title,
			[]domain.StoryActor{
				makeActor(t, "Customer", domain.ActorTypePerson),
			},
			[]domain.WorkObject{
				makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
			},
			[]domain.StorySentence{
				makeSentence(t, i+1, "Customer", "submits", "Order"),
			},
		))
	}

	validator := domain.CoherenceValidator{}
	report := validator.Validate(stories)

	assert.True(t, report.IsCoherent())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_CaseVariation_SameType_NoFalsePositive(t *testing.T) {
	t.Parallel()

	// "CUSTOMER" and "customer" with the same type (person) -> no conflict.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "CUSTOMER", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "CUSTOMER", "submits", "Order"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Invoice", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "customer", "pays", "Invoice"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.True(t, report.IsCoherent())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_MultipleConflicts_ThreeTypeAndOneUndeclared(t *testing.T) {
	t.Parallel()

	// Three distinct term type conflicts + one undeclared reference = 4 findings.
	// Conflict 1: "Customer" person vs system
	// Conflict 2: "Report" document vs folder
	// Conflict 3: "Gateway" actor(system) vs work-object(data) (cross-dimension)
	// Undeclared: sentence references "Phantom" which is not in any story.
	storyA := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
			makeActor(t, "Gateway", domain.ActorTypeSystem),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Report", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "reads", "Report"),
			makeSentence(t, 2, "Gateway", "routes", "Report"),
		},
	)

	storyB := buildCoherenceStory(t, "Story B",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypeSystem),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Report", domain.WorkObjectTypeFolder),
			makeWorkObject(t, "Gateway", domain.WorkObjectTypeData),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "archives", "Report"),
		},
	)

	// Add a sentence referencing undeclared object "Phantom".
	undeclared := makeSentence(t, 2, "Customer", "sends", "Phantom")
	require.NoError(t, storyB.AddSentence(undeclared))

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{storyA, storyB})

	assert.False(t, report.IsCoherent())

	findings := report.Findings()
	// Expect exactly 4 findings: 3 term conflicts + 1 undeclared ref.
	assert.Len(t, findings, 4, "expected 3 type conflicts + 1 undeclared ref, got %d: %+v", len(findings), findings)
}

func TestCoherenceValidator_Validate_StoryWithNoSentences_NoUndeclaredRefFindings(t *testing.T) {
	t.Parallel()

	// A story with actors and work objects but NO sentences should not produce
	// any undeclared-ref findings (there are no sentences to walk).
	story := buildCoherenceStory(t, "Empty Sentences Story",
		[]domain.StoryActor{
			makeActor(t, "Admin", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Config", domain.WorkObjectTypeData),
		},
		[]domain.StorySentence{}, // explicitly empty
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{story})

	assert.True(t, report.IsCoherent())
	assert.Empty(t, report.Findings())
}

func TestCoherenceValidator_Validate_SentenceWithoutIndirectObject_SkipsIndirectCheck(t *testing.T) {
	t.Parallel()

	// Sentence has no indirect object -> the indirect object check path
	// should be completely skipped (no findings from it).
	story := buildCoherenceStory(t, "Story A",
		[]domain.StoryActor{
			makeActor(t, "Customer", domain.ActorTypePerson),
		},
		[]domain.WorkObject{
			makeWorkObject(t, "Order", domain.WorkObjectTypeDocument),
		},
		[]domain.StorySentence{
			makeSentence(t, 1, "Customer", "submits", "Order"),
		},
	)

	validator := domain.CoherenceValidator{}
	report := validator.Validate([]*domain.DomainStory{story})

	assert.True(t, report.IsCoherent())
	assert.Empty(t, report.Findings())

	// Verify HasIndirectObject is false for our sentence to confirm the skip.
	sentences := story.Sentences()
	require.Len(t, sentences, 1)
	assert.False(t, sentences[0].HasIndirectObject())
}
