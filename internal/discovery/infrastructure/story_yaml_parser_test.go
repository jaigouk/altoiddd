package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// sampleStoryDir returns the path to the sample story files.
func sampleStoryDir(t *testing.T) string {
	t.Helper()

	dir := filepath.Join("..", "..", "..", "docs", "research", "samples")

	info, err := os.Stat(dir)
	require.NoError(t, err, "sample directory must exist")
	require.True(t, info.IsDir(), "sample path must be a directory")

	return dir
}

func TestStoryYAMLParser_Parse_WhenEcommerceSample_ExpectValidStory(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join(sampleStoryDir(t), "ecommerce.story.yaml"))
	require.NoError(t, err)

	parser := &StoryYAMLParser{}

	story, err := parser.Parse(data)
	require.NoError(t, err)

	assert.Equal(t, "Customer Purchases Product from Marketplace", story.Title())
	assert.Equal(t, domain.StoryTypeCoarseGrained, story.Type())
	assert.Equal(t, domain.TimeTypeToBe, story.Time())
	assert.Equal(t, domain.PurityTypeDigitalized, story.Purity())
	assert.Equal(t, "Customer searches for a product", story.Trigger())
	assert.Len(t, story.Actors(), 5)
	assert.Len(t, story.WorkObjects(), 7)
	assert.Len(t, story.Sentences(), 12)
	assert.Len(t, story.Annotations(), 5)
	assert.Len(t, story.Variations(), 4)

	// Verify first actor details.
	actors := story.Actors()
	assert.Equal(t, "Customer", actors[0].Name())
	assert.Equal(t, domain.ActorTypePerson, actors[0].Type())
	assert.Equal(t, vo.UserStated, actors[0].Trust())

	// Verify actor with source (ai_researched).
	assert.Equal(t, "Payment Gateway", actors[2].Name())
	assert.Equal(t, domain.ActorTypeSystem, actors[2].Type())
	assert.Equal(t, vo.AIResearched, actors[2].Trust())
	assert.Equal(t, "Stripe/PayPal integration patterns", actors[2].Source())

	// Verify sentence with preposition.
	sentences := story.Sentences()
	assert.Equal(t, 2, sentences[1].Step())
	assert.Equal(t, "Customer", sentences[1].Subject())
	assert.Equal(t, "adds", sentences[1].Activity())
	assert.Equal(t, "Product Listing", sentences[1].Object())
	assert.Equal(t, "to", sentences[1].Preposition())
	assert.Equal(t, "Shopping Cart", sentences[1].IndirectObject())

	// Verify annotation with sentence ref.
	annotations := story.Annotations()
	assert.Equal(t, "Customer must be authenticated before checkout", annotations[0].Text())

	ref := annotations[0].SentenceRef()
	require.NotNil(t, ref)
	assert.Equal(t, 3, *ref)
	assert.Equal(t, domain.AnnotationTypeConstraint, annotations[0].Type())

	// Verify story-wide annotation (no sentence ref).
	assert.Equal(t, "Payment must be authorized before Order is created", annotations[1].Text())
	assert.Nil(t, annotations[1].SentenceRef())
	assert.Equal(t, domain.AnnotationTypeInvariant, annotations[1].Type())
}

func TestStoryYAMLParser_Parse_WhenVetclinicSample_ExpectValidStory(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join(sampleStoryDir(t), "vetclinic.story.yaml"))
	require.NoError(t, err)

	parser := &StoryYAMLParser{}

	story, err := parser.Parse(data)
	require.NoError(t, err)

	assert.Equal(t, "Pet Owner Brings Pet for Examination", story.Title())
	assert.Equal(t, domain.StoryTypeCoarseGrained, story.Type())
	assert.Equal(t, domain.TimeTypeToBe, story.Time())
	assert.Equal(t, domain.PurityTypePure, story.Purity())
	assert.Equal(t, "Pet Owner calls to book appointment", story.Trigger())
	assert.Len(t, story.Actors(), 5)
	assert.Len(t, story.WorkObjects(), 8)
	assert.Len(t, story.Sentences(), 12)
	assert.Len(t, story.Annotations(), 5)
	assert.Len(t, story.Variations(), 4)

	// Verify group actor type.
	actors := story.Actors()
	assert.Equal(t, "Pharmacy", actors[4].Name())
	assert.Equal(t, domain.ActorTypeGroup, actors[4].Type())
}

func TestStoryYAMLParser_Parse_WhenAltoSample_ExpectValidStory(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join(sampleStoryDir(t), "alto.story.yaml"))
	require.NoError(t, err)

	parser := &StoryYAMLParser{}

	story, err := parser.Parse(data)
	require.NoError(t, err)

	assert.Equal(t, "Developer Bootstraps New Project with alto", story.Title())
	assert.Len(t, story.Actors(), 5)
	assert.Len(t, story.WorkObjects(), 8)
	assert.Len(t, story.Sentences(), 15)
	assert.Len(t, story.Annotations(), 5)
	assert.Len(t, story.Variations(), 4)

	// Verify multi-word preposition "based on".
	sentences := story.Sentences()
	assert.Equal(t, "based on", sentences[3].Preposition())
	assert.Equal(t, "README", sentences[3].IndirectObject())
}

func TestStoryYAMLParser_Parse_WhenInvalidInput_ExpectError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: "title",
		},
		{
			name:    "invalid YAML syntax",
			input:   "{{invalid yaml",
			wantErr: "parsing story YAML",
		},
		{
			name:    "missing title",
			input:   minimalStoryYAML("", "coarse_grained", "as_is", "pure", "trigger"),
			wantErr: "title",
		},
		{
			name:    "missing type",
			input:   minimalStoryYAML("Title", "", "as_is", "pure", "trigger"),
			wantErr: "type",
		},
		{
			name:    "missing time",
			input:   minimalStoryYAML("Title", "coarse_grained", "", "pure", "trigger"),
			wantErr: "time",
		},
		{
			name:    "missing purity",
			input:   minimalStoryYAML("Title", "coarse_grained", "as_is", "", "trigger"),
			wantErr: "purity",
		},
		{
			name:    "missing trigger",
			input:   minimalStoryYAML("Title", "coarse_grained", "as_is", "pure", ""),
			wantErr: "trigger",
		},
		{
			name: "empty actors",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors: []
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`,
			wantErr: "actors",
		},
		{
			name: "empty work_objects",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects: []
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`,
			wantErr: "work_objects",
		},
		{
			name: "empty sentences",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences: []
`,
			wantErr: "sentences",
		},
		{
			name: "invalid story type",
			input: `
title: Test
type: medium_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`,
			wantErr: "type",
		},
		{
			name: "invalid actor type",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: robot
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`,
			wantErr: "actors[0].type",
		},
		{
			name: "invalid work object type",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: spreadsheet
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`,
			wantErr: "work_objects[0].type",
		},
		{
			name: "invalid preposition",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
  - name: WO2
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    preposition: towards
    indirect_object: WO2
    trust: user_stated
`,
			wantErr: "sentences[0]",
		},
		{
			name: "missing source for ai_researched actor",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: ai_researched
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`,
			wantErr: "actors[0]",
		},
		{
			name: "invalid annotation type",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
annotations:
  - text: Some note
    type: observation
    trust: user_stated
`,
			wantErr: "annotations[0].type",
		},
		{
			name: "annotation with zero sentence ref",
			input: `
title: Test
type: coarse_grained
time: as_is
purity: pure
trigger: something
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
annotations:
  - text: Some note
    type: constraint
    sentence: 0
    trust: user_stated
`,
			wantErr: "annotations[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := &StoryYAMLParser{}

			_, err := parser.Parse([]byte(tt.input))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestStoryYAMLParser_Parse_WhenAnnotationTrustOmitted_ExpectDefaultToUserStated(t *testing.T) {
	t.Parallel()

	input := `
title: Test Story
type: coarse_grained
time: as_is
purity: pure
trigger: something happens
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
annotations:
  - text: Some annotation without trust
    type: constraint
`

	parser := &StoryYAMLParser{}

	story, err := parser.Parse([]byte(input))
	require.NoError(t, err)

	annotations := story.Annotations()
	require.Len(t, annotations, 1)
	assert.Equal(t, vo.UserStated, annotations[0].Trust())
}

func TestStoryYAMLParser_Parse_WhenVersionMissing_ExpectNoError(t *testing.T) {
	t.Parallel()

	input := `
title: Test Story
type: coarse_grained
time: as_is
purity: pure
trigger: something happens
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`

	parser := &StoryYAMLParser{}

	story, err := parser.Parse([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, "Test Story", story.Title())
}

func TestStoryYAMLParser_Parse_WhenEmptyOptionalLists_ExpectNoError(t *testing.T) {
	t.Parallel()

	input := `
title: Test Story
type: coarse_grained
time: as_is
purity: pure
trigger: something happens
actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`

	parser := &StoryYAMLParser{}

	story, err := parser.Parse([]byte(input))
	require.NoError(t, err)
	assert.Empty(t, story.Annotations())
	assert.Empty(t, story.Variations())
}

func TestStoryYAMLParser_Serialize_WhenValidStory_ExpectVersionOne(t *testing.T) {
	t.Parallel()

	story := buildMinimalStory(t)

	parser := &StoryYAMLParser{}

	data, err := parser.Serialize(story)
	require.NoError(t, err)
	assert.Contains(t, string(data), "version: 1")
}

func TestStoryYAMLParser_RoundTrip_WhenSampleFiles_ExpectFieldPreservation(t *testing.T) {
	t.Parallel()

	samples := []string{
		"ecommerce.story.yaml",
		"vetclinic.story.yaml",
		"alto.story.yaml",
	}

	for _, sample := range samples {
		t.Run(sample, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(sampleStoryDir(t), sample))
			require.NoError(t, err)

			parser := &StoryYAMLParser{}

			// Parse original.
			story1, err := parser.Parse(data)
			require.NoError(t, err)

			// Serialize.
			serialized, err := parser.Serialize(story1)
			require.NoError(t, err)

			// Parse serialized.
			story2, err := parser.Parse(serialized)
			require.NoError(t, err)

			// Compare all fields.
			assertDomainStoryEqual(t, story1, story2)
		})
	}
}

func TestStoryYAMLParser_Read_WhenNonExistentFile_ExpectError(t *testing.T) {
	t.Parallel()

	parser := &StoryYAMLParser{}

	_, err := parser.Read(context.Background(), "/nonexistent/path/story.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading story file")
}

func TestStoryYAMLParser_ReadWrite_WhenRoundTrip_ExpectFieldPreservation(t *testing.T) {
	t.Parallel()

	story := buildMinimalStory(t)
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.story.yaml")

	parser := &StoryYAMLParser{}

	err := parser.Write(context.Background(), path, story)
	require.NoError(t, err)

	loaded, err := parser.Read(context.Background(), path)
	require.NoError(t, err)

	assertDomainStoryEqual(t, story, loaded)
}

func TestStoryYAMLParser_Read_WhenContextCancelled_ExpectError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parser := &StoryYAMLParser{}

	_, err := parser.Read(ctx, "any-path.yaml")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestStoryYAMLParser_Write_WhenContextCancelled_ExpectError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	story := buildMinimalStory(t)

	parser := &StoryYAMLParser{}

	err := parser.Write(ctx, filepath.Join(t.TempDir(), "test.yaml"), story)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// buildMinimalStory creates a minimal valid DomainStory for testing.
func buildMinimalStory(t *testing.T) *domain.DomainStory {
	t.Helper()

	st, err := domain.NewStoryType("coarse_grained")
	require.NoError(t, err)

	tt, err := domain.NewTimeType("as_is")
	require.NoError(t, err)

	pt, err := domain.NewPurityType("pure")
	require.NoError(t, err)

	story, err := domain.NewDomainStory("Test Story", st, tt, pt, "something happens")
	require.NoError(t, err)

	actor, err := domain.NewStoryActor("Actor1", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := domain.NewWorkObject("WO1", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := domain.NewStorySentence(1, "Actor1", "does", "WO1", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	return story
}

// minimalStoryYAML builds a minimal story YAML string with the given top-level fields.
// Empty string values for a field cause it to be omitted, triggering validation errors.
func minimalStoryYAML(title, storyType, timeType, purity, trigger string) string {
	var yaml string

	if title != "" {
		yaml += "title: " + title + "\n"
	}

	if storyType != "" {
		yaml += "type: " + storyType + "\n"
	}

	if timeType != "" {
		yaml += "time: " + timeType + "\n"
	}

	if purity != "" {
		yaml += "purity: " + purity + "\n"
	}

	if trigger != "" {
		yaml += "trigger: " + trigger + "\n"
	}

	yaml += `actors:
  - name: Actor1
    type: person
    trust: user_stated
work_objects:
  - name: WO1
    type: document
    trust: user_stated
sentences:
  - step: 1
    subject: Actor1
    activity: does
    object: WO1
    trust: user_stated
`

	return yaml
}

// assertDomainStoryEqual compares two DomainStory instances field by field.
func assertDomainStoryEqual(t *testing.T, expected, actual *domain.DomainStory) {
	t.Helper()

	assert.Equal(t, expected.Title(), actual.Title(), "title mismatch")
	assert.Equal(t, expected.Type(), actual.Type(), "type mismatch")
	assert.Equal(t, expected.Time(), actual.Time(), "time mismatch")
	assert.Equal(t, expected.Purity(), actual.Purity(), "purity mismatch")
	assert.Equal(t, expected.Trigger(), actual.Trigger(), "trigger mismatch")

	// Actors.
	expectedActors := expected.Actors()
	actualActors := actual.Actors()
	require.Len(t, actualActors, len(expectedActors), "actor count mismatch")

	for i := range expectedActors {
		assert.Equal(t, expectedActors[i].Name(), actualActors[i].Name(), "actor[%d].name", i)
		assert.Equal(t, expectedActors[i].Type(), actualActors[i].Type(), "actor[%d].type", i)
		assert.Equal(t, expectedActors[i].Trust(), actualActors[i].Trust(), "actor[%d].trust", i)
		assert.Equal(t, expectedActors[i].Source(), actualActors[i].Source(), "actor[%d].source", i)
	}

	// Work objects.
	expectedWOs := expected.WorkObjects()
	actualWOs := actual.WorkObjects()
	require.Len(t, actualWOs, len(expectedWOs), "work object count mismatch")

	for i := range expectedWOs {
		assert.Equal(t, expectedWOs[i].Name(), actualWOs[i].Name(), "work_objects[%d].name", i)
		assert.Equal(t, expectedWOs[i].Type(), actualWOs[i].Type(), "work_objects[%d].type", i)
		assert.Equal(t, expectedWOs[i].Trust(), actualWOs[i].Trust(), "work_objects[%d].trust", i)
		assert.Equal(t, expectedWOs[i].Source(), actualWOs[i].Source(), "work_objects[%d].source", i)
	}

	// Sentences.
	expectedSentences := expected.Sentences()
	actualSentences := actual.Sentences()
	require.Len(t, actualSentences, len(expectedSentences), "sentence count mismatch")

	for i := range expectedSentences {
		assert.Equal(t, expectedSentences[i].Step(), actualSentences[i].Step(), "sentences[%d].step", i)
		assert.Equal(t, expectedSentences[i].Subject(), actualSentences[i].Subject(), "sentences[%d].subject", i)
		assert.Equal(t, expectedSentences[i].Activity(), actualSentences[i].Activity(), "sentences[%d].activity", i)
		assert.Equal(t, expectedSentences[i].Object(), actualSentences[i].Object(), "sentences[%d].object", i)
		assert.Equal(t, expectedSentences[i].Preposition(), actualSentences[i].Preposition(), "sentences[%d].preposition", i)
		assert.Equal(t, expectedSentences[i].IndirectObject(), actualSentences[i].IndirectObject(), "sentences[%d].indirect_object", i)
		assert.Equal(t, expectedSentences[i].Trust(), actualSentences[i].Trust(), "sentences[%d].trust", i)
		assert.Equal(t, expectedSentences[i].Source(), actualSentences[i].Source(), "sentences[%d].source", i)
	}

	// Annotations.
	expectedAnns := expected.Annotations()
	actualAnns := actual.Annotations()
	require.Len(t, actualAnns, len(expectedAnns), "annotation count mismatch")

	for i := range expectedAnns {
		assert.Equal(t, expectedAnns[i].Text(), actualAnns[i].Text(), "annotations[%d].text", i)
		assert.Equal(t, expectedAnns[i].Type(), actualAnns[i].Type(), "annotations[%d].type", i)
		assert.Equal(t, expectedAnns[i].Trust(), actualAnns[i].Trust(), "annotations[%d].trust", i)
		assert.Equal(t, expectedAnns[i].Source(), actualAnns[i].Source(), "annotations[%d].source", i)

		expRef := expectedAnns[i].SentenceRef()
		actRef := actualAnns[i].SentenceRef()

		if expRef == nil {
			assert.Nil(t, actRef, "annotations[%d].sentence_ref should be nil", i)
		} else {
			require.NotNil(t, actRef, "annotations[%d].sentence_ref should not be nil", i)
			assert.Equal(t, *expRef, *actRef, "annotations[%d].sentence_ref", i)
		}
	}

	// Variations.
	assert.Equal(t, expected.Variations(), actual.Variations(), "variations mismatch")
}
