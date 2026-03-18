package infrastructure_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	challengedomain "github.com/alto-cli/alto/internal/challenge/domain"
	challengeinfra "github.com/alto-cli/alto/internal/challenge/infrastructure"
)

// --- ParseDDDVersionFromContent tests ---

func TestParseDDDVersionFromContent_WithFrontmatter(t *testing.T) {
	t.Parallel()

	content := `---
version: 2
round: challenge
updated: 2026-03-10
convergence_delta: 3
---

# Domain Model

Some content here.
`

	version, err := challengeinfra.ParseDDDVersionFromContent(content)

	require.NoError(t, err)
	assert.Equal(t, 2, version.Version())
	assert.Equal(t, "challenge", version.Round())
	assert.Equal(t, 3, version.ConvergenceDelta())
	assert.Equal(t, "2026-03-10", version.Updated())
}

func TestParseDDDVersionFromContent_NoFrontmatter(t *testing.T) {
	t.Parallel()

	content := `# Domain Model

No frontmatter here, just content.
`

	version, err := challengeinfra.ParseDDDVersionFromContent(content)

	require.NoError(t, err)
	// Should return version 0 (unversioned)
	assert.Equal(t, 0, version.Version())
	assert.Empty(t, version.Round())
}

func TestParseDDDVersionFromContent_EmptyContent(t *testing.T) {
	t.Parallel()

	version, err := challengeinfra.ParseDDDVersionFromContent("")

	require.NoError(t, err)
	assert.Equal(t, 0, version.Version())
}

func TestParseDDDVersionFromContent_OnlyFrontmatterDelimiters(t *testing.T) {
	t.Parallel()

	content := `---
---
# Content
`

	version, err := challengeinfra.ParseDDDVersionFromContent(content)

	require.NoError(t, err)
	assert.Equal(t, 0, version.Version())
}

func TestParseDDDVersionFromContent_PartialFrontmatter(t *testing.T) {
	t.Parallel()

	// Only version, no other fields
	content := `---
version: 1
---

# Content
`

	version, err := challengeinfra.ParseDDDVersionFromContent(content)

	require.NoError(t, err)
	assert.Equal(t, 1, version.Version())
	assert.Empty(t, version.Round())
	assert.Equal(t, 0, version.ConvergenceDelta())
}

func TestParseDDDVersionFromContent_MalformedYAML(t *testing.T) {
	t.Parallel()

	content := `---
version: not_a_number
round: [invalid yaml
---
`

	_, err := challengeinfra.ParseDDDVersionFromContent(content)

	assert.Error(t, err)
}

func TestParseDDDVersionFromContent_UnclosedFrontmatter(t *testing.T) {
	t.Parallel()

	// Frontmatter started but never closed
	content := `---
version: 1
# This is content, not frontmatter end
`

	version, err := challengeinfra.ParseDDDVersionFromContent(content)

	// Should treat as no frontmatter (unclosed)
	require.NoError(t, err)
	assert.Equal(t, 0, version.Version())
}

func TestParseDDDVersionFromContent_ExtraFieldsIgnored(t *testing.T) {
	t.Parallel()

	content := `---
version: 2
round: challenge
updated: 2026-03-10
convergence_delta: 3
extra_field: should be ignored
another: also ignored
---

# Content
`

	version, err := challengeinfra.ParseDDDVersionFromContent(content)

	require.NoError(t, err)
	assert.Equal(t, 2, version.Version())
}

func TestParseDDDVersionFromContent_WhitespaceHandling(t *testing.T) {
	t.Parallel()

	// Extra whitespace around delimiters
	content := `---
version: 1
round: express
updated: 2026-03-10
---

# Content
`

	version, err := challengeinfra.ParseDDDVersionFromContent(content)

	require.NoError(t, err)
	assert.Equal(t, 1, version.Version())
	assert.Equal(t, "express", version.Round())
}

// --- ApplyVersionToContent tests ---

func TestApplyVersionToContent_WithFrontmatter(t *testing.T) {
	t.Parallel()

	content := `---
version: 1
round: express
updated: 2026-03-01
convergence_delta: 0
---

# Domain Model

Content here.
`

	version := challengedomain.NewDDDVersion(2, "challenge", "2026-03-12", 5)
	result := challengeinfra.ApplyVersionToContent(content, version)

	// Should replace frontmatter
	assert.Contains(t, result, "version: 2")
	assert.Contains(t, result, "round: challenge")
	assert.Contains(t, result, "convergence_delta: 5")
	assert.Contains(t, result, "# Domain Model")
	assert.Contains(t, result, "Content here.")
	// Old values should be gone
	assert.NotContains(t, result, "version: 1")
	assert.NotContains(t, result, "round: express")
}

func TestApplyVersionToContent_WithoutFrontmatter(t *testing.T) {
	t.Parallel()

	content := `# Domain Model

Content here.
`

	version := challengedomain.NewDDDVersion(1, "express", "2026-03-12", 0)
	result := challengeinfra.ApplyVersionToContent(content, version)

	// Should prepend frontmatter
	assert.Greater(t, len(result), len(content))
	assert.Contains(t, result, "---")
	assert.Contains(t, result, "version: 1")
	assert.Contains(t, result, "# Domain Model")
}

func TestApplyVersionToContent_ToEmptyContent(t *testing.T) {
	t.Parallel()

	version := challengedomain.NewDDDVersion(1, "express", "2026-03-12", 0)
	result := challengeinfra.ApplyVersionToContent("", version)

	assert.Contains(t, result, "version: 1")
	assert.Contains(t, result, "---")
}

func TestApplyVersionToContent_PreservesBody(t *testing.T) {
	t.Parallel()

	content := `---
version: 1
---

# Domain Model

## Bounded Contexts

- Sales
- Shipping

## Aggregates

Important content that must be preserved.
`

	version := challengedomain.NewDDDVersion(2, "challenge", "2026-03-12", 3)
	result := challengeinfra.ApplyVersionToContent(content, version)

	// All original content should be preserved
	assert.Contains(t, result, "# Domain Model")
	assert.Contains(t, result, "## Bounded Contexts")
	assert.Contains(t, result, "- Sales")
	assert.Contains(t, result, "- Shipping")
	assert.Contains(t, result, "## Aggregates")
	assert.Contains(t, result, "Important content that must be preserved.")
}

func TestApplyVersionToContent_RoundTrip(t *testing.T) {
	t.Parallel()

	// Parse -> Increment -> Apply -> Parse should be consistent
	original := `---
version: 1
round: express
updated: 2026-03-01
convergence_delta: 0
---

# Domain Model
`

	v1, err := challengeinfra.ParseDDDVersionFromContent(original)
	require.NoError(t, err)

	v2 := v1.Increment("challenge", 5, time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC))
	updated := challengeinfra.ApplyVersionToContent(original, v2)

	v2Parsed, err := challengeinfra.ParseDDDVersionFromContent(updated)
	require.NoError(t, err)

	assert.Equal(t, v2.Version(), v2Parsed.Version())
	assert.Equal(t, v2.Round(), v2Parsed.Round())
	assert.Equal(t, v2.ConvergenceDelta(), v2Parsed.ConvergenceDelta())
}

// --- YAMLFrontmatterParser interface tests ---

func TestYAMLFrontmatterParser_ImplementsInterface(t *testing.T) {
	t.Parallel()

	parser := challengeinfra.NewYAMLFrontmatterParser()

	// Test ParseVersion
	version, err := parser.ParseVersion(`---
version: 3
round: simulate
---
# Content
`)
	require.NoError(t, err)
	assert.Equal(t, 3, version.Version())
	assert.Equal(t, "simulate", version.Round())

	// Test ApplyVersion
	newVersion := challengedomain.NewDDDVersion(4, "challenge", "2026-03-15", 2)
	result := parser.ApplyVersion("# Content", newVersion)
	assert.Contains(t, result, "version: 4")
	assert.Contains(t, result, "# Content")
}
