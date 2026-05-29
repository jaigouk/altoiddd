package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- titleCaseTokens tests ---

func TestTitleCase_HandlesMultipleWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"two-word lowercase", "context map", "Context Map"},
		{"three-word lowercase", "one two three", "One Two Three"},
		{"single word", "order", "Order"},
		{"already title-cased", "Context Map", "Context Map"},
		{"mixed case preserved after first rune", "cOnTeXt mAp", "COnTeXt MAp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, titleCaseTokens(tt.input))
		})
	}
}

func TestTitleCase_HandlesEmpty(t *testing.T) {
	t.Parallel()
	assert.Empty(t, titleCaseTokens(""))
}

// --- canonicalizeName tests ---

func TestCanonicalizeName_StripsUpdatedPrefix(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Context Map", canonicalizeName("Updated Context Map"))
}

func TestCanonicalizeName_CaseInsensitivePrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase updated", "updated context map", "Context Map"},
		{"uppercase updated", "UPDATED Context Map", "Context Map"},
		{"lowercase new", "NEW invoice", "Invoice"},
		{"lowercase modified", "modified order", "Order"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, canonicalizeName(tt.input))
		})
	}
}

func TestCanonicalizeName_PreservesCanonical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain canonical", "Context Map", "Context Map"},
		{"name containing updated but not prefixed", "Order Updated Log", "Order Updated Log"},
		{"name containing new but not prefixed", "Brand New Order", "Brand New Order"},
		{"name containing modified but not prefixed", "Pre-Modified Item", "Pre-Modified Item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, canonicalizeName(tt.input))
		})
	}
}

// --- deriveName tests ---

// makeStoryForDerive constructs a DomainStory with the given sentences for
// use in deriveName tests. Every sentence's subject must be in actors and
// object must be in actors or workObjects to satisfy aggregate invariants.
func makeStoryForDerive(
	t *testing.T,
	title string,
	actors, workObjects []string,
	sentences []struct {
		step     int
		subject  string
		activity string
		object   string
	},
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

	for _, name := range actors {
		actor, aErr := domain.NewStoryActor(name, domain.ActorTypePerson, vo.UserStated, "")
		require.NoError(t, aErr)
		require.NoError(t, story.AddActor(actor))
	}

	for _, name := range workObjects {
		wo, wErr := domain.NewWorkObject(name, domain.WorkObjectTypeDocument, vo.UserStated, "")
		require.NoError(t, wErr)
		require.NoError(t, story.AddWorkObject(wo))
	}

	for _, s := range sentences {
		sentence, sErr := domain.NewStorySentence(s.step, s.subject, s.activity, s.object, vo.UserStated, "")
		require.NoError(t, sErr)
		require.NoError(t, story.AddSentence(sentence))
	}

	require.NoError(t, story.Validate())

	return story
}

func TestDeriveName_PicksDominantByCount(t *testing.T) {
	t.Parallel()

	// "order" mentioned in 3 sentences, "invoice" in 1.
	story := makeStoryForDerive(t, "Dominant Story",
		[]string{"alice"},
		[]string{"order", "invoice"},
		[]struct {
			step     int
			subject  string
			activity string
			object   string
		}{
			{1, "alice", "creates", "order"},
			{2, "alice", "reviews", "order"},
			{3, "alice", "submits", "order"},
			{4, "alice", "issues", "invoice"},
		},
	)

	actors := map[string]struct{}{"alice": {}}
	workObjects := map[string]struct{}{"order": {}, "invoice": {}}

	name := deriveName([]*domain.DomainStory{story}, actors, workObjects, 1)
	assert.Equal(t, "Order", name)
}

func TestDeriveName_TieBreaksAlphabetically(t *testing.T) {
	t.Parallel()

	// "apple" and "banana" each mentioned in 2 sentences.
	story := makeStoryForDerive(t, "Tied Story",
		[]string{"alice"},
		[]string{"apple", "banana"},
		[]struct {
			step     int
			subject  string
			activity string
			object   string
		}{
			{1, "alice", "picks", "apple"},
			{2, "alice", "sells", "apple"},
			{3, "alice", "picks", "banana"},
			{4, "alice", "sells", "banana"},
		},
	)

	actors := map[string]struct{}{"alice": {}}
	workObjects := map[string]struct{}{"apple": {}, "banana": {}}

	name := deriveName([]*domain.DomainStory{story}, actors, workObjects, 1)
	assert.Equal(t, "Apple", name)
}

func TestDeriveName_MergesPrefixedVariantCounts(t *testing.T) {
	t.Parallel()

	// "Context Map" count=1, "Updated Context Map" count=2 -> merged to "Context Map"=3.
	// "Order" count=2. After merge: "Context Map" wins with 3.
	story := makeStoryForDerive(t, "Merging Story",
		[]string{"alice"},
		[]string{"context map", "updated context map", "order"},
		[]struct {
			step     int
			subject  string
			activity string
			object   string
		}{
			{1, "alice", "draws", "context map"},
			{2, "alice", "updates", "updated context map"},
			{3, "alice", "shares", "updated context map"},
			{4, "alice", "creates", "order"},
			{5, "alice", "submits", "order"},
		},
	)

	actors := map[string]struct{}{"alice": {}}
	workObjects := map[string]struct{}{
		"context map":         {},
		"updated context map": {},
		"order":               {},
	}

	name := deriveName([]*domain.DomainStory{story}, actors, workObjects, 1)
	assert.Equal(t, "Context Map", name)
}

func TestDeriveName_FallsBackToContextN(t *testing.T) {
	t.Parallel()

	// Empty actors AND workObjects -> "Context 3".
	name := deriveName(nil, map[string]struct{}{}, map[string]struct{}{}, 3)
	assert.Equal(t, "Context 3", name)
}
