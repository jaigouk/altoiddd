package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	challengeapp "github.com/alto-cli/alto/internal/challenge/application"
	challengedomain "github.com/alto-cli/alto/internal/challenge/domain"
	challengeinfra "github.com/alto-cli/alto/internal/challenge/infrastructure"
)

// --- Mocks ---

type mockFileReader struct {
	content map[string]string
	err     error
}

func (m *mockFileReader) ReadFile(_ context.Context, path string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	content, ok := m.content[path]
	if !ok {
		return "", errors.New("file not found")
	}
	return content, nil
}

type mockFileWriter struct {
	written map[string]string
	err     error
}

func (m *mockFileWriter) WriteFile(_ context.Context, path, content string) error {
	if m.err != nil {
		return m.err
	}
	if m.written == nil {
		m.written = make(map[string]string)
	}
	m.written[path] = content
	return nil
}

// mockDDDVersionParser wraps the real parser for testing.
// Most tests need real YAML parsing behavior.
type mockDDDVersionParser struct {
	real     *challengeinfra.YAMLFrontmatterParser
	parseErr error
}

func newMockParser() *mockDDDVersionParser {
	return &mockDDDVersionParser{
		real: challengeinfra.NewYAMLFrontmatterParser(),
	}
}

func (m *mockDDDVersionParser) ParseVersion(content string) (challengedomain.DDDVersion, error) {
	if m.parseErr != nil {
		return challengedomain.DDDVersion{}, m.parseErr
	}
	return m.real.ParseVersion(content)
}

func (m *mockDDDVersionParser) ApplyVersion(content string, version challengedomain.DDDVersion) string {
	return m.real.ApplyVersion(content, version)
}

// --- VersionDDDDocument tests ---

func TestVersionDDDDocument_HappyPath(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": `---
version: 1
round: express
updated: 2026-03-01
convergence_delta: 0
---

# Domain Model

Content here.
`,
		},
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"challenge",
		5,
		time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Contains(t, writer.written, "docs/DDD.md")

	written := writer.written["docs/DDD.md"]
	assert.Contains(t, written, "version: 2")
	assert.Contains(t, written, "round: challenge")
	assert.Contains(t, written, "convergence_delta: 5")
	assert.Contains(t, written, "2026-03-12")
	assert.Contains(t, written, "# Domain Model")
	assert.Contains(t, written, "Content here.")
}

func TestVersionDDDDocument_FirstVersion(t *testing.T) {
	t.Parallel()

	// Document without frontmatter
	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": `# Domain Model

No version yet.
`,
		},
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"express",
		0,
		time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)

	written := writer.written["docs/DDD.md"]
	assert.Contains(t, written, "version: 1")
	assert.Contains(t, written, "round: express")
	assert.Contains(t, written, "# Domain Model")
}

func TestVersionDDDDocument_FileNotFound(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{
		content: map[string]string{}, // Empty - file not found
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"challenge",
		3,
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading DDD document")
}

func TestVersionDDDDocument_ReadError(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{
		err: errors.New("permission denied"),
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"challenge",
		3,
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading DDD document")
}

func TestVersionDDDDocument_WriteError(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": "# Content",
		},
	}
	writer := &mockFileWriter{
		err: errors.New("disk full"),
	}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"challenge",
		3,
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing DDD document")
}

func TestVersionDDDDocument_MalformedFrontmatter(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": `---
version: not_a_number
---
# Content
`,
		},
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"challenge",
		3,
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing version")
}

func TestVersionDDDDocument_IncrementMultipleTimes(t *testing.T) {
	t.Parallel()

	// Simulate v1 -> v2 -> v3
	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": `---
version: 2
round: challenge
updated: 2026-03-11
convergence_delta: 5
---

# Domain Model
`,
		},
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"simulate",
		2,
		time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)

	written := writer.written["docs/DDD.md"]
	assert.Contains(t, written, "version: 3")
	assert.Contains(t, written, "round: simulate")
	assert.Contains(t, written, "convergence_delta: 2")
}

func TestVersionDDDDocument_PreservesAllContent(t *testing.T) {
	t.Parallel()

	complexContent := `---
version: 1
round: express
---

# Domain Model

## Bounded Contexts

### Sales
- Handles orders
- Manages customers

### Shipping
- Tracks deliveries

## Aggregates

| Name | Context | Invariants |
|------|---------|------------|
| Order | Sales | Total > 0 |

## Code Examples

` + "```go\n" + `type Order struct {
    ID string
}
` + "```\n"

	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": complexContent,
		},
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"challenge",
		3,
		time.Now(),
	)

	require.NoError(t, err)

	written := writer.written["docs/DDD.md"]
	// All content sections must be preserved
	assert.Contains(t, written, "# Domain Model")
	assert.Contains(t, written, "## Bounded Contexts")
	assert.Contains(t, written, "### Sales")
	assert.Contains(t, written, "- Handles orders")
	assert.Contains(t, written, "### Shipping")
	assert.Contains(t, written, "## Aggregates")
	assert.Contains(t, written, "| Order | Sales | Total > 0 |")
	assert.Contains(t, written, "type Order struct")
}

func TestVersionDDDDocument_EmptyDocument(t *testing.T) {
	t.Parallel()

	reader := &mockFileReader{
		content: map[string]string{
			"docs/DDD.md": "",
		},
	}
	writer := &mockFileWriter{}

	handler := challengeapp.NewVersionHandler(reader, writer, newMockParser())
	err := handler.VersionDDDDocument(
		context.Background(),
		"docs/DDD.md",
		"express",
		0,
		time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)

	written := writer.written["docs/DDD.md"]
	assert.Contains(t, written, "version: 1")
	assert.Contains(t, written, "round: express")
}
