package markdown_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/infrastructure/markdown"
)

func TestExtractFrontmatter_WhenNoLeadingDelimiter_ReturnsBodyAndFalse(t *testing.T) {
	t.Parallel()

	content := "# heading\nbody text\n"
	raw, body, has, err := markdown.ExtractFrontmatter(content)

	require.NoError(t, err)
	assert.False(t, has)
	assert.Empty(t, raw)
	assert.Equal(t, content, body)
}

func TestExtractFrontmatter_WhenEmpty_ReturnsEmptyAndFalse(t *testing.T) {
	t.Parallel()

	raw, body, has, err := markdown.ExtractFrontmatter("")

	require.NoError(t, err)
	assert.False(t, has)
	assert.Empty(t, raw)
	assert.Empty(t, body)
}

func TestExtractFrontmatter_WhenClosingDelimiterMissing_ReturnsErrMissingClosingDelimiter(t *testing.T) {
	t.Parallel()

	content := "---\nname: foo\nbody without close"
	raw, body, has, err := markdown.ExtractFrontmatter(content)

	require.ErrorIs(t, err, markdown.ErrMissingClosingDelimiter)
	assert.False(t, has)
	assert.Empty(t, raw)
	assert.Equal(t, content, body)
}

func TestExtractFrontmatter_WhenLeadingDelimiterWithoutNewline_TolerantParse(t *testing.T) {
	t.Parallel()

	// `---xyz\n---\nbody` — no newline after the opening delimiter. The
	// tolerant scanner treats "xyz" as part of the frontmatter and still
	// locates the closing `\n---` correctly.
	content := "---name: foo\n---\nbody"
	raw, body, has, err := markdown.ExtractFrontmatter(content)

	require.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, "name: foo", raw)
	assert.Equal(t, "body", body)
}

func TestExtractFrontmatter_StripsLeadingNewlineFromBody(t *testing.T) {
	t.Parallel()

	content := "---\nfoo: bar\n---\nbody line\n"
	raw, body, has, err := markdown.ExtractFrontmatter(content)

	require.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, "foo: bar", raw)
	assert.Equal(t, "body line\n", body)
}

func TestExtractFrontmatter_PreservesNumericKeysInRawForCallerUnmarshal(t *testing.T) {
	t.Parallel()

	// Caller (e.g., challenge's frontmatterData) needs the raw bytes
	// preserved so its typed Unmarshal sees the original `version: 2`
	// numeric value without any whitespace mangling beyond TrimSpace.
	content := "---\nversion: 2\nround: challenge\nconvergence_delta: 7\n---\nbody"
	raw, _, has, err := markdown.ExtractFrontmatter(content)

	require.NoError(t, err)
	assert.True(t, has)
	assert.Contains(t, raw, "version: 2")
	assert.Contains(t, raw, "round: challenge")
	assert.Contains(t, raw, "convergence_delta: 7")
}

func TestExtractFrontmatter_EmptyFrontmatterDelimitersOnly(t *testing.T) {
	t.Parallel()

	// `---\n---\nbody` — empty frontmatter between two delimiters.
	content := "---\n---\nbody"
	raw, body, has, err := markdown.ExtractFrontmatter(content)

	require.NoError(t, err)
	assert.True(t, has)
	assert.Empty(t, raw)
	assert.Equal(t, "body", body)
}

func TestExtractFrontmatter_OnlyDelimiterNoBody_ReturnsErrMissingClosingDelimiter(t *testing.T) {
	t.Parallel()

	// `---` alone with nothing after — no closing delimiter exists.
	content := "---"
	raw, body, has, err := markdown.ExtractFrontmatter(content)

	require.ErrorIs(t, err, markdown.ErrMissingClosingDelimiter)
	assert.False(t, has)
	assert.Empty(t, raw)
	assert.Equal(t, content, body)
}
