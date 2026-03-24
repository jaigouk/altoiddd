package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- Helpers ---

func mustActor(t *testing.T, name string) domain.StoryActor {
	t.Helper()

	a, err := domain.NewStoryActor(name, domain.ActorTypePerson, vo.UserConfirmed, "")
	require.NoError(t, err)

	return a
}

func mustActorWithType(t *testing.T, name string, actorType domain.ActorType) domain.StoryActor {
	t.Helper()

	a, err := domain.NewStoryActor(name, actorType, vo.UserConfirmed, "")
	require.NoError(t, err)

	return a
}

func mustWorkObject(t *testing.T, name string) domain.WorkObject {
	t.Helper()

	wo, err := domain.NewWorkObject(name, domain.WorkObjectTypeDocument, vo.UserConfirmed, "")
	require.NoError(t, err)

	return wo
}

func mustSentence(t *testing.T, step int, subject, activity, object string) domain.StorySentence {
	t.Helper()

	s, err := domain.NewStorySentence(step, subject, activity, object, vo.UserConfirmed, "")
	require.NoError(t, err)

	return s
}

func mustSentenceWithIndirect(t *testing.T, step int, subject, activity, object, prep, indirect string) domain.StorySentence {
	t.Helper()

	s := mustSentence(t, step, subject, activity, object)

	s2, err := s.WithPreposition(prep, indirect)
	require.NoError(t, err)

	return s2
}

func mustAnnotation(t *testing.T, text string, aType domain.AnnotationType, ref *int) domain.Annotation {
	t.Helper()

	a, err := domain.NewAnnotation(text, aType, ref, vo.UserConfirmed, "")
	require.NoError(t, err)

	return a
}

func intPtr(v int) *int {
	return &v
}

func mustValidStory(t *testing.T) *domain.DomainStory {
	t.Helper()

	ds, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "creates", "Order")))

	return ds
}

// --- Construction ---

func TestNewDomainStory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		title      string
		storyType  domain.StoryType
		timeType   domain.TimeType
		purityType domain.PurityType
		trigger    string
		wantErr    bool
	}{
		{
			name:       "valid",
			title:      "Order Processing",
			storyType:  domain.StoryTypeCoarseGrained,
			timeType:   domain.TimeTypeAsIs,
			purityType: domain.PurityTypePure,
			trigger:    "Customer places order",
		},
		{
			name:       "empty title",
			title:      "",
			storyType:  domain.StoryTypeCoarseGrained,
			timeType:   domain.TimeTypeAsIs,
			purityType: domain.PurityTypePure,
			trigger:    "trigger",
			wantErr:    true,
		},
		{
			name:       "whitespace title",
			title:      "  \t  ",
			storyType:  domain.StoryTypeCoarseGrained,
			timeType:   domain.TimeTypeAsIs,
			purityType: domain.PurityTypePure,
			trigger:    "trigger",
			wantErr:    true,
		},
		{
			name:       "empty trigger",
			title:      "Title",
			storyType:  domain.StoryTypeCoarseGrained,
			timeType:   domain.TimeTypeAsIs,
			purityType: domain.PurityTypePure,
			trigger:    "",
			wantErr:    true,
		},
		{
			name:       "invalid story type",
			title:      "Title",
			storyType:  domain.StoryType("invalid"),
			timeType:   domain.TimeTypeAsIs,
			purityType: domain.PurityTypePure,
			trigger:    "trigger",
			wantErr:    true,
		},
		{
			name:       "invalid time type",
			title:      "Title",
			storyType:  domain.StoryTypeCoarseGrained,
			timeType:   domain.TimeType("invalid"),
			purityType: domain.PurityTypePure,
			trigger:    "trigger",
			wantErr:    true,
		},
		{
			name:       "invalid purity type",
			title:      "Title",
			storyType:  domain.StoryTypeCoarseGrained,
			timeType:   domain.TimeTypeAsIs,
			purityType: domain.PurityType("invalid"),
			trigger:    "trigger",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ds, err := domain.NewDomainStory(tt.title, tt.storyType, tt.timeType, tt.purityType, tt.trigger)
			if tt.wantErr {
				require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.title, ds.Title())
			assert.Equal(t, tt.storyType, ds.Type())
			assert.Equal(t, tt.timeType, ds.Time())
			assert.Equal(t, tt.purityType, ds.Purity())
			assert.Equal(t, tt.trigger, ds.Trigger())
		})
	}
}

// --- AddActor ---

func TestDomainStory_AddActor_Success(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddActor(mustActor(t, "Admin")))

	assert.Len(t, ds.Actors(), 2)
}

func TestDomainStory_AddActor_Duplicate(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))

	// Case-insensitive duplicate.
	err = ds.AddActor(mustActor(t, "CUSTOMER"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Len(t, ds.Actors(), 1)
}

// --- AddWorkObject ---

func TestDomainStory_AddWorkObject_Success(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Invoice")))

	assert.Len(t, ds.WorkObjects(), 2)
}

func TestDomainStory_AddWorkObject_Duplicate(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))

	// Case-insensitive duplicate.
	err = ds.AddWorkObject(mustWorkObject(t, "ORDER"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Len(t, ds.WorkObjects(), 1)
}

// --- AddSentence ---

func TestDomainStory_AddSentence_Success(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Actor", "does", "Thing")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 2, "Actor", "does", "Other")))

	assert.Equal(t, 2, ds.SentenceCount())
}

// --- AddAnnotation ---

func TestDomainStory_AddAnnotation_StoryWide(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	ann := mustAnnotation(t, "All orders require approval", domain.AnnotationTypeInvariant, nil)
	require.NoError(t, ds.AddAnnotation(ann))

	annotations := ds.Annotations()
	require.Len(t, annotations, 1)
	assert.True(t, annotations[0].IsStoryWide())
}

func TestDomainStory_AddAnnotation_SentenceSpecific(t *testing.T) {
	t.Parallel()

	// Annotation with sentence ref added BEFORE sentence — no validation at add time.
	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	ann := mustAnnotation(t, "Must have payment", domain.AnnotationTypeConstraint, intPtr(5))
	require.NoError(t, ds.AddAnnotation(ann))

	annotations := ds.Annotations()
	require.Len(t, annotations, 1)

	ref := annotations[0].SentenceRef()
	require.NotNil(t, ref)
	assert.Equal(t, 5, *ref)
}

// --- AddVariation ---

func TestDomainStory_AddVariation_Success(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddVariation("Customer cancels order"))

	assert.Len(t, ds.Variations(), 1)
}

func TestDomainStory_AddVariation_Empty(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	err = ds.AddVariation("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)

	err = ds.AddVariation("   ")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)

	assert.Empty(t, ds.Variations())
}

// --- Validate ---

func TestDomainStory_Validate_Valid(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	require.NoError(t, ds.Validate())
}

func TestDomainStory_Validate_NoActors(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Someone", "does", "Order")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_NoWorkObjects(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "does", "Something")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_NoSentences(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_NonSequentialSteps(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "creates", "Order")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 3, "Customer", "submits", "Order")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_DuplicateSteps(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "creates", "Order")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "submits", "Order")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_SubjectNotInActors(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	// Subject "Unknown" is not a registered actor.
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Unknown", "creates", "Order")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_ObjectNotInActorsOrWO(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	// Object "Unknown" is neither an actor nor a work object.
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "creates", "Unknown")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_IndirectObjectNotInActorsOrWO(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	// Indirect object "Unknown" is neither an actor nor a work object.
	require.NoError(t, ds.AddSentence(mustSentenceWithIndirect(t, 1, "Customer", "sends", "Order", "to", "Unknown")))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_AnnotationBadSentenceRef(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t) // 1 actor, 1 WO, 1 sentence (step 1)

	require.NoError(t, ds.AddAnnotation(mustAnnotation(t, "Bad ref", domain.AnnotationTypeConstraint, intPtr(99))))

	require.ErrorIs(t, ds.Validate(), domainerrors.ErrInvariantViolation)
}

func TestDomainStory_Validate_ActorAsObject(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActor(t, "Customer")))
	require.NoError(t, ds.AddActor(mustActor(t, "Admin")))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	// Object is "Admin" (an actor, not a work object) — should be valid.
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "notifies", "Admin")))

	require.NoError(t, ds.Validate())
}

// --- Query methods ---

func TestDomainStory_SentenceCount(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	assert.Equal(t, 1, ds.SentenceCount())

	require.NoError(t, ds.AddSentence(mustSentence(t, 2, "Customer", "submits", "Order")))

	assert.Equal(t, 2, ds.SentenceCount())
}

func TestDomainStory_HasBranching_True(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory("Story", domain.StoryTypeCoarseGrained, domain.TimeTypeAsIs, domain.PurityTypePure, "trigger")
	require.NoError(t, err)

	// "sometimes validates" contains branching keyword.
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Actor", "sometimes validates", "Thing")))

	assert.True(t, ds.HasBranching())
}

func TestDomainStory_HasBranching_False(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	assert.False(t, ds.HasBranching())
}

func TestDomainStory_FormatText(t *testing.T) {
	t.Parallel()

	ds, err := domain.NewDomainStory(
		"Order Processing",
		domain.StoryTypeCoarseGrained,
		domain.TimeTypeAsIs,
		domain.PurityTypePure,
		"Customer places order",
	)
	require.NoError(t, err)

	require.NoError(t, ds.AddActor(mustActorWithType(t, "Customer", domain.ActorTypePerson)))
	require.NoError(t, ds.AddActor(mustActorWithType(t, "OrderSystem", domain.ActorTypeSystem)))
	require.NoError(t, ds.AddWorkObject(mustWorkObject(t, "Order")))
	require.NoError(t, ds.AddSentence(mustSentence(t, 1, "Customer", "creates", "Order")))
	require.NoError(t, ds.AddSentence(mustSentenceWithIndirect(t, 2, "OrderSystem", "validates", "Order", "for", "Customer")))
	require.NoError(t, ds.AddAnnotation(mustAnnotation(t, "Must have valid payment", domain.AnnotationTypeConstraint, intPtr(1))))
	require.NoError(t, ds.AddAnnotation(mustAnnotation(t, "All orders require approval", domain.AnnotationTypeInvariant, nil)))

	expected := `# Order Processing (coarse_grained, as_is, pure)
Trigger: Customer places order

## Actors
- Customer (person)
- OrderSystem (system)

## Work Objects
- Order (document)

## Story
1. Customer creates Order
2. OrderSystem validates Order for Customer

## Annotations
- [constraint] (sentence 1) Must have valid payment
- [invariant] (story-wide) All orders require approval`

	assert.Equal(t, expected, ds.FormatText())
}

// --- Defensive copies ---

func TestDomainStory_Actors_DefensiveCopy(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	actors := ds.Actors()
	original := ds.Actors()

	require.Len(t, actors, 1)

	// Mutate the returned slice.
	actors[0] = mustActor(t, "Intruder")

	// Original aggregate is unaffected.
	assert.Equal(t, original[0].Name(), ds.Actors()[0].Name())
}

func TestDomainStory_WorkObjects_DefensiveCopy(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	wos := ds.WorkObjects()
	original := ds.WorkObjects()

	require.Len(t, wos, 1)

	// Mutate the returned slice.
	wos[0] = mustWorkObject(t, "Intruder")

	// Original aggregate is unaffected.
	assert.Equal(t, original[0].Name(), ds.WorkObjects()[0].Name())
}

func TestDomainStory_Sentences_DefensiveCopy(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	sentences := ds.Sentences()
	original := ds.Sentences()

	require.Len(t, sentences, 1)

	// Mutate the returned slice.
	sentences[0] = mustSentence(t, 99, "Intruder", "attacks", "System")

	// Original aggregate is unaffected.
	assert.Equal(t, original[0].Step(), ds.Sentences()[0].Step())
}

func TestDomainStory_Annotations_DefensiveCopy(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	require.NoError(t, ds.AddAnnotation(mustAnnotation(t, "Original note", domain.AnnotationTypeConstraint, intPtr(1))))

	annotations := ds.Annotations()
	original := ds.Annotations()

	require.Len(t, annotations, 1)

	// Mutate the returned slice.
	annotations[0] = mustAnnotation(t, "Intruder", domain.AnnotationTypeAssumption, nil)

	// Original aggregate is unaffected.
	assert.Equal(t, original[0].Text(), ds.Annotations()[0].Text())
}

func TestDomainStory_Variations_DefensiveCopy(t *testing.T) {
	t.Parallel()

	ds := mustValidStory(t)

	require.NoError(t, ds.AddVariation("Original variation"))

	variations := ds.Variations()
	original := ds.Variations()

	require.Len(t, variations, 1)

	// Mutate the returned slice.
	variations[0] = "Intruder"

	// Original aggregate is unaffected.
	assert.Equal(t, original[0], ds.Variations()[0])
}
