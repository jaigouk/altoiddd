package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
)

// --- ParseDDDVersion tests ---

func TestParseDDDVersion_WithFrontmatter(t *testing.T) {
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

	version, err := challengedomain.ParseDDDVersion(content)

	require.NoError(t, err)
	assert.Equal(t, 2, version.Version())
	assert.Equal(t, "challenge", version.Round())
	assert.Equal(t, 3, version.ConvergenceDelta())
	assert.Equal(t, "2026-03-10", version.Updated())
}

func TestParseDDDVersion_NoFrontmatter(t *testing.T) {
	t.Parallel()

	content := `# Domain Model

No frontmatter here, just content.
`

	version, err := challengedomain.ParseDDDVersion(content)

	require.NoError(t, err)
	// Should return version 0 (unversioned)
	assert.Equal(t, 0, version.Version())
	assert.Empty(t, version.Round())
}

func TestParseDDDVersion_EmptyContent(t *testing.T) {
	t.Parallel()

	version, err := challengedomain.ParseDDDVersion("")

	require.NoError(t, err)
	assert.Equal(t, 0, version.Version())
}

func TestParseDDDVersion_OnlyFrontmatterDelimiters(t *testing.T) {
	t.Parallel()

	content := `---
---
# Content
`

	version, err := challengedomain.ParseDDDVersion(content)

	require.NoError(t, err)
	assert.Equal(t, 0, version.Version())
}

func TestParseDDDVersion_PartialFrontmatter(t *testing.T) {
	t.Parallel()

	// Only version, no other fields
	content := `---
version: 1
---

# Content
`

	version, err := challengedomain.ParseDDDVersion(content)

	require.NoError(t, err)
	assert.Equal(t, 1, version.Version())
	assert.Empty(t, version.Round())
	assert.Equal(t, 0, version.ConvergenceDelta())
}

func TestParseDDDVersion_MalformedYAML(t *testing.T) {
	t.Parallel()

	content := `---
version: not_a_number
round: [invalid yaml
---
`

	_, err := challengedomain.ParseDDDVersion(content)

	assert.Error(t, err)
}

func TestParseDDDVersion_UnclosedFrontmatter(t *testing.T) {
	t.Parallel()

	// Frontmatter started but never closed
	content := `---
version: 1
# This is content, not frontmatter end
`

	version, err := challengedomain.ParseDDDVersion(content)

	// Should treat as no frontmatter (unclosed)
	require.NoError(t, err)
	assert.Equal(t, 0, version.Version())
}

// --- DDDVersion.Increment tests ---

func TestDDDVersion_Increment(t *testing.T) {
	t.Parallel()

	original, _ := challengedomain.ParseDDDVersion(`---
version: 1
round: express
updated: 2026-03-01
convergence_delta: 0
---
`)

	updated := original.Increment("challenge", 5, time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, 2, updated.Version())
	assert.Equal(t, "challenge", updated.Round())
	assert.Equal(t, 5, updated.ConvergenceDelta())
	assert.Equal(t, "2026-03-12", updated.Updated())
}

func TestDDDVersion_IncrementFromZero(t *testing.T) {
	t.Parallel()

	// Starting from unversioned document
	original, _ := challengedomain.ParseDDDVersion("# Just content")

	updated := original.Increment("express", 0, time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, 1, updated.Version())
	assert.Equal(t, "express", updated.Round())
}

func TestDDDVersion_IncrementPreservesImmutability(t *testing.T) {
	t.Parallel()

	original, _ := challengedomain.ParseDDDVersion(`---
version: 1
round: express
---
`)

	_ = original.Increment("challenge", 3, time.Now())

	// Original should be unchanged
	assert.Equal(t, 1, original.Version())
	assert.Equal(t, "express", original.Round())
}

// --- ApplyVersion tests ---

func TestApplyVersion_ToContentWithFrontmatter(t *testing.T) {
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
	result := challengedomain.ApplyVersion(content, version)

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

func TestApplyVersion_ToContentWithoutFrontmatter(t *testing.T) {
	t.Parallel()

	content := `# Domain Model

Content here.
`

	version := challengedomain.NewDDDVersion(1, "express", "2026-03-12", 0)
	result := challengedomain.ApplyVersion(content, version)

	// Should prepend frontmatter
	assert.Greater(t, len(result), len(content))
	assert.Contains(t, result, "---")
	assert.Contains(t, result, "version: 1")
	assert.Contains(t, result, "# Domain Model")
}

func TestApplyVersion_ToEmptyContent(t *testing.T) {
	t.Parallel()

	version := challengedomain.NewDDDVersion(1, "express", "2026-03-12", 0)
	result := challengedomain.ApplyVersion("", version)

	assert.Contains(t, result, "version: 1")
	assert.Contains(t, result, "---")
}

func TestApplyVersion_PreservesContentAfterFrontmatter(t *testing.T) {
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
	result := challengedomain.ApplyVersion(content, version)

	// All original content should be preserved
	assert.Contains(t, result, "# Domain Model")
	assert.Contains(t, result, "## Bounded Contexts")
	assert.Contains(t, result, "- Sales")
	assert.Contains(t, result, "- Shipping")
	assert.Contains(t, result, "## Aggregates")
	assert.Contains(t, result, "Important content that must be preserved.")
}

// --- NewDDDVersion constructor tests ---

func TestNewDDDVersion(t *testing.T) {
	t.Parallel()

	version := challengedomain.NewDDDVersion(3, "simulate", "2026-03-15", 7)

	assert.Equal(t, 3, version.Version())
	assert.Equal(t, "simulate", version.Round())
	assert.Equal(t, "2026-03-15", version.Updated())
	assert.Equal(t, 7, version.ConvergenceDelta())
}

// --- Edge cases ---

func TestParseDDDVersion_ExtraFieldsIgnored(t *testing.T) {
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

	version, err := challengedomain.ParseDDDVersion(content)

	require.NoError(t, err)
	assert.Equal(t, 2, version.Version())
}

func TestParseDDDVersion_WhitespaceHandling(t *testing.T) {
	t.Parallel()

	// Extra whitespace around delimiters
	content := `---
version: 1
round: express
updated: 2026-03-10
---

# Content
`

	version, err := challengedomain.ParseDDDVersion(content)

	require.NoError(t, err)
	assert.Equal(t, 1, version.Version())
	assert.Equal(t, "express", version.Round()) // Should trim whitespace
}

func TestApplyVersion_RoundTrip(t *testing.T) {
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

	v1, err := challengedomain.ParseDDDVersion(original)
	require.NoError(t, err)

	v2 := v1.Increment("challenge", 5, time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC))
	updated := challengedomain.ApplyVersion(original, v2)

	v2Parsed, err := challengedomain.ParseDDDVersion(updated)
	require.NoError(t, err)

	assert.Equal(t, v2.Version(), v2Parsed.Version())
	assert.Equal(t, v2.Round(), v2Parsed.Round())
	assert.Equal(t, v2.ConvergenceDelta(), v2Parsed.ConvergenceDelta())
}
