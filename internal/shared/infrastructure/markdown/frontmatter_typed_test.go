package markdown_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/infrastructure/markdown"
)

func TestParseGeneric_UnmarshalsScalarsArraysAndMaps(t *testing.T) {
	t.Parallel()

	raw := "name: foo\ncount: 3\ntags:\n  - a\n  - b\nnested:\n  key: value\n"
	m, err := markdown.ParseGeneric(raw)

	require.NoError(t, err)
	assert.Equal(t, "foo", m["name"])
	assert.Equal(t, 3, m["count"])
	tags, ok := m["tags"].([]any)
	require.True(t, ok, "tags should unmarshal as []any")
	assert.Equal(t, []any{"a", "b"}, tags)
	nested, ok := m["nested"].(map[string]any)
	require.True(t, ok, "nested should unmarshal as map[string]any")
	assert.Equal(t, "value", nested["key"])
}

func TestParseGeneric_WhenYAMLInvalid_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	raw := "version: not_a_number\nround: [invalid yaml"
	m, err := markdown.ParseGeneric(raw)

	require.Error(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "unmarshalling frontmatter")
}

type typedFixture struct {
	Name  string `yaml:"name"`
	Count int    `yaml:"count"`
	Flag  bool   `yaml:"flag"`
}

func TestParseTyped_UnmarshalsIntoCallerStruct(t *testing.T) {
	t.Parallel()

	raw := "name: foo\ncount: 7\nflag: true\n"
	got, err := markdown.ParseTyped[typedFixture](raw)

	require.NoError(t, err)
	assert.Equal(t, "foo", got.Name)
	assert.Equal(t, 7, got.Count)
	assert.True(t, got.Flag)
}

func TestParseTyped_WhenYAMLInvalid_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	raw := "name: foo\ncount: [invalid"
	_, err := markdown.ParseTyped[typedFixture](raw)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling frontmatter")
}
