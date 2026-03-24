package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestGlossaryYAMLParser_Parse_SampleFile(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/research/samples/glossary.yaml")
	require.NoError(t, err)

	parser := &infrastructure.GlossaryYAMLParser{}
	entries, err := parser.Parse(data)
	require.NoError(t, err)
	assert.Len(t, entries, 12)

	// Spot-check first entry.
	assert.Equal(t, "Product Listing", entries[0].Term())
	assert.Equal(t, "Catalog", entries[0].Context())
	assert.Equal(t, vo.UserStated, entries[0].Trust())
	assert.Len(t, entries[0].Stories(), 1)

	// Check entry with aliases.
	assert.Equal(t, "Inventory", entries[1].Term())
	assert.Equal(t, []string{"stock"}, entries[1].Aliases())

	// Check entry with note.
	assert.Equal(t, "Shopping Cart", entries[2].Term())
	assert.Contains(t, entries[2].Note(), "Basket")

	// Check entry with source (ai_researched).
	assert.Equal(t, "Delivery Confirmation", entries[7].Term())
	assert.Equal(t, vo.AIResearched, entries[7].Trust())
	assert.NotEmpty(t, entries[7].Source())
}

func TestGlossaryYAMLParser_RoundTrip(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/research/samples/glossary.yaml")
	require.NoError(t, err)

	parser := &infrastructure.GlossaryYAMLParser{}

	// Parse original.
	entries1, err := parser.Parse(data)
	require.NoError(t, err)

	// Serialize.
	out, err := parser.Serialize(entries1)
	require.NoError(t, err)

	// Parse again.
	entries2, err := parser.Parse(out)
	require.NoError(t, err)

	// Compare field by field.
	require.Len(t, entries2, len(entries1))
	for i := range entries1 {
		assert.Equal(t, entries1[i].Term(), entries2[i].Term(), "term mismatch at index %d", i)
		assert.Equal(t, entries1[i].Definition(), entries2[i].Definition(), "definition mismatch at index %d", i)
		assert.Equal(t, entries1[i].Context(), entries2[i].Context(), "context mismatch at index %d", i)
		assert.Equal(t, entries1[i].Trust(), entries2[i].Trust(), "trust mismatch at index %d", i)
		assert.Equal(t, entries1[i].Source(), entries2[i].Source(), "source mismatch at index %d", i)
		assert.Equal(t, entries1[i].Aliases(), entries2[i].Aliases(), "aliases mismatch at index %d", i)
		assert.Equal(t, entries1[i].Note(), entries2[i].Note(), "note mismatch at index %d", i)
		assert.Equal(t, entries1[i].Stories(), entries2[i].Stories(), "stories mismatch at index %d", i)
	}
}

func TestGlossaryYAMLParser_Serialize_IncludesVersion(t *testing.T) {
	t.Parallel()
	entry, err := vo.NewUbiquitousLanguageEntry("Foo", "A foo", "TestCtx", vo.UserStated, "")
	require.NoError(t, err)
	entry = entry.WithStories([]string{"story1"})

	parser := &infrastructure.GlossaryYAMLParser{}
	out, err := parser.Serialize([]vo.UbiquitousLanguageEntry{entry})
	require.NoError(t, err)
	assert.Contains(t, string(out), "version: 1")
}

func TestGlossaryYAMLParser_Parse_MissingVersion_DefaultsToOne(t *testing.T) {
	t.Parallel()
	yaml := `terms:
  - term: Foo
    definition: A foo
    context: TestCtx
    trust: user_stated
    stories:
      - story1
`
	parser := &infrastructure.GlossaryYAMLParser{}
	entries, err := parser.Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Foo", entries[0].Term())
}

func TestGlossaryYAMLParser_Parse_MissingRequiredField(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		yaml string
	}{
		{"missing term", `terms:
  - definition: A foo
    context: TestCtx
    trust: user_stated
    stories:
      - story1`},
		{"missing definition", `terms:
  - term: Foo
    context: TestCtx
    trust: user_stated
    stories:
      - story1`},
		{"missing context", `terms:
  - term: Foo
    definition: A foo
    trust: user_stated
    stories:
      - story1`},
		{"missing trust", `terms:
  - term: Foo
    definition: A foo
    context: TestCtx
    stories:
      - story1`},
		{"missing stories", `terms:
  - term: Foo
    definition: A foo
    context: TestCtx
    trust: user_stated`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := &infrastructure.GlossaryYAMLParser{}
			_, err := parser.Parse([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "terms[0]")
		})
	}
}

func TestGlossaryYAMLParser_Parse_InvalidTrust(t *testing.T) {
	t.Parallel()
	yaml := `terms:
  - term: Foo
    definition: A foo
    context: TestCtx
    trust: invalid_trust
    stories:
      - story1
`
	parser := &infrastructure.GlossaryYAMLParser{}
	_, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestGlossaryYAMLParser_Parse_EmptyYAML(t *testing.T) {
	t.Parallel()
	parser := &infrastructure.GlossaryYAMLParser{}
	_, err := parser.Parse([]byte(""))
	require.Error(t, err)
}

func TestGlossaryYAMLParser_ReadWrite_RoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "glossary.yaml")

	entry, err := vo.NewUbiquitousLanguageEntry("Foo", "A foo", "TestCtx", vo.UserStated, "")
	require.NoError(t, err)
	entry = entry.WithStories([]string{"story1"}).WithAliases([]string{"Bar"}).WithNote("a note")

	parser := &infrastructure.GlossaryYAMLParser{}
	ctx := context.Background()

	err = parser.Write(ctx, path, []vo.UbiquitousLanguageEntry{entry})
	require.NoError(t, err)

	entries, err := parser.Read(ctx, path)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "Foo", entries[0].Term())
	assert.Equal(t, []string{"Bar"}, entries[0].Aliases())
	assert.Equal(t, "a note", entries[0].Note())
}

func TestGlossaryYAMLParser_Read_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parser := &infrastructure.GlossaryYAMLParser{}
	_, err := parser.Read(ctx, "/nonexistent")
	require.Error(t, err)
}

func TestGlossaryYAMLParser_Write_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parser := &infrastructure.GlossaryYAMLParser{}
	err := parser.Write(ctx, "/nonexistent", nil)
	require.Error(t, err)
}

func TestGlossaryYAMLParser_Read_NonExistentFile(t *testing.T) {
	t.Parallel()

	parser := &infrastructure.GlossaryYAMLParser{}
	_, err := parser.Read(context.Background(), "/nonexistent/glossary.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading glossary file")
}

func TestGlossaryYAMLParser_Parse_AIResearchedWithoutSource(t *testing.T) {
	t.Parallel()
	yaml := `terms:
  - term: Foo
    definition: A foo
    context: TestCtx
    trust: ai_researched
    stories:
      - story1
`
	parser := &infrastructure.GlossaryYAMLParser{}
	_, err := parser.Parse([]byte(yaml))
	require.Error(t, err)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "terms[0]")
}

func TestGlossaryYAMLParser_Parse_MalformedYAML(t *testing.T) {
	t.Parallel()

	parser := &infrastructure.GlossaryYAMLParser{}
	_, err := parser.Parse([]byte("{{invalid yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing glossary YAML")
}
