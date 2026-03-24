package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- UbiquitousLanguageEntry constructor tests --

func TestNewUbiquitousLanguageEntry_Valid(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "A request to purchase goods", "Ordering", vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Order", entry.Term())
	assert.Equal(t, "A request to purchase goods", entry.Definition())
	assert.Equal(t, "Ordering", entry.Context())
	assert.Equal(t, vo.UserStated, entry.Trust())
	assert.Empty(t, entry.Source())
	assert.Empty(t, entry.Aliases())
	assert.Empty(t, entry.Note())
	assert.Empty(t, entry.Stories())
}

func TestNewUbiquitousLanguageEntry_EmptyTerm(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("", "def", "ctx", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_WhitespaceTerm(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("   ", "def", "ctx", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_EmptyDefinition(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("Order", "", "ctx", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_WhitespaceDefinition(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("Order", "   ", "ctx", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_EmptyContext(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("Order", "def", "", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_WhitespaceContext(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("Order", "def", "   ", vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_InvalidTrust(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.TrustLevel(99), "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_AIResearchedWithoutSource(t *testing.T) {
	t.Parallel()
	_, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.AIResearched, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewUbiquitousLanguageEntry_AIResearchedWithSource(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.AIResearched, "research-doc")
	require.NoError(t, err)
	assert.Equal(t, "research-doc", entry.Source())
	assert.Equal(t, vo.AIResearched, entry.Trust())
}

func TestNewUbiquitousLanguageEntry_UserStatedWithoutSource(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)
	assert.Empty(t, entry.Source())
}

func TestNewUbiquitousLanguageEntry_TrimsFields(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("  Order  ", "  def  ", "  ctx  ", vo.UserStated, "  src  ")
	require.NoError(t, err)
	assert.Equal(t, "Order", entry.Term())
	assert.Equal(t, "def", entry.Definition())
	assert.Equal(t, "ctx", entry.Context())
	assert.Equal(t, "src", entry.Source())
}

// -- With* immutable builder tests --

func TestUbiquitousLanguageEntry_WithAliases(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	aliases := []string{"Purchase", "Request"}
	withAliases := entry.WithAliases(aliases)

	// Original unchanged.
	assert.Empty(t, entry.Aliases())

	// New instance has aliases.
	assert.Equal(t, []string{"Purchase", "Request"}, withAliases.Aliases())

	// Preserves other fields.
	assert.Equal(t, "Order", withAliases.Term())
	assert.Equal(t, "def", withAliases.Definition())
	assert.Equal(t, "ctx", withAliases.Context())
	assert.Equal(t, vo.UserStated, withAliases.Trust())
}

func TestUbiquitousLanguageEntry_WithAliases_DefensiveCopy(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	aliases := []string{"Purchase", "Request"}
	withAliases := entry.WithAliases(aliases)

	// Mutate original slice.
	aliases[0] = "Hacked"
	assert.Equal(t, "Purchase", withAliases.Aliases()[0])

	// Mutate returned slice.
	got := withAliases.Aliases()
	got[0] = "Hacked"
	assert.Equal(t, "Purchase", withAliases.Aliases()[0])
}

func TestUbiquitousLanguageEntry_WithNote(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	withNote := entry.WithNote("important context")

	// Original unchanged.
	assert.Empty(t, entry.Note())

	// New instance has note.
	assert.Equal(t, "important context", withNote.Note())

	// Preserves other fields.
	assert.Equal(t, "Order", withNote.Term())
}

func TestUbiquitousLanguageEntry_WithStories(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	stories := []string{"Place Order", "Cancel Order"}
	withStories := entry.WithStories(stories)

	// Original unchanged.
	assert.Empty(t, entry.Stories())

	// New instance has stories.
	assert.Equal(t, []string{"Place Order", "Cancel Order"}, withStories.Stories())

	// Preserves other fields.
	assert.Equal(t, "Order", withStories.Term())
}

func TestUbiquitousLanguageEntry_WithStories_DefensiveCopy(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	stories := []string{"Place Order", "Cancel Order"}
	withStories := entry.WithStories(stories)

	// Mutate original slice.
	stories[0] = "Hacked"
	assert.Equal(t, "Place Order", withStories.Stories()[0])

	// Mutate returned slice.
	got := withStories.Stories()
	got[0] = "Hacked"
	assert.Equal(t, "Place Order", withStories.Stories()[0])
}

func TestUbiquitousLanguageEntry_WithChaining(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	result := entry.
		WithAliases([]string{"Purchase"}).
		WithNote("note").
		WithStories([]string{"Place Order"})

	assert.Equal(t, []string{"Purchase"}, result.Aliases())
	assert.Equal(t, "note", result.Note())
	assert.Equal(t, []string{"Place Order"}, result.Stories())

	// Original still empty.
	assert.Empty(t, entry.Aliases())
	assert.Empty(t, entry.Note())
	assert.Empty(t, entry.Stories())
}

// -- HasAlias tests --

func TestUbiquitousLanguageEntry_HasAlias_Found(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)
	entry = entry.WithAliases([]string{"Purchase", "Request"})

	assert.True(t, entry.HasAlias("Purchase"))
	assert.True(t, entry.HasAlias("Request"))
}

func TestUbiquitousLanguageEntry_HasAlias_CaseInsensitive(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)
	entry = entry.WithAliases([]string{"Purchase"})

	assert.True(t, entry.HasAlias("purchase"))
	assert.True(t, entry.HasAlias("PURCHASE"))
	assert.True(t, entry.HasAlias("Purchase"))
}

func TestUbiquitousLanguageEntry_HasAlias_NotFound(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)
	entry = entry.WithAliases([]string{"Purchase"})

	assert.False(t, entry.HasAlias("Refund"))
}

func TestUbiquitousLanguageEntry_HasAlias_EmptyAliases(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	assert.False(t, entry.HasAlias("anything"))
}

// -- String tests --

func TestUbiquitousLanguageEntry_String(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "Ordering", vo.UserStated, "")
	require.NoError(t, err)
	entry = entry.WithAliases([]string{"Ordering"})

	assert.Equal(t, "UbiquitousLanguageEntry: Order (Ordering, user_stated)", entry.String())
}

func TestUbiquitousLanguageEntry_String_NoAliases(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)

	assert.Equal(t, "UbiquitousLanguageEntry: Order (, user_stated)", entry.String())
}

func TestUbiquitousLanguageEntry_String_MultipleAliases(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Order", "def", "ctx", vo.UserStated, "")
	require.NoError(t, err)
	entry = entry.WithAliases([]string{"Purchase", "Request"})

	assert.Equal(t, "UbiquitousLanguageEntry: Order (Purchase, Request, user_stated)", entry.String())
}
